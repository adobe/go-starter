package console

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"os"
)

type Console struct {
	r io.Reader
	w io.Writer
}

func New(r io.Reader, w io.Writer) *Console {
	return &Console{r: r, w: w}
}

func (l *Console) Scanln(arg interface{}) {
	_, _ = fmt.Fscanln(l.r, arg)
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
