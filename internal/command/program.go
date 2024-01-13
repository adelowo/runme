package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

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

func programAndArgsFromCodeBlock(block *document.CodeBlock) (program string, args []string, err error) {
	interpreter := ""

	lang := block.Language()

	// If the language is a shell language, then infer the interpreter from the FrontMatter.
	if isShellLanguage(lang) {
		interpreter = shellFromFrontmatter(block)
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
