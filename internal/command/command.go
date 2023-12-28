package command

import (
	"context"
	"io"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
	"go.uber.org/zap"
)

type Command struct {
	// Name comes from the block. It is only used for
	// human-friendly messages and errors.
	Name string

	// Path is the path to the command.
	Path string

	// Args are the arguments to the command.
	Args []string

	PreEnv  []string
	PostEnv []string

	Dir string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	// pty and tty are pseudo-terminal primary and secondary.
	// They may be nil if the command is not interactive.
	// pty *os.File
	// tty *os.File

	// wg  sync.WaitGroup
	// mu  sync.Mutex
	// err error

	// logger *zap.Logger
}

type CommandOptions struct {
	ParentDir string
	Env       []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Logger    *zap.Logger
}

func CommandFromCodeBlock(
	block *document.CodeBlock,
	options *CommandOptions,
) (*Command, error) {
	if options == nil {
		options = &CommandOptions{}
	}

	builder, err := newCommandBuilder(block, options)
	if err != nil {
		return nil, err
	}

	return builder.Build()
}

func (c *Command) Start(ctx context.Context) error {
	c.cmd = exec.CommandContext(
		ctx,
		c.Path,
		c.Args...,
	)
	c.cmd.Dir = c.Dir
	c.cmd.Env = c.PreEnv
	c.cmd.Stdin = c.Stdin
	c.cmd.Stdout = c.Stdout
	c.cmd.Stderr = c.Stderr

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *Command) Wait() error {
	return errors.WithStack(c.cmd.Wait())
}
