package dosoracle

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

// R3 resolution timing: the order handler (02dc:0004) runs the attacking
// side's whole list at the phase change. RunOrders executes the loop with the
// destroy helper in place, and ResolveOrders, which chains the engagement
// model order by order, must agree on every order and on the state left.

func resolveOrders(t *testing.T, e *Export, what string, g *OrderList) (Resolution, bool) {
	t.Helper()
	want, werr := ResolveOrders(g, e.TypeRecord, e.TerrainTerms())
	got, gerr := e.RunOrders(g)
	if werr != nil || gerr != nil {
		if !errors.Is(werr, ErrDivideError) || !errors.Is(gerr, ErrDivideError) {
			t.Fatalf("%s: model %v, emulator %v", what, werr, gerr)
		}
		return want, false
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: %s\nlist %+v\nunits %+v moved %v", what, diffResolution(got, want), g.Orders, g.Units, g.Moved)
	}
	return want, true
}

// diffResolution names the first fields where the emulator's resolution
// (got) and the model's (want) differ.
func diffResolution(got, want Resolution) string {
	for i := range max(len(got.Orders), len(want.Orders)) {
		if i >= len(got.Orders) || i >= len(want.Orders) {
			return fmt.Sprintf("%d orders run, model %d", len(got.Orders), len(want.Orders))
		}
		if !reflect.DeepEqual(got.Orders[i], want.Orders[i]) {
			return fmt.Sprintf("order %d:\nemulator %+v\nmodel    %+v", i, got.Orders[i], want.Orders[i])
		}
	}
	for i, u := range got.Units {
		if u != want.Units[i] {
			return fmt.Sprintf("unit %d: emulator %+v, model %+v", i, u, want.Units[i])
		}
	}
	for c := range got.Cells {
		if got.Cells[c] != want.Cells[c] {
			return fmt.Sprintf("cell %d: emulator %#x, model %#x", c, got.Cells[c], want.Cells[c])
		}
	}
	return fmt.Sprintf("marked: emulator %v, model %v", got.Marked, want.Marked)
}

func TestOrdersCoverage(t *testing.T) { eachExport(t, ordersCoverage) }

func ordersCoverage(t *testing.T, e *Export) {
	needHandler(t, e)
	seen := map[string]int{}
	run := func(what string, g *OrderList) (Resolution, bool) {
		r, ok := resolveOrders(t, e, what, g)
		if !ok {
			seen["divide error"]++
			return r, false
		}
		gone := map[int]bool{}
		for i, o := range r.Orders {
			if gone[g.Orders[i].Attacker] {
				t.Errorf("%s: order %d by a destroyed attacker", what, i)
			}
			if !o.Battle {
				seen["order skipped"]++
				for _, p := range g.Orders[:i] {
					if p.Target == g.Orders[i].Target {
						seen["target destroyed before its order"]++
						break
					}
				}
				continue
			}
			if gone[o.Defender] {
				seen["destroyed unit fought again"]++
			}
			if i > 0 && r.Orders[i-1].Battle && r.Orders[i-1].Defender == o.Defender {
				seen["surviving target fought again"]++
			}
			for _, k := range o.Destroyed {
				gone[k] = true
				if k == g.Orders[i].Attacker && i+1 < len(r.Orders) {
					seen["attacker destroyed before a later order"]++
				}
			}
		}
		if len(g.Idle) > 0 {
			seen["other side's list present"]++
		}
		if len(g.Moved) > 0 {
			seen["other side moved"]++
		}
		return r, true
	}

	// Two attackers on one target, every size of the target and of the
	// first attacker, from opposite and from adjacent neighbours; each list
	// in both orders.
	n := ring(centre)
	for side := range 2 {
		for _, pair := range [][2]int{{0, 3}, {0, 1}} {
			for _, typ := range []int{infantry, panzer} {
				for ds := 1; ds <= 6; ds++ {
					for as := 1; as <= 6; as++ {
						a1, a2 := n[pair[0]], n[pair[1]]
						units := map[int]Fighter{
							0: fighter(e, typ, 1, a1), 1: fighter(e, infantry, 1, a2),
							2: fighter(e, infantry, 2, centre),
						}
						u := units[2]
						u.Size = ds
						units[2] = u
						u = units[0]
						u.Size = min(as, u.Size)
						units[0] = u
						g := &OrderList{Engagement: Engagement{Width: sceneW, Height: sceneH, Units: units, Side: side},
							Orders: []Order{{0, centre}, {1, centre}}}
						what := fmt.Sprintf("side %d pair %v type %d sizes %d/%d", side, pair, typ, as, ds)
						r1, ok1 := run(what, g)
						rev := *g
						rev.Orders = []Order{{1, centre}, {0, centre}}
						r2, ok2 := run(what+" reversed", &rev)
						if ok1 && ok2 && !reflect.DeepEqual(r1.Units, r2.Units) {
							seen["list order changes the result"]++
						}
					}
				}
			}
		}
	}

	// A weak attacker that dies to return fire, then a second attacker on
	// the same target that it would have supported: pincer and blocking are
	// read from the attacking side's map after the first battle.
	for side := range 2 {
		for k := range 6 {
			for as := 1; as <= 3; as++ {
				units := map[int]Fighter{
					0: fighter(e, infantry, 1, n[k]), 1: fighter(e, panzer, 1, n[(k+3)%6]),
					2: fighter(e, panzer, 2, centre),
				}
				u := units[0]
				u.Size = as
				units[0] = u
				g := &OrderList{Engagement: Engagement{Width: sceneW, Height: sceneH, Units: units, Side: side},
					Orders: []Order{{0, centre}, {1, centre}}, Idle: []int{2}}
				run(fmt.Sprintf("side %d supporter %d size %d", side, k, as), g)
			}
		}
	}

	// Random lists: two to five attackers of mixed types, one to four
	// defenders, targets that may be empty, the other side's view moved, and
	// a list for the other side. The attack order takes only cells with the
	// side's target bit (0ad1:000c), so no order targets an own unit and no
	// attacker can be destroyed before its own order.
	types := []int{infantry, lightArt, heavyArt, depotBuild, transport, panzer}
	rng := rand.New(rand.NewPCG(3, uint64(e.B.SHA256[0])))
	for i := range 300 {
		units := map[int]Fighter{}
		used := map[int]bool{}
		cell := func() int {
			for {
				c := rng.IntN(sceneW * sceneH)
				if !used[c] {
					used[c] = true
					return c
				}
			}
		}
		na, nd := 2+rng.IntN(4), 1+rng.IntN(4)
		for k := range na + nd {
			side := 1
			if k >= na {
				side = 2
			}
			u := fighter(e, types[rng.IntN(len(types))], side, cell())
			u.Size = 1 + rng.IntN(u.Size)
			u.Exp = rng.IntN(7)
			units[k] = u
		}
		g := &OrderList{Engagement: Engagement{Width: sceneW, Height: sceneH, Units: units, Side: rng.IntN(2)}}
		for k := range na {
			var target int
			switch rng.IntN(6) {
			case 0:
				for target = rng.IntN(sceneW * sceneH); used[target]; target = rng.IntN(sceneW * sceneH) {
				}
			default:
				target = units[na+rng.IntN(nd)].Cell
			}
			g.Orders = append(g.Orders, Order{k, target})
		}
		if rng.IntN(3) == 0 {
			g.Moved = map[int]int{}
			for k := na; k < na+nd; k++ {
				if rng.IntN(2) == 0 {
					g.Moved[k] = cell()
				}
			}
		}
		if rng.IntN(2) == 0 {
			for k := na; k < na+nd; k++ {
				g.Idle = append(g.Idle, k)
			}
		}
		run(fmt.Sprintf("random %d", i), g)
	}

	for _, name := range []string{"order skipped", "target destroyed before its order", "surviving target fought again",
		"attacker destroyed before a later order", "list order changes the result", "other side's list present",
		"other side moved"} {
		if seen[name] == 0 {
			t.Errorf("no list with %s", name)
		}
	}
	t.Logf("%v", seen)
}
