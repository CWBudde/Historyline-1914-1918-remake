// Package finextract writes every map of the local original DOS full game
// and demo (internal/origmaps) as the versioned level JSON of internal/finmap,
// one file per map plus index.json, for a look at the decoded maps.
//
//	go run ./cmd/hl levels extract [-original dir] [-out build/scratch/levels] [-build full,demo]
package finextract

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
)

// repoRoot is the directory of go.mod from dir upwards.
func repoRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.mod above the working directory")
		}
		dir = parent
	}
}

// selectBuilds returns the builds named in the comma-separated list.
func selectBuilds(all []origmaps.Build, names string) ([]origmaps.Build, error) {
	var out []origmaps.Build
	for _, name := range strings.Split(names, ",") {
		i := slices.IndexFunc(all, func(b origmaps.Build) bool { return b.Name == name })
		switch {
		case i < 0:
			return nil, fmt.Errorf("unknown build %q", name)
		case slices.ContainsFunc(out, func(b origmaps.Build) bool { return b.Name == name }):
			return nil, fmt.Errorf("build %q named twice", name)
		}
		out = append(out, all[i])
	}
	return out, nil
}

// code is the level's password for the log, "" without a mission.
func code(l *finmap.Level) string {
	if l.Mission == nil {
		return ""
	}
	return l.Mission.Code
}

// indexEntry summarises one written level.
type indexEntry struct {
	Build     string `json:"build"`
	Map       string `json:"map"`
	File      string `json:"file"`
	Players   int    `json:"players,omitempty"`
	Winter    bool   `json:"winter,omitempty"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Units     int    `json:"units"`
	Buildings int    `json:"buildings"`
}

// extract writes the levels of every build to out/<build>/<map>.json and
// out/index.json, and returns how many it wrote.
func extract(builds []origmaps.Build, out string, logw io.Writer) (int, error) {
	var index []indexEntry
	for _, b := range builds {
		ls, err := origmaps.Read(b)
		if err != nil {
			return 0, err
		}
		if err := os.MkdirAll(filepath.Join(out, b.Name), 0o755); err != nil {
			return 0, err
		}
		for _, l := range ls {
			raw, err := json.Marshal(l)
			if err != nil {
				return 0, err
			}
			file := filepath.Join(b.Name, l.Source.Map+".json")
			if err := os.WriteFile(filepath.Join(out, file), append(raw, '\n'), 0o644); err != nil {
				return 0, err
			}
			e := indexEntry{
				Build: b.Name, Map: l.Source.Map, File: filepath.ToSlash(file), Width: l.Width, Height: l.Height,
				Units: len(l.Units), Buildings: len(l.Buildings),
			}
			if m := l.Mission; m != nil {
				e.Players, e.Winter = m.Players, m.Winter
			}
			index = append(index, e)
			if _, err := fmt.Fprintf(logw, "%s %-7s %2dx%-2d %5s units=%d buildings=%d\n", b.Name, l.Source.Map, l.Width, l.Height, code(l), e.Units, e.Buildings); err != nil {
				return 0, err
			}
		}
	}
	raw, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(filepath.Join(out, "index.json"), append(raw, '\n'), 0o644); err != nil {
		return 0, err
	}
	return len(index), nil
}
