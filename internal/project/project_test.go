package project

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

func TestNewGitProject(t *testing.T) {
	testdataDir := testdataDir()
	gitProjectDir := filepath.Join(testdataDir, "git-project")
	projectDir := filepath.Join(testdataDir, "dir-project")

	t.Run("NotGitProject", func(t *testing.T) {
		_, err := NewGitProject(projectDir)
		require.ErrorIs(t, err, git.ErrRepositoryNotExists)
	})

	t.Run("UnknownDir", func(t *testing.T) {
		unknownDir := filepath.Join(testdataDir, "unknown-project")
		_, err := NewGitProject(unknownDir)
		require.ErrorIs(t, err, git.ErrRepositoryNotExists)
	})

	t.Run("ProperGitProject", func(t *testing.T) {
		_, err := NewGitProject(gitProjectDir)
		require.NoError(t, err)
	})
}

func TestNewDirProject(t *testing.T) {
	testdataDir := testdataDir()
	gitProjectDir := filepath.Join(testdataDir, "git-project")
	projectDir := filepath.Join(testdataDir, "dir-project")

	t.Run("ProperDirProject", func(t *testing.T) {
		_, err := NewDirProject(projectDir)
		require.NoError(t, err)
	})

	t.Run("ProperGitProject", func(t *testing.T) {
		_, err := NewDirProject(gitProjectDir)
		require.NoError(t, err)
	})

	t.Run("UnknownDir", func(t *testing.T) {
		unknownDir := filepath.Join(testdataDir, "unknown-project")
		_, err := NewDirProject(unknownDir)
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestNewFileProject(t *testing.T) {
	testdataDir := testdataDir()

	t.Run("UnknownFile", func(t *testing.T) {
		fileProject := filepath.Join(testdataDir, "unknown-file.md")
		_, err := NewFileProject(fileProject)
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("ProperFileProject", func(t *testing.T) {
		fileProject := filepath.Join(testdataDir, "file-project.md")
		_, err := NewFileProject(fileProject)
		require.NoError(t, err)
	})
}

func TestProjectLoad(t *testing.T) {
	testdataDir := testdataDir()
	gitProjectDir := filepath.Join(testdataDir, "git-project")

	t.Run("GitProject", func(t *testing.T) {
		p, err := NewGitProject(gitProjectDir, WithRespectGitignore())
		require.NoError(t, err)

		eventc := make(chan LoadEvent)

		events := make([]LoadEvent, 0)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for e := range eventc {
				events = append(events, e)
			}
		}()

		p.Load(context.Background(), eventc, false)
		<-done

		require.NotEmpty(t, events)

		expectedEvents := []LoadEvent{
			{Type: LoadEventStartedWalk},
			{Type: LoadEventFinishedWalk},
		}
		require.EqualValues(t, expectedEvents, events)
	})
}

// TODO(adamb): a better approach is to store "testdata" during build time.
func testdataDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(
		filepath.Dir(b),
		"testdata",
	)
}
