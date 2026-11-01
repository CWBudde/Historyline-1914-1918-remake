package dosrules

import (
	"slices"
	"testing"
)

// builderScene is a game on grass with side 1's construction unit 1 at from
// and side 0's infantry 0 at enemy; the records in terrain override the
// grass and holds are the container records. Side 1 is in the action phase.
func builderScene(t *testing.T, p Profile, from, enemy int, terrain map[int]int, holds map[int]Hold) *Game {
	t.Helper()
	tr := make([]int, cycleW*cycleH)
	for c, r := range terrain {
		tr[c] = r
	}
	g, err := NewGame(p, cycleW, cycleH, tr, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: enemy}, 1: {Type: cBuilder, Owner: 1, Cell: from},
	}, holds)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// construct gives construction unit u of side its order on cell, resolves
// the phase and returns the work's events (EvHalfDepot, EvDepot).
func construct(t *testing.T, g *Game, side, u, cell int) []Event {
	t.Helper()
	evs := apply(t, g, Command{Kind: CmdAttack, Side: side, Unit: u, Cell: cell})
	if len(evs) != 1 || evs[0].Kind != EvWork || evs[0].Messages[0].Code != MsgDepot {
		t.Fatalf("order events %+v", evs)
	}
	if l := g.Units[u]; l.Flags4&0x200 == 0 || len(g.Players[side].Orders) != 0 {
		t.Fatalf("after the order: flags %#x, list %v", l.Flags4, g.Players[side].Orders)
	}
	var done []Event
	for _, ev := range resolve(t, g) {
		switch ev.Kind {
		case EvBattle, EvSkipped:
			t.Errorf("the work fought: %+v", ev)
		case EvHalfDepot, EvDepot:
			done = append(done, ev)
		}
	}
	return done
}

// records are the terrain records of cells, -1 where the two maps differ.
func records(g *Game, cells ...int) []int {
	out := make([]int, len(cells))
	for k, c := range cells {
		out[k] = terrainOf(g, c)
	}
	return out
}

func TestBuilderSites(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			// Even column (4, 4): the ten anchors whose depot touches the
			// builder without covering it.
			g := builderScene(t, p, 40, 0, nil, nil)
			want := []int{21, 22, 23, 29, 33, 38, 39, 41, 42, 49}
			if got := slices.Sorted(slices.Values(g.Sites(1, 1))); !slices.Equal(got, want) {
				t.Errorf("sites %v, want %v", got, want)
			}
			for c := range cycleW * cycleH {
				err := g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: c})
				if slices.Contains(want, c) {
					if err != nil {
						t.Errorf("site %d: %v", c, err)
					}
				} else {
					refusal(t, "no site", err, MsgNoTarget)
				}
			}
			for _, a := range want {
				if slices.Contains(g.DepotCells(a), 40) {
					t.Errorf("the depot at %d covers the builder: %v", a, g.DepotCells(a))
				}
			}

			// Odd column (5, 4).
			g = builderScene(t, p, 41, 0, nil, nil)
			want = []int{23, 30, 31, 33, 34, 39, 43, 49, 50, 51}
			if got := slices.Sorted(slices.Values(g.Sites(1, 1))); !slices.Equal(got, want) {
				t.Errorf("odd column: sites %v, want %v", got, want)
			}

			// A unit on a cell or terrain sappers leave alone rules out the
			// depots over it; off the map's edge none is offered.
			g = builderScene(t, p, 40, 50, map[int]int{22: unworked(t, p)}, nil)
			for _, c := range []int{41, 49, 22} {
				refusal(t, "a site over a unit or rock", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: c}), MsgNoTarget)
			}
			g = builderScene(t, p, 0, 80, nil, nil)
			for _, c := range g.Sites(1, 1) {
				for _, d := range g.DepotCells(c) {
					if d < 0 || d >= cycleW*cycleH {
						t.Errorf("site %d reaches off the map: %v", c, g.DepotCells(c))
					}
				}
			}
		})
	}
}

func TestBuilderRefusals(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			rock := unworked(t, p)
			walls := map[int]int{}
			for c := range cycleW * cycleH {
				if c != 40 {
					walls[c] = rock
				}
			}
			g := builderScene(t, p, 40, 0, walls, nil)
			refusal(t, "nowhere to build", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 41}), MsgNoSite)

			g = builderScene(t, p, 40, 0, nil, nil)
			u := g.Units[1]
			u.Hold = 0
			g.Units[1] = u
			refusal(t, "no charges", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 41}), MsgNoCharge)
		})
	}
}

func TestBuilderHalfDepot(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := builderScene(t, p, 40, 0, nil, nil)
			evs := construct(t, g, 1, 1, 41)
			if len(evs) != 1 || evs[0].Kind != EvHalfDepot || evs[0].Cell != 41 || !slices.Equal(evs[0].Cells, []int{41, 51, 50, 49}) {
				t.Fatalf("work %+v", evs)
			}
			if got := records(g, 41, 51, 50, 49); !slices.Equal(got, []int{21, 24, 22, 23}) {
				t.Errorf("records %v, want the half depot", got)
			}
			if g.Units[1].Hold != 1 {
				t.Errorf("charges %d, want 1", g.Units[1].Hold)
			}
			if _, ok := g.Holds[HoldFlag100]; ok {
				t.Error("a half depot has a container record")
			}
			// The half depot is the anchor's site, whatever stands on it.
			if s := g.Sites(1, 1); !slices.Contains(s, 41) {
				t.Errorf("sites %v lack the half depot", s)
			}
		})
	}
}

func TestBuilderCompletesDepot(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := builderScene(t, p, 40, 0, nil, nil)
			construct(t, g, 1, 1, 41)
			resolve(t, g) // side 0's action phase
			evs := construct(t, g, 1, 1, 41)
			if len(evs) != 1 || evs[0].Kind != EvDepot || evs[0].Hold != HoldFlag100 || !slices.Equal(evs[0].Cells, []int{41, 51, 50, 49}) {
				t.Fatalf("work %+v", evs)
			}
			if got := records(g, 41, 51, 50, 49); !slices.Equal(got, []int{17, 20, 18, 19}) {
				t.Errorf("records %v, want side 1's depot", got)
			}
			h, ok := g.Holds[HoldFlag100]
			if !ok {
				t.Fatal("no container record")
			}
			if h.Cells != [4]int{82, 98, 100, 102} || h.Flags != 1|KindDepot || h.Occupant != 0xFF {
				t.Errorf("record %+v", h)
			}
			for _, s := range h.Slots {
				if s <= SlotFree {
					t.Errorf("slots %v", h.Slots)
				}
			}
			if b, ok := g.Buildings[HoldFlag100]; !ok || b.Credits != 0 {
				t.Errorf("building %+v, %v: want one that credits nothing", b, ok)
			}
			if r, ok := g.HoldAt(1, 41); !ok || r != HoldFlag100 {
				t.Errorf("HoldAt: %d, %v", r, ok)
			}
			if g.Units[1].Hold != 0 {
				t.Errorf("charges %d, want 0", g.Units[1].Hold)
			}
			resolve(t, g)
			refusal(t, "no charges left", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 33}), MsgNoCharge)
		})
	}
}

func TestBuilderCompletesOthersHalf(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			// Side 1 lays half a depot; side 0's builder completes it and owns it.
			tr := make([]int, cycleW*cycleH)
			g, err := NewGame(p, cycleW, cycleH, tr, map[int]Placement{
				0: {Type: cBuilder, Owner: 0, Cell: 58}, 1: {Type: cBuilder, Owner: 1, Cell: 40},
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
			construct(t, g, 1, 1, 41)
			evs := construct(t, g, 0, 0, 41)
			if len(evs) != 1 || evs[0].Kind != EvDepot {
				t.Fatalf("work %+v", evs)
			}
			if got := records(g, 41); got[0] != 16 {
				t.Errorf("anchor record %d, want side 0's 16", got[0])
			}
			if h := g.Holds[HoldFlag100]; h.Flags != 0|KindDepot {
				t.Errorf("record flags %#x", h.Flags)
			}
		})
	}
}

func TestBuilderChecksOtherMap(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			// The first half counts the depot's cells on the other side's
			// map: side 0 moves onto one in the same cycle and nothing is
			// built, nor a charge spent.
			g := builderScene(t, p, 40, 59, nil, nil)
			apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 41})
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: 50})
			if g.occupant(0, 50) != 0 {
				t.Fatalf("the move did not arrive: occupant %#x", g.occupant(0, 50))
			}
			for _, ev := range resolve(t, g) {
				if ev.Kind == EvHalfDepot || ev.Kind == EvDepot {
					t.Errorf("built: %+v", ev)
				}
			}
			if got := records(g, 41, 51, 50, 49); !slices.Equal(got, []int{0, 0, 0, 0}) {
				t.Errorf("records %v", got)
			}
			if g.Units[1].Hold != 2 {
				t.Errorf("charges %d, want 2", g.Units[1].Hold)
			}
		})
	}
}

func TestBuilderDepotLimit(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			holds := map[int]Hold{}
			for k := range holdTable {
				h := EmptyHold()
				h.Cells[0] = 2 * (k + 1)
				holds[HoldFlag100+k] = h
			}
			g := builderScene(t, p, 40, 0, map[int]int{41: 21, 51: 24, 50: 22, 49: 23}, holds)
			refusal(t, "the eleventh depot", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 41}), MsgDepotLimit)
			if err := g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 33}); err != nil {
				t.Errorf("a first half beyond the limit: %v", err)
			}
		})
	}
}

func TestBuilderSaveResume(t *testing.T) {
	g := builderScene(t, Full, 40, 0, nil, nil)
	construct(t, g, 1, 1, 41)
	resolve(t, g)
	construct(t, g, 1, 1, 41)
	b, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	r, err := LoadGame(b)
	if err != nil {
		t.Fatal(err)
	}
	if r.Holds[HoldFlag100] != g.Holds[HoldFlag100] || r.Units[1].Hold != 0 {
		t.Errorf("resumed %+v / %+v", r.Holds[HoldFlag100], r.Units[1])
	}
	if _, ok := r.Buildings[HoldFlag100]; !ok {
		t.Error("the resumed depot is no building")
	}
}
