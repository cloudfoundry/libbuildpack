package libbuildpack

import (
	"io"
	"os"
	"os/exec"
)

type CommandRunner interface {
	SetOutput(io.Writer)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	SetDir(string)
	Reset()
	Run(program string, args ...string) error
}

type commandRunner struct {
	dir    string
	stdout io.Writer
	stderr io.Writer
}

func NewCommandRunner() CommandRunner {
	return &commandRunner{
		dir:    "",
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (c *commandRunner) SetOutput(output io.Writer) {
	c.SetStderr(output)
	c.SetStdout(output)
}

func (c *commandRunner) SetStderr(output io.Writer) {
	c.stderr = output
}
func (c *commandRunner) SetStdout(output io.Writer) {
	c.stdout = output
}
func (c *commandRunner) SetDir(dir string) {
	c.dir = dir
}

func (c *commandRunner) Run(program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	cmd.Dir = c.dir

	return cmd.Run()
}

func (c *commandRunner) Reset() {
	c.dir = ""
	c.stdout = os.Stdout
	c.stderr = os.Stderr
}
