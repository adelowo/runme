package command

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

func resolveDirectory(blockDir string) (string, error) {
	if blockDir != "" {
		return blockDir, nil
	}
	directory, err := os.Getwd()
	return directory, errors.WithStack(err)
}

func resolveDirectoryFromCodeBlock(block *document.CodeBlock, parentDir string, otherDirs ...string) string {
	dirs := []string{
		block.Cwd(),
	}

	// TODO(adamb): consider handling this error or add a comment it can be skipped.
	fmtr, _ := block.Document().Frontmatter()
	if fmtr != nil && fmtr.Cwd != "" {
		dirs = append(dirs, fmtr.Cwd)
	}

	dirs = append(dirs, otherDirs...)

	for _, dir := range dirs {
		dir := filepath.FromSlash(dir)
		newDir := resolveOrAbsolute(parentDir, dir)
		if stat, err := os.Stat(newDir); err == nil && stat.IsDir() {
			parentDir = newDir
		}
	}

	return parentDir
}

func resolveOrAbsolute(parent string, child string) string {
	if child == "" {
		return parent
	}

	if filepath.IsAbs(child) {
		return child
	}

	if parent != "" {
		return filepath.Join(parent, child)
	}

	return child
}
