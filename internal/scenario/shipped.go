package scenario

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Named is a setup of a shipped scenario file as a level.
type Named struct {
	Name  string // "monthly/1914-08:cp"
	Level *level.Level
}

// shippedDirs are the embedded scenario sets in order.
var shippedDirs = []struct {
	dir string
	fs  embed.FS
}{{"monthly", levels.Monthly}, {"skirmish", levels.Skirmish}}

// Shipped are the levels of the shipped scenario files (levels.Monthly,
// then levels.Skirmish, each in name order): every setup of the modes
// given, in that order, or of every mode without any. Shipped(CP) is every
// map once (a skirmish map's cp setup is its two-player units).
func Shipped(modes ...Mode) ([]Named, error) {
	if len(modes) == 0 {
		modes = Modes
	}
	var out []Named
	for _, d := range shippedDirs {
		names, err := fs.Glob(d.fs, d.dir+"/*.json")
		if err != nil {
			return nil, err
		}
		slices.Sort(names)
		for _, name := range names {
			f, err := shippedFile(d.fs, name)
			if err != nil {
				return nil, err
			}
			for _, m := range modes {
				if f.Setups[m] == nil {
					continue
				}
				lv, err := f.Level(m)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", name, err)
				}
				out = append(out, Named{Name: strings.TrimSuffix(name, ".json") + ":" + string(m), Level: lv})
			}
		}
	}
	return out, nil
}

// ShippedLevel is the level of setup m of the shipped scenario file name
// ("monthly/1916-02", "skirmish/11-uboot").
func ShippedLevel(name string, m Mode) (*level.Level, error) {
	for _, d := range shippedDirs {
		if strings.HasPrefix(name, d.dir+"/") {
			f, err := shippedFile(d.fs, name+".json")
			if err != nil {
				return nil, err
			}
			lv, err := f.Level(m)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			return lv, nil
		}
	}
	return nil, fmt.Errorf("scenario: %q is in no shipped set", name)
}

func shippedFile(fsys fs.FS, name string) (*File, error) {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}
	f, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return f, nil
}
