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
	"regexp"
)

type console interface {
	Titlef(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	ReadString(string) string
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

			answer = ui.ReadString(fmt.Sprintf("Enter %v: ", q.Name))
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
