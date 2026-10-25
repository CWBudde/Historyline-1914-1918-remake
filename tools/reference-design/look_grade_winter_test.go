package main

import (
	"os"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

// The winter maps meet the direction's winter targets after the grade: snow
// at L* 86–90, winter forest at least 35 under it, and open water at L* 24–32
// and chroma 12 or less (water_w, the winter water grade). Winter
// relief keeps its snow: only the rock takes the relief row, so a relief cell
// loses at most 4 L*.
func TestTerrainGradeWinterTargets(t *testing.T) {
	for _, name := range []string{"monthly/1914-10:cp", "monthly/1916-12:cp"} {
		lvl := testLevel(t, name)
		families := map[string][]cell{}
		g := lvl.Grid()
		for l, line := range lvl.Lines {
			for c, h := range line {
				if len(h.Infra.IDs) == 0 && h.Content == nil {
					f := h.Terrain
					// Open water of every depth is measured as one.
					if mapcompose.OpenWater(f) {
						f = "water"
					}
					// Snow beside winter forest carries the crowns that the woodland
					// edge spills over the hex edge, and their shadows: measured apart.
					if f == "grass_w" && slices.ContainsFunc(g.Neighbors(hexgrid.Coord{Line: l, Col: c}), func(n hexgrid.Coord) bool {
						return lvl.Lines[n.Line][n.Col].Terrain == "forest_w"
					}) {
						f = "grass_w by forest"
					}
					families[f] = append(families[f], cell{Row: l + 1, Col: c + 1})
				}
			}
		}
		relief := []string{"mountain_w", "mountE_w", "mountN_w", "mountS_w", "mountW_w"}
		measure := func(ungraded bool) map[string]metrics {
			r, err := maprender.New(os.DirFS("../../assets"))
			if err != nil {
				t.Fatal(err)
			}
			r.Ungraded = ungraded
			m, err := r.Render(lvl, compareDensity, false)
			if err != nil {
				t.Fatal(err)
			}
			plane := newLabPlane(halve(m.Image))
			out := map[string]metrics{}
			for _, f := range append([]string{"grass_w", "grass_w by forest", "forest_w", "water"}, relief...) {
				if len(families[f]) > 0 {
					out[f] = plane.measure(families[f])
				}
			}
			return out
		}
		graded, painted := measure(false), measure(true)
		for f, m := range graded {
			t.Logf("%s %s (%d cells): L* %.2f (painted %.2f), σ %.2f, b* %.2f (painted %.2f), chroma %.2f",
				lvl.MapName, f, len(families[f]), m.L, painted[f].L, m.Contrast, m.B, painted[f].B, m.Chroma)
		}
		snow, forest := graded["grass_w"], graded["forest_w"]
		if snow.L < 86 || snow.L > 90 {
			t.Errorf("%s: snow L* %.2f, want 86–90", lvl.MapName, snow.L)
		}
		if d := forest.L - snow.L; d > -35 {
			t.Errorf("%s: winter forest %.2f against snow, want at most −35", lvl.MapName, d)
		}
		if water, ok := graded["water"]; ok && (water.L < 24 || water.L > 32 || water.Chroma > 12) {
			t.Errorf("%s: open water L* %.2f, chroma %.2f, want L* 24–32 and chroma at most 12", lvl.MapName, water.L, water.Chroma)
		}
		for _, f := range relief {
			if m, ok := graded[f]; ok && m.L < painted[f].L-4 {
				t.Errorf("%s: %s L* %.2f, painted %.2f: the grade darkens its snow", lvl.MapName, f, m.L, painted[f].L)
			}
		}
	}
}
