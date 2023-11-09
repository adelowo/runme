package project

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

type LoadEventType uint8

const (
	LoadEventStartedWalk LoadEventType = iota + 1
	LoadEventFoundDir
	LoadEventFoundFile
	LoadEventFinishedWalk
	LoadEventStartedParsingDocument
	LoadEventFinishedParsingDocument
	LoadEventFoundTask
	LoadEventError
)

type LoadEvent struct {
	Type LoadEventType
	Data any
}

type ProjectOption func(*Project)

func WithRespectGitignore() ProjectOption {
	return func(p *Project) {
		p.respectGitignore = true
	}
}

func WithIgnoreFilePatterns(patterns ...string) ProjectOption {
	return func(p *Project) {
		p.ignoreFilePatterns = patterns
	}
}

func WithGitPlainOpenOptions(options *git.PlainOpenOptions) ProjectOption {
	return func(p *Project) {
		p.plainOpenOptions = options
	}
}

type Project struct {
	// fs is used for git- or dir-based projects.
	fs billy.Filesystem
	// filePath is used for file-based projects.
	filePath string

	// Used when creating a git-based project.
	repo             *git.Repository
	plainOpenOptions *git.PlainOpenOptions
	respectGitignore bool

	// Used when creating a git- or dir-based project.
	ignoreFilePatterns []string
}

func NewGitProject(
	dir string,
	opts ...ProjectOption,
) (*Project, error) {
	p := &Project{}

	for _, opt := range opts {
		opt(p)
	}

	if p.plainOpenOptions == nil {
		p.plainOpenOptions = &git.PlainOpenOptions{}
	}

	var err error

	p.repo, err = git.PlainOpenWithOptions(
		dir,
		p.plainOpenOptions,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	wt, err := p.repo.Worktree()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	p.fs = wt.Filesystem

	return p, nil
}

func NewDirProject(
	dir string,
	opts ...ProjectOption,
) (*Project, error) {
	p := &Project{}

	for _, opt := range opts {
		opt(p)
	}

	if _, err := os.Stat(dir); err != nil {
		return nil, errors.WithStack(err)
	}

	p.fs = osfs.New(dir)

	return p, nil
}

func NewFileProject(
	path string,
	opts ...ProjectOption,
) (*Project, error) {
	p := &Project{}

	for _, opt := range opts {
		opt(p)
	}

	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if _, err := os.Stat(path); err != nil {
		return nil, errors.WithStack(err)
	}

	p.filePath = path

	return p, nil
}

func (p *Project) Load(
	ctx context.Context,
	eventc chan<- LoadEvent,
	onlyFiles bool,
) {
	defer close(eventc)

	switch {
	case p.repo != nil:
		// Git-based project.
		// TODO: confirm if the order of appending to ignorePatterns is important.
		ignorePatterns := []gitignore.Pattern{
			// Ignore .git by default.
			gitignore.ParsePattern(".git", nil),
			gitignore.ParsePattern(".git/*", nil),
		}

		if p.respectGitignore {
			patterns, err := gitignore.ReadPatterns(p.fs, nil)
			if err != nil {
				eventc <- LoadEvent{
					Type: LoadEventError,
					Data: errors.WithStack(err),
				}
			}
			ignorePatterns = append(ignorePatterns, patterns...)
		}

		for _, p := range p.ignoreFilePatterns {
			ignorePatterns = append(ignorePatterns, gitignore.ParsePattern(p, nil))
		}

		p.loadFromDirectory(ctx, eventc, ignorePatterns, onlyFiles)
	case p.fs != nil:
		// Dir-based project.
		ignorePatterns := make([]gitignore.Pattern, 0, len(p.ignoreFilePatterns))

		// It's allowed for a dir-based project to read
		// .gitignore and interpret it.
		if p.respectGitignore {
			patterns, err := gitignore.ReadPatterns(p.fs, nil)
			if err != nil {
				eventc <- LoadEvent{
					Type: LoadEventError,
					Data: errors.WithStack(err),
				}
			}
			ignorePatterns = append(ignorePatterns, patterns...)
		}

		for _, p := range p.ignoreFilePatterns {
			ignorePatterns = append(ignorePatterns, gitignore.ParsePattern(p, nil))
		}

		p.loadFromDirectory(ctx, eventc, ignorePatterns, onlyFiles)
	case p.filePath != "":
		p.loadFromFile(ctx, eventc, p.filePath, onlyFiles)
	default:
		eventc <- LoadEvent{
			Type: LoadEventError,
			Data: errors.New("invariant violation: Project struct initialized incorrectly"),
		}
	}
}

func (p *Project) loadFromDirectory(
	ctx context.Context,
	eventc chan<- LoadEvent,
	ignorePatterns []gitignore.Pattern,
	onlyFiles bool,
) {
	filesToSearchBlocks := make([]string, 0)
	onFileFound := func(path string) {
		if !onlyFiles {
			filesToSearchBlocks = append(filesToSearchBlocks, path)
		}
	}

	ignoreMatcher := gitignore.NewMatcher(ignorePatterns)

	eventc <- LoadEvent{Type: LoadEventStartedWalk}

	err := util.Walk(p.fs, ".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ignored := ignoreMatcher.Match(
			[]string{path},
			info.IsDir(),
		)
		if !ignored {
			absPath := p.fs.Join(p.fs.Root(), path)

			if info.IsDir() {
				eventc <- LoadEvent{
					Type: LoadEventFoundDir,
					Data: absPath,
				}
			} else if isMarkdown(path) {
				eventc <- LoadEvent{
					Type: LoadEventFoundFile,
					Data: absPath,
				}

				onFileFound(absPath)
			}
		} else if info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		eventc <- LoadEvent{
			Type: LoadEventError,
			Data: err,
		}
	}

	eventc <- LoadEvent{
		Type: LoadEventFinishedWalk,
	}

	if len(filesToSearchBlocks) == 0 {
		return
	}

	for _, file := range filesToSearchBlocks {
		extractTasksFromFile(ctx, eventc, file)
	}
}

func (p *Project) loadFromFile(
	ctx context.Context,
	eventc chan<- LoadEvent,
	path string,
	onlyFiles bool,
) {
	eventc <- LoadEvent{Type: LoadEventStartedWalk}

	eventc <- LoadEvent{
		Type: LoadEventFoundFile,
		Data: path,
	}

	eventc <- LoadEvent{
		Type: LoadEventFinishedWalk,
	}

	if onlyFiles {
		return
	}

	extractTasksFromFile(ctx, eventc, path)
}

func extractTasksFromFile(
	ctx context.Context,
	eventc chan<- LoadEvent,
	path string,
) {
	eventc <- LoadEvent{
		Type: LoadEventStartedParsingDocument,
		Data: path,
	}

	codeBlocks, err := getCodeBlocksFromFile(path)

	eventc <- LoadEvent{
		Type: LoadEventFinishedParsingDocument,
		Data: path,
	}

	if err != nil {
		eventc <- LoadEvent{
			Type: LoadEventError,
			Data: err,
		}
	}

	for _, b := range codeBlocks {
		eventc <- LoadEvent{
			Type: LoadEventFoundTask,
			Data: Task{
				Filename:  path,
				CodeBlock: b,
			},
		}
	}
}

func getCodeBlocksFromFile(path string) (document.CodeBlocks, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return getCodeBlocks(data)
}

func getCodeBlocks(data []byte) (document.CodeBlocks, error) {
	d := document.New(data)
	node, err := d.Root()
	if err != nil {
		return nil, err
	}
	return document.CollectCodeBlocks(node), nil
}

func isMarkdown(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".md" || ext == ".mdx" || ext == ".mdi" || ext == ".mdr" || ext == ".run" || ext == ".runme"
}
