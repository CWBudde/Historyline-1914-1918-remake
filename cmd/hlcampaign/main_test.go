package main

import (
	"errors"
	"fmt"
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
// with the generator: every generated folder is shipped.
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
	shipped, err := scenario.Months(levels.Levels)
	if err != nil {
		t.Fatal(err)
	}
	for name := range out.folders {
		if !slices.Contains(shipped, name) {
			t.Errorf("%s is missing: run just levels-campaign", name)
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
	for month, want := range map[string]bool{"1914-08": true, "1914-09": false, "1914-11": false, "1915-01": false, "1915-03": true, "1918-06": true, "1918-07": true, "1918-11": true} {
		if got := out.folders[month] != nil; got != want {
			t.Errorf("%s written %v, want %v", month, got, want)
		}
	}
}

// The two-player maps fill the months from March 1915 to the armistice
// that no original campaign map takes, each once and September 1918 for
// October too, each Entente setup with a code word of its own.
func TestTwoPlayerMonths(t *testing.T) {
	months, codes := map[string]bool{}, map[string]bool{}
	for _, tp := range twoPlayer {
		if months[tp.month] || codes[tp.ecCode] || len(tp.ecCode) != 5 || strings.ToUpper(tp.ecCode) != tp.ecCode {
			t.Errorf("%s %s: month or code twice, or not a code word", tp.month, tp.ecCode)
		}
		months[tp.month], codes[tp.ecCode] = true, true
	}
	var want []string
	for m := 14; m <= 54; m += 2 { // months after January 1914: March 1915 to July 1918
		want = append(want, fmt.Sprintf("%d-%02d", 1914+m/12, m%12+1))
	}
	want = append(want, "1918-08", "1918-09", "1918-11") // September for October too
	if len(months) != len(want) || want[0] != "1915-03" {
		t.Fatalf("months %v, want %v", months, want)
	}
	for _, m := range want {
		if !months[m] {
			t.Errorf("%s is not filled", m)
		}
	}
}
