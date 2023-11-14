package cmd

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/pkg/project"
)

func printCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:               "print",
		Short:             "Print a selected snippet",
		Long:              "Print will display the details of the corresponding command block based on its name.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: validCmdNames,
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

			tasks, err := loader.LoadTasks(proj, &loadTasksConfig{AllowUnknown: fAllowUnknown, AllowUnnamed: fAllowUnnamed})
			if err != nil {
				return err
			}

			task, err := lookupTaskWithPrompt(cmd, args[0], tasks)
			if err != nil {
				if project.IsCodeBlockNotFoundError(err) && !fAllowUnnamed {
					fAllowUnnamed = true
					goto generateBlocks
				}

				return err
			}

			block := task.CodeBlock

			w := bulkWriter{
				Writer: cmd.OutOrStdout(),
			}
			w.Write([]byte(block.Value()))
			w.Write([]byte{'\n'})
			return errors.Wrap(w.Err(), "failed to write to stdout")
		},
	}

	setDefaultFlags(&cmd)

	return &cmd
}

type bulkWriter struct {
	io.Writer
	n   int
	err error
}

func (w *bulkWriter) Err() error {
	return w.err
}

func (w *bulkWriter) Result() (int, error) {
	return w.n, w.err
}

func (w *bulkWriter) Write(p []byte) {
	if w.err != nil {
		return
	}
	n, err := w.Writer.Write(p)
	w.n += n
	w.err = err
}
