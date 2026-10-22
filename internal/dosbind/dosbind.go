// Package dosbind binds a level of levels/*.json to the records of the
// original rules engine (internal/dosrules): a terrain record per original
// cell, the units as placements, the container records of the buildings and
// transporters, and the buildings' energy (dosrules.Building).
//
// The original map is Width × Height cells, cell = row*Width + col, odd
// columns half a cell lower. Level line 2*row holds the even columns of a
// row and line 2*row+1 the odd ones, so hexgrid.Coord{Line, Col} is cell
// (Line/2)*Width + 2*Col + Line%2. The last column and the last row are the
// map edge, the record Edge, whatever the level shows there; the levels leave
// out the last column and the odd cells of the last row.
//
// Sides: side 0 is the Central Powers, side 1 the Entente, 2 neutral
// (Factions). The level's owner field (game.Player: 0 none, 1 Central Powers,
// 2 Entente) maps onto them.
package dosbind

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Factions are the remake's players of side 0, side 1 and neutral (2).
var Factions = [3]game.Player{game.CentralPowers, game.Entente, game.None}

// Side is the side of a remake player: 0, 1, or 2 for none.
func Side(p game.Player) int {
	switch p {
	case game.CentralPowers:
		return 0
	case game.Entente:
		return 1
	}
	return 2
}

//go:embed records.json
var recordsJSON []byte

// Records is the binding table (records.json).
type Records struct {
	// Edge is the terrain record of the map edge.
	Edge int `json:"edge"`
	// Terrain maps a hex key (Key) to its terrain record.
	Terrain map[string]int `json:"terrain"`
	// Carries are the classes (container flags 0xfc0) a transporter type
	// takes, by UNIT.DAT type.
	Carries map[string]int `json:"carries"`
	// Maps are the scenario numbers of a level, by its mapNo.
	Maps map[string]MapRecords `json:"maps"`
}

// MapRecords are what a scenario adds to the table: terrain keys whose
// record differs on this map, each building record's energy credits per
// cycle, and the types its factories build.
type MapRecords struct {
	Terrain map[string]int `json:"terrain,omitempty"`
	Credits map[string]int `json:"credits"`
	Builds  []int          `json:"builds"`
}

// Table is the embedded binding table.
var Table = mustRecords(recordsJSON)

func mustRecords(b []byte) *Records {
	var r Records
	if err := json.Unmarshal(b, &r); err != nil {
		panic("dosbind: " + err.Error())
	}
	return &r
}

// Board is a level bound to the original records.
type Board struct {
	Width, Height int
	Winter        bool                       // the winter season: winter terrain (names ending in _w), an original map's mission
	Terrain       []int                      // the terrain record of every cell
	Units         map[int]dosrules.Placement // every unit by live record; a held unit's Cell is its container's
	IDs           map[int]string             // the remake unit id of every unit (FromOriginal: where the remake has one)
	Holds         map[int]dosrules.Hold      // the container records of the buildings and transporters
	Buildings     map[int]dosrules.Building  // the energy of every building record
	// Dropped is the level content the original has no container for (a
	// building cell other than the entrance), by hex; it is not placed.
	Dropped map[hexgrid.Coord][]string

	grid hexgrid.Grid
}

// Cell is the original cell of a level hex; false for a hex off the level.
func (b *Board) Cell(c hexgrid.Coord) (int, bool) {
	if !b.grid.Contains(c) {
		return 0, false
	}
	return c.Line/2*b.Width + 2*c.Col + c.Line%2, true
}

// Coord is the level hex of an original cell; false for a cell the level
// leaves out.
func (b *Board) Coord(cell int) (hexgrid.Coord, bool) {
	if cell < 0 || cell >= b.Width*b.Height {
		return hexgrid.Coord{}, false
	}
	row, col := cell/b.Width, cell%b.Width
	c := hexgrid.Coord{Line: 2*row + col%2, Col: col / 2}
	return c, b.grid.Contains(c)
}

// Edge reports whether a cell is on the map edge.
func (b *Board) Edge(cell int) bool {
	return cell/b.Width == b.Height-1 || cell%b.Width == b.Width-1
}

// Game is a dosrules.Game on the board with the profile's season tables and
// the board's buildings.
func (b *Board) Game(p dosrules.Profile) (*dosrules.Game, error) {
	g, err := dosrules.NewGame(p.Season(b.Winter), b.Width, b.Height, b.Terrain, b.Units, b.Holds)
	if err != nil {
		return nil, err
	}
	g.Buildings = make(map[int]dosrules.Building, len(b.Buildings))
	for i, bd := range b.Buildings {
		bd.Builds = slices.Clone(bd.Builds)
		g.Buildings[i] = bd
	}
	return g, nil
}

// Key is the table key of a hex: its terrain without the winter suffix and
// its infrastructure ids, each once, joined by "+" in the level's order.
func Key(h level.Hex) string {
	k := strings.TrimSuffix(h.Terrain, "_w")
	var seen []string
	for _, id := range h.Infra.IDs {
		if !slices.Contains(seen, id) {
			seen = append(seen, id)
			k += "+" + id
		}
	}
	return k
}

// The building families: their cells carry the name prefix, the entrance
// holds the level's content.
var families = []struct{ prefix, entrance string }{{"hq", "hqN"}, {"cmp", "cmpN"}, {"factory", "factoryE"}}

func family(terrain string) int {
	t := strings.TrimSuffix(terrain, "_w")
	for i, f := range families {
		if strings.HasPrefix(t, f.prefix) {
			return i
		}
	}
	return -1
}

// building is the cells of one building on the level.
type building struct {
	cells    []hexgrid.Coord
	entrance hexgrid.Coord
	content  *level.Content
}

// state is the suffix a building cell's key takes: the owning side, or "-"
// when the level gives the building no content.
func (bd *building) state() string {
	if bd.content == nil {
		return "-"
	}
	return strconv.Itoa(Side(game.Player(bd.content.Owner)))
}

// buildings groups the building cells of a level: adjacent cells of one
// family form a building.
func buildings(lv *level.Level) (map[hexgrid.Coord]*building, error) {
	g := lv.Grid()
	at := map[hexgrid.Coord]*building{}
	for l, line := range lv.Lines {
		for c, h := range line {
			start := hexgrid.Coord{Line: l, Col: c}
			fam := family(h.Terrain)
			if fam < 0 || at[start] != nil {
				continue
			}
			bd := &building{entrance: hexgrid.Coord{Line: -1}}
			stack := []hexgrid.Coord{start}
			at[start] = bd
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				bd.cells = append(bd.cells, p)
				ph := lv.Lines[p.Line][p.Col]
				if strings.TrimSuffix(ph.Terrain, "_w") == families[fam].entrance {
					if bd.entrance.Line >= 0 {
						return nil, fmt.Errorf("line%d[%d]: a building with two entrances", p.Line+1, p.Col)
					}
					bd.entrance, bd.content = p, ph.Content
				}
				for _, n := range g.Neighbors(p) {
					if at[n] == nil && family(lv.Lines[n.Line][n.Col].Terrain) == fam {
						at[n] = bd
						stack = append(stack, n)
					}
				}
			}
			if bd.entrance.Line < 0 {
				return nil, fmt.Errorf("line%d[%d]: a building without its entrance", l+1, c)
			}
		}
	}
	for l, line := range lv.Lines {
		for c, h := range line {
			if h.Content != nil && at[hexgrid.Coord{Line: l, Col: c}] == nil {
				return nil, fmt.Errorf("line%d[%d]: content outside a building", l+1, c)
			}
		}
	}
	return at, nil
}

// Terrain flags of the building records and the container records they use.
const (
	flagHQ      = 0x40
	flagDepot   = 0x100
	flagFactory = 0x400
	tableLen    = 10   // depot and factory records
	carrier     = 0x20 // a transporter's container record
	transporter = 0x1000
	twoCell     = 0x40
)

// owner is the side a building record's flags give (dosrules.Owns): 1 for
// bit 1, 2 (neutral) for bit 2, else 0.
func owner(flags int) int {
	switch {
	case flags&1 != 0:
		return 1
	case flags&2 != 0:
		return 2
	}
	return 0
}

// Bind binds a level. Every hex must have a terrain key in the table and
// every unit a type; an error names the hex.
//
// Assumed, not verified:
//   - Units are numbered in the order of their cells, then the units in
//     buildings in the order of the building records and their content.
//   - Both sides' slots of a building record list its units, as both maps
//     show every unit at the start; a unit in a neutral building is neutral.
//   - A building record covers the building's entrance first, then its
//     other cells in cell order, at most four; its occupant byte is
//     dosrules.InBuilding. Depot and factory records are numbered in the
//     order of each building's first cell.
//   - A transporter's record is the next free one from 0 in unit order; it
//     takes the classes of Carries, its occupant byte is the transporter.
//   - A building record's credits are the scenario's (Maps.Credits); only
//     factories build, the scenario's types (Maps.Builds).
func Bind(lv *level.Level, set *data.Set) (*Board, error) {
	g := lv.Grid()
	if g.Width < 2 || g.Lines < 2 {
		return nil, fmt.Errorf("dosbind: a level of %d lines of %d hexes", g.Lines, g.Width)
	}
	b := &Board{Width: 2 * g.Width, Height: (g.Lines + 1) / 2, grid: g,
		Units: map[int]dosrules.Placement{}, IDs: map[int]string{}, Holds: map[int]dosrules.Hold{},
		Buildings: map[int]dosrules.Building{}}
	b.Terrain = make([]int, b.Width*b.Height)
	for i := range b.Terrain {
		b.Terrain[i] = Table.Edge
	}
	mr := Table.Maps[strconv.Itoa(lv.MapNo)]
	winter := 0
	at, err := buildings(lv)
	if err != nil {
		return nil, fmt.Errorf("dosbind: %w", err)
	}
	for l, line := range lv.Lines {
		for c, h := range line {
			p := hexgrid.Coord{Line: l, Col: c}
			if strings.HasSuffix(h.Terrain, "_w") {
				winter++
			}
			key := Key(h)
			if bd := at[p]; bd != nil {
				if _, ok := lookup(mr, key+"@"+bd.state()); ok {
					key += "@" + bd.state()
				}
			}
			rec, ok := lookup(mr, key)
			if !ok {
				return nil, fmt.Errorf("dosbind: line%d[%d]: no terrain record for %q", l+1, c, key)
			}
			cell, _ := b.Cell(p)
			if !b.Edge(cell) {
				b.Terrain[cell] = rec
			}
		}
	}
	switch winter {
	case 0:
	case g.Size():
		b.Winter = true
	default:
		return nil, fmt.Errorf("dosbind: %d of %d hexes are winter terrain", winter, g.Size())
	}
	if err := b.placeUnits(lv, set); err != nil {
		return nil, err
	}
	if err := b.bindBuildings(set, at, mr); err != nil {
		return nil, err
	}
	for p, bd := range at {
		if h := lv.Lines[p.Line][p.Col]; h.Content != nil && p != bd.entrance && len(h.Content.Units) > 0 {
			if b.Dropped == nil {
				b.Dropped = map[hexgrid.Coord][]string{}
			}
			b.Dropped[p] = slices.Clone(h.Content.Units)
		}
	}
	if err := b.bindTransporters(); err != nil {
		return nil, err
	}
	return b, nil
}

func lookup(mr MapRecords, key string) (int, bool) {
	if r, ok := mr.Terrain[key]; ok {
		return r, true
	}
	r, ok := Table.Terrain[key]
	return r, ok
}

// unitType is the UNIT.DAT type and faction side of a level unit id.
func unitType(set *data.Set, id string) (int, int, error) {
	u, ok := set.Units[id]
	if !ok {
		return 0, 0, fmt.Errorf("unknown unit %q", id)
	}
	types := dosrules.Full.Tables.Move.Types
	if u.Dos < 0 || u.Dos >= len(types) {
		return 0, 0, fmt.Errorf("unit %q: no type %d", id, u.Dos)
	}
	if types[u.Dos].Class&twoCell != 0 {
		return 0, 0, fmt.Errorf("unit %q: two-cell ships are not bound", id)
	}
	return u.Dos, Side(game.FactionOf(u.Country)), nil
}

// add places the next unit; an empty id names none.
func (b *Board) add(id string, p dosrules.Placement) (int, error) {
	i := len(b.Units)
	if i >= dosrules.SlotFree {
		return 0, fmt.Errorf("more than %d units", dosrules.SlotFree)
	}
	b.Units[i] = p
	if id != "" {
		b.IDs[i] = id
	}
	return i, nil
}

// placeUnits places the units on the map in cell order.
func (b *Board) placeUnits(lv *level.Level, set *data.Set) error {
	for cell := range b.Width * b.Height {
		p, ok := b.Coord(cell)
		if !ok {
			continue
		}
		h := lv.Lines[p.Line][p.Col]
		if h.Unit == "" {
			continue
		}
		typ, side, err := unitType(set, h.Unit)
		if err == nil && b.Edge(cell) {
			err = fmt.Errorf("unit %q on the map edge", h.Unit)
		}
		if f := dosrules.Full.Tables.Move.Terrains[b.Terrain[cell]].Flags; err == nil && f&(flagHQ|flagDepot|flagFactory) != 0 {
			err = fmt.Errorf("unit %q on a building entrance", h.Unit)
		}
		if err == nil {
			_, err = b.add(h.Unit, dosrules.Placement{Type: typ, Owner: side, Cell: cell})
		}
		if err != nil {
			return fmt.Errorf("dosbind: line%d[%d]: %w", p.Line+1, p.Col, err)
		}
	}
	return nil
}

// bindBuildings gives every building entrance its container record, in
// the order of the buildings' first cells, with the units of its content.
func (b *Board) bindBuildings(set *data.Set, at map[hexgrid.Coord]*building, mr MapRecords) error {
	var order []*building
	first := map[*building]int{}
	for p, bd := range at {
		cell, _ := b.Cell(p)
		if f, ok := first[bd]; !ok || cell < f {
			first[bd] = cell
		}
		if !slices.Contains(order, bd) {
			order = append(order, bd)
		}
	}
	slices.SortFunc(order, func(x, y *building) int { return first[x] - first[y] })
	base := map[int]int{flagDepot: dosrules.HoldFlag100, flagFactory: dosrules.HoldFlag400}
	next := map[int]int{flagDepot: base[flagDepot], flagFactory: base[flagFactory]}
	terrains := dosrules.Full.Tables.Move.Terrains
	for _, bd := range order {
		entrance, _ := b.Cell(bd.entrance)
		f := terrains[b.Terrain[entrance]].Flags
		var idx int
		switch {
		case f&flagHQ != 0:
			idx = dosrules.HoldFlag40 + f&1
		case f&flagDepot != 0, f&flagFactory != 0:
			kind := f & (flagDepot | flagFactory)
			idx = next[kind]
			if idx == base[kind]+tableLen {
				return fmt.Errorf("dosbind: line%d[%d]: more than %d buildings of the kind", bd.entrance.Line+1, bd.entrance.Col, tableLen)
			}
			next[kind]++
		default:
			if bd.content != nil {
				return fmt.Errorf("dosbind: line%d[%d]: content in terrain record %d, which is no container", bd.entrance.Line+1, bd.entrance.Col, b.Terrain[entrance])
			}
			continue
		}
		if _, dup := b.Holds[idx]; dup {
			return fmt.Errorf("dosbind: line%d[%d]: a second building for record %d", bd.entrance.Line+1, bd.entrance.Col, idx)
		}
		side := owner(f)
		var rest []int
		for _, p := range bd.cells {
			if c, _ := b.Cell(p); c != entrance {
				rest = append(rest, c)
			}
		}
		h := buildingHold(side, entrance, rest)
		if bd.content != nil {
			if len(bd.content.Units) > 7 {
				return fmt.Errorf("dosbind: line%d[%d]: %d units in a building of 7 slots", bd.entrance.Line+1, bd.entrance.Col, len(bd.content.Units))
			}
			for k, id := range bd.content.Units {
				typ, faction, err := unitType(set, id)
				if err == nil && side != 2 && faction != side {
					err = fmt.Errorf("unit %q in a building of side %d", id, side)
				}
				var u int
				if err == nil {
					u, err = b.add(id, dosrules.Placement{Type: typ, Owner: side, Cell: entrance})
				}
				if err != nil {
					return fmt.Errorf("dosbind: line%d[%d]: %w", bd.entrance.Line+1, bd.entrance.Col, err)
				}
				hold(&h, k, u)
			}
		}
		b.Holds[idx] = h
		bd2 := dosrules.Building{Credits: mr.Credits[strconv.Itoa(idx)]}
		if f&flagFactory != 0 {
			bd2.Builds = slices.Clone(mr.Builds)
		}
		b.Buildings[idx] = bd2
	}
	if mr.Credits != nil {
		for k := range mr.Credits {
			i, err := strconv.Atoi(k)
			if _, ok := b.Holds[i]; err != nil || !ok {
				return fmt.Errorf("dosbind: credits for record %s, which the level has no building for", k)
			}
		}
	}
	return nil
}

// bindTransporters gives every transporter on the map a container record.
func (b *Board) bindTransporters() error {
	types := dosrules.Full.Tables.Move.Types
	next := 0
	for i := range len(b.Units) {
		u := b.Units[i]
		if types[u.Type].Class&transporter == 0 {
			continue
		}
		h, ok := carrierHold(u, i)
		if !ok {
			return fmt.Errorf("dosbind: unit %d (%s): no classes for transporter type %d", i, b.IDs[i], u.Type)
		}
		if next == dosrules.HoldFlag40 {
			return fmt.Errorf("dosbind: more than %d transporters", dosrules.HoldFlag40)
		}
		b.Holds[next] = h
		next++
	}
	return nil
}

// buildingHold is the container record of a building of side: its entrance
// first, then its other cells in cell order, at most four; its occupant
// byte is dosrules.InBuilding.
func buildingHold(side, entrance int, rest []int) dosrules.Hold {
	h := dosrules.EmptyHold()
	h.Flags, h.Occupant = side, dosrules.InBuilding
	h.Cells[0] = 2 * entrance
	rest = slices.Sorted(slices.Values(rest))
	for k := 1; k < len(h.Cells) && k-1 < len(rest); k++ {
		h.Cells[k] = 2 * rest[k-1]
	}
	return h
}

// carrierHold is the container record of transporter unit i: its position
// first, the classes of Carries and its side; false for a type without
// classes.
func carrierHold(u dosrules.Placement, i int) (dosrules.Hold, bool) {
	classes, ok := Table.Carries[strconv.Itoa(u.Type)]
	if !ok {
		return dosrules.Hold{}, false
	}
	h := dosrules.EmptyHold()
	h.Cells[0] = 2*u.Cell + 1
	h.Flags = carrier | classes | u.Owner&3
	h.Occupant = i
	return h, true
}

// hold lists unit u in slot k of both sides of h.
func hold(h *dosrules.Hold, k, u int) { h.Slots[k], h.Slots[7+k] = u, u }
