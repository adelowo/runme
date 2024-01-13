package command

import (
	"context"
	"io"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type virtualCommand struct {
	cfg  *Config
	opts *VirtualCommandOptions

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	pty *os.File
	tty *os.File

	wg sync.WaitGroup

	mx  sync.Mutex
	err error

	logger *zap.Logger
}

func newVirtualCommand(cfg *Config, opts *VirtualCommandOptions) *virtualCommand {
	return &virtualCommand{
		cfg:    cfg,
		opts:   opts,
		logger: opts.Logger.With(zap.String("name", cfg.Name)),
	}
}

func (c *virtualCommand) IsRunning() bool {
	return c.cmd != nil && c.cmd.ProcessState == nil
}

func (c *virtualCommand) PID() int {
	if c.cmd == nil || c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

func (c *virtualCommand) Start(ctx context.Context) error {
	var err error

	c.pty, c.tty, err = pty.Open()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := disableEcho(c.tty.Fd()); err != nil {
		return err
	}

	c.cmd = exec.CommandContext(
		ctx,
		c.cfg.ProgramPath,
		c.cfg.Args...,
	)
	c.cmd.Dir = c.cfg.Dir
	c.cmd.Env = c.opts.Env
	c.cmd.Stdin = c.tty
	c.cmd.Stdout = c.tty
	c.cmd.Stderr = c.tty

	setSysProcAttrCtty(c.cmd)

	if !isNil(c.opts.Stdin) {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			n, err := io.Copy(c.pty, c.opts.Stdin)
			if err != nil {
				c.logger.Info("failed to copy from stdin to pty", zap.Error(err))
				c.setErr(err)
			} else {
				c.logger.Debug("finished copying from stdin to pty", zap.Int64("count", n))
			}
		}()
	}

	if !isNil(c.opts.Stdout) {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			n, err := io.Copy(c.opts.Stdout, c.pty)
			if err != nil {
				// Linux kernel returns EIO when attempting to read from
				// a master pseudo-terminal which no longer has an open slave.
				// See https://github.com/creack/pty/issues/21.
				if errors.Is(err, syscall.EIO) {
					c.logger.Debug("failed to copy from pty to stdout; handled EIO")
					return
				}
				if errors.Is(err, os.ErrClosed) {
					c.logger.Debug("failed to copy from pty to stdout; handled ErrClosed")
					return
				}

				c.logger.Info("failed to copy from pty to stdout", zap.Error(err))

				c.setErr(err)
			} else {
				c.logger.Debug("finished copying from pty to stdout", zap.Int64("count", n))
			}
		}()
	}

	c.logger.Info("starting a remote command", zap.Any("command", c))

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	c.logger.Info("a remote command started")

	return nil
}

func (c *virtualCommand) StopWithSignal(sig os.Signal) error {
	if c.pty != nil {
		if sig == os.Interrupt {
			_, _ = c.pty.Write([]byte{0x3})
		}

		if err := c.pty.Close(); err != nil {
			c.logger.Info("failed to close pty; continuing")
		}
	}

	// Try to terminate the whole process group. If it fails, fall back to stdlib methods.
	if err := signalPgid(c.cmd.Process.Pid, sig); err != nil {
		c.logger.Info("failed to terminate process group; trying Process.Signal()", zap.Error(err))
		if err := c.cmd.Process.Signal(sig); err != nil {
			c.logger.Info("failed to signal process; trying Process.Kill()", zap.Error(err))
			return errors.WithStack(c.cmd.Process.Kill())
		}
	}
	return nil
}

func (c *virtualCommand) Wait() error {
	c.logger.Info("waiting for the remote command to finish")
	err := c.cmd.Wait()
	c.logger.Info("the remote command finished", zap.Error(err))
	if err != nil {
		return errors.WithStack(err)
	}

	c.wg.Wait()

	c.mx.Lock()
	err = c.err
	c.mx.Unlock()

	return errors.WithStack(err)
}

func (c *virtualCommand) setErr(err error) {
	if err == nil {
		return
	}
	c.mx.Lock()
	if c.err == nil {
		c.err = err
	}
	c.mx.Unlock()
}

func isNil(val interface{}) bool {
	return val == nil || reflect.ValueOf(val).IsNil()
}
