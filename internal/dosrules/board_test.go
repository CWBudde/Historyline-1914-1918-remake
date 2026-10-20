package dosrules

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"testing"
)

// The typed API is checked against the recorded fixtures: the records it
// decodes must be the recorded bytes, and its routes, refusals and
// occupants what the record-level outcome holds.

func record(r FixtureRecord) []byte {
	b := make([]byte, len(r.Bytes))
	for i, v := range r.Bytes {
		b[i] = byte(v)
	}
	return b
}

// TestLiveRoundTrip reads and writes back every recorded live and
// container record.
func TestLiveRoundTrip(t *testing.T) {
	var lives, holds int
	for build := range profiles {
		var live, hold []FixtureRecord
		for _, f := range load[StepFixture](t, build, "steps.json") {
			live = append(live, f.Expected.Live...)
		}
		for _, f := range load[ContainerFixture](t, build, "containers.json") {
			live, hold = append(live, f.Expected.Live...), append(hold, f.Expected.Holds...)
		}
		for _, f := range load[VisibilityFixture](t, build, "visibility.json") {
			live, hold = append(live, f.Expected.Live...), append(hold, f.Expected.Holds...)
		}
		for _, r := range live {
			if b := record(r); !bytes.Equal(DecodeLive(b).Bytes(), b) {
				t.Fatalf("%s: live record %d % x reads back as % x", build, r.Index, b, DecodeLive(b).Bytes())
			}
			lives++
		}
		for _, r := range hold {
			if b := record(r); !bytes.Equal(DecodeHold(b).Bytes(), b) {
				t.Fatalf("%s: container record %d % x reads back as % x", build, r.Index, b, DecodeHold(b).Bytes())
			}
			holds++
		}
	}
	if lives == 0 || holds == 0 {
		t.Fatalf("%d live and %d container records", lives, holds)
	}
	t.Logf("%d live and %d container records", lives, holds)
}

func TestLiveFields(t *testing.T) {
	b := make([]byte, LiveLen)
	for i := range b {
		b[i] = byte(i + 1)
	}
	u := DecodeLive(b)
	if u.Size != 3 || u.Flags4 != 0x0605 || u.Type != 9 || u.Hold != 11 || u.Pos != [2]int{0x0D0C, 0x0F0E} ||
		u.Facing != [2]int{16, 17} || u.Target != 0x1312 || u.Action[6] != 26 {
		t.Fatalf("%+v", u)
	}
	if c := (Live{Pos: [2]int{2*57 + 1, 2*12 + 1}}); c.Cell(0) != 57 || c.Cell(1) != 12 {
		t.Fatalf("cells %d %d", c.Cell(0), c.Cell(1))
	}
}

func TestNoMovementTables(t *testing.T) {
	p := Profile{Name: "dos-0.1", Tables: &Tables{}}
	if _, err := p.Move(&Movement{}, MoveOrder{}); err != ErrNoMovement {
		t.Fatalf("Move without tables: %v", err)
	}
}

// TestTypedRoutes asks every recorded movement query for its destinations
// and routes through the typed API.
func TestTypedRoutes(t *testing.T) {
	seen := map[string]int{}
	for build, p := range profiles {
		for _, f := range load[MovementFixture](t, build, "movements.json") {
			g := f.Query()
			if x := f.Expected.Destinations; x != nil {
				got, err := p.Destinations(g)
				if err != nil || !slices.Equal(got, x) {
					t.Errorf("%s: destinations %v (%v), recorded %v", f.ID, got, err, x)
				}
				seen["destinations"]++
			}
			for k, goal := range f.Goals {
				got, err := p.Route(g, goal)
				want := slices.Clone(f.Expected.Paths[k])
				slices.Reverse(want)
				if err != nil || !slices.Equal(got, want) {
					t.Errorf("%s: route to %d is %v (%v), recorded %v", f.ID, goal, got, err, want)
				}
				if want != nil && len(got) > 0 && got[0] == g.Units[g.Unit].Cell && got[len(got)-1] == goal {
					seen["route from the unit to the goal"]++
				}
			}
		}
	}
	need(t, seen, "destinations", "route from the unit to the goal")
}

// movedAgrees compares a typed outcome with the recorded one.
func movedAgrees(g *Movement, m Moved, x StepExpected, seen map[string]int) error {
	if refused := x.Order&1 == 0 || x.Start&1 == 0; m.Refused != refused {
		return fmt.Errorf("refused %v, recorded order %#x start %#x", m.Refused, x.Order, x.Start)
	}
	if m.TwoCell != (x.Order&8 != 0) || !slices.Equal(m.Route, x.Cells) {
		return fmt.Errorf("two-cell %v route %v, recorded order %#x cells %v", m.TwoCell, m.Route, x.Order, x.Cells)
	}
	if len(m.Messages) != len(x.Messages) {
		return fmt.Errorf("messages %v, recorded %v", m.Messages, x.Messages)
	}
	for i, msg := range m.Messages {
		if msg != (Message{x.Messages[i][0], x.Messages[i][1]}) {
			return fmt.Errorf("messages %v, recorded %v", m.Messages, x.Messages)
		}
	}
	if len(m.Units) != len(x.Live) {
		return fmt.Errorf("%d units, recorded %d", len(m.Units), len(x.Live))
	}
	for _, r := range x.Live {
		if !bytes.Equal(m.Units[r.Index].Bytes(), record(r)) {
			return fmt.Errorf("unit %d is %+v, recorded % x", r.Index, m.Units[r.Index], record(r))
		}
	}
	for k := range 2 {
		want := map[int]int{}
		for _, o := range x.Occupants[k] {
			want[o[0]] = o[1]
		}
		if !maps.Equal(m.Occupants[k], want) {
			return fmt.Errorf("map %d occupants %v, recorded %v", k, m.Occupants[k], want)
		}
	}
	switch u := m.Units[g.Unit]; {
	case m.Refused:
		seen["refused"]++
	case m.TwoCell:
		seen["two-cell ship"]++
	case len(m.Route) > 0 && u.Cell(g.Side) == m.Route[len(m.Route)-1]:
		seen["arrived at the route's end"]++
	}
	return nil
}

func TestTypedMoves(t *testing.T) {
	seen := map[string]int{}
	for build, p := range profiles {
		for _, f := range load[StepFixture](t, build, "steps.json") {
			g := f.Movement()
			m, err := p.Move(g, f.Order)
			if err == nil {
				err = movedAgrees(g, m, f.Expected, seen)
			}
			if err != nil {
				t.Errorf("%s: %v", f.ID, err)
			}
		}
	}
	need(t, seen, "refused", "two-cell ship", "arrived at the route's end")
}

func TestTypedContainers(t *testing.T) {
	seen := map[string]int{}
	for build, p := range profiles {
		for _, f := range load[ContainerFixture](t, build, "containers.json") {
			g := f.Movement()
			var m Moved
			var err error
			if f.Order.Hold < 0 {
				m, err = p.Move(g, f.Order.MoveOrder)
			} else {
				m, err = p.Exit(g, f.Order)
			}
			if err == nil {
				err = movedAgrees(g, m, f.Expected.StepExpected, seen)
			}
			if err == nil {
				err = holdsAgree(m.Holds, f.Expected.Holds)
			}
			if err == nil && f.Order.Hold >= 0 {
				want := map[int]int{}
				for _, r := range f.Expected.Remaining {
					want[r[0]] = r[1]
				}
				if !maps.Equal(m.Remaining, want) {
					err = fmt.Errorf("exit reach %v, recorded %v", m.Remaining, want)
				}
				seen["exit"]++
			}
			if err != nil {
				t.Errorf("%s: %v", f.ID, err)
			}
			if len(m.Holds) > 0 {
				seen["container records changed"]++
			}
		}
	}
	need(t, seen, "exit", "container records changed", "refused", "arrived at the route's end")
}

func holdsAgree(got map[int]Hold, want []FixtureRecord) error {
	if len(got) != len(want) {
		return fmt.Errorf("%d container records changed, recorded %d", len(got), len(want))
	}
	for _, r := range want {
		if h, ok := got[r.Index]; !ok || !bytes.Equal(h.Bytes(), record(r)) {
			return fmt.Errorf("container record %d is %+v, recorded % x", r.Index, h, record(r))
		}
	}
	return nil
}

func TestTypedPhaseChanges(t *testing.T) {
	seen := map[string]int{}
	for build, p := range profiles {
		for _, f := range load[VisibilityFixture](t, build, "visibility.json") {
			g := f.PhaseChange()
			ph, err := p.ChangePhase(g)
			if err == nil {
				err = phasedAgrees(ph, f.Expected, seen)
			}
			if err != nil {
				t.Errorf("%s: %v", f.ID, err)
			}
		}
	}
	need(t, seen, "unit in both views", "unit left the views", "container records changed")
}

func phasedAgrees(ph Phased, x VisibilityExpected, seen map[string]int) error {
	for _, r := range x.Live {
		u, ok := ph.Units[r.Index]
		if !ok || !bytes.Equal(u.Bytes(), record(r)) {
			return fmt.Errorf("unit %d is %+v, recorded % x", r.Index, u, record(r))
		}
		if u.Flags4&0xC000 != 0 {
			seen["unit left the views"]++
		} else if ph.Occupants[0][u.Cell(0)] == r.Index && ph.Occupants[1][u.Cell(0)] == r.Index {
			seen["unit in both views"]++
		}
	}
	if err := holdsAgree(ph.Holds, x.Holds); err != nil {
		return err
	}
	if len(ph.Holds) > 0 {
		seen["container records changed"]++
	}
	for k := range 2 {
		if ph.Phase[k] != x.Players[k][0x16] {
			return fmt.Errorf("side %d phase %d, recorded %d", k, ph.Phase[k], x.Players[k][0x16])
		}
	}
	if ph.Count != x.Count || len(ph.Messages) != len(x.Messages) {
		return fmt.Errorf("count %d messages %v, recorded %d %v", ph.Count, ph.Messages, x.Count, x.Messages)
	}
	return nil
}

func need(t *testing.T, seen map[string]int, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if seen[k] == 0 {
			t.Errorf("no case showed %q", k)
		}
	}
	t.Logf("%v", seen)
}
