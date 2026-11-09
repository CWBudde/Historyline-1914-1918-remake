package dosrules

import (
	"encoding/json"
	"slices"
	"testing"
)

const (
	cGas     = 51 // the gas pioneers (extension type)
	cArmCar  = 0  // a ground vehicle
	cFighter = 5  // an aircraft
)

func TestExtendedTables(t *testing.T) {
	x, f := ExtendedTables, FullTables
	if len(f.Types) != 51 || len(f.Move.Types) != 51 {
		t.Fatalf("the full tables changed: %d types", len(f.Types))
	}
	if len(x.Types) != len(x.Move.Types) || len(x.Types) <= 51 {
		t.Fatalf("extended: %d types, %d move types", len(x.Types), len(x.Move.Types))
	}
	if !slices.Equal(x.Types[:51], f.Types) || !slices.Equal(x.Move.Types[:51], f.Move.Types) {
		t.Error("the extension changes an original record")
	}
	if x.Build != f.Build+"+units1" || x.Terms != f.Terms || len(x.Move.Terrains) != len(f.Move.Terrains) {
		t.Errorf("extended build %q, terms or terrains differ", x.Build)
	}
	if !Extended.gasType(cGas) || Full.gasType(cGas) || Extended.gasType(cSapper) {
		t.Error("gasType")
	}
	for i, m := range f.Move.Types {
		if m.Flags&flagGas != 0 {
			t.Errorf("original type %d has the gas flag", i)
		}
	}
}

// gasScene is an Extended game on open terrain with side 1's gas pioneer 1
// at from and the units of more.
func gasScene(t *testing.T, from int, more map[int]Placement) *Game {
	t.Helper()
	units := map[int]Placement{1: {Type: cGas, Owner: 1, Cell: from}}
	for i, u := range more {
		units[i] = u
	}
	return newCycle(t, Extended, units)
}

// release gives pioneer 1 its order on cell and resolves side 1's phase;
// it returns the resolution's events.
func release(t *testing.T, g *Game, cell int) []Event {
	t.Helper()
	evs := apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: cell})
	if len(evs) != 1 || evs[0].Kind != EvWork || evs[0].Messages[0].Code != MsgGas {
		t.Fatalf("order events %+v", evs)
	}
	if u := g.Units[1]; u.Flags4&0x200 == 0 || !u.Releasing() || len(g.Players[1].Orders) != 0 {
		t.Fatalf("after the order: %+v, list %v", u, g.Players[1].Orders)
	}
	return resolve(t, g)
}

// kinds are the events of kind k.
func kinds(evs []Event, k EventKind) []Event {
	var out []Event
	for _, ev := range evs {
		if ev.Kind == k {
			out = append(out, ev)
		}
	}
	return out
}

func TestGasTargets(t *testing.T) {
	n := neighbours(40, cycleW, cycleH)
	g := gasScene(t, 40, map[int]Placement{0: {Type: cInfantry, Owner: 0, Cell: n[0]}})
	for k, c := range n {
		if err := g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: c}); err != nil {
			t.Errorf("neighbour %d (%d): %v", k, c, err)
		}
	}
	refusal(t, "a cell beyond the neighbours", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 42}), MsgNoTarget)
	refusal(t, "its own cell", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40}), MsgNoTarget)
	if got := g.GasCells(1, 1, 31); !slices.Equal(got, []int{30, 31, 32}) {
		t.Errorf("cloud above an even column: %v", got)
	}
	if got := g.GasCells(1, 1, 42); got != nil {
		t.Errorf("cloud beyond the neighbours: %v", got)
	}
	u := g.Units[1]
	u.Hold = 0
	g.Units[1] = u
	refusal(t, "no charge", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 31}), MsgNoGas)
}

func TestGasCloud(t *testing.T) {
	// The infantry 0 of side 0 stands in the cloud, side 1's infantry 2
	// too (friendly fire); side 0's infantry 3 beside it stays clear.
	g := gasScene(t, 40, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: 31},
		2: {Type: cInfantry, Owner: 1, Cell: 30},
		3: {Type: cInfantry, Owner: 0, Cell: 22},
	})
	full := Extended.Tables.Types[cInfantry].Group
	if g.Units[1].Hold != GasCharges {
		t.Fatalf("charges %d", g.Units[1].Hold)
	}
	evs := release(t, g, 31)
	cloud := kinds(evs, EvGasCloud)
	if len(cloud) != 1 || cloud[0].Cell != 31 || !slices.Equal(cloud[0].Cells, []int{30, 31, 32}) {
		t.Fatalf("cloud %+v", cloud)
	}
	if len(kinds(evs, EvBattle)) != 0 || len(kinds(evs, EvGassed)) != 0 {
		t.Errorf("side 1's resolution fought or gassed: %+v", evs)
	}
	if g.Units[1].Hold != GasCharges-1 || g.GasAt(31) != GasTurns || g.GasAt(22) != 0 {
		t.Errorf("after the release: charges %d, gas %v", g.Units[1].Hold, g.Gas)
	}

	// Side 0's resolution counts the turn: the units in the cloud lose 30 %
	// of the group, rounded up, and the cloud thins.
	loss := (full*GasLoss + 99) / 100
	evs = resolve(t, g)
	gassed := kinds(evs, EvGassed)
	if len(gassed) != 2 || gassed[0].Unit != 0 || gassed[1].Unit != 2 || gassed[0].Loss != loss || gassed[1].Side != 1 {
		t.Fatalf("gassed %+v", gassed)
	}
	for i, want := range map[int]int{0: full - loss, 2: full - loss, 3: full, 1: full} {
		if g.Units[i].Size != want {
			t.Errorf("unit %d: size %d, want %d", i, g.Units[i].Size, want)
		}
	}
	if g.GasAt(31) != GasTurns-1 {
		t.Errorf("gas %v", g.Gas)
	}

	// Side 1's next resolution leaves the cloud; side 0's thins it away.
	if evs := resolve(t, g); len(kinds(evs, EvGassed)) != 0 || g.GasAt(31) != GasTurns-1 {
		t.Errorf("side 1's resolution: %+v, gas %v", evs, g.Gas)
	}
	evs = resolve(t, g)
	if len(kinds(evs, EvGassed)) != 2 || g.Units[0].Size != full-2*loss {
		t.Errorf("second turn: %+v, size %d", evs, g.Units[0].Size)
	}
	if cl := kinds(evs, EvGasCleared); len(cl) != 1 || !slices.Equal(cl[0].Cells, []int{30, 31, 32}) || g.Gas != nil {
		t.Errorf("cleared %+v, gas %v", cl, g.Gas)
	}
}

func TestGasLossByKind(t *testing.T) {
	g := gasScene(t, 40, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: 31},
		2: {Type: cArmCar, Owner: 0, Cell: 30},
		3: {Type: cFighter, Owner: 0, Cell: 32},
	})
	g.GasLoss = 15 // masks
	release(t, g, 31)
	got := map[int]int{}
	for _, ev := range kinds(resolve(t, g), EvGassed) {
		got[ev.Unit] = ev.Loss
	}
	// 15 % of 6 is 1 for the infantry; 5 % of the car's 6 rounds up to 1;
	// the fighter loses nothing.
	if len(got) != 2 || got[0] != 1 || got[2] != 1 {
		t.Errorf("losses %v", got)
	}
	if g.Units[3].Size != Extended.Tables.Types[cFighter].Group {
		t.Errorf("the fighter lost strength: %+v", g.Units[3])
	}
}

func TestGasDestroys(t *testing.T) {
	g := gasScene(t, 40, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: 31, Size: 1},
		2: {Type: cInfantry, Owner: 0, Cell: 13},
	})
	g.GasTurns = 4
	release(t, g, 31)
	if g.GasAt(31) != 4 {
		t.Fatalf("duration %v", g.Gas)
	}
	evs := resolve(t, g)
	if ev := kinds(evs, EvGassed); len(ev) != 1 || ev[0].Loss != 1 {
		t.Fatalf("gassed %+v", ev)
	}
	u := g.Units[0]
	if !u.Destroyed() || u.Flags4&0x8000 == 0 {
		t.Errorf("unit 0 not destroyed: %+v", u)
	}
	for k := range 2 {
		if g.occupant(k, 31) == 0 {
			t.Errorf("map %d still shows unit 0", k)
		}
	}
	if g.Destroyed() == nil {
		t.Error("Destroyed lists nothing")
	}
}

func TestGasCharges(t *testing.T) {
	g := gasScene(t, 40, map[int]Placement{0: {Type: cInfantry, Owner: 0, Cell: 4}})
	release(t, g, 31)
	resolve(t, g) // side 0
	release(t, g, 39)
	if g.Units[1].Hold != 0 {
		t.Fatalf("charges %d", g.Units[1].Hold)
	}
	resolve(t, g) // side 0
	refusal(t, "spent", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 31}), MsgNoGas)
	if n, ok := Extended.charges(cGas); !ok || n != GasCharges {
		t.Errorf("charges(%d) = %d, %v", cGas, n, ok)
	}
	if _, ok := Extended.charges(cInfantry); ok {
		t.Error("infantry holds charges")
	}
}

func TestGasSave(t *testing.T) {
	g := gasScene(t, 40, map[int]Placement{0: {Type: cInfantry, Owner: 0, Cell: 31}})
	g.GasLoss = 15
	b, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(b, &keys); err != nil {
		t.Fatal(err)
	}
	if _, ok := keys["gas"]; ok {
		t.Error("a save without a cloud writes gas")
	}
	release(t, g, 31)
	b, err = g.Save()
	if err != nil {
		t.Fatal(err)
	}
	h, err := LoadGame(b)
	if err != nil {
		t.Fatal(err)
	}
	if h.Profile.Tables != ExtendedTables || !slices.Equal(h.Gas, g.Gas) || h.GasLoss != 15 || h.GasTurns != 0 {
		t.Errorf("loaded: tables %q, gas %v, loss %d", h.Profile.Tables.Build, h.Gas, h.GasLoss)
	}
	again, err := h.Save()
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(b) {
		t.Error("the loaded game saves differently")
	}
	var s savedGame
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	s.Gas = s.Gas[:3]
	bad, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGame(bad); err == nil {
		t.Error("a cloud of the wrong size loads")
	}
}
