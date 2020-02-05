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
