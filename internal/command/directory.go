package command

import (
	"os"
	"path/filepath"
)

func resolveDir(parentDir string, candidates []string) string {
	for _, dir := range candidates {
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
