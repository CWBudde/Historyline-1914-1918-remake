package main

import (
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

// The terrain grade (terraingrade) brings the PULSE render to the direction's
// summer value targets: the 5–95 % L* range about 8–64 with the map mean at
// 39–42, land 4–5 b* less yellow than painted, relief darker than grass, the
// forest at 26–31 and deep water at 31–36.
func TestTerrainGradeValueTargets(t *testing.T) {
	p, err := load("../..", "../../art/reference/pulse/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	measure := func(ungraded bool) lookSource {
		r, err := maprender.New(os.DirFS("../../assets"))
		if err != nil {
			t.Fatal(err)
		}
		r.Ungraded = ungraded
		m, err := r.Render(p.level, compareDensity, false)
		if err != nil {
			t.Fatal(err)
		}
		return p.measureLookSource("render", halve(m.Image))
	}
	graded, painted := measure(false), measure(true)
	family := func(s lookSource, id string) metrics {
		for _, f := range s.Families {
			if f.ID == id && f.Colour != nil {
				return *f.Colour
			}
		}
		t.Fatalf("%s has no measured %s cells", s.ID, id)
		return metrics{}
	}
	w := graded.Whole
	if mean, p05, p95 := *w.MeanL.Value, *w.P05.Value, *w.P95.Value; mean < 39 || mean > 42 || p05 < 5 || p05 > 11 || p95 < 61 || p95 > 67 {
		t.Errorf("whole map: mean L* %.2f (39–42), 5–95 %% %.2f–%.2f (about 8–64)", mean, p05, p95)
	}
	// Every pixel loses the same b*, but a family's mean is taken over composited
	// pixels (ungraded cast shadows, colours clipped to sRGB), so each may miss by
	// half a unit.
	land := []string{"grass", "hills", "forest", "ground", "mount", "mountain"}
	sum := 0.0
	for _, id := range land {
		db := family(graded, id).B - family(painted, id).B
		sum += db
		if db < -5.5 || db > -3.5 {
			t.Errorf("%s: b* %.2f against painted %.2f, want 4–5 lower", id, family(graded, id).B, family(painted, id).B)
		}
	}
	if mean := sum / float64(len(land)); mean < -5 || mean > -4 {
		t.Errorf("land b* %.2f against painted, want 4–5 lower", mean)
	}
	grass := family(graded, "grass").L
	for _, id := range []string{"hills", "mount", "mountain"} {
		if l := family(graded, id).L; l >= grass {
			t.Errorf("%s: L* %.2f, not darker than grass %.2f", id, l, grass)
		}
	}
	for _, tc := range []struct {
		id     string
		lo, hi float64
	}{{"grass", 42, 45}, {"forest", 26, 31}, {"water", 31, 36}} {
		if l := family(graded, tc.id).L; l < tc.lo || l > tc.hi {
			t.Errorf("%s: L* %.2f, want %v–%v", tc.id, l, tc.lo, tc.hi)
		}
	}
}
