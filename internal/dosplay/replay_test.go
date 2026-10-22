package dosplay

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// repairBuild repairs and builds once in side 0's buildings of s, as
// TestRepairAndBuild does; it returns the events.
func repairBuild(t *testing.T, s *Session) []Event {
	t.Helper()
	var evs []Event
	for _, p := range Coords(toSet(s.View(0).Buildings)) {
		hold, slots, ok := s.Contents(0, p)
		if !ok {
			continue
		}
		if len(slots) > 0 {
			ev, err := s.Repair(0, hold, slots[0].Slot)
			if err != nil {
				t.Fatal(err)
			}
			evs = append(evs, ev...)
		}
		if opts := s.Buildable(0, hold); len(opts) > 0 && len(slots) < Slots {
			ev, err := s.Build(0, hold, opts[0].Type)
			if err != nil {
				t.Fatal(err)
			}
			evs = append(evs, ev...)
		}
	}
	return evs
}

// played is a session on lv after a few command cycles (a confirmation
// withdrawn, on Civil a repair and a build) and an open move, and the
// events of all its commands.
func played(t *testing.T, lv *level.Level, o Options) (*Session, []Event) {
	t.Helper()
	s := session(t, lv, o)
	s.Game.Energy[0] = 1000
	evs := s.Confirm(0)
	evs = append(evs, s.Unconfirm(0)...)
	evs = append(evs, s.Unconfirm(0)...) // not confirmed: no command
	if lv.MapName == "Civil" {
		evs = append(evs, repairBuild(t, s)...)
	}
	for range 3 {
		evs = append(evs, cycle(t, s, approach)...)
	}
	mv := 1 - must(s.Game.Action())
	for _, i := range own(s, mv) {
		if d, err := s.Destinations(mv, i); err == nil && len(d) > 0 {
			evs = append(evs, must(s.Move(mv, i, Coords(d)[0]))...)
			break
		}
	}
	return s, evs
}

// The log replayed on an equal new session gives the same events and the
// same state; a mark replays the commands since.
func TestReplay(t *testing.T) {
	attacks, exits := 0, 0
	for _, lv := range shipped(t) {
		o := Options{Limit: 8}
		a, want := played(t, lv, o)
		cmds := a.CommandsSince(0)
		if len(cmds) != a.Mark() || a.CommandsSince(a.Mark()) != nil || a.CommandsSince(-1) != nil || a.CommandsSince(a.Mark()+1) != nil {
			t.Fatalf("%s: %d commands, mark %d", lv.MapName, len(cmds), a.Mark())
		}
		attacks += countKind(cmds, dosrules.CmdAttack)
		exits += countKind(cmds, dosrules.CmdExit)
		for _, k := range []dosrules.CommandKind{dosrules.CmdMove, CmdConfirm, CmdUnconfirm, CmdResolve} {
			if !slices.ContainsFunc(cmds, func(c dosrules.Command) bool { return c.Kind == k }) {
				t.Errorf("%s: no %s command in the log", lv.MapName, k)
			}
		}
		if n := countKind(cmds, CmdUnconfirm); n != 1 {
			t.Errorf("%s: %d withdrawals logged, want 1", lv.MapName, n)
		}
		if lv.MapName == "Civil" && (countKind(cmds, dosrules.CmdRepair) == 0 || countKind(cmds, dosrules.CmdBuild) == 0) {
			t.Errorf("%s: no repair or build logged", lv.MapName)
		}
		b := session(t, lv, o)
		b.Game.Energy[0] = 1000
		got, err := b.Replay(cmds)
		if err != nil {
			t.Fatalf("%s: Replay: %v", lv.MapName, err)
		}
		checkBound(t, b)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the replay gave other events", lv.MapName)
		}
		if b.Version() != a.Version() || string(must(b.Save())) != string(must(a.Save())) {
			t.Errorf("%s: the replay left another state (version %d, want %d)", lv.MapName, b.Version(), a.Version())
		}
		if !reflect.DeepEqual(b.CommandsSince(0), cmds) {
			t.Errorf("%s: the replay logged other commands", lv.MapName)
		}
		// A mark taken now replays only the cycle that follows, and the
		// replay reveals the moves made before the mark.
		m, c := a.Mark(), must(a.Clone())
		want = cycle(t, a, approach)
		got, err = c.Replay(a.CommandsSince(m))
		if err != nil || !reflect.DeepEqual(got, want) || string(must(c.Save())) != string(must(a.Save())) {
			t.Errorf("%s: the replay since the mark: %v", lv.MapName, err)
		}
	}
	if attacks == 0 || exits == 0 {
		t.Errorf("%d attack and %d exit commands in the logs", attacks, exits)
	}
}

func countKind(cmds []dosrules.Command, k dosrules.CommandKind) int {
	n := 0
	for _, c := range cmds {
		if c.Kind == k {
			n++
		}
	}
	return n
}

// A replay stops at the first command the session refuses, with the events
// so far.
func TestReplayRefused(t *testing.T) {
	s := session(t, shipped(t)[0], Options{})
	evs, err := s.Replay([]dosrules.Command{{Kind: CmdConfirm, Side: 1}, {Kind: dosrules.CmdMove, Side: 0, Unit: 999}})
	var oe *OrderError
	if !errors.As(err, &oe) || len(evs) != 1 || evs[0].Kind != Confirmed {
		t.Errorf("Replay: %+v, %v", evs, err)
	}
	for _, c := range []dosrules.Command{{Kind: "fly"}, {Kind: CmdConfirm, Side: 2}, {Kind: dosrules.CmdMove, Cell: -1}} {
		if _, err := s.Replay([]dosrules.Command{c}); err == nil {
			t.Errorf("Replay(%+v) passed", c)
		}
	}
}

// A clone saves, sees and plays on like its original; changing it leaves
// the original as it was.
func TestClone(t *testing.T) {
	for _, lv := range shipped(t) {
		s, _ := played(t, lv, Options{Limit: 8})
		raw, v := must(s.Save()), s.Version()
		c, err := s.Clone()
		if err != nil {
			t.Fatalf("%s: Clone: %v", lv.MapName, err)
		}
		ref := must(s.Clone())
		if string(must(c.Save())) != string(raw) || c.Version() != v || !reflect.DeepEqual(c.CommandsSince(0), s.CommandsSince(0)) {
			t.Errorf("%s: the clone differs", lv.MapName)
		}
		for k := range 2 {
			if !reflect.DeepEqual(c.View(k), s.View(k)) || c.OrdersLeft(k) != s.OrdersLeft(k) {
				t.Errorf("%s: side %d sees another battle in the clone", lv.MapName, k)
			}
		}
		// Change the clone through its commands and in every part of its
		// state.
		for range 2 {
			cycle(t, c, first)
		}
		g := c.Game
		for k := range 2 {
			g.Maps[k][0]++
			g.Players[k].Orders = append(g.Players[k].Orders[:0], 1, 2, 3)
			g.Energy[k] = -5
		}
		for i, u := range g.Units {
			u.Size, u.Pos[0], u.Action[0] = 99, 0, 77
			g.Units[i] = u
		}
		for i, h := range g.Holds {
			h.Slots[0], h.Cells[0] = 1, 0
			g.Holds[i] = h
		}
		for i, b := range g.Buildings {
			for k := range b.Builds {
				b.Builds[k] = 0
			}
			b.Credits = 0
			g.Buildings[i] = b
		}
		for k := range c.routes {
			for i := range c.routes[k] {
				c.routes[k][i].Unit = -1
			}
		}
		c.ids[0] = "changed"
		c.Game.Ended = &dosrules.End{Over: true}
		if string(must(s.Save())) != string(raw) || s.Version() != v {
			t.Errorf("%s: changing the clone changed the original", lv.MapName)
		}
		// The original plays on as an untouched clone does (its routes
		// reveal the same moves).
		if !reflect.DeepEqual(cycle(t, s, approach), cycle(t, ref, approach)) {
			t.Errorf("%s: the original plays on differently after the clone changed", lv.MapName)
		}
	}
	// The clone keeps the session's profile, which a save does not hold.
	p := dosrules.Full
	p.TransposedDistance = false
	s := session(t, shipped(t)[0], Options{Profile: &p})
	if c := must(s.Clone()); !reflect.DeepEqual(c.Game.Profile, s.Game.Profile) {
		t.Errorf("the clone has profile %+v, want %+v", c.Game.Profile, s.Game.Profile)
	}
}
