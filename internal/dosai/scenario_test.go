package dosai

import (
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// scenarioEnergy is each side's account at the start of the integrated
// scenario, enough for its repairs and builds.
const scenarioEnergy = 1000

// TestIntegratedScenario plays every shipped level (PULSE in summer, Civil
// and Mouse in winter) computer against computer for up to maxTurns
// resolutions, each side repairing once and building once in the first of
// its action phases that allows it, and replays the command log on a new
// session: the same events and the same state. Over the three levels both
// sides move, leave buildings, attack, fight, repair, build and capture.
func TestIntegratedScenario(t *testing.T) {
	var cmds [2]map[dosrules.CommandKind]int
	var evs [2]map[dosplay.EventKind]int
	for k := range 2 {
		cmds[k], evs[k] = map[dosrules.CommandKind]int{}, map[dosplay.EventKind]int{}
	}
	levels := shipped(t)
	winter := 0
	for _, lv := range levels {
		a := scenarioSession(t, lv)
		if a.Board.Winter {
			winter++
		}
		var want []dosplay.Event
		var repaired, built [2]bool
		p := [2]*Planner{{}, {}}
		for range maxTurns {
			if a.Ended() != nil {
				break
			}
			for k := range 2 {
				if a.Phase(k) == dosrules.PhaseAction {
					e := service(t, a, k, !repaired[k], !built[k])
					repaired[k] = repaired[k] || count(e, dosplay.Repaired) > 0
					built[k] = built[k] || count(e, dosplay.Built) > 0
					want = append(want, e...)
				}
				e, err := p[k].Plan(a, k)
				if err != nil {
					t.Fatalf("%s: %v", lv.MapName, err)
				}
				want = append(want, e...)
			}
			e, err := a.Resolve()
			if err != nil {
				t.Fatalf("%s: %v", lv.MapName, err)
			}
			want = append(want, e...)
		}
		log := a.CommandsSince(0)
		for _, c := range log {
			cmds[c.Side][c.Kind]++
		}
		for _, ev := range want {
			if ev.Kind == dosplay.Refused {
				t.Fatalf("%s: refused: %s", lv.MapName, ev.Text)
			}
			if ev.Side == 0 || ev.Side == 1 {
				evs[ev.Side][ev.Kind]++
			}
		}
		b := scenarioSession(t, lv)
		got, err := b.Replay(log)
		if err != nil {
			t.Fatalf("%s: Replay: %v", lv.MapName, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the replayed log gave other events (%d, want %d)", lv.MapName, len(got), len(want))
		}
		same(t, a, b, "after the replay")
		t.Logf("%s: %d commands, %d events, turn %d, end %+v", lv.MapName, len(log), len(want), a.Turn(), a.Ended())
	}
	if winter == 0 || winter == len(levels) {
		t.Errorf("%d of %d shipped levels are winter maps, want both seasons", winter, len(levels))
	}
	if cmds[0][dosplay.CmdResolve] == 0 {
		t.Error("no resolution logged")
	}
	for k := range 2 {
		t.Logf("side %d: commands %v, events %v", k, cmds[k], evs[k])
		for _, c := range []dosrules.CommandKind{dosrules.CmdMove, dosrules.CmdExit, dosrules.CmdAttack, dosrules.CmdRepair, dosrules.CmdBuild, dosplay.CmdConfirm} {
			if cmds[k][c] == 0 {
				t.Errorf("side %d gave no %s command", k, c)
			}
		}
		for _, e := range []dosplay.EventKind{dosplay.Battle, dosplay.Captured, dosplay.Repaired, dosplay.Built} {
			if evs[k][e] == 0 {
				t.Errorf("side %d has no %v event", k, e)
			}
		}
	}
}

// scenarioSession is a computer-against-computer session on lv with
// scenarioEnergy on both accounts.
func scenarioSession(t *testing.T, lv *level.Level) *dosplay.Session {
	t.Helper()
	s := session(t, lv)
	s.Game.Energy = [2]int{scenarioEnergy, scenarioEnergy}
	return s
}

// service repairs the first unit (repair) and builds the first type it can
// (build) in the first of side's buildings of s that allows it; it returns
// the events.
func service(t *testing.T, s *dosplay.Session, side int, repair, build bool) []dosplay.Event {
	t.Helper()
	var evs []dosplay.Event
	for _, at := range sorted(s.View(side).Buildings) {
		hold, slots, ok := s.Contents(side, at)
		if !ok {
			continue
		}
		if repair && len(slots) > 0 {
			ev, err := s.Repair(side, hold, slots[0].Slot)
			if err != nil {
				t.Fatalf("%s: repair at %v: %v", s.Level.MapName, at, err)
			}
			evs, repair = append(evs, ev...), false
		}
		if opts := s.Buildable(side, hold); build && len(opts) > 0 && len(slots) < dosplay.Slots {
			ev, err := s.Build(side, hold, opts[0].Type)
			if err != nil {
				t.Fatalf("%s: build at %v: %v", s.Level.MapName, at, err)
			}
			evs, build = append(evs, ev...), false
		}
	}
	return evs
}
