package dosrules

import "slices"

// Energy. Each side has one energy account that its buildings pay into every
// command cycle and that any of its buildings draws on: a repair costs 15
// energy, 50 for a construction unit, restores the group size, keeps the
// experience and resets a construction unit's charges to 2; a factory
// builds a unit of its build list for the type's price (UNIT.DAT +0x3f,
// 255 never sold). What a building pays and what it builds are the
// scenario's (Game.Buildings). Repairs and builds are given in the side's
// action phase, as the manual's action cycle; a new unit leaves the building
// from the next movement phase on.
//
// Assumed, not verified:
//   - Each side is credited after every phase change, with the credits of
//     every building record it owns (not neutral).
//   - Repairs and builds are given in an own building; they mark the unit
//     moved (one order per unit and phase) but do not enter the order list,
//     so the move limit does not count them.
//   - A repair leaves the alternate strength as it is.
//   - A new unit takes the lowest free live record, the side's highest
//     serial plus one, a free slot of the side and the building's first
//     cell as its position, as an arrival into an own building leaves it;
//     a construction unit starts with 2 charges. A transporter takes as
//     +0x0a the lowest transporter record (below HoldFlag40) that none
//     holds, as a scenario's transporter gets one: flag 0x20 with the
//     classes its type carries and its side, the building's first cell as
//     its position (2*cell+1) and itself as the occupant.

// Building is what the scenario gives a building record: the energy it
// credits each cycle and the types it builds (none: it only repairs).
type Building struct {
	Credits int   `json:"credits"`
	Builds  []int `json:"builds,omitempty"`
}

// builderCharges are a construction unit's charges after a repair.
const builderCharges = 2

// builder reports whether typ is a construction unit.
func (p Profile) builder(typ int) bool { return slices.Contains(p.Builders, typ) }

// repairCost is the energy a repair of a unit of typ costs.
func (p Profile) repairCost(typ int) int {
	if p.builder(typ) {
		return p.BuilderRepair
	}
	return p.Repair
}

// checkBuilding is Check for a repair or a build.
func (g *Game) checkBuilding(c Command) error {
	if g.Players[c.Side].Phase != PhaseAction {
		return refuse(0, "repairs and builds are given in the action phase")
	}
	b, ok := g.Buildings[c.Hold]
	h, held := g.Holds[c.Hold]
	switch {
	case !ok || !held:
		return refuse(0, "no such building")
	case h.Flags&2 != 0 || !Owns(h.Flags, c.Side):
		return refuse(0, "not the side's building")
	}
	if c.Kind == CmdRepair {
		if c.Slot < 0 || c.Slot >= 7 {
			return refuse(0, "no such slot")
		}
		u, ok := g.Units[h.Slots[7*c.Side+c.Slot]]
		switch {
		case !ok:
			return refuse(0, "the slot is empty")
		case u.Flags4&2 != 0 || !Owns(u.Flags4, c.Side):
			return refuse(0, "not the side's unit")
		case u.Flags4&0x200 != 0:
			return refuse(0, "the unit has its order")
		case g.Energy[c.Side] < g.Profile.repairCost(u.Type):
			return refuse(0, "not enough energy")
		}
		return nil
	}
	types := g.Profile.Tables.Types
	switch {
	case !slices.Contains(b.Builds, c.Type) || c.Type < 0 || c.Type >= len(types) || types[c.Type].Cost == 0xFF:
		return refuse(0, "the building does not build the type")
	case g.Energy[c.Side] < types[c.Type].Cost:
		return refuse(0, "not enough energy")
	case freeSlot(h, c.Side) < 0:
		return refuse(0, "no free slot")
	case g.freeRecord() < 0:
		return refuse(0, "no free unit record")
	case g.Profile.Tables.Move.Types[c.Type].Class&0x1000 != 0 && g.freeCarrier() < 0:
		return refuse(0, "no free container record")
	}
	return nil
}

// freeSlot is the first free slot of side's part of h, -1 if none.
func freeSlot(h Hold, side int) int {
	for k := range 7 {
		if h.Slots[7*side+k] > SlotFree {
			return k
		}
	}
	return -1
}

// freeRecord is the lowest live record without a unit, -1 if none.
func (g *Game) freeRecord() int {
	for i := range SlotFree {
		if _, ok := g.Units[i]; !ok {
			return i
		}
	}
	return -1
}

// freeCarrier is the lowest transporter record that none holds, -1 if none.
func (g *Game) freeCarrier() int {
	for i := range HoldFlag40 {
		if h, ok := g.Holds[i]; !ok || h == EmptyHold() {
			return i
		}
	}
	return -1
}

// repair carries out a checked repair.
func (g *Game) repair(c Command) []Event {
	i := g.Holds[c.Hold].Slots[7*c.Side+c.Slot]
	u := g.Units[i]
	g.Energy[c.Side] -= g.Profile.repairCost(u.Type)
	u.Size = g.Profile.Tables.Types[u.Type].Group
	if g.Profile.builder(u.Type) {
		u.Hold = builderCharges
	}
	u.Flags4 |= 0x200
	g.Units[i] = u
	return []Event{{Kind: EvRepaired, Side: c.Side, Unit: i, Cell: g.Holds[c.Hold].Cells[0] / 2, Energy: g.Energy[c.Side]}}
}

// build carries out a checked build.
func (g *Game) build(c Command) []Event {
	m := g.Profile.Tables.Move.Types[c.Type]
	h := g.Holds[c.Hold]
	i, k := g.freeRecord(), freeSlot(h, c.Side)
	u := Live{
		Budget: m.Budget, Size: g.Profile.Tables.Types[c.Type].Group, Flags4: m.Class | c.Side | 0x4200,
		Flags6: m.Flags, Type: c.Type, Serial: g.serial(c.Side), Hold: 0xFF, Pos: [2]int{h.Cells[0], h.Cells[0]},
	}
	if g.Profile.builder(c.Type) {
		u.Hold = builderCharges
	}
	if m.Class&0x1000 != 0 {
		u.Hold = g.freeCarrier()
		r := EmptyHold()
		r.Cells[0], r.Flags, r.Occupant = h.Cells[0]+1, 0x20|m.Carries|c.Side, i
		g.Holds[u.Hold] = r
	}
	g.Units[i] = u
	h.Slots[7*c.Side+k] = i
	g.Holds[c.Hold] = h
	if at := h.Cells[0] + 1; g.Maps[c.Side][at] == 0xFF {
		g.Maps[c.Side][at] = InBuilding
	}
	g.Energy[c.Side] -= g.Profile.Tables.Types[c.Type].Cost
	return []Event{{Kind: EvBuilt, Side: c.Side, Unit: i, Cell: h.Cells[0] / 2, Energy: g.Energy[c.Side]}}
}

// Income is the energy side's buildings pay it at a phase change: the
// credits of every building record it owns.
func (g *Game) Income(side int) int {
	paid := 0
	for i, b := range g.Buildings {
		if h, ok := g.Holds[i]; ok && h.Flags&2 == 0 && Owns(h.Flags, side) {
			paid += b.Credits
		}
	}
	return paid
}

// credit pays each side its Income after the phase change ph; a side
// credited nothing gets no event.
func (g *Game) credit(ph Event) []Event {
	var evs []Event
	for k := range 2 {
		paid := g.Income(k)
		if paid == 0 {
			continue
		}
		g.Energy[k] += paid
		evs = append(evs, Event{Kind: EvCredit, Side: k, Energy: g.Energy[k], Phase: ph.Phase, Turn: ph.Turn})
	}
	return evs
}
