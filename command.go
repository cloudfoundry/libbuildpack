package libbuildpack

import (
	"io"
	"os"
	"os/exec"
)

type Command interface {
	SetOutput(io.Writer)
	Run() error
}

type command struct {
	cmd *exec.Cmd
	out io.Writer
}

func NewCommand(program string, args ...string) Command {
	c := &command{
		cmd: exec.Command(program, args...),
	}

	c.SetOutput(os.Stdout)
	return c
}

func (c *command) SetOutput(output io.Writer) {
	c.out = output
	c.cmd.Stdout = output
	c.cmd.Stderr = output
}

func (c *command) Run() error {
	return c.cmd.Run()
}
