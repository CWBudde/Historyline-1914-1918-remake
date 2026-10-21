package dosplay

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// shipped are the shipped levels in file order.
func shipped(t *testing.T) []*level.Level {
	t.Helper()
	names, err := fs.Glob(levels.FS, "level?_cp.json")
	if err != nil || len(names) == 0 {
		t.Fatalf("no shipped levels: %v", err)
	}
	var out []*level.Level
	for _, n := range names {
		raw, err := levels.FS.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		lv, err := level.Parse(raw)
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		out = append(out, lv)
	}
	return out
}

func dataSet(t *testing.T) *data.Set {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func session(t *testing.T, lv *level.Level, o Options) *Session {
	t.Helper()
	s, err := New(lv, dataSet(t), o)
	if err != nil {
		t.Fatalf("%s: New: %v", lv.MapName, err)
	}
	checkBound(t, s)
	return s
}

// checkBound is the invariant every command keeps: each side's map shows
// live units on the level's hexes only, its View round-trips to those
// cells, and every held unit sits in a container on a hex of the level.
func checkBound(t *testing.T, s *Session) {
	t.Helper()
	g := s.Game
	for k := range 2 {
		m := g.Maps[k]
		n := 0
		for c := range g.Width * g.Height {
			occ := m[2*c+1]
			if occ > dosrules.SlotFree {
				continue
			}
			n++
			if _, ok := g.Units[occ]; !ok {
				t.Fatalf("%s: side %d's map shows unit %d, which has no record", s.Level.MapName, k, occ)
			}
			if _, ok := s.Board.Coord(c); !ok {
				t.Fatalf("%s: side %d's map shows unit %d on unbound cell %d", s.Level.MapName, k, occ, c)
			}
		}
		v := s.View(k)
		if len(v.Units) != n {
			t.Fatalf("%s: View(%d) shows %d units, the map %d", s.Level.MapName, k, len(v.Units), n)
		}
		for p, u := range v.Units {
			cell, ok := s.Board.Cell(p)
			if !ok || u.At != p || m[2*cell+1] != u.Index {
				t.Fatalf("%s: View(%d) unit %d at %v (cell %d, map %d)", s.Level.MapName, k, u.Index, p, cell, m[2*cell+1])
			}
		}
	}
	for i, u := range g.Units {
		if u.Destroyed() || u.Flags4&0x4000 == 0 {
			continue
		}
		side := sideOf(u.Flags4) & 1
		h, _, ok := s.container(side, i)
		if !ok {
			t.Fatalf("%s: held unit %d is in no container record of side %d", s.Level.MapName, i, side)
		}
		if _, ok := s.Board.Coord(startCell(g.Holds[h])); !ok {
			t.Fatalf("%s: held unit %d is in record %d off the level", s.Level.MapName, i, h)
		}
	}
}

// own are the live units of side, ascending.
func own(s *Session, side int) []int {
	var out []int
	for _, i := range slices.Sorted(maps.Keys(s.Game.Units)) {
		u := s.Game.Units[i]
		if !u.Destroyed() && sideOf(u.Flags4) == side {
			out = append(out, i)
		}
	}
	return out
}

// policy picks a destination for a move from the sorted choices.
type policy func(s *Session, side, unit int, choices []hexgrid.Coord) hexgrid.Coord

func first(_ *Session, _, _ int, choices []hexgrid.Coord) hexgrid.Coord { return choices[0] }

// approach picks the choice nearest to an enemy unit the side's map shows.
func approach(s *Session, side, _ int, choices []hexgrid.Coord) hexgrid.Coord {
	var enemies []hexgrid.Coord
	for _, p := range Coords(toSet(s.View(side).Units)) {
		if u := s.View(side).Units[p]; u.Side != side && u.Side != 2 {
			enemies = append(enemies, p)
		}
	}
	best, bd := choices[0], 1<<30
	for _, c := range choices {
		for _, e := range enemies {
			if d := hexgrid.Distance(c, e); d < bd {
				best, bd = c, d
			}
		}
	}
	return best
}

func toSet[V any](m map[hexgrid.Coord]V) map[hexgrid.Coord]bool {
	out := map[hexgrid.Coord]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}

// cycle plays one command cycle: the movement side moves every unit it can
// (exits included) by pick, the action side attacks the first target of
// every unit that has one, both confirm and the orders resolve. It returns
// all events.
func cycle(t *testing.T, s *Session, pick policy) []Event {
	t.Helper()
	a, err := s.Game.Action()
	if err != nil {
		t.Fatal(err)
	}
	mv := 1 - a
	var evs []Event
	for _, i := range own(s, mv) {
		if s.CanOrder(mv, i) != nil {
			continue
		}
		u := s.Unit(i)
		var dests map[hexgrid.Coord]bool
		if u.Held {
			if dests, err = s.ExitDestinations(mv, u.In, u.Slot); err != nil {
				t.Fatalf("%s: ExitDestinations(%d, %d, %d): %v", s.Level.MapName, mv, u.In, u.Slot, err)
			}
		} else if dests, err = s.Destinations(mv, i); err != nil {
			t.Fatalf("%s: Destinations(%d, %d): %v", s.Level.MapName, mv, i, err)
		}
		if len(dests) == 0 {
			continue
		}
		to := pick(s, mv, i, Coords(dests))
		var ev []Event
		if u.Held {
			ev, err = s.Exit(mv, u.In, u.Slot, to)
		} else {
			ev, err = s.Move(mv, i, to)
		}
		if err != nil {
			t.Fatalf("%s: move of unit %d to %v: %v", s.Level.MapName, i, to, err)
		}
		checkBound(t, s)
		evs = append(evs, ev...)
	}
	for _, i := range own(s, a) {
		ts := Coords(s.Targets(a, i))
		if len(ts) == 0 {
			continue
		}
		ev, err := s.Attack(a, i, ts[0])
		if err != nil {
			t.Fatalf("%s: attack of unit %d on %v: %v", s.Level.MapName, i, ts[0], err)
		}
		checkBound(t, s)
		evs = append(evs, ev...)
	}
	evs = append(evs, s.Confirm(0)...)
	evs = append(evs, s.Confirm(1)...)
	ev, err := s.Resolve()
	if err != nil {
		t.Fatalf("%s: Resolve: %v", s.Level.MapName, err)
	}
	checkBound(t, s)
	return append(evs, ev...)
}

func TestNewViews(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{})
		if s.Phase(0) != dosrules.PhaseMovement || s.Phase(1) != dosrules.PhaseAction || s.Ended() != nil {
			t.Errorf("%s: phases %d/%d", lv.MapName, s.Phase(0), s.Phase(1))
		}
		for k := range 2 {
			v := s.View(k)
			for l, line := range lv.Lines {
				for c, h := range line {
					p := hexgrid.Coord{Line: l, Col: c}
					u, ok := v.Units[p]
					if h.Unit == "" {
						if ok && !u.Carrier {
							t.Errorf("%s: View(%d) shows %s at empty %v", lv.MapName, k, u.ID, p)
						}
						continue
					}
					if !ok || u.ID != h.Unit || u.Owner != s.Faction(u.Side) || u.Name == "" || u.Size != u.Full {
						t.Errorf("%s: View(%d) at %v: %+v, want %s", lv.MapName, k, p, u, h.Unit)
					}
				}
			}
			for p, b := range v.Buildings {
				h := lv.Lines[p.Line][p.Col]
				if h.Content == nil {
					continue
				}
				if b.Owner != s.Faction(sideOf(s.Game.Holds[b.Hold].Flags)) || int(b.Owner) != h.Content.Owner && h.Content.Owner != 0 {
					t.Errorf("%s: building at %v owned by %v, level %d", lv.MapName, p, b.Owner, h.Content.Owner)
				}
				var ids []string
				for _, u := range b.Units {
					ids = append(ids, u.ID)
					if !u.Held || u.At != p || u.In != b.Hold {
						t.Errorf("%s: building at %v: %+v", lv.MapName, p, u)
					}
				}
				if !slices.Equal(ids, h.Content.Units) {
					t.Errorf("%s: building at %v holds %v, level %v", lv.MapName, p, ids, h.Content.Units)
				}
			}
		}
		for i := range s.Game.Units {
			u := s.Unit(i)
			if cell, ok := s.Board.Cell(u.At); !ok || s.coord(cell) != u.At {
				t.Errorf("%s: unit %d at %v", lv.MapName, i, u.At)
			}
		}
	}
}

func TestMoveAndResolve(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{})
		var unit int
		var dests []hexgrid.Coord
		for _, i := range own(s, 0) {
			if s.CanOrder(0, i) != nil || s.Unit(i).Held {
				continue
			}
			d, err := s.Destinations(0, i)
			if err != nil {
				t.Fatal(err)
			}
			if len(d) > 0 {
				unit, dests = i, Coords(d)
				break
			}
		}
		if dests == nil {
			t.Fatalf("%s: no unit of side 0 can move", lv.MapName)
		}
		from := s.Unit(unit).At
		to := dests[len(dests)-1]
		route, err := s.Route(0, unit, to)
		if err != nil || len(route) < 2 || route[0] != from || route[len(route)-1] != to {
			t.Fatalf("%s: Route(%d, %v) = %v, %v", lv.MapName, unit, to, route, err)
		}
		v0 := s.Version()
		evs, err := s.Move(0, unit, to)
		if err != nil {
			t.Fatalf("%s: Move: %v", lv.MapName, err)
		}
		checkBound(t, s)
		if len(evs) != 1 || evs[0].Kind != Moved || evs[0].Path[0] != from || evs[0].Path[len(evs[0].Path)-1] != to || s.Version() == v0 {
			t.Fatalf("%s: Move events %+v", lv.MapName, evs)
		}
		if !s.Unit(unit).Moved {
			t.Errorf("%s: unit %d not marked moved", lv.MapName, unit)
		}
		var oe *OrderError
		var r *dosrules.Refusal
		if err := s.CanOrder(0, unit); !errors.As(err, &oe) || !errors.As(err, &r) {
			t.Errorf("%s: CanOrder after the move: %v", lv.MapName, err)
		}
		// Side 1 still sees the unit where it stood.
		if u, ok := s.View(1).Units[from]; !s.Unit(unit).Held && (!ok || u.Index != unit) {
			t.Errorf("%s: side 1 does not see unit %d at %v", lv.MapName, unit, from)
		}
		if _, err := s.Resolve(); !errors.Is(err, dosrules.ErrNotReady) {
			t.Errorf("%s: Resolve before confirmation: %v", lv.MapName, err)
		}
		s.Confirm(0)
		if ev := s.Unconfirm(0); len(ev) != 1 || ev[0].Kind != Unconfirmed || ev[0].Text == "" {
			t.Errorf("%s: Unconfirm: %+v", lv.MapName, ev)
		}
		s.Confirm(0)
		s.Confirm(1)
		evs, err = s.Resolve()
		if err != nil {
			t.Fatalf("%s: Resolve: %v", lv.MapName, err)
		}
		checkBound(t, s)
		var phase, revealed int
		for i, ev := range evs {
			switch ev.Kind {
			case PhaseChanged:
				phase = i + 1
			case Revealed:
				if phase == 0 || ev.Unit != unit || !slices.Equal(ev.Path, route) || ev.Side != 0 {
					t.Errorf("%s: Revealed %+v, route %v", lv.MapName, ev, route)
				}
				revealed++
			}
		}
		if phase == 0 || revealed != 1 || s.Phase(0) != dosrules.PhaseAction || s.Phase(1) != dosrules.PhaseMovement {
			t.Errorf("%s: resolution %+v, phases %d/%d", lv.MapName, evs, s.Phase(0), s.Phase(1))
		}
		if !reflect.DeepEqual(s.View(0).Units, s.View(1).Units) {
			t.Errorf("%s: the maps differ after the phase change", lv.MapName)
		}
	}
}

func TestSmoke(t *testing.T) {
	for _, pick := range []struct {
		name string
		p    policy
	}{{"first", first}, {"approach", approach}} {
		for _, lv := range shipped(t) {
			s := session(t, lv, Options{})
			kinds := map[EventKind]int{}
			n := 0
			for ; n < 40 && s.Ended() == nil; n++ {
				for _, ev := range cycle(t, s, pick.p) {
					kinds[ev.Kind]++
				}
			}
			end := "open"
			if e := s.Ended(); e != nil {
				end = fmt.Sprintf("side %d won (%v), score %d", e.Winner, e.Won, e.Score)
			}
			t.Logf("%s/%s: %d resolutions, turn %d, %d moves, %d refused, %d attack orders, %d battles, %d skipped, %d revealed, %d/%d units left, %s",
				pick.name, lv.MapName, n, s.Turn(), kinds[Moved], kinds[Refused], kinds[Ordered], kinds[Battle], kinds[Skipped], kinds[Revealed],
				len(own(s, 0)), len(own(s, 1)), end)
			if kinds[Refused] > 0 {
				t.Errorf("%s/%s: %d moves to a destination were refused", pick.name, lv.MapName, kinds[Refused])
			}
		}
	}
}

// battleChecked is cycle with the battles checked against the state: each
// fighter's Before is what the previous battle left (or the size before the
// resolution), and the last After of every unit is its size afterwards.
func battleChecked(t *testing.T, s *Session, pick policy) []Event {
	t.Helper()
	pre := map[int]int{}
	for i, u := range s.Game.Units {
		pre[i] = u.Size
	}
	evs := cycle(t, s, pick)
	running := maps.Clone(pre)
	fought := map[int]bool{}
	for _, ev := range evs {
		if ev.Kind != Battle {
			continue
		}
		for _, f := range []Fighter{ev.Battle.Attacker, ev.Battle.Defender} {
			i := f.Unit.Index
			if f.Before != running[i] || f.Unit.Size != f.Before || f.Destroyed != s.Game.Units[i].Destroyed() && f.Destroyed {
				t.Fatalf("%s: battle %+v: unit %d before %d, running %d", s.Level.MapName, ev.Battle, i, f.Before, running[i])
			}
			running[i], fought[i] = f.After, true
		}
		if ev.From != ev.Battle.Attacker.Unit.At || ev.To != ev.Battle.Defender.Unit.At || ev.Battle.Attacker.Unit.Side != ev.Side {
			t.Fatalf("%s: battle event %+v", s.Level.MapName, ev)
		}
	}
	for i := range fought {
		if got := s.Game.Units[i].Size; got != running[i] {
			t.Fatalf("%s: unit %d has size %d after the battles, the last After says %d", s.Level.MapName, i, got, running[i])
		}
		if s.Game.Units[i].Destroyed() {
			for k := range 2 {
				if v, ok := s.View(k).Units[s.Unit(i).At]; ok && v.Index == i {
					t.Fatalf("%s: destroyed unit %d still shown to side %d", s.Level.MapName, i, k)
				}
			}
		}
	}
	return evs
}

func TestBattles(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{})
		battles, destroyed := 0, 0
		for n := 0; n < 40 && s.Ended() == nil; n++ {
			for _, ev := range battleChecked(t, s, approach) {
				if ev.Kind == Battle {
					battles++
					if ev.Battle.Defender.Destroyed {
						destroyed++
					}
				}
			}
		}
		if battles == 0 || destroyed == 0 {
			t.Errorf("%s: %d battles, %d units destroyed", lv.MapName, battles, destroyed)
		}
	}
}

func TestAttackOrders(t *testing.T) {
	s := session(t, shipped(t)[0], Options{})
	for n := 0; n < 20; n++ {
		a, _ := s.Game.Action()
		for _, i := range own(s, a) {
			ts := Coords(s.Targets(a, i))
			if len(ts) == 0 {
				continue
			}
			from := s.Unit(i).At
			evs, err := s.Attack(a, i, ts[0])
			if err != nil || len(evs) != 1 || evs[0].Kind != Ordered || evs[0].From != from || evs[0].To != ts[0] || evs[0].Text == "" {
				t.Fatalf("Attack: %+v, %v", evs, err)
			}
			if os := s.Orders(a); len(os) != 1 || os[0].Unit != i || os[0].To != ts[0] {
				t.Errorf("Orders(%d) = %+v", a, os)
			}
			if u := s.View(a).Units[from]; u.Target == nil || *u.Target != ts[0] || !u.Moved {
				t.Errorf("View(%d) of the attacker: %+v", a, u)
			}
			if u := s.View(1 - a).Units[from]; u.Target != nil {
				t.Errorf("View(%d) shows the other side's target", 1-a)
			}
			if _, err := s.Attack(a, i, ts[0]); err == nil {
				t.Errorf("a second attack order of unit %d was taken", i)
			}
			if _, err := s.Move(a, i, ts[0]); err == nil {
				t.Errorf("a move in the action phase was taken")
			}
			return
		}
		cycle(t, s, approach)
	}
	t.Fatal("no unit got a target")
}

func TestExit(t *testing.T) {
	var lv *level.Level
	for _, l := range shipped(t) {
		if l.MapName == "Civil" {
			lv = l
		}
	}
	s := session(t, lv, Options{})
	for _, p := range Coords(toSet(s.View(0).Buildings)) {
		hold, slots, ok := s.Contents(0, p)
		if !ok || len(slots) == 0 {
			continue
		}
		u := slots[0]
		if u.In != hold || u.At != p || !u.Held || s.CanOrder(0, u.Index) != nil {
			t.Fatalf("Contents(0, %v): %+v", p, u)
		}
		dests, err := s.ExitDestinations(0, hold, u.Slot)
		if err != nil || len(dests) == 0 || dests[p] {
			t.Fatalf("ExitDestinations(0, %d, %d) = %v, %v", hold, u.Slot, dests, err)
		}
		to := Coords(dests)[0]
		evs, err := s.Exit(0, hold, u.Slot, to)
		if err != nil || len(evs) != 1 || evs[0].Kind != Moved || evs[0].Path[0] != p || evs[0].Path[len(evs[0].Path)-1] != to {
			t.Fatalf("Exit: %+v, %v", evs, err)
		}
		checkBound(t, s)
		if v := s.View(0).Units[to]; v.Index != u.Index || v.Held || !v.Moved {
			t.Errorf("after the exit side 0 sees %+v at %v", v, to)
		}
		if _, after, _ := s.Contents(0, p); slices.ContainsFunc(after, func(v UnitView) bool { return v.Index == u.Index }) {
			t.Errorf("the building still lists unit %d", u.Index)
		}
		if _, ok := s.View(1).Units[to]; ok {
			t.Errorf("side 1 sees the exit before the phase change")
		}
		return
	}
	t.Fatal("no building of side 0 holds a unit")
}

// script plays n cycles and returns every event.
func script(t *testing.T, lv *level.Level, n int) ([]Event, *Session) {
	s := session(t, lv, Options{Limit: 8})
	var evs []Event
	for range n {
		if s.Ended() != nil {
			break
		}
		evs = append(evs, cycle(t, s, approach)...)
	}
	return evs, s
}

func TestDeterminism(t *testing.T) {
	for _, lv := range shipped(t) {
		a, _ := script(t, lv, 8)
		b, _ := script(t, lv, 8)
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s: two equal command sequences gave different events", lv.MapName)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	for _, lv := range shipped(t) {
		_, s := script(t, lv, 3)
		if s.Ended() != nil {
			t.Fatalf("%s: ended", lv.MapName)
		}
		// A move and an exit given, not confirmed.
		mv := 1 - must(s.Game.Action())
		for _, i := range own(s, mv) {
			if d, err := s.Destinations(mv, i); err == nil && len(d) > 0 {
				if _, err := s.Move(mv, i, Coords(d)[0]); err != nil {
					t.Fatal(err)
				}
				break
			}
		}
		held := 0
		for _, u := range s.Game.Units {
			if u.Flags4&0x4000 != 0 && !u.Destroyed() {
				held++
			}
		}
		if held == 0 {
			t.Errorf("%s: no unit in a container", lv.MapName)
		}
		raw, err := s.Save()
		if err != nil {
			t.Fatal(err)
		}
		g, err := dosrules.LoadGame(raw)
		if err != nil {
			t.Fatalf("%s: LoadGame: %v", lv.MapName, err)
		}
		again, err := g.Save()
		if err != nil || string(again) != string(raw) {
			t.Errorf("%s: the loaded game saves differently (%v)", lv.MapName, err)
		}
		r, err := Load(lv, dataSet(t), raw)
		if err != nil {
			t.Fatalf("%s: Load: %v", lv.MapName, err)
		}
		checkBound(t, r)
		for k := range 2 {
			if !reflect.DeepEqual(r.View(k), s.View(k)) || r.OrdersLeft(k) != s.OrdersLeft(k) {
				t.Errorf("%s: side %d sees another battle after Load", lv.MapName, k)
			}
		}
		// The loaded session has no routes from before the save to reveal.
		revealed := func(e Event) bool { return e.Kind == Revealed }
		c1 := slices.DeleteFunc(cycle(t, s, approach), revealed)
		c2 := slices.DeleteFunc(cycle(t, r, approach), revealed)
		if !reflect.DeepEqual(c1, c2) {
			t.Errorf("%s: the loaded battle plays on differently", lv.MapName)
		}
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestLimitAndRefusals(t *testing.T) {
	s := session(t, shipped(t)[0], Options{Limit: 4})
	if s.OrdersLeft(0) != 4 {
		t.Fatalf("OrdersLeft = %d", s.OrdersLeft(0))
	}
	moved := 0
	for _, i := range own(s, 0) {
		d, err := s.Destinations(0, i)
		if err != nil || len(d) == 0 {
			continue
		}
		_, err = s.Move(0, i, Coords(d)[0])
		if moved == 4 {
			var oe *OrderError
			if !errors.As(err, &oe) || oe.Refusal.Code != dosrules.MsgLimit || oe.Text != MessageText(dosrules.MsgLimit) {
				t.Errorf("the fifth move: %v", err)
			}
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		moved++
	}
	if s.OrdersLeft(0) != 0 || s.OrdersLeft(1) != 4 {
		t.Errorf("OrdersLeft = %d/%d", s.OrdersLeft(0), s.OrdersLeft(1))
	}
	if _, err := s.Move(0, 0, hexgrid.Coord{Line: 999}); err == nil {
		t.Error("a move off the level was taken")
	}
	for _, code := range []int{dosrules.MsgNoTarget, dosrules.MsgAttack, dosrules.MsgLimit, dosrules.MsgReadyUndone,
		msgEnemyThere, msgArrived, msgFriendThere, msgNoCarry, msgFull, msgTerrain, msgNoRoute, msgMovementTurn, msgActionTurn} {
		if MessageText(code) == "" {
			t.Errorf("no text for message %#x", code)
		}
	}
}

func TestEnterBuilding(t *testing.T) {
	s := session(t, shipped(t)[0], Options{})
	for _, i := range own(s, 0) {
		d, err := s.Destinations(0, i)
		if err != nil {
			continue
		}
		for _, p := range Coords(d) {
			if b, ok := s.View(0).Buildings[p]; !ok || b.Owner != s.Faction(0) {
				continue
			}
			from := s.Unit(i).At
			evs, err := s.Move(0, i, p)
			if err != nil || len(evs) != 1 || evs[0].Kind != Moved || evs[0].Path[0] != from || evs[0].Path[len(evs[0].Path)-1] != p {
				t.Fatalf("Move into the building at %v: %+v, %v", p, evs, err)
			}
			checkBound(t, s)
			u := s.Unit(i)
			hold, slots, ok := s.Contents(0, p)
			if !u.Held || u.At != p || u.In != hold || !ok || !slices.ContainsFunc(slots, func(v UnitView) bool { return v.Index == i }) {
				t.Errorf("unit %d after entering %v: %+v, contents %v", i, p, u, slots)
			}
			if _, ok := s.View(0).Units[p]; ok {
				t.Errorf("side 0's map shows a unit on the entrance")
			}
			return
		}
	}
	t.Fatal("no unit of side 0 reaches an own building")
}

func TestRepairAndBuild(t *testing.T) {
	var lv *level.Level
	for _, l := range shipped(t) {
		if l.MapName == "Civil" {
			lv = l
		}
	}
	s := session(t, lv, Options{})
	s.Game.Energy[0] = 1000
	repaired, built := false, false
	for _, p := range Coords(toSet(s.View(0).Buildings)) {
		hold, slots, ok := s.Contents(0, p)
		if !ok {
			continue
		}
		if !repaired && len(slots) > 0 {
			u := s.Game.Units[slots[0].Index]
			u.Size--
			s.Game.Units[slots[0].Index] = u
			evs, err := s.Repair(0, hold, slots[0].Slot)
			if err != nil || len(evs) != 1 || evs[0].Kind != Repaired || evs[0].To != p || evs[0].Energy != s.Energy(0) ||
				s.Unit(slots[0].Index).Size != slots[0].Full {
				t.Errorf("Repair: %+v, %v", evs, err)
			}
			repaired = true
		}
		if b := s.Game.Buildings[hold]; !built && len(b.Builds) > 0 && len(slots) < 7 {
			evs, err := s.Build(0, hold, b.Builds[0])
			if err != nil || len(evs) != 1 || evs[0].Kind != Built || evs[0].To != p {
				t.Fatalf("Build: %+v, %v", evs, err)
			}
			if u := s.Unit(evs[0].Unit); u.ID == "" || u.Name == "" || u.Type != b.Builds[0] || !u.Held || u.In != hold || u.At != p {
				t.Errorf("the built unit: %+v", u)
			}
			built = true
		}
		checkBound(t, s)
	}
	if !repaired || !built {
		t.Errorf("repaired %v, built %v", repaired, built)
	}
}

func TestCosts(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{})
		hexes := 0
		for _, line := range lv.Lines {
			hexes += len(line)
		}
		closed, held := false, false
		for _, i := range own(s, 0) {
			costs, err := s.Costs(0, i)
			if err != nil {
				t.Fatalf("%s: Costs(0, %d): %v", lv.MapName, i, err)
			}
			for p, c := range costs {
				if _, ok := s.Board.Cell(p); !ok || c <= 0 {
					t.Fatalf("%s: unit %d: cost %d at %v", lv.MapName, i, c, p)
				}
			}
			closed = closed || len(costs) < hexes
			u := s.Unit(i)
			if u.Held {
				held = true
				continue
			}
			d, err := s.Destinations(0, i)
			if err != nil {
				t.Fatal(err)
			}
			for p := range d {
				if _, ok := costs[p]; !ok {
					t.Errorf("%s: unit %d reaches %v, which Costs closes", lv.MapName, i, p)
				}
			}
		}
		if !closed {
			t.Errorf("%s: no unit has a closed hex", lv.MapName)
		}
		if lv.MapName == "Civil" && !held {
			t.Errorf("%s: no held unit checked", lv.MapName)
		}
	}
	s := session(t, shipped(t)[0], Options{})
	if _, err := s.Costs(2, 0); err == nil {
		t.Error("Costs for side 2")
	}
	if _, err := s.Costs(0, 999); err == nil {
		t.Error("Costs for no unit")
	}
}
