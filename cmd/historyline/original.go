package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// originalPrefix names an imported original map in -level, like
// "original/05" for levels/original/full/05.json.
const originalPrefix = "original/"

// loadMap loads -level name: a shipped level, or an imported original map
// with its board bound exactly to the original records (nil for a shipped
// level). An original map is drawn from its view level (internal/origview)
// and plays on the original rules only.
func loadMap(dir, name string, set *data.Set) (*level.Level, *dosbind.Board, error) {
	stem, ok := strings.CutPrefix(name, originalPrefix)
	if !ok {
		lvl, err := loadLevel(dir, name)
		return lvl, nil, err
	}
	_, lvl, b, err := loadOriginal(dir, stem, set)
	return lvl, b, err
}

// loadOriginal loads the imported original map stem ("05") from dir where
// present, else the embedded copy: the map, its view level and its board.
func loadOriginal(dir, stem string, set *data.Set) (*finmap.Level, *level.Level, *dosbind.Board, error) {
	name := originalPrefix + stem
	file := "original/full/" + stem + ".json"
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(file)))
	if errors.Is(err, fs.ErrNotExist) {
		raw, err = levels.Original.ReadFile(file)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	l, err := finmap.Parse(raw)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	b, err := dosbind.FromOriginal(l, set)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	lvl, err := origview.Level(l)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	return l, lvl, b, nil
}
