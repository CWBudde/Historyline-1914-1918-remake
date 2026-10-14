package game

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestVisionMatchesPython(t *testing.T) {
	set := loadSet(t)
	want := map[int]Fog{0: Visible, 1: Unseen, 2: Seen}
	for _, run := range loadPython(t).Levels {
		s, err := New(loadLevel(t, run.Level), set)
		if err != nil {
			t.Fatal(err)
		}
		bad := 0
		for i, line := range run.Fog {
			for j, f := range line {
				if got := s.Fog(run.Player, hexgrid.Coord{Line: i, Col: j}); got != want[f] && bad < 5 {
					bad++
					t.Errorf("%s player %d (%d,%d): Fog = %v, Python %d", run.Level, run.Player, i, j, got, f)
				}
			}
		}
	}
}

// find returns the coordinate of the first map unit of type id.
func find(t *testing.T, s *State, id string) hexgrid.Coord {
	t.Helper()
	for i := range s.Hexes {
		for j := range s.Hexes[i] {
			if u := s.Hexes[i][j].Unit; u != nil && u.Type == id {
				return hexgrid.Coord{Line: i, Col: j}
			}
		}
	}
	t.Fatalf("no %s on the map", id)
	return hexgrid.Coord{}
}

func TestVisionSeenAndBuildings(t *testing.T) {
	s, err := New(loadLevel(t, "level1_cp.json"), loadSet(t))
	if err != nil {
		t.Fatal(err)
	}
	if s.Fog(CentralPowers, hexgrid.Coord{Line: 46, Col: 0}) != Unseen {
		t.Fatal("far corner is not unseen at the start")
	}

	// Without units, a CP HQ still shows itself and its neighbours.
	var hq hexgrid.Coord
	for i := range s.Hexes {
		for j := range s.Hexes[i] {
			h := &s.Hexes[i][j]
			if h.Terrain == "hqN" && h.Owner == CentralPowers {
				hq = hexgrid.Coord{Line: i, Col: j}
			}
			h.Unit = nil
		}
	}
	s.vision = [3]vision{} // forget what the starting units saw
	s.UpdateVision(CentralPowers)
	for _, c := range append(s.Grid.Neighbors(hq), hq) {
		if f := s.Fog(CentralPowers, c); f != Visible {
			t.Errorf("HQ hex %v: Fog = %v, want Visible", c, f)
		}
	}

	// Losing the HQ leaves its surroundings seen, not unseen as in Python.
	s.At(hq).Owner = Entente
	s.updateVision()
	for _, c := range append(s.Grid.Neighbors(hq), hq) {
		if f := s.Fog(CentralPowers, c); f != Seen {
			t.Errorf("lost HQ hex %v: Fog = %v, want Seen", c, f)
		}
	}
	if f := s.Fog(Entente, hq); f != Visible {
		t.Errorf("captured HQ: Entente Fog = %v, want Visible", f)
	}
	if f := s.Fog(None, hq); f != Unseen {
		t.Errorf("Fog(None) = %v, want Unseen", f)
	}
	if f := s.Fog(CentralPowers, hexgrid.Coord{Line: -1}); f != Unseen {
		t.Errorf("off-map Fog = %v, want Unseen", f)
	}
}

func TestVisionUnitSight(t *testing.T) {
	s, err := New(loadLevel(t, "level1_cp.json"), loadSet(t))
	if err != nil {
		t.Fatal(err)
	}
	from := find(t, s, "CP_lArtillery")
	u := s.At(from).Unit
	for i := range s.Hexes {
		for j := range s.Hexes[i] {
			s.Hexes[i][j].Unit = nil
			s.Hexes[i][j].Container = nil
		}
	}
	s.At(from).Unit = u
	s.UpdateVision(CentralPowers)
	for i := range s.Hexes {
		for j := range s.Hexes[i] {
			c := hexgrid.Coord{Line: i, Col: j}
			inSight := hexgrid.Distance(c, from) <= u.Def.Sight
			if got := s.Fog(CentralPowers, c); (got == Visible) != inSight {
				t.Errorf("%v at distance %d: Fog = %v, in sight %v", c, hexgrid.Distance(c, from), got, inSight)
			}
		}
	}
}
