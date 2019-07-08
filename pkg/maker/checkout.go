package maker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Checkout template repository
func Checkout(destination, template, branch string) error {
	if err := run("git", "clone", "--depth=1", "--branch="+branch, template, destination); err != nil {
		return fmt.Errorf("unable to clone template repository %#v into destination folder: %v", template, err)
	}

	if err := os.RemoveAll(filepath.Join(destination, ".git")); err != nil {
		return fmt.Errorf("unable to remote .git folder of template repository: %v", err)
	}

	return nil
}

// run a cli command
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
