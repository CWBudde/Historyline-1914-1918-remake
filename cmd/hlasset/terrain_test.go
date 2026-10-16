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

var (
	repoRoot        = filepath.Join("..", "..")
	terrainCoverage = filepath.Join(repoRoot, "docs", "art", "runtime-tile-coverage.csv")
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
	packs, err := buildTerrain(terrainCoverage, unitRecipes, repoRoot, out)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"backgrounds": 60, "infrastructure": 74, "context-ground": 110, "context-forest": 82, "context-hills": 159}
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
	src, err := os.ReadFile(filepath.Join(repoRoot, "art", "exports", "compound-v1", "hqN-4x.png"))
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
	if _, err := buildTerrain(terrainCoverage, unitRecipes, repoRoot, out); err == nil {
		t.Fatal("existing output directory reused")
	}
}

// copyTerrainInputs copies the coverage ledger and the terrain recipes to a
// temporary directory, applying edit to the named file.
func copyTerrainInputs(t *testing.T, file string, edit func(string) string) (coverage, recipes string) {
	t.Helper()
	dir := t.TempDir()
	recipes = filepath.Join(dir, "recipes")
	if err := os.Mkdir(recipes, 0o755); err != nil {
		t.Fatal(err)
	}
	copyFile := func(src, dst string) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if filepath.Base(src) == file {
			if s2 := edit(s); s2 != s {
				s = s2
			} else {
				t.Fatalf("edit of %s changed nothing", file)
			}
		}
		if err := os.WriteFile(dst, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	paths, _ := filepath.Glob(filepath.Join(unitRecipes, "*-v1.json"))
	for _, p := range paths {
		if !strings.HasPrefix(filepath.Base(p), "unit-") && !strings.HasPrefix(filepath.Base(p), "portrait-") {
			copyFile(p, filepath.Join(recipes, filepath.Base(p)))
		}
	}
	coverage = filepath.Join(dir, "coverage.csv")
	copyFile(terrainCoverage, coverage)
	return coverage, recipes
}

func TestTerrainBuildRefusesBadInputs(t *testing.T) {
	cases := []struct {
		name, file string
		edit       func(string) string
		want       string
	}{
		{"tampered hash", "compound-hq-v1.json", func(s string) string {
			return strings.ReplaceAll(s, `"sha256": "345ccd216dbe0a4ca1820b3cb4906fa17861107311782e28a710509ec001227c"`,
				`"sha256": "`+strings.Repeat("0", 64)+`"`)
		}, "tile/hqN"},
		{"unpinned export", "runtime-tile-coverage.csv", func(s string) string {
			return strings.Replace(s, "art/exports/compound-v1/hqN-2x.png", "art/exports/compound-v1/hqN-copy-2x.png", 1)
		}, "no pin"},
		{"missing density", "runtime-tile-coverage.csv", func(s string) string {
			return strings.Replace(s, ",art/exports/compound-v1/hqN-4x.png,", ",,", 1)
		}, "hqN"},
		{"context key", "hills-terrain-v1.json", func(s string) string {
			return strings.Replace(s, `"context_key": "e26f98f07c675934102b2039d788895d492df0a98dcbbdcf1762136b1ffe100f"`,
				`"context_key": "`+strings.Repeat("1", 64)+`"`, 1)
		}, "context key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			coverage, recipes := copyTerrainInputs(t, c.file, c.edit)
			out := filepath.Join(t.TempDir(), "terrain")
			_, err := buildTerrain(coverage, recipes, repoRoot, out)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to mention %q", err, c.want)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Fatal("failed build left its output directory")
			}
		})
	}
}

func TestTerrainGapsReport(t *testing.T) {
	out := filepath.Join(t.TempDir(), "terrain")
	var b bytes.Buffer
	if err := run([]string{"build-terrain", "-coverage", terrainCoverage, "-recipes", unitRecipes, "-root", repoRoot, "-out", out}, &b); err != nil {
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
