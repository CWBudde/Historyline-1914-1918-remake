package game

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// reveal marks the whole map as seen for the side to move, so that fog does not
// limit what the tests move across.
func reveal(s *State) {
	v := &s.vision[s.Turn]
	for i := range v.seen {
		for j := range v.seen[i] {
			v.seen[i][j] = true
		}
	}
}

func terrain(id string, infra ...string) level.Hex {
	return level.Hex{Terrain: id, Infra: level.Infra{IDs: infra, List: len(infra) > 1}}
}

func withUnit(h level.Hex, id string) level.Hex {
	h.Unit = id
	return h
}

// routes puts a unit of type id at (0,2) and returns what it can reach.
func routes(t *testing.T, place map[hexgrid.Coord]level.Hex, id string) (*State, map[hexgrid.Coord]Route) {
	t.Helper()
	place[at(0, 2)] = withUnit(place[at(0, 2)], id)
	if place[at(0, 2)].Terrain == "" {
		h := place[at(0, 2)]
		h.Terrain = "grass"
		place[at(0, 2)] = h
	}
	s := board(t, place)
	reveal(s)
	return s, s.Movable(at(0, 2), s.At(at(0, 2)).Unit)
}

func TestPointsCostByTerrain(t *testing.T) {
	// Infantry has speed 4: 8 points.
	_, r := routes(t, map[hexgrid.Coord]level.Hex{
		at(2, 2): terrain("grass", "road25"), at(4, 2): terrain("grass", "road25"),
		at(6, 2): terrain("grass", "road25"), at(8, 2): terrain("grass", "road25"),
		at(10, 2): terrain("grass", "road25"),
	}, "CP_infantry")
	for i, want := range []int{1, 2, 3, 4, 5} {
		if got := r[at(2*(i+1), 2)]; got.Cost != want || got.Left != 8-want || len(got.Path) != i+2 {
			t.Errorf("road %d: %+v, want cost %d", i+1, got, want)
		}
	}

	_, r = routes(t, map[hexgrid.Coord]level.Hex{at(2, 2): terrain("forest")}, "CP_infantry")
	if got := r[at(2, 2)]; got.Cost != 3 || got.Left != 5 {
		t.Errorf("forest: %+v, want cost 3, left 5", got)
	}
	if got := r[at(4, 2)]; got.Cost != 5 || got.Left != 3 || !slices.Equal(got.Path, []hexgrid.Coord{at(0, 2), at(2, 2), at(4, 2)}) {
		t.Errorf("forest then grass: %+v, want cost 5", got)
	}
	if _, ok := r[at(6, 2)]; ok && r[at(6, 2)].Cost > 8 {
		t.Errorf("(6,2) costs %d, beyond the budget", r[at(6, 2)].Cost)
	}
	for c, route := range r {
		if route.Cost > 8 {
			t.Errorf("%v costs %d, beyond the budget of 8", c, route.Cost)
		}
	}
}

func TestPointsFirstStep(t *testing.T) {
	// Heavy artillery has 2 points; a mountain costs 5.
	_, r := routes(t, map[hexgrid.Coord]level.Hex{at(2, 2): terrain("mountain")}, "CP_hArtillery")
	got, ok := r[at(2, 2)]
	if !ok || got.Cost != 5 || got.Left != 0 {
		t.Fatalf("mountain first step: %+v, ok %v", got, ok)
	}
	for c, route := range r {
		if route.Cost > 2 && len(route.Path) > 2 {
			t.Errorf("%v: %+v goes on beyond the mountain", c, route)
		}
	}
	// A second step is only allowed within the budget: grass costs the full 2.
	_, r = routes(t, map[hexgrid.Coord]level.Hex{}, "CP_hArtillery")
	if _, ok := r[at(2, 2)]; !ok {
		t.Error("no step onto grass")
	}
	if _, ok := r[at(4, 2)]; ok {
		t.Error("two grass hexes with 2 points")
	}
}

func TestPointsImpassableAndBlocked(t *testing.T) {
	_, r := routes(t, map[hexgrid.Coord]level.Hex{
		at(2, 2): terrain("house"), at(3, 1): withUnit(terrain("grass"), "EC_infantry"),
	}, "CP_infantry")
	if _, ok := r[at(2, 2)]; ok {
		t.Error("walked onto a house")
	}
	if _, ok := r[at(3, 1)]; ok {
		t.Error("walked onto an enemy")
	}
	got, ok := r[at(4, 2)]
	if !ok {
		t.Fatal("no way around the house")
	}
	for _, c := range got.Path {
		if c == at(2, 2) || c == at(3, 1) {
			t.Errorf("path %v crosses a blocked hex", got.Path)
		}
	}
}

func TestPointsOwnUnitsPass(t *testing.T) {
	_, r := routes(t, map[hexgrid.Coord]level.Hex{
		at(1, 1): withUnit(terrain("grass"), "CP_cavalry"), at(1, 2): withUnit(terrain("grass"), "CP_cavalry"),
	}, "CP_infantry")
	if _, ok := r[at(1, 1)]; ok {
		t.Error("stopped on an own unit that is no carrier")
	}
	if _, ok := r[at(2, 2)]; !ok {
		t.Error("cannot pass own units")
	}
}

func TestPointsMovementTypes(t *testing.T) {
	_, r := routes(t, map[hexgrid.Coord]level.Hex{
		at(2, 2): terrain("grass", "rail25"), at(4, 2): terrain("grass", "rail25"), at(6, 2): terrain("grass"),
	}, "CP_armTrain")
	if len(r) != 2 || r[at(2, 2)].Cost != 1 || r[at(4, 2)].Cost != 2 {
		t.Errorf("train routes = %v, want only the two rail hexes at cost 1 and 2", r)
	}

	_, r = routes(t, map[hexgrid.Coord]level.Hex{
		at(2, 2): terrain("grass", "stream25"), at(4, 2): terrain("grass", "stream25"), at(6, 2): terrain("grass"),
	}, "CP_patrolBoat")
	if len(r) != 2 || r[at(4, 2)].Cost != 2 {
		t.Errorf("boat routes = %v, want only the two water hexes", r)
	}

	_, r = routes(t, map[hexgrid.Coord]level.Hex{at(2, 2): terrain("grass", "stream25")}, "CP_infantry")
	if _, ok := r[at(2, 2)]; ok {
		t.Error("infantry walked into a stream")
	}
}

func TestPointsRoadBridge(t *testing.T) {
	place := map[hexgrid.Coord]level.Hex{at(2, 2): terrain("grass", "stream25", "rBridge36")}
	_, r := routes(t, place, "CP_infantry")
	if got, ok := r[at(2, 2)]; !ok || got.Cost != 1 {
		t.Errorf("bridge over a stream: %+v, ok %v", got, ok)
	}

	s := boardWith(t, Classic, map[hexgrid.Coord]level.Hex{at(0, 2): withUnit(terrain("grass"), "CP_infantry"),
		at(2, 2): terrain("grass", "stream25", "rBridge36")})
	reveal(s)
	if _, ok := s.Movable(at(0, 2), s.At(at(0, 2)).Unit)[at(2, 2)]; ok {
		t.Error("Classic read rBridge as a road")
	}
}

func TestOwnStatsOnExit(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): unit("CP_supplyCar")})
	inf := s.At(at(4, 2)).Unit
	if _, err := s.Move(at(4, 2), inf, at(6, 2)); err != nil {
		t.Fatal(err)
	}
	inf.Moved = false
	reveal(s)
	r := s.Movable(at(6, 2), inf)
	if len(r) == 0 {
		t.Fatal("no exit from the car")
	}
	for c, route := range r {
		if route.Cost > inf.Def.Speed*pointsPerSpeed {
			t.Errorf("exit to %v costs %d, more than the infantry's %d points", c, route.Cost, inf.Def.Speed*pointsPerSpeed)
		}
	}
	// The car's weight does not decide where the infantry may stop.
	car := s.At(at(6, 2)).Unit
	if _, err := s.Move(at(6, 2), inf, at(10, 2)); err != nil || car.Content.Used() != 0 {
		t.Errorf("leaving the car: %v, car holds %d", err, car.Content.Used())
	}
}

func TestFullDepotBlocks(t *testing.T) {
	// Eight heavy artillery of weight 5 fill the depot's 40.
	full := slices.Repeat([]string{"CP_hArtillery"}, 8)
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): building("cmpN", CentralPowers, full...)})
	inf := s.At(at(4, 2)).Unit
	if _, ok := s.Movable(at(4, 2), inf)[at(6, 2)]; ok {
		t.Error("a full depot is a legal target")
	}
	if _, err := s.Move(at(4, 2), inf, at(6, 2)); err == nil {
		t.Error("moved onto a full depot")
	}
	if s.At(at(4, 2)).Unit != inf || inf.Moved {
		t.Error("a refused move changed the unit")
	}
}

func TestBuildingTargets(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{
		at(4, 2): unit("CP_infantry"), at(4, 4): unit("CP_lArtillery"),
		at(6, 2): building("cmpN", None), at(6, 4): building("cmpN", CentralPowers),
	})
	reveal(s)
	if got, ok := s.Movable(at(4, 2), s.At(at(4, 2)).Unit)[at(6, 2)]; !ok || got.Cost != 1 {
		t.Errorf("capturing a neutral depot: %+v, ok %v", got, ok)
	}
	art := s.At(at(4, 4)).Unit
	if _, ok := s.Movable(at(4, 4), art)[at(6, 2)]; ok {
		t.Error("a unit without the capture skill may enter a neutral depot")
	}
	if _, ok := s.Movable(at(4, 4), art)[at(6, 4)]; !ok {
		t.Error("cannot enter an own depot")
	}
}

func TestClassicRouteCost(t *testing.T) {
	s := boardWith(t, Classic, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry")})
	r := s.Movable(at(4, 2), s.At(at(4, 2)).Unit)[at(10, 2)]
	if r.Cost != 3 || r.Left != 1 || len(r.Path) != 4 {
		t.Errorf("Classic route = %+v, want 3 steps of speed 4", r)
	}
}
