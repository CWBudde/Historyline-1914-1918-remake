package main

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// testLevel is a shipped level: a remake level of levels.FS
// ("level2_cp.json") or a setup of a shipped scenario
// ("monthly/1916-02-parts:cp").
func testLevel(t *testing.T, name string) *level.Level {
	t.Helper()
	if file, mode, ok := strings.Cut(name, ":"); ok {
		l, err := scenario.ShippedLevel(file, scenario.Mode(mode))
		if err != nil {
			t.Fatal(err)
		}
		return l
	}
	raw, err := levels.FS.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	l, err := level.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
