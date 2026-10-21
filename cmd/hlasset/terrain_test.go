package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func openTerrainPack(t *testing.T, out string, d int, group string) *assets.Pack {
	t.Helper()
	r, err := os.OpenRoot(filepath.Join(out, terrainDir(d, group)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	p, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRetainedTerrainCatalogBuilds(t *testing.T) {
	out := filepath.Join(t.TempDir(), "terrain")
	packs, err := buildTerrain(assetsRoot, out)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"backgrounds": 76, "infrastructure": 74, "context-ground": 110, "context-forest": 82, "context-hills": 159}
	if len(packs) != len(want)*2 {
		t.Fatalf("built %d packs, want %d", len(packs), len(want)*2)
	}
	for _, p := range packs {
		if n := len(p.Manifest.Entries); n != want[p.Group] || p.Manifest.Status != "candidate" {
			t.Errorf("%s: %d entries (want %d), status %s", p.Manifest.ID, n, want[p.Group], p.Manifest.Status)
		}
	}
	// Every terrain and infrastructure ID of the shipped levels has a tile at
	// both densities, registered on the hex centre.
	for _, d := range terrainDensities {
		bg := openTerrainPack(t, out, d, "backgrounds")
		inf := openTerrainPack(t, out, d, "infrastructure")
		names, err := fs.Glob(levels.FS, "*.json")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			b, _ := fs.ReadFile(levels.FS, name)
			l, err := level.Parse(b)
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range l.Lines {
				for _, h := range line {
					e, err := bg.Entry(mapcompose.TileID(h.Terrain), d)
					if err != nil {
						t.Fatal(err)
					}
					if e.Anchor != [2]float64{48, 40} || e.Width != 96*d || e.Height != 80*d {
						t.Fatalf("%s: %v %dx%d", e.ID, e.Anchor, e.Width, e.Height)
					}
					for _, id := range h.Infra.IDs {
						if _, err := inf.Entry(mapcompose.TileID(id), d); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		}
	}
	// Bytes are copied, not re-encoded.
	src, err := os.ReadFile(filepath.Join(assetsRoot, "terrain", "hq", "hqN-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := openTerrainPack(t, out, 4, "backgrounds").WritePNG(&got, "tile/hqN", 4); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Bytes(), src) {
		t.Fatal("runtime build changed the source PNG bytes")
	}
	if _, err := buildTerrain(assetsRoot, out); err == nil {
		t.Fatal("existing output directory reused")
	}
}

func TestTerrainPlanIncludesGrassVariations(t *testing.T) {
	packs, err := planTerrain(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[int]int{}
	for _, p := range packs {
		for _, e := range p.Manifest.Entries {
			if strings.HasPrefix(e.ID, "tile/grass-detail-") {
				if p.Group != "backgrounds" {
					t.Errorf("grass detail outside the background pack: %+v", e)
				}
				counts[p.Density]++
			}
		}
	}
	if counts[2] != 16 || counts[4] != 16 {
		t.Fatalf("terrain catalog has grass details %v, want sixteen per density", counts)
	}
}

func TestTerrainBuildRefusesBadInputs(t *testing.T) {
	entry := func(es []indexEntry, pack, id string) *indexEntry {
		for i := range es {
			if es[i].Pack == pack && es[i].ID == id {
				return &es[i]
			}
		}
		t.Fatalf("no %s in %s", id, pack)
		return nil
	}
	cases := []struct {
		name   string
		damage func([]indexEntry) []indexEntry
		want   string
	}{
		{"wrong registration", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-backgrounds-2x", "tile/hqN").Anchor = &[2]float64{0, 0}
			return es
		}, "tile/hqN"},
		{"wrong size", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-backgrounds-4x", "tile/hqN").Height = 160
			return es
		}, "tile/hqN"},
		{"wrong group", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-infrastructure-2x", "tile/road124").Pack = "terrain-context-hills-2x"
			return es
		}, "tile/road124"},
		{"missing pack", func(es []indexEntry) []indexEntry {
			var kept []indexEntry
			for _, e := range es {
				if e.Pack != "terrain-context-forest-4x" {
					kept = append(kept, e)
				}
			}
			return kept
		}, "terrain-context-forest-4x"},
		{"unknown pack", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-backgrounds-2x", "tile/hqN").Pack = "terrain-roofs-2x"
			return es
		}, "terrain-roofs-2x"},
	}
	for _, c := range cases {
		if _, err := planTerrain(c.damage(retainedIndex(t))); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want it to mention %q", c.name, err, c.want)
		}
	}
	// A tampered pin fails validation before the output directory exists.
	packs, err := planTerrain(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	m := packs[0].Manifest
	m.Entries[0].SHA256 = strings.Repeat("0", 64)
	out := filepath.Join(t.TempDir(), "terrain")
	if _, err := writePacks(assetsRoot, out, []assets.Manifest{m}, func(int) string { return "." }); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatal("tampered pin accepted", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed build left its output directory")
	}
}

func TestTerrainGapsReport(t *testing.T) {
	out := filepath.Join(t.TempDir(), "terrain")
	var b bytes.Buffer
	if err := run([]string{"build-terrain", "-assets", assetsRoot, "-out", out}, &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "built 10 terrain packs") {
		t.Errorf("build output %q", b.String())
	}
	b.Reset()
	if err := run([]string{"terrain-gaps", "-terrain", out}, &b); err != nil {
		t.Fatal(err)
	}
	s := b.String()
	for _, want := range []string{
		"level1_cp: 156 context tiles selected, 0 missing, 0 undefined, 0 compound issues",
		"level2_cp: 233 context tiles selected, 0 missing, 0 undefined",
		"level3_cp: 175 context tiles selected, 0 missing, 0 undefined",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("report lacks %q", want)
		}
	}
}
