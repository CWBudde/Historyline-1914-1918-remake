package dosai

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// GRAND starts each side with one cavalry among neutral buildings: against a
// side that only confirms, the computer takes buildings first and several of
// them, whichever side and setup it plays.
func TestGrandTakesBuildings(t *testing.T) {
	for _, mode := range []scenario.Mode{scenario.CP, scenario.EC} {
		for side := range 2 {
			s := setupSession(t, "1918-09", mode)
			kinds := map[int]dosplay.BuildingKind{}
			for _, b := range s.TrueView().Buildings {
				kinds[b.Hold] = b.Kind
			}
			var taken []dosplay.Event
			var p [2]*Planner
			p[side] = &Planner{}
			end, err := playOut(s, p, 60, func(evs []dosplay.Event) {
				for _, ev := range evs {
					if ev.Kind == dosplay.Captured && ev.Side == side {
						taken = append(taken, ev)
					}
				}
			})
			if err != nil {
				t.Fatalf("%s side %d: %v", mode, side, err)
			}
			if len(taken) < 3 {
				t.Errorf("%s side %d: %d captures, want at least 3 (end %+v)", mode, side, len(taken), end)
				continue
			}
			for i, ev := range taken[:3] {
				if kinds[ev.Hold] == dosplay.HQ {
					t.Errorf("%s side %d: capture %d is the HQ", mode, side, i)
				}
			}
			t.Logf("%s side %d: %d captures in %d turns, end %+v", mode, side, len(taken), s.Turn(), end)
		}
	}
}

// planFor is a plan for side on a shipped setup's opening, its missions given.
func planFor(t *testing.T, month string, mode scenario.Mode, side int) *plan {
	t.Helper()
	s := setupSession(t, month, mode)
	pl, err := newPlan(s, side)
	if err != nil {
		t.Fatal(err)
	}
	if err := pl.assignMissions(); err != nil {
		t.Fatal(err)
	}
	return pl
}

// Depots and factories count alike and above the enemy HQ; the side's own
// buildings and neutral HQs are no goals.
func TestCaptureGoals(t *testing.T) {
	pl := planFor(t, "1918-02", scenario.CP, 1)
	goals := pl.captureGoals()
	if len(goals) == 0 {
		t.Fatal("no goals")
	}
	for i, g := range goals {
		b := pl.view.Buildings[g.at]
		want := buildingValue
		if b.HQ {
			want = hqValue
		}
		if b.Owner == pl.s.Faction(1) || b.HQ && b.Owner != pl.s.Faction(0) {
			t.Errorf("goal %v owned by %v (HQ %v)", g.at, b.Owner, b.HQ)
		}
		if g.value != want || i > 0 && goals[i-1].value < g.value {
			t.Errorf("goal %d %v: value %d, want %d, by value", i, g.at, g.value, want)
		}
	}
}

// The lone capturer on GRAND takes the building with the shortest route.
func TestMissionNearest(t *testing.T) {
	pl := planFor(t, "1918-09", scenario.CP, 0)
	if len(pl.missions) != 1 {
		t.Fatalf("%d missions, want 1", len(pl.missions))
	}
	for unit, m := range pl.missions {
		u := pl.s.Unit(unit)
		var at hexgrid.Coord
		for _, c := range sorted(pl.view.Units) {
			if pl.view.Units[c].Index == unit {
				at = c
			}
		}
		own, ok, err := pl.hops(u, m.goal, at)
		if err != nil || !ok {
			t.Fatalf("no route to the goal %v: %v", m.goal, err)
		}
		for _, c := range pl.prizes {
			if h, ok, err := pl.hops(u, c, at); err == nil && ok && h < own {
				t.Errorf("goal %v is %d hexes off, building %v only %d", m.goal, own, c, h)
			}
		}
	}
}

// Each capturer has one mission and no goal two capturers; only capturers on
// the map get one.
func TestMissionsOnePerGoal(t *testing.T) {
	pl := planFor(t, "1918-02", scenario.CP, 1)
	if len(pl.missions) == 0 {
		t.Fatal("no missions on RHEIN")
	}
	goals := map[hexgrid.Coord]int{}
	for unit, m := range pl.missions {
		if other, dup := goals[m.goal]; dup {
			t.Errorf("goal %v has units %d and %d", m.goal, other, unit)
		}
		goals[m.goal] = unit
		u := pl.s.Unit(unit)
		if u.Held {
			t.Errorf("unit %d is held and has a mission", unit)
		}
		if pl.profile.Tables.Move.Types[u.Type].Flags&captureFlag == 0 {
			t.Errorf("unit %d of type %d cannot capture", unit, u.Type)
		}
	}
}
