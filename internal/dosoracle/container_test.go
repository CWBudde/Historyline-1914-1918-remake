package dosoracle

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// holdCase is a move order onto a container (O.Hold < 0) or a unit leaving
// one (an exit from O.Hold's slot O.Slot).
type holdCase struct {
	class, name string
	g           *Movement
	o           ExitOrder
}

// Container case maps are 7 x 7 cells of open terrain with the moving unit
// or the container in the middle.
const (
	holdW, holdH  = 7, 7
	holdMiddle    = 24
	holdUnit      = 1  // the transporter or the other container user
	heldFirst     = 10 // the first held unit's index
	uLight        = 18 // weight 1 (infantry)
	uSupply       = 22 // a ground transporter: capacity 9, weight 4
	uTroopTrain   = 26 // capacity 15
	uCargoShip    = 31 // capacity 30, mask 0x05
	uBalloon      = 23 // live +6 bit 0x10: the exit budget 4
	tOpenBuilding = 13 // flags 0x0e0c: a 0x400 building the flag mask 3 leaves open to side 0
	transportRec  = 3  // the container record of the transporter on the map
	carriedRec    = 5  // the container record of the moving transporter
)

// holdMovement is a container map with the moving unit (index 0) of type
// mover for side at cell.
func holdMovement(tab *MoveTables, mover, side, cell int) *Movement {
	g := uniform(tab, mover, tOpen, holdW, holdH)
	p := piece(tab, mover, cell, side)
	g.Units = map[int]Piece{0: p}
	g.Side, g.Budget, g.Block = side, tab.Types[mover].Budget, playerBlock(p)
	g.Held, g.Holds = map[int]Piece{}, map[int]Hold{}
	return g
}

// addTransporter puts an own transporter of type typ (index holdUnit) at
// cell with container record rec: flag 0x20, the class mask and its first
// cell its position. A ship stands on deep water.
func addTransporter(tab *MoveTables, g *Movement, typ, cell, rec, mask int) {
	p := piece(tab, typ, cell, g.Side)
	p.Hold = rec
	g.Units[holdUnit] = p
	if tab.Types[typ].Mask&tab.Terrains[g.Terrain[cell]].Mask == 0 {
		g.Terrain[cell] = tDeep
	}
	h := dosrules.EmptyHold()
	h.Cells[0], h.Flags, h.Occupant = 2*cell+1, 0x20|mask, holdUnit
	g.Holds[rec] = h
}

// addBuilding makes cell a building of terrain with the record its flags
// select (the fourth of a table found by cell) for owner: 0 the moving
// side, 1 the other side, 2 neutral. It returns the record.
func addBuilding(tab *MoveTables, g *Movement, terrain, cell, owner int) int {
	g.Terrain[cell] = terrain
	fl := tab.Terrains[terrain].Flags
	rec := dosrules.HoldFlag40 + fl&1
	switch {
	case fl&0x400 != 0:
		rec = dosrules.HoldFlag400 + 3
	case fl&0x100 != 0:
		rec = dosrules.HoldFlag100 + 3
	}
	h := dosrules.EmptyHold()
	h.Cells[0], h.Occupant = 2*cell, dosrules.InBuilding
	h.Flags = [3]int{g.Side, 1 - g.Side, 2}[owner]
	g.Holds[rec] = h
	return rec
}

// fill puts held units of the given types into the side's first free slots
// of record rec, positioned at cell.
func fill(tab *MoveTables, g *Movement, rec, side, cell int, types ...int) {
	h := g.Holds[rec]
	for _, ty := range types {
		idx := heldFirst + len(g.Held)
		g.Held[idx] = piece(tab, ty, cell, side)
		h.Slots[7*side+dosrules.FreeSlot(h.Bytes(), side)] = idx
	}
	g.Holds[rec] = h
}

// cargoFor are units, at most six, whose weights add up to w, heaviest
// first; nil if none do.
func cargoFor(tab *MoveTables, w int) []int {
	byWeight := map[int]int{}
	for i, ty := range tab.Types {
		if _, ok := byWeight[ty.Weight]; !ok && ty.Weight > 0 {
			byWeight[ty.Weight] = i
		}
	}
	weights := slices.Sorted(maps.Keys(byWeight))
	slices.Reverse(weights)
	var out []int
	for _, wt := range weights {
		for w >= wt && len(out) < 6 {
			out = append(out, byWeight[wt])
			w -= wt
		}
	}
	if w != 0 {
		return nil
	}
	return out
}

// holdCases enumerates the container classes.
func holdCases(tab *MoveTables) []holdCase {
	var cs []holdCase
	next := toward(holdMiddle, 0, holdW, holdH)
	move := func(class, name string, g *Movement, goal int) {
		cs = append(cs, holdCase{class, name, g, ExitOrder{Hold: -1, MoveOrder: MoveOrder{Goal: goal, Retarget: -1}}})
	}

	// Loading: every type of distinct movement fields or weight into every
	// transporter type, for class masks that each type's live +6 matches or
	// misses, for both sides.
	var loaders []int
	seen := map[MoveType]bool{}
	for i, ty := range tab.Types {
		k := ty
		k.Speed, k.Capacity = 0, 0
		if !seen[k] {
			seen[k] = true
			loaders = append(loaders, i)
		}
	}
	masks := []int{0, 0x40, 0x100, 0x200, 0x300, 0xFC0}
	for _, side := range []int{0, 1} {
		for _, ty := range loaders {
			for _, tr := range []int{uSupply, uTroopTrain, uCargoShip} {
				for _, mask := range masks {
					g := holdMovement(tab, ty, side, holdMiddle)
					addTransporter(tab, g, tr, next, transportRec, mask)
					move("load", fmt.Sprintf("side %d type %d into %d mask %#x", side, ty, tr, mask), g, next)
				}
			}
		}
	}

	// The seven slots: k of the side's slots taken, the other side's all
	// taken.
	for _, side := range []int{0, 1} {
		for k := range 8 {
			g := holdMovement(tab, uLight, side, holdMiddle)
			addTransporter(tab, g, uTroopTrain, next, transportRec, 0xFC0)
			fill(tab, g, transportRec, side, next, slices.Repeat([]int{uLight}, k)...)
			fill(tab, g, transportRec, 1-side, next, slices.Repeat([]int{uLight}, 7)...)
			move("slots", fmt.Sprintf("side %d %d taken", side, k), g, next)
		}
	}

	// Weights one below, at and one above each transporter's capacity.
	for _, tr := range []int{uSupply, uTroopTrain, uCargoShip} {
		for d := -1; d <= 1; d++ {
			mover := uLight
			cargo := cargoFor(tab, tab.Types[tr].Capacity-tab.Types[mover].Weight+d)
			if cargo == nil {
				panic(fmt.Sprintf("no cargo of weight %d", tab.Types[tr].Capacity-tab.Types[mover].Weight+d))
			}
			g := holdMovement(tab, mover, 0, holdMiddle)
			addTransporter(tab, g, tr, next, transportRec, 0xFC0)
			fill(tab, g, transportRec, 0, next, cargo...)
			move("weight", fmt.Sprintf("into %d at capacity %+d", tr, d), g, next)
		}
	}

	// A transporter that has moved (+4 bit 0x200).
	g := holdMovement(tab, uLight, 0, holdMiddle)
	addTransporter(tab, g, uTroopTrain, next, transportRec, 0xFC0)
	p := g.Units[holdUnit]
	p.Flags4 |= 0x200
	g.Units[holdUnit] = p
	move("moved", "transporter moved", g, next)

	// Nested cargo: a transporter carrying k units into a transporter or an
	// own building with f free slots.
	for _, k := range []int{0, 1, 3} {
		for _, f := range []int{k, k + 1, 7} {
			for _, into := range []string{"transporter", "building"} {
				g := holdMovement(tab, uSupply, 0, holdMiddle)
				p := g.Units[0]
				p.Hold = carriedRec
				g.Units[0] = p
				carried := dosrules.EmptyHold()
				carried.Cells[0], carried.Flags, carried.Occupant = 2*holdMiddle+1, 0x20|0xFC0, 0
				g.Holds[carriedRec] = carried
				fill(tab, g, carriedRec, 0, holdMiddle, slices.Repeat([]int{uLight}, k)...)
				rec := transportRec
				if into == "transporter" {
					addTransporter(tab, g, uCargoShip, next, transportRec, 0xFC0)
				} else {
					rec = addBuilding(tab, g, tOpenBuilding, next, 0)
				}
				fill(tab, g, rec, 0, next, slices.Repeat([]int{uLight}, 7-f)...)
				move("nested", fmt.Sprintf("%d carried into %s with %d free", k, into, f), g, next)
			}
		}
	}

	// Buildings of every building terrain, own, foreign and neutral, empty
	// and full, for both sides; a cell no record lists, and one listed as a
	// building's second cell.
	for _, side := range []int{0, 1} {
		for tr := range tab.Terrains {
			if !building(tab, tr) {
				continue
			}
			for owner := range 3 {
				for _, full := range []bool{false, true} {
					g := holdMovement(tab, uLight, side, holdMiddle)
					rec := addBuilding(tab, g, tr, next, owner)
					if full {
						fill(tab, g, rec, side, next, slices.Repeat([]int{uLight}, 7)...)
					}
					move("building", fmt.Sprintf("side %d terrain %d owner %d full %t", side, tr, owner, full), g, next)
				}
			}
		}
	}
	g = holdMovement(tab, uLight, 0, holdMiddle)
	rec := addBuilding(tab, g, tBuilding, next, 0)
	delete(g.Holds, rec)
	move("building", "unlisted cell", g, next)
	g = holdMovement(tab, uLight, 0, holdMiddle)
	rec = addBuilding(tab, g, tBuilding, next, 0)
	other := toward(next, 0, holdW, holdH)
	g.Terrain[other] = tBuilding
	h := g.Holds[rec]
	h.Cells[0], h.Cells[1] = 2*other, 2*next
	g.Holds[rec] = h
	move("building", "second cell", g, next)

	// Carried cargo on a plain move and onto a foreign building: it stays.
	for _, onto := range []string{"open", "foreign building"} {
		g := holdMovement(tab, uSupply, 0, holdMiddle)
		p := g.Units[0]
		p.Hold = carriedRec
		g.Units[0] = p
		carried := dosrules.EmptyHold()
		carried.Cells[0], carried.Flags = 2*holdMiddle+1, 0x20|0xFC0
		g.Holds[carriedRec] = carried
		fill(tab, g, carriedRec, 0, holdMiddle, uLight, uLight)
		goal := toward(next, 0, holdW, holdH)
		if onto != "open" {
			addBuilding(tab, g, tOpenBuilding, goal, 1)
		}
		move("carry", "cargo onto "+onto, g, goal)
	}

	// Exits from a transporter (a ship on water too) and an own building,
	// for units with and without live +6 bit 0x10 (in slot 1), live byte 0
	// at 0 and 255, to every third cell the exit reaches and one out of
	// reach.
	exit := func(name string, g *Movement, rec int) {
		probe, err := tab.Exit(g, ExitOrder{Hold: rec, Slot: 1, MoveOrder: MoveOrder{Goal: 0, Retarget: -1}})
		if err != nil {
			panic(fmt.Sprintf("exit %s: %v", name, err))
		}
		var goals []int
		for c := range g.Width * g.Height {
			if v := probe.Remaining[c/g.Width*64+c%g.Width]; v >= 0 && c != holdMiddle {
				goals = append(goals, c)
			}
		}
		for i, goal := range goals {
			if i%3 == 0 {
				cs = append(cs, holdCase{"exit", name, g, ExitOrder{Hold: rec, Slot: 1, MoveOrder: MoveOrder{Goal: goal, Retarget: -1}}})
			}
		}
		cs = append(cs, holdCase{"exit", name + " out of reach", g, ExitOrder{Hold: rec, Slot: 1, MoveOrder: MoveOrder{Goal: 0, Retarget: -1}}})
	}
	for _, side := range []int{0, 1} {
		for _, ty := range []int{uLight, uBalloon, uInfantry} {
			for _, b0 := range []int{0, 255} {
				for _, from := range []string{"transporter", "ship", "building"} {
					g := holdMovement(tab, ty, side, holdMiddle)
					delete(g.Units, 0)
					rec := transportRec
					switch from {
					case "transporter":
						addTransporter(tab, g, uSupply, holdMiddle, transportRec, 0xFC0)
					case "ship":
						addTransporter(tab, g, uCargoShip, holdMiddle, transportRec, 0xFC0)
					default:
						rec = addBuilding(tab, g, tBuilding, holdMiddle, 0)
					}
					fill(tab, g, rec, side, holdMiddle, uLight, ty)
					g.Unit, g.Budget = heldFirst+1, b0
					exit(fmt.Sprintf("side %d type %d from %s byte 0 %d", side, ty, from, b0), g, rec)
				}
			}
		}
	}
	return cs
}

// holdsAgree runs a container case in the emulator and in the
// transcription and compares every output, the container records and the
// exit's reach buffer among them.
func holdsAgree(t *testing.T, e *Export, tab *MoveTables, c holdCase) (r ExitRun, compared, executed bool) {
	t.Helper()
	var got, want ExitRun
	var gerr, werr error
	if c.o.Hold < 0 {
		got.StepRun, gerr = e.RunMove(c.g, c.o.MoveOrder)
		want.StepRun, werr = tab.Execute(c.g, c.o.MoveOrder)
	} else {
		got, gerr = e.RunExit(c.g, c.o)
		want, werr = tab.Exit(c.g, c.o)
	}
	switch {
	case errors.Is(werr, ErrUndefined) && gerr == nil:
		return got, false, true
	case gerr != nil || werr != nil:
		t.Fatalf("%s/%s to %d: executed %v, transcribed %v", c.class, c.name, c.o.Goal, gerr, werr)
	}
	d := stepDiff(got.StepRun, want.StepRun)
	if d == "" {
		d = holdDiff(got, want)
	}
	if d != "" {
		t.Errorf("%s/%s to %d: %s", c.class, c.name, c.o.Goal, d)
	}
	return got, true, true
}

// holdDiff names the first container record or exit output that differs.
func holdDiff(got, want ExitRun) string {
	for i := range dosrules.HoldRecords {
		if !bytes.Equal(got.Holds[i], want.Holds[i]) {
			return fmt.Sprintf("container record %d executed % x, transcribed % x", i, got.Holds[i], want.Holds[i])
		}
	}
	if got.Unit != want.Unit && want.Remaining != nil {
		return fmt.Sprintf("exit unit executed %d, transcribed %d", got.Unit, want.Unit)
	}
	for i := range want.Remaining {
		if got.Remaining[i] != want.Remaining[i] {
			return fmt.Sprintf("exit buffer at row %d column %d executed %d, transcribed %d", i/64, i%64, got.Remaining[i], want.Remaining[i])
		}
	}
	return ""
}

// TestContainerCoverage runs every container case in the emulator and the
// transcription and checks the rules the cases are there to show.
func TestContainerCoverage(t *testing.T) { eachExport(t, containerCoverage) }

func containerCoverage(t *testing.T, e *Export) {
	needMove(t, e)
	if e.B.Move.Exit[0] == 0 {
		t.Skip("the exit branch is not located in this build")
	}
	tab := moveTablesOf(e)
	perClass, skipped := map[string]int{}, map[string]int{}
	seen := map[string]int{}
	began := time.Now()
	for _, c := range holdCases(tab) {
		r, compared, executed := holdsAgree(t, e, tab, c)
		if compared {
			perClass[c.class]++
		} else {
			skipped[c.class]++
		}
		if executed {
			holdRules(t, tab, c, r, seen)
		}
	}
	for _, class := range []string{"load", "slots", "weight", "moved", "nested", "building", "carry", "exit"} {
		if perClass[class] == 0 {
			t.Errorf("class %s: no case compared", class)
		}
	}
	t.Logf("compared %v, not compared %v in %v", perClass, skipped, time.Since(began).Round(time.Millisecond))
	keys := slices.Collect(maps.Keys(seen))
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("%s: %d", k, seen[k])
	}
	for _, k := range []string{"loaded", "refused: classes (0x3a)", "refused: moved (0x3a)", "refused: slots taken (0x3b)",
		"refused: weight (0x3b)", "loaded at capacity", "entered an own building", "refused: building full (0x3b)",
		"onto a foreign building", "onto a neutral building", "cargo handed over", "cargo kept: too few slots (0x3b)",
		"cargo carried along", "exited", "exit budget 4", "exit budget 8", "exit refused"} {
		if seen[k] == 0 {
			t.Errorf("no case showed %q", k)
		}
	}
}

// slotted maps every unit in a slot to its record.
func slotted(holds map[int][]byte) map[int]int {
	out := map[int]int{}
	for i, h := range holds {
		for k := range 14 {
			if h[k] <= dosrules.SlotFree {
				out[int(h[k])] = i
			}
		}
	}
	return out
}

// holdRules checks the findings on the executed outcome and counts them.
func holdRules(t *testing.T, tab *MoveTables, c holdCase, r ExitRun, seen map[string]int) {
	t.Helper()
	g := c.g
	u, side := g.Unit, g.Side
	fail := func(format string, args ...any) {
		t.Errorf("%s/%s to %d: "+format, append([]any{c.class, c.name, c.o.Goal}, args...)...)
	}
	before, after := slotted(dosrules.HoldBytes(g)), slotted(r.Holds)
	moved := r.Order&1 != 0 && r.Start&1 != 0
	shown := 0
	if len(r.Messages) > 0 {
		shown = r.Messages[0][0]
	}

	// No unit leaves a slot except the exiting unit, and a carried unit only
	// for another slot when its transporter enters a container.
	for idx, rec := range before {
		if c.o.Hold >= 0 && idx == r.Unit {
			continue
		}
		if _, ok := after[idx]; !ok {
			fail("unit %d left record %d without an exit", idx, rec)
		}
	}
	// Every unit taken in has its record's capacity respected and the
	// side's seven slots.
	weights := map[int]int{}
	for idx, rec := range after {
		if _, ok := before[idx]; !ok || before[idx] != rec {
			weights[rec] = 0
		}
	}
	for rec := range weights {
		if rec >= dosrules.HoldFlag40 {
			continue
		}
		for k := range 7 {
			if v := r.Holds[rec][7*side+k]; v <= dosrules.SlotFree {
				weights[rec] += tab.Types[r.Live[int(v)][8]].Weight
			}
		}
		if cap := tab.Types[g.Units[holdUnit].Type].Capacity; weights[rec] > cap {
			fail("record %d carries weight %d over capacity %d", rec, weights[rec], cap)
		}
	}

	if c.o.Hold < 0 {
		_, a, _ := tab.EndCheck(g, c.o.Goal)
		switch {
		case moved && a.Kind == 2:
			seen["loaded"]++
			if after[u] != a.Hold || r.Maps[0][2*c.o.Goal+1] != byte(a.Occupant) {
				fail("not in the transporter's record, or the map lost the transporter")
			}
			if c.class == "weight" {
				seen["loaded at capacity"] += boolInt(c.name[len(c.name)-2:] == "+0")
			}
		case moved && a.Kind == 5:
			seen["entered an own building"]++
			if after[u] != a.Hold || r.Maps[0][2*c.o.Goal+1] != dosrules.InBuilding || word(r.Live[u], 0x0B+2*side)&1 != 0 {
				fail("not in the building's record, the map not 0xf4 or the position odd")
			}
		case moved && a.Kind == 4:
			if word(g.Holds[a.Hold].Bytes(), 0x19)&2 != 0 {
				seen["onto a neutral building"]++
			} else {
				seen["onto a foreign building"]++
			}
			if _, ok := after[u]; ok || r.Maps[0][2*c.o.Goal+1] != byte(u) {
				fail("a foreign building took the unit in")
			}
		case shown == 0x3A && g.Units[holdUnit].Flags4&0x200 != 0:
			seen["refused: moved (0x3a)"]++
		case shown == 0x3A:
			seen["refused: classes (0x3a)"]++
		case shown == 0x3B && c.class == "slots":
			seen["refused: slots taken (0x3b)"]++
		case shown == 0x3B && c.class == "weight":
			seen["refused: weight (0x3b)"]++
		case shown == 0x3B && c.class == "building":
			seen["refused: building full (0x3b)"]++
		case shown == 0x3B && c.class == "nested":
			seen["cargo kept: too few slots (0x3b)"]++
		}
		if moved && g.Units[0].Flags4&0x1000 != 0 && g.Holds[carriedRec].Slots[0] <= dosrules.SlotFree {
			cargo := int(g.Holds[carriedRec].Slots[0])
			switch a.Kind {
			case 2, 5:
				if after[cargo] != a.Hold {
					fail("the carried unit %d is in record %d, not the target's %d", cargo, after[cargo], a.Hold)
				}
				seen["cargo handed over"]++
			default:
				if after[cargo] != carriedRec || word(r.Live[cargo], 0x0B+2*side) != word(r.Live[u], 0x0B+2*side) {
					fail("the carried unit %d left its transporter or its position", cargo)
				}
				seen["cargo carried along"]++
			}
		}
		return
	}

	// Exits: the reach starts at 4 or 8, whatever live byte 0 holds; the
	// cell left gets the record's occupant back.
	l := dosrules.LiveRecords(g)[r.Unit]
	start := word(g.Holds[c.o.Hold].Bytes(), 0x0E) &^ 1
	budget := int(r.Remaining[start/2/g.Width*64+start/2%g.Width])
	want := 8
	if word(l, 6)&0x10 != 0 {
		want = 4
	}
	if budget != want {
		fail("exit budget %d, not %d", budget, want)
	}
	seen[fmt.Sprintf("exit budget %d", budget)]++
	if !moved {
		seen["exit refused"]++
		return
	}
	seen["exited"]++
	if r.Maps[0][start+1] != byte(g.Holds[c.o.Hold].Occupant) {
		fail("the container's cell shows %#x, not its occupant %#x", r.Maps[0][start+1], g.Holds[c.o.Hold].Occupant)
	}
	if _, ok := after[r.Unit]; ok {
		fail("the exiting unit is still in a slot")
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
