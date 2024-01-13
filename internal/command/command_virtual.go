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

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

type VirtualCommand struct {
	cfg  *Config
	opts *VirtualCommandOptions

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	// stdin is Opts.Stdin wrapped in readCloser.
	stdin io.ReadCloser

	tempFile *os.File

	pty *os.File
	tty *os.File

	wg sync.WaitGroup

	mx  sync.Mutex
	err error

	logger *zap.Logger
}

// readCloser allows to wrap a io.Reader into io.ReadCloser.
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

func newVirtualCommand(cfg *Config, opts *VirtualCommandOptions) *VirtualCommand {
	var stdin io.ReadCloser

	if opts.Stdin != nil {
		stdin = &readCloser{r: opts.Stdin, done: make(chan struct{})}
	}

	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}

	return &VirtualCommand{
		cfg:    cfg,
		opts:   opts,
		stdin:  stdin,
		logger: opts.Logger,
	}
}

func (c *VirtualCommand) IsRunning() bool {
	return c.cmd != nil && c.cmd.ProcessState == nil
}

func (c *VirtualCommand) PID() int {
	if c.cmd == nil || c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

func (c *VirtualCommand) SetWinsize(rows, cols, x, y uint16) (err error) {
	if c.pty == nil {
		return
	}

	err = pty.Setsize(c.pty, &pty.Winsize{
		Rows: rows,
		Cols: cols,
		X:    x,
		Y:    y,
	})
	return errors.WithStack(err)
}

func (c *VirtualCommand) Start(ctx context.Context) (err error) {
	switch c.cfg.Mode {
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_UNSPECIFIED.Enum():
		fallthrough
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_INLINE.Enum():
		// no additional setup
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_FILE.Enum():
		f, err := os.CreateTemp("", "runme-script-*")
		if err != nil {
			return errors.WithMessage(err, "failed to create a temporary file for script execution")
		}
		c.tempFile = f

		if _, err := f.Write([]byte(c.cfg.GetScript())); err != nil {
			return errors.WithMessage(err, "failed to write the script to the temporary file")
		}

		_ = f.Close()

		defer func() {
			if err != nil {
				_ = c.cleanup()
			}
		}()
	}

	c.pty, c.tty, err = pty.Open()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := disableEcho(c.tty.Fd()); err != nil {
		return err
	}

	// TODO(adamb): this should not work this way...
	args := append([]string{}, c.cfg.Arguments...)
	if c.tempFile != nil {
		args = append(args, c.tempFile.Name())
	}

	c.cmd = exec.CommandContext(
		ctx,
		c.cfg.ProgramName,
		args...,
	)
	c.cmd.Dir = c.cfg.Directory
	// TODO(adamb): verify the order
	c.cmd.Env = append(append([]string{}, c.opts.Env...), c.cfg.Env...)
	c.cmd.Stdin = c.tty
	c.cmd.Stdout = c.tty
	c.cmd.Stderr = c.tty

	setSysProcAttrCtty(c.cmd)

	if !isNil(c.stdin) {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			n, err := io.Copy(c.pty, c.stdin)
			c.logger.Info("finished copying from stdin to pty", zap.Error(err), zap.Int64("count", n))
			if err != nil {
				c.setErr(errors.WithStack(err))
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

				c.setErr(errors.WithStack(err))
			} else {
				c.logger.Debug("finished copying from pty to stdout", zap.Int64("count", n))
			}
		}()
	}

	c.logger.Info("starting a virtual command", zap.String("program", c.cfg.ProgramName), zap.Strings("args", args))

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	c.logger.Info("a virtual command started")

	return nil
}

func (c *VirtualCommand) StopWithSignal(sig os.Signal) error {
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

func (c *VirtualCommand) Wait() error {
	c.logger.Info("waiting for the virtual command to finish")

	defer func() { _ = c.cleanup() }()

	waitErr := c.cmd.Wait()
	c.logger.Info("the virtual command finished", zap.Error(waitErr))

	_ = c.closeIOLoops()

	c.wg.Wait()

	if waitErr != nil {
		return errors.WithStack(waitErr)
	}

	c.mx.Lock()
	err := c.err
	c.mx.Unlock()

	return err
}

func (c *VirtualCommand) closeIOLoops() (err error) {
	if !isNil(c.stdin) {
		err = c.stdin.Close()
	}
	return
}

func (c *VirtualCommand) setErr(err error) {
	if err == nil {
		return
	}
	c.mx.Lock()
	if c.err == nil {
		c.err = err
	}
	c.mx.Unlock()
}

func (c *VirtualCommand) cleanup() error {
	if c.tempFile == nil {
		return nil
	}
	return os.Remove(c.tempFile.Name())
}

func isNil(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)

	if v.Type().Kind() == reflect.Struct {
		return false
	}

	return reflect.ValueOf(val).IsNil()
}
