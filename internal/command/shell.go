package command

import (
	"path/filepath"
	"strings"

	"github.com/stateful/runme/internal/document"
)

func isShellLanguage(languageID string) bool {
	switch strings.ToLower(languageID) {
	// shellscripts
	// TODO(adamb): breaking change: shellscript was removed to indicate
	// that it should be executed as a file. Consider adding it back and
	// using attributes to decide how a code block should be executed.
	case "sh", "bash", "zsh", "ksh", "shell":
		return true

	// dos
	case "bat", "cmd":
		return true

	// powershell
	case "powershell", "pwsh":
		return true

	// fish
	case "fish":
		return true

	default:
		return false
	}
}

func prepareScript(block *document.CodeBlock, programPath string) string {
	var buf strings.Builder

	_, _ = buf.WriteString(shellOptionsFromProgram(programPath) + ";")

	for _, cmd := range block.Lines() {
		_, _ = buf.WriteString(cmd)
		_, _ = buf.WriteRune(';')
	}

	return buf.String()
}

func shellOptionsFromProgram(programPath string) (res string) {
	shell := shellFromProgramPath(programPath)

	// TODO(mxs): powershell and DOS are missing
	switch shell {
	case "zsh", "ksh", "bash":
		res += "set -e -o pipefail"
	case "sh":
		res += "set -e"
	}

	return
}

// TODO(mxs): this method for determining shell is not strong, since shells can
// be aliased. we should probably run the shell to get this information
func shellFromProgramPath(programPath string) string {
	programFile := filepath.Base(programPath)
	return programFile[:len(programFile)-len(filepath.Ext(programFile))]
}
