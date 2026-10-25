package dosplay

import (
	"fmt"
	"maps"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// UnitView is a unit as a side sees it.
type UnitView struct {
	Index  int    // live record
	Type   int    // UNIT.DAT type
	ID     string // unit id of data.Set.Units
	Name   string
	Domain Domain // the kind of fighting
	Side   int    // 0, 1, 2 neutral
	Owner  game.Player
	// At is the hex the unit stands on, for a held unit its container's.
	At                                   hexgrid.Coord
	Size, Full, Alt, Exp, Budget, Armour int
	Moved                                bool // the unit has its order this phase
	Held                                 bool // in a building or a transporter
	Carrier                              bool // a transporter
	Destroyed                            bool
	Hold                                 int // a transporter's own container record, else -1
	// In and Slot are the container record and slot of a held unit, else
	// -1.
	In, Slot int
	// Target is the hex of the unit's queued attack or sapper's work; set
	// for the viewing side's own units.
	Target *hexgrid.Coord
	// A two-cell ship is two records: Rear is set on its head, the hex of
	// its rear, and Head on the rear, the hex of its head.
	Rear, Head *hexgrid.Coord
}

// BuildingView is a building: its container record, owner, kind, whether
// it is an HQ, and the units the viewing side's slots list.
type BuildingView struct {
	Hold  int
	Owner game.Player
	Kind  BuildingKind
	HQ    bool
	Units []UnitView
}

// View is what a side sees: the occupants of its map by hex and the
// buildings by entrance. Side is -1 for TrueView.
type View struct {
	Side      int
	Units     map[hexgrid.Coord]UnitView
	Buildings map[hexgrid.Coord]BuildingView
}

// Coords are the hexes of a set in line, then column order.
func Coords(set map[hexgrid.Coord]bool) []hexgrid.Coord {
	out := make([]hexgrid.Coord, 0, len(set))
	for c, ok := range set {
		if ok {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b hexgrid.Coord) int {
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return a.Col - b.Col
	})
	return out
}

// sideOf is the side the side bits of a live +4 word or a container
// record's flags give: 1 for bit 1, 2 (neutral) for bit 2, else 0.
func sideOf(flags int) int {
	switch {
	case flags&1 != 0:
		return 1
	case flags&2 != 0:
		return 2
	}
	return 0
}

// unitID is the unit id of a live record: the level's, its head's for a ship's
// rear, or for a built unit the first id of its type and side.
func (s *Session) unitID(idx int) string {
	if id, ok := s.Board.IDs[idx]; ok {
		return id
	}
	if id, ok := s.ids[idx]; ok {
		return id
	}
	u := s.Game.Units[idx]
	if u.Flags4&0xC0 == 0x40 {
		// A two-cell ship's rear is its head's unit.
		return s.unitID(idx - 1)
	}
	id := s.typeID(u.Type, sideOf(u.Flags4))
	s.ids[idx] = id
	return id
}

// typeID is the first unit id of UNIT.DAT type typ and side (any side for
// 2), "" for a type the remake has no unit for yet.
func (s *Session) typeID(typ, side int) string {
	for _, id := range slices.Sorted(maps.Keys(s.set.Units)) {
		d := s.set.Units[id]
		if d.Dos == typ && (side == 2 || dosbind.Side(game.FactionOf(d.Country)) == side) {
			return id
		}
	}
	return ""
}

// typeName is the name of unit id, for "" a placeholder naming type typ.
func (s *Session) typeName(id string, typ int) string {
	if id == "" {
		return fmt.Sprintf("Type %d", typ)
	}
	return s.set.Units[id].Name
}

// unitView is unit idx as side (0 or 1) sees its position; Target is not
// set.
func (s *Session) unitView(idx, side int) UnitView {
	u := s.Game.Units[idx]
	id := s.unitID(idx)
	t := s.Game.Profile.Tables.Types[u.Type]
	v := UnitView{
		Index: idx, Type: u.Type, ID: id, Name: s.typeName(id, u.Type), Side: sideOf(u.Flags4),
		At: s.coord(u.Pos[side] / 2), Size: u.Size, Full: t.Group, Alt: u.Alt, Exp: u.Exp, Budget: u.Budget, Armour: t.Armour,
		Moved: u.Flags4&0x200 != 0, Held: u.Flags4&0x4000 != 0, Carrier: u.Flags4&0x1000 != 0, Destroyed: u.Destroyed(),
		Hold: -1, In: -1, Slot: -1,
	}
	v.Owner, v.Domain = s.Faction(v.Side), TypeDomain(s.Game.Profile.Tables, u.Type)
	if v.Carrier {
		v.Hold = u.Hold
	}
	if v.Held {
		if h, k, ok := s.container(side, idx); ok {
			v.In, v.Slot = h, k
		}
	}
	return v
}

// target is the hex of side's queued attack or sapper's work of unit, nil
// if none.
func (s *Session) target(side, unit int) *hexgrid.Coord {
	if side != 0 && side != 1 || s.Phase(side) != dosrules.PhaseAction ||
		!slices.Contains(s.Game.Players[side].Orders, unit) && !s.Game.Units[unit].Working() {
		return nil
	}
	c := s.coord(s.Game.Units[unit].Target / 2)
	return &c
}

// Unit is a unit by live record, at the position its own side sees (side 0's
// for a neutral one), without its target.
func (s *Session) Unit(idx int) UnitView {
	if _, ok := s.Game.Units[idx]; !ok {
		return UnitView{Index: idx, Hold: -1, In: -1, Slot: -1, Destroyed: true}
	}
	return s.unitView(idx, sideOf(s.Game.Units[idx].Flags4)&1)
}

// View is what side sees: the units its map shows on the level's hexes, and
// the buildings with the units its slots list.
func (s *Session) View(side int) View {
	v := View{Side: side, Units: map[hexgrid.Coord]UnitView{}, Buildings: s.buildings(side)}
	m := s.Game.Maps[side]
	for c := range s.Game.Width * s.Game.Height {
		occ := m[2*c+1]
		p, ok := s.Board.Coord(c)
		if _, live := s.Game.Units[occ]; occ > dosrules.SlotFree || !live || !ok {
			continue
		}
		u := s.unitView(occ, side)
		u.At = p
		if u.Side == side {
			u.Target = s.target(side, occ)
		}
		v.Units[p] = u
	}
	s.linkShips(v.Units)
	return v
}

// TrueView shows every unit where its own side sees it (side 0's position
// for a neutral one) and each building with its owner's slots.
func (s *Session) TrueView() View {
	v := View{Side: -1, Units: map[hexgrid.Coord]UnitView{}, Buildings: s.buildings(-1)}
	for _, i := range slices.Sorted(maps.Keys(s.Game.Units)) {
		u := s.Game.Units[i]
		if u.Destroyed() || u.Flags4&0x4000 != 0 {
			continue
		}
		side := sideOf(u.Flags4) & 1
		cell := u.Cell(side)
		if !s.bound(cell) {
			continue
		}
		uv := s.unitView(i, side)
		uv.Target = s.target(side, i)
		v.Units[uv.At] = uv
	}
	s.linkShips(v.Units)
	return v
}

// buildings are the building records by entrance with side's slots (-1:
// the owner's, side 0's for a neutral building).
func (s *Session) buildings(side int) map[hexgrid.Coord]BuildingView {
	out := map[hexgrid.Coord]BuildingView{}
	for _, i := range slices.Sorted(maps.Keys(s.Game.Holds)) {
		if i < dosrules.HoldFlag40 {
			continue
		}
		h := s.Game.Holds[i]
		p, ok := s.Board.Coord(h.Cells[0] / 2)
		if !ok {
			continue
		}
		owner := sideOf(h.Flags)
		k := side
		if k < 0 {
			k = owner & 1
		}
		out[p] = BuildingView{Hold: i, Owner: s.Faction(owner), Kind: kindOf(i), HQ: kindOf(i) == HQ, Units: s.slots(i, k)}
	}
	return out
}

// Slots is the number of slots a side has in a container record.
const Slots = 7

// SlotView is one of a side's slots of a container record: its index and
// the unit in it, nil for an empty slot.
type SlotView struct {
	Slot int
	Unit *UnitView
}

// slotViews are side's slots of container record hold, in slot order.
func (s *Session) slotViews(hold, side int) []SlotView {
	h := s.Game.Holds[hold]
	out := make([]SlotView, Slots)
	for k := range Slots {
		out[k].Slot = k
		u := h.Slots[Slots*side+k]
		if _, ok := s.Game.Units[u]; u > dosrules.SlotFree || !ok {
			continue
		}
		v := s.unitView(u, side)
		v.At, v.Held, v.In, v.Slot = s.coord(startCell(h)), true, hold, k
		out[k].Unit = &v
	}
	return out
}

// slots are the units of side's slots of container record hold, in slot
// order.
func (s *Session) slots(hold, side int) []UnitView {
	var out []UnitView
	for _, v := range s.slotViews(hold, side) {
		if v.Unit != nil {
			out = append(out, *v.Unit)
		}
	}
	return out
}

// Contents is the container record and the occupied slots of side's own
// building or own transporter on a hex.
func (s *Session) Contents(side int, c hexgrid.Coord) (hold int, slots []UnitView, ok bool) {
	if hold, ok = s.holdOf(side, c); !ok {
		return 0, nil, false
	}
	return hold, s.slots(hold, side), true
}

// ContentSlots is Contents with all of the side's slots, the empty ones
// too.
func (s *Session) ContentSlots(side int, c hexgrid.Coord) (hold int, slots []SlotView, ok bool) {
	if hold, ok = s.holdOf(side, c); !ok {
		return 0, nil, false
	}
	return hold, s.slotViews(hold, side), true
}

// holdOf is the container record of side's own building or own transporter
// on a hex.
func (s *Session) holdOf(side int, c hexgrid.Coord) (int, bool) {
	cell, in := s.Board.Cell(c)
	if !in || side != 0 && side != 1 {
		return 0, false
	}
	hold, ok := s.Game.HoldAt(side, cell)
	if !ok {
		return 0, false
	}
	h, ok := s.Game.Holds[hold]
	if !ok {
		return 0, false
	}
	if hold < dosrules.HoldFlag40 {
		// HoldAt found the transporter as the cell's occupant.
		u := s.Game.Units[s.Game.Maps[side][2*cell+1]]
		if sideOf(u.Flags4) != side {
			return 0, false
		}
	} else if sideOf(h.Flags) != side {
		return 0, false
	}
	return hold, true
}

// linkShips sets Rear and Head on the two records of each two-cell ship in
// units (class bit 0x40; the head, bit 0x80, is the record before its rear).
func (s *Session) linkShips(units map[hexgrid.Coord]UnitView) {
	at := make(map[int]hexgrid.Coord, len(units))
	for c, u := range units {
		at[u.Index] = c
	}
	for c, u := range units {
		f := s.Game.Units[u.Index].Flags4
		if f&0x40 == 0 || f&0x80 != 0 {
			continue
		}
		h, ok := at[u.Index-1]
		if !ok {
			continue
		}
		head := units[h]
		head.Rear, u.Head = &c, &h
		units[h], units[c] = head, u
	}
}
