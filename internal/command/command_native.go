package command

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

type NativeCommand struct {
	cfg  *Config
	opts *NativeCommandOptions

	// cmd is populated when the command is started.
	cmd *exec.Cmd

	// tempFile is a temporary file created for the file mode execution.
	tempFile *os.File

	logger *zap.Logger
}

func newNativeCommand(cfg *Config, opts *NativeCommandOptions) *NativeCommand {
	return &NativeCommand{
		cfg:    cfg,
		opts:   opts,
		logger: opts.Logger,
	}
}

func (c *NativeCommand) Start(ctx context.Context) (err error) {
	switch c.cfg.Mode {
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_UNSPECIFIED.Enum():
		fallthrough
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_INLINE.Enum():
		// no additional setup
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_FILE.Enum():
		c.tempFile, err = createTempFileFromScript(c.cfg)

		defer func() {
			if err != nil {
				c.cleanup()
			}
		}()

		if err != nil {
			return
		}
	}

	stdin := c.opts.Stdin

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

	args := append([]string{}, c.cfg.Arguments...)

	// TODO(adamb): it's not always true that the script-based program
	// takes the filename as a last argument.
	if c.tempFile != nil {
		args = append(args, c.tempFile.Name())
	}

	c.cmd = exec.CommandContext(
		ctx,
		c.cfg.ProgramName,
		args...,
	)
	c.cmd.Dir = c.cfg.Directory
	// TODO(adamb): verify if it's ok to use local env.
	c.cmd.Env = append(os.Environ(), envFromConfigAndOptions(c.cfg, c.opts)...)
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

	c.logger.Info("starting a local command", zap.String("program", c.cfg.ProgramName), zap.Strings("args", args))

	if err := c.cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	c.logger.Info("a local command started")

	return nil
}

func (c *NativeCommand) StopWithSignal(sig os.Signal) error {
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

func (c *NativeCommand) Wait() error {
	c.logger.Info("waiting for the local command to finish")

	defer c.cleanup()

	var stderr []byte

	err := c.cmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stderr = exitErr.Stderr
		}
	}

	c.logger.Info("the local command finished", zap.Error(err), zap.ByteString("stderr", stderr))

	return errors.WithStack(err)
}

func (c *NativeCommand) cleanup() {
	if c.tempFile == nil {
		return
	}
	if err := os.Remove(c.tempFile.Name()); err != nil {
		c.logger.Info("failed to remove temporary file", zap.Error(err))
	}
}
