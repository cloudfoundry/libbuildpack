package buildpack

import "fmt"

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
