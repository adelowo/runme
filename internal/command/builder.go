package command

import (
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
	"go.uber.org/zap"
)

type CommandOptions struct {
	// TODO(adamb): figure out what this dir really is
	ParentDir string

	Env []string

	// Tty, if true, allocates a pseudo-terminal, which is used
	// as stdin, stdout, and stderr in exec.Cmd.
	Tty bool

	Stdin  io.ReadCloser
	Stdout io.Writer
	Stderr io.Writer

	Logger *zap.Logger
}

func CommandFromCodeBlock(
	block *document.CodeBlock,
	options *CommandOptions,
) (*Command, error) {
	if options == nil {
		options = &CommandOptions{
			Logger: zap.NewNop(),
		}
	}

	builder, err := newCommandBuilder(block, options)
	if err != nil {
		return nil, err
	}

	return builder.Build()
}

type commandBuilder interface {
	Build() (*Command, error)
}

func newCommandBuilder(
	block *document.CodeBlock,
	options *CommandOptions,
) (commandBuilder, error) {
	base := &baseBuilder{
		block:   block,
		options: options,
	}

	if isShellLanguage(block.Language()) {
		return &inlineShellBuilder{
			baseBuilder: base,
		}, nil
	}

	return &fileGenericBuilder{
		baseBuilder: base,
	}, nil
}

type baseBuilder struct {
	block   *document.CodeBlock
	options *CommandOptions
}

func (b *baseBuilder) Build() (*Command, error) {
	dir, err := b.dir()
	if err != nil {
		return nil, err
	}

	programPath, args, err := programPathAndArgsFromCodeBlock(b.block)
	if err != nil {
		return nil, err
	}

	preEnv := make([]string, 0, len(b.options.Env))
	copy(preEnv, b.options.Env)

	// Duplicate /dev/stdin.
	newStdinFd, err := syscall.Dup(int(b.options.Stdin.(*os.File).Fd()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to dup stdin")
	}
	syscall.CloseOnExec(newStdinFd)

	if err := syscall.SetNonblock(newStdinFd, true); err != nil {
		return nil, errors.Wrap(err, "failed to set new stdin fd in non-blocking mode")
	}

	stdin := os.NewFile(uintptr(newStdinFd), "")

	cmd := &Command{
		Name:   b.block.Name(),
		Path:   programPath,
		Args:   args,
		PreEnv: preEnv,
		Dir:    dir,
		Stdin:  stdin,
		Stdout: b.options.Stdout,
		Stderr: b.options.Stderr,
		Logger: b.options.Logger.With(zap.String("name", b.block.Name())),
	}

	if b.options.Tty {
		if err := cmd.setTty(); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func (b *baseBuilder) dir() (string, error) {
	dir := resolveDirFromCodeBlock(b.block, b.options.ParentDir)
	if dir != "" {
		return dir, nil
	}

	dir, err := os.Getwd()
	return dir, errors.WithStack(err)
}

type inlineShellBuilder struct {
	*baseBuilder
}

func (b *inlineShellBuilder) Build() (*Command, error) {
	cmd, err := b.baseBuilder.Build()
	if err != nil {
		return nil, err
	}

	if b.options.Tty {
		cmd.Args = append(cmd.Args, "-i")
	}

	cmd.Args = append(cmd.Args, "-c", b.prepareScript(cmd.Path))

	return cmd, nil
}

func (b *inlineShellBuilder) prepareScript(programPath string) string {
	var buf strings.Builder

	_, _ = buf.WriteString(shellOptionsFromProgram(programPath))

	for _, cmd := range b.block.Lines() {
		_, _ = buf.WriteString(cmd)
		_, _ = buf.WriteRune('\n')
	}

	// TODO(adamb): verify if this is needed
	_, _ = buf.WriteRune('\n')

	return buf.String()
}

type fileGenericBuilder struct {
	*baseBuilder
}

func (b *fileGenericBuilder) Build() (*Command, error) {
	return nil, nil
}
