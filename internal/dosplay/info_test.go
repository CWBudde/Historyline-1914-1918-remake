package dosplay

import (
	"fmt"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// original is a session on imported original map i and the map itself.
func original(t *testing.T, i int, o Options) (*Session, *finmap.Level) {
	t.Helper()
	raw, err := levels.Original.ReadFile(fmt.Sprintf("original/full/%02d.json", i))
	if err != nil {
		t.Fatal(err)
	}
	l, err := finmap.Parse(raw)
	if err != nil {
		t.Fatalf("map %02d: %v", i, err)
	}
	b, err := dosbind.FromOriginal(l, dataSet(t))
	if err != nil {
		t.Fatalf("map %02d: %v", i, err)
	}
	lv, err := origview.Level(l)
	if err != nil {
		t.Fatalf("map %02d: %v", i, err)
	}
	s, err := NewBoard(b, lv, dataSet(t), o)
	if err != nil {
		t.Fatalf("map %02d: %v", i, err)
	}
	return s, l
}

// civil is a session on the shipped level with factories.
func civil(t *testing.T) *Session {
	t.Helper()
	var lv *level.Level
	for _, l := range shipped(t) {
		if l.MapName == "Civil" {
			lv = l
		}
	}
	return session(t, lv, Options{})
}

// toAction resolves the opening phase, so side 0 is in its action phase,
// where repairs and builds are given.
func toAction(t *testing.T, s *Session) {
	t.Helper()
	s.Confirm(0)
	s.Confirm(1)
	if _, err := s.Resolve(); err != nil || s.Phase(0) != dosrules.PhaseAction {
		t.Fatalf("resolve: %v, side 0 in phase %d", err, s.Phase(0))
	}
}

// The unit info of a light artillery as the original shows it: weapon power
// and range against air, ground and water, armour, speed, weight and group.
func TestTypeInfo(t *testing.T) {
	s := session(t, shipped(t)[0], Options{})
	got, ok := s.TypeInfo(0, 10)
	want := TypeView{Type: 10, ID: "CP_lArtillery", Name: "Light Artillery", Air: Weapon{0, 0}, Ground: Weapon{45, 3},
		Water: Weapon{45, 3}, Armour: 25, Speed: 2, Weight: 3, Group: 6, Cost: 55, Domain: Artillery}
	if !ok || got != want {
		t.Errorf("TypeInfo(0, 10) = %+v, %v\nwant %+v", got, ok, want)
	}
	if got, ok := s.TypeInfo(1, 10); !ok || got.ID != "EC_lArtillery" {
		t.Errorf("TypeInfo(1, 10) = %+v, %v", got, ok)
	}
	for _, typ := range []int{-1, len(s.Game.Profile.Tables.Types)} {
		if _, ok := s.TypeInfo(0, typ); ok {
			t.Errorf("TypeInfo(0, %d) is a type", typ)
		}
	}
	// Every unit's type: the view's armour and group, the rules' weapons with
	// the range shown one less than stored.
	for _, i := range own(s, 0) {
		u := s.Unit(i)
		v, ok := s.TypeInfo(u.Side, u.Type)
		rt := s.Game.Profile.Tables.Types[u.Type]
		if !ok || v.Name != u.Name || v.Armour != u.Armour || v.Group != u.Full ||
			v.Air != (Weapon{rt.AirPower, rt.AirRange - 1}) || v.Ground.Range != rt.SurfaceRange-1 || v.Water.Range != v.Ground.Range {
			t.Errorf("unit %d (%s): %+v", i, u.ID, v)
		}
	}
}

// Building kinds follow the container record: the imported maps name them.
func TestBuildingKind(t *testing.T) {
	for i := range 72 {
		if testing.Short() && i%6 != 0 {
			continue
		}
		s, l := original(t, i, Options{})
		kinds := map[int]string{}
		for _, b := range l.Buildings {
			kinds[b.Container] = b.Kind
		}
		v := s.TrueView()
		for p, b := range v.Buildings {
			if b.Kind.String() != kinds[b.Hold] || b.HQ != (b.Kind == HQ) {
				t.Errorf("map %02d: building %d at %v: kind %v, map %q", i, b.Hold, p, b.Kind, kinds[b.Hold])
			}
		}
	}
}

// The game info at the start of PULSE (original map 00) as the original
// shows it to side 0 against the computer.
func TestStatsPulse(t *testing.T) {
	s, _ := original(t, 0, Options{Computer: [2]bool{false, true}})
	got := s.Stats()
	want := Stats{Round: 0, Level: "00", LevelNo: 1, Mode: [2]int{dosrules.PhaseMovement, dosrules.PhaseAction}, Score: 861,
		Sides: [3]Holdings{{Units: 14}, {Units: 14}, {Units: 2, Depots: 2}}, Map: Holdings{Units: 30, Depots: 2},
		Income: [2]int{5, 5}}
	if !equalStats(got, want) {
		t.Errorf("Stats = %+v\nwant %+v", got, want)
	}
}

func equalStats(a, b Stats) bool {
	return a.Round == b.Round && a.Level == b.Level && a.LevelNo == b.LevelNo && a.Mode == b.Mode && a.Score == b.Score &&
		a.Sides == b.Sides && a.Map == b.Map && a.Energy == b.Energy && a.Income == b.Income && a.Orders == b.Orders &&
		a.Limit == b.Limit
}

// Stats counts what the true state holds, sums the sides into the map, pays
// the income it shows at the phase change and counts the orders given.
func TestStats(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{Limit: 8})
		st := s.Stats()
		var want [3]Holdings
		for _, u := range s.Game.Units {
			if !u.Destroyed() {
				want[sideOf(u.Flags4)].Units++
			}
		}
		for _, b := range s.TrueView().Buildings {
			k := s.SideOf(b.Owner)
			switch b.Kind {
			case Depot:
				want[k].Depots++
			case Factory:
				want[k].Factories++
			}
		}
		sum := Holdings{}
		for _, h := range want {
			sum.Units, sum.Depots, sum.Factories = sum.Units+h.Units, sum.Depots+h.Depots, sum.Factories+h.Factories
		}
		if st.Sides != want || st.Map != sum || st.Level != lv.MapName || st.LevelNo != lv.MapNo || st.Limit != 8 ||
			st.Score != s.Game.Score() || st.Round != s.Turn() || st.Mode != [2]int{s.Phase(0), s.Phase(1)} {
			t.Errorf("%s: %+v, want holdings %+v, map %+v", lv.MapName, st, want, sum)
		}
		if st.Income[0] == 0 || st.Income[1] == 0 {
			t.Errorf("%s: income %v", lv.MapName, st.Income)
		}

		moved := 0
		for _, i := range own(s, 0) {
			if d, err := s.Destinations(0, i); err == nil && len(d) > 0 {
				if _, err := s.Move(0, i, Coords(d)[0]); err != nil {
					t.Fatal(err)
				}
				if moved++; moved == 2 {
					break
				}
			}
		}
		if got := s.Stats().Orders; got != [2]int{moved, 0} {
			t.Errorf("%s: orders %v after %d moves", lv.MapName, got, moved)
		}

		// Each phase change pays every side the income shown before it.
		for range 3 {
			before := s.Stats()
			s.Confirm(0)
			s.Confirm(1)
			if _, err := s.Resolve(); err != nil {
				t.Fatal(err)
			}
			after := s.Stats()
			for k := range 2 {
				if after.Energy[k] != before.Energy[k]+before.Income[k] || after.Energy[k] != s.Energy(k) {
					t.Errorf("%s: side %d energy %d -> %d with income %d", lv.MapName, k, before.Energy[k], after.Energy[k], before.Income[k])
				}
			}
			if after.Orders != [2]int{} {
				t.Errorf("%s: orders %v after the phase change", lv.MapName, after.Orders)
			}
		}
	}
}

// CanRepair and CanBuild check without changing the state and agree with
// Repair and Build; Buildable lists a factory's types with their prices.
func TestCanRepairAndBuild(t *testing.T) {
	s := civil(t)
	toAction(t, s)
	repaired, built := false, false
	for _, p := range Coords(toSet(s.View(0).Buildings)) {
		hold, slots, ok := s.Contents(0, p)
		if !ok {
			continue
		}
		if !repaired && len(slots) > 0 {
			k := slots[0].Slot
			s.Game.Energy[0] = 0
			if err := s.CanRepair(0, hold, k); err == nil {
				t.Errorf("CanRepair without energy")
			}
			s.Game.Energy[0] = 1000
			v := s.Version()
			if err := s.CanRepair(0, hold, k); err != nil || s.Version() != v || s.Energy(0) != 1000 {
				t.Errorf("CanRepair: %v, version %d -> %d, energy %d", err, v, s.Version(), s.Energy(0))
			}
			for _, bad := range []struct{ side, hold, slot int }{{1, hold, k}, {2, hold, k}, {0, hold, 7}, {0, 999, k}} {
				if err := s.CanRepair(bad.side, bad.hold, bad.slot); err == nil {
					t.Errorf("CanRepair(%d, %d, %d) passed", bad.side, bad.hold, bad.slot)
				}
			}
			if _, err := s.Repair(0, hold, k); err != nil {
				t.Fatal(err)
			}
			if err := s.CanRepair(0, hold, k); err == nil {
				t.Errorf("CanRepair of a unit with its order")
			}
			repaired = true
		}
		opts := s.Buildable(0, hold)
		if built || len(opts) == 0 {
			continue
		}
		var types []int
		for _, o := range opts {
			types = append(types, o.Type)
			tv, _ := s.TypeInfo(0, o.Type)
			if o.Cost != s.Game.Profile.Tables.Types[o.Type].Cost || o.Cost == 0xFF || o.Name == "" || o.Name != tv.Name {
				t.Errorf("Buildable(%d): %+v", hold, o)
			}
		}
		if !slices.Equal(types, s.Game.Buildings[hold].Builds) {
			t.Errorf("Buildable(%d) = %v, builds %v", hold, types, s.Game.Buildings[hold].Builds)
		}
		o := opts[0]
		s.Game.Energy[0] = o.Cost - 1
		if err := s.CanBuild(0, hold, o.Type); err == nil {
			t.Errorf("CanBuild with %d energy for a price of %d", o.Cost-1, o.Cost)
		}
		s.Game.Energy[0] = o.Cost
		if err := s.CanBuild(0, hold, o.Type); err != nil || s.Energy(0) != o.Cost {
			t.Errorf("CanBuild: %v", err)
		}
		if err := s.CanBuild(0, hold, 1); err == nil {
			t.Errorf("CanBuild of a type the factory does not build")
		}
		if err := s.CanBuild(1, hold, o.Type); err == nil {
			t.Errorf("CanBuild in the other side's factory")
		}
		if _, err := s.Build(0, hold, o.Type); err != nil || s.Energy(0) != 0 {
			t.Errorf("Build: %v, energy %d", err, s.Energy(0))
		}
		built = true
	}
	if !repaired || !built {
		t.Errorf("repaired %v, built %v", repaired, built)
	}
	// A building that only repairs builds nothing; neither does a record that
	// is no building.
	for _, b := range s.TrueView().Buildings {
		if b.Kind != Factory && len(s.Buildable(0, b.Hold)) != 0 {
			t.Errorf("Buildable(%d) of a %v", b.Hold, b.Kind)
		}
	}
	if len(s.Buildable(0, -1)) != 0 || len(s.Buildable(0, 0)) != 0 || len(s.Buildable(2, 82)) != 0 {
		t.Error("Buildable of a record that is no building")
	}
}

// ContentSlots lists all of a side's slots of its buildings and
// transporters, empty ones too, and agrees with Contents.
func TestContentSlots(t *testing.T) {
	s := civil(t)
	empty, held, carriers := 0, 0, 0
	v := s.TrueView()
	hexes := Coords(toSet(v.Buildings))
	for _, p := range Coords(toSet(v.Units)) {
		if v.Units[p].Carrier {
			hexes = append(hexes, p)
		}
	}
	for _, p := range hexes {
		for side := range 3 {
			hold, slots, ok := s.ContentSlots(side, p)
			ch, units, cok := s.Contents(side, p)
			if ok != cok || hold != ch {
				t.Fatalf("%v, side %d: ContentSlots %d, %v; Contents %d, %v", p, side, hold, ok, ch, cok)
			}
			if !ok {
				continue
			}
			if b, isB := v.Buildings[p]; !isB {
				carriers++
			} else if s.SideOf(b.Owner) != side {
				t.Errorf("%v: side %d sees the slots of a building of %v", p, side, b.Owner)
			}
			if len(slots) != Slots {
				t.Fatalf("%v: %d slots", p, len(slots))
			}
			var occupied []UnitView
			for k, sv := range slots {
				u := s.Game.Holds[hold].Slots[7*side+k]
				switch {
				case sv.Slot != k:
					t.Errorf("%v: slot %d has index %d", p, k, sv.Slot)
				case sv.Unit == nil:
					empty++
					if u <= dosrules.SlotFree {
						t.Errorf("%v: slot %d holds unit %d, shown empty", p, k, u)
					}
				default:
					held++
					occupied = append(occupied, *sv.Unit)
					if sv.Unit.Index != u || sv.Unit.Slot != k || sv.Unit.In != hold || !sv.Unit.Held {
						t.Errorf("%v: slot %d: %+v", p, k, *sv.Unit)
					}
				}
			}
			if !slices.EqualFunc(occupied, units, func(a, b UnitView) bool { return a == b }) {
				t.Errorf("%v: the occupied slots %v differ from Contents %v", p, occupied, units)
			}
		}
	}
	if empty == 0 || held == 0 {
		t.Errorf("%d empty and %d occupied slots checked", empty, held)
	}
	t.Logf("%d empty and %d occupied slots, %d transporters", empty, held, carriers)
	if _, _, ok := s.ContentSlots(0, hexgrid.Coord{Line: 999}); ok {
		t.Error("ContentSlots off the level")
	}
}
