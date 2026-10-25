package dosbind_test

import (
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Every setup of the shipped scenarios (levels/monthly, levels/skirmish)
// must bind for -rules dos.
func TestBindScenarios(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	months, _ := filepath.Glob("../../levels/monthly/*.json")
	skirmish, _ := filepath.Glob("../../levels/skirmish/*.json")
	if len(months) < 25 || len(skirmish) != 24 {
		t.Fatalf("%d months, %d skirmish maps", len(months), len(skirmish))
	}
	for _, path := range append(months, skirmish...) {
		f, err := scenario.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range scenario.Modes {
			lv, err := f.Level(m)
			if err != nil {
				t.Fatalf("%s: %v", filepath.Base(path), err)
			}
			b, err := dosbind.Bind(lv, set)
			if err != nil {
				t.Errorf("%s %s: %v", filepath.Base(path), m, err)
				continue
			}
			if _, err := b.Game(dosrules.Full); err != nil {
				t.Errorf("%s %s: game: %v", filepath.Base(path), m, err)
			}
		}
	}
}

// A setup's credits and builds (level.Extra) are the buildings' energy and
// what factories build, in place of records.json: Pulse binds as map 00
// did with the table's credits for mapNo 1.
func TestBindScenarioCredits(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lv, err := scenario.LoadLevel("../../levels/monthly/1914-08.json", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	b, err := dosbind.Bind(lv, set)
	if err != nil {
		t.Fatal(err)
	}
	want := dosbind.Table.Maps["1"]
	for idx, bd := range b.Buildings {
		if c := want.Credits[strconv.Itoa(idx)]; bd.Credits != c {
			t.Errorf("building record %d: %d credits, want %d", idx, bd.Credits, c)
		}
		if len(bd.Builds) > 0 && !slices.Equal(slices.Sorted(slices.Values(bd.Builds)), slices.Sorted(slices.Values(want.Builds))) {
			t.Errorf("building record %d builds %v, want %v", idx, bd.Builds, want.Builds)
		}
	}
	if len(b.Buildings) != len(want.Credits) {
		t.Errorf("%d buildings, want %d", len(b.Buildings), len(want.Credits))
	}
}
