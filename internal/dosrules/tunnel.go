package dosrules

import (
	"maps"
	"slices"
)

// Tunnellers (live +6 bit 0x0400, an extension type) dig a mine under a
// neighbouring cell and blow it. The attack order on any neighbouring cell
// on the map that is no water starts the tunnel, with or without a unit on
// it, unless the unit digs one already. The order sets the moved flag, +0x11
// to the cell and +0x13 = actionTunnel; the work (work) after the battles
// opens the tunnel: Game.Tunnels holds it until it blows.
//
// Once per game turn, with the gas (turnEffects), every tunnel whose
// tunnellers still stand on the cell it started from gains a turn; the
// others are abandoned. A tunnel with TunnelTurns blows: every unit on the
// mined cell that is no aircraft and not held in a building or transporter
// loses TunnelLoss percent of its full group, at least 1, own units too; a
// unit left with no strength is destroyed. Where the cell's terrain is
// worked ground, both sides' maps take TerrainCrater, which costs more to
// cross and shelters like a trench.

const (
	flagTunnel   = 0x0400 // live +6: tunnellers (extension types only)
	actionTunnel = 0x17   // +0x13: the tunnellers' order

	// TunnelTurns are the game turns a tunnel takes to blow.
	TunnelTurns = 2
	// TunnelLoss is the loss of a unit on a blown mine, in percent of its
	// full group.
	TunnelLoss = 70

	// MsgTunnel is the message of the tunnellers' order, MsgTunnelBusy
	// refuses tunnellers that dig a tunnel already.
	MsgTunnel     = 0x104
	MsgTunnelBusy = 0x105
)

// Tunnel is a mine being dug: the tunnellers (Unit) of Side dig from cell
// From under the neighbouring Cell and have dug Dug turns.
type Tunnel struct {
	Unit int `json:"unit"`
	Side int `json:"side"`
	From int `json:"from"`
	Cell int `json:"cell"`
	Dug  int `json:"dug"`
}

// Tunneller reports whether the unit digs mines instead of attacking.
func (u Live) Tunneller() bool { return u.Flags6&flagTunnel != 0 }

// TunnelOf is the tunnel unit digs and whether it digs one.
func (g *Game) TunnelOf(unit int) (Tunnel, bool) {
	for _, t := range g.Tunnels {
		if t.Unit == unit {
			return t, true
		}
	}
	return Tunnel{}, false
}

// mineable reports whether tunnellers may dig under cell: on the map and no
// water.
func (g *Game) mineable(cell int) bool {
	t, err := g.Profile.moveTables()
	if err != nil || cell < 0 || cell >= g.Width*g.Height {
		return false
	}
	rec := g.Maps[0][2*cell]
	return rec < len(t.Terrains) && t.Terrains[rec].Mask&0x80 == 0
}

// checkTunnel is Check for the tunnellers' attack order.
func (g *Game) checkTunnel(side int, u Live, idx, cell int) error {
	if _, busy := g.TunnelOf(idx); busy {
		return refuse(MsgTunnelBusy, "the tunnellers dig a tunnel already")
	}
	n := neighbours(u.Cell(side), g.Width, g.Height)
	if !slices.Contains(n[:], cell) || !g.mineable(cell) {
		return refuse(MsgNoTarget, "no tunnel there")
	}
	return nil
}

// startTunnel is the work of tunnellers i of side a: the tunnel opens.
func (g *Game) startTunnel(a, i int) Event {
	u := g.Units[i]
	t := Tunnel{Unit: i, Side: a, From: u.Cell(a), Cell: u.Target / 2}
	g.Tunnels = append(slices.DeleteFunc(g.Tunnels, func(x Tunnel) bool { return x.Unit == i }), t)
	return Event{Kind: EvTunnel, Side: a, Unit: i, Cell: t.Cell}
}

// mineTurn is the tunnels' turn: abandoned tunnels close, the others gain a
// turn, and those with TunnelTurns blow. Units the mines destroy join gone.
func (g *Game) mineTurn(gone *[]int) []Event {
	var evs []Event
	var left []Tunnel
	for _, t := range g.Tunnels {
		u, ok := g.Units[t.Unit]
		if !ok || u.Destroyed() || u.Action[1]&4 != 0 || gasCell(u) != t.From || ownerOf(u) != t.Side {
			evs = append(evs, Event{Kind: EvTunnelLost, Side: t.Side, Unit: t.Unit, Cell: t.Cell})
			continue
		}
		t.Dug++
		if t.Dug < TunnelTurns {
			left = append(left, t)
			continue
		}
		evs = append(evs, g.blow(t, gone)...)
	}
	g.Tunnels = left
	return evs
}

// blow is the mine of tunnel t going up: the crater, then the losses.
func (g *Game) blow(t Tunnel, gone *[]int) []Event {
	ev := Event{Kind: EvMine, Side: t.Side, Unit: t.Unit, Cell: t.Cell}
	for k := range g.Maps {
		if g.worked(g.Maps[k][2*t.Cell]) {
			g.Maps[k][2*t.Cell] = TerrainCrater
			ev.Cells = []int{t.Cell}
		}
	}
	evs := []Event{ev}
	for _, i := range slices.Sorted(maps.Keys(g.Units)) {
		u := g.Units[i]
		if u.Destroyed() || u.Action[1]&4 != 0 || u.Flags4&0x10 != 0 || gasCell(u) != t.Cell {
			continue
		}
		group := g.Profile.Tables.Types[u.Type].Group
		loss := min(max(1, (group*TunnelLoss+99)/100), u.Size)
		u.Size -= loss
		if u.Size == 0 {
			u.Action[1] |= 4
			*gone = append(*gone, i)
		}
		g.Units[i] = u
		evs = append(evs, Event{Kind: EvMined, Side: ownerOf(u), Unit: i, Cell: t.Cell, Loss: loss})
	}
	return evs
}
