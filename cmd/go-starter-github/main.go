package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/magento-mcom/go-starter/pkg/console"
	"github.com/magento-mcom/go-starter/pkg/keychainx"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var version, commit string

func usage() {
	out := flag.CommandLine.Output()

	_, _ = fmt.Fprintf(out, "go-starter-github version %v (commit %v)\n", version, commit)
	_, _ = fmt.Fprintf(out, "\n")
	_, _ = fmt.Fprintf(out, "Usage: %s [flags] <github-org> <github-repo>\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nExample:\n")
	_, _ = fmt.Fprintf(out, "    %s magento-mcom awesome-project\n", os.Args[0])

	if flag.NFlag() != 0 {
		_, _ = fmt.Fprintf(out, "\nFlags:\n")
		flag.PrintDefaults()
	}

	_, _ = fmt.Fprintf(out, "\n")
}

func main() {
	var remote string
	var public, issues, projects, wiki bool

	flag.Usage = usage
	flag.StringVar(&remote, "remote", "upstream", "Name of the remote in local repository")
	flag.BoolVar(&issues, "with-issues", false, "Enable issues in GitHub")
	flag.BoolVar(&projects, "with-projects", false, "Enable projects in GitHub")
	flag.BoolVar(&wiki, "with-wiki", false, "Enable wiki page in GitHub")
	flag.BoolVar(&public, "public", false, "Make repository public")
	flag.Parse()

	ui := console.New(os.Stdin, os.Stdout)

	org, name := flag.Arg(0), flag.Arg(1)
	if org == "" {
		flag.Usage()
		ui.Fatalf("GitHub organisation is empty\n")
	}

	if name == "" {
		flag.Usage()
		ui.Fatalf("GitHub repository name is empty\n")
	}

	// get credentials from OSX keychain
	user, pass, err := keychainx.Load("github.com")
	if err != nil && err != keychainx.ErrNotFound {
		ui.Fatalf("An error occurred while loading keychain: %v\n", err)
	}

	if err == keychainx.ErrNotFound {
		user, pass = AskCredentials(ui)
		if err := keychainx.Save("github.com", user, pass); err != nil {
			ui.Errorf("An error occurred while saving keychain: %v\n", err)
		}
	}

	// build github client
	cli := github.NewClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &github.BasicAuthTransport{
			Username: user,
			Password: pass,
		},
	})

	// create repository
	repo, _, err := cli.Repositories.Create(context.Background(), org, &github.Repository{
		Name:        github.String(name),
		Private:     github.Bool(!public),
		HasIssues:   github.Bool(issues),
		HasProjects: github.Bool(projects),
		HasWiki:     github.Bool(wiki),
	})

	if err != nil {
		if !strings.Contains(err.Error(), "name already exists on this account") {
			ui.Fatalf("An error occurred while creating GitHub repository: %v\n", err)
		}

		ui.Printf("Repository %v/%v already exists\n", name, org)
		return
	}

	ui.Successf("New repository created at %v\n", *repo.HTMLURL)

	// init repository
	if err := run("git", "init"); err != nil {
		ui.Fatalf("An error occurred while running git init: %v\n", err)
	}

	// add all files
	if err := run("git", "add", "-A"); err != nil {
		ui.Fatalf("An error occurred while running git add: %v\n", err)
	}

	// remove starter files
	if err := run("git", "rm", "--cached", "--ignore-unmatch", ".starter*"); err != nil {
		ui.Fatalf("An error occurred while running git rm: %v\n", err)
	}

	// commit
	if err := run("git", "commit", "-m", "Initial commit"); err != nil {
		ui.Fatalf("An error occurred while running git commit: %v\n", err)
	}

	// add remote
	if err := run("git", "remote", "add", remote, *repo.CloneURL); err != nil {
		ui.Fatalf("An error occurred while running git remote add: %v\n", err)
	}

	// push to remote
	if err := run("git", "push", "--set-upstream", "upstream", "master"); err != nil {
		ui.Fatalf("An error occurred while running git push: %v\n", err)
	}
}

// Run a cli command
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}

func AskCredentials(ui *console.Console) (user string, pass string) {
	for {
		ui.Printf("Enter your GitHub username: ")
		ui.Scanln(&user)
		ui.Printf("\n")
		ui.Printf("Follow this link and generate personal token: https://git.io/fjisU. Scroll to the bottom of the page and click \"Generate token\".\n")
		ui.Printf("Enter your personal token: ")
		ui.Scanln(&pass)

		// build github client
		cli := github.NewClient(&http.Client{
			Timeout: 30 * time.Second,
			Transport: &github.BasicAuthTransport{
				Username: user,
				Password: pass,
			},
		})

		_, _, err := cli.Zen(context.Background())
		if err == nil {
			return
		}

		ui.Errorf("Credentials do not appear to be valid, try again...\n")
	}
}
