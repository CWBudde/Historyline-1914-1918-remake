package dosrules

import (
	"errors"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestMovementGate is the R2 gate on each recorded build. Like the R1 gate
// (TestGate) every clause reads the recorded expectations,
// never the resolver's own results, and checks them against the rule the
// spec states; the replays show that the resolver reproduces them. The
// inputs are the fixtures' own set-up and the profile's tables.
func TestMovementGate(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			ms := load[MovementFixture](t, build, "movements.json")
			ss := load[StepFixture](t, build, "steps.json")
			cs := load[ContainerFixture](t, build, "containers.json")
			t.Run("costs", func(t *testing.T) { gateCosts(t, p, ms) })
			t.Run("endpoints", func(t *testing.T) { gateEndpoints(t, p, ms) })
			t.Run("domain", func(t *testing.T) { gateDomain(t, p, ms) })
			t.Run("containers", func(t *testing.T) { gateContainers(t, p, cs) })
			t.Run("exits", func(t *testing.T) { gateExits(t, cs) })
			t.Run("no consumables", func(t *testing.T) { gateConsumables(t, ss, cs) })
		})
	}
	t.Run("no foreign rules", gateForeign)
	t.Run("gated off", gateUndefined)
}

// cellsOf are the cells of a map's neighbours64 list; on a map one cell wide
// or high the edge clips (else-ifs) leak past the map, and those are left out.
func cellsOf(g *Movement, cell int) []int {
	var out []int
	for _, n := range neighbours64(cell%g.Width, cell/g.Width, g.Width, g.Height) {
		if n >= 0 && n%64 < g.Width && n/64 < g.Height {
			out = append(out, n/64*g.Width+n%64)
		}
	}
	return out
}

// closed reports whether the cell's terrain is closed to the moving unit:
// the masks share no bit, or the flags hit the caller's block mask.
func closed(tab *MoveTables, g *Movement, cell int) bool {
	tr, typ := tab.Terrains[g.Terrain[cell]], tab.Types[g.PieceAt(g.Unit).Type]
	return tr.Mask&typ.Mask == 0 || g.Block != 0xFFFF && tr.Flags&g.Block != 0
}

// Costs (*Raw budget vs 2 × speed*): on a map of one terrain with only the
// mover, each neighbour of the start keeps the budget minus the entry cost:
// terrain byte 3 (of the season's table), byte 4 for an air unit, 2 for live
// +6 bit 0x10, and 80 for closed terrain. The budget of the terrain cases is
// type byte 0, which is not 2 × speed for every type.
func gateCosts(t *testing.T, p Profile, ms []MovementFixture) {
	seen := map[string]int{}
	for _, f := range ms {
		if f.Budget >= 128 {
			// The budget is a signed byte in the cost buffer.
			for c, v := range f.Expected.Remaining {
				if v >= 0 {
					t.Errorf("%s: budget %d reaches cell %d", f.ID, f.Budget, c)
				}
			}
			seen["budget >= 128 reaches nothing"]++
			continue
		}
		if len(f.Cells) != 0 || len(f.Units) != 1 || len(f.Unseen) != 0 {
			continue
		}
		tab, g := p.Tables.Move.Season(f.Winter), f.Query()
		u := g.PieceAt(g.Unit)
		typ, tr := tab.Types[u.Type], tab.Terrains[f.Fill]
		class, _, _ := strings.Cut(f.ID, "/")
		if class == "terrain" || class == "winter" {
			if f.Budget != typ.Budget {
				t.Errorf("%s: budget %d, type byte 0 is %d", f.ID, f.Budget, typ.Budget)
			}
			if typ.Budget == 2*typ.Speed {
				seen["byte 0 is 2 × speed"]++
			} else {
				seen["byte 0 is not 2 × speed"]++
			}
		}
		cost, kind := tr.Cost, "ground cost (byte 3)"
		switch {
		case closed(tab, g, u.Cell):
			cost, kind = 80, "closed terrain (80)"
		case u.Flags6&0x10 != 0:
			cost, kind = 2, "flat cost 2"
		case u.Flags4&0x10 != 0:
			cost, kind = tr.AirCost, "air cost (byte 4)"
		}
		want := f.Budget - cost
		if f.Budget == 0 || want < 0 {
			want = -1
		}
		for _, n := range cellsOf(g, u.Cell) {
			got := f.Expected.Remaining[n]
			if want >= 0 && got != want || want < 0 && got >= 0 {
				t.Errorf("%s: neighbour %d keeps %d, budget %d cost %d", f.ID, n, got, f.Budget, cost)
			}
		}
		seen[kind]++
		if f.Winter && tr.Cost != p.Tables.Move.Terrains[f.Fill].Cost {
			seen["winter cost differs from summer"]++
		}
	}
	keys := []string{"byte 0 is 2 × speed", "byte 0 is not 2 × speed", "budget >= 128 reaches nothing",
		"ground cost (byte 3)", "closed terrain (80)", "flat cost 2", "air cost (byte 4)"}
	if p.Tables.Move.Winter != nil {
		keys = append(keys, "winter cost differs from summer")
	}
	need(t, seen, keys...)
}

// enemy reports whether live flags +4 are another player's unit (a neutral
// one, bit 2, blocks its cell but casts no zone of control).
func enemy(flags, side int) bool { return flags&2 == 0 && !Owns(flags, side) }

// Endpoints (*Zone of control*): a destination is reached, is not the
// unit's own cell and holds no unit of the side's map; a cell next to an
// enemy is entered with 0 left, whatever it costs; own units are passed; a
// unit only on the other side's map is a free cell.
func gateEndpoints(t *testing.T, p Profile, ms []MovementFixture) {
	seen := map[string]int{}
	for _, f := range ms {
		tab, g := p.Tables.Move.Season(f.Winter), f.Query()
		x, start := f.Expected, g.PieceAt(g.Unit).Cell
		occ := g.Occupants()
		for _, d := range x.Destinations {
			_, held := occ[d]
			switch {
			case d == start:
				t.Errorf("%s: the unit's own cell is a destination", f.ID)
			case x.Remaining[d] < 0:
				t.Errorf("%s: destination %d was not reached", f.ID, d)
			case held:
				t.Errorf("%s: destination %d holds unit %d", f.ID, d, occ[d])
			}
		}
		for _, u := range g.Unseen {
			for _, d := range x.Destinations {
				if d == u.Cell {
					seen["unseen unit's cell is a destination"]++
				}
			}
		}
		zoc := map[int]bool{}
		for idx, u := range g.Units {
			switch {
			case idx == g.Unit:
			case Owns(u.Flags4, g.Side):
				if x.Remaining[u.Cell] >= 0 {
					seen["own unit passed"]++
				}
			case enemy(u.Flags4, g.Side):
				for _, n := range cellsOf(g, u.Cell) {
					zoc[n] = true
				}
			}
		}
		for n := range zoc {
			if _, held := occ[n]; held || closed(tab, g, n) {
				continue
			}
			switch v := x.Remaining[n]; {
			case v > 0:
				t.Errorf("%s: cell %d next to an enemy keeps %d", f.ID, n, v)
			case v == 0:
				seen["zone of control entered with 0 left"]++
			}
		}
	}
	need(t, seen, "own unit passed", "zone of control entered with 0 left", "unseen unit's cell is a destination")
}

// Domain: terrain the unit's mask does not share a bit with, or whose flags
// hit the block mask, is entered only from 80 budget up, and with the mask
// closed it is never a destination; an air unit never ends on terrain flag
// 0x4000.
func gateDomain(t *testing.T, p Profile, ms []MovementFixture) {
	seen := map[string]int{}
	for _, f := range ms {
		tab, g := p.Tables.Move.Season(f.Winter), f.Query()
		x, u := f.Expected, g.PieceAt(g.Unit)
		occ := g.Occupants()
		dest := map[int]bool{}
		for _, d := range x.Destinations {
			dest[d] = true
		}
		for c := range g.Width * g.Height {
			if _, held := occ[c]; held || !closed(tab, g, c) {
				continue
			}
			reached := x.Remaining[c] >= 0
			switch {
			case reached && f.Budget < 80:
				t.Errorf("%s: closed cell %d reached on budget %d", f.ID, c, f.Budget)
			case reached:
				seen["closed cell entered from budget 80"]++
			default:
				seen["closed cell not entered"]++
			}
			if tab.Terrains[g.Terrain[c]].Mask&tab.Types[u.Type].Mask == 0 {
				if dest[c] {
					t.Errorf("%s: cell %d outside the unit's mask is a destination", f.ID, c)
				} else if reached {
					seen["reached outside the mask, no destination"]++
				}
			}
		}
		if u.Flags4&0x10 != 0 {
			for c := range g.Width * g.Height {
				if tab.Terrains[g.Terrain[c]].Flags&0x4000 == 0 {
					continue
				}
				if dest[c] {
					t.Errorf("%s: an air unit may end on flag 0x4000 at %d", f.ID, c)
				} else if x.Remaining[c] >= 0 {
					seen["air unit reaches flag 0x4000, no destination"]++
				}
			}
		}
	}
	need(t, seen, "closed cell entered from budget 80", "closed cell not entered",
		"reached outside the mask, no destination", "air unit reaches flag 0x4000, no destination")
}

// codes are the message codes of a recorded outcome.
func codes(x StepExpected) map[int]bool {
	out := map[int]bool{}
	for _, m := range x.Messages {
		out[m[0]] = true
	}
	return out
}

// taken are a side's taken slots of a container record.
func taken(h Hold, side int) []int {
	var out []int
	for _, v := range h.Slots[7*side : 7*side+7] {
		if v <= SlotFree {
			out = append(out, v)
		}
	}
	return out
}

func holdOf(g *Movement, idx int) Hold {
	if h, ok := g.Holds[idx]; ok {
		return h
	}
	return EmptyHold()
}

// Containers: a load into an own transporter is refused with 0x3a when the
// unit's live +6 shares none of the record's classes (+0x19 & 0xfc0) or the
// transporter has moved, with 0x3b when the side's seven slots are taken, a
// carrying unit's cargo needs more free slots than there are less one, or
// the weights exceed the transporter type's capacity; otherwise the unit
// takes the side's first free slot. A building never refuses with 0x3a or
// for weight.
func gateContainers(t *testing.T, p Profile, cs []ContainerFixture) {
	seen := map[string]int{}
	tab := p.Tables.Move
	for _, f := range cs {
		if f.Order.Hold >= 0 {
			continue
		}
		g, x := f.Movement(), f.Expected
		mover, side, goal := g.PieceAt(g.Unit), g.Side, f.Order.Goal
		got := codes(x.StepExpected)
		if x.Order&1 == 0 && len(x.Messages) == 0 {
			// Refused before the end-cell check: no reach bit on the goal.
			seen["goal not reached"]++
			continue
		}
		idx, onUnit := g.Occupants()[goal]
		if !onUnit {
			if tab.Terrains[g.Terrain[goal]].Flags&(0x400|0x100|0x40) != 0 {
				if got[0x3A] || got[0x40] {
					t.Errorf("%s: a building refused with %v", f.ID, x.Messages)
				}
				switch {
				case got[0x3B]:
					seen["building full"]++
				case x.Order&1 != 0:
					seen["building entered"]++
				}
			}
			continue
		}
		tr := g.PieceAt(idx)
		if tr.Flags4&0x40 != 0 && tr.Flags4&0x80 == 0 {
			idx--
			tr = g.PieceAt(idx)
		}
		if !Owns(tr.Flags4, side) || tr.Flags4&0x1000 == 0 {
			continue
		}
		h := holdOf(g, int(int8(tr.Hold)))
		in := taken(h, side)
		weight := tab.Types[mover.Type].Weight
		for _, u := range in {
			weight += tab.Types[g.PieceAt(u).Type].Weight
		}
		fits := true
		if mover.Flags4&0x1000 != 0 {
			cargo := taken(holdOf(g, int(int8(mover.Hold))), side)
			for _, u := range cargo {
				weight += tab.Types[g.PieceAt(u).Type].Weight
			}
			fits = 7-len(in) >= len(cargo)+1
			seen["carrying transporter"]++
		}
		capacity := tab.Types[tr.Type].Capacity
		var want int
		switch {
		case mover.Flags6&h.Flags&0xFC0 == 0:
			want, seen["0x3a: no shared class"] = 0x3A, seen["0x3a: no shared class"]+1
		case tr.Flags4&0x200 != 0:
			want, seen["0x3a: transporter moved"] = 0x3A, seen["0x3a: transporter moved"]+1
		case len(in) == 7:
			want, seen["0x3b: seven slots taken"] = 0x3B, seen["0x3b: seven slots taken"]+1
		case weight > capacity || !fits:
			want, seen["0x3b: over capacity"] = 0x3B, seen["0x3b: over capacity"]+1
		case weight == capacity:
			seen["loaded at capacity"]++
		default:
			seen["loaded"]++
		}
		if want != 0 {
			if !got[want] || x.Order&1 != 0 {
				t.Errorf("%s: recorded order %#x messages %v, the rule gives %#x", f.ID, x.Order, x.Messages, want)
			}
			continue
		}
		if got[0x3A] || got[0x3B] || x.Order&1 == 0 {
			t.Errorf("%s: recorded order %#x messages %v, the rule loads", f.ID, x.Order, x.Messages)
			continue
		}
		// The unit is in the side's first free slot of the record.
		slot := len(in)
		for i, v := range h.Slots[7*side : 7*side+7] {
			if v > SlotFree {
				slot = i
				break
			}
		}
		loaded := false
		for _, r := range x.Holds {
			if r.Index == int(int8(tr.Hold)) && r.Bytes[7*side+slot] == g.Unit {
				loaded = true
			}
		}
		if !loaded {
			t.Errorf("%s: unit %d not in slot %d of record %d", f.ID, g.Unit, slot, tr.Hold)
		}
	}
	need(t, seen, "loaded", "loaded at capacity", "carrying transporter", "0x3a: no shared class",
		"0x3a: transporter moved", "0x3b: seven slots taken", "0x3b: over capacity", "building full", "building entered")
}

// Exits: the exit's reach search starts from budget 4 when the unit's live
// +6 has bit 0x10, else 8, whatever live byte 0 holds.
func gateExits(t *testing.T, cs []ContainerFixture) {
	seen := map[string]int{}
	for _, f := range cs {
		if f.Order.Hold < 0 {
			continue
		}
		g := f.Movement()
		u := g.PieceAt(g.Unit)
		want := 8
		if u.Flags6&0x10 != 0 {
			want = 4
		}
		top := -1
		for _, r := range f.Expected.Remaining {
			top = max(top, r[1])
		}
		if top != want {
			t.Errorf("%s: the exit reach starts from %d, the rule gives %d", f.ID, top, want)
		}
		seen["exit budget "+strconv.Itoa(want)]++
		if f.Budget != want {
			seen["live byte 0 not the exit budget"]++
		}
	}
	need(t, seen, "exit budget 4", "exit budget 8", "live byte 0 not the exit budget")
}

// No consumables: no move, load or exit changes a unit's live byte 0, its
// experience, group size or alternate strength, so there is no fuel, ammo or
// resupply; a unit leaves a container slot only by the exit or the hand-over
// of a carrying transporter moving into a building.
func gateConsumables(t *testing.T, ss []StepFixture, cs []ContainerFixture) {
	seen := map[string]int{}
	check := func(id string, g *Movement, live []FixtureRecord) {
		before := LiveRecords(g)
		for _, r := range live {
			b := before[r.Index]
			if b == nil {
				b = make([]byte, LiveLen)
			}
			for _, k := range []int{0, 1, 2, 3} {
				if r.Bytes[k] != int(b[k]) {
					t.Errorf("%s: unit %d live byte %d changed from %d to %d", id, r.Index, k, b[k], r.Bytes[k])
				}
			}
			seen["live records kept"]++
		}
	}
	for _, f := range ss {
		g := f.Movement()
		check(f.ID, g, f.Expected.Live)
		// A moving transporter rewrites its record's cells; its slots stay.
		if !slices.Equal(f.Expected.Container[:14], ints(HoldBytes(g)[0])[:14]) {
			t.Errorf("%s: a move changed the slots of container record 0", f.ID)
		}
	}
	for _, f := range cs {
		g := f.Movement()
		check(f.ID, g, f.Expected.Live)
		exit := f.Order.Hold >= 0
		handOver := f.Order.Hold < 0 && g.PieceAt(g.Unit).Flags4&0x1000 != 0
		for _, r := range f.Expected.Holds {
			h := holdOf(g, r.Index)
			for i, v := range h.Slots {
				if v > SlotFree || r.Bytes[i] == v {
					continue
				}
				switch {
				case exit && r.Index == f.Order.Hold && i == 7*g.Side+f.Order.Slot:
					seen["exit freed its slot"]++
				case handOver:
					seen["hand-over moved cargo"]++
				default:
					t.Errorf("%s: unit %d left slot %d of record %d", f.ID, v, i, r.Index)
				}
			}
		}
	}
	need(t, seen, "live records kept", "exit freed its slot", "hand-over moved cargo")
}

// liveWidths are the bytes each field of Live names.
var liveWidths = map[string]int{"Budget": 1, "Exp": 1, "Size": 1, "Alt": 1, "Flags4": 2, "Flags6": 2,
	"Type": 1, "Serial": 1, "Hold": 1, "Pos": 4, "Facing": 2, "Target": 2, "Action": 7}

// No foreign rules: the profile's code imports only the standard library,
// so neither the Classic/Standard tables (internal/data) nor the Python
// rules (internal/game) can enter it, and a live unit is the 26 bytes of
// the original record and nothing else (no ammo, fuel or sight field).
func gateForeign(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if first, _, _ := strings.Cut(path, "/"); strings.Contains(first, ".") {
				t.Errorf("%s imports %s", name, path)
			}
		}
		n++
	}
	if n == 0 {
		t.Fatal("no source files")
	}
	width := 0
	lt := reflect.TypeFor[Live]()
	for i := range lt.NumField() {
		w, ok := liveWidths[lt.Field(i).Name]
		if !ok {
			t.Errorf("Live.%s is not a field of the original record", lt.Field(i).Name)
		}
		width += w
	}
	if width != LiveLen {
		t.Errorf("Live names %d bytes, the record has %d", width, LiveLen)
	}
}

// Gated off: what the original leaves to memory that is not its own stays
// undefined in the profile. A path search that runs dry reads past the
// reach-bit array (the demo's image then refused with 0x48, the full
// build's walked), and on a map one row high the path search reads past the
// map; both give ErrUndefined.
func gateUndefined(t *testing.T) {
	p, tab := Full, Full.Tables.Move
	const infantry, open, deep = 0, 121, 47
	piece := Piece{Type: infantry, Flags4: tab.Types[infantry].Class, Flags6: tab.Types[infantry].Flags}
	islands := &Movement{Width: 9, Height: 9, Terrain: make([]int, 81), Units: map[int]Piece{}, Unit: 0,
		Budget: tab.Types[infantry].Budget, Block: 3}
	for i := range islands.Terrain {
		islands.Terrain[i] = deep
	}
	islands.Terrain[40], islands.Terrain[0] = open, open
	piece.Cell = 40
	islands.Units[0] = piece
	if _, err := p.Move(islands, MoveOrder{Goal: 0, Retarget: -1, Stale: []int{0}}); !errors.Is(err, ErrUndefined) {
		t.Errorf("a dry path search gives %v, not ErrUndefined", err)
	}
	row := &Movement{Width: 9, Height: 1, Terrain: make([]int, 9), Units: map[int]Piece{}, Unit: 0,
		Budget: tab.Types[infantry].Budget, Block: 3}
	for i := range row.Terrain {
		row.Terrain[i] = open
	}
	piece.Cell = 4
	row.Units[0] = piece
	if _, err := p.Route(row, 0); !errors.Is(err, ErrUndefined) {
		t.Errorf("a route on a map one row high gives %v, not ErrUndefined", err)
	}
}
