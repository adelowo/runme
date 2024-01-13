package command

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

type ErrUnsupportedLanguage struct {
	langID string
}

func (e ErrUnsupportedLanguage) Error() string {
	return fmt.Sprintf("unsupported language %s", e.langID)
}

type ErrUnknownInterpreters struct {
	interpreters []string
}

func (e ErrUnknownInterpreters) Error() string {
	return fmt.Sprintf("unable to loop up any of interpreters %q", e.interpreters)
}

func interpretersFromCodeBlock(block *document.CodeBlock) ([]string, error) {
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

	if interpreter != "" {
		return []string{interpreter}, nil
	}

	// If the interpreter is empty, then infer it from the language.
	interpreters := inferInterpreterFromLanguage(lang)
	if len(interpreters) > 0 {
		return interpreters, nil
	}
	return nil, errors.WithStack(&ErrUnsupportedLanguage{lang})
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
