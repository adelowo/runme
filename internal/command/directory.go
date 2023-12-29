package command

import (
	"os"
	"path/filepath"

	"github.com/stateful/runme/internal/document"
)

func resolveDirFromCodeBlock(block *document.CodeBlock, parentDir string) string {
	var dirs []string

	// The order is immportant starting from a frontmatter dir,
	// next is the block dir, and finally all other dirs.

	// TODO(adamb): consider handling this error or add a comment it can be skipped.
	fmtr, err := block.Document().Frontmatter()
	if err == nil && fmtr != nil && fmtr.Cwd != "" {
		dirs = append(dirs, fmtr.Cwd)
	}

	dirs = append(dirs, block.Cwd())

	for _, dir := range dirs {
		dir := filepath.FromSlash(dir)
		newDir := resolveDirUsingParentAndChild(parentDir, dir)
		if stat, err := os.Stat(newDir); err == nil && stat.IsDir() {
			parentDir = newDir
		}
	}

	return parentDir
}

// TODO(adamb): figure out if it's needed and for which cases.
func resolveDirUsingParentAndChild(parent, child string) string {
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
