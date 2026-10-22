package dosrules

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

// The command cycle on small scenes: validation, a whole round of moves,
// attacks and the phase change, destroyed targets, exits, and the attack
// step against the recorded order lists.

const (
	cycleW, cycleH = 9, 9
	cInfantry      = 18
	cPanzer        = 47
	cSupply        = 22 // a ground transporter
	cTransportRec  = 3
)

// openTerrain is the cheapest terrain record without blocking or building
// flags that the scene's types may enter.
func openTerrain(t testing.TB, p Profile) int {
	t.Helper()
	m := moveTables(t, p)
	best := -1
	for i, tr := range m.Terrains {
		ok := tr.Flags&0x5547 == 0
		for _, ty := range []int{cInfantry, cPanzer, cSupply} {
			ok = ok && tr.Mask&m.Types[ty].Mask != 0
		}
		if ok && (best < 0 || tr.Cost < m.Terrains[best].Cost) {
			best = i
		}
	}
	if best < 0 {
		t.Fatal("no open terrain")
	}
	return best
}

func newCycle(t *testing.T, p Profile, units map[int]Placement) *Game {
	t.Helper()
	return newCycleHolds(t, p, units, nil)
}

// newCycleHolds is newCycle with container records.
func newCycleHolds(t *testing.T, p Profile, units map[int]Placement, holds map[int]Hold) *Game {
	t.Helper()
	tr := openTerrain(t, p)
	terrain := make([]int, cycleW*cycleH)
	for i := range terrain {
		terrain[i] = tr
	}
	g, err := NewGame(p, cycleW, cycleH, terrain, units, holds)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func apply(t *testing.T, g *Game, c Command) []Event {
	t.Helper()
	evs, err := g.Apply(c)
	if err != nil {
		t.Fatalf("%+v: %v", c, err)
	}
	return evs
}

func resolve(t *testing.T, g *Game) []Event {
	t.Helper()
	g.Confirm(0)
	g.Confirm(1)
	evs, err := g.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

// refusal checks that err is the cycle's refusal with the message code.
func refusal(t *testing.T, what string, err error, code int) {
	t.Helper()
	var r *Refusal
	if !errors.As(err, &r) {
		t.Errorf("%s: %v, want a refusal", what, err)
		return
	}
	if r.Code != code {
		t.Errorf("%s: message %#x, want %#x (%v)", what, r.Code, code, err)
	}
}

func TestCycleValidation(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			g := newCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cInfantry, Owner: 1, Cell: n[0]},
				2: {Type: cInfantry, Owner: 2, Cell: 10}, 3: {Type: cInfantry, Owner: 1, Cell: n[3]},
				4: {Type: cInfantry, Owner: 0, Cell: 70},
			})
			for _, c := range []struct {
				what string
				cmd  Command
				code int
			}{
				{"attack in the movement phase", Command{Kind: CmdAttack, Side: 0, Unit: 0, Cell: n[0]}, 0},
				{"move in the action phase", Command{Kind: CmdMove, Side: 1, Unit: 1, Cell: 0}, 0},
				{"the other side's unit", Command{Kind: CmdMove, Side: 0, Unit: 1, Cell: 0}, 0},
				{"a neutral unit", Command{Kind: CmdMove, Side: 0, Unit: 2, Cell: 0}, 0},
				{"no unit", Command{Kind: CmdMove, Side: 0, Unit: 9, Cell: 0}, 0},
				{"attack on an empty cell", Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 0}, MsgNoTarget},
				{"attack on an own unit", Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: n[3]}, MsgNoTarget},
				{"exit from an empty slot", Command{Kind: CmdExit, Side: 0, Hold: cTransportRec, Slot: 0, Cell: 0}, 0},
			} {
				refusal(t, c.what, g.Check(c.cmd), c.code)
			}

			attack := Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40}
			if err := g.Check(attack); err != nil {
				t.Fatalf("attack: %v", err)
			}
			evs := apply(t, g, attack)
			u := g.Units[1]
			if want := (Event{Kind: EvAttack, Side: 1, Unit: 1, Cell: 40, Messages: []Message{{MsgAttack, 1}}}); !reflect.DeepEqual(evs, []Event{want}) {
				t.Errorf("attack events %+v", evs)
			}
			if u.Flags4&0x200 == 0 || u.Target != 80 || u.Action[0] != 9 || u.Action[1] != 2 || !slices.Equal(g.Players[1].Orders, []int{1}) {
				t.Errorf("attack order: unit %+v, list %v", u, g.Players[1].Orders)
			}
			refusal(t, "a second order", g.Check(attack), 0)

			// The move limit counts the list; a computer side has none.
			g.Limit = 1
			second := Command{Kind: CmdAttack, Side: 1, Unit: 3, Cell: 40}
			refusal(t, "the move limit", g.Check(second), MsgLimit)
			g.Players[1].Computer = true
			if err := g.Check(second); err != nil {
				t.Errorf("computer side at the limit: %v", err)
			}
			g.Limit = 0

			// A confirmed side gives no orders until it cancels.
			move := Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: 41}
			g.Confirm(0)
			refusal(t, "a confirmed side", g.Check(move), 0)
			if _, err := g.Resolve(); !errors.Is(err, ErrNotReady) {
				t.Errorf("resolve with one side confirmed: %v", err)
			}
			if evs := g.Unconfirm(0); len(evs) != 1 || evs[0].Messages[0].Code != MsgReadyUndone {
				t.Errorf("unconfirm: %+v", evs)
			}
			if err := g.Check(move); err != nil {
				t.Errorf("move after the cancel: %v", err)
			}

			// A unit in a container is not on the map.
			l := g.Units[4]
			l.Flags4 |= 0x4000
			g.Units[4] = l
			g.Maps[0][2*70+1] = 0xFF
			refusal(t, "a held unit's move", g.Check(Command{Kind: CmdMove, Side: 0, Unit: 4, Cell: 71}), 0)
		})
	}
}

// TestCycleTargets checks the assumed target rule: an attack needs an
// opposing unit of a class the attacker's target flags cover, within the
// range of the weapon for that class.
func TestCycleTargets(t *testing.T) {
	const (
		cAircraft  = 2  // air range 1, attacks air, ground and water
		cArtillery = 24 // surface range 7, attacks ground and water
		cGround    = 21 // surface range 1, attacks ground only
		cBoat      = 29 // a water unit
	)
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			far := 40 - 2*cycleW // two rows up: distance 2
			g := newCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cAircraft, Owner: 0, Cell: n[3]},
				2: {Type: cInfantry, Owner: 0, Cell: 80}, 3: {Type: cBoat, Owner: 0, Cell: 70},
				10: {Type: cInfantry, Owner: 1, Cell: n[0]}, 11: {Type: cInfantry, Owner: 1, Cell: far},
				12: {Type: cArtillery, Owner: 1, Cell: 0}, 13: {Type: cAircraft, Owner: 1, Cell: n[2]},
				14: {Type: cGround, Owner: 1, Cell: 71}, 15: {Type: cInfantry, Owner: 1, Cell: 62},
			})
			if p.LayoutDistance(cycleW, 70, 71) != 1 || p.LayoutDistance(cycleW, 71, 80) != 1 || p.LayoutDistance(cycleW, 71, 62) != 1 {
				t.Fatal("cell 71 is not next to 62, 70 and 80")
			}
			if d := p.LayoutDistance(cycleW, 40, far); d != 2 {
				t.Fatalf("distance 40-%d is %d", far, d)
			}
			if d := p.LayoutDistance(cycleW, 0, 80); d <= 7 {
				t.Fatalf("distance 0-80 is %d, want beyond the artillery's range", d)
			}
			for _, c := range []struct {
				what string
				cmd  Command
				ok   bool
			}{
				{"infantry next to its target", Command{Kind: CmdAttack, Side: 1, Unit: 10, Cell: 40}, true},
				{"infantry two cells away", Command{Kind: CmdAttack, Side: 1, Unit: 11, Cell: 40}, false},
				{"infantry against an aircraft", Command{Kind: CmdAttack, Side: 1, Unit: 10, Cell: n[3]}, false},
				{"aircraft against an aircraft", Command{Kind: CmdAttack, Side: 1, Unit: 13, Cell: n[3]}, true},
				{"aircraft against infantry", Command{Kind: CmdAttack, Side: 1, Unit: 13, Cell: 40}, true},
				{"artillery in range", Command{Kind: CmdAttack, Side: 1, Unit: 12, Cell: 40}, true},
				{"artillery out of range", Command{Kind: CmdAttack, Side: 1, Unit: 12, Cell: 80}, false},
				{"artillery against an aircraft", Command{Kind: CmdAttack, Side: 1, Unit: 12, Cell: n[3]}, false},
				{"a ground-only unit against a boat in range", Command{Kind: CmdAttack, Side: 1, Unit: 14, Cell: 70}, false},
				{"a ground-only unit against infantry in range", Command{Kind: CmdAttack, Side: 1, Unit: 14, Cell: 80}, true},
				{"a unit of its own side next to it", Command{Kind: CmdAttack, Side: 1, Unit: 14, Cell: 62}, false},
			} {
				err := g.Check(c.cmd)
				if c.ok && err != nil {
					t.Errorf("%s: %v", c.what, err)
				}
				if !c.ok {
					refusal(t, c.what, err, MsgNoTarget)
				}
			}
		})
	}
}

// TestCycleRound plays a round: side 0 moves its unit away while side 1
// attacks it where side 1's map still shows it; both confirm, the attack is
// fought at the old cell, and the phase change brings the maps into line.
func TestCycleRound(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			g := newCycle(t, p, map[int]Placement{
				0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cPanzer, Owner: 1, Cell: n[0]},
			})
			dests, err := p.Destinations(g.Movement(0, 0))
			if err != nil {
				t.Fatal(err)
			}
			dest := -1
			for _, c := range dests {
				if !slices.Contains(n[:], c) && c != 40 {
					dest = c
					break
				}
			}
			if dest < 0 {
				t.Fatalf("no destination away from cell 40 in %v", dests)
			}
			evs := apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: dest})
			if len(evs) != 1 || evs[0].Kind != EvMoved || evs[0].Route[len(evs[0].Route)-1] != dest {
				t.Fatalf("move events %+v", evs)
			}
			if g.occupant(0, dest) != 0 || g.occupant(0, 40) != 0xFF || g.occupant(1, 40) != 0 || g.occupant(1, dest) != 0xFF {
				t.Errorf("after the move: side 0 sees %d at %d, side 1 sees %d at 40", g.occupant(0, dest), dest, g.occupant(1, 40))
			}
			if !slices.Equal(g.Players[0].Orders, []int{0}) || g.Units[0].Flags4&0x200 == 0 {
				t.Errorf("after the move: list %v, flags %#x", g.Players[0].Orders, g.Units[0].Flags4)
			}
			refusal(t, "a second move", g.Check(Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: 40}), 0)

			apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40})
			evs = resolve(t, g)
			if len(evs) != 2 || evs[0].Kind != EvBattle || evs[0].Outcome.Defender != 0 || !evs[0].Outcome.ReturnFire || evs[1].Kind != EvPhase {
				t.Fatalf("resolution events %+v", evs)
			}
			o := evs[0].Outcome
			if g.Units[0].Size != o.DefenderSize || g.Units[1].Size != o.AttackerSize {
				t.Errorf("strengths %d/%d, outcome %+v", g.Units[1].Size, g.Units[0].Size, o)
			}
			if evs[1].Phase != [2]int{PhaseAction, PhaseMovement} || evs[1].Turn != 0 {
				t.Errorf("phase change %+v", evs[1])
			}
			if !slices.Equal(g.Maps[0], g.Maps[1]) {
				t.Error("the maps differ after the phase change")
			}
			if o.DefenderSize > 0 && (g.occupant(1, dest) != 0 || g.Units[0].Cell(1) != dest) {
				t.Errorf("side 1 does not see unit 0 at %d: map %d, position %d", dest, g.occupant(1, dest), g.Units[0].Cell(1))
			}
			for i, u := range g.Units {
				if u.Flags4&0x200 != 0 {
					t.Errorf("unit %d keeps its moved flag", i)
				}
			}
			for k, pl := range g.Players {
				if pl.Ready || len(pl.Orders) != 0 {
					t.Errorf("player %d after the change: %+v", k, pl)
				}
			}

			// The next change brings side 0 back to the movement phase and
			// counts the turn.
			if evs := resolve(t, g); evs[len(evs)-1].Phase != [2]int{PhaseMovement, PhaseAction} || g.Turn != 1 {
				t.Errorf("second change %+v, turn %d", evs, g.Turn)
			}
		})
	}
}

// TestCycleDestroyedTarget gives two attacks on one target: where the first
// destroys it, the second is skipped, and the phase change marks it.
func TestCycleDestroyedTarget(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			skipped := 0
			for size := 1; size <= 6; size++ {
				g := newCycle(t, p, map[int]Placement{
					0: {Type: cInfantry, Owner: 0, Cell: 40, Size: size},
					1: {Type: cPanzer, Owner: 1, Cell: n[0]}, 2: {Type: cPanzer, Owner: 1, Cell: n[3]},
				})
				apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40})
				apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 2, Cell: 40})
				evs := resolve(t, g)
				if evs[0].Kind != EvBattle || evs[2].Kind != EvPhase {
					t.Fatalf("size %d: %+v", size, evs)
				}
				if !slices.Contains(evs[0].Outcome.Destroyed, 0) {
					if evs[1].Kind != EvBattle {
						t.Errorf("size %d: the surviving target is not fought again: %+v", size, evs[1])
					}
					continue
				}
				skipped++
				if evs[1].Kind != EvSkipped {
					t.Errorf("size %d: the second order on a destroyed target: %+v", size, evs[1])
				}
				// The phase change marks it 0x8000 and clears the action
				// bytes, the battle's bit 4 of +0x14 too.
				u := g.Units[0]
				if u.Flags4&0x8000 == 0 || u.Action != [7]int{} || g.occupant(0, 40) != 0xFF || g.occupant(1, 40) != 0xFF {
					t.Errorf("size %d: destroyed unit %+v, maps %d/%d", size, u, g.occupant(0, 40), g.occupant(1, 40))
				}
				if !slices.Equal(g.Destroyed(), []int{0}) {
					t.Errorf("size %d: destroyed %v", size, g.Destroyed())
				}
			}
			if skipped == 0 {
				t.Error("no first attack destroyed the target")
			}
		})
	}
}

// transportHold is the container record of transporter unit at cell of
// side, taking every class, with the units of cargo in the side's slots.
func transportHold(cell, side, unit int, cargo ...int) Hold {
	h := EmptyHold()
	h.Cells[0], h.Flags, h.Occupant = 2*cell+1, 0x20|0xFC0|side, unit
	for k, u := range cargo {
		h.Slots[7*side+k] = u
	}
	return h
}

// TestCycleExit takes a unit out of a transporter during the movement phase.
func TestCycleExit(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, cycleW, cycleH)
			g := newCycleHolds(t, p, map[int]Placement{
				1: {Type: cSupply, Owner: 0, Cell: 40}, 10: {Type: cInfantry, Owner: 0},
			}, map[int]Hold{cTransportRec: transportHold(40, 0, 1, 10)})
			refusal(t, "a held unit's move", g.Check(Command{Kind: CmdMove, Side: 0, Unit: 10, Cell: n[0]}), 0)

			evs := apply(t, g, Command{Kind: CmdExit, Side: 0, Hold: cTransportRec, Slot: 0, Cell: n[0]})
			if len(evs) != 1 || evs[0].Kind != EvMoved || evs[0].Unit != 10 {
				t.Fatalf("exit events %+v", evs)
			}
			u := g.Units[10]
			if g.Holds[cTransportRec].Slots[0] != 0xFF || g.occupant(0, n[0]) != 10 || g.occupant(0, 40) != 1 ||
				u.Flags4&0x4000 != 0 || u.Flags4&0x200 == 0 || !slices.Equal(g.Players[0].Orders, []int{10}) {
				t.Errorf("after the exit: slot %d, map %d/%d, flags %#x, list %v", g.Holds[cTransportRec].Slots[0],
					g.occupant(0, n[0]), g.occupant(0, 40), u.Flags4, g.Players[0].Orders)
			}
			resolve(t, g)
			if g.occupant(1, n[0]) != 10 || g.occupant(1, 40) != 1 {
				t.Errorf("side 1 after the change: %d at %d, %d at 40", g.occupant(1, n[0]), n[0], g.occupant(1, 40))
			}
		})
	}
}

// TestCycleOrdersFixtures runs the cycle's attack step on every recorded
// order list (orders.json): the terrain records stand for their combat
// class, as in the recorded scenes, and the outcomes, strengths, destroyed
// units and the attacking side's map must be the recorded ones.
func TestCycleOrdersFixtures(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			tables := *p.Tables
			mt := *tables.Move
			mt.Terrains = make([]MoveTerrain, 9)
			for k := range mt.Terrains {
				mt.Terrains[k].Class = k
			}
			tables.Move = &mt
			p.Tables = &tables
			for _, f := range load[OrdersFixture](t, build, "orders.json") {
				g := ordersGame(p, &f)
				evs, err := g.attack(f.Side)
				if err != nil {
					t.Fatalf("%s: %v", f.ID, err)
				}
				got := OrdersExpected{Orders: []EngagementExpected{}, Units: []FixtureUnit{}, Gone: []int{}, Occupants: [][2]int{}}
				for _, ev := range evs {
					o := Outcome{}
					if ev.Outcome != nil {
						o = *ev.Outcome
					}
					got.Orders = append(got.Orders, EngagementResult(o, nil))
				}
				for _, fu := range f.Units {
					u := g.Units[fu.Index]
					fu.Size, fu.AlternateSize, fu.Experience = u.Size, u.Alt, u.Exp
					got.Units = append(got.Units, fu)
					if u.Action[1]&4 != 0 {
						got.Gone = append(got.Gone, fu.Index)
					}
				}
				for c := range f.Width * f.Height {
					if occ := g.occupant(f.Side, c); occ != 0xFF {
						got.Occupants = append(got.Occupants, [2]int{c, occ})
					}
				}
				if err := fieldDiff(got, f.Expected); err != nil {
					t.Errorf("%s: %v", f.ID, err)
				}
			}
		})
	}
}

// ordersGame is the state an order-list fixture gives the handler: the
// attacking side's map with the units and terrain classes, its list, and
// each attacker's target in +0x11.
func ordersGame(p Profile, f *OrdersFixture) *Game {
	b := f.Battlefield()
	g := &Game{Profile: p, Width: f.Width, Height: f.Height, Units: map[int]Live{}, Holds: map[int]Hold{}}
	for k := range 2 {
		g.Maps[k] = make([]int, 2*f.Width*f.Height)
		for c := range f.Width * f.Height {
			g.Maps[k][2*c], g.Maps[k][2*c+1] = b.class(c), 0xFF
		}
	}
	for i, u := range b.Units {
		l := Live{Size: u.Size, Alt: u.Alt, Exp: u.Exp, Flags4: u.Flags4, Flags6: u.Flags6, Type: u.Type, Hold: 0xFF}
		l.Pos[f.Side] = 2*u.Cell + 1
		g.Units[i] = l
		g.Maps[f.Side][2*u.Cell+1] = i
	}
	for _, o := range f.List() {
		l := g.Units[o.Attacker]
		l.Target = 2 * o.Target
		g.Units[o.Attacker] = l
		g.Players[f.Side].Orders = append(g.Players[f.Side].Orders, o.Attacker)
	}
	return g
}

// BenchmarkDestinations is the destination query of a fast unit on an open
// 40×30 map with units around it.
func BenchmarkDestinations(b *testing.B) {
	const w, h = 40, 30
	tr := openTerrain(b, Full)
	terrain := make([]int, w*h)
	for i := range terrain {
		terrain[i] = tr
	}
	units := map[int]Placement{0: {Type: cPanzer, Owner: 0, Cell: 15*w + 20}}
	for i := 1; i < 20; i++ {
		units[i] = Placement{Type: cInfantry, Owner: i % 2, Cell: (3 + i) * w / 2 * 3 % (w * h)}
	}
	g, err := NewGame(Full, w, h, terrain, units, nil)
	if err != nil {
		b.Fatal(err)
	}
	m := g.Movement(0, 0)
	b.ResetTimer()
	for b.Loop() {
		if _, err := Full.Destinations(m); err != nil {
			b.Fatal(err)
		}
	}
}

// A two-cell ship's rear takes no orders: it moves and fights with its head.
func TestCycleShipRear(t *testing.T) {
	const cShip = 32 // a head; its rear is type 33
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := newCycle(t, p, map[int]Placement{
				0: {Type: cShip, Owner: 0, Cell: 20}, 1: {Type: cShip + 1, Owner: 0, Cell: 11},
				2: {Type: cShip, Owner: 1, Cell: 60}, 3: {Type: cShip + 1, Owner: 1, Cell: 69},
			})
			refusal(t, "the rear's move", g.Check(Command{Kind: CmdMove, Side: 0, Unit: 1, Cell: 12}), 0)
			refusal(t, "the rear's attack", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 3, Cell: 20}), 0)
			if err := g.Check(Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: 21}); err != nil {
				t.Errorf("the head's move: %v", err)
			}
			// The head passes the rear check; the empty cell holds no target.
			refusal(t, "the head's attack", g.Check(Command{Kind: CmdAttack, Side: 1, Unit: 2, Cell: 0}), MsgNoTarget)
		})
	}
}
