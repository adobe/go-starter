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
	"github.com/drone/drone-go/drone"
	"github.com/magento-mcom/go-starter/pkg/console"
	"github.com/magento-mcom/go-starter/pkg/keychainx"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

var version, commit string

var (
	fileSecrets        SliceFlag
	fileSecretsPull    SliceFlag
	literalSecrets     SliceFlag
	literalSecretsPull SliceFlag
)

func usage() {
	out := flag.CommandLine.Output()
	_, _ = fmt.Fprintf(out, "go-starter-drone version %v (commit %v)\n", version, commit)
	_, _ = fmt.Fprintf(out, "\n")
	_, _ = fmt.Fprintf(out, "Usage: %s [flags] <drone-url> <github-org> <github-repo>\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nExample:\n")
	_, _ = fmt.Fprintf(out, "    %s https://cloud.drone.io magento-mcom awesome-project\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nFlags:\n")
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(out, "\n")
}

func main() {
	flag.Usage = usage
	flag.Var(&fileSecrets, "secret-file", "Create a secret from file (eq. --secret-file=secret_name=./path/to/file)")
	flag.Var(&fileSecretsPull, "pull-secret-file", "Create a secret from file available for pull-requests (eq. --pull-secret-file=secret_name=./path/to/file)")
	flag.Var(&literalSecrets, "secret-literal", "Create a secret from literal (eq. --secret-literal=secret_name=value)")
	flag.Var(&literalSecretsPull, "pull-secret-literal", "Create a secret from literal available for pull-requests (eq. --pull-secret-literal=secret_name=value)")
	flag.Parse()

	ui := console.New(os.Stdin, os.Stdout)

	uri, err := url.Parse(flag.Arg(0))
	if err != nil {
		flag.Usage()
		ui.Fatalf("An error occurred while parsing Drone URL: %v. Enter URL with protocol, for example: https://drone.bcn.magento.com.\n", err)
	}

	if uri.Host == "" {
		flag.Usage()
		ui.Fatalf("Drone hostname is empty. Enter URL with protocol, for example: https://drone.bcn.magento.com.\n")
	}

	org, repo := flag.Arg(1), flag.Arg(2)
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

	// create an http client with oauth authentication
	auth := new(oauth2.Config).Client(context.Background(), &oauth2.Token{AccessToken: pass})
	auth.Timeout = 30 * time.Second

	// create the drone client with authenticator
	dcli := drone.NewClient(uri.String(), auth)

	ui.Printf("Sync repository list in drone\n")
	if _, err := dcli.RepoListSync(); err != nil {
		ui.Fatalf("An error occurred: %v\n", err)
	}

	ui.Printf("Activating repository in drone\n")
	if _, err = dcli.RepoEnable(org, repo); err != nil && !strings.Contains(err.Error(), "Repository is already active") {
		ui.Fatalf("An error occurred: %v\n", err)
	}

	ImportSecrets(ui, dcli, org, repo)

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

func ImportSecrets(ui *console.Console, dcli drone.Client, org string, repo string) {
	if len(fileSecrets)+len(fileSecretsPull)+len(literalSecrets)+len(literalSecretsPull) == 0 {
		return
	}

	ui.Titlef("Importing repository secrets...\n")

	for _, secret := range literalSecretsPull {
		key, value := SplitKeyValue(secret)

		ui.Printf("Adding secret %#v from literal...\n", key)
		if err := CreateOrUpdateSecret(dcli, org, repo, key, value, true); err != nil {
			ui.Errorf("An error occurred while adding secret: %v\n", err)
		}
	}

	for _, secret := range literalSecrets {
		key, value := SplitKeyValue(secret)

		ui.Printf("Adding secret %#v from literal...\n", key)
		if err := CreateOrUpdateSecret(dcli, org, repo, key, value, false); err != nil {
			ui.Errorf("An error occurred while adding secret: %v\n", err)
		}
	}

	for _, secret := range fileSecretsPull {
		key, file := SplitKeyValue(secret)
		value, err := ioutil.ReadFile(file)
		if err != nil {
			ui.Errorf("An error occurred while reading secret file %#v: %v\n", file, err)
		}

		ui.Printf("Adding secret %#v from file...\n", key)
		if err := CreateOrUpdateSecret(dcli, org, repo, key, string(value), true); err != nil {
			ui.Errorf("An error occurred while adding secret: %v\n", err)
		}
	}

	for _, secret := range fileSecrets {
		key, file := SplitKeyValue(secret)
		value, err := ioutil.ReadFile(file)
		if err != nil {
			ui.Errorf("An error occurred while reading secret file %#v: %v\n", file, err)
		}

		ui.Printf("Adding secret %#v from file...\n", key)
		if err := CreateOrUpdateSecret(dcli, org, repo, key, string(value), false); err != nil {
			ui.Errorf("An error occurred while adding secret: %v\n", err)
		}
	}
}

// CreateOrUpdateSecret in drone
func CreateOrUpdateSecret(dcli drone.Client, owner, repo, key string, value string, pulls bool) error {
	secret, err := dcli.Secret(owner, repo, key)
	if err != nil && strings.Contains(err.Error(), "client error 404") {
		_, err = dcli.SecretCreate(owner, repo, &drone.Secret{
			Name:            key,
			Data:            value,
			PullRequest:     pulls,
			PullRequestPush: pulls,
		})

		return err
	}

	if err != nil {
		return err
	}

	secret.Name = key
	secret.Data = value
	secret.PullRequest = pulls
	secret.PullRequestPush = pulls

	_, err = dcli.SecretUpdate(owner, repo, secret)
	return err
}

// SplitKeyValue for string in format key=value
func SplitKeyValue(c string) (string, string) {
	if parts := strings.SplitN(c, "=", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}

	return c, ""
}

// AskCredentials (token) for drone
func AskCredentials(ui *console.Console, u *url.URL) (pass string) {
	for {
		ui.Printf("Follow this link to get your Personal Token: %v://%v/account.\n", u.Scheme, u.Host)

		pass = ui.ReadString("Enter your personal token: ")

		// build drone client
		auth := new(oauth2.Config).Client(context.Background(), &oauth2.Token{AccessToken: pass})
		auth.Timeout = 30 * time.Second

		// create the drone client with authenticator
		client := drone.NewClient(u.String(), auth)

		// gets the current user
		_, err := client.Self()
		if err == nil {
			return
		}

		ui.Errorf("An error occurred while validating your personal token: %v\n", err)
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
