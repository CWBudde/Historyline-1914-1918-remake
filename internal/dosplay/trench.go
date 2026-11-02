package dosplay

import (
	"errors"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Sappers dig and fill trenches (dosrules sapper.go), construction units
// build depots (dosrules builder.go). The level stays as it was loaded;
// Reworked gives the hexes whose terrain the work changed, from the terrain
// records, so a resumed battle shows them as well.

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
	depotN // a complete depot's pieces
	depotE
	depotS
	depotW
	halfN // a half depot's pieces
	halfE
	halfS
	halfW
)

// depotGround are the depot records, complete (15-20) and half (21-24).
var depotGround = map[int]ground{
	15: depotN, 16: depotN, 17: depotN, 18: depotS, 19: depotW, 20: depotE,
	21: halfN, 22: halfS, 23: halfW, 24: halfE,
}

// depotOwner is the owner of a complete depot by its N piece's record.
var depotOwner = map[int]int{15: 2, 16: 0, 17: 1}

func groundOf(rec int) ground {
	switch {
	case dosrules.Trench(rec):
		return trench
	case rec == dosrules.TerrainFilled:
		return filled
	}
	if g, ok := depotGround[rec]; ok {
		return g
	}
	return plain
}

// Sapper reports whether unit digs and fills trenches instead of attacking:
// its action order goes to a neighbouring hex, which it digs or fills after
// the battles, and fights no battle.
func (s *Session) Sapper(unit int) bool { return s.Game.Units[unit].Sapper() }

// Builder reports whether unit builds depots instead of attacking: its
// action order goes to a depot site next to it (Targets), where it lays half
// a depot after the battles, or completes one, and fights no battle.
func (s *Session) Builder(unit int) bool { return s.Game.Units[unit].Builder() }

// CanConstruct is why construction unit of side can build no depot now (no
// charge left, no site), nil when it can.
func (s *Session) CanConstruct(side, unit int) error {
	u, ok := s.Game.Units[unit]
	if !ok || !u.Builder() {
		return refusal("not a construction unit")
	}
	err := s.Game.Check(dosrules.Command{Kind: dosrules.CmdAttack, Side: side, Unit: unit, Cell: u.Cell(side)})
	var r *dosrules.Refusal
	if errors.As(err, &r) && r.Code == dosrules.MsgNoTarget {
		return nil // a site elsewhere: its own cell is none
	}
	return wrap(err)
}

// Work is the work sapper unit of side would do on hex to as the side's own
// map shows it now, without ordering it: the hexes it changes, the terrain
// record each takes, and whether it fills (a trench on to) rather than
// digs. The hexes are nil for a unit that is no sapper or a hex that is no
// neighbour of it.
func (s *Session) Work(side, unit int, to hexgrid.Coord) (hexes []hexgrid.Coord, recs []int, fill bool) {
	cell, ok := s.Board.Cell(to)
	if !ok || !s.Sapper(unit) {
		return nil, nil, false
	}
	cells, rs, fill := s.Game.WorkCells(side, unit, cell)
	for k, c := range cells {
		if p, ok := s.Board.Coord(c); ok {
			hexes = append(hexes, p)
			recs = append(recs, rs[k])
		}
	}
	return hexes, recs, fill
}

// Reworked are the level's hexes whose terrain sappers or construction
// units changed, as the current terrain records make them: a dug hex is
// grass with a trench whose ports lead to its trench neighbours, a filled one
// grass without its trench (its record, 0x51, plays as grass), a depot's hex,
// half or complete, the depot's piece without infrastructure. A complete
// depot's N piece holds its owner as content; a half depot has none, which
// the map draws as the construction site (a depot the level gives no content
// binds to the half depot's records). The other hexes are the level's.
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
		if g := groundOf(rec); g >= depotN {
			h.Terrain = s.season("cmp" + string("NESW"[(g-depotN)%4]))
			h.Infra = level.Infra{}
			h.Content = nil
			if side, ok := depotOwner[rec]; ok {
				h.Content = &level.Content{Owner: int(s.Faction(side))}
			}
			out[c] = h
			continue
		}
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
