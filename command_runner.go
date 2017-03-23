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
	ResetOutput()
	Run(program string, args ...string) error
}

type commandRunner struct {
	stdout io.Writer
	stderr io.Writer
}

func NewCommandRunner() CommandRunner {
	return &commandRunner{
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

func (c *commandRunner) Run(program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr

	return cmd.Run()
}

func (c *commandRunner) ResetOutput() {
	c.stdout = os.Stdout
	c.stderr = os.Stderr
}
