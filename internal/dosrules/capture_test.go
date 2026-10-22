package dosrules

import (
	"reflect"
	"slices"
	"testing"
)

// Capture: a unit of the movement side that ends on another side's or a
// neutral building takes it and the units in it at the next resolution.

// buildingTerrain is the terrain record of kind flag (0x40, 0x100, 0x400)
// with owner bits owner (4 side 0, 1 side 1, 2 neutral).
func buildingTerrain(t *testing.T, p Profile, flag, owner int) int {
	t.Helper()
	for i, tr := range moveTables(t, p).Terrains {
		if tr.Flags&flag != 0 && tr.Flags&7 == owner {
			return i
		}
	}
	t.Fatalf("no terrain of flag %#x and owner %d", flag, owner)
	return 0
}

// depotCycle is newCycle with a depot record of owner (0, 1, 2 neutral) at
// cell depot, its entrance terrain showing the owner, the units of inside in
// both sides' slots.
func depotCycle(t *testing.T, p Profile, owner int, units map[int]Placement, inside ...int) *Game {
	t.Helper()
	tr := openTerrain(t, p)
	terrain := make([]int, cycleW*cycleH)
	for i := range terrain {
		terrain[i] = tr
	}
	terrain[depot] = buildingTerrain(t, p, 0x100, [3]int{4, 1, 2}[owner])
	h := EmptyHold()
	h.Flags, h.Cells[0] = owner, 2*depot
	for k, u := range inside {
		h.Slots[k], h.Slots[7+k] = u, u
	}
	g, err := NewGame(p, cycleW, cycleH, terrain, units, map[int]Hold{depotRec: h})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestCaptureNeutralDepot(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(depot, cycleW, cycleH)
			g := depotCycle(t, p, 2, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: n[0]},
				1: {Type: cPanzer, Owner: 1, Cell: 0},
				2: {Type: cPanzer, Owner: 2, Cell: depot},
			}, 2)
			g.Buildings = map[int]Building{depotRec: {Credits: 7}}
			if f := g.Holds[depotRec].Flags; f != 2|KindDepot {
				t.Fatalf("depot flags %#x, want the neutral bit and the depot kind", f)
			}
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: depot})
			evs := resolve(t, g)
			i := slices.IndexFunc(evs, func(e Event) bool { return e.Kind == EvCaptured })
			if i < 0 || !reflect.DeepEqual(evs[i], Event{Kind: EvCaptured, Side: 0, Unit: 0, Cell: depot, Hold: depotRec}) {
				t.Fatalf("no capture event: %+v", evs)
			}
			if j := slices.IndexFunc(evs, func(e Event) bool { return e.Kind == EvPhase }); j < i {
				t.Errorf("the capture follows the phase change: %+v", evs)
			}
			h := g.Holds[depotRec]
			if h.Flags != KindDepot {
				t.Errorf("depot flags %#x, want side 0's depot", h.Flags)
			}
			for k := range 2 {
				if s := h.Slots[7*k : 7*k+2]; !slices.Equal(s, []int{2, 0}) {
					t.Errorf("side %d's slots %v, want the unit found and the capturer", k, s)
				}
				if tr := g.Maps[k][2*depot]; tr != buildingTerrain(t, p, 0x100, 4) {
					t.Errorf("side %d's map shows terrain %d at the entrance, want side 0's depot", k, tr)
				}
				if occ := g.occupant(k, depot); occ == 0 || occ == 2 {
					t.Errorf("side %d's map shows unit %d on the entrance", k, occ)
				}
			}
			found, capturer := g.Units[2], g.Units[0]
			if !Owns(found.Flags4, 0) || found.Flags4&3 != 0 {
				t.Errorf("the unit found is not side 0's: %#x", found.Flags4)
			}
			if capturer.Flags4&0x4000 == 0 || capturer.Pos != [2]int{2 * depot, 2 * depot} {
				t.Errorf("the capturer is not in the depot: %+v", capturer)
			}
			if g.Energy[0] != 7 || g.Energy[1] != 0 {
				t.Errorf("energy %v, want the depot's 7 for side 0", g.Energy)
			}
		})
	}
}

// TestCaptureEnemyDepotFull takes a depot whose seven slots are full: the
// unit in the first slot is destroyed and the capturer takes its place.
func TestCaptureEnemyDepotFull(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(depot, cycleW, cycleH)
			units := map[int]Placement{0: {Type: cInfantry, Owner: 0, Cell: n[0]}}
			var inside []int
			for i := 1; i <= 7; i++ {
				units[i] = Placement{Type: cInfantry, Owner: 1, Cell: depot}
				inside = append(inside, i)
			}
			g := depotCycle(t, p, 1, units, inside...)
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: depot})
			resolve(t, g)
			h := g.Holds[depotRec]
			if h.Flags != KindDepot {
				t.Errorf("depot flags %#x", h.Flags)
			}
			if !slices.Equal(h.Slots[:7], []int{0, 2, 3, 4, 5, 6, 7}) {
				t.Errorf("side 0's slots %v", h.Slots[:7])
			}
			if !g.Units[1].Destroyed() {
				t.Error("the unit in the first slot survives")
			}
			for i := 2; i <= 7; i++ {
				if u := g.Units[i]; u.Flags4&3 != 0 || u.Destroyed() {
					t.Errorf("unit %d: %#x", i, u.Flags4)
				}
			}
		})
	}
}

// TestCaptureOnlyFoot keeps a unit without type flag bit 1 off a foreign
// building's entrance.
func TestCaptureOnlyFoot(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			nb := neighbours(depot, cycleW, cycleH)
			n := slices.DeleteFunc(nb[:], func(c int) bool { return c < 0 })
			g := depotCycle(t, p, 2, map[int]Placement{
				0: {Type: cPanzer, Owner: 0, Cell: n[0]}, 1: {Type: cInfantry, Owner: 0, Cell: n[1]},
			})
			for u, want := range []bool{false, true} {
				cells, err := p.Destinations(g.Movement(0, u))
				if err != nil {
					t.Fatal(err)
				}
				if got := slices.Contains(cells, depot); got != want {
					t.Errorf("unit %d reaches the neutral depot: %v, want %v", u, got, want)
				}
			}
		})
	}
}

// TestCaptureOnce: the action bytes are cleared at the phase change, so a
// later resolution does not capture again.
func TestCaptureOnce(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(depot, cycleW, cycleH)
			g := depotCycle(t, p, 2, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: n[0]}, 1: {Type: cPanzer, Owner: 1, Cell: 0},
			})
			apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: depot})
			resolve(t, g)
			for range 2 {
				for _, e := range resolve(t, g) {
					if e.Kind == EvCaptured {
						t.Errorf("captured again: %+v", e)
					}
				}
			}
		})
	}
}
