# Go Starter

Go-starter allows to bootstrap a new project from a template. It uses Git repositories as templates and is shipped with batch of utilities to make bootstarpping easier.

## Installation

Download latest release from [release](https://github.com/adobe/go-starter/releases) page using one of the commands below.

**Mac OS**

```
curl https://github.com/adobe/go-starter/releases/latest/download/go-starter-darwin-amd64.tgz \
  -sSfL -o go-starter.tgz
```

**Linux**

```
curl https://github.com/adobe/go-starter/releases/latest/download/go-starter-linux-amd64.tgz \
  -sSfL -o go-starter.tgz
```

Unpack content of the archive to a directory listed in `$PATH`. The archive includes multiple binaries shipped with `go-starter`.

```
tar -xvzf go-starter.tgz -C /usr/local/bin
rm go-starter.tgz
```

Run `go-starter` to verify it's installed correctly.

## Usage

Run go-starter with template repository URL and path where you would like to create a new project, for example:

```bash
go-starter starter-template/hello-world-starter awesome-project
```

You can specify full GitHub URL or just repository name (like so `starter-template/hello-world-starter`). 

Now, go-starter will clone [hello-world-starter](https://github.com/starter-template/hello-world-starter) into `./awesome-project` directory and run tasks defined in [`.starter.yml`](https://github.com/starter-template/hello-world-starter/blob/master/.starter.yml). See "Templates" for more details.

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

#### Usage

```bash
Usage: go-starter-replace [flags]

Example:
    STARTER_PLACEHOLDER1=VALUE1 STARTER_PLACEHOLDER2=VALUE2 go-starter-replace

Flags:
  -prefix string
        Placeholder prefix (default "<")
  -suffix string
        Placeholder suffix (default ">")
```

### go-starter-github

This binary automatically created GitHub repository, initiates local Git repository, adds GitHub remote and pushes changes to GitHub.

#### Usage

```bash
Usage: go-starter-github [flags] <github-org> <github-repo>

Example:
    go-starter-github adobe awesome-project

Flags:
  -branch string
        Name of the master branch (default "master")
  -collaborator value
        Add collaborators to the repository by GitHub username. You can grant permissions using following format: <username>:<permission>. Permission can be: pull (read only), push (read and write) or admin (everything), default is push. Can be specified multiple times. Example: --collaborator octocat:pull
  -deploy-key string
        Add SSH deployment key to the repository, add ':rw' suffix to grant write permissions to the key
  -public
        Make repository public
  -remote string
        Name of the remote in local repository (default "upstream")
  -with-issues
        Enable issues in GitHub
  -with-projects
        Enable projects in GitHub
  -with-wiki
        Enable wiki page in GitHub
```

### go-starter-drone

This binary configures drone integration and runs first build.

#### Usage

```bash
Usage: go-starter-drone [flags] <drone-url> <org-name> <repo-name>

Example:
    go-starter-drone https://cloud.drone.io adobe awesome-project

Flags:
  -pull-secret-file value
        Create a secret from file available for pull-requests (eq. --pull-secret-file=secret_name=./path/to/file)
  -pull-secret-literal value
        Create a secret from literal available for pull-requests (eq. --pull-secret-literal=secret_name=value)
  -secret-file value
        Create a secret from file (eq. --secret-file=secret_name=./path/to/file)
  -secret-literal value
        Create a secret from literal (eq. --secret-literal=secret_name=value)
```
