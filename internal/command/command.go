package command

import (
	"context"
	"io"
	"os"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
	"go.uber.org/zap"
)

type Command interface {
	Start(context.Context) error
	Wait() error
}

// TODO(adamb): consider changing the strategy and have separate functions
// for commands executed locally: NewLocal() and from remote: NewRemote().
// For the local, it's very simple to reuse OS stdin, stderr, stdout.
// For the remote, a PTY will be allocated and data will flow accordingly.
// Sources:
// - https://www.baeldung.com/linux/pty-vs-tty#:~:text=Software%20terminal%2C%20i.e.%2C%20virtual%20TeleTYpe,actual%20or%20CLI%2Demulated%20GUI
// - For example, when you ssh in to a machine and run ls, the ls command is sending its output to a pseudo-terminal, the other side of which is attached to the SSH daemon.
type LocalOptions struct {
	ParentDir string

	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func NewLocal(
	block *document.CodeBlock,
	options *LocalOptions,
) (Command, error) {
	if options == nil {
		options = &LocalOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}

	cfg, err := NewConfigBuilder(block).Build()
	if err != nil {
		return nil, err
	}

	stdin := options.Stdin

	if f, ok := stdin.(*os.File); ok && f != nil {
		// Duplicate /dev/stdin.
		newStdinFd, err := syscall.Dup(int(f.Fd()))
		if err != nil {
			return nil, errors.Wrap(err, "failed to dup stdin")
		}
		syscall.CloseOnExec(newStdinFd)

		// TODO(adamb): setting it to the non-block mode fails on the simple "read" command.
		// if err := syscall.SetNonblock(newStdinFd, true); err != nil {
		// 	return nil, errors.Wrap(err, "failed to set new stdin fd in non-blocking mode")
		// }

		stdin = os.NewFile(uintptr(newStdinFd), "")
	}

	cmd := &localCommand{
		Name:   cfg.Name,
		Path:   cfg.Path,
		Args:   cfg.Args,
		PreEnv: options.Env,
		Dir:    resolveDir(options.ParentDir, cfg.Dirs),
		Stdin:  stdin,
		Stdout: options.Stdout,
		Stderr: options.Stderr,
		Logger: options.Logger.With(zap.String("name", cfg.Name)),
	}

	return cmd, nil
}
