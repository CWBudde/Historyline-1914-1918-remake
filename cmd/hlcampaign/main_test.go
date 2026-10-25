package main

import (
	"io/fs"
	"path"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
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

// The shipped scenarios began as what hlcampaign writes from
// levels/original and are hand-editable, so this test does not require
// byte identity with the generator: every generated file is shipped under
// its name, and every shipped file parses.
func TestScenariosCurrent(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	src, err := fs.Sub(levels.Original, "original")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generate(src, set)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.monthly) != 24 || len(out.skirmish) != 24 {
		t.Fatalf("%d months, %d skirmish maps, want 24 each", len(out.monthly), len(out.skirmish))
	}
	for _, set := range []struct {
		fsys  fs.FS
		dir   string
		files map[string][]byte
	}{{levels.Monthly, "monthly", out.monthly}, {levels.Skirmish, "skirmish", out.skirmish}} {
		names, err := fs.Glob(set.fsys, set.dir+"/*.json")
		if err != nil {
			t.Fatal(err)
		}
		shipped := map[string]bool{}
		for _, n := range names {
			shipped[path.Base(n)] = true
			b, err := fs.ReadFile(set.fsys, n)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := scenario.Parse(b); err != nil {
				t.Errorf("%s does not parse: %v", n, err)
			}
		}
		for n := range set.files {
			if !shipped[n] {
				t.Errorf("%s/%s is missing: run just levels-campaign", set.dir, n)
			}
		}
		if set.dir == "skirmish" && len(names) != len(set.files) {
			t.Errorf("%d skirmish files, the import writes %d", len(names), len(set.files))
		}
	}
}

// The months the import writes are every second month from August 1914;
// the import leaves the others, the newly authored levels, alone.
func TestMonthsLeaveTheSecondMonths(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	src, _ := fs.Sub(levels.Original, "original")
	out, err := generate(src, set)
	if err != nil {
		t.Fatal(err)
	}
	if !out.months["1914-08"] || out.months["1914-09"] || !out.months["1918-06"] || len(out.months) != 24 {
		t.Errorf("months %v", out.months)
	}
}
