package maker

import (
	"fmt"
	"regexp"
)

type console interface {
	Titlef(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Scanln(arg interface{})
}

// Ask questions from questions.yml
func Ask(ui console, questions []Question, vars map[string]string) (map[string]string, error) {
	var answer string

	answers := vars

	for _, q := range questions {
		// check if there is valid value in vars
		if v, ok := vars[q.Name]; ok {
			if ok, _ := Valid(q, v); ok {
				answers[q.Name] = v
				continue
			}

			ui.Errorf("Invalid input! %v\n", q.ValidationMessage)
		}

		// run loop to read value from stdin
		for {
			ui.Titlef(q.Message + "\n")

			if q.HelpMessage != "" {
				ui.Printf("Help: %v\n", q.HelpMessage)
			}

			if q.Default != "" {
				ui.Printf("Default: %v\n", q.Default)
			}

			ui.Printf("Enter %v: ", q.Name)
			ui.Scanln(&answer)

			if answer == "" {
				answer = q.Default
			}

			ok, err := Valid(q, answer)
			if err != nil {
				return nil, err
			}

			if ok {
				break
			}

			ui.Errorf("Invalid input! %v\n", q.ValidationMessage)
		}

		answers[q.Name] = answer
	}

	return answers, nil
}

func Valid(q Question, val string) (bool, error) {
	if q.RegExp == "" {
		return true, nil
	}

	re, err := regexp.Compile(q.RegExp)
	if err != nil {
		return true, fmt.Errorf("unable to parse regexp for variable %v, questions file is invalid", q.Name)
	}

	return re.MatchString(val), nil
}
