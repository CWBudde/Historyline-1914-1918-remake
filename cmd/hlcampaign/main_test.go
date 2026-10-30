package main

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func TestMonth(t *testing.T) {
	for k, want := range map[int]string{0: "1914-08", 1: "1914-10", 2: "1914-12", 3: "1915-02", 23: "1918-06"} {
		if got := Month(k); got != want {
			t.Errorf("Month(%d) = %s, want %s", k, got, want)
		}
	}
}

// originalMaps are the full game's maps of the local original files; the
// test is skipped without them.
func originalMaps(t *testing.T) []*finmap.Level {
	t.Helper()
	maps, err := origmaps.Full()
	if errors.Is(err, origmaps.ErrNotFound) {
		t.Skip("no original files")
	}
	if err != nil {
		t.Fatal(err)
	}
	return maps
}

// The shipped levels began as what hlcampaign writes from the original
// maps and are hand-editable, so this test does not require byte identity
// with the generator: every generated folder is shipped, and the extra
// maps are the generated ones.
func TestScenariosCurrent(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	out, err := generate(originalMaps(t), set)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.folders) != 48 {
		t.Fatalf("%d level folders, want 48", len(out.folders))
	}
	shipped, err := scenario.Folders(levels.Levels)
	if err != nil {
		t.Fatal(err)
	}
	for name := range out.folders {
		if !slices.Contains(shipped, name) {
			t.Errorf("%s is missing: run just levels-campaign", name)
		}
	}
	for _, name := range shipped {
		if strings.HasPrefix(name, scenario.ExtraDir+"/") && out.folders[name] == nil {
			t.Errorf("%s is no extra map the import writes", name)
		}
	}
}

// The import writes the original campaign maps' months and leaves the
// newly authored ones alone.
func TestLeavesTheAuthoredMonths(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	out, err := generate(originalMaps(t), set)
	if err != nil {
		t.Fatal(err)
	}
	for month, want := range map[string]bool{"1914-08": true, "1914-09": false, "1914-11": false, "1915-01": false, "1918-06": true} {
		if got := out.folders[month] != nil; got != want {
			t.Errorf("%s written %v, want %v", month, got, want)
		}
	}
}
