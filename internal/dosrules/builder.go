package dosrules

import "slices"

// Construction units build depots, half a depot per charge. Addresses in
// this file are the full build's (data segment 32f7). A unit whose live +6
// has bit 0x4000 (the construction unit's type word +0x11; tested before the
// sappers' 0x2000) gets the attack order on a depot site (Sites): with no
// charge left (+0x0a, a signed byte, not above 0) the selection shows
// message 0x4c, with no site message 0x4b (0ca6:1277-13bf). The order
// (0ad0:000d) sets the moved flag, +0x11 to the anchor and +0x13 = 0x0b, the
// first half, or 0x0d, the completion, where the anchor holds the half
// depot's N piece; a completion with ten depots on the map is refused
// (message 0x47). Either shows message 0x46 and stays off the order list.
//
// The work runs in the action pass with the captures and the sappers'
// (02dc:17b8, 02dc:1893), on the action side's map; the phase change copies
// it to the other side's.
//   - The first half lays the half depot's four pieces, provided all four
//     cells, as the OTHER side's map shows them, still have terrain sappers
//     work (flag 0x8000) and no unit; otherwise nothing happens and the
//     charge stays. A charge is spent.
//   - The completion checks nothing: the next depot record (HoldFlag100 plus
//     the number of depots, 15a4:0330) covers the four cells, owned by the
//     completing side, the anchor takes that side's N piece and the others
//     the depot's E, S and W pieces. A charge is spent. The record credits
//     no energy: the work clears the energy bytes +0x16/+0x17 the creator
//     set (the SHP record's energy, which a scenario depot gets at the level
//     load), and its +0x18 (the computer's goal weight) is 10.
//
// Either side may complete either side's half depot. Not modelled: the
// per-side depot count (32f7:4d3f) the completion raises, which nothing is
// known to read.

const (
	flagBuilder = 0x4000 // live +6: a construction unit
	actionHalf  = 0x0B   // +0x13: the first half of a depot
	actionDepot = 0x0D   // +0x13: the completion of a depot
	depotLimit  = holdTable

	// MsgDepot is the message of a construction unit's order.
	MsgDepot = 0x46
	// MsgDepotLimit refuses a completion with ten depots on the map.
	MsgDepotLimit = 0x47
	// MsgNoSite refuses a construction unit with no depot site.
	MsgNoSite = 0x4B
	// MsgNoCharge refuses a construction unit with no charge left.
	MsgNoCharge = 0x4C
)

// Depot terrain records (the constants file the game loads at 32f7:4c82),
// by the depot's pieces in the order N, E, S, W: the half depot (4c9a), the
// complete depot's E, S and W pieces (4c9f), and its N piece by owner
// (4c85: side 0, side 1).
var (
	halfDepot  = [4]int{21, 24, 22, 23}
	depotPiece = [4]int{-1, 20, 18, 19}
	depotN     = [2]int{16, 17}
)

// depotSites are the anchors a construction unit may build on, as (column,
// row) steps from its cell by its column parity (32f7:0029/0033 even,
// 003d/0047 odd): the depots that touch the unit without covering it.
var depotSites = [2][10][2]int{
	{{0, -2}, {1, -2}, {2, -1}, {2, 0}, {1, 0}, {0, 1}, {-1, 0}, {-2, 0}, {-2, -1}, {-1, -2}},
	{{0, -2}, {1, -1}, {2, -1}, {2, 0}, {1, 1}, {0, 1}, {-1, 1}, {-2, 0}, {-2, -1}, {-1, -1}},
}

// depotSteps are a depot's cells N, E, S, W from its anchor, by the anchor's
// column parity (32f7:0051/0055 even, 0059/005d odd; 0e47:1a8c).
var depotSteps = [2][4][2]int{
	{{0, 0}, {1, 0}, {0, 1}, {-1, 0}},
	{{0, 0}, {1, 1}, {0, 1}, {-1, 1}},
}

// Builder reports whether the unit builds depots instead of attacking.
func (u Live) Builder() bool { return u.Flags6&flagBuilder != 0 }

// charged reports whether a construction unit has a charge left.
func (u Live) charged() bool { return int8(byte(u.Hold)) > 0 }

// DepotCells are the cells of the depot at anchor, N, E, S, W; a cell off
// the map is -1.
func (g *Game) DepotCells(anchor int) []int {
	return g.depotCells(anchor%g.Width, anchor/g.Width)
}

// depotCells are the cells of the depot at column x, row y.
func (g *Game) depotCells(x, y int) []int {
	out := make([]int, 4)
	for k, s := range depotSteps[x&1] {
		cx, cy := x+s[0], y+s[1]
		out[k] = -1
		if cx >= 0 && cx < g.Width && cy >= 0 && cy < g.Height {
			out[k] = cy*g.Width + cx
		}
	}
	return out
}

// buildable reports whether a cell of side's map takes a depot piece:
// terrain sappers work and no unit.
func (g *Game) buildable(side, cell int) bool {
	return cell >= 0 && g.worked(g.Maps[side][2*cell]) && g.occupant(side, cell) > SlotFree
}

// Sites are the anchors construction unit of side may build on as the
// side's map shows it (0e47:1898): of the ten around it, those that hold a
// half depot's N piece, and those whose four cells are on the map with
// terrain sappers work and no unit. Charges do not count here. A unit that
// builds no depots has none.
func (g *Game) Sites(side, unit int) []int {
	u, ok := g.Units[unit]
	if !ok || !u.Builder() || side != 0 && side != 1 {
		return nil
	}
	from := u.Cell(side)
	if from < 0 || from >= g.Width*g.Height {
		return nil
	}
	x, y := from%g.Width, from/g.Width
	var out []int
	for _, s := range depotSites[x&1] {
		ax, ay := x+s[0], y+s[1]
		// The anchor's own record is read without a bounds check, so a
		// column off either edge reads the neighbouring row's cell.
		if a := ay*g.Width + ax; a >= 0 && a < g.Width*g.Height && g.Maps[side][2*a] == halfDepot[0] {
			out = append(out, a)
			continue
		}
		cells := g.depotCells(ax, ay)
		if !slices.ContainsFunc(cells, func(c int) bool { return !g.buildable(side, c) }) {
			out = append(out, cells[0])
		}
	}
	return out
}

// checkBuilder is Check for a construction unit's order on cell.
func (g *Game) checkBuilder(side int, u Live, unit, cell int) error {
	if !u.charged() {
		return refuse(MsgNoCharge, "no charge left")
	}
	sites := g.Sites(side, unit)
	switch {
	case len(sites) == 0:
		return refuse(MsgNoSite, "nowhere to build a depot")
	case !slices.Contains(sites, cell):
		return refuse(MsgNoTarget, "no depot site there")
	case g.Maps[side][2*cell] == halfDepot[0] && g.depots() >= depotLimit:
		return refuse(MsgDepotLimit, "no more depots")
	}
	return nil
}

// depots is the number of depots on the map (32f7:4cb6): the depot records.
func (g *Game) depots() int {
	n := 0
	for r := range g.Holds {
		if r >= HoldFlag100 && r < HoldFlag400 {
			n++
		}
	}
	return n
}

// orderDepot is the order of construction unit i of side on cell, its
// event.
func (g *Game) orderDepot(side, i, cell int) Event {
	u := g.Units[i]
	u.Flags4 |= 0x200
	u.Target = 2 * cell
	u.Action[0] = actionHalf
	if g.Maps[side][2*cell] == halfDepot[0] {
		u.Action[0] = actionDepot
	}
	g.Units[i] = u
	return Event{Kind: EvWork, Side: side, Unit: i, Cell: cell, Messages: []Message{{MsgDepot, side}}}
}

// halfDepotWork is the first half by construction unit i of side a on a's
// map; ok is false where the other side's map leaves the depot no room.
func (g *Game) halfDepotWork(a, i int) (Event, bool) {
	u := g.Units[i]
	anchor := u.Target / 2
	cells := g.DepotCells(anchor)
	for _, c := range cells {
		if !g.buildable(1-a, c) {
			return Event{}, false
		}
	}
	for k, c := range cells {
		g.Maps[a][2*c] = halfDepot[k]
	}
	u.Hold--
	g.Units[i] = u
	return Event{Kind: EvHalfDepot, Side: a, Unit: i, Cell: anchor, Cells: cells}, true
}

// depotWork is the completion by construction unit i of side a on a's map.
func (g *Game) depotWork(a, i int) Event {
	u := g.Units[i]
	p := u.Target
	r := HoldFlag100 + g.depots()
	// 15a4:0330: the cells N, W, S, E, the W and E steps by the parity of the
	// cell number (the column's where the map's width is even).
	h := EmptyHold()
	row := 2 * g.Width
	h.Cells = [4]int{p, p - 2, p + row, p + 2}
	if (p/2)&1 != 0 {
		h.Cells[1], h.Cells[3] = p+row-2, p+row+2
	}
	h.Flags, h.Occupant = a|KindDepot, 0xFF
	g.Holds[r] = h
	if g.Buildings == nil {
		g.Buildings = map[int]Building{}
	}
	g.Buildings[r] = Building{}
	cells := g.DepotCells(p / 2)
	for k, c := range cells {
		if c < 0 {
			continue
		}
		rec := depotPiece[k]
		if k == 0 {
			rec = depotN[a]
		}
		g.Maps[a][2*c] = rec
	}
	u.Hold--
	g.Units[i] = u
	return Event{Kind: EvDepot, Side: a, Unit: i, Cell: p / 2, Cells: cells, Hold: r}
}
