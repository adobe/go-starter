package main

import (
	"flag"
	"fmt"
	"github.com/magento-mcom/go-starter/pkg/console"
	"github.com/magento-mcom/go-starter/pkg/maker"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

var version, commit string

func usage() {
	out := flag.CommandLine.Output()

	_, _ = fmt.Fprintf(out, "go-starter version %v (commit %v)\n", version, commit)
	_, _ = fmt.Fprintf(out, "\n")
	_, _ = fmt.Fprintf(out, "Usage: %s [flags] <template> <destination>\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nExample:\n")
	_, _ = fmt.Fprintf(out, "    %s -var \"app_name=awesome-project\" skolodyazhnyy/go-starter-template awesome-project\n", os.Args[0])

	if flag.NFlag() != 0 {
		_, _ = fmt.Fprintf(out, "\nFlags:\n")
		flag.PrintDefaults()
	}

	_, _ = fmt.Fprintf(out, "\n")
}

func main() {
	var skipClone bool
	var template, destination, branch string
	var vars = make(maker.Vars)

	flag.Usage = usage
	flag.Var(&vars, "var", "An additional variable. Can be used multiple times. Example: -var \"variable_name=value\"")
	flag.BoolVar(&skipClone, "skip-clone", false, "Skip clone step, just enter destination directory and run tasks.")
	flag.StringVar(&branch, "branch", "master", "Branch to checkout in template repository.")
	flag.Parse()

	ui := console.New(os.Stdin, os.Stdout)

	template, destination = flag.Arg(0), flag.Arg(1)
	if template == "" {
		flag.Usage()
		ui.Fatalf("ERROR: template should not be empty, use Git repository URL\n")
	}

	if destination == "" {
		flag.Usage()
		ui.Fatalf("ERROR: destination should not be empty, enter folder where you want to deploy new application\n")
	}

	// Get clone URL
	cloneURL := maker.ResolveTemplateURL(template)

	// Add service variables
	vars["template_url"] = cloneURL
	vars["template_branch"] = branch
	vars["destination"] = destination

	// Clone template repository
	if !skipClone {
		ui.Titlef("Cloning template %v\n", template)

		if err := maker.Checkout(destination, cloneURL, branch); err != nil {
			ui.Fatalf("An error occurred: %v\n", err)
		}
	}

	// Enter destination folder so all next steps are executed in current work dir
	if err := os.Chdir(destination); err != nil {
		ui.Fatalf("An error occurred when chdir to destination directory: %v\n", err)
	}

	// Read config
	config, err := load(".starter.yml")
	if err != nil {
		ui.Fatalf("An error occurred when reading .starter.yml from repository: %v\n", err)
	}

	// Ask questions
	vars, err = maker.Ask(ui, config.Questions, vars)
	if err != nil {
		ui.Fatalf("An error occurred when reading user input", err)
	}

	// Run tasks
	for _, task := range config.Tasks {
		if len(task.Command) == 0 {
			ui.Fatalf("Task command can not be empty, check your .starter.yml\n")
		}

		name, args := task.Command[0], subst(task.Command[1:], vars)

		ui.Titlef("Running task %v...\n", name)

		if err := maker.Run(vars, name, args...); err != nil {
			ui.Fatalf("An error occurred when executing task: %v\n", err)
		}
	}

	ui.Successf("You're all set, happy coding!\n")
}

// load .starter.yml
func load(file string) (c maker.Config, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &c)
	return
}

// subst variables in list of strings
func subst(args []string, vars map[string]string) (out []string) {
	for _, arg := range args {
		value := arg
		for k, v := range vars {
			value = strings.ReplaceAll(value, "$"+k, v)
		}

		out = append(out, value)
	}

	return
}
