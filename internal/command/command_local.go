package command

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type localCommand struct {
	cfg  *Config
	opts *NativeCommandOptions

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	logger *zap.Logger
}

func newNativeCommand(cfg *Config, opts *NativeCommandOptions) *localCommand {
	return &localCommand{
		cfg:    cfg,
		opts:   opts,
		logger: opts.Logger.With(zap.String("name", cfg.Name)),
	}
}

func (c *localCommand) Start(ctx context.Context) error {
	stdin := c.opts.Stdin
	// TODO(adamb): it does not make sense really; just pass stdin as nil.
	if !c.cfg.Interactive {
		stdin = nil
	}

	if f, ok := stdin.(*os.File); ok && f != nil {
		// Duplicate /dev/stdin.
		newStdinFd, err := syscall.Dup(int(f.Fd()))
		if err != nil {
			return errors.Wrap(err, "failed to dup stdin")
		}
		syscall.CloseOnExec(newStdinFd)

		// Setting stdin to the non-block mode fails on the simple "read" command.
		// On the other hand, it allows to use SetReadDeadline().
		// It turned out it's not needed, but keeping the code here for now.
		// if err := syscall.SetNonblock(newStdinFd, true); err != nil {
		// 	return nil, errors.Wrap(err, "failed to set new stdin fd in non-blocking mode")
		// }

		stdin = os.NewFile(uintptr(newStdinFd), "")
	}

	c.cmd = exec.CommandContext(
		ctx,
		c.cfg.ProgramPath,
		c.cfg.Args...,
	)
	c.cmd.Dir = c.cfg.Dir
	c.cmd.Env = c.opts.Env
	c.cmd.Stdin = stdin
	c.cmd.Stdout = c.opts.Stdout
	c.cmd.Stderr = c.opts.Stderr

	// Set the process group ID of the program.
	// It is helpful to stop the program and its
	// children.
	// Note that Setsid set in setSysProcAttrCtty()
	// already starts a new process group.
	// Warning: it does not work with interactive programs
	// like "python", hence, it's commented out.
	// setSysProcAttrPgid(c.cmd)

	c.logger.Info("starting a local command", zap.Any("command", c))

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	c.logger.Info("a local command started")

	return nil
}

func (c *localCommand) Wait() error {
	c.logger.Info("waiting for the local command to finish")
	err := c.cmd.Wait()
	c.logger.Info("the local command finished", zap.Error(err))
	return errors.WithStack(err)
}
