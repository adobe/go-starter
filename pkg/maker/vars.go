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
