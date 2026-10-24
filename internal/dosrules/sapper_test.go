package dosrules

import (
	"slices"
	"testing"
)

const cSapper = 20

// sapperScene is a game on grass (record 0, which sappers work) with
// side 1's sapper 1 at from and side 0's infantry 0 at enemy; the records
// in terrain override the grass. Side 1 gives its orders first.
func sapperScene(t *testing.T, p Profile, from, enemy int, terrain map[int]int) *Game {
	t.Helper()
	tr := make([]int, cycleW*cycleH)
	for c, r := range terrain {
		tr[c] = r
	}
	g, err := NewGame(p, cycleW, cycleH, tr, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: enemy}, 1: {Type: cSapper, Owner: 1, Cell: from},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// unworked is a terrain record sappers leave alone.
func unworked(t *testing.T, p Profile) int {
	t.Helper()
	for i, tr := range moveTables(t, p).Terrains {
		if tr.Flags&(terrainWorked|0x5547) == 0 {
			return i
		}
	}
	t.Fatal("no terrain without the sapper flag")
	return 0
}

// work gives sapper 1 its order on cell, resolves the phase and returns the
// work's event.
func work(t *testing.T, g *Game, cell int) Event {
	t.Helper()
	evs := apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: cell})
	if len(evs) != 1 || evs[0].Kind != EvWork || evs[0].Messages[0].Code != MsgTrench {
		t.Fatalf("order events %+v", evs)
	}
	if u := g.Units[1]; u.Flags4&0x200 == 0 || len(g.Players[1].Orders) != 0 {
		t.Fatalf("after the order: flags %#x, list %v", u.Flags4, g.Players[1].Orders)
	}
	g.Confirm(0)
	g.Confirm(1)
	var done []Event
	for _, ev := range resolve(t, g) {
		switch ev.Kind {
		case EvBattle, EvSkipped:
			t.Errorf("the work fought: %+v", ev)
		case EvDug, EvFilled:
			done = append(done, ev)
		}
	}
	if len(done) != 1 {
		t.Fatalf("work events %+v", done)
	}
	return done[0]
}

// terrainOf is the terrain record of cell on both maps, -1 where they differ.
func terrainOf(g *Game, cell int) int {
	if g.Maps[0][2*cell] != g.Maps[1][2*cell] {
		return -1
	}
	return g.Maps[0][2*cell]
}

func TestSapperTargets(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			g := sapperScene(t, p, 40, n[0], map[int]int{n[3]: unworked(t, p)})
			for k, c := range n {
				err := g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: c})
				if k == 3 {
					refusal(t, "terrain the sapper leaves alone", err, MsgNoTarget)
				} else if err != nil {
					t.Errorf("neighbour %d (%d): %v", k, c, err)
				}
			}
			refusal(t, "a cell beyond the neighbours", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 42}), MsgNoTarget)
			refusal(t, "its own cell", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40}), MsgNoTarget)
		})
	}
}

func TestSapperDigs(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			// Even column: the target straight above, with an enemy on it,
			// bends the trench round the sapper's front.
			g := sapperScene(t, p, 40, 31, nil)
			ev := work(t, g, 31)
			if ev.Kind != EvDug || ev.Cell != 31 || !slices.Equal(ev.Cells, []int{30, 31, 32}) {
				t.Errorf("dig %+v", ev)
			}
			for c, want := range map[int]int{30: 0xAA, 31: 0xA7, 32: 0xAB, 40: 0, 22: 0} {
				if got := terrainOf(g, c); got != want {
					t.Errorf("cell %d: record %#x, want %#x", c, got, want)
				}
			}
			if g.Units[0].Size != p.Tables.Types[cInfantry].Group {
				t.Errorf("the unit on the target lost strength: %+v", g.Units[0])
			}

			// Odd column, the target to the lower right: a straight line.
			g = sapperScene(t, p, 41, 0, nil)
			if ev := work(t, g, 51); !slices.Equal(ev.Cells, []int{43, 51, 50}) {
				t.Errorf("dig %+v", ev)
			}
			for c, want := range map[int]int{43: 0xAA, 51: 0xA7, 50: 0xAA} {
				if got := terrainOf(g, c); got != want {
					t.Errorf("cell %d: record %#x, want %#x", c, got, want)
				}
			}
		})
	}
}

func TestSapperFills(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			// A trench target is filled; of its line only the cells sappers
			// work change.
			rock := unworked(t, p)
			g := sapperScene(t, p, 40, 0, map[int]int{31: 0xA7, 32: rock})
			ev := work(t, g, 31)
			if ev.Kind != EvFilled || !slices.Equal(ev.Cells, []int{30, 31}) {
				t.Errorf("fill %+v", ev)
			}
			for c, want := range map[int]int{30: TerrainFilled, 31: TerrainFilled, 32: rock} {
				if got := terrainOf(g, c); got != want {
					t.Errorf("cell %d: record %#x, want %#x", c, got, want)
				}
			}
		})
	}
}
