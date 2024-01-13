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
	Cfg  *Config
	Opts *VirtualCommandOptions

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	// stdin is Opts.Stdin wrapped in ReadCloser.
	stdin io.ReadCloser

	pty *os.File
	tty *os.File

	wg sync.WaitGroup

	mx  sync.Mutex
	err error

	logger *zap.Logger
}

// readClose allows to wrap a io.Reader into io.ReadCloser.
//
// When Close() is called, the underlying read operation is ignored.
// A disadvantage is that it may leak and hang indefinitely, or
// the read data is lost. It's caller's responsibility to interrupt
// the underlying reader when the virtualCommand exits.
type readCloser struct {
	r    io.Reader
	done chan struct{}
}

func (r *readCloser) Read(p []byte) (n int, err error) {
	readc := make(chan struct{})

	go func() {
		n, err = r.r.Read(p)
		close(readc)
	}()

	select {
	case <-readc:
		return n, err
	case <-r.done:
		return 0, io.EOF
	}
}

func (r *readCloser) Close() error {
	close(r.done)
	return nil
}

func newVirtualCommand(cfg *Config, opts *VirtualCommandOptions) *virtualCommand {
	var stdin io.ReadCloser

	if opts.Stdin != nil {
		stdin = &readCloser{r: opts.Stdin, done: make(chan struct{})}
	}

	return &virtualCommand{
		Cfg:    cfg,
		Opts:   opts,
		stdin:  stdin,
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
		c.Cfg.ProgramPath,
		c.Cfg.Args...,
	)
	c.cmd.Dir = c.Cfg.Dir
	c.cmd.Env = c.Opts.Env
	c.cmd.Stdin = c.tty
	c.cmd.Stdout = c.tty
	c.cmd.Stderr = c.tty

	setSysProcAttrCtty(c.cmd)

	if !isNil(c.stdin) {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			n, err := io.Copy(c.pty, c.stdin)
			c.logger.Info("copy from stdin to pty", zap.Error(err), zap.Int64("count", n))
			if err != nil {
				c.setErr(errors.WithStack(err))
			}
		}()
	}

	if !isNil(c.Opts.Stdout) {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			n, err := io.Copy(c.Opts.Stdout, c.pty)
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

				c.setErr(errors.WithStack(err))
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
	c.logger.Info("stopping the virtual command with signal", zap.String("signal", sig.String()))

	if c.pty != nil {
		c.logger.Info("closing pty due to the signal")

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
	c.setErr(err)
	c.logger.Info("the remote command finished", zap.Error(err))

	_ = c.closeIOLoops()

	c.wg.Wait()

	c.mx.Lock()
	err = c.err
	c.mx.Unlock()

	return errors.WithStack(err)
}

func (c *virtualCommand) closeIOLoops() (err error) {
	if !isNil(c.stdin) {
		err = c.stdin.Close()
	}
	return
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
