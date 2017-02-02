package buildpack

import (
	"fmt"
	"io"
	"os"
)

// Buildpack Custom Error
type Error interface {
	error

	// Error safe to return to Users
	BuildpackError() string
}

type bperror struct {
	e   string
	bpe string
}

func (e *bperror) Error() string {
	return e.e
}

func (e *bperror) BuildpackError() string {
	return e.bpe
}

func newBuildpackError(bpe, format string, args ...interface{}) error {
	return &bperror{e: fmt.Sprintf(format, args...), bpe: bpe}
}

const defaultVersionsError = "The buildpack manifest is misconfigured for 'default_versions'. " +
	"Contact your Cloud Foundry operator/admin. For more information, see " +
	"https://docs.cloudfoundry.org/buildpacks/custom.html#specifying-default-versions"

type Logger struct {
	w io.Writer
}

func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Fprintf(l.w, "       ")
	fmt.Fprintf(l.w, format, args...)
	fmt.Fprintf(l.w, "\n")
}
func (l *Logger) Warning(format string, args ...interface{}) {
	fmt.Fprintf(l.w, "       **WARNING** ")
	fmt.Fprintf(l.w, format, args...)
	fmt.Fprintf(l.w, "\n")
}
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(l.w, "       **ERROR** ")
	fmt.Fprintf(l.w, format, args...)
	fmt.Fprintf(l.w, "\n")
}
func (l *Logger) BeginStep(format string, args ...interface{}) {
	fmt.Fprintf(l.w, "----->  ")
	fmt.Fprintf(l.w, format, args...)
	fmt.Fprintf(l.w, "\n")
}
func (l *Logger) SetOutput(w io.Writer) {
	l.w = w
}

var Log = &Logger{w: os.Stdout}
