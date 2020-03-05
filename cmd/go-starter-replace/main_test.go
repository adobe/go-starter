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
	"fmt"
	"github.com/adobe/go-starter/pkg/console"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Test happy path: replacing placeholders with values
func TestReplace(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	create(t, ".starter.yml", "")
	create(t, "file.txt", "foo <PLACEHOLDER> bar")
	create(t, "nested/file.txt", "foo <PLACEHOLDER> bar")
	create(t, "nested/<PLACEHOLDER>/file.txt", "foo <PLACEHOLDER> bar")

	ui := console.New(bytes.NewBuffer(nil), ioutil.Discard)

	replace(ui, map[string]string{
		"PLACEHOLDER": "REPLACED",
	})

	assert(t, "file.txt", "foo REPLACED bar")
	assert(t, "nested/file.txt", "foo REPLACED bar")
	assert(t, "nested/REPLACED/file.txt", "foo REPLACED bar")
}

// Test reverse replace: values back to placeholders
func TestReplaceReverse(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	create(t, ".starter.yml", "")
	create(t, "file.txt", "foo REPLACED bar")
	create(t, "nested/file.txt", "foo REPLACED bar")
	create(t, "nested/REPLACED/file.txt", "foo REPLACED bar")

	reverse = true

	ui := console.New(bytes.NewBuffer(nil), ioutil.Discard)

	replace(ui, map[string]string{
		"PLACEHOLDER": "REPLACED",
	})

	assert(t, "file.txt", "foo <PLACEHOLDER> bar")
	assert(t, "nested/file.txt", "foo <PLACEHOLDER> bar")
	assert(t, "nested/<PLACEHOLDER>/file.txt", "foo <PLACEHOLDER> bar")
}

// Make sure go-starter-replace won't run in non-template directory
func TestReplaceChecksStarterConfig(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	create(t, "file.txt", "foo <PLACEHOLDER> bar")

	out := bytes.NewBuffer(nil)

	ui := console.New(bytes.NewBuffer(nil), out)

	replace(ui, map[string]string{
		"PLACEHOLDER": "REPLACED",
	})

	assert(t, "file.txt", "foo <PLACEHOLDER> bar")

	want, got := ".starter.yml does not exist", out.String()
	if !strings.Contains(got, want) {
		t.Errorf("Error message should be printed: want %v, got %v", want, got)
	}
}

// setup test, create temporary path and chdir into it
func setup(t *testing.T) func() {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Unable to get working dir:", err)
	}

	path := filepath.Join(os.TempDir(), fmt.Sprint(rand.Int()%10000))

	if err := os.Mkdir(path, 0777); err != nil {
		t.Fatalf("Unable to create test workspace %v: %v", path, err)
	}

	if err := os.Chdir(path); err != nil {
		t.Fatalf("Unable to chdir to test workspace: %v", err)
	}

	return func() {
		// chdir back to cwd
		_ = os.Chdir(cwd)

		// reset flags
		prefix, suffix, reverse = "<", ">", false

		// remove test workspace
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Unable to remove test workspace: %v", err)
		}
	}
}

// create test file at given path
func create(t *testing.T, filename, content string) {
	path := filepath.Dir(filename)
	if path != "" {
		if err := os.MkdirAll(path, 0777); err != nil {
			t.Fatalf("Unable to create test path %v: %v", path, err)
		}
	}

	if err := ioutil.WriteFile(filename, []byte(content), 0666); err != nil {
		t.Fatalf("Unable to create test file %v: %v", filename, err)
	}
}

// assert file content matching
func assert(t *testing.T, filename, want string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Unable to read test file %v: %v", filename, err)
		return
	}

	got := string(data)

	if want != got {
		t.Errorf("Test file %#v: content does not match, want: %v, got %v", filename, want, got)
	}
}
