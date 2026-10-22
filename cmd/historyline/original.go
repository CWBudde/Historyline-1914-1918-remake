package main

import (
	"embed"
	"fmt"
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
// level). An original map's level is its campaign level (levels/campaign),
// which the original rules draw and the remake rules play.
func loadMap(dir, name string, set *data.Set) (*level.Level, *dosbind.Board, error) {
	stem, ok := strings.CutPrefix(name, originalPrefix)
	if !ok {
		lvl, err := loadLevel(dir, name)
		return lvl, nil, err
	}
	_, lvl, b, err := loadOriginal(dir, stem, set)
	return lvl, b, err
}

// loadOriginal loads the imported original map stem ("05") and its campaign
// level from dir where present, else the embedded copies: the map, the level
// and the map's board.
func loadOriginal(dir, stem string, set *data.Set) (*finmap.Level, *level.Level, *dosbind.Board, error) {
	name := originalPrefix + stem
	raw, err := readLevelFile(dir, levels.Original, "original/full/"+stem+".json")
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
	file, err := origview.FileName(l)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	if raw, err = readLevelFile(dir, levels.Campaign, "campaign/"+file+".json"); err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	lvl, err := level.Parse(raw)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %s: %w", name, file, err)
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
