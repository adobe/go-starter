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
	"bytes"
	"flag"
	"fmt"
	"github.com/adobe/go-starter/pkg/console"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var version, commit string
var skips = []string{".starter/", ".starter.yml", ".git/"}

func usage() {
	out := flag.CommandLine.Output()
	_, _ = fmt.Fprintf(out, "go-starter-replace version %v (commit %v)\n", version, commit)
	_, _ = fmt.Fprintf(out, "\n")
	_, _ = fmt.Fprintf(out, "Usage: %s [flags]\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nExample:\n")
	_, _ = fmt.Fprintf(out, "    STARTER_PLACEHOLDER1=VALUE1 STARTER_PLACEHOLDER2=VALUE2 %s\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nFlags:\n")
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(out, "\n")
}

func main() {
	var prefix, suffix string

	flag.Usage = usage
	flag.StringVar(&prefix, "prefix", "<", "Placeholder prefix")
	flag.StringVar(&suffix, "suffix", ">", "Placeholder suffix")
	flag.Parse()

	ui := console.New(os.Stdin, os.Stdout)

	vars := variables()

	var renames []string

	// walk through current folder and replace variables
	err := filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
		if err != nil {
			ui.Errorf("Unable to process path %#v: %v\n", path, err)
			return nil
		}

		for _, skip := range skips {
			if strings.HasPrefix(path, skip) {
				return nil
			}
		}

		name := file.Name()

		if renamed := rename(name, prefix, suffix, vars); renamed != name {
			renames = append(renames, path)
		}

		if file.IsDir() {
			return nil
		}

		ok, err := replace(path, prefix, suffix, vars)
		if err != nil {
			return err
		}

		if ok {
			ui.Printf("Updating %#v\n", path)
		}

		return nil
	})

	for _, path := range renames {
		renamed := rename(path, prefix, suffix, vars)

		ui.Printf("Renaming %#v to %#v\n", path, renamed)
		if err := os.Rename(path, renamed); err != nil {
			ui.Errorf("Unable to rename path %#v: %v\n", path, err)
		}
	}

	if err != nil {
		ui.Fatalf("An error occurred while traversing file system: %v\n", err)
	}
}

// variables from environment
func variables() map[string]string {
	vars := make(map[string]string)

	for _, pair := range os.Environ() {
		if !strings.HasPrefix(pair, "STARTER_") {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)

		key := strings.TrimPrefix(parts[0], "STARTER_")
		value := "1"
		if len(parts) == 2 {
			value = parts[1]
		}

		vars[key] = value
	}

	return vars
}

// rename - replace placeholders in file name
func rename(filename, prefix, suffix string, params map[string]string) string {
	for k, v := range params {
		filename = strings.Replace(filename, prefix+k+suffix, v, -1)
	}

	return filename
}

// replace placeholders in file
func replace(filename, prefix, suffix string, params map[string]string) (bool, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}

	output := input

	for k, v := range params {
		output = bytes.Replace(output, []byte(prefix+k+suffix), []byte(v), -1)
	}

	if bytes.Equal(input, output) {
		return false, nil
	}

	if err = ioutil.WriteFile(filename, output, 0666); err != nil {
		return false, err
	}

	return true, nil
}
