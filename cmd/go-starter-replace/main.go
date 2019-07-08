package main

import (
	"bytes"
	"flag"
	"github.com/magento-mcom/go-starter/pkg/console"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var skips = []string{".starter/", ".starter.yml", ".git/"}

func main() {
	var prefix, suffix string

	flag.StringVar(&prefix, "prefix", "<", "Placeholder prefix")
	flag.StringVar(&suffix, "suffix", ">", "Placeholder suffix")
	flag.Parse()

	ui := console.New(os.Stdin, os.Stdout)

	vars := variables()

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

		dir := filepath.Dir(path)
		name := file.Name()

		if renamed := rename(name, prefix, suffix, vars); renamed != name {
			ui.Printf("Renaming %#v to %#v\n", path, filepath.Join(dir, renamed))
			if err := os.Rename(path, filepath.Join(dir, renamed)); err != nil {
				return err
			}

			name = renamed
			path = filepath.Join(dir, name)
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

	if err = ioutil.WriteFile(filename, output, 0666); err != nil {
		return false, err
	}

	return !bytes.Equal(input, output), nil
}
