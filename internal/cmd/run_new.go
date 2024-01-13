package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/stateful/runme/internal/command"
	"github.com/stateful/runme/internal/command/blockconfig"
	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/project"
)

// TODO(adamb): missing options:
// - --dry-run: Print the final command without executing.
// - --replace, -r: Replace instructions using sed.
// - --parallel, -p: Run tasks in parallel.
// - --all, -a: Run all tasks.
// - --skip-prompts: Skip prompting for variables.
// - --category, -c: Run from a specific category.
// - --index, -i: Index of command to run, 0-based. (Ignored in project mode.)
//
// Missing features:
// - [ ] Select tasks by index, if provided.
// - [ ] Select tasks by category, if provided.
// - [ ] Lookup tasks with prompt.
// - [ ] Selecting runner based on runner options.
// - [ ] Confirm execution using prompt.
// - [ ] Run tasks in parallel using multi runner.

func runNewCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "run-new <commands>",
		Short: "Run a selected command",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := getProject()
			if err != nil {
				return err
			}

			tasks, err := project.LoadTasks(cmd.Context(), proj)
			if err != nil {
				return err
			}

			tasks, err = project.FilterTasksByFileAndTaskName(tasks, "", args[0])
			if err != nil {
				return err
			}

			logger, err := getLogger(true)
			if err != nil {
				return err
			}
			defer logger.Sync()

			return runCommandNatively(cmd, tasks[0].CodeBlock, logger)
		},
	}

	return &cmd
}

func runCommandNatively(cmd *cobra.Command, block *document.CodeBlock, logger *zap.Logger) error {
	cfg, err := blockconfig.New(block)
	if err != nil {
		return err
	}

	opts := &command.NativeCommandOptions{
		Stdin:  cmd.InOrStdin(),
		Stdout: cmd.OutOrStdout(),
		Stderr: cmd.ErrOrStderr(),
		Logger: logger,
	}

	nativeCmd, err := command.NewNative(cfg, opts)
	if err != nil {
		return err
	}

	err = nativeCmd.Start(cmd.Context())
	if err != nil {
		return err
	}

	return nativeCmd.Wait()
}
