package dosai

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// gasLevel is a session with side 1's gas pioneers at from and the units
// placed on the hexes of the cloud they would release on to, by index in
// the cloud's order; more adds other units.
func gasLevel(t *testing.T, from, to hexgrid.Coord, inCloud map[int]string, more map[hexgrid.Coord]string) (*dosplay.Session, []hexgrid.Coord) {
	t.Helper()
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_gasPioneers"}))
	cloud := s.Cloud(1, s.View(1).Units[from].Index, to)
	if len(cloud) != 3 {
		t.Fatalf("cloud %v", cloud)
	}
	units := map[hexgrid.Coord]string{from: "EC_gasPioneers"}
	for k, id := range inCloud {
		units[cloud[k]] = id
	}
	maps.Copy(units, more)
	return session(t, grassLevel(units)), cloud
}

func TestGasRelease(t *testing.T) {
	from := hexgrid.Coord{Line: 6, Col: 3}
	to := hexgrid.Around(from)[3]
	far := map[hexgrid.Coord]string{{Line: 18, Col: 3}: "CP_infantry"}

	// Two enemy foot units under the cloud: the pioneers release it.
	s, cloud := gasLevel(t, from, to, map[int]string{0: "CP_infantry", 1: "CP_infantry"}, nil)
	tasked := digs(t, s)
	if len(tasked) != 1 || tasked[0].From != from || !slices.Contains(cloud, tasked[0].To) {
		t.Fatalf("release orders %+v", tasked)
	}
	if got := s.Cloud(1, tasked[0].Unit, tasked[0].To); !slices.Contains(got, cloud[0]) || !slices.Contains(got, cloud[1]) {
		t.Errorf("the cloud on %v covers %v, not both enemies at %v", tasked[0].To, got, cloud[:2])
	}

	// One enemy is not enough; an own unit in the cloud forbids it (no
	// other cloud covers both ends of this one).
	for name, units := range map[string]map[int]string{
		"one enemy":       {0: "CP_infantry"},
		"own unit inside": {0: "CP_infantry", 1: "EC_infantry", 2: "CP_infantry"},
	} {
		s, _ := gasLevel(t, from, to, units, far)
		for _, ev := range digs(t, s) {
			if s.GasPioneer(ev.Unit) {
				t.Errorf("%s: released %+v", name, ev)
			}
		}
	}
}

func TestMayBuildExtension(t *testing.T) {
	for side := range 2 {
		if !mayBuild(nil, side, 51) || !mayBuild([]int{18}, side, 51) {
			t.Errorf("side %d may not build the gas pioneers", side)
		}
	}
	if mayBuild([]int{18}, 0, 36) {
		t.Error("side 0 builds an Entente aircraft")
	}
}

// On a shipped map where some infantry are gas pioneers, the computer sides
// release gas over the enemy in a played-out battle and keep their own units
// out of it.
func TestGasPlayOut(t *testing.T) {
	lv, err := scenario.ShippedLevel("1915-05", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for l, line := range lv.Lines {
		for c, h := range line {
			if strings.HasSuffix(h.Unit, "_infantry") {
				if n%3 == 0 {
					lv.Lines[l][c].Unit = strings.TrimSuffix(h.Unit, "_infantry") + "_gasPioneers"
				}
				n++
			}
		}
	}
	s := session(t, lv)
	_, evs := play(t, s, [2]*Planner{{}, {}})
	released, enemy, own := 0, 0, 0
	sideOf := map[int]int{}
	for _, ev := range evs {
		switch ev.Kind {
		case dosplay.Released:
			released++
			for _, c := range ev.Hexes {
				sideOf[c.Line*1000+c.Col] = ev.Side
			}
		case dosplay.Gassed:
			if ev.Side == sideOf[ev.To.Line*1000+ev.To.Col] {
				own++
			} else {
				enemy++
			}
		}
	}
	t.Logf("%d releases, %d enemy and %d own losses to gas", released, enemy, own)
	if released == 0 || enemy == 0 {
		t.Errorf("%d releases, %d enemy losses", released, enemy)
	}
	if own > enemy/2 {
		t.Errorf("own losses %d against enemy %d", own, enemy)
	}
}
