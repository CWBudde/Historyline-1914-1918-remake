package game

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

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
	s, err := New(loadLevel(t, "pulse.json"), loadSet(t))
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

	// Losing the HQ leaves its surroundings seen, not unseen.
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
	s, err := New(loadLevel(t, "pulse.json"), loadSet(t))
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

func TestRememberedSnapshot(t *testing.T) {
	obs, enemy, empty, depot := at(4, 2), at(4, 3), at(6, 2), at(2, 2)
	s := board(t, map[hexgrid.Coord]level.Hex{
		obs: unit("CP_infantry"), enemy: unit("EC_infantry"),
		depot: building("cmpN", Entente),
	})
	observer := s.At(obs).Unit
	u := s.At(enemy).Unit
	snap := Memory{Unit: &UnitMemory{Type: "EC_infantry", Owner: Entente, Size: 6}}
	show := func(m Memory) string {
		if m.Unit == nil {
			return fmt.Sprintf("{no unit, owner %v}", m.Owner)
		}
		return fmt.Sprintf("{%+v, owner %v}", *m.Unit, m.Owner)
	}
	remembered := func(c hexgrid.Coord, want Memory) {
		t.Helper()
		if got, ok := s.Remembered(CentralPowers, c); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Remembered(%v) = %v, %v; want %v, true", c, show(got), ok, show(want))
		}
	}

	// In sight, the memory is the live hex.
	remembered(enemy, snap)
	remembered(empty, Memory{})
	remembered(depot, Memory{Owner: Entente})

	// Out of sight, it keeps the last seen state while the enemy shrinks,
	// moves to an empty hex and the depot changes hands.
	s.At(obs).Unit = nil
	s.UpdateVision(CentralPowers)
	if f := s.Fog(CentralPowers, enemy); f != Seen {
		t.Fatalf("enemy hex Fog = %v, want Seen", f)
	}
	u.Size = 3
	remembered(enemy, snap)
	s.At(enemy).Unit = nil
	s.At(empty).Unit = u
	s.At(depot).Owner = None
	s.UpdateVision(CentralPowers)
	remembered(enemy, snap)
	remembered(empty, Memory{})
	remembered(depot, Memory{Owner: Entente})

	// Seen again, the memory is refreshed.
	s.At(obs).Unit = observer
	s.UpdateVision(CentralPowers)
	remembered(enemy, Memory{})
	remembered(empty, Memory{Unit: &UnitMemory{Type: "EC_infantry", Owner: Entente, Size: 3}})
	remembered(depot, Memory{})
	s.At(obs).Unit = nil
	s.UpdateVision(CentralPowers)
	remembered(empty, Memory{Unit: &UnitMemory{Type: "EC_infantry", Owner: Entente, Size: 3}})

	for _, c := range []struct {
		p Player
		c hexgrid.Coord
	}{{CentralPowers, at(11, 0)}, {CentralPowers, at(-1, 0)}, {CentralPowers, at(0, 9)}, {None, enemy}, {Player(3), enemy}} {
		if got, ok := s.Remembered(c.p, c.c); ok || got != (Memory{}) {
			t.Errorf("Remembered(%v, %v) = %+v, %v; want nothing", c.p, c.c, got, ok)
		}
	}
}
