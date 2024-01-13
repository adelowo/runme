package command

import (
	"context"
	"io"
	"os/exec"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type localCommand struct {
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

	Logger *zap.Logger

	// cmd is populated when the command is started.
	cmd *exec.Cmd
}

func (c *localCommand) Start(ctx context.Context) error {
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

	// Set the process group ID of the program.
	// It is helpful to stop the program and its
	// children. See command.Stop().
	// Note that Setsid set in setSysProcAttrCtty()
	// already starts a new process group, hence,
	// this call is inside this branch.
	// TODO(adamb): it does not work with interactive programs like "python".
	// setSysProcAttrPgid(c.cmd)

	c.Logger.Info("starting a local command", zap.Any("command", c))

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	c.Logger.Info("a local command started")

	return nil
}

func (c *localCommand) Wait() error {
	c.Logger.Info("waiting for the local command to finish")
	err := c.cmd.Wait()
	c.Logger.Info("the local command finished", zap.Error(err))
	return errors.WithStack(err)
}
