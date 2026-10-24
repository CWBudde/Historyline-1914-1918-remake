package main

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// A selected ground context lives only in a paired second pack part.
func terrainPartFixture(t *testing.T) (fstest.MapFS, *level.Level, mapcompose.Gap) {
	t.Helper()
	l := testLevel(t)
	s := mapcompose.Compose(l.Grid(), mapcompose.LevelHex(l), func(string) bool { return false })
	var gap mapcompose.Gap
	for _, g := range s.Gaps {
		if g.Family == mapcompose.Ground {
			gap = g
			break
		}
	}
	if gap.Descriptor == "" {
		t.Fatal("fixture has no ground context")
	}
	id := mapcompose.ContextID(gap.Family, mapcompose.Key(gap.Descriptor))
	fsys := terrainCatalog(t, l, "", id)
	for _, d := range terrainDensities {
		base := fmt.Sprintf("%dx/context-ground", d)
		part := fmt.Sprintf("%dx/context-ground-2", d)
		m, err := assets.ParseManifest(fsys[base+"/manifest.json"].Data)
		if err != nil {
			t.Fatal(err)
		}
		extra := assets.Manifest{Schema: 1, ID: fmt.Sprintf("terrain-context-ground-2-%dx", d), Status: "candidate"}
		kept := m.Entries[:0]
		for _, e := range m.Entries {
			if e.ID == id {
				extra.Entries = append(extra.Entries, e)
				fsys[path.Join(part, e.Path)] = fsys[path.Join(base, e.Path)]
			} else {
				kept = append(kept, e)
			}
		}
		m.Entries = kept
		for dir, manifest := range map[string]assets.Manifest{base: m, part: extra} {
			b, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			fsys[dir+"/manifest.json"] = &fstest.MapFile{Data: b}
		}
	}
	return fsys, l, gap
}

func TestOpenTerrainArtUsesPackParts(t *testing.T) {
	fsys, l, gap := terrainPartFixture(t)
	id := mapcompose.ContextID(gap.Family, mapcompose.Key(gap.Descriptor))
	a, err := openTerrainArt(fsys, assets.Options{AllowCandidates: true}, l)
	if err != nil {
		t.Fatal(err)
	}
	if a.group[id] != "context-ground-2" || !a.bleed["context-ground-2"] || a.scene.Cells[gap.At.Line][gap.At.Col].Background != id {
		t.Errorf("second part not selected with opaque borders: group=%q bleed=%t", a.group[id], a.bleed["context-ground-2"])
	}
	bg, _, _, _, dim := terrainLayers(&a.scene, gap.At, func(hexgrid.Coord) state.Fog { return state.Unseen })
	if bg != mapcompose.TileID("unseen") || dim {
		t.Errorf("part changed fog: %q,dim=%t", bg, dim)
	}
	manifest := fsys["4x/context-ground-2/manifest.json"]
	delete(fsys, "4x/context-ground-2/manifest.json")
	if _, err := openTerrainArt(fsys, assets.Options{AllowCandidates: true}, l); err == nil {
		t.Error("accepted a part without its 4x manifest")
	}
	fsys["4x/context-ground-2/manifest.json"] = manifest
	for _, d := range terrainDensities {
		base := fmt.Sprintf("%dx/context-ground/manifest.json", d)
		m, _ := assets.ParseManifest(fsys[base].Data)
		part, _ := assets.ParseManifest(fsys[fmt.Sprintf("%dx/context-ground-2/manifest.json", d)].Data)
		m.Entries = append(m.Entries, part.Entries...)
		b, _ := json.Marshal(m)
		fsys[base] = &fstest.MapFile{Data: b}
	}
	if _, err := openTerrainArt(fsys, assets.Options{AllowCandidates: true}, l); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("duplicate context across parts: %v", err)
	}
}

func TestHTTPTerrainPackParts(t *testing.T) {
	files, l, gap := terrainPartFixture(t)
	fsys, done, err := openArtDir(serveCatalog(t, files), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	a, err := openTerrainArt(fsys, assets.Options{AllowCandidates: true}, l)
	if err != nil {
		t.Fatal(err)
	}
	id := mapcompose.ContextID(gap.Family, mapcompose.Key(gap.Descriptor))
	if a.group[id] != "context-ground-2" {
		t.Errorf("HTTP ignored a context part: %q", a.group[id])
	}
}

func TestOpenTerrainArtRejectsDuplicateInHigherDensity(t *testing.T) {
	f, l, _ := terrainPartFixture(t)
	name := "4x/context-ground/manifest.json"
	m, _ := assets.ParseManifest(f[name].Data)
	part, _ := assets.ParseManifest(f["4x/context-ground-2/manifest.json"].Data)
	e := part.Entries[0]
	e.Path = "images/duplicate.png"
	m.Entries = append(m.Entries, e)
	b, _ := json.Marshal(m)
	f[name] = &fstest.MapFile{Data: b}
	f["4x/context-ground/"+e.Path] = f["4x/context-ground-2/"+part.Entries[0].Path]
	if _, err := openTerrainArt(f, assets.Options{AllowCandidates: true}, l); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate only at 4x: %v", err)
	}
}

func TestOpenTerrainArtRejectsUnpairedPartEntries(t *testing.T) {
	for _, change := range []string{"id", "anchor"} {
		t.Run(change, func(t *testing.T) {
			f, l, _ := terrainPartFixture(t)
			name := "4x/context-ground-2/manifest.json"
			m, _ := assets.ParseManifest(f[name].Data)
			if change == "id" {
				m.Entries[0].ID += "-extra"
			} else {
				m.Entries[0].Anchor[0]++
			}
			b, _ := json.Marshal(m)
			f[name] = &fstest.MapFile{Data: b}
			if _, err := openTerrainArt(f, assets.Options{AllowCandidates: true}, l); err == nil {
				t.Fatal("accepted unmatched part entries")
			}
		})
	}
}
