package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestTrenchIDs(t *testing.T) {
	for _, s := range []TrenchShape{
		{Ring: true}, {Ring: true, Links: 0b011000}, {Links: 0b001001}, {Spurs: 0b001001}, {Links: 0b000001, Spurs: 0b001000}, {Links: 0b011011},
	} {
		if got, ok := ParseTrench(s.ID()); !ok || got != s {
			t.Errorf("ParseTrench(%s) = %+v, %v", s.ID(), got, ok)
		}
	}
	for _, id := range []string{"trench", "trench0-", "trench41", "trench14-s", "trench14-s4", "trench7", "trench5a", "road14"} {
		if s, ok := ParseTrench(id); ok {
			t.Errorf("ParseTrench(%s) accepted %+v", id, s)
		}
	}
}

// trenchGrid is a grass map with trench pieces at the given hexes.
func trenchGrid(pieces map[hexgrid.Coord]string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	g := hexgrid.Grid{Width: 3, Lines: 5}
	return g, func(c hexgrid.Coord) Hex {
		h := Hex{Terrain: "grass"}
		if id, ok := pieces[c]; ok {
			h.Infra = []string{id}
		}
		return h
	}
}

func TestTrenchShapes(t *testing.T) {
	centre := hexgrid.Coord{Line: 2, Col: 1}
	around := hexgrid.Around(centre)
	for _, tt := range []struct {
		name   string
		pieces map[hexgrid.Coord]string
		want   string
	}{
		{"lone position", map[hexgrid.Coord]string{centre: "trench0"}, "trench0"},
		// A position takes the courses that lead into it.
		{"position", map[hexgrid.Coord]string{centre: "trench0", around[0]: "trench14", around[3]: "trench25"}, "trench0-1"},
		// A run toward no trench ends inside the hex.
		{"lone run", map[hexgrid.Coord]string{centre: "trench14"}, "trench-s14"},
		{"end of a line", map[hexgrid.Coord]string{centre: "trench14", around[0]: "trench14"}, "trench1-s4"},
		{"line", map[hexgrid.Coord]string{centre: "trench14", around[0]: "trench14", around[3]: "trench124"}, "trench14"},
		// A piece that leads into the hex links it, either way round.
		{"branch", map[hexgrid.Coord]string{centre: "trench25", around[0]: "trench24", around[1]: "trench25", around[4]: "trench0"}, "trench125"},
		// Side by side, neither piece leads into the other.
		{"parallel", map[hexgrid.Coord]string{centre: "trench14", around[1]: "trench14"}, "trench-s14"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g, at := trenchGrid(tt.pieces)
			s, ok := trenchShape(g, at, centre)
			if !ok || s.ID() != tt.want {
				t.Errorf("shape %s (%+v, %v), want %s", s.ID(), s, ok, tt.want)
			}
		})
	}
}

func TestComposeDrawsDerivedTrenches(t *testing.T) {
	centre := hexgrid.Coord{Line: 2, Col: 1}
	g, at := trenchGrid(map[hexgrid.Coord]string{centre: "trench14", hexgrid.Around(centre)[0]: "trench14"})
	s := Compose(g, at, func(id string) bool { return id == "tile/trench1-s4" })
	if got := s.Cells[2][1].Overlays; !slices.Equal(got, []string{"tile/trench1-s4"}) {
		t.Errorf("overlays %v", got)
	}
	// Without the tile the piece stays and the gap names the tile.
	s = Compose(g, at, none)
	if got := s.Cells[2][1].Overlays; !slices.Equal(got, []string{"tile/trench14"}) {
		t.Errorf("overlays without the tile %v", got)
	}
	if !slices.ContainsFunc(s.Gaps, func(gap Gap) bool { return gap.At == centre && gap.Family == Trench && gap.Descriptor == "trench1-s4" }) {
		t.Errorf("gaps %+v", s.Gaps)
	}
	// A course the catalog lacks, such as the end of a line a sapper dug,
	// is a dug-in position with its links.
	g, at = trenchGrid(map[hexgrid.Coord]string{centre: "trench2", hexgrid.Around(centre)[1]: "trench5"})
	s = Compose(g, at, func(id string) bool { return id == "tile/trench0-2" || id == "tile/trench5" })
	if got := s.Cells[2][1].Overlays; !slices.Equal(got, []string{"tile/trench0-2"}) || len(s.Gaps) != 0 {
		t.Errorf("overlays %v, gaps %+v", got, s.Gaps)
	}
}

// The pieces of the original map 03 ("VENOM", campaign level 4) that stood
// apart: a junction, a dug-in position and a run, each leading into the next.
func TestVenomTrenchesJoin(t *testing.T) {
	l, err := campaignLevelErr("1915-02:cp")
	if err != nil {
		t.Fatal(err)
	}
	at := func(c hexgrid.Coord) Hex {
		h := l.Lines[c.Line][c.Col]
		return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}
	for _, tt := range []struct {
		line, col int // 1-based, as the hex numbers show them
		want      string
	}{
		{4, 10, "trench2-s15"},
		{5, 11, "trench0-45"},
		{6, 10, "trench1-s4"},
	} {
		c := hexgrid.Coord{Line: tt.line - 1, Col: tt.col - 1}
		if s, _ := trenchShape(l.Grid(), at, c); s.ID() != tt.want {
			t.Errorf("%d,%d draws %s, want %s", tt.line, tt.col, s.ID(), tt.want)
		}
	}
}

// On every level two neighbouring trench hexes agree on their shared edge,
// and every port of a piece is drawn, as a link or a blind spur.
func TestTrenchesAgreeAcrossEdges(t *testing.T) {
	for _, nl := range allLevels(t) {
		name, l := nl.Name, nl.Level
		g := l.Grid()
		at := func(c hexgrid.Coord) Hex {
			h := l.Lines[c.Line][c.Col]
			return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}
		for line, hexes := range l.Lines {
			for col := range hexes {
				c := hexgrid.Coord{Line: line, Col: col}
				s, ok := trenchShape(g, at, c)
				if !ok {
					continue
				}
				if p, ok := ParseTrench(s.ID()); !ok || p != s {
					t.Errorf("%s %+v: shape %s does not parse back", name, c, s.ID())
				}
				_, ports, _, _ := trenchPiece(at(c).Infra)
				if ports&^(s.Links|s.Spurs) != 0 {
					t.Errorf("%s %+v: %s drops a port of %v", name, c, s.ID(), at(c).Infra)
				}
				for i, n := range hexgrid.Around(c) {
					if !g.Contains(n) {
						continue
					}
					o, ok := trenchShape(g, at, n)
					if mine, theirs := s.Links&(1<<i) != 0, ok && o.Links&(1<<((i+3)%6)) != 0; mine != theirs {
						t.Errorf("%s %+v and %+v disagree on their edge: %s / %s", name, c, n, s.ID(), o.ID())
					}
				}
			}
		}
	}
}
