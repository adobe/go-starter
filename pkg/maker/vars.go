package maker

import (
	"strings"
)

type Vars map[string]string

func (i *Vars) String() string {
	return ""
}

func (i *Vars) Set(value string) error {
	if i == nil {
		*i = make(Vars)
	}

	parts := strings.SplitN(value, "=", 2)
	if len(parts) == 2 {
		(*i)[parts[0]] = parts[1]
	} else {
		(*i)[parts[0]] = "1"
	}

	return nil
}
