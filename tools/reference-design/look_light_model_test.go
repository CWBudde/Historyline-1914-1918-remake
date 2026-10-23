package main

import (
	"image/png"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terrainlight"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Every cast shadow on PULSE (scattered crowns, forest contexts beside
// infrastructure) and on campaign map 23 (also tree props on the
// meadow) follows the one light: it falls along 135°, 0.25 × its item's
// diameter away. A scattered crown's diameter is its size; it and its shadow
// have their own anchors, each rounded to the density grid, so the offset may
// miss by up to 1/density world px per axis. Winter forest casts no shadow yet.
func TestLightModelShadows(t *testing.T) {
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range []struct {
		fsys fs.FS
		name string
		want []string
	}{
		{levels.FS, "level1_cp.json", []string{"standing/scatter/crown", "standing/context/forest"}},
		{levels.Campaign, "campaign/level23_cp.json", []string{"standing/scatter/crown", "standing/props/trees"}},
	} {
		raw, err := fs.ReadFile(src.fsys, src.name)
		if err != nil {
			t.Fatal(err)
		}
		lvl, err := level.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, density := range []int{2, 4} {
			got, err := r.CastShadowPairs(lvl, density)
			if err != nil {
				t.Fatal(err)
			}
			diameters := map[string]float64{}
			kinds := map[string]int{}
			for _, p := range got.Pairs {
				if p.Reason != "" {
					continue
				}
				d, tolerance := 0.0, 0.05
				if size, ok := strings.CutPrefix(p.Standing.ID, "standing/scatter/crown/"); ok {
					if d, err = strconv.ParseFloat(size[strings.LastIndex(size, "-")+1:], 64); err != nil {
						t.Fatal(err)
					}
					tolerance = 1/float64(density) + 0.05
				} else if d, ok = diameters[p.Standing.Path]; !ok {
					d = diameter(t, p.Standing.Path, density)
					diameters[p.Standing.Path] = d
				}
				kinds[strings.Join(strings.Split(p.Standing.ID, "/")[:3], "/")]++
				along := terrainlight.Length * d / math.Sqrt2
				if math.Abs(p.Offset[0]-along) > tolerance || math.Abs(p.Offset[1]-along) > tolerance {
					t.Errorf("%s %dx %v %s: offset %.2f,%.2f world px for diameter %.1f, want %.2f,%.2f (135°)",
						lvl.MapName, density, p.At, p.Standing.ID, p.Offset[0], p.Offset[1], d, along, along)
				}
			}
			for _, k := range src.want {
				if kinds[k] == 0 {
					t.Errorf("%s %dx: no measured %s shadow", lvl.MapName, density, k)
				}
			}
			t.Logf("%s %dx: measured pairs %v, item diameters %v", lvl.MapName, density, kinds, diameters)
		}
	}
}

// diameter is terrainlight.Diameter of the runtime asset at path.
func diameter(t *testing.T, path string, density int) float64 {
	t.Helper()
	f, err := os.Open(filepath.Join("../../assets", path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return terrainlight.Diameter(im, density)
}
