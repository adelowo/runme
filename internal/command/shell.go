package command

import (
	"path/filepath"
	"strings"

	"github.com/stateful/runme/internal/document"
)

func isShellLanguage(languageID string) bool {
	switch strings.ToLower(languageID) {
	// shellscripts
	case "sh", "bash", "zsh", "ksh", "shell", "shellscript":
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

// func resolveShellPath(customShell string) string {
// 	if customShell != "" {
// 		if path, err := exec.LookPath(customShell); err == nil {
// 			return path
// 		}
// 	}
// 	return systemShellPath()
// }

// func systemShellPath() string {
// 	shell, ok := os.LookupEnv("SHELL")
// 	if !ok {
// 		shell = "sh"
// 	}
// 	if path, err := exec.LookPath(shell); err == nil {
// 		return path
// 	}
// 	return "/bin/sh"
// }

// func isShellProgram(path string) bool {
// 	name := filepath.Base(path)

// 	for _, candidate := range []string{
// 		"bash", "sh", "ksh", "zsh", "fish", "powershell", "pwsh", "cmd",
// 	} {
// 		if name == candidate {
// 			return true
// 		}
// 	}

// 	return false
// }

func shellOptionsFromProgram(programPath string) (res string) {
	shell := shellFromProgramPath(programPath)

	// TODO(mxs): powershell and DOS are missing
	switch shell {
	case "zsh", "ksh", "bash":
		res += "set -e -x -o pipefail"
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

func prepareScript(block *document.CodeBlock, programPath string) string {
	var buf strings.Builder

	_, _ = buf.WriteString(shellOptionsFromProgram(programPath) + ";")

	for _, cmd := range block.Lines() {
		_, _ = buf.WriteString(cmd)
		_, _ = buf.WriteRune(';')
	}

	return buf.String()
}
