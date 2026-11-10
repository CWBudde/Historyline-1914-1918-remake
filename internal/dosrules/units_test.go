package dosrules

import (
	"encoding/json"
	"slices"
	"testing"
)

// The extension types beyond the gas pioneers (extension.json).
const (
	cMortar    = 52
	cStorm     = 53
	cZeppelin  = 54
	cWhippet   = 55
	cMarkV     = 56
	cSchneider = 57
	cParisGun  = 58
	cFlame     = 59
	cWiring    = 60
	cTunnel    = 61
	cAmbulance = 62
	cCavalry   = 21
)

func TestExtensionRecords(t *testing.T) {
	x, f := ExtendedTables, FullTables
	if len(x.Types) != 63 || len(x.Move.Types) != 63 {
		t.Fatalf("extended: %d types, %d move types", len(x.Types), len(x.Move.Types))
	}
	if len(x.Move.Terrains) != len(f.Move.Terrains)+2 || len(x.Move.Winter) != len(f.Move.Winter)+2 {
		t.Fatalf("extended: %d terrains, %d winter", len(x.Move.Terrains), len(x.Move.Winter))
	}
	if !slices.Equal(x.Move.Terrains[:len(f.Move.Terrains)], f.Move.Terrains) ||
		!slices.Equal(x.Move.Winter[:len(f.Move.Winter)], f.Move.Winter) {
		t.Error("the extension changes an original terrain record")
	}
	if len(f.Move.Terrains) != TerrainWire {
		t.Errorf("the wire record is %d, the full tables end at %d", TerrainWire, len(f.Move.Terrains))
	}
	for i, m := range f.Move.Types {
		if m.Flags&(flagFlame|flagWire|flagTunnel|flagMedic|flagGas) != 0 {
			t.Errorf("original type %d uses an extension flag: %#x", i, m.Flags)
		}
	}
	for typ, flag := range map[int]int{cFlame: flagFlame, cWiring: flagWire, cTunnel: flagTunnel, cAmbulance: flagMedic} {
		if !Extended.typeFlag(typ, flag) {
			t.Errorf("type %d lacks flag %#x", typ, flag)
		}
	}
	// Wire lets foot units and heavy tanks in, not cavalry or light tanks.
	wire := x.Move.Terrains[TerrainWire]
	for typ, in := range map[int]bool{cInfantry: true, cMarkV: true, cCavalry: false, cWhippet: false, cWiring: true} {
		if got := wire.Mask&x.Move.Types[typ].Mask != 0; got != in {
			t.Errorf("type %d on wire: %v, want %v", typ, got, in)
		}
	}
	if wire.Cost < x.Move.Types[cInfantry].Budget {
		t.Errorf("wire costs %d, a foot move is %d", wire.Cost, x.Move.Types[cInfantry].Budget)
	}
	if x.Types[cParisGun].SurfaceRange <= x.Types[24].SurfaceRange || !x.Types[cParisGun].setsUp() {
		t.Error("the Paris Gun outranges no train gun or does not set up")
	}
	if x.Move.Types[cZeppelin].Class&0x10 == 0 {
		t.Error("the Zeppelin does not fly")
	}
}

func TestFlameIgnoresTerrain(t *testing.T) {
	terms := ExtendedTables.Terms
	class := -1
	for c, tm := range terms {
		if tm.Defence > 0 && (class < 0 || tm.Defence > terms[class].Defence) {
			class = c
		}
	}
	if class < 0 {
		t.Fatal("no terrain with a defence term")
	}
	ty := ExtendedTables.Types
	a := Fighter{Type: ty[cFlame], Size: 6, Exp: 0, Flags: 1 | 4}
	d := Fighter{Type: ty[cInfantry], Size: 6, Exp: 0, Flags: 4}
	ctx := Context{Distance: 1, ATerrain: 8, DTerrain: class, ReturnFire: true}
	plain, err := Extended.Casualties(a, d, ctx)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Flame = true
	flame, err := Extended.Casualties(a, d, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if flame.DefenderSize >= plain.DefenderSize {
		t.Errorf("in cover (class %d): %d left with the flame, %d without", class, flame.DefenderSize, plain.DefenderSize)
	}
	if flame.AttackerSize != plain.AttackerSize {
		t.Errorf("the flame changed the return fire: %d vs %d", flame.AttackerSize, plain.AttackerSize)
	}
}

// workScene is an Extended game on worked open ground (the grass record 0)
// with side 1's unit 1 of typ at from and the units of more.
func workScene(t *testing.T, typ, from int, more map[int]Placement) *Game {
	t.Helper()
	units := map[int]Placement{1: {Type: typ, Owner: 1, Cell: from}}
	for i, u := range more {
		units[i] = u
	}
	g := newCycle(t, Extended, units)
	for k := range g.Maps {
		for c := range g.Width * g.Height {
			g.Maps[k][2*c] = 0
		}
	}
	return g
}

// order gives unit 1 of side 1 its work order on cell with message msg.
func order(t *testing.T, g *Game, cell, msg int) {
	t.Helper()
	evs := apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: cell})
	if len(evs) != 1 || evs[0].Kind != EvWork || evs[0].Messages[0].Code != msg {
		t.Fatalf("order events %+v", evs)
	}
	if u := g.Units[1]; u.Flags4&0x200 == 0 || len(g.Players[1].Orders) != 0 {
		t.Fatalf("after the order: %+v, list %v", u, g.Players[1].Orders)
	}
}

func TestWire(t *testing.T) {
	g := workScene(t, cWiring, 40, map[int]Placement{0: {Type: cInfantry, Owner: 0, Cell: 31}})
	if g.Units[1].Hold != WireCharges {
		t.Fatalf("charges %d", g.Units[1].Hold)
	}
	if got := g.WireCells(1, 1, 31); !slices.Equal(got, []int{30, 31, 32}) {
		t.Errorf("wire cells %v", got)
	}
	g.Maps[1][2*32] = trenchFirst
	if got := g.WireCells(1, 1, 31); !slices.Equal(got, []int{30, 31}) {
		t.Errorf("wire cells beside a trench %v", got)
	}
	refusal(t, "a trench", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 32}), MsgNoTarget)
	refusal(t, "beyond the neighbours", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 42}), MsgNoTarget)
	order(t, g, 31, MsgWire)
	evs := resolve(t, g)
	w := kinds(evs, EvWired)
	if len(w) != 1 || w[0].Cell != 31 || !slices.Equal(w[0].Cells, []int{30, 31}) {
		t.Fatalf("wired %+v", w)
	}
	for k := range 2 {
		if !Wire(g.Maps[k][2*30]) || !Wire(g.Maps[k][2*31]) || !Trench(g.Maps[1][2*32]) {
			t.Errorf("map %d after the work: %d %d %d", k, g.Maps[k][2*30], g.Maps[k][2*31], g.Maps[k][2*32])
		}
	}
	if g.Units[1].Hold != WireCharges-1 || g.Units[0].Size != Extended.Tables.Types[cInfantry].Group {
		t.Errorf("charges %d, the infantry %+v", g.Units[1].Hold, g.Units[0])
	}
	resolve(t, g)
	refusal(t, "wire already", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 31}), MsgNoTarget)
	u := g.Units[1]
	u.Hold = 0
	g.Units[1] = u
	refusal(t, "no charge", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 39}), MsgNoWire)
	// A sapper clears wire as it fills a trench.
	if !g.worked(TerrainWire) || !g.worked(TerrainCrater) {
		t.Error("sappers leave wire or craters alone")
	}
}

// tunnelScene has side 1's tunnellers at 40 under side 0's infantry 0 at
// 31 (size size) and side 0's infantry 2 at 22.
func tunnelScene(t *testing.T, size int) *Game {
	t.Helper()
	return workScene(t, cTunnel, 40, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: 31, Size: size},
		2: {Type: cInfantry, Owner: 0, Cell: 22},
	})
}

func TestTunnel(t *testing.T) {
	g := tunnelScene(t, 0)
	full := Extended.Tables.Types[cInfantry].Group
	order(t, g, 31, MsgTunnel)
	evs := resolve(t, g) // side 1
	if tn := kinds(evs, EvTunnel); len(tn) != 1 || tn[0].Cell != 31 || len(g.Tunnels) != 1 {
		t.Fatalf("tunnel %+v, %+v", tn, g.Tunnels)
	}
	if len(kinds(evs, EvBattle)) != 0 {
		t.Error("the tunnellers fought")
	}
	resolve(t, g) // side 0: the first turn
	if g.Tunnels[0].Dug != 1 {
		t.Fatalf("dug %+v", g.Tunnels)
	}
	refusal(t, "busy", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 39}), MsgTunnelBusy)
	resolve(t, g)       // side 1
	evs = resolve(t, g) // side 0: the second turn blows it
	mine := kinds(evs, EvMine)
	if len(mine) != 1 || mine[0].Cell != 31 || !slices.Equal(mine[0].Cells, []int{31}) || g.Tunnels != nil {
		t.Fatalf("mine %+v, tunnels %+v", mine, g.Tunnels)
	}
	loss := (full*TunnelLoss + 99) / 100
	if m := kinds(evs, EvMined); len(m) != 1 || m[0].Unit != 0 || m[0].Loss != loss || m[0].Side != 0 {
		t.Errorf("mined %+v", m)
	}
	if g.Units[0].Size != full-loss || g.Units[2].Size != full {
		t.Errorf("sizes %d %d", g.Units[0].Size, g.Units[2].Size)
	}
	for k := range 2 {
		if g.Maps[k][2*31] != TerrainCrater {
			t.Errorf("map %d: record %d", k, g.Maps[k][2*31])
		}
	}
	if err := g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 39}); err != nil {
		t.Errorf("a new tunnel after the blast: %v", err)
	}
}

func TestTunnelDestroys(t *testing.T) {
	g := tunnelScene(t, 2)
	order(t, g, 31, MsgTunnel)
	for range 4 {
		resolve(t, g)
	}
	if u := g.Units[0]; !u.Destroyed() || g.occupant(0, 31) == 0 {
		t.Errorf("unit 0 survived the mine: %+v", u)
	}
}

func TestTunnelLost(t *testing.T) {
	g := tunnelScene(t, 0)
	order(t, g, 31, MsgTunnel)
	resolve(t, g)
	u := g.Units[1]
	u.Size = 0
	g.Units[1] = u
	evs := resolve(t, g)
	if l := kinds(evs, EvTunnelLost); len(l) != 1 || l[0].Unit != 1 || g.Tunnels != nil || len(kinds(evs, EvMine)) != 0 {
		t.Errorf("lost %+v, tunnels %+v", l, g.Tunnels)
	}
	refusal(t, "water", tunnelWater(t), MsgNoTarget)
}

// tunnelWater is Check of a tunnel under water.
func tunnelWater(t *testing.T) error {
	t.Helper()
	g := tunnelScene(t, 0)
	for k := range 2 {
		g.Maps[k][2*39] = 3 // water
	}
	return g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 39})
}

func TestAmbulance(t *testing.T) {
	g := workScene(t, cAmbulance, 40, map[int]Placement{
		0: {Type: cInfantry, Owner: 1, Cell: 31, Size: 3}, // next to it: heals
		2: {Type: cInfantry, Owner: 1, Cell: 22, Size: 3}, // too far
		3: {Type: cInfantry, Owner: 0, Cell: 39, Size: 3}, // the enemy
		4: {Type: cArmCar, Owner: 1, Cell: 41, Size: 3},   // no foot unit
		5: {Type: cInfantry, Owner: 1, Cell: 49},          // at full strength
	})
	if evs := resolve(t, g); len(kinds(evs, EvHealed)) != 0 {
		t.Errorf("side 1's resolution healed: %+v", evs)
	}
	evs := resolve(t, g)
	h := kinds(evs, EvHealed)
	if len(h) != 1 || h[0].Unit != 0 || h[0].Gain != AmbulanceHeal || h[0].Side != 1 || h[0].Cell != 31 {
		t.Fatalf("healed %+v", h)
	}
	for i, want := range map[int]int{0: 3 + AmbulanceHeal, 2: 3, 3: 3, 4: 3} {
		if g.Units[i].Size != want {
			t.Errorf("unit %d: size %d, want %d", i, g.Units[i].Size, want)
		}
	}
	if len(kinds(resolve(t, g), EvBattle)) != 0 {
		t.Error("the ambulance fought")
	}
}

func TestTunnelSave(t *testing.T) {
	g := tunnelScene(t, 0)
	order(t, g, 31, MsgTunnel)
	resolve(t, g)
	b, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	h, err := LoadGame(b)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(h.Tunnels, g.Tunnels) || len(h.Tunnels) != 1 {
		t.Errorf("loaded tunnels %+v", h.Tunnels)
	}
	if c := g.Clone(); !slices.Equal(c.Tunnels, g.Tunnels) {
		t.Errorf("cloned tunnels %+v", c.Tunnels)
	}
	var s savedGame
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	s.Tunnels[0].Cell = -1
	bad, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGame(bad); err == nil {
		t.Error("a tunnel off the map loads")
	}
}
