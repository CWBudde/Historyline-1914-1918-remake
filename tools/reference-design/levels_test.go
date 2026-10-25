package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// referenceLevel is the PULSE reference design's map (art/reference/pulse).
const referenceLevel = "reference"

// testLevel is a setup of a shipped scenario ("monthly/1916-02:cp")
// or the reference design's map (referenceLevel).
func testLevel(t *testing.T, name string) *level.Level {
	t.Helper()
	if name == referenceLevel {
		l, err := level.Load(filepath.Join("..", "..", "art", "reference", "pulse", "level.json"))
		if err != nil {
			t.Fatal(err)
		}
		return l
	}
	file, mode, _ := strings.Cut(name, ":")
	l, err := scenario.ShippedLevel(file, scenario.Mode(mode))
	if err != nil {
		t.Fatal(err)
	}
	return l
}
