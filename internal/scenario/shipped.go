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
	Name  string // "1914-08:cp"
	Level *level.Level
}

// Months are the level folders of fsys (levels.Levels, or the levels
// directory on disk), the campaign months, in calendar order.
func Months(fsys fs.FS) ([]string, error) {
	maps, err := fs.Glob(fsys, "19*/"+MapFile)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, m := range maps {
		if dir := path.Dir(m); monthRE.MatchString(dir) {
			dirs = append(dirs, dir)
		}
	}
	slices.Sort(dirs)
	return dirs, nil
}

// Shipped are the levels of the shipped level folders (Months of
// levels.Levels): every setup of the modes given, in that order, or of
// every mode without any. Shipped(CP) is every map once.
func Shipped(modes ...Mode) ([]Named, error) {
	if len(modes) == 0 {
		modes = Modes
	}
	dirs, err := Months(levels.Levels)
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
// ("1916-02").
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
