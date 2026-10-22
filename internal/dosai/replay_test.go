package dosai

import (
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Planning off the frame: the computer side (side 1) plans on a clone in a
// goroutine while side 0 gives its orders on the session, then the clone's
// commands are replayed onto the session. That gives the same events and
// the same state as planning on the session itself before side 0's orders,
// in both phases, resolution after resolution.
func TestPlanOnClone(t *testing.T) {
	const computer, human = 1, 0
	for _, lv := range shipped(t) {
		a, b := session(t, lv), session(t, lv)
		var p Planner
		phases := map[int]bool{}
		moves, attacks := 0, 0
		for n := 0; n < 10 && a.Ended() == nil; n++ {
			phases[a.Phase(computer)] = true
			// (a) On the session.
			want, err := p.Plan(a, computer)
			if err != nil {
				t.Fatal(err)
			}
			wantHuman, err := p.Plan(a, human)
			if err != nil {
				t.Fatal(err)
			}
			// (b) On a clone, the other side's orders given meanwhile.
			c, err := b.Clone()
			if err != nil {
				t.Fatal(err)
			}
			type result struct {
				cmds []dosrules.Command
				err  error
			}
			done := make(chan result)
			go func() {
				cmds, err := p.Commands(c, computer)
				done <- result{cmds, err}
			}()
			gotHuman, err := p.Plan(b, human)
			if err != nil {
				t.Fatal(err)
			}
			r := <-done
			if r.err != nil {
				t.Fatal(r.err)
			}
			got, err := b.Replay(r.cmds)
			if err != nil {
				t.Fatalf("%s: resolution %d: Replay: %v", lv.MapName, n, err)
			}
			if len(r.cmds) == 0 || r.cmds[len(r.cmds)-1].Kind != dosplay.CmdConfirm {
				t.Fatalf("%s: resolution %d: the plan's commands %v do not end with the confirmation", lv.MapName, n, r.cmds)
			}
			if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(gotHuman, wantHuman) {
				t.Fatalf("%s: resolution %d: the replayed plan gave other events", lv.MapName, n)
			}
			moves, attacks = moves+count(got, dosplay.Moved), attacks+count(got, dosplay.Ordered)
			same(t, a, b, "after the plans")
			ra, err := a.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			rb, err := b.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(ra, rb) {
				t.Fatalf("%s: resolution %d resolved differently", lv.MapName, n)
			}
			same(t, a, b, "after the resolution")
		}
		if !phases[dosrules.PhaseMovement] || !phases[dosrules.PhaseAction] {
			t.Errorf("%s: the computer planned in phases %v only", lv.MapName, phases)
		}
		if moves == 0 || attacks == 0 {
			t.Errorf("%s: the replayed plans hold %d moves and %d attack orders", lv.MapName, moves, attacks)
		}
	}
}

// same fails unless a and b save alike and count the same version.
func same(t *testing.T, a, b *dosplay.Session, when string) {
	t.Helper()
	sa, err := a.Save()
	if err != nil {
		t.Fatal(err)
	}
	sb, err := b.Save()
	if err != nil {
		t.Fatal(err)
	}
	if string(sa) != string(sb) || a.Version() != b.Version() {
		t.Fatalf("%s: the sessions differ %s (version %d, %d)", a.Level.MapName, when, a.Version(), b.Version())
	}
}

// Planning on a clone leaves the session as it was.
func TestCommandsOnClone(t *testing.T) {
	s := session(t, shipped(t)[0])
	var p Planner
	if _, err := PlayOut(s, [2]*Planner{&p, &p}, 3); err != nil {
		t.Fatal(err)
	}
	raw, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	v, mark := s.Version(), s.Mark()
	c, err := s.Clone()
	if err != nil {
		t.Fatal(err)
	}
	for k := range 2 {
		if _, err := p.Commands(c, k); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := c.Resolve(); err != nil {
		t.Fatal(err)
	}
	if again, _ := s.Save(); string(again) != string(raw) || s.Version() != v || s.Mark() != mark {
		t.Error("planning on the clone changed the session")
	}
	if cmds, err := p.Commands(s, 2); err == nil || cmds != nil {
		t.Errorf("Commands for side 2: %v, %v", cmds, err)
	}
}
