package libbuildpack

import (
	"io"
	"os"
	"os/exec"
)

type Command interface {
	SetOutput(io.Writer)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	Run() error
}

type command struct {
	cmd    *exec.Cmd
	stdout io.Writer
	stderr io.Writer
}

func NewCommand(program string, args ...string) Command {
	c := &command{
		cmd: exec.Command(program, args...),
	}

	c.SetOutput(os.Stdout)
	return c
}

func (c *command) SetOutput(output io.Writer) {
	c.SetStderr(output)
	c.SetStdout(output)
}

func (c *command) SetStderr(output io.Writer) {
	c.stderr = output
	c.cmd.Stderr = output
}
func (c *command) SetStdout(output io.Writer) {
	c.stdout = output
	c.cmd.Stdout = output
}

func (c *command) Run() error {
	return c.cmd.Run()
}
