package command

import (
	"context"
	"io"

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

	return newLocalCommand(cfg, options), nil
}

type RemoteOptions struct {
	ParentDir string

	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func NewRemote(
	block *document.CodeBlock,
	options *RemoteOptions,
) (Command, error) {
	if options == nil {
		options = &RemoteOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}

	cfg, err := NewConfigBuilder(block).Build()
	if err != nil {
		return nil, err
	}

	return newRemoteCommand(cfg, options), nil
}
