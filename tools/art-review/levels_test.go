package main

import (
	"io/fs"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// shippedLevels are the shipped levels: the remake levels of levels.FS,
// then every shipped scenario map once (scenario.Shipped(scenario.CP)).
func shippedLevels() ([]scenario.Named, error) {
	names, err := fs.Glob(levels.FS, "*.json")
	if err != nil {
		return nil, err
	}
	var out []scenario.Named
	for _, name := range names {
		b, err := fs.ReadFile(levels.FS, name)
		if err != nil {
			return nil, err
		}
		l, err := level.Parse(b)
		if err != nil {
			return nil, err
		}
		out = append(out, scenario.Named{Name: name, Level: l})
	}
	sh, err := scenario.Shipped(scenario.CP)
	if err != nil {
		return nil, err
	}
	return append(out, sh...), nil
}
