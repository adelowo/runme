package command

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

// Config represents a Command configuration. It's built
// from a CodeBlock and unifies the configuration for both
// inline shell and file-based commands.
type Config struct {
	// Name is the name of the command.
	Name string

	// Dir is a directory in which the command should be executed.
	Dir string

	ProgramPath string

	Args []string

	// Interactive, if true, allows the input from the user.
	Interactive bool

	// TempDir is a temporary directory.
	TempDir string

	// ScriptPath is a path to a script file.
	// It is also the last argument in Args.
	ScriptPath string
}

type ConfigBuilder interface {
	Build() (*Config, error)
}

func NewConfigBuilder(block *document.CodeBlock) ConfigBuilder {
	base := &baseConfigBuilder{
		block: block,
	}

	switch {
	case isShellLanguage(block.Language()):
		return &inlineShellConfigBuilder{
			baseConfigBuilder: base,
		}
	default:
		return &fileConfigBuilder{
			baseConfigBuilder: base,
		}
	}
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

func (b *baseConfigBuilder) Build() (*Config, error) {
	path, args, err := b.pathAndArgs()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Name:        b.block.Name(),
		Dir:         b.dir(),
		ProgramPath: path,
		Args:        args,
	}

	return cfg, nil
}

type inlineShellConfigBuilder struct {
	*baseConfigBuilder
}

func (b *inlineShellConfigBuilder) Build() (*Config, error) {
	cfg, err := b.baseConfigBuilder.Build()
	if err != nil {
		return nil, err
	}

	// Using "-i" options seems to be not needed.

	cfg.Interactive = b.block.Interactive()

	if script := prepareScript(b.block, cfg.ProgramPath); script != "" {
		cfg.Args = append(cfg.Args, "-c", script)
	}

	return cfg, nil
}

type fileConfigBuilder struct {
	*baseConfigBuilder
}

func (b *fileConfigBuilder) Build() (*Config, error) {
	cfg, err := b.baseConfigBuilder.Build()
	if err != nil {
		return nil, err
	}

	cfg.Interactive = b.block.Interactive()

	cfg.TempDir, err = os.MkdirTemp("", "runme-*")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	scriptFile := filepath.Join(cfg.TempDir, "script-file-"+b.block.Name()+"."+b.block.Language())

	err = os.WriteFile(scriptFile, b.block.Content(), 0o600)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	cfg.ScriptPath = scriptFile

	cfg.Args = append(cfg.Args, scriptFile)

	return cfg, nil
}
