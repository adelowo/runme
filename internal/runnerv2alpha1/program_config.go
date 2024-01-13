package runnerv2alpha1

import (
	"path/filepath"
	"strings"

	"github.com/stateful/runme/internal/command"
	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

func newConfigFromProtoProgramConfig(cfg *runnerv2alpha1.ProgramConfig) (*command.Config, error) {
	return (&configBuilder{cfg: cfg}).Build()
}

type configBuilder struct {
	cfg *runnerv2alpha1.ProgramConfig
}

func (b *configBuilder) script(programPath string) string {
	var buf strings.Builder

	_, _ = buf.WriteString(shellOptionsFromProgram(programPath) + ";")

	for _, cmd := range b.cfg.GetCommands().Commands {
		_, _ = buf.WriteString(cmd)
		_, _ = buf.WriteRune('\n')
	}

	return buf.String()

}

func (b *configBuilder) Build() (*command.Config, error) {
	cfg := &command.Config{
		ProgramPath: b.cfg.ProgramName,
	}

	// Using "-i" options seems to be not needed.

	cfg.Interactive = b.cfg.Interactive

	if script := b.script(b.cfg.ProgramName); script != "" {
		cfg.Args = append(cfg.Args, "-c", script)
	}

	return cfg, nil
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
