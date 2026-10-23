package dosrules

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"
)

// Energy: the building credits, repair and production.

const cBuilder = 16 // the construction unit

// depotRec is a building record of terrain flag 0x100 that the scenes use,
// covering cell depot.
const depotRec, depot = HoldFlag100, 76

// inSlot is a unit placed in a slot of side's part of a record.
type inSlot struct{ rec, side, slot, unit int }

// energyCycle is hqCycle in side 0's action phase (repairs and builds are
// given there) with side 0's HQ record covering hq0, a depot
// record of side 1 covering depot, the HQ crediting 20 and building
// infantry and panzers, the depot crediting 7, a neutral record crediting
// 99, the units of units and the slots of slots.
func energyCycle(t *testing.T, p Profile, units map[int]Placement, slots ...inSlot) *Game {
	t.Helper()
	h := EmptyHold()
	h.Cells[0] = 2 * hq0
	d := EmptyHold()
	d.Flags, d.Cells[0] = 1, 2*depot
	n := EmptyHold()
	n.Flags = 2
	holds := map[int]Hold{HoldFlag40: h, depotRec: d, depotRec + 1: n}
	for _, s := range slots {
		r := holds[s.rec]
		r.Slots[7*s.side+s.slot] = s.unit
		holds[s.rec] = r
	}
	g := hqCycleHolds(t, p, units, holds)
	g.Buildings = map[int]Building{
		HoldFlag40: {Credits: 20, Builds: []int{cInfantry, cPanzer, cBuilder, 13}},
		depotRec:   {Credits: 7}, depotRec + 1: {Credits: 99},
	}
	g.Players[0].Phase, g.Players[1].Phase = PhaseAction, PhaseMovement
	return g
}

// held puts unit i into slot of side's part of record r.
func held(g *Game, r, side, slot, i int) {
	h := g.Holds[r]
	h.Slots[7*side+slot] = i
	g.Holds[r] = h
	u := g.Units[i]
	u.Flags4 |= 0x4000
	u.Pos = [2]int{h.Cells[0], h.Cells[0]}
	g.Units[i] = u
	for k := range 2 {
		for c := range g.Width * g.Height {
			if g.Maps[k][2*c+1] == i {
				g.Maps[k][2*c+1] = 0xFF
			}
		}
	}
}

func TestEnergyCredits(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := energyCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cInfantry, Owner: 1, Cell: 44},
			})
			g.Energy = [2]int{5, 1}
			evs := resolve(t, g)
			if g.Energy != [2]int{25, 8} {
				t.Errorf("energy %v, want [25 8]", g.Energy)
			}
			var got [][3]int
			for _, e := range evs {
				if e.Kind == EvCredit {
					got = append(got, [3]int{e.Side, e.Energy, e.Turn})
				}
			}
			if want := [][3]int{{0, 25, g.Turn}, {1, 8, g.Turn}}; !slices.Equal(got, want) {
				t.Errorf("credit events %v, want %v", got, want)
			}
			if i := slices.IndexFunc(evs, func(e Event) bool { return e.Kind == EvCredit }); i < 1 || evs[i-1].Kind != EvPhase {
				t.Errorf("credits do not follow the phase change: %+v", evs)
			}
		})
	}
}

func TestEnergyRepair(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := energyCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Size: 2, Exp: 4}, 1: {Type: cInfantry, Owner: 1, Cell: 44},
				2: {Type: cBuilder, Owner: 0, Size: 1}, 3: {Type: cInfantry, Owner: 0, Cell: 42, Size: 1},
			}, inSlot{HoldFlag40, 0, 0, 0}, inSlot{HoldFlag40, 0, 1, 2})
			g.Energy[0] = 65
			b := g.Units[2]
			if b.Hold != 2 {
				t.Errorf("a new construction unit has %d charges, want 2", b.Hold)
			}
			b.Hold = 0
			g.Units[2] = b

			refusal(t, "an empty slot", g.Check(Command{Kind: CmdRepair, Side: 0, Hold: HoldFlag40, Slot: 3}), 0)
			refusal(t, "the other side's slot", g.Check(Command{Kind: CmdRepair, Side: 1, Hold: HoldFlag40, Slot: 0}), 0)
			refusal(t, "no building", g.Check(Command{Kind: CmdRepair, Side: 0, Hold: cTransportRec, Slot: 0}), 0)

			evs := apply(t, g, Command{Kind: CmdRepair, Side: 0, Hold: HoldFlag40, Slot: 0})
			u := g.Units[0]
			if len(evs) != 1 || evs[0].Kind != EvRepaired || evs[0].Unit != 0 || evs[0].Energy != 50 {
				t.Errorf("repair events %+v", evs)
			}
			if u.Size != p.Tables.Types[cInfantry].Group || u.Exp != 4 || u.Flags4&0x200 == 0 || g.Energy[0] != 50 {
				t.Errorf("after the repair: size %d exp %d flags %#x energy %d", u.Size, u.Exp, u.Flags4, g.Energy[0])
			}
			if len(g.Players[0].Orders) != 0 {
				t.Errorf("the repair entered the order list: %v", g.Players[0].Orders)
			}
			refusal(t, "a second repair", g.Check(Command{Kind: CmdRepair, Side: 0, Hold: HoldFlag40, Slot: 0}), 0)

			apply(t, g, Command{Kind: CmdRepair, Side: 0, Hold: HoldFlag40, Slot: 1})
			if b := g.Units[2]; b.Size != p.Tables.Types[cBuilder].Group || b.Hold != 2 || g.Energy[0] != 0 {
				t.Errorf("after the builder's repair: size %d charges %d energy %d", b.Size, b.Hold, g.Energy[0])
			}
		})
	}
}

func TestEnergyRepairRefused(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := energyCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Size: 2}, 1: {Type: cInfantry, Owner: 1, Size: 2},
				2: {Type: cBuilder, Owner: 0, Size: 1},
			}, inSlot{HoldFlag40, 0, 0, 0}, inSlot{depotRec, 1, 0, 1}, inSlot{HoldFlag40, 0, 1, 2})
			g.Energy = [2]int{49, 100}
			refusal(t, "a builder without 50", g.Check(Command{Kind: CmdRepair, Side: 0, Hold: HoldFlag40, Slot: 1}), 0)
			refusal(t, "the movement phase", g.Check(Command{Kind: CmdRepair, Side: 1, Hold: depotRec, Slot: 0}), 0)
			g.Energy[0] = 14
			refusal(t, "less than 15", g.Check(Command{Kind: CmdRepair, Side: 0, Hold: HoldFlag40, Slot: 0}), 0)

			// A unit of side 0 left in a record side 1 holds is not repaired there.
			held(g, depotRec, 0, 1, 2)
			g.Energy[0] = 100
			refusal(t, "a foreign building", g.Check(Command{Kind: CmdRepair, Side: 0, Hold: depotRec, Slot: 1}), 0)
		})
	}
}

func TestEnergyBuild(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := energyCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cInfantry, Owner: 1, Cell: 44},
				3: {Type: cInfantry, Owner: 0, Cell: 42},
			})
			cost := p.Tables.Types[cPanzer].Cost
			g.Energy[0] = cost + p.Tables.Types[cBuilder].Cost
			s := g.Units[3]
			s.Serial = 5
			g.Units[3] = s

			evs := apply(t, g, Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cPanzer})
			if len(evs) != 1 || evs[0].Kind != EvBuilt || evs[0].Unit != 2 || evs[0].Cell != hq0 {
				t.Fatalf("build events %+v", evs)
			}
			u, h := g.Units[2], g.Holds[HoldFlag40]
			m := moveTables(t, p).Types[cPanzer]
			want := Live{Budget: m.Budget, Size: p.Tables.Types[cPanzer].Group, Flags4: m.Class | 0x4200,
				Flags6: m.Flags, Type: cPanzer, Serial: 6, Hold: 0xFF, Pos: [2]int{2 * hq0, 2 * hq0}}
			if u != want {
				t.Errorf("built unit\n got %+v\nwant %+v", u, want)
			}
			if h.Slots[0] != 2 || g.Energy[0] != p.Tables.Types[cBuilder].Cost || g.occupant(0, hq0) != InBuilding {
				t.Errorf("after the build: slots %v energy %d map %#x", h.Slots, g.Energy[0], g.occupant(0, hq0))
			}
			if len(g.Players[0].Orders) != 0 {
				t.Errorf("the build entered the order list: %v", g.Players[0].Orders)
			}

			apply(t, g, Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cBuilder})
			if b := g.Units[4]; b.Type != cBuilder || b.Hold != 2 || g.Energy[0] != 0 || g.Holds[HoldFlag40].Slots[1] != 4 {
				t.Errorf("built builder %+v, energy %d, slots %v", b, g.Energy[0], g.Holds[HoldFlag40].Slots)
			}

			g.Energy[0] = 1000
			g.Players[0].Phase = PhaseMovement
			refusal(t, "the movement phase", g.Check(Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cInfantry}), 0)
			g.Players[0].Phase = PhaseAction
			refusal(t, "a type the building does not build", g.Check(Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cSupply}), 0)
			refusal(t, "a type never sold", g.Check(Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: 13}), 0)
			refusal(t, "no build list", g.Check(Command{Kind: CmdBuild, Side: 1, Hold: depotRec, Type: cInfantry}), 0)
			refusal(t, "another side's building", g.Check(Command{Kind: CmdBuild, Side: 1, Hold: HoldFlag40, Type: cInfantry}), 0)
			g.Energy[0] = p.Tables.Types[cInfantry].Cost - 1
			refusal(t, "too little energy", g.Check(Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cInfantry}), 0)
			g.Energy[0] = 1000
			for range 5 {
				apply(t, g, Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cInfantry})
			}
			refusal(t, "no free slot", g.Check(Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cInfantry}), 0)
		})
	}
}

// TestEnergyBuildExit builds a unit in side 0's action phase and takes it out
// of the HQ in the next movement phase.
func TestEnergyBuildExit(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := energyCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cInfantry, Owner: 1, Cell: 44},
			})
			g.Energy[0] = 100
			apply(t, g, Command{Kind: CmdBuild, Side: 0, Hold: HoldFlag40, Type: cInfantry})
			refusal(t, "an exit in the phase it was built", g.Check(Command{Kind: CmdExit, Side: 0, Hold: HoldFlag40, Slot: 0, Cell: hq0 + 1}), 0)
			resolve(t, g)
			if g.Players[0].Phase != PhaseMovement {
				t.Fatalf("phases %+v", g.Players)
			}
			n := neighbours(hq0, cycleW, cycleH)
			evs := apply(t, g, Command{Kind: CmdExit, Side: 0, Hold: HoldFlag40, Slot: 0, Cell: n[0]})
			if evs[0].Kind != EvMoved || g.occupant(0, n[0]) != 2 {
				t.Errorf("exit events %+v, occupant %d", evs, g.occupant(0, n[0]))
			}
		})
	}
}

func TestEnergySave(t *testing.T) {
	g := midRound(t, Full)
	g.Energy = [2]int{12, 34}
	g.Buildings = map[int]Building{HoldFlag40: {Credits: 20, Builds: []int{cInfantry}}, depotRec: {Credits: 7}}
	if h := reload(t, g); !reflect.DeepEqual(g, h) {
		t.Errorf("loaded game differs:\n got %+v %+v\nwant %+v %+v", h.Energy, h.Buildings, g.Energy, g.Buildings)
	}
}

// TestCosts checks the prices against the full build's unit table.
func TestCosts(t *testing.T) {
	b, err := os.ReadFile("../data/testdata/unitdat.json")
	if err != nil {
		t.Fatal(err)
	}
	var recs []struct {
		Cost int `json:"cost"`
	}
	if err := json.Unmarshal(b, &recs); err != nil {
		t.Fatal(err)
	}
	if len(recs) != len(FullTables.Types) {
		t.Fatalf("%d records, %d types", len(recs), len(FullTables.Types))
	}
	for i, r := range recs {
		if FullTables.Types[i].Cost != r.Cost {
			t.Errorf("type %d: cost %d, record %d", i, FullTables.Types[i].Cost, r.Cost)
		}
	}
	if !slices.Equal(Full.Builders, []int{cBuilder}) || !slices.Equal(Demo.Builders, Full.Builders) {
		t.Errorf("builders %v / %v", Full.Builders, Demo.Builders)
	}
}
