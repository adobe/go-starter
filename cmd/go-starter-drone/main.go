package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/drone/drone-go/drone"
	"github.com/magento-mcom/go-starter/pkg/console"
	"github.com/magento-mcom/go-starter/pkg/keychainx"
	"golang.org/x/oauth2"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

var version, commit string

func usage() {
	out := flag.CommandLine.Output()

	_, _ = fmt.Fprintf(out, "go-starter-drone version %v (commit %v)\n", version, commit)
	_, _ = fmt.Fprintf(out, "\n")
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
	flag.Usage = usage
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
	if _, err := dcli.RepoListSync(); err != nil {
		ui.Fatalf("An error occurred: %v\n", err)
	}

	ui.Printf("Activating repository in drone\n")
	if _, err = dcli.RepoEnable(org, repo); err != nil && !strings.Contains(err.Error(), "Repository is already active") {
		ui.Fatalf("An error occurred: %v\n", err)
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

// AskCredentials (token) for drone
func AskCredentials(ui *console.Console, u *url.URL) (pass string) {
	for {
		ui.Printf("Follow this link to get your Personal Token: %v://%v/account.\n", u.Scheme, u.Host)
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

// run a cli command
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}
