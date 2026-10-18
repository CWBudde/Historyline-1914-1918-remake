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
)

// stepCase is one move order of a coverage class: the map, the goal and the
// destination the order is changed to before the start (-1: none).
type stepCase struct {
	class, name string
	g           *Movement
	o           MoveOrder
}

// stepsAgree runs an order in the emulator and in the transcription and
// compares every output. It returns the emulator's outcome, whether the two
// were compared and whether the emulator ran it: the transcription is
// undefined in some (ErrUndefined).
func stepsAgree(t *testing.T, e *Export, tab *MoveTables, c stepCase) (r StepRun, compared, executed bool) {
	t.Helper()
	got, gerr := e.RunMove(c.g, c.o)
	want, werr := tab.Execute(c.g, c.o)
	switch {
	case errors.Is(werr, ErrUndefined) && gerr == nil:
		return got, false, true
	case errors.Is(werr, ErrUndefined):
		return got, false, false
	case gerr != nil || werr != nil:
		t.Fatalf("%s/%s to %d: executed %v, transcribed %v", c.class, c.name, c.o.Goal, gerr, werr)
	}
	if d := stepDiff(got, want); d != "" {
		t.Errorf("%s/%s to %d: %s", c.class, c.name, c.o.Goal, d)
	}
	return got, true, true
}

// stepDiff names the first output in which two outcomes differ.
func stepDiff(got, want StepRun) string {
	switch {
	case got.Order != want.Order || got.Start != want.Start:
		return fmt.Sprintf("order/start executed %#x/%#x, transcribed %#x/%#x", got.Order, got.Start, want.Order, want.Start)
	case !slices.EqualFunc(got.Messages, want.Messages, func(a, b [2]int) bool { return a == b }):
		return fmt.Sprintf("messages executed %v, transcribed %v", got.Messages, want.Messages)
	case !slices.Equal(got.Cells, want.Cells):
		return fmt.Sprintf("cells executed %v, transcribed %v", got.Cells, want.Cells)
	case got.Count != want.Count:
		return fmt.Sprintf("count executed %d, transcribed %d", got.Count, want.Count)
	case !slices.Equal(got.Orders, want.Orders):
		return fmt.Sprintf("order list executed %v, transcribed %v", got.Orders, want.Orders)
	case !bytes.Equal(got.Container, want.Container):
		return fmt.Sprintf("container executed % x, transcribed % x", got.Container, want.Container)
	}
	for _, idx := range slices.Sorted(maps.Keys(got.Live)) {
		if !bytes.Equal(got.Live[idx], want.Live[idx]) {
			return fmt.Sprintf("live record %d executed % x, transcribed % x", idx, got.Live[idx], want.Live[idx])
		}
	}
	for k := range 2 {
		for i := range got.Maps[k] {
			if got.Maps[k][i] != want.Maps[k][i] {
				return fmt.Sprintf("map %d byte %d executed %#x, transcribed %#x", k, i, got.Maps[k][i], want.Maps[k][i])
			}
		}
	}
	for i := range got.Bits {
		if got.Bits[i] != want.Bits[i] {
			return fmt.Sprintf("bits at row %d column %d executed %#x, transcribed %#x", i/64, i%64, got.Bits[i], want.Bits[i])
		}
	}
	return ""
}

// stepGoals caps the routes taken from a movement case per class (6 where
// the class is not named).
var stepGoals = map[string]int{"terrain": 2, "transition": 1, "random": 3}

// stepCases are the movement cases' routes (stepGoals of them per case; for
// a case whose destination list stops at a building, the reached cells with
// end code 0) and the order cases.
func stepCases(tab *MoveTables) []stepCase {
	var cs []stepCase
	for _, mc := range moveCases(tab) {
		// The game places a two-cell ship's head and rear together.
		if tab.Types[mc.g.Units[mc.g.Unit].Type].Class&0x40 != 0 {
			if _, ok := mc.g.Units[mc.g.Unit+1]; !ok {
				continue
			}
		}
		goals := mc.g.Goals
		if len(goals) == 0 {
			_, bits, _ := tab.Reach(mc.g)
			for c := range mc.g.Width * mc.g.Height {
				code, err := tab.EndCode(mc.g, c)
				if bits[c/mc.g.Width*64+c%mc.g.Width]&byte(1+mc.g.Side) != 0 && err == nil && code == 0 && !holdCell(tab, mc.g, c) {
					goals = append(goals, c)
				}
			}
		}
		n := stepGoals[mc.class]
		if n == 0 {
			n = 6
		}
		if len(goals) > n {
			kept := make([]int, n)
			for i := range kept {
				kept[i] = goals[i*len(goals)/n]
			}
			goals = kept
		}
		for _, goal := range goals {
			if holdCell(tab, mc.g, goal) {
				continue // containers have their own cases (container_test.go)
			}
			cs = append(cs, stepCase{mc.class, mc.name, mc.g, MoveOrder{Goal: goal, Retarget: -1}})
		}
	}

	// Orders the order code refuses: a cell out of reach, the unit's own
	// cell, an own, an enemy and a neutral unit, terrain closed to air units.
	add := func(name string, g *Movement, goal int) {
		cs = append(cs, stepCase{"order", name, g, MoveOrder{Goal: goal, Retarget: -1}})
	}
	g := uniform(tab, uInfantry, tOpen, 9, 9)
	add("out of reach", g, 0)
	add("own cell", g, 40)
	for _, rel := range []struct {
		name string
		side int
	}{{"own unit", 0}, {"enemy", 1}, {"neutral", 2}} {
		g := uniform(tab, uInfantry, tOpen, 9, 9)
		g.Units[1] = piece(tab, uInfantry, toward(40, 0, 9, 9), rel.side)
		add(rel.name, g, toward(40, 0, 9, 9))
	}
	g = uniform(tab, uAir, tOpen, 9, 9)
	g.Terrain[toward(40, 2, 9, 9)] = tClosed
	add("air onto closed terrain", g, toward(40, 2, 9, 9))
	g = uniform(tab, uInfantry, tShallow, 9, 9)
	g.Terrain[40] = tOpen
	add("terrain the mask does not enter", g, toward(40, 1, 9, 9))

	// An order changed between the order code and the start.
	g = uniform(tab, uInfantry, tOpen, 9, 9)
	cs = append(cs, stepCase{"retarget", "another cell", g, MoveOrder{Goal: toward(40, 3, 9, 9), Retarget: toward(40, 0, 9, 9)}},
		stepCase{"retarget", "the same cell", g, MoveOrder{Goal: toward(40, 3, 9, 9), Retarget: toward(40, 3, 9, 9)}})

	// A reach bit left on a cell the search cannot reach: the path search
	// runs dry. An infantry unit and the stale cell on two islands of open
	// terrain in deep water.
	g = uniform(tab, uInfantry, tDeep, 9, 9)
	g.Terrain[40], g.Terrain[0] = tOpen, tOpen
	cs = append(cs, stepCase{"order", "stale reach bit", g, MoveOrder{Goal: 0, Retarget: -1, Stale: []int{0}}})

	// Passing: a line of own units, a hidden enemy on the route and on the
	// destination, a two-cell ship and an empty transporter.
	g = uniform(tab, uInfantry, tCheap, 9, 9)
	for i, c := range []int{3*9 + 4, 2*9 + 4} {
		g.Units[i+1] = piece(tab, uInfantry, c, 0)
	}
	cs = append(cs, stepCase{"pass", "two own units", g, MoveOrder{Goal: 9 + 4, Retarget: -1}})
	for _, at := range []int{3*9 + 4, 2*9 + 4} {
		g := uniform(tab, uInfantry, tCheap, 9, 9)
		g.Unseen = map[int]Piece{1: piece(tab, uInfantry, at, 1)}
		name := "hidden enemy on the route"
		if at == 2*9+4 {
			name = "hidden enemy on the destination"
		}
		cs = append(cs, stepCase{"pass", name, g, MoveOrder{Goal: 2*9 + 4, Retarget: -1}})
	}
	for _, goal := range []int{9 + 4, 2*9 + 6, 6*9 + 2} {
		g := uniform(tab, uShipHead, tShallow, 9, 9)
		g.Units[1] = piece(tab, uShipRear, 3*9+4, 0)
		cs = append(cs, stepCase{"pass", fmt.Sprintf("two-cell ship to %d", goal), g, MoveOrder{Goal: goal, Retarget: -1}})
	}
	for _, goal := range []int{9 + 4, 6*9 + 6} {
		g := uniform(tab, uTransporter, tOpen, 9, 9)
		cs = append(cs, stepCase{"pass", fmt.Sprintf("empty transporter to %d", goal), g, MoveOrder{Goal: goal, Retarget: -1}})
	}
	return cs
}

// TestStepCoverage runs every move order in the emulator and the
// transcription and checks the rules the cases are there to show.
func TestStepCoverage(t *testing.T) { eachExport(t, stepCoverage) }

func stepCoverage(t *testing.T, e *Export) {
	needMove(t, e)
	tab := moveTablesOf(e)
	perClass, skipped := map[string]int{}, map[string]int{}
	seen := map[string]int{}
	began := time.Now()
	for _, c := range stepCases(tab) {
		r, compared, executed := stepsAgree(t, e, tab, c)
		if compared {
			perClass[c.class]++
		} else {
			skipped[c.class]++
		}
		if executed {
			stepRules(t, tab, c, r, seen)
		}
	}
	for _, class := range []string{"terrain", "transition", "units", "unseen", "block", "budget", "edge", "random", "order", "retarget", "pass"} {
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
	for _, k := range []string{"moved", "through a zone of control", "over the budget", "passed an own unit",
		"passed a hidden unit", "ended on a hidden unit", "refused", "retargeted", "refused with message 0x48",
		"refused with message 0x39", "two-cell ship", "transporter"} {
		if seen[k] == 0 {
			t.Errorf("no case showed %q", k)
		}
	}
}

// stepRules checks the findings on the executed outcome and counts them.
func stepRules(t *testing.T, tab *MoveTables, c stepCase, r StepRun, seen map[string]int) {
	t.Helper()
	g := c.g
	u, side := g.Unit, g.Side
	fail := func(format string, args ...any) {
		t.Errorf("%s/%s to %d: "+format, append([]any{c.class, c.name, c.o.Goal}, args...)...)
	}
	before := liveRecords(g)
	ownMap, other := movementMaps(g)
	if !bytes.Equal(r.Maps[1], other) {
		fail("the other side's map changed")
	}
	for idx, rec := range r.Live {
		if rec[0] != before[idx][0] {
			fail("live byte 0 of unit %d changed from %d to %d", idx, before[idx][0], rec[0])
		}
		if word(rec, 0x0B+2*(1-side)) != word(before[idx], 0x0B+2*(1-side)) {
			fail("unit %d moved as the other side sees it", idx)
		}
	}
	if r.Start&1 == 0 {
		if len(r.Cells) != 0 || !bytes.Equal(r.Maps[0], ownMap) {
			fail("a refused order moved the unit")
		}
		seen["refused"]++
		for _, m := range r.Messages {
			seen[fmt.Sprintf("refused with message %#x", m[0])]++
		}
		if c.o.Retarget >= 0 {
			seen["retargeted"]++
		}
		return
	}
	seen["moved"]++
	if r.Cells[len(r.Cells)-1] != c.o.Goal {
		fail("arrived at %d", r.Cells[len(r.Cells)-1])
	}
	if !slices.Equal(r.Orders, []int{u}) {
		fail("order list %v", r.Orders)
	}
	if n := len(r.Messages); n == 0 || r.Messages[n-1] != [2]int{0x38, side} {
		fail("messages %v", r.Messages)
	}
	if word(r.Live[u], 4)&0x200 == 0 {
		fail("no moved bit")
	}
	route := append([]int{g.Units[u].Cell}, r.Cells...)
	spent, zoc := 0, false
	occ := g.occupants()
	enemyNext := func(cell int) bool {
		for _, n := range Ring(cell, g.Width, g.Height) {
			if o, ok := occ[n]; ok && n >= 0 && g.Units[o].Flags4&2 == 0 && !own(g.Units[o].Flags4, side) {
				return true
			}
		}
		return false
	}
	for i, cell := range route[1:] {
		spent += stepCost(tab, g, cell)
		if i < len(route)-2 && enemyNext(cell) {
			zoc = true
		}
	}
	if zoc {
		seen["through a zone of control"]++
	}
	if spent > g.Budget {
		seen["over the budget"]++
	}
	ship := word(r.Live[u], 4)&0x40 != 0
	if ship {
		seen["two-cell ship"]++
	}
	if word(r.Live[u], 4)&0x1000 != 0 {
		seen["transporter"]++
		if word(r.Container, 0x0E) != 2*c.o.Goal+1 {
			fail("container record at %#x", word(r.Container, 0x0E))
		}
	}
	// On the side's map only the start, the destination and, for a ship,
	// the rear's cells change; passed occupants are back.
	for cell := range g.Width * g.Height {
		want := ownMap[2*cell+1]
		switch {
		case cell == c.o.Goal:
			want = byte(u)
		case ship && cell == route[len(route)-2]:
			want = byte(u + 1)
		case cell == g.Units[u].Cell || ship && cell == g.Units[u+1].Cell:
			want = 0xFF
		}
		if r.Maps[0][2*cell+1] != want {
			fail("side's map at %d holds %#x, want %#x", cell, r.Maps[0][2*cell+1], want)
		}
	}
	for _, cell := range route[1 : len(route)-1] {
		if _, ok := occ[cell]; ok && (!ship || g.Units[u+1].Cell != cell) {
			seen["passed an own unit"]++
		}
		for _, h := range g.Unseen {
			if h.Cell == cell {
				seen["passed a hidden unit"]++
			}
		}
	}
	for _, h := range g.Unseen {
		if h.Cell == c.o.Goal {
			seen["ended on a hidden unit"]++
		}
	}
}
