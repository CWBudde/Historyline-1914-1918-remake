package dosai

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// setupSession is a computer-against-computer session on a shipped setup.
func setupSession(t *testing.T, month string, mode scenario.Mode) *dosplay.Session {
	t.Helper()
	lv, err := scenario.LoadLevel("../../levels/"+month, mode)
	if err != nil {
		t.Fatalf("%s: %v", month, err)
	}
	return session(t, lv)
}

// toPhase confirms both sides and resolves until side is in phase.
func toPhase(t *testing.T, s *dosplay.Session, side int, phase int) {
	t.Helper()
	for range 4 {
		if int(s.Phase(side)) == phase {
			return
		}
		s.Confirm(0)
		s.Confirm(1)
		if _, err := s.Resolve(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatalf("side %d does not reach phase %d", side, phase)
}

// factories are the side's factory records with a build the computer may
// give, in record order.
func factories(s *dosplay.Session, side int) []int {
	var out []int
	v := s.View(side)
	for _, c := range sorted(v.Buildings) {
		b := v.Buildings[c]
		if b.Owner == s.Faction(side) && b.Kind == dosplay.Factory && len(allowedBuilds(s, side, b.Hold)) > 0 {
			out = append(out, b.Hold)
		}
	}
	slices.Sort(out)
	return out
}

// builds are the type of each unit a phase's commands build.
func builds(cmds []dosrules.Command) []int {
	var out []int
	for _, c := range cmds {
		if c.Kind == dosrules.CmdBuild {
			out = append(out, c.Type)
		}
	}
	return out
}

// planAction plans side's action phase with the given energy and returns
// its commands.
func planAction(t *testing.T, s *dosplay.Session, side, energy int) []dosrules.Command {
	t.Helper()
	toPhase(t, s, side, int(dosrules.PhaseAction))
	s.Game.Energy[side] = energy
	cmds, err := (&Planner{}).Commands(s, side)
	if err != nil {
		t.Fatal(err)
	}
	return cmds
}

// cheapest is the lowest price of the computer's builds in a record.
func cheapest(s *dosplay.Session, side, hold int) int {
	low := -1
	for _, o := range allowedBuilds(s, side, hold) {
		if low < 0 || o.Cost < low {
			low = o.Cost
		}
	}
	return low
}

// The computer's lists hold neither the construction unit nor a transporter,
// and each side's only its own types.
func TestComputerBuildLists(t *testing.T) {
	want := [2][]int{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 15, 17, 18, 19, 21},
		{10, 11, 12, 14, 15, 17, 18, 21, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50},
	}
	for side := range 2 {
		for typ := range 51 {
			if got := mayBuild(nil, side, typ); got != slices.Contains(want[side], typ) {
				t.Errorf("side %d, type %d: mayBuild = %v", side, typ, got)
			}
		}
	}
}

// pick stays in range and spreads over the candidates.
func TestPick(t *testing.T) {
	seen := map[int]bool{}
	for turn := range 50 {
		n := pick(turn, 1, dosrules.HoldFlag400, turn%7, 5)
		if n < 0 || n >= 5 {
			t.Fatalf("pick = %d", n)
		}
		seen[n] = true
	}
	if len(seen) != 5 {
		t.Errorf("picked only %v of 5", seen)
	}
	// The same values as the web edition's ai.test.ts.
	got := [4]int{pick(3, 0, 1024, 2, 9), pick(17, 1, 1030, 6, 23), pick(255, 1, 1, 0, 1000), pick(12345, 0, 4096, 5, 997)}
	if want := [4]int{3, 20, 698, 832}; got != want {
		t.Errorf("picks %v, want %v", got, want)
	}
}

// With the reserve alone nothing is built; with far more every factory fills,
// up to the cap, with types the computer may build.
func TestProductionReserveAndCap(t *testing.T) {
	const month, mode, side = "1918-02", scenario.Mode("cp"), 1
	s := setupSession(t, month, mode)
	if got := builds(planAction(t, s, side, reserve)); len(got) != 0 {
		t.Errorf("at the reserve: built %v", got)
	}

	s = setupSession(t, month, mode)
	got := builds(planAction(t, s, side, 100000))
	if len(got) == 0 || len(got) > maxBuilds {
		t.Fatalf("built %d units, want 1 to %d", len(got), maxBuilds)
	}
	for _, typ := range got {
		if !slices.Contains([]int{10, 11, 12, 14, 15, 17, 18, 19, 21}, typ) {
			t.Errorf("built type %d, outside RHEIN's list for side 1", typ)
		}
	}
}

// One unit's price above the reserve builds one unit in one factory; nothing
// dearer than the energy above the reserve is built.
func TestProductionBudget(t *testing.T) {
	s := setupSession(t, "1918-02", "cp")
	const side = 1
	toPhase(t, s, side, int(dosrules.PhaseAction))
	fs := factories(s, side)
	if len(fs) == 0 {
		t.Fatal("no factory")
	}
	price := cheapest(s, side, fs[0])
	types := s.Game.Profile.Tables.Types
	for _, typ := range builds(planAction(t, s, side, reserve+price)) {
		if types[typ].Cost > price {
			t.Errorf("built type %d for %d with %d above the reserve", typ, types[typ].Cost, price)
		}
	}
}

// Each factory plans from the phase's starting energy, so two factories
// together can spend into the reserve.
func TestProductionResetsPerFactory(t *testing.T) {
	const side = 1
	s := setupSession(t, "1918-02", "cp")
	toPhase(t, s, side, int(dosrules.PhaseAction))
	fs := factories(s, side)
	if len(fs) < 2 {
		t.Skip("the setup has fewer than two factories")
	}
	price := 0
	for _, f := range fs {
		price = max(price, cheapest(s, side, f))
	}
	got := builds(planAction(t, s, side, reserve+price))
	if len(got) < 2 {
		t.Errorf("built %v, want a unit from each of %d factories", got, len(fs))
	}
	if s.Energy(side) >= reserve {
		t.Errorf("energy %d: the factories did not spend into the reserve", s.Energy(side))
	}
}

// A damaged unit in an own building is repaired in record order while the
// phase started above the threshold, and not at or below it.
func TestRepairs(t *testing.T) {
	const side = 1
	damaged := func() (*dosplay.Session, []int) {
		s := setupSession(t, "1918-02", "cp")
		toPhase(t, s, side, int(dosrules.PhaseAction))
		var units []int
		v := s.View(side)
		for _, c := range sorted(v.Buildings) {
			if b := v.Buildings[c]; b.Owner == s.Faction(side) {
				for _, u := range b.Units {
					rec := s.Game.Units[u.Index]
					rec.Size = 1
					s.Game.Units[u.Index] = rec
					units = append(units, u.Index)
				}
			}
		}
		if len(units) == 0 {
			t.Skip("no unit in a building")
		}
		slices.Sort(units)
		return s, units
	}

	s, _ := damaged()
	for _, c := range planAction(t, s, side, repairFloor) {
		if c.Kind == dosrules.CmdRepair {
			t.Errorf("repair at the floor: %+v", c)
		}
	}

	s, units := damaged()
	n := 0
	last := -1
	for _, c := range planAction(t, s, side, 1000) {
		if c.Kind != dosrules.CmdRepair {
			continue
		}
		u := s.Game.Holds[c.Hold].Slots[7*side+c.Slot]
		if u <= last {
			t.Errorf("repaired unit %d after %d", u, last)
		}
		last = u
		n++
	}
	if n != len(units) {
		t.Errorf("repaired %d of %d", n, len(units))
	}
}

// The movement phase neither builds nor repairs.
func TestNoServiceInMovement(t *testing.T) {
	const side = 1
	s := setupSession(t, "1918-02", "cp")
	toPhase(t, s, side, int(dosrules.PhaseMovement))
	s.Game.Energy[side] = 100000
	cmds, err := (&Planner{}).Commands(s, side)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cmds {
		if c.Kind == dosrules.CmdBuild || c.Kind == dosrules.CmdRepair {
			t.Errorf("movement phase gave %+v", c)
		}
	}
}

// After a movement phase's orders no unit stays in an own building that had
// a hex to leave to, with enemies next to it or units in its doorway. Equal
// sessions plan alike, and the log replays to the same state.
func TestUnitsLeaveBuildings(t *testing.T) {
	cases := []struct {
		month string
		mode  scenario.Mode
	}{{"1914-10", "cp"}, {"1915-12", "cp"}, {"1916-04", "ec"}}
	for _, c := range cases {
		s := setupSession(t, c.month, c.mode)
		other := setupSession(t, c.month, c.mode)
		left := 0
		for range 30 {
			if s.Ended() != nil {
				break
			}
			for side := range 2 {
				movement := s.Phase(side) == dosrules.PhaseMovement
				mark := s.Mark()
				if _, err := (&Planner{}).Plan(s, side); err != nil {
					t.Fatal(err)
				}
				cmds := s.CommandsSince(mark)
				for _, cmd := range cmds {
					if cmd.Kind == dosrules.CmdExit {
						left++
					}
				}
				if _, err := other.Replay(cmds); err != nil {
					t.Fatal(err)
				}
				if movement {
					checkLeft(t, s, side, c.month)
				}
			}
			mark := s.Mark()
			if _, err := s.Resolve(); err != nil {
				t.Fatal(err)
			}
			if _, err := other.Replay(s.CommandsSince(mark)); err != nil {
				t.Fatal(err)
			}
		}
		if left == 0 {
			t.Errorf("%s: no unit left a building", c.month)
		}
		same(t, s, other, c.month)
	}
}

// checkLeft fails for a unit still in an own building of side that may get
// an order and has a free hex to leave to.
func checkLeft(t *testing.T, s *dosplay.Session, side int, month string) {
	t.Helper()
	pl, err := newPlan(s, side)
	if err != nil {
		t.Fatal(err)
	}
	if len(pl.enemies) == 0 && len(pl.hqs) == 0 {
		return
	}
	for _, c := range sorted(pl.view.Buildings) {
		b := pl.view.Buildings[c]
		if b.Owner != s.Faction(side) {
			continue
		}
		for _, v := range b.Units {
			if s.CanOrder(side, v.Index) != nil {
				continue
			}
			u := s.Unit(v.Index)
			dests, err := s.ExitDestinations(side, u.In, u.Slot)
			if err != nil {
				continue
			}
			for d := range dests {
				if _, taken := pl.view.Units[d]; !taken && d != c {
					t.Errorf("%s, turn %d: %s stays at %v with %v free", month, s.Turn(), u.Name, c, d)
					break
				}
			}
		}
	}
}

// A map's list holds both sides' types, each side builds its own part of it;
// a map without a list leaves the side's default.
func TestMapBuildList(t *testing.T) {
	list := []int{0, 9, 10, 19, 36, 50}
	for _, c := range []struct {
		side, typ int
		want      bool
	}{{0, 9, true}, {0, 10, true}, {0, 19, true}, {0, 36, false}, {1, 9, false}, {1, 10, true}, {1, 19, true}, {1, 50, true}, {1, 11, false}} {
		if got := mayBuild(list, c.side, c.typ); got != c.want {
			t.Errorf("side %d, type %d: mayBuild = %v", c.side, c.typ, got)
		}
	}
	if !mayBuild(nil, 1, 50) || mayBuild(nil, 1, 19) {
		t.Error("no list: not the side's default")
	}
}
