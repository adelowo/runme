package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/command"
	"github.com/stateful/runme/internal/project"
)

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

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			localOptions := &command.LocalOptions{
				ParentDir: dir,
				Stdin:     cmd.InOrStdin(),
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.ErrOrStderr(),
				Logger:    logger,
			}

			localCmd, err := command.NewLocal(
				tasks[0].CodeBlock,
				localOptions,
			)
			if err != nil {
				return err
			}

			err = localCmd.Start(cmd.Context())
			if err != nil {
				return err
			}

			return localCmd.Wait()
		},
	}

	return &cmd
}
