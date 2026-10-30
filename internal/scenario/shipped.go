package scenario

import (
	"fmt"
	"io/fs"
	"path"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Named is a setup of a shipped level folder as a level.
type Named struct {
	Name  string // "1914-08:cp", "extra/13-grand:2p"
	Level *level.Level
}

// ExtraDir holds the extra maps' folders below the levels.
const ExtraDir = "extra"

// Folders are the level folders of fsys (levels.Levels, or the levels
// directory on disk): the campaign months in calendar order, then the
// extra maps in name order.
func Folders(fsys fs.FS) ([]string, error) {
	var dirs []string
	for _, pattern := range []string{"19*/" + MapFile, ExtraDir + "/*/" + MapFile} {
		maps, err := fs.Glob(fsys, pattern)
		if err != nil {
			return nil, err
		}
		slices.Sort(maps)
		for _, m := range maps {
			dirs = append(dirs, path.Dir(m))
		}
	}
	return dirs, nil
}

// Months are the campaign months' folders of fsys in calendar order.
func Months(fsys fs.FS) ([]string, error) {
	dirs, err := Folders(fsys)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(dirs, func(d string) bool { return !monthRE.MatchString(d) }), nil
}

// Shipped are the levels of the shipped level folders (Folders of
// levels.Levels): every setup of the modes given, in that order, or of
// every mode without any. Shipped(CP) is every map once (an extra map's cp
// setup is its two-player units).
func Shipped(modes ...Mode) ([]Named, error) {
	if len(modes) == 0 {
		modes = Modes
	}
	dirs, err := Folders(levels.Levels)
	if err != nil {
		return nil, err
	}
	var out []Named
	for _, dir := range dirs {
		f, err := Read(levels.Levels, dir)
		if err != nil {
			return nil, err
		}
		for _, m := range modes {
			if f.Setups[m] == nil {
				continue
			}
			lv, err := f.Level(m)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", dir, err)
			}
			out = append(out, Named{Name: dir + ":" + string(m), Level: lv})
		}
	}
	return out, nil
}

// ShippedLevel is the level of setup m of the shipped level folder name
// ("1916-02", "extra/15-water").
func ShippedLevel(name string, m Mode) (*level.Level, error) {
	f, err := Read(levels.Levels, name)
	if err != nil {
		return nil, err
	}
	lv, err := f.Level(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return lv, nil
}
