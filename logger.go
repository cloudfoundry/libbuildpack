package libbuildpack

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
	l.printWithHeader("       **WARNING** ", format, args...)

}
func (l *Logger) Error(format string, args ...interface{}) {
	l.printWithHeader("       **ERROR** ", format, args...)
}

func (l *Logger) BeginStep(format string, args ...interface{}) {
	l.printWithHeader("-----> ", format, args...)
}

func (l *Logger) printWithHeader(header string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	msg = strings.Replace(msg, "\n", "\n       ", -1)
	fmt.Fprintf(l.w, "%s%s\n", header, msg)
}

func (l *Logger) SetOutput(w io.Writer) {
	l.w = w
}

var Log = &Logger{w: os.Stdout}
