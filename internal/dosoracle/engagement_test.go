package dosoracle

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// R1 controlled battles: the order handler (02dc:0004) resolves one attack
// order on a map with the build's own UNIT.DAT records. Its neighbour list,
// support caller, battle-screen caller, casualty routine and experience
// window run as in the game; only drawing, scrolling and sound are hooked.
// Engage composes the independent transcriptions in the handler's order, and
// the two must agree field for field.

// Unit types of the demo's UNIT.DAT used below.
const (
	infantry   = 18 // target flags 0x0c, surface range 2
	lightArt   = 10 // surface range 4
	heavyArt   = 12 // long range: target flags 0x40
	depotBuild = 16 // target flags 0x80: never returns fire
	transport  = 22 // target flags 0: hits nothing
	panzer     = 47
)

const sceneW, sceneH = 9, 9

// centre is the attacker's cell in most scenes; all its rings stay on the map.
var centre = 4*sceneW + 4

func TestDistanceAgrees(t *testing.T) {
	eachExport(t, func(t *testing.T, e *Export) {
		for r1 := range 8 {
			for c1 := range 8 {
				for r2 := range 8 {
					for c2 := range 8 {
						got, err := e.RunDistance(r1, c1, r2, c2)
						if err != nil {
							t.Fatal(err)
						}
						if want := Distance(r1, c1, r2, c2); got != want {
							t.Errorf("(%d,%d)-(%d,%d): emulator %d, model %d", r1, c1, r2, c2, got, want)
						}
					}
				}
			}
		}
	})
}

// fighter is a unit of a type at full group size; side is the live side bit
// (1 or 2) and the class bits are the type's own (+0x0d).
func fighter(e *Export, typ, side, cell int) Fighter {
	r := e.TypeRecord(typ)
	return Fighter{Type: typ, Cell: cell, Size: int(r[3]), Flags4: side | word(r, 0x0D)&0x3C}
}

// duel puts an infantry attacker (side 1) in cell a and an infantry defender
// (side 2) in cell d.
func duel(e *Export, a, d int) *Engagement {
	return &Engagement{
		Width: sceneW, Height: sceneH,
		Units:    map[int]Fighter{0: fighter(e, infantry, 1, a), 1: fighter(e, infantry, 2, d)},
		Attacker: 0, Target: d,
	}
}

func ring(cell int) [6]int { return Ring(cell, sceneW, sceneH) }

// layoutDistance calls the distance helper with (column, row), as the
// movement code at 0b06:117f does; it is 1 for every cell of Ring.
func layoutDistance(a, b int) int {
	return Distance(a%sceneW, a/sceneW, b%sceneW, b/sceneW)
}

// battleDistance is the distance the casualty routine computes: it passes
// (row, column), the transposed order.
func battleDistance(a, b int) int {
	return Distance(a/sceneW, a%sceneW, b/sceneW, b%sceneW)
}

// The casualty routine's transposed call puts one of the six neighbours at
// distance 2, so an adjacent attack in that direction takes the range
// penalty of a weapon with encoded range above 1; farther away the battle
// distance is up to two more or less than the layout distance.
func TestBattleDistanceTransposed(t *testing.T) {
	for _, c := range []int{30, 31, 40, 41} {
		far := 0
		for _, n := range ring(c) {
			if layoutDistance(c, n) != 1 {
				t.Errorf("cell %d: neighbour %d at layout distance %d", c, n, layoutDistance(c, n))
			}
			if battleDistance(c, n) != 1 {
				far++
			}
		}
		if far != 1 {
			t.Errorf("cell %d: %d neighbours not at battle distance 1", c, far)
		}
	}
	spread := map[int]int{}
	for a := range sceneW * sceneH {
		for b := range sceneW * sceneH {
			spread[battleDistance(a, b)-layoutDistance(a, b)]++
		}
	}
	for d := range spread {
		if d < -2 || d > 2 {
			t.Errorf("battle distance differs by %d", d)
		}
	}
}

func indexOf(r [6]int, cell int) int {
	for k, c := range r {
		if c == cell {
			return k
		}
	}
	return -1
}

// needHandler skips a build whose order handler is not located.
func needHandler(t *testing.T, e *Export) {
	t.Helper()
	if e.B.Handler == nil {
		t.Skipf("order handler not located in the %s build", e.B.Name)
	}
}

// engage runs a scene through the handler and the model and compares them.
func engage(t *testing.T, e *Export, what string, g *Engagement) Outcome {
	t.Helper()
	want, werr := Engage(g, e.TypeRecord, e.TerrainTerms())
	got, gerr := e.RunEngagement(g)
	if werr != nil || gerr != nil {
		if !errors.Is(werr, ErrDivideError) || !errors.Is(gerr, ErrDivideError) {
			t.Fatalf("%s: model %v, emulator %v", what, werr, gerr)
		}
		return want
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s:\nemulator %+v\nmodel    %+v", what, got, want)
	}
	return want
}

func TestEngagementDuels(t *testing.T) { eachExport(t, engagementDuels) }

func engagementDuels(t *testing.T, e *Export) {
	needHandler(t, e)
	for side := range 2 {
		for k, d := range ring(centre) {
			g := duel(e, centre, d)
			g.Side = side
			o := engage(t, e, fmt.Sprintf("side %d direction %d", side, k), g)
			if !o.Battle || !o.ReturnFire || o.Pincer != 0 || o.Blocking != 0 {
				t.Errorf("side %d direction %d: adjacent duel %+v", side, k, o)
			}
		}
	}
	// Terrain under either unit.
	d := ring(centre)[3]
	for ca := range 9 {
		for cd := range 9 {
			g := duel(e, centre, d)
			g.Terrain = map[int]int{centre: ca, d: cd}
			engage(t, e, fmt.Sprintf("terrain %d/%d", ca, cd), g)
		}
	}
	// Experience 0-6 on both sides, and the roles swapped.
	for ea := range 7 {
		for ed := range 7 {
			g := duel(e, centre, d)
			u := g.Units[0]
			u.Exp = ea
			g.Units[0] = u
			u = g.Units[1]
			u.Exp = ed
			g.Units[1] = u
			engage(t, e, fmt.Sprintf("experience %d/%d", ea, ed), g)
			g.Attacker, g.Target = 1, centre
			engage(t, e, fmt.Sprintf("experience %d/%d swapped", ea, ed), g)
		}
	}
	// No unit at the target: the handler skips the order.
	g := duel(e, centre, d)
	g.Target = ring(centre)[0]
	if o := engage(t, e, "empty target", g); o.Battle {
		t.Errorf("empty target fought: %+v", o)
	}
}

func TestEngagementBlocking(t *testing.T) { eachExport(t, engagementBlocking) }

// Blocking: units friendly to the defender in the two cells beside the
// defender's direction around the attacker. The manual says each adds a
// quarter; the handler feeds 80 each, capped at 150.
func engagementBlocking(t *testing.T, e *Export) {
	needHandler(t, e)
	for k, d := range ring(centre) {
		left, right := ring(centre)[(k+5)%6], ring(centre)[(k+1)%6]
		for _, c := range []struct {
			name   string
			cells  []int
			typ    int
			side   int
			expect int
		}{
			{"left", []int{left}, infantry, 2, 80},
			{"right", []int{right}, infantry, 2, 80},
			{"both", []int{left, right}, infantry, 2, 150},
			{"attacker's side", []int{left, right}, infantry, 1, 0},
			{"long range", []int{left, right}, heavyArt, 2, 0},
			{"no target class", []int{left, right}, transport, 2, 0},
		} {
			g := duel(e, centre, d)
			for i, cell := range c.cells {
				g.Units[2+i] = fighter(e, c.typ, c.side, cell)
			}
			what := fmt.Sprintf("direction %d blocking %s", k, c.name)
			if o := engage(t, e, what, g); o.Blocking != c.expect {
				t.Errorf("%s: blocking %d, want %d", what, o.Blocking, c.expect)
			}
		}
	}
}

func TestEngagementPincer(t *testing.T) { eachExport(t, engagementPincer) }

// Pincer: units friendly to the attacker around the defender, weighted
// 25/50/100/50/25 by their step from the attacker's direction.
func engagementPincer(t *testing.T, e *Export) {
	needHandler(t, e)
	weights := [6]int{0, 25, 50, 100, 50, 25}
	for k, d := range ring(centre) {
		back := indexOf(ring(d), centre)
		all := 0
		for step := 1; step < 6; step++ {
			cell := ring(d)[(back+step)%6]
			if cell < 0 {
				continue
			}
			all += weights[step]
			for _, s := range []struct {
				name   string
				typ    int
				side   int
				expect int
			}{
				{"infantry", infantry, 1, weights[step]},
				{"defender's side", infantry, 2, 0},
				{"long range", heavyArt, 1, 0},
				{"no target class", transport, 1, 0},
			} {
				g := duel(e, centre, d)
				g.Units[2] = fighter(e, s.typ, s.side, cell)
				what := fmt.Sprintf("direction %d step %d %s", k, step, s.name)
				if o := engage(t, e, what, g); o.Pincer != s.expect {
					t.Errorf("%s: pincer %d, want %d", what, o.Pincer, s.expect)
				}
			}
		}
		g := duel(e, centre, d)
		for step := 1; step < 6; step++ {
			if cell := ring(d)[(back+step)%6]; cell >= 0 {
				g.Units[1+step] = fighter(e, infantry, 1, cell)
			}
		}
		what := fmt.Sprintf("direction %d all five", k)
		if o := engage(t, e, what, g); o.Pincer != all&0xFF {
			t.Errorf("%s: pincer %d, want %d", what, o.Pincer, all&0xFF)
		}
	}
}

func TestEngagementRanged(t *testing.T) { eachExport(t, engagementRanged) }

// Ranged: a light artillery attacker two to four cells away gets neither
// pincer nor blocking, and the defender does not return fire.
func engagementRanged(t *testing.T, e *Export) {
	needHandler(t, e)
	for dist := 2; dist <= 4; dist++ {
		for d := range sceneW * sceneH {
			if layoutDistance(centre, d) != dist {
				continue
			}
			g := duel(e, centre, d)
			g.Units[0] = fighter(e, lightArt, 1, centre)
			// Attacker-side units around the defender and defender-side
			// units around the attacker would support an adjacent battle.
			n := 2
			for _, c := range ring(d) {
				if c >= 0 && c != centre && indexOf(ring(centre), c) < 0 {
					g.Units[n] = fighter(e, infantry, 1, c)
					n++
				}
			}
			for _, c := range ring(centre) {
				if c >= 0 && c != d && indexOf(ring(d), c) < 0 {
					g.Units[n] = fighter(e, infantry, 2, c)
					n++
				}
			}
			what := fmt.Sprintf("distance %d cell %d", dist, d)
			if o := engage(t, e, what, g); o.ReturnFire || o.Pincer != 0 || o.Blocking != 0 {
				t.Errorf("%s: ranged battle %+v", what, o)
			}
		}
	}
	// Adjacent defenders that never return fire: the long-range flag and the
	// special flag.
	for _, typ := range []int{heavyArt, depotBuild} {
		g := duel(e, centre, ring(centre)[2])
		g.Units[1] = fighter(e, typ, 2, ring(centre)[2])
		before := g.Units[0].Size
		if o := engage(t, e, fmt.Sprintf("adjacent type %d", typ), g); o.AttackerSize != before {
			t.Errorf("type %d returned fire: %+v", typ, o)
		}
	}
}

func TestEngagementExperience(t *testing.T) { eachExport(t, engagementExperience) }

// Kill experience: a loss gives the opponent 1, a destruction 1 more, unless
// the loser's live word at +6 has bit 0x20; experience stops at 6.
func engagementExperience(t *testing.T, e *Export) {
	needHandler(t, e)
	d := ring(centre)[1]
	for size := 1; size <= 6; size++ {
		for _, flags6 := range []int{0, noAwardFlag} {
			for _, exp := range []int{0, 5, 6} {
				g := duel(e, centre, d)
				g.Units[0] = fighter(e, panzer, 1, centre)
				u := g.Units[0]
				u.Exp = exp
				g.Units[0] = u
				u = g.Units[1]
				u.Size, u.Flags6 = size, flags6
				g.Units[1] = u
				what := fmt.Sprintf("defender size %d flags6 %#x attacker exp %d", size, flags6, exp)
				o := engage(t, e, what, g)
				gain := 0
				if o.DefenderSize != size && flags6 == 0 {
					gain++
				}
				if o.DefenderSize == 0 {
					gain++
					if !reflect.DeepEqual(o.Destroyed, []int{1}) {
						t.Errorf("%s: destroyed %v", what, o.Destroyed)
					}
				}
				if want := min(6, exp+gain); o.AttackerExp != want {
					t.Errorf("%s: attacker experience %d, want %d", what, o.AttackerExp, want)
				}
			}
		}
	}
}
