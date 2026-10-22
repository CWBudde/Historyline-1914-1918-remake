package game

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// A straight line of a stream, forest, mountain and a house: impassable or
// dear on the ground, one point per hex in the air.
func airLine() map[hexgrid.Coord]level.Hex {
	return map[hexgrid.Coord]level.Hex{
		at(2, 2): terrain("grass", "stream25"), at(4, 2): terrain("forest"),
		at(6, 2): terrain("mountain"), at(8, 2): terrain("house"),
	}
}

func TestAircraftFliesOverEveryTerrain(t *testing.T) {
	// The Fokker E.I has speed 7: 14 points.
	_, r := routes(t, airLine(), "CP_fokkerEI")
	for i := 1; i <= 4; i++ {
		c := at(2*i, 2)
		got, ok := r[c]
		if !ok || got.Cost != i || got.Left != 14-i || len(got.Path) != i+1 {
			t.Errorf("%v: %+v, ok %v, want cost %d straight on", c, got, ok, i)
		}
	}
	for c, route := range r {
		if route.Cost != len(route.Path)-1 {
			t.Errorf("%v: cost %d for %d hexes", c, route.Cost, len(route.Path)-1)
		}
	}

	// Classic counts steps; the terrain does not stop an aircraft either.
	s := boardWith(t, Classic, map[hexgrid.Coord]level.Hex{at(0, 2): withUnit(terrain("grass"), "CP_fokkerEI")})
	for c, h := range airLine() {
		*s.At(c) = mustHex(t, h)
	}
	reveal(s)
	r = s.Movable(at(0, 2), s.At(at(0, 2)).Unit)
	for i := 1; i <= 4; i++ {
		if got, ok := r[at(2*i, 2)]; !ok || got.Cost != i {
			t.Errorf("Classic %v: %+v, ok %v, want %d steps", at(2*i, 2), got, ok, i)
		}
	}

	// The ground units still see the line as before.
	_, r = routes(t, airLine(), "CP_infantry")
	if _, ok := r[at(2, 2)]; ok {
		t.Error("infantry walked into a stream")
	}
}

func TestAircraftKeepsOccupancyAndContainerRules(t *testing.T) {
	for _, rules := range []Rules{Standard, Classic} {
		s := boardWith(t, rules, map[hexgrid.Coord]level.Hex{
			at(4, 2): unit("CP_fokkerEI"),
			at(6, 2): unit("CP_supplyCar"),            // own carrier with room
			at(2, 2): building("cmpN", CentralPowers), // own depot
			at(3, 2): unit("EC_infantry"),             // enemy
			at(5, 2): unit("CP_infantry"),             // own unit, no carrier
		})
		reveal(s)
		r := s.Movable(at(4, 2), s.At(at(4, 2)).Unit)
		for _, c := range []hexgrid.Coord{at(6, 2), at(2, 2), at(3, 2), at(5, 2)} {
			if _, ok := r[c]; ok {
				t.Errorf("%+v: the aircraft may stop on %v", rules, c)
			}
		}
		if len(r) == 0 {
			t.Errorf("%+v: the aircraft cannot move at all", rules)
		}
		for c, route := range r {
			if slices.Contains(route.Path, at(3, 2)) || slices.Contains(route.Path, at(2, 2)) {
				t.Errorf("%+v: route to %v %v crosses the enemy or the depot", rules, c, route.Path)
			}
		}
	}
}

func TestAircraftOnlyHitByAirWeapons(t *testing.T) {
	// Infantry has only ground and water weapons: no target, no attack.
	s, _, _ := duel(t, Standard, "CP_infantry", "EC_spad7")
	if got := s.Targets(at(4, 2)); len(got) != 0 {
		t.Errorf("infantry Targets = %v, want none", got)
	}
	for slot := range 3 {
		if _, err := s.Attack(at(4, 2), at(6, 2), slot, zero); err == nil {
			t.Errorf("infantry slot %d hit an aircraft", slot)
		}
	}

	// Elite infantry carries an air machine gun (slot 2) besides its ground one.
	s, _, _ = duel(t, Standard, "CP_eliteInf", "EC_spad7")
	if got := s.Targets(at(4, 2)); !slices.Equal(got, []hexgrid.Coord{at(6, 2)}) {
		t.Errorf("elite infantry Targets = %v", got)
	}
	if _, err := s.Attack(at(4, 2), at(6, 2), 1, zero); err == nil {
		t.Error("a ground machine gun hit an aircraft")
	}
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err != nil {
		t.Errorf("air machine gun: %v", err)
	}

	// An anti-aircraft gun has no ground target; an aircraft attacking ground
	// troops gets no counter from them.
	s, _, _ = duel(t, Standard, "CP_staticAA", "EC_infantry")
	if got := s.Targets(at(4, 2)); len(got) != 0 {
		t.Errorf("anti-aircraft Targets = %v, want none on the ground", got)
	}
	s, _, _ = duel(t, Standard, "CP_fokkerDVII", "EC_infantry")
	if b, err := s.Attack(at(4, 2), at(6, 2), 0, zero); err != nil || b.Counter != -1 {
		t.Errorf("strafing: counter %d, err %v, want no counter", b.Counter, err)
	}
}

func TestAircraftOverImpassableTerrainCanBeAttacked(t *testing.T) {
	// A house has no battle modifier; an aircraft above it fights without one.
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_eliteInf"), at(6, 2): withUnit(terrain("house"), "EC_spad7")})
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err != nil {
		t.Errorf("aircraft over a house: %v", err)
	}
	s = board(t, map[hexgrid.Coord]level.Hex{at(4, 2): withUnit(terrain("house"), "CP_fokkerDVII"), at(6, 2): unit("EC_eliteInf")})
	if _, err := s.Attack(at(4, 2), at(6, 2), 0, zero); err != nil {
		t.Errorf("aircraft attacking from above a house: %v", err)
	}
}

// mustHex builds a game hex from a level hex.
func mustHex(t *testing.T, lh level.Hex) Hex {
	t.Helper()
	h, err := newHex(loadSet(t), lh, Classic)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
