package unitpanel

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const repoModule = "github.com/CWBudde/Historyline-1914-1918-remake"

// sourceBase returns the directory the source paths of the manifest at path
// resolve against: the repository root above the manifest (the directory whose
// go.mod declares this module), else the manifest's own directory.
func sourceBase(path string) string {
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

func isRepoRoot(dir string) bool {
	f, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	return s.Scan() && strings.TrimSpace(s.Text()) == "module "+repoModule
}
