package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/project"
)

type tuiProjectLoader struct {
	w          io.Writer
	r          io.Reader
	isTerminal bool
}

func newTUIProjectLoader(cmd *cobra.Command) (*tuiProjectLoader, error) {
	fd := os.Stdout.Fd()

	if int(fd) < 0 {
		return nil, fmt.Errorf("invalid file descriptor due to restricted environments, redirected standard output, system configuration issues, or testing/simulation setups")
	}

	return &tuiProjectLoader{
		w:          cmd.OutOrStdout(),
		r:          cmd.InOrStdin(),
		isTerminal: isTerminal(fd),
	}, nil
}

type loadTasksModel struct {
	spinner spinner.Model

	status   string
	filename string

	clear bool

	err error

	tasks project.Tasks
	files []string

	nextTaskMsg tea.Cmd
}

type loadTaskFinished struct{}

func (pl tuiProjectLoader) newLoadTasksModel(nextTaskMsg tea.Cmd) loadTasksModel {
	return loadTasksModel{
		spinner:     spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		nextTaskMsg: nextTaskMsg,
		status:      "Initializing...",
	}
}

func (pl tuiProjectLoader) LoadFiles(proj *project.Project) ([]string, error) {
	m, err := pl.runTasksModel(proj, true)
	if err != nil {
		return nil, err
	}
	return m.files, nil
}

type loadTasksConfig struct {
	AllowUnknown bool
	AllowUnnamed bool
}

func (pl tuiProjectLoader) LoadTasks(proj *project.Project, cfg *loadTasksConfig) (project.Tasks, error) {
	m, err := pl.runTasksModel(proj, false)
	if err != nil {
		return nil, err
	}

	tasks := m.tasks

	if cfg != nil {
		filteredTasks := make(project.Tasks, 0, len(tasks))

		for _, task := range tasks {
			if !cfg.AllowUnknown && task.CodeBlock.IsUnknown() {
				continue
			}

			if !cfg.AllowUnnamed && task.CodeBlock.IsUnnamed() {
				continue
			}

			filteredTasks = append(filteredTasks, task)
		}

		tasks = filteredTasks
	}

	return tasks, nil
}

func (pl tuiProjectLoader) runTasksModel(proj *project.Project, onlyFiles bool) (*loadTasksModel, error) {
	eventc := make(chan project.LoadEvent)

	go proj.Load(context.TODO(), eventc, onlyFiles)

	nextTaskMsg := func() tea.Msg {
		msg, ok := <-eventc
		if !ok {
			return loadTaskFinished{}
		}
		return msg
	}

	m := pl.newLoadTasksModel(nextTaskMsg)

	resultModel := m

	if pl.isTerminal {
		p := tea.NewProgram(m, tea.WithOutput(pl.w), tea.WithInput(pl.r))
		result, err := p.Run()
		if err != nil {
			return nil, err
		}

		resultModel = result.(loadTasksModel)
	} else {
		// TODO(adamb): move to a flag
		if strings.ToLower(os.Getenv("RUNME_VERBOSE")) != "true" {
			pl.w = io.Discard
		}

		_, _ = fmt.Fprintln(pl.w, "Initializing...")

	outer:
		for {
			if resultModel.err != nil {
				break
			}

			switch msg := nextTaskMsg().(type) {
			case loadTaskFinished:
				_, _ = fmt.Fprintln(pl.w, "")
				break outer
			case project.LoadEvent:
				switch msg.Type {
				case project.LoadEventStartedWalk:
					_, _ = fmt.Fprintln(pl.w, "Searching for files...")
				case project.LoadEventFinishedWalk:
					_, _ = fmt.Fprintln(pl.w, "Parsing files...")
				}
			default:
				if newModel, ok := resultModel.TaskUpdate(msg).(loadTasksModel); ok {
					resultModel = newModel
				}
			}
		}
	}

	if resultModel.err != nil {
		return nil, resultModel.err
	}

	return &resultModel, nil
}

func (m loadTasksModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return m.spinner.Tick()
		},
		m.nextTaskMsg,
	)
}

func (m loadTasksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case loadTaskFinished:
		m.clear = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "crtl+d":
			m.err = errors.New("aborted")
			return m, tea.Quit
		default:
			return m, nil
		}
	}

	if m, ok := m.TaskUpdate(msg).(loadTasksModel); ok {
		return m, m.nextTaskMsg
	}

	return m, nil
}

func (m loadTasksModel) TaskUpdate(msg tea.Msg) tea.Model {
	switch msg := msg.(type) {
	case project.LoadEvent:
		switch msg.Type {
		case project.LoadEventError:
			m.err = msg.Data.(error)
		case project.LoadEventStartedWalk:
			m.filename = ""
			m.status = "Searching for files..."
		case project.LoadEventFinishedWalk:
			m.filename = ""
			m.status = "Parsing files..."
		case project.LoadEventFoundDir:
			m.filename = msg.Data.(string)
		case project.LoadEventStartedParsingDocument:
			m.filename = msg.Data.(string)
		case project.LoadEventFoundFile:
			m.files = append(m.files, msg.Data.(string))
		case project.LoadEventFoundTask:
			m.tasks = append(m.tasks, msg.Data.(project.Task))
		}
	default:
		panic("invariant: TaskUpdate called with invalid message type")
	}

	return m
}

func (m loadTasksModel) View() (s string) {
	if m.clear {
		return
	}

	s += m.spinner.View()
	s += " "

	s += m.status

	if m.filename != "" {
		s += fmt.Sprintf(" (%s)", m.filename)
	}

	return
}
