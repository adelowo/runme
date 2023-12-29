package command

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"github.com/pkg/errors"
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

	Stdin  io.ReadCloser
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	// pty and tty are pseudo-terminal primary and secondary.
	// They may be nil if the command is not interactive.
	pty          *os.File
	ptyWriteDone chan struct{}
	ptyReadDone  chan struct{}
	tty          *os.File

	mu  sync.Mutex
	err error
}

func (c *Command) Start(ctx context.Context) error {
	c.cmd = exec.CommandContext(
		ctx,
		c.Path,
		c.Args...,
	)
	c.cmd.Dir = c.Dir
	c.cmd.Env = c.PreEnv

	if c.tty != nil {
		c.cmd.Stdin = c.tty
		c.cmd.Stdout = c.tty
		c.cmd.Stderr = c.tty

		setSysProcAttrCtty(c.cmd)
	} else {
		c.cmd.Stdin = c.Stdin
		c.cmd.Stdout = c.Stdout
		c.cmd.Stderr = c.Stderr

		// Set the process group ID of the program.
		// It is helpful to stop the program and its
		// children. See command.Stop().
		// Note that Setsid set in setSysProcAttrCtty()
		// already starts a new process group, hence,
		// this call is inside this branch.
		setSysProcAttrPgid(c.cmd)
	}

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	if c.tty != nil {
		// Close tty as not needed anymore.
		if err := c.tty.Close(); err != nil {
			c.Logger.Info("failed to close tty after starting the command", zap.Error(err))
		}

		c.tty = nil
	}

	if c.pty != nil {
		go func() {
			defer close(c.ptyWriteDone)

			for {
				// It's possible to set read deadline because the stdin
				// is in the non-blocking mode. This is set in the builder.
				// if err := c.Stdin.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				// 	c.Logger.Info("failed to set read deadline", zap.Error(err))
				// }

				// TODO(adamb): this generally works, but it still may leak some input data
				// that should be passed to the next command.
				n, err := io.Copy(c.pty, c.Stdin)
				if err != nil {
					c.Logger.Info("failed to copy from stdin to pty", zap.Error(err))

					if errors.Is(err, os.ErrDeadlineExceeded) && c.pty != nil {
						continue
					}

					if errors.Is(err, os.ErrClosed) {
						break
					}

					c.setErr(err)
				} else {
					c.Logger.Debug("finished copying from stdin to pty", zap.Int64("count", n))
				}

				break
			}
		}()

		go func() {
			defer close(c.ptyReadDone)
			n, err := io.Copy(c.Stdout, c.pty)
			if err != nil {
				// Linux kernel returns EIO when attempting to read from
				// a master pseudo-terminal which no longer has an open slave.
				// See https://github.com/creack/pty/issues/21.
				if errors.Is(err, syscall.EIO) {
					c.Logger.Debug("failed to copy from pty to stdout; handled EIO")
					return
				}
				if errors.Is(err, os.ErrClosed) {
					c.Logger.Debug("failed to copy from pty to stdout; handled ErrClosed")
					return
				}

				c.Logger.Info("failed to copy from pty to stdout", zap.Error(err))

				c.setErr(err)
			} else {
				c.Logger.Debug("finished copying from pty to stdout", zap.Int64("count", n))
			}
		}()
	}

	return nil
}

func (c *Command) Wait() error {
	err := c.cmd.Wait()
	if err != nil {
		var pErr *exec.ExitError
		if errors.As(err, &pErr) {
			c.Logger.Info("command has finished", zap.ByteString("stderr", pErr.Stderr), zap.Error(err))
		} else {
			c.Logger.Info("command has finished", zap.Error(err))
		}
	}

	if c.Stdin != nil {
		err := c.Stdin.Close()
		c.Logger.Info("stdin exists and has been closed", zap.Error(err))
	}

	if c.tty != nil {
		err := c.tty.Close()
		c.Logger.Debug("tty exists and has been closed", zap.Error(err))
		c.tty = nil
	}

	if c.pty != nil {
		<-c.ptyWriteDone

		err := c.pty.Close()
		c.Logger.Debug("pty exists and has been closed", zap.Error(err))
		c.pty = nil

		<-c.ptyReadDone
	}

	c.Logger.Info("all commands I/O has been closed")

	return errors.WithStack(err)
}

func (c *Command) setTty() (err error) {
	c.pty, c.tty, err = pty.Open()
	err = errors.WithStack(err)
	if err == nil {
		c.ptyWriteDone = make(chan struct{})
		c.ptyReadDone = make(chan struct{})
	}
	return
}

func (c *Command) setErr(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	if c.err == nil {
		c.err = err
	}
	c.mu.Unlock()
}
