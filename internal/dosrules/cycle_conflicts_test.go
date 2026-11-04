package dosrules

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

// Moves happen immediately on the moving side's map. A second order to
// the same destination is refused without consuming that unit's order.
func TestCycleConflictingDestinations(t *testing.T) {
	for build, p := range profiles {
		for side := range 2 {
			for first := range 2 {
				t.Run(fmt.Sprintf("%s/side%d/first%d", build, side, first), func(t *testing.T) {
					g := newCycle(t, p, map[int]Placement{
						0: {Type: cInfantry, Owner: side, Cell: 40},
						1: {Type: cInfantry, Owner: side, Cell: 42},
					})
					g.Players[side].Phase, g.Players[1-side].Phase = PhaseMovement, PhaseAction
					second := 1 - first
					evs := apply(t, g, Command{Kind: CmdMove, Side: side, Unit: first, Cell: 41})
					if len(evs) != 1 || evs[0].Kind != EvMoved || g.occupant(side, 41) != first {
						t.Fatalf("first move: %+v, occupant %d", evs, g.occupant(side, 41))
					}
					refusedMoveUnchanged(t, g, Command{Kind: CmdMove, Side: side, Unit: second, Cell: 41})
					if !slices.Equal(g.Players[side].Orders, []int{first}) {
						t.Fatalf("orders after refusal: %v", g.Players[side].Orders)
					}
					// The waiting side still sees both units at their old cells.
					if g.occupant(1-side, 40) != 0 || g.occupant(1-side, 42) != 1 || g.occupant(1-side, 41) != 0xFF {
						t.Fatal("the moves changed the waiting side's map")
					}
					resolve(t, g)
					for viewer := range 2 {
						if g.occupant(viewer, 41) != first || g.occupant(viewer, 40+2*second) != second || g.occupant(viewer, 40+2*first) != 0xFF {
							t.Errorf("side %d: wrong occupants after resolution", viewer)
						}
					}
				})
			}
		}
	}
}

// Two units cannot exchange occupied cells with queued moves. Once a unit
// moves elsewhere, its old cell is available to a later move in that phase.
func TestCycleSwapsAndVacatedCells(t *testing.T) {
	for build, p := range profiles {
		for side := range 2 {
			t.Run(fmt.Sprintf("%s/side%d", build, side), func(t *testing.T) {
				g := newCycle(t, p, map[int]Placement{
					0: {Type: cInfantry, Owner: side, Cell: 40},
					1: {Type: cInfantry, Owner: side, Cell: 41},
				})
				g.Players[side].Phase, g.Players[1-side].Phase = PhaseMovement, PhaseAction
				refusedMoveUnchanged(t, g, Command{Kind: CmdMove, Side: side, Unit: 0, Cell: 41})
				refusedMoveUnchanged(t, g, Command{Kind: CmdMove, Side: side, Unit: 1, Cell: 40})
				for _, c := range []Command{
					{Kind: CmdMove, Side: side, Unit: 0, Cell: 31},
					{Kind: CmdMove, Side: side, Unit: 1, Cell: 40},
				} {
					if evs := apply(t, g, c); len(evs) != 1 || evs[0].Kind != EvMoved {
						t.Fatalf("move after refusal: %+v", evs)
					}
				}
				if !slices.Equal(g.Players[side].Orders, []int{0, 1}) {
					t.Errorf("orders: %v, want [0 1]", g.Players[side].Orders)
				}
				resolve(t, g)
				for viewer := range 2 {
					if g.occupant(viewer, 31) != 0 || g.occupant(viewer, 40) != 1 || g.occupant(viewer, 41) != 0xFF {
						t.Errorf("side %d: wrong occupants after resolution", viewer)
					}
				}
			})
		}
	}
}

func refusedMoveUnchanged(t *testing.T, g *Game, c Command) {
	t.Helper()
	before, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	if evs := apply(t, g, c); len(evs) != 1 || evs[0].Kind != EvRefused {
		t.Fatalf("occupied destination: %+v", evs)
	}
	after, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("a refused move changed the game state")
	}
}
