package artreview

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// sourceRoot is the -root flag: the directory manifest source paths are relative
// to. Empty selects sourceBase's default.
var sourceRoot string

const repoModule = "github.com/CWBudde/Historyline-1914-1918-remake"

// sourceBase returns the directory the source paths of the manifest at path
// resolve against: -root when given, else the repository root above the
// manifest (the directory whose go.mod declares this module), else the
// manifest's own directory.
func sourceBase(path string) string {
	if sourceRoot != "" {
		return sourceRoot
	}
	if abs, err := filepath.Abs(path); err == nil {
		for dir := filepath.Dir(abs); ; dir = filepath.Dir(dir) {
			if isRepoRoot(dir) {
				return dir
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return filepath.Dir(path)
}

func isRepoRoot(dir string) (found bool) {
	f, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			found = false
		}
	}()
	s := bufio.NewScanner(f)
	return s.Scan() && strings.TrimSpace(s.Text()) == "module "+repoModule
}
