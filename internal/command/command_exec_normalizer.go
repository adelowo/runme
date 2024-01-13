package command

import (
	"os"
	"strings"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type argsNormalizer struct {
	tempFile *os.File
	logger   *zap.Logger
}

func (n *argsNormalizer) Normalize(cfg *Config) (_ *Config, err error) {
	args := append([]string{}, cfg.Arguments...)

	switch cfg.Mode {
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_UNSPECIFIED.Enum():
		panic("invariant: mode unspecified")
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_INLINE.Enum():
		var buf strings.Builder

		if options := shellOptionsFromProgram(cfg.ProgramName); options != "" {
			_, _ = buf.WriteString(options)
			_, _ = buf.WriteString("\n\n")
		}

		if commands := cfg.GetCommands(); commands != nil {
			for _, cmd := range commands.Items {
				_, _ = buf.WriteString(cmd)
				_, _ = buf.WriteRune('\n')
			}
		}

		if script := buf.String(); script != "" {
			args = append(args, "-c", script)
		}
	case *runnerv2alpha1.CommandMode_COMMAND_MODE_FILE.Enum():
		n.tempFile, err = createTempFileFromScript(cfg)
		if err != nil {
			return
		}

		// TODO(adamb): it's not always true that the script-based program
		// takes the filename as a last argument.
		args = append(args, n.tempFile.Name())
	}

	result := proto.Clone(cfg).(*Config)
	result.Arguments = args
	return result, nil
}

func (n *argsNormalizer) cleanup() {
	if n.tempFile == nil {
		return
	}
	if err := os.Remove(n.tempFile.Name()); err != nil {
		n.logger.Info("failed to remove temporary file", zap.Error(err))
	}
}

type envNormalizer struct {
	opts interface{ GetEnv() []string }
}

func (n *envNormalizer) Normalize(cfg *Config) (*Config, error) {
	result := proto.Clone(cfg).(*Config)

	env := os.Environ()
	env = append(env, cfg.Env...)
	env = append(env, n.opts.GetEnv()...)

	result.Env = env

	return result, nil
}
