package dosplay

import (
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Sappers dig and fill trenches (dosrules sapper.go). The level stays as it
// was loaded; Reworked gives the hexes whose terrain the work changed, from
// the terrain records, so a resumed battle shows them as well.

// trenchPorts are the directions of hexgrid.Around a trench course may take:
// the original's trench pieces and a sapper's work run only along the four
// diagonals (NE, SE, SW, NW).
var trenchPorts = []int{0, 1, 3, 4}

// ground is what a terrain record is to the sappers' work.
type ground int

const (
	plain ground = iota
	trench
	filled
)

func groundOf(rec int) ground {
	switch {
	case dosrules.Trench(rec):
		return trench
	case rec == dosrules.TerrainFilled:
		return filled
	}
	return plain
}

// Sapper reports whether unit digs and fills trenches instead of attacking:
// its action order goes to a neighbouring hex, which it digs or fills after
// the battles, and fights no battle.
func (s *Session) Sapper(unit int) bool { return s.Game.Units[unit].Sapper() }

// Reworked are the level's hexes whose terrain sappers changed, as the
// current terrain records make them: a dug hex is grass with a trench whose
// ports lead to its trench neighbours, a filled one grass without its trench
// (its record, 0x51, plays as grass). The other hexes are the level's.
func (s *Session) Reworked() map[hexgrid.Coord]level.Hex {
	out := map[hexgrid.Coord]level.Hex{}
	m := s.Game.Maps[0]
	for cell, orig := range s.Board.Terrain {
		rec := m[2*cell]
		c, ok := s.Board.Coord(cell)
		if !ok || groundOf(rec) == groundOf(orig) {
			continue
		}
		h := s.Level.Lines[c.Line][c.Col]
		ids := slices.DeleteFunc(slices.Clone(h.Infra.IDs), func(id string) bool { return strings.HasPrefix(id, "trench") })
		h.Terrain = s.season("grass")
		if groundOf(rec) == trench {
			ids = append(ids, s.trenchPiece(c))
		}
		h.Infra = level.Infra{IDs: ids, List: len(ids) > 1}
		out[c] = h
	}
	return out
}

// season is a terrain name in the board's season.
func (s *Session) season(terrain string) string {
	if s.Board.Winter {
		return terrain + "_w"
	}
	return terrain
}

// trenchPiece is the trench id of a dug hex: its ports toward the trench
// hexes on the diagonals, or the dug-in position trench0 without one.
func (s *Session) trenchPiece(c hexgrid.Coord) string {
	id := "trench"
	around := hexgrid.Around(c)
	for _, k := range trenchPorts {
		if cell, ok := s.Board.Cell(around[k]); ok && dosrules.Trench(s.Game.Maps[0][2*cell]) {
			id += string(rune('1' + k))
		}
	}
	if id == "trench" {
		return "trench0"
	}
	return id
}
