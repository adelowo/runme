package project

import (
	"bufio"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const (
	commentPrefix   = "#"
	gitDir          = ".git"
	gitignoreFile   = ".gitignore"
	infoExcludeFile = "info/exclude"
)

func replaceTildeWithHome(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		firstSlash := strings.Index(path, "/")
		if firstSlash == 1 {
			home, err := os.UserHomeDir()
			if err != nil {
				return path, err
			}
			return strings.Replace(path, "~", home, 1), nil
		} else if firstSlash > 1 {
			username := path[1:firstSlash]
			userAccount, err := user.Lookup(username)
			if err != nil {
				return path, err
			}
			return strings.Replace(path, path[:firstSlash], userAccount.HomeDir, 1), nil
		}
	}

	return path, nil
}

func readIgnoreFile(fs billy.Filesystem, ignoreFile string) (ps []gitignore.Pattern, err error) {
	ignoreFile, _ = replaceTildeWithHome(ignoreFile)

	f, err := fs.Open(ignoreFile)
	if err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			s := scanner.Text()
			if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
				ps = append(ps, gitignore.ParsePattern(s, filepath.SplitList(ignoreFile)))
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}

func ReadPatterns(bfs billy.Filesystem) (ps []gitignore.Pattern, err error) {
	err = util.Walk(bfs, ".", func(path string, info fs.FileInfo, err error) error {
		if info.Name() == gitignoreFile {
			subps, _ := readIgnoreFile(bfs, path)
			ps = append(ps, subps...)
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		if info.Name() == gitDir {
			subps, _ := readIgnoreFile(bfs, filepath.Join(path, infoExcludeFile))
			ps = append(ps, subps...)
			return filepath.SkipDir
		}

		return nil
	})

	return
}
