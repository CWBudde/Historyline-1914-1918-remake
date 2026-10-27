package main

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
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
// (origview.Campaign), which the battle draws and plays.
func loadMap(dir, name string, set *data.Set) (*level.Level, *dosbind.Board, error) {
	if stem, ok := strings.CutPrefix(name, originalPrefix); ok {
		_, lvl, b, err := loadOriginal(stem, set)
		return lvl, b, err
	}
	lvl, err := loadLevel(dir, name)
	return lvl, nil, err
}

// scenarioDirs are the -level prefixes of the scenario files.
var scenarioDirs = map[string]embed.FS{"monthly/": levels.Monthly, "skirmish/": levels.Skirmish}

// loadScenario loads -level name "dir/file:mode" from dir where present,
// else the embedded copy.
func loadScenario(dir string, emb embed.FS, name string) (*level.Level, error) {
	file, mode, _ := strings.Cut(name, ":")
	raw, err := readLevelFile(dir, emb, file+".json")
	if err != nil {
		return nil, err
	}
	f, err := scenario.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
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

// scenarioLevel is the level of setup m of the scenario file f, which is
// file ("monthly/1914-08") below the levels directory; its backdrop
// becomes a path below the levels directory too (levelBackdrop).
func scenarioLevel(f *scenario.File, file string, m scenario.Mode) (*level.Level, error) {
	lvl, err := f.Level(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if b := lvl.Extra.Backdrop; b != "" {
		lvl.Extra.Backdrop = path.Join(path.Dir(file), b)
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

// readLevelFile reads file (slash-separated) below dir, else from emb. It
// falls back on the embedded copy for any error: the web build has no file
// system and fails with "not implemented", not fs.ErrNotExist.
func readLevelFile(dir string, emb embed.FS, file string) ([]byte, error) {
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file)))
	if err != nil {
		if b, embErr := emb.ReadFile(file); embErr == nil {
			return b, nil
		}
	}
	return raw, err
}
