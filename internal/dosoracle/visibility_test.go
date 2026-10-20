package dosoracle

import (
	"bytes"
	"fmt"
	"maps"
	"math/rand/v2"
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// R2 visibility: the phase change 02dc:1a09-1cc6 (RunPhaseChange) against
// ChangePhase, and attack orders resolved on the attacking side's view
// while the other side's view differs (TestEngagementHidden).

const phaseW, phaseH = 8, 6

// phaseCase is one phase change of a coverage class.
type phaseCase struct {
	class, name string
	g           *PhaseChange
}

// phaseMaps are two maps with the terrain records 1-7 by cell, the second
// shifted by shift, and no occupants.
func phaseMaps(shift int) [2][]int {
	var out [2][]int
	for k := range 2 {
		out[k] = make([]int, 2*phaseW*phaseH)
		for c := range phaseW * phaseH {
			out[k][2*c], out[k][2*c+1] = 1+(c+k*shift)%7, 0xFF
		}
	}
	return out
}

// at is the position word of a cell.
func at(cell int) int { return 2*cell + 1 }

// phaseBase is a phase change with no units: side a attacked, the phases
// as the handler leaves them (a's phase byte 2).
func phaseBase(a int) *PhaseChange {
	g := &PhaseChange{Width: phaseW, Height: phaseH, Maps: phaseMaps(0), Units: map[int]Sighted{},
		Holds: map[int]Hold{}, Attacker: a}
	g.Phase[a], g.Phase[1-a] = 2, 1
	return g
}

// place puts a unit into both maps where each side sees it.
func place(g *PhaseChange, idx int, u Sighted) {
	g.Units[idx] = u
	for k := range 2 {
		if u.Pos[k] < len(g.Maps[k]) {
			g.Maps[k][u.Pos[k]] = idx
		}
	}
}

// sideFlags are live +4 values of side 0, side 1, a neutral unit and both
// side bits.
var sideFlags = []int{0, 1, 2, 3}

func phaseCases() []phaseCase {
	var cases []phaseCase
	add := func(class, name string, g *PhaseChange) {
		cases = append(cases, phaseCase{class, fmt.Sprintf("attacker %d %s", g.Attacker, name), g})
	}
	for a := range 2 {
		// Phase bytes and the turn counter.
		for _, ph := range [][2]int{{2, 1}, {1, 2}, {1, 1}, {2, 2}, {0, 0}, {3, 0}} {
			for _, count := range []int{0, 7, 0xFFFF} {
				g := phaseBase(a)
				g.Phase, g.Count = ph, count
				place(g, 3, Sighted{Flags4: 0, Size: 5, Pos: [2]int{at(9), at(9)}})
				add("phase", fmt.Sprintf("phases %v count %d", ph, count), g)
			}
		}
		// Each side flag with views that differ either way, and facing.
		for _, f := range sideFlags {
			for _, pos := range [][2]int{{at(10), at(10)}, {at(10), at(20)}, {at(20), at(10)}} {
				g := phaseBase(a)
				place(g, 5, Sighted{Flags4: f | 0x1C, Size: 4, Pos: pos, Facing: [2]int{1, 4}})
				place(g, 6, Sighted{Flags4: 1 - f&1, Size: 3, Pos: [2]int{at(30), at(31)}, Facing: [2]int{2, 5}})
				add("sides", fmt.Sprintf("flags %d at %v", f, pos), g)
			}
		}
		// Destroyed, dying and held units.
		for _, f := range sideFlags[:2] {
			for _, life := range []struct {
				name   string
				flags  int
				size   int
				onMaps bool
			}{
				{"alive", 0, 6, true}, {"dying", 0, 0, true}, {"destroyed", 0x8000, 0, false},
				{"destroyed with strength", 0x8000, 3, false}, {"in a container", 0x4000, 4, false},
				{"in a container dying", 0x4000, 0, false},
			} {
				g := phaseBase(a)
				u := Sighted{Flags4: f | life.flags, Size: life.size, Pos: [2]int{at(12), at(13)}}
				if life.onMaps {
					place(g, 40, u)
				} else {
					g.Units[40] = u
				}
				place(g, 41, Sighted{Flags4: 1 - f, Size: 2, Pos: [2]int{at(14), at(14)}})
				add("life", fmt.Sprintf("flags %d %s", f, life.name), g)
			}
		}
		// Two units on one side-0 position word; the higher index is written
		// last.
		for _, f := range [][2]int{{0, 0}, {0, 1}, {1, 1}} {
			g := phaseBase(a)
			place(g, 7, Sighted{Flags4: f[0], Size: 2, Pos: [2]int{at(16), at(17)}})
			place(g, 70, Sighted{Flags4: f[1], Size: 2, Pos: [2]int{at(16), at(18)}})
			add("stale", fmt.Sprintf("flags %v on one cell", f), g)
		}
		// A side-1 unit whose side-0 word is stale when side 1 attacks.
		for _, pos := range [][2]int{{at(22), at(23)}, {at(0), at(23)}, {0, at(23)}, {2 * 25, at(25)}} {
			g := phaseBase(a)
			place(g, 9, Sighted{Flags4: 1, Size: 3, Pos: pos})
			add("stale", fmt.Sprintf("side 1 at %v", pos), g)
		}
		// Two-cell ships: head (0x40|0x80) and rear (0x40) records.
		for _, f := range sideFlags[:2] {
			g := phaseBase(a)
			place(g, 20, Sighted{Flags4: f | 0xC0, Size: 4, Pos: [2]int{at(26), at(27)}})
			place(g, 21, Sighted{Flags4: f | 0x40, Size: 4, Pos: [2]int{at(34), at(35)}})
			add("ships", fmt.Sprintf("flags %d", f), g)
		}
		// Terrain that differs between the maps, and building markers.
		for _, shift := range []int{0, 3} {
			g := phaseBase(a)
			g.Maps = phaseMaps(shift)
			g.Maps[0][at(5)], g.Maps[1][at(6)] = dosrules.InBuilding, dosrules.InBuilding
			place(g, 11, Sighted{Flags4: 1, Size: 3, Pos: [2]int{at(8), at(8)}})
			add("terrain", fmt.Sprintf("shift %d", shift), g)
		}
		cases = append(cases, holdPhaseCases(a)...)
		rng := rand.New(rand.NewPCG(uint64(a), 1914))
		for i := range 60 {
			add("random", fmt.Sprintf("%d", i), randomPhase(rng, a))
		}
	}
	return cases
}

// holdPhaseCases sync the slots of a transporter record and of each
// building table under every owner, with live, dying, destroyed and
// unlisted cargo.
func holdPhaseCases(a int) []phaseCase {
	var cases []phaseCase
	for _, rec := range []int{3, dosrules.HoldFlag40, dosrules.HoldFlag40 + 1, dosrules.HoldFlag100 + 4, dosrules.HoldFlag400 + 9} {
		for _, owner := range []int{0, 1, 2, 3, 0x8000, 0x8001} {
			g := phaseBase(a)
			h := dosrules.EmptyHold()
			h.Flags = owner | 0x20*b2i(rec < dosrules.HoldFlag40)
			// Side 0's slots: live, dying, destroyed, unlisted; side 1's: live
			// and a destroyed unit.
			g.Units[50] = Sighted{Flags4: 0x4000, Size: 3, Pos: [2]int{at(2), at(2)}}
			g.Units[51] = Sighted{Flags4: 0x4000, Size: 0, Pos: [2]int{at(2), at(2)}}
			g.Units[52] = Sighted{Flags4: 0x8000 | 1, Size: 0, Pos: [2]int{at(2), at(2)}}
			g.Units[53] = Sighted{Flags4: 0x4001, Size: 2, Pos: [2]int{at(2), at(2)}}
			g.Units[54] = Sighted{Flags4: 0x8000, Size: 1, Pos: [2]int{at(2), at(2)}}
			copy(h.Slots[:], []int{50, 0xFF, 51, 52, 99, 0xFF, 0xFF, 53, 54, 0xFF, 0xFF, 0xFF, 0xFF, 50})
			g.Holds[rec] = h
			cases = append(cases, phaseCase{"holds", fmt.Sprintf("attacker %d record %d owner %#x", a, rec, owner), g})
		}
	}
	return cases
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// randomPhase is a random map with up to 40 units of any side and life,
// views that differ, and a few container records.
func randomPhase(rng *rand.Rand, a int) *PhaseChange {
	g := phaseBase(a)
	g.Maps = phaseMaps(rng.IntN(7))
	g.Phase = [2]int{rng.IntN(3), rng.IntN(3)}
	g.Count = rng.IntN(0x10000)
	cells := phaseW * phaseH
	for range rng.IntN(40) {
		idx := rng.IntN(dosrules.LiveUnits)
		u := Sighted{Flags4: sideFlags[rng.IntN(4)] | rng.IntN(0x40)<<2, Size: rng.IntN(7),
			Pos: [2]int{at(rng.IntN(cells)), at(rng.IntN(cells))}, Facing: [2]int{rng.IntN(6), rng.IntN(6)}}
		switch rng.IntN(8) {
		case 0:
			u.Flags4 |= 0x8000
		case 1:
			u.Flags4 |= 0x4000
		}
		place(g, idx, u)
	}
	for range rng.IntN(4) {
		h := dosrules.EmptyHold()
		h.Flags = []int{0, 1, 2, 0x20, 0x21, 0x8000}[rng.IntN(6)]
		for k := range h.Slots {
			if rng.IntN(2) == 0 {
				h.Slots[k] = rng.IntN(dosrules.LiveUnits)
			}
		}
		g.Holds[rng.IntN(dosrules.HoldRecords)] = h
	}
	return g
}

func TestVisibilityCoverage(t *testing.T) { eachExport(t, visibilityCoverage) }

func needPhase(t *testing.T, e *Export) {
	t.Helper()
	if e.B.Handler == nil || e.B.Move == nil || e.B.Handler.Phase[1] == 0 {
		t.Skipf("the phase change is not located in the %s build", e.B.Name)
	}
}

func visibilityCoverage(t *testing.T, e *Export) {
	needPhase(t, e)
	perClass, seen := map[string]int{}, map[string]int{}
	began := time.Now()
	for _, c := range phaseCases() {
		want, werr := dosrules.ChangePhase(c.g)
		got, gerr := e.RunPhaseChange(c.g)
		if werr != nil || gerr != nil {
			t.Errorf("%s/%s: model %v, emulator %v", c.class, c.name, werr, gerr)
			continue
		}
		if d := fieldDiff(got, want); d != "" {
			t.Errorf("%s/%s: %s", c.class, c.name, d)
			continue
		}
		perClass[c.class]++
		phaseRules(t, c, got, seen)
	}
	for _, class := range []string{"phase", "sides", "life", "stale", "ships", "terrain", "holds", "random"} {
		if perClass[class] == 0 {
			t.Errorf("class %s: no case compared", class)
		}
	}
	t.Logf("compared %v in %v", perClass, time.Since(began).Round(time.Millisecond))
	keys := slices.Collect(maps.Keys(seen))
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("%s: %d", k, seen[k])
	}
	for _, k := range []string{"maps identical", "unit written at its side-0 word", "unit dropped: destroyed",
		"unit dropped: in a container", "unit dropped: dying", "view copied to the attacking side",
		"stale side-0 word used", "slots synced", "record skipped", "destroyed cargo cleared",
		"building marker lost", "terrain of the attacking side's map", "turn counted"} {
		if seen[k] == 0 {
			t.Errorf("no case showed %q", k)
		}
	}
}

// fieldDiff names the first field where two structs of one type differ.
func fieldDiff(got, want any) string {
	gv, wv := reflect.ValueOf(got), reflect.ValueOf(want)
	for i := range gv.NumField() {
		if !reflect.DeepEqual(gv.Field(i).Interface(), wv.Field(i).Interface()) {
			return fmt.Sprintf("%s is %v, want %v", gv.Type().Field(i).Name, gv.Field(i).Interface(), wv.Field(i).Interface())
		}
	}
	return ""
}

// phaseRules checks the rules on an executed outcome and counts what the
// case showed.
func phaseRules(t *testing.T, c phaseCase, r PhaseRun, seen map[string]int) {
	t.Helper()
	g, a, b := c.g, c.g.Attacker, 1-c.g.Attacker
	fail := func(format string, args ...any) {
		t.Errorf("%s/%s: "+format, append([]any{c.class, c.name}, args...)...)
	}
	if !bytes.Equal(r.Maps[0], r.Maps[1]) {
		fail("the maps differ afterwards")
	} else {
		seen["maps identical"]++
	}
	for c := range phaseW * phaseH {
		if r.Maps[0][2*c] == byte(g.Maps[a][2*c]) && g.Maps[a][2*c] != g.Maps[b][2*c] {
			seen["terrain of the attacking side's map"]++
			break
		}
	}
	for idx, u := range g.Units {
		l := r.Live[idx]
		f := word(l, 4)
		other := b == 1 && u.Flags4&1 != 0 || b == 0 && u.Flags4&3 == 0
		switch {
		case u.Flags4&0x8000 != 0:
			seen["unit dropped: destroyed"]++
		case f&0x8000 != 0:
			seen["unit dropped: dying"]++
		case f&0x4000 != 0:
			seen["unit dropped: in a container"]++
		default:
			p := word(l, 0x0B)
			if int(r.Maps[0][p]) < idx {
				fail("unit %d is not at its side-0 word %d", idx, p)
			}
			seen["unit written at its side-0 word"]++
			if !other && u.Pos[0] != u.Pos[a] {
				seen["stale side-0 word used"]++
			}
		}
		if u.Flags4&0x8000 == 0 && other {
			if word(l, 0x0B+2*a) != u.Pos[b] || int(l[0x0F+a]) != u.Facing[b] {
				fail("unit %d of the other side keeps %d in the attacking side's view", idx, word(l, 0x0B+2*a))
			}
			if u.Pos[a] != u.Pos[b] {
				seen["view copied to the attacking side"]++
			}
		}
	}
	for i, h := range g.Holds {
		rb := r.Holds[i]
		if h.Flags&0x8002 != 0 {
			seen["record skipped"]++
			continue
		}
		if !bytes.Equal(rb[:7], rb[7:14]) {
			fail("record %d: the slot lists differ", i)
		}
		seen["slots synced"]++
		for k := range 7 {
			if u := h.Slots[7*b+k]; dosrules.Owns(b, h.Flags) && u <= dosrules.SlotFree && rb[7*b+k] == 0xFF {
				seen["destroyed cargo cleared"]++
			}
		}
	}
	for c := range phaseW * phaseH {
		if g.Maps[a][2*c+1] == dosrules.InBuilding || g.Maps[b][2*c+1] == dosrules.InBuilding {
			if r.Maps[0][2*c+1] != dosrules.InBuilding {
				seen["building marker lost"]++
			}
		}
	}
	if r.Count != g.Count {
		seen["turn counted"]++
	}
}

// TestEngagementHidden runs attack orders while the other side's view
// differs from the attacking side's: the defender left the target cell, a
// unit entered it, and the defender's neighbours moved away. The handler
// reads only the attacking side's map and position words, so the model,
// which knows nothing of Moved, must agree.
func TestEngagementHidden(t *testing.T) { eachExport(t, engagementHidden) }

func engagementHidden(t *testing.T, e *Export) {
	needHandler(t, e)
	far := 0
	for side := range 2 {
		for k, d := range ring(centre) {
			left, right := ring(centre)[(k+5)%6], ring(centre)[(k+1)%6]
			g := duel(e, centre, d)
			g.Side = side
			g.Moved = map[int]int{1: far}
			if o := engage(t, e, fmt.Sprintf("side %d direction %d defender left", side, k), g); !o.Battle {
				t.Errorf("side %d direction %d: the defender that left was not fought", side, k)
			}
			g.Moved = map[int]int{1: ring(centre)[(k+3)%6]}
			engage(t, e, fmt.Sprintf("side %d direction %d defender moved round", side, k), g)

			g = duel(e, centre, d)
			g.Side = side
			g.Units[2], g.Units[3] = fighter(e, infantry, 2, left), fighter(e, infantry, 2, right)
			g.Moved = map[int]int{2: far, 3: far + 1}
			if o := engage(t, e, fmt.Sprintf("side %d direction %d blockers left", side, k), g); o.Blocking != 150 {
				t.Errorf("side %d direction %d: blockers that left count %d", side, k, o.Blocking)
			}

			g = duel(e, centre, d)
			g.Side = side
			g.Target = ring(centre)[(k+3)%6]
			g.Moved = map[int]int{1: g.Target}
			if o := engage(t, e, fmt.Sprintf("side %d direction %d unit entered the target", side, k), g); o.Battle {
				t.Errorf("side %d direction %d: a unit the attacking side does not see was fought", side, k)
			}
		}
	}
}
