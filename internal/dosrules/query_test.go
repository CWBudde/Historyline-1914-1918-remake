package dosrules

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

// The container state NewGame derives, the queries a session asks of a
// game, attacks of units off the map and the save of a game with a unit in
// a building in winter.

// containerScene is a game on hqCycle's map: side 0's transporter at 40
// carrying unit 10, infantry 0 next to it, infantry 5 in side 0's HQ, a
// construction unit 6 and side 1's infantry 7.
func containerScene(t *testing.T, p Profile) *Game {
	t.Helper()
	n := neighbours(40, cycleW, cycleH)
	hq := EmptyHold()
	hq.Cells[0], hq.Slots[0] = 2*hq0, 5
	return hqCycleHolds(t, p, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: n[0]}, 1: {Type: cSupply, Owner: 0, Cell: 40},
		5: {Type: cInfantry, Owner: 0}, 6: {Type: cBuilder, Owner: 0, Cell: 60},
		7: {Type: cInfantry, Owner: 1, Cell: 44}, 10: {Type: cInfantry, Owner: 0},
	}, map[int]Hold{cTransportRec: transportHold(40, 0, 1, 10), HoldFlag40: hq})
}

func TestNewGameContainers(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := containerScene(t, p)
			tr, cargo, in := g.Units[1], g.Units[10], g.Units[5]
			if tr.Hold != cTransportRec || tr.Flags4&0x4000 != 0 || g.occupant(0, 40) != 1 || g.occupant(1, 40) != 1 {
				t.Errorf("transporter %+v, map %d/%d", tr, g.occupant(0, 40), g.occupant(1, 40))
			}
			if cargo.Flags4&0x4000 == 0 || cargo.Pos != [2]int{2*40 + 1, 2*40 + 1} || cargo.Hold != 0xFF {
				t.Errorf("cargo %+v", cargo)
			}
			if in.Flags4&0x4000 == 0 || in.Pos != [2]int{2 * hq0, 2 * hq0} {
				t.Errorf("unit in the HQ %+v", in)
			}
			if g.occupant(0, hq0) != InBuilding || g.occupant(1, hq0) != 0xFF {
				t.Errorf("HQ cell shows %#x/%#x", g.occupant(0, hq0), g.occupant(1, hq0))
			}
			for k := range 2 {
				for c := range cycleW * cycleH {
					if o := g.occupant(k, c); o == 5 || o == 10 {
						t.Errorf("map %d shows held unit %d at %d", k, o, c)
					}
				}
			}
			if g.Units[6].Hold != builderCharges || g.Units[0].Hold != 0xFF {
				t.Errorf("charges %d, plain unit's hold %#x", g.Units[6].Hold, g.Units[0].Hold)
			}

			// A unit that reaches its own transporter: the destinations work
			// and the move boards it.
			dests, err := p.Destinations(g.Movement(0, 0))
			if err != nil {
				t.Fatalf("destinations: %v", err)
			}
			if !slices.Contains(dests, 40) {
				t.Errorf("the transporter's cell is no destination: %v", dests)
			}
			evs := apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: 40})
			if evs[0].Kind != EvMoved || g.Holds[cTransportRec].Slots[1] != 0 || g.Units[0].Flags4&0x4000 == 0 || g.occupant(0, 40) != 1 {
				t.Errorf("boarding: events %+v, slots %v", evs, g.Holds[cTransportRec].Slots)
			}
		})
	}
}

func TestNewGameContainersRefused(t *testing.T) {
	terrain := make([]int, cycleW*cycleH)
	for i := range terrain {
		terrain[i] = openTerrain(t, Full)
	}
	two := EmptyHold()
	two.Slots[0] = 3
	for what, c := range map[string]struct {
		units map[int]Placement
		holds map[int]Hold
	}{
		"an unplaced unit in a slot": {map[int]Placement{1: {Type: cInfantry, Cell: 4}}, map[int]Hold{HoldFlag40: two}},
		"a unit in two records": {map[int]Placement{3: {Type: cInfantry}},
			map[int]Hold{HoldFlag40: two, HoldFlag100: two}},
		"a transporter without a record": {map[int]Placement{1: {Type: cSupply, Cell: 40}},
			map[int]Hold{0: transportHold(41, 0, 1)}},
		"two transporters on one record": {map[int]Placement{1: {Type: cSupply, Cell: 40}, 2: {Type: cSupply, Cell: 40}},
			map[int]Hold{0: transportHold(40, 0, 1)}},
		"a record past the table": {nil, map[int]Hold{HoldRecords: EmptyHold()}},
	} {
		if _, err := NewGame(Full, cycleW, cycleH, terrain, c.units, c.holds); err == nil {
			t.Errorf("%s: accepted", what)
		}
	}
}

// TestCycleAttackOffMap checks the assumed rule that an attacker must be on
// its side's map.
func TestCycleAttackOffMap(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			g := newCycleHolds(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40},
				1: {Type: cSupply, Owner: 1, Cell: n[0]}, 2: {Type: cInfantry, Owner: 1},
				3: {Type: cInfantry, Owner: 1, Cell: n[3]}, 4: {Type: cInfantry, Owner: 1, Cell: n[1]},
			}, map[int]Hold{cTransportRec: transportHold(n[0], 1, 1, 2)})
			refusal(t, "a held unit's attack", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 2, Cell: 40}), 0)
			g.Maps[1][2*n[3]+1] = 4
			refusal(t, "an attacker its map does not show", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 3, Cell: 40}), 0)
			if err := g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 4, Cell: 40}); err != nil {
				t.Errorf("an attacker on the map: %v", err)
			}
		})
	}
}

func TestExitMovement(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			g := containerScene(t, p)
			m, err := g.ExitMovement(0, cTransportRec, 0)
			if err != nil {
				t.Fatal(err)
			}
			if _, on := m.Units[10]; m.Unit != 10 || m.Budget != 8 || m.Block != 3 || m.Held[10].Cell != 40 || on {
				t.Errorf("query: unit %d budget %d block %d held %+v", m.Unit, m.Budget, m.Block, m.Held[10])
			}
			dests, err := p.Destinations(m)
			if err != nil || !slices.Contains(dests, n[2]) {
				t.Errorf("destinations %v, %v", dests, err)
			}

			// The exit's own reach search is the query's.
			tab := moveTables(t, p)
			r, err := tab.Exit(g.Movement(0, 10), ExitOrder{Hold: cTransportRec, Slot: 0, MoveOrder: MoveOrder{Goal: n[2], Retarget: -1}})
			if err != nil {
				t.Fatal(err)
			}
			buf, _, _ := tab.Reach(m)
			if !slices.Equal(buf, r.Remaining) {
				t.Error("the exit's reach differs from ExitMovement's")
			}

			// From the HQ, at the record's first cell.
			if m, err := g.ExitMovement(0, HoldFlag40, 0); err != nil || m.Unit != 5 || m.Held[5].Cell != hq0 {
				t.Errorf("HQ exit: %+v, %v", m, err)
			}
			for what, c := range map[string][3]int{"an empty slot": {0, cTransportRec, 1}, "no record": {0, 9, 0},
				"a bad slot": {0, cTransportRec, 7}, "a bad side": {2, cTransportRec, 0}} {
				if _, err := g.ExitMovement(c[0], c[1], c[2]); err == nil {
					t.Errorf("%s: no error", what)
				}
			}
		})
	}
}

func TestHoldAt(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := containerScene(t, p)
			d := EmptyHold()
			d.Cells[0] = 2 * depot
			g.Holds[depotRec] = d
			depotTerrain := -1
			for i, tr := range moveTables(t, p).Terrains {
				if tr.Flags&0x500 == 0x100 {
					depotTerrain = i
					break
				}
			}
			for k := range 2 {
				g.Maps[k][2*depot] = depotTerrain
				g.Maps[k][2*(depot+1)] = depotTerrain
			}
			for _, c := range []struct {
				side, cell, rec int
				ok              bool
			}{
				{0, 40, cTransportRec, true}, {1, 40, cTransportRec, true}, {0, hq0, HoldFlag40, true},
				{0, hq1, HoldFlag40 + 1, true}, {0, depot, depotRec, true}, {0, depot + 1, 0, false},
				{0, 44, 0, false}, {0, 41, 0, false}, {0, -1, 0, false},
			} {
				if rec, ok := g.HoldAt(c.side, c.cell); rec != c.rec || ok != c.ok {
					t.Errorf("side %d cell %d: %d %v, want %d %v", c.side, c.cell, rec, ok, c.rec, c.ok)
				}
			}
		})
	}
}

func TestWeaponAgainst(t *testing.T) {
	for i, ty := range Full.Tables.Types {
		for _, c := range []struct{ flags, power, rng int }{
			{airTarget, ty.AirPower, ty.AirRange}, {waterTarget, ty.WaterPower, ty.SurfaceRange},
			{4, ty.GroundPower, ty.SurfaceRange},
		} {
			if pw, r := Full.WeaponAgainst(i, c.flags); pw != c.power || r != c.rng {
				t.Errorf("type %d against %#x: %d/%d, want %d/%d", i, c.flags, pw, r, c.power, c.rng)
			}
		}
	}
}

// TestSaveBuilding saves a winter game with a unit in a building, which the
// map shows as InBuilding, and loads it back; a save of schema 2 loads as
// summer.
func TestSaveBuilding(t *testing.T) {
	g := containerScene(t, Full.Season(true))
	if !g.Profile.Winter() || g.occupant(0, hq0) != InBuilding {
		t.Fatalf("set-up: winter %v, HQ cell %#x", g.Profile.Winter(), g.occupant(0, hq0))
	}
	h := reload(t, g)
	if !reflect.DeepEqual(g, h) {
		t.Fatalf("loaded game differs:\n got %+v\nwant %+v", h, g)
	}
	if !h.Profile.Winter() || &h.Profile.Tables.Move.Terrains[0] != &FullTables.Move.Winter[0] {
		t.Error("the loaded game does not play in winter")
	}

	b, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	edit := func(f func(m map[string]any)) []byte {
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		f(m)
		out, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	old, err := LoadGame(edit(func(m map[string]any) { m["schema"] = 2; delete(m, "winter") }))
	if err != nil || old.Profile.Winter() || old.occupant(0, hq0) != InBuilding {
		t.Errorf("schema 2: %v", err)
	}
	if _, err := LoadGame(edit(func(m map[string]any) { m["build"] = DemoTables.Build })); err == nil {
		t.Error("a winter save of the demo's tables loaded")
	}
	if _, err := LoadGame(edit(func(m map[string]any) {
		m["maps"].([]any)[0].([]any)[2*hq0+1] = 0x100
	})); err == nil {
		t.Error("an occupant past a byte loaded")
	}
}
