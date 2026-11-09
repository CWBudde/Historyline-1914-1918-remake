package dosrules

import (
	"maps"
	"slices"
)

// Gas pioneers release a lasting gas cloud. They are an extension type
// (extension.json, docs/new-units.md), not the original's: the original
// tables have no type with live +6 bit 0x8000.
//
// A gas pioneer gets the attack order on any neighbouring cell, with or
// without a unit on it, while it has a charge (live +0x0a, as a
// construction unit's; a repair refills them). The order sets the moved
// flag, +0x11 to the cell and +0x13 = actionGas and stays off the order
// list, like a sapper's. The release runs in the action pass after the
// battles (work): the cloud covers the three cells of a sapper's work on
// that neighbour (trenchWork) that lie on the map, whatever their terrain,
// each set to the game's gas duration, and a charge is spent.
//
// Once per game turn, in the resolution whose phase change counts the turn
// (side 0's action phase), every unit on the map standing on a gassed cell
// loses part of its full group, own units too: foot units (live +6 bits
// 0x01 or 0x40: infantry, cavalry, sappers and the guns pushed by hand) the
// game's rate, other ground units and ships a third of it, air units
// nothing; at least 1. A unit held in a building or a transporter is safe.
// A unit left with no strength is destroyed by the phase change that
// follows. Then each gassed cell thins by one turn.

const (
	flagGas   = 0x8000 // live +6: a gas pioneer (extension types only)
	actionGas = 0x15   // +0x13: the gas pioneers' release order

	// GasCharges are a gas pioneer's charges when placed, built or repaired.
	GasCharges = 2
	// GasTurns is the default duration of a cloud in game turns.
	GasTurns = 2
	// GasLoss is the default loss of a foot unit per turn, in percent of its
	// full group.
	GasLoss = 30

	// MsgGas is the message of a gas pioneer's release order (an extension
	// code above the original's message bytes).
	MsgGas = 0x100
	// MsgNoGas refuses a gas pioneer with no charge left.
	MsgNoGas = 0x101
)

// GasPioneer reports whether the unit releases gas instead of attacking.
func (u Live) GasPioneer() bool { return u.Flags6&flagGas != 0 }

// Releasing reports whether the unit has a gas pioneer's release order this
// phase.
func (u Live) Releasing() bool { return u.Action[0] == actionGas }

// gasType reports whether typ is a gas pioneer.
func (p Profile) gasType(typ int) bool {
	t := p.Tables.Move
	return t != nil && typ >= 0 && typ < len(t.Types) && t.Types[typ].Flags&flagGas != 0
}

// charges are the charges a unit of typ holds in live +0x0a when placed,
// built or repaired, and whether it holds any.
func (p Profile) charges(typ int) (int, bool) {
	switch {
	case p.builder(typ):
		return builderCharges, true
	case p.gasType(typ):
		return GasCharges, true
	}
	return 0, false
}

// gasTurns is the game's cloud duration.
func (g *Game) gasTurns() int {
	if g.GasTurns > 0 {
		return g.GasTurns
	}
	return GasTurns
}

// gasLoss is the game's foot rate in percent.
func (g *Game) gasLoss() int {
	if g.GasLoss > 0 {
		return g.GasLoss
	}
	return GasLoss
}

// checkGas is Check for a gas pioneer's attack order.
func (g *Game) checkGas(side int, u Live, cell int) error {
	if !u.charged() {
		return refuse(MsgNoGas, "no gas left")
	}
	n := neighbours(u.Cell(side), g.Width, g.Height)
	if !slices.Contains(n[:], cell) {
		return refuse(MsgNoTarget, "gas is released on a neighbouring cell")
	}
	return nil
}

// GasCells are the cells a cloud released by unit of side on cell would
// cover, ascending in the order of the work; nil where the unit may not
// release there.
func (g *Game) GasCells(side, unit, cell int) []int {
	u, ok := g.Units[unit]
	if !ok || side != 0 && side != 1 || cell < 0 || cell >= g.Width*g.Height || !u.GasPioneer() {
		return nil
	}
	n := neighbours(u.Cell(side), g.Width, g.Height)
	if !slices.Contains(n[:], cell) {
		return nil
	}
	return g.gasCells(u.Cell(side), cell)
}

// gasCells are the cells of a sapper's work from cell from on neighbour to
// that lie on the map.
func (g *Game) gasCells(from, to int) []int {
	n := neighbours(from, g.Width, g.Height)
	k := slices.Index(n[:], to)
	if k < 0 {
		return nil
	}
	x, y := from%g.Width, from/g.Width
	var cells []int
	for _, c := range trenchWork[x&1][k] {
		cx, cy := x+c.dx, y+c.dy
		if cx >= 0 && cx < g.Width && cy >= 0 && cy < g.Height {
			cells = append(cells, cy*g.Width+cx)
		}
	}
	return cells
}

// orderGas carries out a checked release order.
func (g *Game) orderGas(side, unit, cell int) Event {
	u := g.Units[unit]
	u.Flags4 |= 0x200
	u.Target = 2 * cell
	u.Action[0] = actionGas
	g.Units[unit] = u
	return Event{Kind: EvWork, Side: side, Unit: unit, Cell: cell, Messages: []Message{{MsgGas, side}}}
}

// releaseGas is the release of gas pioneer i of side a: the cloud's cells
// take the full duration and a charge is spent.
func (g *Game) releaseGas(a, i int) Event {
	u := g.Units[i]
	to := u.Target / 2
	cells := g.gasCells(u.Cell(a), to)
	if g.Gas == nil {
		g.Gas = make([]int, g.Width*g.Height)
	}
	for _, c := range cells {
		g.Gas[c] = g.gasTurns()
	}
	if u.charged() {
		u.Hold--
	}
	g.Units[i] = u
	return Event{Kind: EvGasCloud, Side: a, Unit: i, Cell: to, Cells: cells}
}

// GasAt is the number of turns the cloud on cell has left, 0 for none.
func (g *Game) GasAt(cell int) int {
	if cell < 0 || cell >= len(g.Gas) {
		return 0
	}
	return g.Gas[cell]
}

// gasLossOf is the loss of unit u per turn on a gassed cell.
func (g *Game) gasLossOf(u Live) int {
	if u.Flags4&0x10 != 0 {
		return 0
	}
	pct := g.gasLoss()
	if u.Flags6&0x41 == 0 {
		pct /= 3
	}
	group := g.Profile.Tables.Types[u.Type].Group
	return max(1, (group*pct+99)/100)
}

// gasCell is the cell a unit stands on as its own side sees it (side 0's
// view for a neutral one), -1 for a held unit.
func gasCell(u Live) int {
	if u.Flags4&0x4000 != 0 {
		return -1
	}
	if u.Flags4&1 != 0 {
		return u.Cell(1)
	}
	return u.Cell(0)
}

// gasTurn is the turn's gas in side a's resolution: the losses of the units
// on gassed cells, then the thinning. It runs only when the coming phase
// change counts the turn.
func (g *Game) gasTurn() []Event {
	if g.Gas == nil || g.Players[0].Phase != PhaseAction {
		return nil
	}
	var evs []Event
	var gone, cleared []int
	for _, i := range slices.Sorted(maps.Keys(g.Units)) {
		u := g.Units[i]
		if u.Destroyed() || u.Flags4&0xC0 == 0x40 {
			continue // a ship's rear goes with its head
		}
		on := g.GasAt(gasCell(u)) > 0
		rear, ship := g.Units[i+1]
		ship = ship && u.Flags4&0xC0 == 0xC0 && rear.Flags4&0xC0 == 0x40
		if ship && !on {
			on = g.GasAt(gasCell(rear)) > 0
		}
		loss := g.gasLossOf(u)
		if !on || loss == 0 {
			continue
		}
		loss = min(loss, u.Size)
		u.Size -= loss
		if u.Size == 0 {
			u.Action[1] |= 4
			gone = append(gone, i)
		}
		g.Units[i] = u
		if ship {
			rear.Size = u.Size
			rear.Action[1] |= u.Action[1] & 4
			g.Units[i+1] = rear
		}
		evs = append(evs, Event{Kind: EvGassed, Side: ownerOf(u), Unit: i, Cell: gasCell(u), Loss: loss})
	}
	left := false
	for c, t := range g.Gas {
		if t <= 0 {
			continue
		}
		g.Gas[c] = t - 1
		if t == 1 {
			cleared = append(cleared, c)
		} else {
			left = true
		}
	}
	if !left {
		g.Gas = nil
	}
	g.vacate(gone)
	if len(cleared) > 0 {
		evs = append(evs, Event{Kind: EvGasCleared, Side: -1, Cells: cleared})
	}
	return evs
}

// ownerOf is the side a unit belongs to: 0, 1 or 2 for neutral.
func ownerOf(u Live) int {
	switch {
	case u.Flags4&1 != 0:
		return 1
	case u.Flags4&2 != 0:
		return 2
	}
	return 0
}
