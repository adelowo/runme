package command

import (
	"io"

	"go.uber.org/zap"
)

type NativeCommandOptions struct {
	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func (o *NativeCommandOptions) GetEnv() []string { return o.Env }

func NewNative(cfg *Config, options *NativeCommandOptions) (*NativeCommand, error) {
	if options == nil {
		options = &NativeCommandOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}

	return newNativeCommand(cfg, options), nil
}

type VirtualCommandOptions struct {
	Env []string

	Stdin  io.Reader
	Stdout io.Writer

	Logger *zap.Logger
}

func (o *VirtualCommandOptions) GetEnv() []string { return o.Env }

func NewVirtual(cfg *Config, options *VirtualCommandOptions) (*VirtualCommand, error) {
	if options == nil {
		options = &VirtualCommandOptions{}
	}

	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}

	return newVirtualCommand(cfg, options), nil
}
