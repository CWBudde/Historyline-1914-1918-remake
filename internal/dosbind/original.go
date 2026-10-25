package dosbind

import (
	"fmt"
	"maps"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Terrain flags of a building's other cells, by the flag of its entrance.
var partFlag = map[int]int{flagHQ: 0x20, flagDepot: 0x80, flagFactory: 0x200}

// hqCredits are the credits of an HQ without a record.
const hqCredits = 50

// FromOriginal binds an imported original map (internal/finmap) as the
// map's files give it: the terrain record of every cell, the units in the
// order the game creates them, the container records of the buildings and
// transporters and the credits of the scenario's records. The season is the
// mission's (summer without one). IDs name the remake unit of a unit's type
// and side where the remake has one.
//
// Assumed, not verified:
//   - The map's units come first, in the map's order, a two-cell ship's
//     rear (type Type+1, on its own cell) as the record after its head, as
//     the move reads it; then the units of the records in file order, an
//     overridden record's too.
//   - A record's units take its side and both sides' slots of their
//     container in order, as Bind lists a building's units; a unit in a
//     container stands at its first cell, a transporter created there gets
//     its own record at that cell.
//   - A building record covers the entrance first, then the cells around it
//     whose terrain carries its kind's flag (0x20 HQ, 0x80 depot, 0x200
//     factory) in cell order, at most four; its occupant byte is
//     dosrules.InBuilding and its side the entrance's flags (owner).
//   - A building's credits are the energy of its record that is not
//     overridden, an HQ's 50 without one; factories build the map's
//     producible types.
//   - A record for a transporter the map lacks creates nothing.
func FromOriginal(l *finmap.Level, set *data.Set) (*Board, error) {
	if l.Width < 2 || l.Width%2 != 0 || l.Height < 2 || len(l.Terrain) != l.Height {
		return nil, fmt.Errorf("dosbind: an original map of %d×%d cells", l.Width, l.Height)
	}
	b := &Board{
		Width: l.Width, Height: l.Height, Winter: l.Mission != nil && l.Mission.Winter,
		Terrain: make([]int, 0, l.Width*l.Height), grid: hexgrid.Grid{Width: l.Width / 2, Lines: 2*l.Height - 1},
		Units: map[int]dosrules.Placement{}, IDs: map[int]string{}, Holds: map[int]dosrules.Hold{}, Parts: map[int]int{},
		Buildings: map[int]dosrules.Building{},
	}
	terrains := dosrules.Full.Tables.Move.Terrains
	for row, line := range l.Terrain {
		if len(line) != l.Width {
			return nil, fmt.Errorf("dosbind: terrain row %d of %d cells", row, len(line))
		}
		for col, t := range line {
			if t < 0 || t >= len(terrains) {
				return nil, fmt.Errorf("dosbind: cell %d,%d: terrain record %d", row, col, t)
			}
		}
		b.Terrain = append(b.Terrain, line...)
	}
	ids := RemakeIDs(set)
	place := func(typ, side, cell int) (int, error) {
		if typ < 0 || typ >= len(dosrules.Full.Tables.Move.Types) || side < 0 || side > 2 || cell < 0 || cell >= len(b.Terrain) {
			return 0, fmt.Errorf("type %d of side %d at cell %d", typ, side, cell)
		}
		return b.add(ids[[2]int{typ, side}], dosrules.Placement{Type: typ, Owner: side, Cell: cell})
	}
	transporter := func(i, idx int) error {
		h, ok := carrierHold(b.Units[i], i)
		switch _, dup := b.Holds[idx]; {
		case !ok:
			return fmt.Errorf("no classes for transporter type %d", b.Units[i].Type)
		case dup || idx < 0 || idx >= dosrules.HoldFlag40:
			return fmt.Errorf("container record %d", idx)
		}
		b.Holds[idx] = h
		return nil
	}

	types := dosrules.Full.Tables.Move.Types
	for _, u := range l.Units {
		i, err := place(u.Type, u.Side, b.cellAt(u.Cell))
		if err == nil {
			err = shape(types[u.Type].Class, u.Rear != nil, u.Container != nil)
		}
		if err == nil && u.Rear != nil {
			_, err = place(u.Type+1, u.Side, b.cellAt(*u.Rear))
		}
		if err == nil && u.Container != nil {
			err = transporter(i, *u.Container)
		}
		if err != nil {
			return nil, fmt.Errorf("dosbind: unit at %d,%d: %w", u.Cell[0], u.Cell[1], err)
		}
	}
	if err := b.originalBuildings(l); err != nil {
		return nil, err
	}

	slots := map[int]int{} // container record -> slots taken
	for _, c := range l.Containers {
		if c.Holder == "none" {
			if len(c.Units) > 0 {
				return nil, fmt.Errorf("dosbind: record for container %d, which the map lacks, creates units", c.Container)
			}
			continue
		}
		h, ok := b.Holds[c.Container]
		if !ok {
			return nil, fmt.Errorf("dosbind: record for container %d, which the map lacks", c.Container)
		}
		if bd, ok := b.Buildings[c.Container]; ok && !c.Overridden {
			bd.Credits = c.Energy
			b.Buildings[c.Container] = bd
		}
		for _, u := range c.Units {
			k := slots[c.Container]
			if k == 7 {
				return nil, fmt.Errorf("dosbind: more than 7 units in container %d", c.Container)
			}
			i, err := place(u.Type, c.Side, h.Cells[0]/2)
			if err == nil {
				err = shape(types[u.Type].Class, false, u.Container != nil)
			}
			if err == nil && u.Container != nil {
				err = transporter(i, *u.Container)
			}
			if err != nil {
				return nil, fmt.Errorf("dosbind: unit in container %d: %w", c.Container, err)
			}
			hold(&h, k, i)
			slots[c.Container] = k + 1
		}
		b.Holds[c.Container] = h
	}
	return b, nil
}

// shape checks a unit of type class against the map: a two-cell ship's head
// (class 0xc0) has its rear and nothing else does, a transporter has its
// container record and nothing else does.
func shape(class int, rear, container bool) error {
	switch {
	case (class&0xC0 == 0xC0) != rear:
		return fmt.Errorf("class %#x with rear %v", class, rear)
	case (class&transporter != 0) != container:
		return fmt.Errorf("class %#x with container record %v", class, container)
	}
	return nil
}

// cellAt is the cell of [row, col]; -1 off the map.
func (b *Board) cellAt(c [2]int) int {
	if c[0] < 0 || c[0] >= b.Height || c[1] < 0 || c[1] >= b.Width {
		return -1
	}
	return c[0]*b.Width + c[1]
}

// originalBuildings gives every building its container record and credits.
func (b *Board) originalBuildings(l *finmap.Level) error {
	terrains := dosrules.Full.Tables.Move.Terrains
	flags := func(cell int) int { return terrains[b.Terrain[cell]].Flags }
	claimed := map[int]int{} // building cell -> entrance
	for _, bd := range l.Buildings {
		entrance := b.cellAt(bd.Cell)
		if entrance < 0 {
			return fmt.Errorf("dosbind: building at %d,%d off the map", bd.Cell[0], bd.Cell[1])
		}
		f := flags(entrance)
		kind := f & (flagHQ | flagDepot | flagFactory)
		part, ok := partFlag[kind]
		if !ok {
			return fmt.Errorf("dosbind: building at %d,%d on terrain flags %#x", bd.Cell[0], bd.Cell[1], f)
		}
		if _, dup := b.Holds[bd.Container]; dup || bd.Container < dosrules.HoldFlag40 || bd.Container >= dosrules.HoldRecords {
			return fmt.Errorf("dosbind: building at %d,%d: container record %d", bd.Cell[0], bd.Cell[1], bd.Container)
		}
		var rest []int
		stack := []int{entrance}
		for len(stack) > 0 {
			c := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			p, _ := b.Coord(c)
			for _, n := range b.grid.Neighbors(p) {
				nc, _ := b.Cell(n)
				if nf := flags(nc); nf&part == 0 || nf&(flagHQ|flagDepot|flagFactory) != 0 || nc == entrance {
					continue
				}
				if e, ok := claimed[nc]; ok {
					if e != entrance {
						return fmt.Errorf("dosbind: cell %d belongs to the buildings at cells %d and %d", nc, e, entrance)
					}
					continue
				}
				claimed[nc] = entrance
				rest = append(rest, nc)
				b.Parts[nc] = entrance
				stack = append(stack, nc)
			}
		}
		b.Holds[bd.Container] = buildingHold(owner(f), entrance, rest)
		var credit dosrules.Building
		switch kind {
		case flagHQ:
			credit.Credits = hqCredits
		case flagFactory:
			credit.Builds = slices.Clone(l.Producible)
		}
		b.Buildings[bd.Container] = credit
	}
	return nil
}

// RemakeIDs are the remake's unit ids by UNIT.DAT type and side, the first
// in id order where several share both.
func RemakeIDs(set *data.Set) map[[2]int]string {
	out := map[[2]int]string{}
	for _, id := range slices.Sorted(maps.Keys(set.Units)) {
		u := set.Units[id]
		k := [2]int{u.Dos, Side(game.FactionOf(u.Country))}
		if _, ok := out[k]; !ok {
			out[k] = id
		}
	}
	return out
}
