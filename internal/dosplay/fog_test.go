package dosplay

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// sighted is the data set with every unit's sight set to n.
func sighted(t *testing.T, n int) *data.Set {
	t.Helper()
	set := *dataSet(t)
	set.Units = map[string]data.Unit{}
	for id, u := range dataSet(t).Units {
		u.Sight = n
		set.Units[id] = u
	}
	return &set
}

// fogSession is a session on lv with fog and the data set set.
func fogSession(t *testing.T, lv *level.Level, set *data.Set, fog bool) *Session {
	t.Helper()
	s, err := New(lv, set, Options{Fog: fog})
	if err != nil {
		t.Fatalf("%s: New: %v", lv.MapName, err)
	}
	return s
}

// TestFogView: without fog every hex is in sight and events pass; with fog
// a side's view keeps its own units and leaves out only units on hexes out
// of its sight, the buildings stay, and TrueView is the same as without.
func TestFogView(t *testing.T) {
	dropped := 0
	for _, lv := range shipped(t) {
		plain := fogSession(t, lv, dataSet(t), false)
		fog := fogSession(t, lv, dataSet(t), true)
		if plain.Fog() || !fog.Fog() {
			t.Fatalf("%s: Fog() %v and %v", lv.MapName, plain.Fog(), fog.Fog())
		}
		if !reflect.DeepEqual(plain.TrueView(), fog.TrueView()) {
			t.Errorf("%s: fog changes TrueView", lv.MapName)
		}
		evs := []Event{{Kind: Moved, Side: 1, Path: []hexgrid.Coord{{}}}}
		for k := range 2 {
			if plain.Sight(k) != nil || !reflect.DeepEqual(plain.EventsFor(k, evs), evs) {
				t.Errorf("%s: side %d has a sight without fog", lv.MapName, k)
			}
			pv, fv := plain.View(k), fog.View(k)
			if !reflect.DeepEqual(pv.Buildings, fv.Buildings) {
				t.Errorf("%s: fog changes side %d's buildings", lv.MapName, k)
			}
			sight := fog.Sight(k)
			for p, u := range pv.Units {
				got, ok := fv.Units[p]
				switch {
				case ok && !reflect.DeepEqual(got, u):
					t.Errorf("%s: side %d sees unit %d otherwise under fog", lv.MapName, k, u.Index)
				case !ok && u.Side == k:
					t.Errorf("%s: fog hides side %d's own unit %d", lv.MapName, k, u.Index)
				case !ok && (sight[p] || fog.InSight(k, p)):
					t.Errorf("%s: fog hides unit %d on %v in side %d's sight", lv.MapName, u.Index, p, k)
				case !ok:
					dropped++
				}
			}
			if len(fv.Units) > len(pv.Units) {
				t.Errorf("%s: side %d sees more under fog", lv.MapName, k)
			}
		}
	}
	t.Logf("%d units hidden", dropped)
	if dropped == 0 {
		t.Error("fog hid no unit on any shipped level")
	}
}

// TestFogTargets: with every unit's sight 0 (only the buildings look out),
// a fogged side's targets are its plain targets but those of hidden units,
// and an attack order on a hidden unit is refused.
func TestFogTargets(t *testing.T) {
	set := sighted(t, 0)
	refused := 0
	for _, lv := range shipped(t) {
		plain := fogSession(t, lv, set, false)
		fog := fogSession(t, lv, set, true)
		a := must(fog.Game.Action())
		view := fog.View(a)
		for _, i := range own(fog, a) {
			want, got := plain.Targets(a, i), fog.Targets(a, i)
			for p := range got {
				if !want[p] {
					t.Errorf("%s: unit %d targets %v only under fog", lv.MapName, i, p)
				}
			}
			for p := range want {
				if got[p] {
					continue
				}
				if _, seen := view.Units[p]; seen || fog.Sapper(i) {
					t.Errorf("%s: fog takes target %v of unit %d", lv.MapName, p, i)
				}
				_, err := fog.Attack(a, i, p)
				var r *dosrules.Refusal
				if !errors.As(err, &r) || r.Code != dosrules.MsgNoTarget {
					t.Errorf("%s: attack of unit %d on hidden %v: %v", lv.MapName, i, p, err)
				}
				refused++
			}
		}
	}
	t.Logf("%d targets hidden", refused)
	if refused == 0 {
		t.Error("no target was hidden on any shipped level")
	}
}

// TestFogReveal: with every unit's sight 2, a hidden enemy unit comes into
// view when a unit of the side moves within 2 hexes of it; the move shows
// to the other side only when it ends in that side's sight, and so does
// its reveal after the phase change.
func TestFogReveal(t *testing.T) {
	set := sighted(t, 2)
	found := 0
	for _, lv := range shipped(t) {
		s := fogSession(t, lv, set, true)
		mv := 1 - must(s.Game.Action())
		enemy, at, ok := hexgrid.Coord{}, -1, false
		var to hexgrid.Coord
		hidden := s.TrueView().Units
		for p, u := range hidden {
			if _, seen := s.View(mv).Units[p]; seen || u.Side != 1-mv {
				delete(hidden, p)
			}
		}
	search:
		for _, i := range own(s, mv) {
			d, err := s.Destinations(mv, i)
			if err != nil {
				continue
			}
			for _, c := range Coords(d) {
				for _, e := range Coords(toSet(hidden)) {
					if hexgrid.Distance(c, e) <= 2 {
						enemy, at, to, ok = e, i, c, true
						break search
					}
				}
			}
		}
		if !ok {
			continue
		}
		evs, err := s.Move(mv, at, to)
		if err != nil {
			t.Fatalf("%s: move of unit %d to %v: %v", lv.MapName, at, to, err)
		}
		if _, seen := s.View(mv).Units[enemy]; !seen {
			t.Errorf("%s: unit %d on %v does not see the enemy on %v", lv.MapName, at, to, enemy)
		}
		moves := func(evs []Event) int {
			n := 0
			for _, ev := range evs {
				if ev.Kind == Moved {
					n++
				}
			}
			return n
		}
		want := 0
		if s.InSight(1-mv, to) {
			want = moves(evs)
		}
		if got := moves(s.EventsFor(1-mv, evs)); moves(evs) != 1 || got != want {
			t.Errorf("%s: the other side gets %d of %d moves (goal in sight %v)", lv.MapName, got, moves(evs), want > 0)
		}
		if !reflect.DeepEqual(s.EventsFor(mv, evs), evs) {
			t.Errorf("%s: the mover does not see its own move", lv.MapName)
		}
		s.Confirm(0)
		s.Confirm(1)
		res, err := s.Resolve()
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range res {
			if ev.Kind == Revealed && !s.InSight(1-mv, ev.Path[len(ev.Path)-1]) {
				t.Errorf("%s: a move to %v out of sight is revealed", lv.MapName, ev.Path[len(ev.Path)-1])
			}
		}
		found++
	}
	t.Logf("%d levels with an enemy coming into view", found)
	if found == 0 {
		t.Error("no unit could come within sight of a hidden enemy on any shipped level")
	}
}

// TestFogSaveLoad: the fog option survives Save, Load and Clone, and a save
// without it loads without fog.
func TestFogSaveLoad(t *testing.T) {
	lv := shipped(t)[1]
	s := fogSession(t, lv, dataSet(t), true)
	cycle(t, s, approach)
	raw := must(s.Save())
	r, err := Load(lv, dataSet(t), raw)
	if err != nil {
		t.Fatal(err)
	}
	c := must(s.Clone())
	if !r.Fog() || !c.Fog() {
		t.Fatalf("fog after Load %v, after Clone %v", r.Fog(), c.Fog())
	}
	for k := range 2 {
		if !reflect.DeepEqual(r.View(k), s.View(k)) {
			t.Errorf("side %d sees another battle after Load", k)
		}
	}
	plain := must(fogSession(t, lv, dataSet(t), false).Save())
	if strings.Contains(string(plain), `"fog"`) {
		t.Error("a save without fog names it")
	}
	if p, err := Load(lv, dataSet(t), plain); err != nil || p.Fog() {
		t.Errorf("a save without fog loads with fog %v (%v)", p != nil && p.Fog(), err)
	}
}
