package main

import (
	"bytes"
	"io/fs"
	"path"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// levels/campaign is what hlcampaign writes from levels/original: every map,
// no stale file, no hand edit.
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
		if w, ok := want[path.Base(n)]; !ok {
			t.Errorf("%s is not generated", n)
		} else if !bytes.Equal(b, w) {
			t.Errorf("%s is stale: run just levels-campaign", n)
		}
	}
	for n := range want {
		if !slices.Contains(got, n) {
			t.Errorf("campaign/%s is missing: run just levels-campaign", n)
		}
	}
}
