package command

import (
	"strings"

	"github.com/stateful/runme/internal/document"
)

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

	program, args, err := programAndArgsFromCodeBlock(b.block)
	if err != nil {
		return nil, err
	}

	preEnv := make([]string, 0, len(b.options.Env))
	copy(preEnv, b.options.Env)

	cmd := &Command{
		Name:   b.block.Name(),
		Path:   program,
		Args:   args,
		PreEnv: preEnv,
		Dir:    dir,
		Stdin:  b.options.Stdin,
		Stdout: b.options.Stdout,
		Stderr: b.options.Stderr,
	}

	return cmd, nil
}

func (b *baseBuilder) dir() (string, error) {
	return resolveDirectory(
		resolveDirectoryFromCodeBlock(b.block, b.options.ParentDir),
	)
}

type inlineShellBuilder struct {
	*baseBuilder
}

func (b *inlineShellBuilder) Build() (*Command, error) {
	cmd, err := b.baseBuilder.Build()
	if err != nil {
		return nil, err
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
