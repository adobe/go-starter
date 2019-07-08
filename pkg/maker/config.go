package maker

import (
	"strings"
)

type Config struct {
	Questions []Question
	Tasks     []Task
}

type Question struct {
	Message           string `yaml:"message"`
	Name              string `yaml:"name"`
	Type              string `yaml:"type"`
	Default           string `yaml:"default"`
	RegExp            string `yaml:"regexp"`
	ValidationMessage string `yaml:"validation_msg"`
	HelpMessage       string `yaml:"help_msg"`
}

type Task struct {
	Command StringOrSlice
}

type StringOrSlice []string

func (ss *StringOrSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// try to parse task as string
	var inline string
	if err := unmarshal(&inline); err == nil {
		*ss = strings.Split(inline, " ")
		return nil
	}

	// to to parse task as slice
	var slice []string
	err := unmarshal(&slice)
	*ss = slice

	return err
}
