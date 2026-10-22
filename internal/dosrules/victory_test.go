package dosrules

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"testing"
)

// The end of a battle: the HQ held, annihilation, a draw, and the score.

const hq0, hq1 = 10, 70 // the HQ entrance cells of side 0 and side 1

// hqCycle is newCycle with side 0's HQ entrance at hq0 and side 1's at hq1.
func hqCycle(t *testing.T, p Profile, units map[int]Placement) *Game {
	t.Helper()
	return hqCycleHolds(t, p, units, nil)
}

// hqCycleHolds is hqCycle with further container records; the HQ records
// are given where holds has none.
func hqCycleHolds(t *testing.T, p Profile, units map[int]Placement, holds map[int]Hold) *Game {
	t.Helper()
	m := moveTables(t, p)
	terrain := make([]int, cycleW*cycleH)
	for i := range terrain {
		terrain[i] = openTerrain(t, p)
	}
	for i, tr := range m.Terrains {
		if tr.Flags&0x40 != 0 {
			terrain[[2]int{hq0, hq1}[tr.Flags&1]] = i
		}
	}
	if m.Terrains[terrain[hq0]].Flags&0x40 == 0 || m.Terrains[terrain[hq1]].Flags&0x40 == 0 {
		t.Fatal("no HQ terrain")
	}
	h0, h1 := EmptyHold(), EmptyHold()
	h0.Cells[0] = 2 * hq0
	h1.Flags, h1.Cells[0] = 1, 2*hq1
	all := map[int]Hold{HoldFlag40: h0, HoldFlag40 + 1: h1}
	maps.Copy(all, holds)
	g, err := NewGame(p, cycleW, cycleH, terrain, units, all)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// armour is the sum of the armour of the units not destroyed.
func armour(g *Game) int {
	s := 0
	for _, u := range g.Units {
		if !u.Destroyed() {
			s += g.Profile.Tables.Types[u.Type].Armour
		}
	}
	return s
}

func endOf(t *testing.T, evs []Event) *End {
	t.Helper()
	e := evs[len(evs)-1]
	if e.Kind != EvEnd || e.End == nil {
		t.Fatalf("no end event: %+v", evs)
	}
	if e.Side != e.End.Winner {
		t.Errorf("end event side %d, winner %d", e.Side, e.End.Winner)
	}
	return e.End
}

func TestVictoryOpen(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := hqCycleHolds(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: hq0}, 1: {Type: cPanzer, Owner: 1, Cell: hq1},
				2: {Type: cSupply, Owner: 2, Cell: 40},
			}, map[int]Hold{0: transportHold(40, 2, 2)})
			if r := g.end(-1); r.Over || r.Winner != -1 || r.Score != 0 || g.Ended != nil {
				t.Errorf("result %+v", r)
			}
			want := 0
			for _, ty := range []int{cInfantry, cPanzer, cSupply} {
				want += p.Tables.Types[ty].Armour
			}
			if g.Score() != want || armour(g) != want {
				t.Errorf("score %d, want %d", g.Score(), want)
			}
			evs := resolve(t, g)
			if evs[len(evs)-1].Kind != EvPhase {
				t.Errorf("an open battle ends: %+v", evs)
			}
		})
	}
}

func TestVictoryAnnihilation(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			for size := 1; size <= 6; size++ {
				g := hqCycle(t, p, map[int]Placement{
					0: {Type: cInfantry, Owner: 0, Cell: 40, Size: size},
					1: {Type: cPanzer, Owner: 1, Cell: n[0]}, 2: {Type: cPanzer, Owner: 1, Cell: n[3]},
				})
				apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40})
				apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 2, Cell: 40})
				evs := resolve(t, g)
				if !slices.Equal(g.Destroyed(), []int{0}) {
					if evs[len(evs)-1].Kind == EvEnd {
						t.Fatalf("size %d: ends with side 0 alive: %+v", size, evs)
					}
					continue
				}
				r := endOf(t, evs)
				if !r.Over || r.Winner != 1 || r.Won != [2]EndReason{"", EndAnnihilation} {
					t.Errorf("result %+v", r)
				}
				if want := armour(g) + 500; r.Score != want || g.Score() != armour(g) {
					t.Errorf("score %d (now %d), want %d", r.Score, g.Score(), want)
				}
				if g.Ended != r {
					t.Errorf("game end %+v, event %+v", g.Ended, r)
				}
				if h := reload(t, g); !reflect.DeepEqual(g, h) {
					t.Errorf("an ended battle reloads as %+v", h.Ended)
				} else {
					g = h
				}
				var ref *Refusal
				if err := g.Check(Command{Kind: CmdMove, Side: 1, Unit: 1, Cell: 0}); !errors.As(err, &ref) {
					t.Errorf("a command after the end: %v", err)
				}
				g.Confirm(0)
				g.Confirm(1)
				if _, err := g.Resolve(); !errors.Is(err, ErrOver) {
					t.Errorf("a resolution after the end: %v", err)
				}
				return
			}
			t.Error("no attack destroyed side 0's unit")
		})
	}
}

// TestVictoryHQ captures the opposing HQ: the final score counts the
// concealed inventory and the move limit.
func TestVictoryHQ(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(hq1, cycleW, cycleH)
			g := hqCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: n[0]}, 1: {Type: cPanzer, Owner: 1, Cell: 40},
			})
			g.Limit, g.Concealed = 8, true
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: hq1})
			r := endOf(t, resolve(t, g))
			if !r.Over || r.Winner != 0 || r.Won != [2]EndReason{EndHQ, ""} {
				t.Errorf("result %+v", r)
			}
			if want := (armour(g) + 100) * 3; r.Score != want {
				t.Errorf("score %d, want %d", r.Score, want)
			}
		})
	}
}

// TestVictoryHQStanding: a unit placed on the opposing HQ has not entered
// it and does not win.
func TestVictoryHQStanding(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := hqCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: hq1}, 1: {Type: cPanzer, Owner: 1, Cell: 40},
			})
			if evs := resolve(t, g); evs[len(evs)-1].Kind == EvEnd || g.Ended != nil {
				t.Errorf("the battle ends: %+v", g.Ended)
			}
		})
	}
}

// TestVictoryHQMove reaches the opposing HQ by a move order.
func TestVictoryHQMove(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(hq1, cycleW, cycleH)
			g := hqCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: n[0]}, 1: {Type: cPanzer, Owner: 1, Cell: 40},
			})
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: hq1})
			if g.Ended != nil {
				t.Error("the battle ends before the resolution")
			}
			r := endOf(t, resolve(t, g))
			if r.Winner != 0 || r.Won[0] != EndHQ {
				t.Errorf("result %+v", r)
			}
		})
	}
}

func TestVictoryOwnHQ(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := hqCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: hq0}, 1: {Type: cPanzer, Owner: 1, Cell: hq1},
			})
			if evs := resolve(t, g); evs[len(evs)-1].Kind == EvEnd || g.Ended != nil {
				t.Errorf("a side wins on its own HQ: %+v", g.Ended)
			}
		})
	}
}

// TestVictoryHQOverAnnihilation: the HQ captured wins as such, without the
// annihilation bonus, even where the other side has no unit left.
func TestVictoryHQOverAnnihilation(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(hq1, cycleW, cycleH)
			g := hqCycle(t, p, map[int]Placement{0: {Type: cInfantry, Owner: 0, Cell: n[0]}})
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: hq1})
			r := endOf(t, resolve(t, g))
			if !r.Over || r.Winner != 0 || r.Won != [2]EndReason{EndHQ, ""} || r.Score != armour(g) {
				t.Errorf("result %+v, want the HQ won at score %d", r, armour(g))
			}
		})
	}
}

func TestFinalScore(t *testing.T) {
	for _, c := range []struct {
		score        int
		annihilation bool
		concealed    bool
		limit, want  int
	}{
		{1000, false, false, 0, 1000},
		{1000, true, false, 0, 1500},
		{1000, false, true, 0, 1100},
		{1000, true, true, 4, 6400},
		{1000, false, false, 8, 3000},
		{1000, false, false, 16, 2000},
		{1000, false, false, 5, 1000},
		{20000, false, false, 4, 32500},
		{32600, false, false, 0, 32500},
	} {
		if got := finalScore(c.score, c.annihilation, c.concealed, c.limit); got != c.want {
			t.Errorf("finalScore(%d, %v, %v, %d) = %d, want %d", c.score, c.annihilation, c.concealed, c.limit, got, c.want)
		}
	}
}
