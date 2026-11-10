package dosrules

import (
	"maps"
	"slices"
)

// Sappers dig and fill trenches. A unit whose live +6 has bit 0x2000 (the
// sappers' type word +0x11) gets the attack order on a neighbouring cell
// whose terrain record has flag 0x8000 (0ca6:13c5, the target marking), with
// or without a unit on it. The order (0ad1:000c) sets the moved flag, +0x11
// to the cell and +0x13 = 0x11 and shows message 0x6e; it does not enter the
// order list, so it does not count for the move limit and fights no battle.
//
// The work runs in the action pass after the battles, in unit order with the
// captures (02dc:197b, 0e47:1b4c), on the action side's map; the phase change
// copies it to the other side's. A target below the first trench record is
// dug, any other filled. The work lays three cells: the target and its two
// neighbours across the sapper's front, a straight line where the hexes allow
// one, a bend toward the sapper where the target lies straight above or
// below it. Each of them whose terrain has flag 0x8000 takes the trench
// record the table gives, or the filled record 0x51 (plain ground that plays
// as grass); the others keep theirs.

const (
	flagSapper    = 0x2000 // live +6: a sapper
	terrainWorked = 0x8000 // terrain flag word: a sapper digs and fills here
	actionTrench  = 0x11   // +0x13: the sapper's work order
	trenchFirst   = 0xA5   // the first trench record; a target from here on is filled
	// TerrainFilled is the record a filled cell takes.
	TerrainFilled = 0x51
	// MsgTrench is the message of a sapper's work order.
	MsgTrench = 0x6E
)

// Sapper reports whether the unit digs and fills trenches instead of
// attacking.
func (u Live) Sapper() bool { return u.Flags6&flagSapper != 0 }

// Working reports whether the unit has a sapper's work order this phase.
func (u Live) Working() bool { return u.Action[0] == actionTrench }

// trenchCell is a cell of a sapper's work: its step from the sapper (column,
// row) and the trench record it takes.
type trenchCell struct{ dx, dy, rec int }

// trenchWork are the three cells of the work on each neighbour of the
// sapper, by the sapper's column parity and the neighbour in the order of
// steps (32f7:0061 and 00cd, indexed there by the column and row
// differences); the target is the middle one.
var trenchWork = [2][6][3]trenchCell{
	{
		{{-1, -1, 0xAA}, {0, -1, 0xA7}, {1, -1, 0xAB}},
		{{0, -1, 0xAB}, {1, -1, 0xA7}, {2, 0, 0xAB}},
		{{2, 0, 0xAA}, {1, 0, 0xA7}, {0, 1, 0xAA}},
		{{1, 0, 0xAA}, {0, 1, 0xA7}, {-1, 0, 0xAB}},
		{{0, 1, 0xAB}, {-1, 0, 0xA7}, {-2, 0, 0xAB}},
		{{-2, 0, 0xAA}, {-1, -1, 0xA7}, {0, -1, 0xAA}},
	},
	{
		{{-1, 0, 0xAA}, {0, -1, 0xA7}, {1, 0, 0xAB}},
		{{0, -1, 0xAB}, {1, 0, 0xA7}, {2, 0, 0xAB}},
		{{2, 0, 0xAA}, {1, 1, 0xA7}, {0, 1, 0xAA}},
		{{1, 1, 0xAA}, {0, 1, 0xA7}, {-1, 1, 0xAB}},
		{{0, 1, 0xAB}, {-1, 1, 0xA7}, {-2, 0, 0xAB}},
		{{-2, 0, 0xAA}, {-1, 0, 0xA7}, {0, -1, 0xAA}},
	},
}

// Trench reports whether a terrain record is a trench.
func Trench(rec int) bool { return rec >= trenchFirst && rec <= 0xAD }

// worked reports whether a sapper digs and fills on a terrain record.
func (g *Game) worked(rec int) bool {
	t, err := g.Profile.moveTables()
	return err == nil && rec < len(t.Terrains) && t.Terrains[rec].Flags&terrainWorked != 0
}

// trenchTarget reports whether sapper u of side may work on cell: a
// neighbour of its cell on the side's map with worked terrain.
func (g *Game) trenchTarget(side int, u Live, cell int) bool {
	from := u.Cell(side)
	n := neighbours(from, g.Width, g.Height)
	return slices.Contains(n[:], cell) && g.worked(g.Maps[side][2*cell])
}

// WorkCells is the work sapper unit of side would do on cell as the side's
// map shows it now, without doing it: the cells it changes, the terrain
// record each takes and whether it fills (a trench on cell) rather than
// digs. The cells are nil for a cell the unit may not work on (no
// neighbour, or terrain sappers leave alone).
func (g *Game) WorkCells(side, unit, cell int) (cells, recs []int, fill bool) {
	u, ok := g.Units[unit]
	if !ok || side != 0 && side != 1 || cell < 0 || cell >= g.Width*g.Height || !g.trenchTarget(side, u, cell) {
		return nil, nil, false
	}
	return g.workCells(side, u.Cell(side), cell)
}

// workCells is the work from cell from on cell to on side's map: the cells
// sappers work among the three, the record each takes, and whether the
// work fills. A cell off the map is left alone (the original's offset reads
// beyond the map's rows).
func (g *Game) workCells(side, from, to int) (cells, recs []int, fill bool) {
	m := g.Maps[side]
	fill = m[2*to] >= trenchFirst
	n := neighbours(from, g.Width, g.Height)
	k := slices.Index(n[:], to)
	if k < 0 {
		return nil, nil, fill
	}
	x, y := from%g.Width, from/g.Width
	for _, c := range trenchWork[x&1][k] {
		cell := (y+c.dy)*g.Width + x + c.dx
		if cell < 0 || cell >= g.Width*g.Height || !g.worked(m[2*cell]) {
			continue
		}
		rec := c.rec
		if fill {
			rec = TerrainFilled
		}
		cells = append(cells, cell)
		recs = append(recs, rec)
	}
	return cells, recs, fill
}

// trench is the work of sapper i of side a on a's map, the terrain the
// cells take: dug below the first trench record, filled from there on.
func (g *Game) trench(a, i int) Event {
	u := g.Units[i]
	to := u.Target / 2
	cells, recs, fill := g.workCells(a, u.Cell(a), to)
	ev := Event{Kind: EvDug, Side: a, Unit: i, Cell: to, Cells: cells}
	if fill {
		ev.Kind = EvFilled
	}
	for k, cell := range cells {
		g.Maps[a][2*cell] = recs[k]
	}
	return ev
}

// work is the action pass for side a's resolution, in unit order: the
// captures (capture), the sappers' work (trench) and the construction units'
// (builder.go). It returns their events and the side that took an HQ (-1 if
// none), the order handler's return value.
func (g *Game) work(a int) ([]Event, int) {
	won := -1
	var evs []Event
	for _, i := range slices.Sorted(maps.Keys(g.Units)) {
		u := g.Units[i]
		if u.Destroyed() || u.Action[1]&4 != 0 {
			continue
		}
		switch u.Action[0] {
		case 4:
			if ev, hq, ok := g.capture(a, i); ok {
				evs = append(evs, ev)
				if hq {
					won = 1 - a
				}
			}
		case actionTrench:
			evs = append(evs, g.trench(a, i))
		case actionHalf:
			if ev, ok := g.halfDepotWork(a, i); ok {
				evs = append(evs, ev)
			}
		case actionDepot:
			evs = append(evs, g.depotWork(a, i))
		case actionGas:
			evs = append(evs, g.releaseGas(a, i))
		case actionWire:
			evs = append(evs, g.layWire(a, i))
		case actionTunnel:
			evs = append(evs, g.startTunnel(a, i))
		}
	}
	return evs, won
}
