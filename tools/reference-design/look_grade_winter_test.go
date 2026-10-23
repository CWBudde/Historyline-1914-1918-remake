package main

import (
	"io/fs"
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// The winter maps meet the direction's winter targets after the grade: snow
// at L* 86–90 and winter forest at least 35 under it. Open winter water (L*
// 24–32) needs winter water art; the maps still draw the summer water. Winter
// relief keeps its snow: only the rock takes the relief row, so a relief cell
// loses at most 4 L*.
func TestTerrainGradeWinterTargets(t *testing.T) {
	for _, src := range []struct {
		fsys fs.FS
		name string
	}{{levels.FS, "level2_cp.json"}, {levels.Campaign, "campaign/level15_cp.json"}} {
		raw, err := fs.ReadFile(src.fsys, src.name)
		if err != nil {
			t.Fatal(err)
		}
		lvl, err := level.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		families := map[string][]cell{}
		for l, line := range lvl.Lines {
			for c, h := range line {
				if len(h.Infra.IDs) == 0 && h.Content == nil {
					families[h.Terrain] = append(families[h.Terrain], cell{Row: l + 1, Col: c + 1})
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
			for _, f := range append([]string{"grass_w", "forest_w", "water"}, relief...) {
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
		for _, f := range relief {
			if m, ok := graded[f]; ok && m.L < painted[f].L-4 {
				t.Errorf("%s: %s L* %.2f, painted %.2f: the grade darkens its snow", lvl.MapName, f, m.L, painted[f].L)
			}
		}
	}
}
