package command

import (
	"context"
	"io"
	"os"

	"github.com/stateful/runme/internal/document"
	"go.uber.org/zap"
)

type Command interface {
	IsRunning() bool
	PID() int
	Start(context.Context) error
	StopWithSignal(os.Signal) error
	Wait() error
}

type NativeCommandOptions struct {
	ParentDir string

	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func NewNative(
	block *document.CodeBlock,
	options *NativeCommandOptions,
) (Command, error) {
	if options == nil {
		options = &NativeCommandOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}

	cfg, err := NewConfigBuilder(block).Build()
	if err != nil {
		return nil, err
	}

	return newNativeCommand(cfg, options), nil
}

type VirtualCommandOptions struct {
	ParentDir string

	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func NewVirtual(
	block *document.CodeBlock,
	options *VirtualCommandOptions,
) (Command, error) {
	if options == nil {
		options = &VirtualCommandOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}

	cfg, err := NewConfigBuilder(block).Build()
	if err != nil {
		return nil, err
	}

	return newVirtualCommand(cfg, options), nil
}
