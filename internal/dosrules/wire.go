package dosrules

import "slices"

// Extension work units (extension.json, docs/new-units.md) beyond the gas
// pioneers (gas.go): the flamethrower, the wiring party, the tunnellers
// (tunnel.go) and the field ambulance (ambulance.go). Their abilities ride
// on live +6 bits the original's types leave unused.
//
// A flamethrower (bit 0x0800) attacks like any unit, but the defender's
// terrain lends it no defence: a positive defence term counts as 0
// (Context.Flame).
//
// A wiring party (bit 0x1000) gets the attack order on a neighbouring cell
// whose terrain a sapper works (flag 0x8000) and that is neither a trench nor
// wire yet, with or without a unit on it, while it has a charge (live +0x0a,
// as a construction unit's; a repair refills them). The order sets the moved
// flag, +0x11 to the cell and +0x13 = actionWire, like a sapper's. The work
// runs in the action pass after the battles (work) on the action side's map:
// of the three cells of a sapper's work on that neighbour (trenchWork), each
// such cell takes TerrainWire, and a charge is spent. Wire plays like a
// trench for who may enter (terrain mask 0x08: foot units, heavy tanks and
// light guns, not cavalry, cars or light tanks) but costs a whole foot move
// and shelters nobody; sappers clear it as they fill a trench.

const (
	flagFlame  = 0x0800 // live +6: a flamethrower (extension types only)
	flagWire   = 0x1000 // live +6: a wiring party
	actionWire = 0x16   // +0x13: the wiring party's work order

	// TerrainWire is the extension terrain record of barbed wire on open
	// ground, TerrainCrater that of a mine's crater (tunnel.go).
	TerrainWire   = 0xAF
	TerrainCrater = 0xB0

	// WireCharges are a wiring party's charges when placed, built or
	// repaired.
	WireCharges = 2

	// MsgWire is the message of a wiring party's order, MsgNoWire refuses one
	// with no charge left (extension codes, like MsgGas).
	MsgWire   = 0x102
	MsgNoWire = 0x103
)

// Flamethrower reports whether the unit's attacks ignore the defender's
// terrain defence.
func (u Live) Flamethrower() bool { return u.Flags6&flagFlame != 0 }

// Wiring reports whether the unit lays barbed wire instead of attacking.
func (u Live) Wiring() bool { return u.Flags6&flagWire != 0 }

// Wire reports whether a terrain record is barbed wire.
func Wire(rec int) bool { return rec == TerrainWire }

// typeFlag reports whether type typ has live +6 bit flag.
func (p Profile) typeFlag(typ, flag int) bool {
	t := p.Tables.Move
	return t != nil && typ >= 0 && typ < len(t.Types) && t.Types[typ].Flags&flag != 0
}

// wireable reports whether a wiring party lays wire on a terrain record:
// worked ground that is no trench and no wire yet.
func (g *Game) wireable(rec int) bool {
	return g.worked(rec) && !Trench(rec) && !Wire(rec)
}

// checkWire is Check for a wiring party's attack order.
func (g *Game) checkWire(side int, u Live, cell int) error {
	if !u.charged() {
		return refuse(MsgNoWire, "no wire left")
	}
	n := neighbours(u.Cell(side), g.Width, g.Height)
	if !slices.Contains(n[:], cell) || !g.wireable(g.Maps[side][2*cell]) {
		return refuse(MsgNoTarget, "no wire work there")
	}
	return nil
}

// WireCells are the cells a wiring party unit of side would wire from its
// cell on cell as the side's map shows it now, ascending in the order of
// the work; nil where it may not.
func (g *Game) WireCells(side, unit, cell int) []int {
	u, ok := g.Units[unit]
	if !ok || side != 0 && side != 1 || cell < 0 || cell >= g.Width*g.Height || !u.Wiring() {
		return nil
	}
	n := neighbours(u.Cell(side), g.Width, g.Height)
	if !slices.Contains(n[:], cell) || !g.wireable(g.Maps[side][2*cell]) {
		return nil
	}
	return g.wireCells(side, u.Cell(side), cell)
}

// wireCells are the cells of a sapper's work from from on to that take wire
// on side's map.
func (g *Game) wireCells(side, from, to int) []int {
	var out []int
	for _, c := range g.gasCells(from, to) {
		if g.wireable(g.Maps[side][2*c]) {
			out = append(out, c)
		}
	}
	return out
}

// orderWork carries out a checked work order with action byte action and
// message msg (the wiring party's, the tunnellers').
func (g *Game) orderWork(side, unit, cell, action, msg int) Event {
	u := g.Units[unit]
	u.Flags4 |= 0x200
	u.Target = 2 * cell
	u.Action[0] = action
	g.Units[unit] = u
	return Event{Kind: EvWork, Side: side, Unit: unit, Cell: cell, Messages: []Message{{msg, side}}}
}

// layWire is the work of wiring party i of side a on a's map.
func (g *Game) layWire(a, i int) Event {
	u := g.Units[i]
	to := u.Target / 2
	cells := g.wireCells(a, u.Cell(a), to)
	for _, c := range cells {
		g.Maps[a][2*c] = TerrainWire
	}
	if u.charged() {
		u.Hold--
	}
	g.Units[i] = u
	return Event{Kind: EvWired, Side: a, Unit: i, Cell: to, Cells: cells}
}
