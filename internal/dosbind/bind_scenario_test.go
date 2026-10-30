package dosbind_test

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Every setup of the shipped level folders must bind for -rules dos.
func TestBindScenarios(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	dirs, err := scenario.Months(os.DirFS(levelsDir))
	if err != nil || len(dirs) < 51 {
		t.Fatalf("%d level folders, %v", len(dirs), err)
	}
	for _, dir := range dirs {
		path := filepath.Join(levelsDir, filepath.FromSlash(dir))
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
	lv, err := scenario.LoadLevel(filepath.Join(levelsDir, "1914-08"), scenario.CP)
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

// Every hex of a hand-authored month binds to a terrain record of its own:
// none loses an overlay to the fallback that drops unrecorded
// infrastructure.
func TestAuthoredHexesHaveRecords(t *testing.T) {
	for _, month := range []string{"1914-09", "1914-11", "1915-01", "1918-10"} {
		if month == "1914-09" || month == "1914-11" {
			continue // predate this check: a double road, paths on mud and forest
		}
		f, err := scenario.Load(filepath.Join(levelsDir, month))
		if err != nil {
			t.Fatal(err)
		}
		for l, line := range f.Map {
			for c, h := range line {
				if strings.HasPrefix(h.Terrain, "hq") || strings.HasPrefix(h.Terrain, "cmp") ||
					strings.HasPrefix(h.Terrain, "factory") {
					continue // buildings bind by owner
				}
				if k := dosbind.Key(h); !recorded(k) {
					t.Errorf("%s line %d col %d: no record %q", month, l+1, c+1, k)
				}
			}
		}
	}
}

func recorded(key string) bool {
	_, ok := dosbind.Table.Terrain[key]
	return ok
}

// The road bridge over the lake of 1917-12 (Amper) binds as a road for
// ground units, not as the water under it.
func TestBindRoadBridgeOverWater(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lv, err := scenario.LoadLevel(filepath.Join(levelsDir, "1917-12"), scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	b, err := dosbind.Bind(lv, set)
	if err != nil {
		t.Fatal(err)
	}
	terrains := dosrules.Full.Tables.Move.Terrains
	road := terrains[dosbind.Table.Terrain["grass+road14"]].Mask
	n := 0
	for l, line := range lv.Lines {
		for c, h := range line {
			if h.Terrain != "water" || !slices.Contains(h.Infra.IDs, "rBridge14") {
				continue
			}
			n++
			cell, _ := b.Cell(hexgrid.Coord{Line: l, Col: c})
			if rec := b.Terrain[cell]; terrains[rec].Mask != road {
				t.Errorf("line%d[%d]: record %d, mask %#x, a road has %#x", l+1, c, rec, terrains[rec].Mask, road)
			}
		}
	}
	if n != 4 {
		t.Errorf("%d bridge hexes, want 4", n)
	}
}

// levelsDir is the levels directory on disk.
var levelsDir = filepath.Join("..", "..", "levels")
