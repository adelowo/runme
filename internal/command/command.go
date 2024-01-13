package command

import (
	"io"

	"go.uber.org/zap"

	"github.com/stateful/runme/internal/document"
)

type NativeCommandOptions struct {
	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func NewNative(
	block *document.CodeBlock,
	options *NativeCommandOptions,
) (*NativeCommand, error) {
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
	Env []string

	Stdin  io.Reader
	Stdout io.Writer

	Logger *zap.Logger
}

func NewVirtual(
	block *document.CodeBlock,
	options *VirtualCommandOptions,
) (*VirtualCommand, error) {
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
