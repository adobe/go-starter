package maker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run a cli command
func Run(vars map[string]string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	for k, v := range vars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("STARTER_%v=%v", strings.ToUpper(k), v))
	}

	return cmd.Run()
}
