package main

import (
	"io/fs"
	"path"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// levels/campaign is the shipped campaign: it began as what hlcampaign
// writes from levels/original and is hand-editable since the merge, so
// this test no longer requires byte identity with the generator — only
// that every generated map is present and every shipped file parses.
func TestCampaignCurrent(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	src, err := fs.Sub(levels.Original, "original")
	if err != nil {
		t.Fatal(err)
	}
	want, err := generate(src, set)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 72 {
		t.Errorf("%d campaign levels, want 72", len(want))
	}
	names, err := fs.Glob(levels.Campaign, "campaign/*.json")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, n := range names {
		got = append(got, path.Base(n))
		b, err := fs.ReadFile(levels.Campaign, n)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := want[path.Base(n)]; !ok {
			t.Errorf("%s is not a campaign map", n)
		}
		if _, err := level.Parse(b); err != nil {
			t.Errorf("%s does not parse: %v", n, err)
		}
	}
	for n := range want {
		if !slices.Contains(got, n) {
			t.Errorf("campaign/%s is missing: run just levels-campaign", n)
		}
	}
}
