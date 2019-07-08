# Go Starter

Go-starter allows to bootstrap a new project from a template. It uses Git repositories as templates and is shipped with batch of utilities to make bootstarpping easier.

## Installation

Use [homebrew tap](https://github.com/magento-mcom/homebrew-tap) to install go-starter. First, follow instructions on how to add magento-mcom/tap [from here](https://github.com/magento-mcom/homebrew-tap/blob/master/README.md#setup). Then, use brew to install go-starter:

```
brew update
brew install go-starter
```

Check if go-starter is properly installed by running it:

```
go-starter 
```

## Usage

Run go-starter with template repository URL and path where you would like to create a new project, for example:

```bash
go-starter https://github.com/magento-mcom/go-scaffolding awesome-project
```

You can specify full GitHub URL or just repository name (like so `go-scaffolding`). Go-starter automatically assumes template is published on GitHub under magento-mcom organisation, so you can ommit these parts.

Now, go-starter will clone [go-scaffolding](https://github.com/magento-mcom/go-scaffolding) into `./awesome-project` directory and run tasks defined in [`.starter.yml`](https://github.com/magento-mcom/go-scaffolding/blob/master/.starter.yml). See "Templates" for more details.

### Advanced usage

You can skip cloning, for example if template is already cloned, but task failed to execute by passing `-skip-clone` flag. 

You can also pass additional variables (or pre-define variables instead of entering them using prompt) using `-var` flag.

## Templates

Templates are regular Git repositories like this one. If you try to use go-starter with random repository it will just clone it to your computer. To make use of go-starter you would need to let it know how to "post-process" template repository after it has been cloned. To do so, you need to define `.starter.yml` configuration file. 

After go-starter clones repository it tries to load `.starter.yml` to execute additional actions and turn template into a project. This configuration file defines two sections:

- `questions` - list of questions to be asked from user, for example: project name, binary name, team etc
- `tasks` - list of commands to be executed (these can be globally installed binaries, or binaries packed with template itself)

Here is an example of `.starter.yml`:

```yaml
questions:
  - message: Application name
    name: "application_name"
    type: input
    regexp: ^[a-z0-9\-]+$
    help_msg: Must be lowercase characters or digits or hyphens
  - message: GitHub owners
    name: "github_owners"
    type: input
    regexp: ^[a-z0-9\-]+$
    help_msg: Must be lowercase characters or digits or hyphens

tasks:
  - command: [ "go-starter-replace" ]
  - command: [ "rm", ".starter.yml" ]
```

This file defines two questions, asking user to enter `application_name` and `github_owners` variables. Then, go-starter will execute `go-starter-replace` binary (shipped with go-starter) to replace placeholders in the files of the template, turning generic tempalte into something more specific to the project. Finally it will use standard `rm ` command to remove `.starter.yml`.

Each template may contain custom tasks placed in `.starter` folder. For example, you can create a bash script which would generate CODEOWNERS file and place it under `.starter/make-owners`. Then, add it as tasks in `.starter.yml` like so:

```yaml
...

tasks:
...
  - command: [ "./.starter/make-owners" ]
```

Custom scripts may access variables (answers to the questions) through environment variables. They are uppercased and prefixed with `STARTER_`. Following example above, `./.starter/make-owners` may get `github_owners` variable using `STARTER_GITHUB_OWNERS` environment variable. 

## Build-in tasks

Go-starter ships with few additional binaries which can be used as tasks in `.starter.yml`.

### go-starter-replace

This binary recursively goes through files in current folder and replaces placeholders to variable values in files and their names. By default, placeholders are surrounded by `<` and `>`.

### go-starter-github

This binary automatically created GitHub repository, initiates local Git repository, adds GitHub remote and pushes changes to GitHub.

### go-starter-drone

This binary configures drone integration and runs first build.

