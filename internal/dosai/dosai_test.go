package dosai

import (
	"io/fs"
	"reflect"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// maxTurns bounds a played-out battle.
const maxTurns = 300

// shipped are the shipped levels in file order.
func shipped(tb testing.TB) []*level.Level {
	tb.Helper()
	names, err := fs.Glob(levels.FS, "level?_cp.json")
	if err != nil || len(names) == 0 {
		tb.Fatalf("no shipped levels: %v", err)
	}
	var out []*level.Level
	for _, n := range names {
		raw, err := levels.FS.ReadFile(n)
		if err != nil {
			tb.Fatal(err)
		}
		lv, err := level.Parse(raw)
		if err != nil {
			tb.Fatalf("%s: %v", n, err)
		}
		out = append(out, lv)
	}
	return out
}

var set *data.Set

func dataSet(tb testing.TB) *data.Set {
	tb.Helper()
	if set == nil {
		s, err := data.Load()
		if err != nil {
			tb.Fatal(err)
		}
		set = s
	}
	return set
}

func session(tb testing.TB, lv *level.Level) *dosplay.Session {
	tb.Helper()
	s, err := dosplay.New(lv, dataSet(tb), dosplay.Options{Computer: [2]bool{true, true}})
	if err != nil {
		tb.Fatalf("%s: %v", lv.MapName, err)
	}
	return s
}

// play plays a battle out and returns its events; a refused move fails.
func play(t *testing.T, s *dosplay.Session, p [2]*Planner) (*dosrules.End, []dosplay.Event) {
	t.Helper()
	var evs []dosplay.Event
	end, err := playOut(s, p, maxTurns, func(e []dosplay.Event) { evs = append(evs, e...) })
	if err != nil {
		t.Fatalf("%s: %v", s.Level.MapName, err)
	}
	for _, ev := range evs {
		if ev.Kind == dosplay.Refused {
			t.Fatalf("%s: a move of unit %d to %v was refused: %s", s.Level.MapName, ev.Unit, ev.To, ev.Text)
		}
	}
	return end, evs
}

func count(evs []dosplay.Event, k dosplay.EventKind) int {
	n := 0
	for _, ev := range evs {
		if ev.Kind == k {
			n++
		}
	}
	return n
}

func TestPlayOut(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv)
		end, evs := play(t, s, [2]*Planner{{}, {}})
		if end == nil || !end.Over {
			t.Errorf("%s: open after %d resolutions", lv.MapName, maxTurns)
			continue
		}
		t.Logf("%s: winner %d (%v), turn %d, score %d; %d moves, %d attack orders, %d battles",
			lv.MapName, end.Winner, end.Won, s.Turn(), end.Score, count(evs, dosplay.Moved), count(evs, dosplay.Ordered),
			count(evs, dosplay.Battle))
	}
}

func TestDeterminism(t *testing.T) {
	for _, lv := range shipped(t) {
		_, a := play(t, session(t, lv), [2]*Planner{{}, {}})
		_, b := play(t, session(t, lv), [2]*Planner{{}, {}})
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s: two playouts gave different events", lv.MapName)
		}
	}
}

func TestAgainstPass(t *testing.T) {
	for _, lv := range shipped(t) {
		for side := range 2 {
			var p [2]*Planner
			p[side] = &Planner{}
			s := session(t, lv)
			end, _ := play(t, s, p)
			if end == nil || end.Winner != side {
				t.Errorf("%s: the planner of side %d against pass: %+v", lv.MapName, side, end)
				continue
			}
			t.Logf("%s: side %d beats pass (%v), turn %d, score %d", lv.MapName, side, end.Won[side], s.Turn(), end.Score)
		}
	}
}

// TestHidden checks that a plan does not depend on what the other side did
// since the last phase change: its moves (on its own map) and its attack
// orders. Two equal sessions plan the sides in opposite order; each side's
// plan must be the same either way.
func TestHidden(t *testing.T) {
	for _, lv := range shipped(t) {
		a, b := session(t, lv), session(t, lv)
		var p Planner
		for n := 0; n < 12 && a.Ended() == nil; n++ {
			var plans [2][2][]dosplay.Event // session, side
			for k := range 2 {
				ev, err := p.Plan(a, k)
				if err != nil {
					t.Fatal(err)
				}
				plans[0][k] = ev
			}
			for k := 1; k >= 0; k-- {
				ev, err := p.Plan(b, k)
				if err != nil {
					t.Fatal(err)
				}
				plans[1][k] = ev
			}
			for k := range 2 {
				if !reflect.DeepEqual(plans[0][k], plans[1][k]) {
					t.Fatalf("%s: resolution %d: side %d planned differently with the other side's orders given", lv.MapName, n, k)
				}
			}
			if _, err := a.Resolve(); err != nil {
				t.Fatal(err)
			}
			if _, err := b.Resolve(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// TestHiddenPositions moves the other side's units on its own map and
// changes their positions there: the plan must not change.
func TestHiddenPositions(t *testing.T) {
	for _, lv := range shipped(t) {
		attacks, moves := 0, 0
		for n := 0; n < 16; n += 3 {
			a, b := session(t, lv), session(t, lv)
			var p Planner
			for range n {
				for _, s := range []*dosplay.Session{a, b} {
					if _, err := PlayOut(s, [2]*Planner{&p, &p}, 1); err != nil {
						t.Fatal(err)
					}
				}
			}
			act, err := a.Game.Action()
			if err != nil {
				t.Fatal(err)
			}
			mv := 1 - act
			// The movement side moves in a, then the action side plans.
			if _, err := p.Plan(a, mv); err != nil {
				t.Fatal(err)
			}
			if scramble(a, mv) == 0 {
				t.Fatalf("%s: nothing to scramble", lv.MapName)
			}
			pa, err := p.Plan(a, act)
			if err != nil {
				t.Fatal(err)
			}
			pb, err := p.Plan(b, act)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(pa, pb) {
				t.Fatalf("%s: after %d resolutions side %d's plan depends on the other side's map", lv.MapName, n, act)
			}
			attacks += count(pa, dosplay.Ordered)
			// The other way round: the movement side plans after the
			// action side's orders, and after the action side's units
			// were scrambled on their own map.
			if _, err := p.Plan(b, mv); err != nil {
				t.Fatal(err)
			}
			c := session(t, lv)
			for range n {
				if _, err := PlayOut(c, [2]*Planner{&p, &p}, 1); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := p.Plan(c, act); err != nil {
				t.Fatal(err)
			}
			scramble(c, act)
			pc, err := p.Plan(c, mv)
			if err != nil {
				t.Fatal(err)
			}
			d := session(t, lv)
			for range n {
				if _, err := PlayOut(d, [2]*Planner{&p, &p}, 1); err != nil {
					t.Fatal(err)
				}
			}
			pd, err := p.Plan(d, mv)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(pc, pd) {
				t.Fatalf("%s: after %d resolutions side %d's moves depend on the other side's map or orders", lv.MapName, n, mv)
			}
			moves += count(pc, dosplay.Moved)
		}
		if attacks == 0 || moves == 0 {
			t.Errorf("%s: the plans compared hold %d attack orders and %d moves", lv.MapName, attacks, moves)
		}
	}
}

// scramble shifts side's units on side's own map and in their records' own
// positions by one cell where that cell is empty on that map; it returns
// the number of units shifted.
func scramble(s *dosplay.Session, side int) int {
	n := 0
	g := s.Game
	m := g.Maps[side]
	for i, u := range g.Units {
		c := u.Cell(side)
		if u.Destroyed() || u.Flags4&0x4000 != 0 || !dosrules.Owns(u.Flags4, side) || m[2*c+1] != i {
			continue
		}
		to := c + 1
		if to >= g.Width*g.Height || m[2*to+1] != 0xFF {
			continue
		}
		m[2*c+1], m[2*to+1] = 0xFF, i
		u.Pos[side] = 2*to + 1
		g.Units[i] = u
		n++
	}
	return n
}

func TestDistanceMatchesLayout(t *testing.T) {
	s := session(t, shipped(t)[0])
	w := s.Board.Width
	for a := 0; a < w*s.Board.Height; a += 7 {
		for b := 0; b < w*s.Board.Height; b += 13 {
			ca, oka := s.Board.Coord(a)
			cb, okb := s.Board.Coord(b)
			if !oka || !okb {
				continue
			}
			if d, l := hexgrid.Distance(ca, cb), s.Game.Profile.LayoutDistance(w, a, b); d != l {
				t.Fatalf("cells %d, %d: hex distance %d, layout distance %d", a, b, d, l)
			}
		}
	}
}

// TestPlanTime logs the slowest Plan call of a played-out battle per level.
func TestPlanTime(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv)
		var p Planner
		var slowest time.Duration
		for n := 0; n < maxTurns && s.Ended() == nil; n++ {
			for k := range 2 {
				t0 := time.Now()
				if _, err := p.Plan(s, k); err != nil {
					t.Fatal(err)
				}
				slowest = max(slowest, time.Since(t0))
			}
			if _, err := s.Resolve(); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("%s: the slowest Plan took %v", lv.MapName, slowest)
	}
}

// midgame is a session of the largest level after some resolutions, saved.
func midgame(b *testing.B) (*level.Level, []byte) {
	b.Helper()
	var lv *level.Level
	for _, l := range shipped(b) {
		if lv == nil || hexes(l) > hexes(lv) {
			lv = l
		}
	}
	s := session(b, lv)
	var p Planner
	if _, err := PlayOut(s, [2]*Planner{&p, &p}, 8); err != nil {
		b.Fatal(err)
	}
	raw, err := s.Save()
	if err != nil {
		b.Fatal(err)
	}
	return lv, raw
}

func hexes(lv *level.Level) int {
	n := 0
	for _, line := range lv.Lines {
		n += len(line)
	}
	return n
}

func benchmarkPlan(b *testing.B, phase int) {
	lv, raw := midgame(b)
	var p Planner
	for b.Loop() {
		b.StopTimer()
		s, err := dosplay.Load(lv, dataSet(b), raw)
		if err != nil {
			b.Fatal(err)
		}
		side := 0
		if s.Phase(1) == phase {
			side = 1
		}
		b.StartTimer()
		if _, err := p.Plan(s, side); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPlanMovement(b *testing.B) { benchmarkPlan(b, dosrules.PhaseMovement) }

func BenchmarkPlanAction(b *testing.B) { benchmarkPlan(b, dosrules.PhaseAction) }

// A Progress callback reports each phase's plan from 0 to its total
// without changing the plans.
func TestProgress(t *testing.T) {
	for _, lv := range shipped(t) {
		_, want := play(t, session(t, lv), [2]*Planner{{}, {}})
		var calls [][2]int
		phases := 0
		report := func(done, total int) {
			if done == 0 {
				phases++
			}
			calls = append(calls, [2]int{done, total})
		}
		_, got := play(t, session(t, lv), [2]*Planner{{Progress: report}, {Progress: report}})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the progress callback changed the plans", lv.MapName)
		}
		if phases == 0 {
			t.Fatalf("%s: no progress reported", lv.MapName)
		}
		for i, c := range calls {
			done, total := c[0], c[1]
			switch {
			case done < 0 || done > total:
				t.Fatalf("%s: progress %d of %d", lv.MapName, done, total)
			case done > 0 && (calls[i-1][1] != total || calls[i-1][0] > done):
				t.Fatalf("%s: progress %v after %v", lv.MapName, c, calls[i-1])
			case i+1 == len(calls) || calls[i+1][0] == 0:
				if done != total {
					t.Fatalf("%s: a plan ended at %d of %d", lv.MapName, done, total)
				}
			}
		}
		t.Logf("%s: %d plans, %d progress reports", lv.MapName, phases, len(calls))
	}
}
