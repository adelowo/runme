package project

import (
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stateful/runme/internal/project/testdata"
)

func TestReadPatterns(t *testing.T) {
	gitProjectDir := testdata.GitProjectPath()
	fs := osfs.New(gitProjectDir)
	patterns, err := ReadPatterns(fs)
	require.NoError(t, err)
	assert.Len(t, patterns, 2)
}
