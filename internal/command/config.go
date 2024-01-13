package command

import (
	"os"

	"github.com/pkg/errors"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

// Config contains a serializable configuration for a command.
// It's agnostic to the runtime or particular execution settings.
type Config = runnerv2alpha1.ProgramConfig

func createTempFileFromScript(cfg *Config) (*os.File, error) {
	f, err := os.CreateTemp("", "runme-script-*")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create a temporary file for script execution")
	}

	if _, err := f.Write([]byte(cfg.GetScript())); err != nil {
		return nil, errors.WithMessage(err, "failed to write the script to the temporary file")
	}

	if err := f.Close(); err != nil {
		return nil, errors.WithMessage(err, "failed to close the temporary file")
	}

	return f, nil
}

func envFromConfigAndOptions(cfg *Config, opts interface{ GetEnv() []string }) []string {
	// TODO(adamb): verify the order
	return append(append([]string{}, opts.GetEnv()...), cfg.Env...)
}
