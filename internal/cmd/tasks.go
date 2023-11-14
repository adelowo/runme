package cmd

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/tasks"
	"github.com/stateful/runme/pkg/project"
)

func tasksCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:    "tasks",
		Short:  "Generates task.json for VS Code editor. Caution, this is experimental.",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := getProject()
			if err != nil {
				return err
			}

		generateBlocks:
			loader, err := newTUIProjectLoader(cmd)
			if err != nil {
				return err
			}

			blocks, err := loader.LoadTasks(proj, &loadTasksConfig{AllowUnknown: fAllowUnknown, AllowUnnamed: fAllowUnnamed})
			if err != nil {
				return err
			}

			block, err := lookupTaskWithPrompt(cmd, args[0], blocks)
			if err != nil {
				if project.IsCodeBlockNotFoundError(err) && !fAllowUnnamed {
					fAllowUnnamed = true
					goto generateBlocks
				}

				return err
			}

			tasksDef, err := tasks.GenerateFromShellCommand(
				block.CodeBlock.Name(),
				block.CodeBlock.Lines()[0],
				&tasks.ShellCommandOpts{
					Cwd: fChdir,
				},
			)
			if err != nil {
				return errors.Wrap(err, "failed to generate tasks.json")
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(tasksDef); err != nil {
				return errors.Wrap(err, "failed to marshal tasks.json")
			}

			return nil
		},
	}

	setDefaultFlags(&cmd)

	return &cmd
}
