/*
Copyright 2019 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package console

import (
	"bufio"
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"os"
	"strings"
)

type Console struct {
	r io.Reader
	w io.Writer
}

func New(r io.Reader, w io.Writer) *Console {
	return &Console{r: r, w: w}
}

func (l *Console) ReadString(prompt string) string {
	l.Printf(prompt)

	input, _ := bufio.NewReader(l.r).ReadString('\n')
	return strings.TrimSuffix(input, "\n")
}

func (l *Console) Debugf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.w, aurora.Gray(0, format).String(), args...)
}

func (l *Console) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.w, format, args...)
}

func (l *Console) Titlef(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.w, aurora.Bold(aurora.Blue(format)).String(), args...)
}

func (l *Console) Successf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.w, aurora.BrightGreen(format).String(), args...)
}

func (l *Console) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.w, aurora.Red(format).String(), args...)
}

func (l *Console) Fatalf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(l.w, aurora.Red(format).String(), args...)
	os.Exit(1)
}
