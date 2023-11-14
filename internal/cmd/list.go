package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/pkg/jsonpretty"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/project"
	"github.com/stateful/runme/internal/shell"
)

type row struct {
	Name         string `json:"name"`
	File         string `json:"file"`
	FirstCommand string `json:"first_command"`
	Description  string `json:"description"`
	Named        bool   `json:"named"`
}

func listCmd() *cobra.Command {
	var formatJSON bool
	cmd := cobra.Command{
		Use:     "list [search]",
		Aliases: []string{"ls"},
		Short:   "List available commands",
		Long:    "Displays list of parsed command blocks, their name, number of commands in a block, and description from a given markdown file, such as README.md. Provide an argument to filter results by file and name using a regular expression.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			search := ""
			if len(args) > 0 {
				search = args[0]
			}

			proj, err := getProject()
			if err != nil {
				return err
			}

			loader, err := newTUIProjectLoader(cmd)
			if err != nil {
				return err
			}

			allTasks, err := loader.LoadTasks(proj, &loadTasksConfig{AllowUnknown: fAllowUnknown, AllowUnnamed: fAllowUnnamed})
			if err != nil {
				return err
			}

			foundTasks, err := allTasks.LookupByID(search)
			if err != nil {
				return err
			}

			if len(foundTasks) == 0 && !fAllowUnnamed {
				return errors.Errorf("no named code blocks, consider adding flag --allow-unnamed")
			}

			foundTasks = sortTasks(foundTasks)

			// TODO: this should be taken from cmd.
			io := iostreams.System()
			var rows []row
			for _, task := range foundTasks {
				block := task.CodeBlock
				lines := block.Lines()
				r := row{
					Name:         block.Name(),
					File:         task.Filename,
					FirstCommand: shell.TryGetNonCommentLine(lines),
					Description:  block.Intro(),
					Named:        !block.IsUnnamed(),
				}
				rows = append(rows, r)
			}
			if !formatJSON {
				return displayTable(io, rows)
			}

			return displayJSON(io, rows)
		},
	}

	cmd.PersistentFlags().BoolVar(&formatJSON, "json", false, "This flag tells the list command to print the output in json")
	setDefaultFlags(&cmd)

	return &cmd
}

func displayTable(io *iostreams.IOStreams, rows []row) error {
	table := tableprinter.New(io.Out, io.IsStdoutTTY(), io.TerminalWidth())

	// table header
	table.AddField(strings.ToUpper("Name"))
	table.AddField(strings.ToUpper("File"))
	table.AddField(strings.ToUpper("First Command"))
	table.AddField(strings.ToUpper("Description"))
	table.AddField(strings.ToUpper("Named"))
	table.EndRow()

	for _, row := range rows {
		named := "Yes"
		if !row.Named {
			named = "No"
		}
		table.AddField(row.Name)
		table.AddField(row.File)
		table.AddField(row.FirstCommand)
		table.AddField(row.Description)
		table.AddField(named)
		table.EndRow()
	}

	return errors.Wrap(table.Render(), "failed to render")
}

func displayJSON(io *iostreams.IOStreams, rows []row) error {
	by, err := json.Marshal(&rows)
	if err != nil {
		return err
	}
	return jsonpretty.Format(io.Out, bytes.NewReader(by), "  ", false)
}

func sortTasks(tasks project.Tasks) project.Tasks {
	tasksByFile := make(map[string]project.Tasks)

	files := make([]string, 0)
	for _, task := range tasks {
		if arr, ok := tasksByFile[task.Filename]; ok {
			tasksByFile[task.Filename] = append(arr, task)
			continue
		}

		tasksByFile[task.Filename] = project.Tasks{task}
		files = append(files, task.Filename)
	}

	sort.SliceStable(files, func(i, j int) bool {
		return getFileDepth(files[i]) < getFileDepth(files[j])
	})

	result := make(project.Tasks, 0, len(tasks))
	for _, file := range files {
		result = append(result, tasksByFile[file]...)
	}
	return result
}

func getFileDepth(fp string) int {
	return len(strings.Split(fp, string(filepath.Separator)))
}
