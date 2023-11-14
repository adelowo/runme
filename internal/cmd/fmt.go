package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/document/editor"
	"github.com/stateful/runme/internal/renderer/cmark"
)

func fmtCmd() *cobra.Command {
	var (
		formatJSON bool
		flatten    bool
		write      bool
	)

	cmd := cobra.Command{
		Use:   "fmt",
		Short: "Format a Markdown file into canonical format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if formatJSON {
				if write {
					return errors.New("invalid usage of --json with --write")
				}
				if !flatten {
					return errors.New("invalid usage of --json without --flatten")
				}
			}

			proj, err := getProject()
			if err != nil {
				return err
			}

			files := args

			if len(files) == 0 {
				loader, err := newTUIProjectLoader(cmd)
				if err != nil {
					return err
				}

				projectFiles, err := loader.LoadFiles(proj)
				if err != nil {
					return err
				}

				files = projectFiles
			}

			return fmtFiles(files, proj.Root(), flatten, formatJSON, write, func(file string, formatted []byte) error {
				out := cmd.OutOrStdout()
				_, _ = fmt.Fprintf(out, "===== %s =====\n", file)
				_, _ = out.Write(formatted)
				_, _ = fmt.Fprint(out, "\n")
				return nil
			})
		},
	}

	setDefaultFlags(&cmd)

	cmd.Flags().BoolVar(&flatten, "flatten", true, "Flatten nested blocks in the output. WARNING: This can currently break frontmatter if turned off.")
	cmd.Flags().BoolVar(&formatJSON, "json", false, "Print out data as JSON. Only possible with --flatten and not allowed with --write.")
	cmd.Flags().BoolVarP(&write, "write", "w", false, "Write result to the source file instead of stdout.")

	return &cmd
}

type funcOutput func(string, []byte) error

func fmtFiles(files []string, root string, flatten bool, formatJSON bool, write bool, outputter funcOutput) error {
	for _, relFile := range files {
		data, err := readMarkdown(filepath.Join(root, relFile))
		if err != nil {
			return err
		}

		var formatted []byte

		if flatten {
			notebook, err := editor.Deserialize(data)
			if err != nil {
				return errors.Wrap(err, "failed to deserialize")
			}

			if formatJSON {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetIndent("", "  ")
				if err := enc.Encode(notebook); err != nil {
					return errors.Wrap(err, "failed to encode to JSON")
				}
				formatted = buf.Bytes()
			} else {
				formatted, err = editor.Serialize(notebook)
				if err != nil {
					return errors.Wrap(err, "failed to serialize")
				}
			}
		} else {
			doc := document.New(data)
			astNode, err := doc.RootASTNode()
			if err != nil {
				return errors.Wrap(err, "failed to parse source")
			}
			formatted, err = cmark.Render(astNode, data)
			if err != nil {
				return errors.Wrap(err, "failed to render")
			}
		}

		if write {
			err = writeMarkdown(filepath.Join(root, relFile), formatted)
		} else {
			err = outputter(relFile, formatted)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func writeMarkdown(destination string, data []byte) error {
	if destination == "-" {
		if _, err := os.Stdout.Write(data); err != nil {
			return errors.Wrap(err, "failed to write to stdout")
		}
	} else if strings.HasPrefix(destination, "https://") {
		return errors.New("cannot write to HTTP location")
	}
	err := os.WriteFile(destination, data, 0o600)
	return errors.Wrapf(err, "failed to write data to %q", destination)
}

func readMarkdown(source string) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	if source == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read from stdin")
		}
	} else if strings.HasPrefix(source, "https://") {
		client := http.Client{
			Timeout: time.Second * 5,
		}
		resp, err := client.Get(source)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get a file %q", source)
		}
		defer func() { _ = resp.Body.Close() }()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read body")
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read from file %q", source)
		}
	}

	return data, nil
}
