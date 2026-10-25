package main

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

func oversizedTerrainIndex(t *testing.T) []indexEntry {
	t.Helper()
	// Keep the fixture at exactly 274 entries as the retained campaign grows.
	var es []indexEntry
	for _, e := range retainedIndex(t) {
		if !strings.HasPrefix(e.Pack, "terrain-context-ground-") {
			es = append(es, e)
		}
	}
	for i := 0; i < 274; i++ {
		for _, d := range terrainDensities {
			es = append(es, indexEntry{
				ID:   fmt.Sprintf("context/ground/campaign-%04d", i),
				Pack: fmt.Sprintf("terrain-context-ground-%dx", d), Density: d,
				Path: fmt.Sprintf("terrain/ground/campaign-%04d-%dx.png", i, d), SHA256: strings.Repeat("0", 64),
				Width: 96 * d, Height: 80 * d, Anchor: &[2]float64{48, 40}, Art: "terrain/ground",
			})
		}
	}
	return es
}

func TestTerrainPlanSplitsContextPacks(t *testing.T) {
	es := oversizedTerrainIndex(t)
	plan := func(entries []indexEntry) map[string]string {
		t.Helper()
		packs, err := planTerrain(entries)
		if err != nil {
			t.Fatal(err)
		}
		assignment := map[string]string{}
		counts := map[string]int{}
		for _, p := range packs {
			if !strings.HasPrefix(p.Group, "context-ground") {
				continue
			}
			var decoded int64
			for _, e := range p.Manifest.Entries {
				decoded += int64(e.Width) * int64(e.Height) * 8
				assignment[fmt.Sprintf("%s@%d", e.ID, p.Density)] = p.Group
			}
			if decoded > 256<<20 {
				t.Errorf("%s exceeds the unchanged decoded limit: %d", p.Manifest.ID, decoded)
			}
			counts[fmt.Sprintf("%s@%d", p.Group, p.Density)] = len(p.Manifest.Entries)
		}
		for _, d := range terrainDensities {
			for group, want := range map[string]int{"context-ground": 273, "context-ground-2": 1} {
				key := fmt.Sprintf("%s@%d", group, d)
				if counts[key] != want {
					t.Errorf("%s has %d entries, want %d", key, counts[key], want)
				}
			}
		}
		return assignment
	}
	got := plan(es)
	shuffled := slices.Clone(es)
	slices.Reverse(shuffled)
	if !reflect.DeepEqual(got, plan(shuffled)) {
		t.Error("index ordering changed the ID-to-part assignment")
	}
	for _, e := range es {
		if e.Pack == "terrain-context-ground-2x" && got[e.ID+"@2"] != got[e.ID+"@4"] {
			t.Errorf("%s moved between parts at different densities", e.ID)
		}
	}
}

func TestTerrainPlanRejectsUnpairedContext(t *testing.T) {
	es := retainedIndex(t)
	for i, e := range es {
		if e.Pack == "terrain-context-ground-4x" {
			es = append(es[:i], es[i+1:]...)
			break
		}
	}
	if _, err := planTerrain(es); err == nil {
		t.Fatal("accepted a context missing its 4x counterpart")
	}
}

func TestTerrainPlanRejectsUnsplitOversizedImage(t *testing.T) {
	es := retainedIndex(t)
	for _, d := range terrainDensities {
		es = append(es, indexEntry{
			ID: "stamp/crater_w/1/2", Pack: fmt.Sprintf("terrain-stamps-%dx", d), Density: d,
			Path: fmt.Sprintf("terrain/ground/oversized-%dx.png", d), SHA256: strings.Repeat("0", 64),
			Width: 2048 * d, Height: 2048 * d, Anchor: &[2]float64{48, 40}, Art: "terrain/ground",
		})
	}
	if _, err := planTerrain(es); err == nil {
		t.Fatal("accepted a single image larger than the decoded pack limit")
	}
}

func TestTerrainPartsEntryLimitAndGeometry(t *testing.T) {
	var input []terrainPack
	for _, d := range terrainDensities {
		m := assets.Manifest{Schema: 1, Status: "candidate"}
		for i := 0; i <= assets.MaxEntriesPerPack; i++ {
			m.Entries = append(m.Entries, assets.Entry{ID: fmt.Sprintf("tile/%05d", i), Density: d, Width: d, Height: d})
		}
		input = append(input, terrainPack{Group: "backgrounds", Density: d, Manifest: m})
	}
	parts, err := splitTerrain(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 4 || len(parts[0].Manifest.Entries) != assets.MaxEntriesPerPack || len(parts[2].Manifest.Entries) != 1 {
		t.Fatalf("entry limit not respected: %d parts", len(parts))
	}
	input[1].Manifest.Entries[0].Anchor[0] = 1
	if _, err := splitTerrain(input); err == nil {
		t.Fatal("accepted different logical anchors")
	}
	input[1].Manifest.Entries[0].Anchor[0] = 0
	input[1].Manifest.Entries[0].ID = "tile/other"
	if _, err := splitTerrain(input); err == nil {
		t.Fatal("accepted different paired ID sets")
	}
}

// A standing item and its cast shadow stay in one part: the pack validator
// requires a shadow's standing item in the same pack.
func TestTerrainPartsKeepShadowWithItsStanding(t *testing.T) {
	var input []terrainPack
	for _, d := range terrainDensities {
		m := assets.Manifest{Schema: 1, Status: "candidate"}
		for _, id := range []string{"standing/a", "standing/b", "standing/shadow/a", "standing/shadow/b"} {
			m.Entries = append(m.Entries, assets.Entry{ID: id, Density: d, Width: 4096 * d / 4, Height: 3072 * d / 4})
		}
		input = append(input, terrainPack{Group: "standing", Density: d, Manifest: m})
	}
	parts, err := splitTerrain(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) < 4 {
		t.Fatalf("%d parts, want the group split", len(parts))
	}
	in := map[string]string{}
	for _, p := range parts {
		for _, e := range p.Manifest.Entries {
			in[e.ID] = p.Group
		}
	}
	for _, id := range []string{"a", "b"} {
		if in["standing/"+id] != in["standing/shadow/"+id] {
			t.Errorf("standing/%s in %s, its shadow in %s", id, in["standing/"+id], in["standing/shadow/"+id])
		}
	}
}
