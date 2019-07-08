package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/drone/drone-go/drone"
	vault "github.com/hashicorp/vault/api"
	"github.com/magento-mcom/go-starter/pkg/console"
	"github.com/magento-mcom/go-starter/pkg/keychainx"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Registries []Registry
	Secrets    []Secret
}

type Registry struct {
	Addr      string `yaml:"addr"`       // Registry address, for example: docker.io
	Username  string `yaml:"username"`   // Username
	Email     string `yaml:"email"`      // Email (has been deprecated for a while now)
	VaultPath string `yaml:"vault_path"` // Value path to access password or token, vault value should have either password or token key
}

type Secret struct {
	Name      string   `yaml:"secret"`     // Secret name in drone
	Events    []string `yaml:"events"`     // Events where secret will be available (push, pull_request, deployment, tag)
	Images    []string `yaml:"images"`     // Docker images for while secret will be available
	VaultPath string   `yaml:"vault_path"` // Vault path to access secret value, vault value should have data key for value of the secret
	ValueKey  string   `yaml:"vault_key"`  // Vault key within path
}

func usage() {
	out := flag.CommandLine.Output()

	_, _ = fmt.Fprintf(out, "Usage: %s [flags] <drone-url> <github-org> <github-repo>\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nExample:\n")
	_, _ = fmt.Fprintf(out, "    %s magento-mcom awesome-project\n", os.Args[0])

	if flag.NFlag() != 0 {
		_, _ = fmt.Fprintf(out, "\nFlags:\n")
		flag.PrintDefaults()
	}

	_, _ = fmt.Fprintf(out, "\n")
}

func main() {
	var noDeploy, noPull, noPush, noTag bool
	var secretsFile string

	flag.Usage = usage
	flag.BoolVar(&noDeploy, "no-deploy", false, "Don't trigger build on deploy event")
	flag.BoolVar(&noPull, "no-pr", false, "Don't trigger build on pull request event")
	flag.BoolVar(&noPush, "no-push", false, "Don't trigger build on push event")
	flag.BoolVar(&noTag, "no-tag", false, "Don't trigger build on tag event")
	flag.StringVar(&secretsFile, "secrets-from", "", "Import secrets and docker registry credentials from a file")
	flag.Parse()

	ui := console.New(os.Stdin, os.Stdout)

	org, repo := flag.Arg(1), flag.Arg(2)

	uri, err := url.Parse(flag.Arg(0))
	if err != nil {
		flag.Usage()
		ui.Fatalf("An error occurred while parsing Drone URL: %v. Enter URL with protocol, for example: https://drone.bcn.magento.com.\n", err)
	}

	if uri.Host == "" {
		flag.Usage()
		ui.Fatalf("Drone hostname is empty. Enter URL with protocol, for example: https://drone.bcn.magento.com.\n")
	}

	if org == "" {
		flag.Usage()
		ui.Fatalf("GitHub organisation is empty\n")
	}

	if repo == "" {
		flag.Usage()
		ui.Fatalf("GitHub repository name is empty\n")
	}

	// get credentials from OSX keychain
	_, pass, err := keychainx.Load(uri.Host)
	if err != nil && err != keychainx.ErrNotFound {
		ui.Fatalf("An error occurred while reading Drone token from keychain: %v\n", err)
	}

	if err == keychainx.ErrNotFound {
		pass = AskCredentials(ui, uri)
		if err := keychainx.Save(uri.Host, "drone", pass); err != nil {
			ui.Errorf("An error occurred while writing Drone token to keychain: %v\n", err)
		}
	}

	// create an http client with oauth authentication.
	auther := new(oauth2.Config).Client(context.Background(), &oauth2.Token{AccessToken: pass})
	auther.Timeout = 30 * time.Second

	// create the drone client with authenticator
	dcli := drone.NewClient(uri.String(), auther)

	ui.Printf("Sync repository list in drone\n")
	sync(ui, uri, auther)

	ui.Printf("Activating repository in drone\n")
	_, err = dcli.RepoPost(org, repo)
	if err != nil && !strings.Contains(err.Error(), "Repository is already active") {
		ui.Fatalf("An error occurred: %v\n", err)
	}

	ui.Printf("Updating repository configuration\n")
	_, err = dcli.RepoPatch(org, repo, &drone.RepoPatch{
		AllowDeploy: ptrBool(!noDeploy),
		AllowPull:   ptrBool(!noPull),
		AllowPush:   ptrBool(!noPush),
		AllowTag:    ptrBool(!noTag),
	})

	if err != nil {
		ui.Fatalf("An error occurred: %v\n", err)
	}

	if secretsFile != "" {
		vcli, err := vault.NewClient(vault.DefaultConfig())
		if err != nil {
			ui.Fatalf("An error occurred while initiating vault client: %v\n", err)
		}

		ui.Titlef("Loading secrets file\n")
		config, err := load(secretsFile)
		if err != nil {
			ui.Fatalf("An error occurred while loading secrets file: %v\n", err)
		}

		if len(config.Registries) != 0 {
			ui.Printf("Configuring registry\n")
			for _, r := range config.Registries {
				if err := ConfigureRegistry(ui, dcli, vcli, org, repo, r); err != nil {
					ui.Fatalf("An error occurred: %v\n", err)
				}
			}
		}

		if len(config.Secrets) != 0 {
			ui.Printf("Importing secrets\n")
			for _, s := range config.Secrets {
				if err := ConfigureSecret(ui, dcli, vcli, org, repo, s); err != nil {
					ui.Fatalf("An error occurred: %v\n", err)
				}
			}
		}
	}

	ui.Titlef("Triggering build by making empty commit\n")
	if err := run("git", "commit", "--allow-empty", "-m", "Trigger CI build"); err != nil {
		ui.Fatalf("An error occurred when making commit: %v\n", err)
	}

	if err := run("git", "push"); err != nil {
		ui.Fatalf("An error occurred when pushing commit: %v\n", err)
	}

	// give drone few seconds to start build
	time.Sleep(3 * time.Second)

	// get last build
	if last, err := dcli.BuildLast(org, repo, ""); err == nil {
		ui.Printf("Build #%v started: %v://%v/%v/%v/%v\n", last.Number, uri.Scheme, uri.Host, org, repo, last.Number)
	}
}

// sync repository list in drone, unfortunately there is no method for syncing in SDK
func sync(ui *console.Console, uri *url.URL, cli *http.Client) {
	endpoint := strings.TrimSuffix(uri.String(), "/") + "/api/user/repos?all=true&flush=true"

	resp, err := cli.Get(endpoint)
	if err != nil {
		ui.Errorf("An error occurred: %v\n", err)
		return
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		ui.Errorf("An error occurred, returned status code is not 200, got %v instead\n", resp.StatusCode)
		return
	}
}

func ConfigureRegistry(ui *console.Console, dcli drone.Client, vcli *vault.Client, org string, repo string, r Registry) error {
	value, err := vcli.Logical().Read("secret/" + r.VaultPath)
	if err != nil {
		return err
	}

	reg, err := dcli.Registry(org, repo, r.Addr)
	if err != nil && !strings.Contains(err.Error(), "client error 404") {
		return err
	}

	if reg != nil && reg.ID != 0 {
		ui.Printf("  - Updating %#v\n", r.Addr)
		_, err := dcli.RegistryUpdate(org, repo, &drone.Registry{
			ID:       reg.ID,
			Address:  r.Addr,
			Username: r.Username,
			Email:    r.Email,
			Password: fmt.Sprint(value.Data["password"]),
			Token:    fmt.Sprint(value.Data["token"]),
		})

		return err
	}

	ui.Printf("  - Adding %#v\n", r.Addr)
	_, err = dcli.RegistryCreate(org, repo, &drone.Registry{
		Address:  r.Addr,
		Username: r.Username,
		Email:    r.Email,
		Password: fmt.Sprint(value.Data["password"]),
		Token:    fmt.Sprint(value.Data["token"]),
	})

	return err
}

func ConfigureSecret(ui *console.Console, dcli drone.Client, vcli *vault.Client, org string, repo string, s Secret) error {
	value, err := vcli.Logical().Read("secret/" + s.VaultPath)
	if err != nil {
		return err
	}

	key := s.ValueKey
	if key == "" {
		key = "data"
	}

	sec, err := dcli.Secret(org, repo, s.Name)
	if err != nil && !strings.Contains(err.Error(), "client error 404") {
		return err
	}

	if sec != nil && sec.ID != 0 {
		ui.Printf("  - Updating %#v\n", s.Name)
		_, err := dcli.SecretUpdate(org, repo, &drone.Secret{
			ID:     sec.ID,
			Name:   s.Name,
			Value:  fmt.Sprint(value.Data[key]),
			Images: s.Images,
			Events: s.Events,
		})

		return err
	}

	ui.Printf("  - Creating %#v\n", s.Name)
	_, err = dcli.SecretCreate(org, repo, &drone.Secret{
		Name:   s.Name,
		Value:  fmt.Sprint(value.Data[key]),
		Images: s.Images,
		Events: s.Events,
	})

	return err
}

// AskCredentials (token) for drone
func AskCredentials(ui *console.Console, u *url.URL) (pass string) {
	for {
		ui.Printf("Follow this link to get your Personal Token: %v://%v/account/token.\n", u.Scheme, u.Host)
		ui.Printf("Enter your personal token: ")
		ui.Scanln(&pass)

		// build drone client
		config := new(oauth2.Config)
		auther := config.Client(context.Background(), &oauth2.Token{
			AccessToken: pass,
		})

		// create the drone client with authenticator
		client := drone.NewClient(u.String(), auther)

		// gets the current user
		if _, err := client.Self(); err == nil {
			return
		}

		ui.Errorf("Credentials do not appear to be valid, try again...\n")
	}
}

// load drone secrets file
func load(file string) (c Config, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	return c, yaml.Unmarshal(data, &c)
}

// run a cli command
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}

func ptrBool(a bool) *bool {
	return &a
}
