package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// originalPrefix names a map of the local original files in -level, like
// "original/05" (internal/origmaps).
const originalPrefix = "original/"

// loadMap loads -level name (loadLevel), or a map of the local original
// files with its board bound exactly to the original records (nil
// otherwise). An original map's level is its remake level
// (origview.Campaign), which the original rules draw and the remake rules
// play.
func loadMap(dir, name string, set *data.Set) (*level.Level, *dosbind.Board, error) {
	if stem, ok := strings.CutPrefix(name, originalPrefix); ok {
		_, lvl, b, err := loadOriginal(stem, set)
		return lvl, b, err
	}
	lvl, err := loadLevel(dir, name)
	return lvl, nil, err
}

// levelFiles is the levels directory on disk over the embedded levels: a
// file below the directory where it opens, else its embedded copy. Any
// error falls back on the embedded copy: the web build has no file system
// and fails with "not implemented", not fs.ErrNotExist.
type levelFiles struct{ disk fs.FS }

// levelsFS is the levels directory dir over the embedded levels.
func levelsFS(dir string) fs.FS {
	if dir == "" {
		dir = "."
	}
	return levelFiles{os.DirFS(dir)}
}

func (l levelFiles) Open(name string) (fs.File, error) {
	f, err := l.disk.Open(name)
	if err != nil {
		if e, embErr := levels.Levels.Open(name); embErr == nil {
			return e, nil
		}
	}
	return f, err
}

// scenarioName reports whether -level name is a setup of a level folder,
// a campaign month ("1914-08:ec").
func scenarioName(name string) bool {
	file, _, _ := strings.Cut(name, ":")
	return monthName.MatchString(file)
}

var monthName = regexp.MustCompile(`^19\d\d-\d\d$`)

// loadScenario loads -level name "folder:mode" from the levels directory
// dir where present, else the embedded copy; without a mode, the folder's
// first setup.
func loadScenario(dir, name string) (*level.Level, error) {
	file, mode, _ := strings.Cut(name, ":")
	f, err := scenario.Read(levelsFS(dir), file)
	if err != nil {
		return nil, err
	}
	m := scenario.Mode(mode)
	if mode == "" {
		for _, m = range scenario.Modes {
			if f.Setups[m] != nil {
				break
			}
		}
	}
	return scenarioLevel(f, file, m)
}

// scenarioLevel is the level of setup m of the level folder f, which is
// file ("1914-08") below the levels directory; its backdrop becomes a path
// below the levels directory too (levelBackdrop).
func scenarioLevel(f *scenario.File, file string, m scenario.Mode) (*level.Level, error) {
	lvl, err := f.Level(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if b := lvl.Extra.Backdrop; b != "" {
		lvl.Extra.Backdrop = path.Join(file, b)
	}
	return lvl, nil
}

// loadOriginal loads the original full game's map stem ("05") from the
// local original files: the map, its remake level and its board.
func loadOriginal(stem string, set *data.Set) (*finmap.Level, *level.Level, *dosbind.Board, error) {
	name := originalPrefix + stem
	l, err := origmaps.Map(stem)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	b, err := dosbind.FromOriginal(l, set)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	lvl, err := origview.Campaign(l, set)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	return l, lvl, b, nil
}
