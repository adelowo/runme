package command

import (
	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

// Config contains a serializable configuration for a command.
// It's agnostic to the runtime or particular execution settings.
type Config = runnerv2alpha1.ProgramConfig
