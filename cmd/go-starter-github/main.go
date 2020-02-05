/*
Copyright 2019 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/magento-mcom/go-starter/pkg/console"
	"github.com/magento-mcom/go-starter/pkg/keychainx"
	"io/ioutil"
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
	_, _ = fmt.Fprintf(out, "\nFlags:\n")
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(out, "\n")
}

func main() {
	var remote, deployKey string
	var public, issues, projects, wiki bool
	var collaborators SliceFlag

	flag.Usage = usage
	flag.StringVar(&remote, "remote", "upstream", "Name of the remote in local repository")
	flag.BoolVar(&issues, "with-issues", false, "Enable issues in GitHub")
	flag.BoolVar(&projects, "with-projects", false, "Enable projects in GitHub")
	flag.BoolVar(&wiki, "with-wiki", false, "Enable wiki page in GitHub")
	flag.StringVar(&deployKey, "deploy-key", "", "Add SSH deployment key to the repository, add ':rw' suffix to grant write permissions to the key")
	flag.BoolVar(&public, "public", false, "Make repository public")
	flag.Var(&collaborators, "collaborator", "Add collaborators to the repository by GitHub username. You can grant permissions using following format: <username>:<permission>. Permission can be: pull (read only), push (read and write) or admin (everything), default is push. Can be specified multiple times. Example: --collaborator octocat:pull")
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

		ui.Printf("Repository %v/%v already exists, skipping repository configuration...\n", org, name)
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

	for _, c := range collaborators {
		user, perm := SplitPermissions(c, "push")

		_, err := cli.Repositories.AddCollaborator(context.Background(), org, name, user, &github.RepositoryAddCollaboratorOptions{
			Permission: perm,
		})

		if err != nil {
			ui.Errorf("An error occurred while adding %#v collaborator: %v\n", c, err)
		}
	}

	if deployKey != "" {
		key, perm := SplitPermissions(deployKey, "ro")

		ui.Printf("Adding deployment key %#v with %#v permissions\n", key, perm)

		data, err := ioutil.ReadFile(key)
		if err != nil {
			ui.Fatalf("Unable to read deployment key: %v\n", err)
		}

		_, _, err = cli.Repositories.CreateKey(context.Background(), org, name, &github.Key{
			Title:    StringPtr("Deploy Key"),
			Key:      StringPtr(string(data)),
			ReadOnly: BoolPtr(perm != "rw"),
		})

		if err != nil {
			ui.Errorf("An error occurred while adding %#v deployment key: %v\n", key, err)
		}
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
		user = ui.ReadString("Enter your GitHub username: ")

		ui.Printf("\n")
		ui.Printf("Follow this link and generate personal token: https://git.io/fjisU. Scroll to the bottom of the page and click \"Generate token\".\n")

		pass = ui.ReadString("Enter your personal token: ")

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

		ui.Errorf("An error occurred while validating credentials: %v\n", err)
		ui.Errorf("Credentials do not appear to be valid, try again...\n")
	}
}

func SplitPermissions(c, d string) (string, string) {
	if parts := strings.SplitN(c, ":", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}

	return c, d
}

type SliceFlag []string

func (s *SliceFlag) Set(v string) error {
	values := strings.Split(v, ",")
	for _, v := range values {
		*s = append(*s, strings.TrimSpace(v))
	}

	return nil
}

func (s *SliceFlag) String() string {
	return strings.Join(*s, ",")
}

func BoolPtr(v bool) *bool {
	return &v
}

func StringPtr(v string) *string {
	return &v
}
