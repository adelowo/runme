package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

type Config struct {
	Name        string
	Dirs        []string
	Path        string
	Args        []string
	Interactive bool
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

func (b *baseConfigBuilder) Build() (*Config, error) {
	var dirs []string

	fmtr, err := b.block.Document().Frontmatter()
	if err == nil && fmtr != nil && fmtr.Cwd != "" {
		dirs = append(dirs, fmtr.Cwd)
	}

	if dir := b.block.Cwd(); dir != "" {
		dirs = append(dirs, dir)
	}

	path, args, err := programPathAndArgsFromCodeBlock(b.block)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Name: b.block.Name(),
		Dirs: dirs,
		Path: path,
		Args: args,
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

	// TODO(adamb): this seems to be completely unnecessary.
	// if b.block.Interactive() {
	// 	cfg.Args = append(cfg.Args, "-i")
	// }
	cfg.Interactive = b.block.Interactive()

	if script := prepareScript(b.block, cfg.Path); script != "" {
		cfg.Args = append(cfg.Args, "-c", script)
	}

	return cfg, nil
}

type fileConfigBuilder struct {
	*baseConfigBuilder
}

func (b *fileConfigBuilder) Build() (*Config, error) {
	return b.baseConfigBuilder.Build()
}

type ErrInvalidLanguage struct {
	langID string
}

func (e ErrInvalidLanguage) Error() string {
	return fmt.Sprintf("unsupported language %s", e.langID)
}

type ErrInvalidProgram struct {
	program string
	inner   error
}

func (e ErrInvalidProgram) Error() string {
	return fmt.Sprintf("unable to locate program %s", e.program)
}

func (e ErrInvalidProgram) Unwrap() error {
	return e.inner
}

func programPathAndArgsFromCodeBlock(block *document.CodeBlock) (program string, args []string, err error) {
	interpreter := ""

	lang := block.Language()

	// If the language is a shell language, then infer the interpreter from the FrontMatter.
	if isShellLanguage(lang) {
		fmtr, err := block.Document().Frontmatter()
		if err == nil && fmtr != nil {
			interpreter = fmtr.Shell
		}
	}

	// Interpreter can be always overwritten at the block level.
	if val := block.Interpreter(); val != "" {
		interpreter = val
	}

	// If the interpreter is empty, then infer it from the language.
	// There might be more than one cadidate. In such case, use LookPath()
	// to find the first one that exists in the system.
	if interpreter == "" {
		interpreters := inferInterpreterFromLanguage(lang)

		if len(interpreters) == 1 {
			interpreter = interpreters[0]
		} else {
			for _, interpreter = range interpreters {
				if path, err := exec.LookPath(interpreter); err == nil {
					interpreter = path
					break
				}
			}
		}
	}

	// If it's still empty, then return an error.
	if interpreter == "" {
		return "", nil, errors.WithStack(&ErrInvalidLanguage{lang})
	}

	program, args = parseInterpreter(interpreter)

	if path, err := exec.LookPath(program); err == nil {
		program = path
	} else {
		return "", nil, errors.WithStack(&ErrInvalidProgram{program, err})
	}

	return
}

// parseInterpreter handles cases when the interpreter is, for instance, "deno run".
// Only the first word is a program name and the rest is arguments.
func parseInterpreter(interpreter string) (program string, args []string) {
	parts := strings.SplitN(interpreter, " ", 2)

	if len(parts) > 0 {
		program = parts[0]
	}

	if len(parts) > 1 {
		args = strings.Split(parts[1], " ")
	}

	return
}

var interpreterByLanguageID = map[string][]string{
	"js":              {"node"},
	"javascript":      {"node"},
	"jsx":             {"node"},
	"javascriptreact": {"node"},

	"ts":              {"ts-node", "deno run", "bun run"},
	"typescript":      {"ts-node", "deno run", "bun run"},
	"tsx":             {"ts-node", "deno run", "bun run"},
	"typescriptreact": {"ts-node", "deno run", "bun run"},

	"sh":         {"bash", "sh"},
	"bash":       {"bash", "sh"},
	"ksh":        {"ksh"},
	"zsh":        {"zsh"},
	"fish":       {"fish"},
	"powershell": {"powershell"},
	"cmd":        {"cmd"},
	"dos":        {"cmd"},

	"lua":    {"lua"},
	"perl":   {"perl"},
	"php":    {"php"},
	"python": {"python3", "python"},
	"py":     {"python3", "python"},
	"ruby":   {"ruby"},
	"rb":     {"ruby"},
}

func inferInterpreterFromLanguage(langID string) []string {
	return interpreterByLanguageID[langID]
}
