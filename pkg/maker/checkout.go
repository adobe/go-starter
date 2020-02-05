/*
Copyright 2019 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

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
