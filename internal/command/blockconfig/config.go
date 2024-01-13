package blockconfig

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"

	"github.com/stateful/runme/internal/command"
	"github.com/stateful/runme/internal/document"
	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

func New(block *document.CodeBlock) (*command.Config, error) {
	base := &baseConfigBuilder{
		block: block,
	}

	var builder interface {
		Build() (*command.Config, error)
	}

	switch {
	case isShellLanguage(block.Language()):
		builder = &inlineShellConfigBuilder{
			baseConfigBuilder: base,
		}
	default:
		builder = &fileConfigBuilder{
			baseConfigBuilder: base,
		}
	}

	return builder.Build()
}

type baseConfigBuilder struct {
	block *document.CodeBlock
}

func (b *baseConfigBuilder) dir() string {
	var dirs []string

	fmtr, err := b.block.Document().Frontmatter()
	if err == nil && fmtr != nil && fmtr.Cwd != "" {
		dirs = append(dirs, fmtr.Cwd)
	}

	if dir := b.block.Cwd(); dir != "" {
		dirs = append(dirs, dir)
	}

	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	// TODO(adamb): figure out the first argument.
	return resolveDir("", dirs)
}

func (b *baseConfigBuilder) pathAndArgs() (string, []string, error) {
	interpreters, err := interpretersFromCodeBlock(b.block)
	if err != nil {
		return "", nil, err
	}

	for _, interpreter := range interpreters {
		program, args := parseInterpreter(interpreter)
		if path, err := exec.LookPath(program); err == nil {
			return path, args, nil
		}
	}

	return "", nil, errors.WithStack(&ErrUnknownInterpreters{interpreters})
}

func (b *baseConfigBuilder) Build() (*command.Config, error) {
	path, args, err := b.pathAndArgs()
	if err != nil {
		return nil, err
	}

	cfg := &command.Config{
		ProgramName: path,
		Directory:   b.dir(),
		Arguments:   args,
	}

	return cfg, nil
}

type inlineShellConfigBuilder struct {
	*baseConfigBuilder
}

func (b *inlineShellConfigBuilder) Build() (*command.Config, error) {
	cfg, err := b.baseConfigBuilder.Build()
	if err != nil {
		return nil, err
	}

	// Using "-i" options seems to be not needed.

	cfg.Interactive = b.block.Interactive()

	if script := prepareScriptFromLines(cfg.ProgramName, b.block.Lines()); script != "" {
		cfg.Arguments = append(cfg.Arguments, "-c", script)
	}

	return cfg, nil
}

type fileConfigBuilder struct {
	*baseConfigBuilder
}

func (b *fileConfigBuilder) Build() (*command.Config, error) {
	cfg, err := b.baseConfigBuilder.Build()
	if err != nil {
		return nil, err
	}

	cfg.Mode = runnerv2alpha1.CommandMode_COMMAND_MODE_FILE

	cfg.Interactive = b.block.Interactive()

	cfg.Source = &runnerv2alpha1.ProgramConfig_Script{
		Script: prepareScriptFromLines(cfg.ProgramName, b.block.Lines()),
	}

	return cfg, nil
}
