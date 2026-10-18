package dosoracle

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// R1 coverage: each input class of the casualty, support and experience
// routines is enumerated, not sampled, and the executed routine must agree
// with the transcription on every case. The branch trace shows that the
// cases reach every path of the casualty routine both ways.

// battleCase is one enumerated battle with the input class it belongs to.
type battleCase struct {
	class string
	a, d  Side
	p     Params
}

// coverRecord is a synthetic type record: armour, encoded air/surface range,
// air/ground/water power, target flags (+6), attacker classes (+0x0d) and the
// alternate-strength bit (+0x0f).
func coverRecord(armour, airRange, surfRange, air, ground, water, flags6 int, alt bool) []byte {
	r := make([]byte, recordLen)
	r[2], r[8], r[9], r[10], r[11], r[12] = byte(armour), byte(airRange), byte(surfRange), byte(air), byte(ground), byte(water)
	r[6], r[7] = byte(flags6), byte(flags6>>8)
	r[0x0D] = 0x3C
	if alt {
		r[0x0F] = 4
	}
	return r
}

func battleCases(e *Export) []battleCase {
	terms := e.TerrainTerms()
	base := func() (Side, Side, Params) {
		a := Side{Rec: coverRecord(20, 1, 1, 15, 30, 25, 0x04, false), Size: 6}
		d := Side{Rec: coverRecord(20, 1, 1, 15, 30, 25, 0x04, false), Size: 6}
		return a, d, Params{Distance: 1, ATerrain: 7, DTerrain: 7, ReturnFire: true, TerrainTerms: terms}
	}
	var cs []battleCase
	add := func(class string, a, d Side, p Params) { cs = append(cs, battleCase{class, a, d, p}) }

	sizes := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 255}
	for _, sa := range sizes {
		for _, sd := range sizes {
			a, d, p := base()
			a.Size, d.Size = sa, sd
			add("group size", a, d, p)
		}
	}
	// Experience 0-6 is the game's range; 7 and 8 reach the parameter clamps,
	// 127/128/255 the byte width.
	exps := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 127, 128, 255}
	for _, ea := range exps {
		for _, ed := range exps {
			a, d, p := base()
			a.Exp, d.Exp = ea, ed
			add("experience", a, d, p)
		}
	}
	flags := []int{0, 0x08, 0x10, 0x18}
	for _, fa := range flags {
		for _, fd := range flags {
			a, d, p := base()
			a.Flg, d.Flg = fa, fd
			add("target domain", a, d, p)
		}
	}
	maxRange := 0
	for i := range len(e.Units) / recordLen {
		r := e.TypeRecord(i)
		maxRange = max(maxRange, int(r[8]), int(r[9]))
	}
	for _, r := range []int{0, 1, 2, 3, maxRange, 255} {
		for _, flg := range []int{0, 0x10} {
			for dist := 1; dist <= min(r, 12)+1; dist++ {
				a, d, p := base()
				a.Rec = coverRecord(20, r, r, 40, 40, 40, 0x04, false)
				d.Flg, p.Distance = flg, dist
				add("encoded range", a, d, p)
			}
		}
	}
	for ta := range 9 {
		for td := range 9 {
			a, d, p := base()
			p.ATerrain, p.DTerrain = ta, td
			add("terrain class", a, d, p)
		}
	}
	for pin := 0; pin <= 250; pin += 25 {
		for _, blk := range []int{0, 80, 150, 160, 255} {
			a, d, p := base()
			p.Pincer, p.Blocking = pin, blk
			add("pincer and blocking", a, d, p)
		}
	}
	// Defender flags: plain, special 0x80, long range 0x40, a class outside
	// the attacker's mask, and both special and long range.
	for _, rf := range []bool{true, false} {
		for _, f6 := range []int{0x04, 0x84, 0x44, 0x02, 0xC4} {
			a, d, p := base()
			a.Rec[0x0D] = 0x04
			d.Rec = coverRecord(20, 1, 1, 15, 30, 25, f6, false)
			p.ReturnFire = rf
			add("return fire", a, d, p)
		}
	}
	for _, altA := range []bool{false, true} {
		for _, altD := range []bool{false, true} {
			for alt := range 7 {
				for _, size := range []int{0, 6} {
					a, d, p := base()
					a.Rec = coverRecord(20, 1, 1, 15, 120, 25, 0x04, altA)
					d.Rec = coverRecord(5, 1, 1, 15, 120, 25, 0x04, altD)
					a.Alt, d.Alt, a.Size, d.Size = alt, alt, size, size
					add("alternate strength", a, d, p)
				}
			}
		}
	}
	// Casualty arithmetic: minimum loss, 6..12 -> 6, twice-size cap, clamp to
	// zero, size cap, destruction, and zero armour (a divide error on both).
	powers := []int{0, 1, 3, 10, 30, 60, 120, 200, 255}
	armours := []int{0, 1, 2, 5, 20, 60, 255}
	for _, pa := range powers {
		for _, aa := range armours {
			for _, pd := range []int{0, 30, 255} {
				for _, ad := range []int{1, 20, 255} {
					for _, sz := range [][2]int{{1, 6}, {6, 1}, {6, 6}} {
						a, d, p := base()
						a.Rec = coverRecord(aa, 1, 1, pa, pa, pa, 0x04, false)
						d.Rec = coverRecord(ad, 1, 1, pd, pd, pd, 0x04, false)
						a.Size, d.Size = sz[0], sz[1]
						add("casualty arithmetic", a, d, p)
					}
				}
			}
		}
	}
	// Every record of the build against every other, adjacent.
	n := len(e.Units) / recordLen
	for i := range n {
		for j := range n {
			a, d, p := base()
			a.Rec, d.Rec = e.TypeRecord(i), e.TypeRecord(j)
			d.Flg = []int{0, 0x08, 0x10}[(i+j)%3]
			add("type records", a, d, p)
		}
	}
	return cs
}

func TestBattleCoverage(t *testing.T) { eachExport(t, battleCoverage) }

func battleCoverage(t *testing.T, e *Export) {
	cases := battleCases(e)
	var seen, unseen Branches
	perClass := map[string]int{}
	divide := 0
	for i, c := range cases {
		want, br, werr := Trace(c.a, c.d, c.p)
		got, gerr := e.Battle(c.a, c.d, c.p)
		if errors.Is(werr, ErrDivideError) != errors.Is(gerr, dos8086.ErrDivideError) || (werr == nil) != (gerr == nil) {
			t.Fatalf("%s case %d: emulator error %v, model error %v", c.class, i, gerr, werr)
		}
		perClass[c.class]++
		f := casualtyFixture(e, c.class, c)
		ra, rd, rc, err := f.Inputs(profileOf(e).Tables)
		if err != nil {
			t.Fatal(err)
		}
		res, rerr := profileOf(e).Casualties(ra, rd, rc)
		if !errors.Is(rerr, dosrules.ErrDivideError) != (gerr == nil) || (gerr == nil && res != dosrules.Result(got.Result)) {
			t.Errorf("%s case %d: emulator %+v %v, dosrules %+v %v", c.class, i, got.Result, gerr, res, rerr)
		}
		if gerr != nil {
			divide++
			continue
		}
		if got.Result != want {
			t.Errorf("%s case %d: emulator %+v, model %+v (a %+v, d %+v, %+v)", c.class, i, got.Result, want, c.a, c.d, c.p)
		}
		merge(&seen, &unseen, br)
	}
	for _, name := range missing(seen, unseen) {
		if !deadBranches[name] {
			t.Errorf("branch %s is not reached both ways", name)
		}
	}
	for name := range deadBranches {
		if reflect.ValueOf(seen).FieldByName(name).Bool() {
			t.Errorf("dead branch %s was reached", name)
		}
	}
	if divide == 0 {
		t.Error("no case reaches the zero-armour divide error")
	}
	t.Logf("%d cases (%v), %d divide errors on both sides", len(cases), perClass, divide)
}

// deadBranches cannot fire from byte inputs: experience is read unsigned, so
// E = exp+1 is at least 1 (experience 255 gives E = 256 in both).
var deadBranches = map[string]bool{"ExpBelow": true}

// merge ORs a trace into seen and its negation into unseen.
func merge(seen, unseen *Branches, b Branches) {
	vs, vu, vb := reflect.ValueOf(seen).Elem(), reflect.ValueOf(unseen).Elem(), reflect.ValueOf(b)
	for i := range vb.NumField() {
		if vb.Field(i).Bool() {
			vs.Field(i).SetBool(true)
		} else {
			vu.Field(i).SetBool(true)
		}
	}
}

func missing(seen, unseen Branches) (names []string) {
	vs, vu := reflect.ValueOf(seen), reflect.ValueOf(unseen)
	for i := range vs.NumField() {
		if !vs.Field(i).Bool() || !vu.Field(i).Bool() {
			names = append(names, vs.Type().Field(i).Name)
		}
	}
	return names
}

// supportMap puts the defender (unit 1) at a cell of the given column parity
// in the middle of a 7x7 map and the attacker (unit 0) in direction dir of it.
func supportMap(parity, dir int) (*Map, int, int) {
	m := emptyMap(7, 7)
	dcell := 3*7 + 2 + parity
	acell := Ring(dcell, 7, 7)[dir]
	m.Types = map[int]int{0: 0x04, 1: 0x04}
	m.Units = map[int]Unit{0: {Flags4: 1 | 0x04, Type: 0, Cell: acell}, 1: {Flags4: 2 | 0x04, Type: 1, Cell: dcell}}
	m.Cells[dcell], m.Cells[acell] = 1, 0
	return m, acell, dcell
}

// supporter is a third unit's eligibility variant.
type supporter struct {
	name           string
	side, flags6   int
	typeFlags      int
	pincer, blocks bool // eligible around the defender (for the attacker) or around the attacker (for the defender)
}

var supporters = []supporter{
	{"attacker's side, shared class", 1, 0, 0x04, true, false},
	{"defender's side, shared class", 2, 0, 0x04, false, true},
	{"neutral side", 0, 0, 0x04, false, true}, // side 0 pairs with any non-side-1 unit
	{"no shared class", 1, 0, 0x08, false, false},
	{"bit 2 path", 1, 2, 0x02, false, false}, // bit 2 counts only for the scanning anchor
	{"long range", 1, 0, 0x44, false, false},
}

func TestSupportCoverage(t *testing.T) { eachExport(t, supportCoverage) }

func supportCoverage(t *testing.T, e *Export) {
	compare := func(what string, m *Map) (int, int) {
		t.Helper()
		wp, wb, err := Support(m, 0, 1)
		if err != nil {
			t.Fatalf("%s: model %v", what, err)
		}
		gp, gb, err := e.RunSupport(m, 0, 1, 0)
		if err != nil {
			t.Fatalf("%s: emulator %v", what, err)
		}
		if gp != wp || gb != wb {
			t.Errorf("%s: emulator %d/%d, model %d/%d", what, gp, gb, wp, wb)
		}
		return wp, wb
	}
	pincerByRel := map[int]map[int]bool{}
	blockSeen := map[int]bool{}
	cases := 0
	for parity := range 2 {
		for dir := range 6 {
			// Nothing around: zero.
			m, acell, dcell := supportMap(parity, dir)
			if p, b := compare("alone", m); p != 0 || b != 0 {
				t.Errorf("alone: %d/%d", p, b)
			}
			cases++
			// One third unit in each free cell around the defender and the
			// attacker, for every eligibility variant.
			for _, anchor := range []int{dcell, acell} {
				for k, c := range Ring(anchor, 7, 7) {
					if c < 0 || c == acell || c == dcell {
						continue
					}
					for _, s := range supporters {
						m, _, _ := supportMap(parity, dir)
						m.Types[2] = s.typeFlags
						m.Units[2] = Unit{Flags4: s.side | 0x04, Flags6: s.flags6, Type: 2, Cell: c}
						m.Cells[c] = 2
						what := fmt.Sprintf("parity %d dir %d anchor %d slot %d %s", parity, dir, anchor, k, s.name)
						p, b := compare(what, m)
						cases++
						if anchor == dcell && p != 0 {
							rel := (k - (dir) + 6) % 6
							if pincerByRel[rel] == nil {
								pincerByRel[rel] = map[int]bool{}
							}
							pincerByRel[rel][p] = true
						}
						if !s.pincer && anchor == dcell && p != 0 {
							t.Errorf("%s: ineligible supporter gave pincer %d", what, p)
						}
						if !s.blocks && b != 0 && anchor == acell {
							t.Errorf("%s: ineligible blocker gave %d", what, b)
						}
						blockSeen[b] = true
					}
				}
			}
			// Full rings: the attacker's side around the defender (all five
			// pincer slots), then the defender's side around the attacker
			// (both blockers; they stand on the cells the two rings share).
			for _, ring := range []struct{ side, typ int }{{1, 0}, {2, 1}} {
				m, acell, dcell = supportMap(parity, dir)
				anchor, other := dcell, acell
				if ring.side == 2 {
					anchor, other = acell, dcell
				}
				idx := 2
				for _, c := range Ring(anchor, 7, 7) {
					if c >= 0 && c != other {
						m.Units[idx] = Unit{Flags4: ring.side | 0x04, Type: ring.typ, Cell: c}
						m.Cells[c] = idx
						idx++
					}
				}
				p, b := compare(fmt.Sprintf("parity %d dir %d full ring of side %d", parity, dir, ring.side), m)
				if ring.side == 1 && p != 250 {
					t.Errorf("parity %d dir %d: full pincer ring gave %d", parity, dir, p)
				}
				blockSeen[b] = true
				cases++
			}
		}
	}
	for rel, want := range map[int]int{1: 25, 2: 50, 3: 100, 4: 50, 5: 25} {
		if !pincerByRel[rel][want] {
			t.Errorf("relative direction %d never gave pincer %d (saw %v)", rel, want, pincerByRel[rel])
		}
	}
	for _, b := range []int{0, 80, 150} {
		if !blockSeen[b] {
			t.Errorf("blocking %d never reached (saw %v)", b, blockSeen)
		}
	}
	t.Logf("%d support cases", cases)
}

func TestExperienceCoverage(t *testing.T) { eachExport(t, experienceCoverage) }

func experienceCoverage(t *testing.T, e *Export) {
	outcomes := [][2]int{{5, 0}, {5, 3}, {5, 5}, {1, 0}, {0, 0}} // before, after: kill, loss, none, last man, empty
	cases, kills, capped := 0, 0, 0
	for ea := range 7 {
		for ed := range 7 {
			for _, oa := range outcomes {
				for _, od := range outcomes {
					for _, fl := range [][2]int{{0, 0}, {0x20, 0}, {0, 0x20}, {0x20, 0x20}} {
						for _, alt := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
							a := ExpSide{ea, oa[0], oa[1], fl[0]}
							d := ExpSide{ed, od[0], od[1], fl[1]}
							ga, gd, gk, err := e.RunExperienceWindow(a, d, alt)
							if err != nil {
								t.Fatal(err)
							}
							wa, wd, wk := Award(a, d)
							if ga != wa || gd != wd || gk != len(wk) {
								t.Fatalf("%+v %+v %v: emulator %d %d %d, model %d %d %d", a, d, alt, ga, gd, gk, wa, wd, len(wk))
							}
							cases++
							kills += len(wk)
							if (ea == expCap && wa == expCap && (od[1] == 0 || od[0] != od[1])) || (ed == expCap && wd == expCap) {
								capped++
							}
						}
					}
				}
			}
		}
	}
	if kills == 0 || capped == 0 {
		t.Fatalf("kills %d, capped %d", kills, capped)
	}
	t.Logf("%d experience cases, %d kills", cases, kills)
}
