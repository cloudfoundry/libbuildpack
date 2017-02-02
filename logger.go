package buildpack

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Logger struct {
	w io.Writer
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.printWithHeader("       ", format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.printWithHeader("       **WARNING**", format, args...)

}
func (l *Logger) Error(format string, args ...interface{}) {
	l.printWithHeader("       **ERROR**", format, args...)
}

func (l *Logger) BeginStep(format string, args ...interface{}) {
	l.printWithHeader("----->", format, args...)
}

func (l *Logger) printWithHeader(header string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	lines := strings.Split(msg, "\n")

	fmt.Fprintf(l.w, "%s %s\n", header, lines[0])

	for i := 1; i < len(lines); i++ {
		fmt.Fprintf(l.w, "       %s\n", lines[i])
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.w = w
}

var Log = &Logger{w: os.Stdout}
