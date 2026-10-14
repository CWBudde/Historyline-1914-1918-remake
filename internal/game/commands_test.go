package game

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func key(c hexgrid.Coord) string { return fmt.Sprintf("%d,%d", c.Line, c.Col) }

func TestMovableMatchesPython(t *testing.T) {
	set := loadSet(t)
	runs := 0
	for _, run := range loadPython(t).Levels {
		s, err := NewWithRules(loadLevel(t, run.Level), set, Classic)
		if err != nil {
			t.Fatal(err)
		}
		s.Turn = run.Player
		for _, m := range run.Moves {
			from := hexgrid.Coord{Line: m.From[0], Col: m.From[1]}
			h := s.At(from)
			var u *Unit
			switch m.In {
			case "map":
				u = h.Unit
			case "building":
				u = h.Container.Slots[m.Row][m.Col]
			case "carrier":
				u = h.Unit.Content.Slots[m.Row][m.Col]
			}
			if u == nil || u.Type != m.Unit {
				t.Fatalf("%s: no %s %s at %v", run.Level, m.In, m.Unit, from)
			}
			got := map[string][][2]int{}
			for to, route := range s.Movable(from, u) {
				path := route.Path
				p := make([][2]int, len(path))
				for i, c := range path {
					p[i] = [2]int{c.Line, c.Col}
				}
				got[key(to)] = p
			}
			if !reflect.DeepEqual(got, m.Targets) && (len(got) > 0 || len(m.Targets) > 0) {
				t.Errorf("%s player %d %s %s at %v:\n got  %v\n want %v", run.Level, run.Player, m.In, m.Unit, from, got, m.Targets)
			}
			runs++
		}
	}
	if runs < 100 {
		t.Fatalf("only %d units compared", runs)
	}
}

func TestBattleMatchesPython(t *testing.T) {
	set := loadSet(t)
	for i, c := range loadPython(t).Battles {
		a, err := NewUnit(set, c.Attacker)
		if err != nil {
			t.Fatal(err)
		}
		d, err := NewUnit(set, c.Defender)
		if err != nil {
			t.Fatal(err)
		}
		a.Experience, a.Size, d.Experience, d.Size = c.AttackerExp, c.AttackerSize, c.DefenderExp, c.DefenderSize
		for _, slot := range c.DefenderEmpty {
			zero := 0
			d.Ammo[slot] = &zero
		}
		rolls := c.Rolls[:]
		roll := func() int { r := rolls[0]; rolls = rolls[1:]; return r }
		counter, f, e := resolve(Classic, a, d, c.Slot, c.Steps, c.AttackerBattle, c.DefenderBattle, roll)
		if counter != c.Counter || f != c.AttackerKills || e != c.DefenderKills {
			t.Errorf("case %d %+v: got counter %d kills %d/%d", i, c, counter, f, e)
		}
	}
}

// board builds a 6-wide, 12-line grass map played by the Central Powers, with
// the given hexes placed on it.
func board(t *testing.T, place map[hexgrid.Coord]level.Hex) *State {
	t.Helper()
	return boardWith(t, Standard, place)
}

// boardWith is board with the given rules.
func boardWith(t *testing.T, r Rules, place map[hexgrid.Coord]level.Hex) *State {
	t.Helper()
	lv := &level.Level{MapName: "test", MapNo: 1, Player: "Central Powers"}
	for l := range 12 {
		line := make([]level.Hex, 6-l%2)
		for c := range line {
			line[c] = level.Hex{Terrain: "grass"}
			if h, ok := place[hexgrid.Coord{Line: l, Col: c}]; ok {
				line[c] = h
			}
		}
		lv.Lines = append(lv.Lines, line)
	}
	s, err := NewWithRules(lv, loadSet(t), r)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func at(l, c int) hexgrid.Coord { return hexgrid.Coord{Line: l, Col: c} }

func unit(id string) level.Hex { return level.Hex{Terrain: "grass", Unit: id} }

func building(terrain string, owner Player, units ...string) level.Hex {
	return level.Hex{Terrain: terrain, Content: &level.Content{Owner: int(owner), Units: units}}
}

func TestMoveAndPath(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry")})
	u := s.At(at(4, 2)).Unit
	path, err := s.Move(at(4, 2), u, at(10, 2))
	if err != nil {
		t.Fatal(err)
	}
	if want := []hexgrid.Coord{at(4, 2), at(6, 2), at(8, 2), at(10, 2)}; !reflect.DeepEqual(path, want) {
		t.Errorf("path = %v, want %v", path, want)
	}
	if s.At(at(4, 2)).Unit != nil || s.At(at(10, 2)).Unit != u || !u.Moved {
		t.Error("unit not moved")
	}
	if s.Fog(CentralPowers, at(4, 2)) != Visible || s.Fog(CentralPowers, at(2, 2)) != Seen {
		t.Error("vision not updated after the move")
	}
	if _, err := s.Move(at(10, 2), u, at(8, 2)); err == nil {
		t.Error("moved a unit twice")
	}
}

func TestMoveErrors(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{
		at(4, 2): unit("CP_infantry"), at(4, 4): unit("CP_bunker"), at(10, 2): unit("EC_infantry"),
		at(0, 0): building("cmpN", Entente, "CP_infantry"), at(0, 4): building("cmpN", None, "CP_infantry"),
	})
	cases := map[string]struct {
		from hexgrid.Coord
		u    *Unit
		to   hexgrid.Coord
	}{
		"too far":          {at(4, 2), s.At(at(4, 2)).Unit, at(11, 2)},
		"onto own unit":    {at(4, 2), s.At(at(4, 2)).Unit, at(4, 4)},
		"speed 0":          {at(4, 4), s.At(at(4, 4)).Unit, at(6, 4)},
		"enemy unit":       {at(10, 2), s.At(at(10, 2)).Unit, at(8, 2)},
		"not there":        {at(6, 2), s.At(at(4, 2)).Unit, at(8, 2)},
		"from enemy depot": {at(0, 0), s.At(at(0, 0)).Container.Slots[0][0], at(2, 0)},
		"onto enemy unit":  {at(4, 2), s.At(at(4, 2)).Unit, at(10, 2)},
		"same hex":         {at(4, 2), s.At(at(4, 2)).Unit, at(4, 2)},
		"off the map":      {at(4, 2), s.At(at(4, 2)).Unit, at(-2, 2)},
		"nil unit":         {at(4, 2), nil, at(6, 2)},
	}
	for name, c := range cases {
		if _, err := s.Move(c.from, c.u, c.to); err == nil {
			t.Errorf("%s: Move succeeded", name)
		}
	}
	// A neutral depot's own-side units may leave (ContentMenu only bars the opponent's buildings).
	u := s.At(at(0, 4)).Container.Slots[0][0]
	if _, err := s.Move(at(0, 4), u, at(2, 4)); err != nil {
		t.Errorf("leave neutral depot: %v", err)
	}
	if len(s.At(at(0, 4)).Container.Units()) != 0 || s.At(at(2, 4)).Unit != u {
		t.Error("unit did not leave the depot")
	}
}

func TestCaptureAndVictory(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{
		at(4, 2): unit("CP_infantry"), at(4, 4): unit("CP_lArtillery"),
		at(6, 2): building("hqN", Entente), at(6, 4): building("cmpN", Entente),
	})
	art := s.At(at(4, 4)).Unit
	if _, ok := s.Movable(at(4, 4), art)[at(6, 4)]; ok {
		t.Error("a unit without the capture skill may enter an enemy depot")
	}
	inf := s.At(at(4, 2)).Unit
	if _, err := s.Move(at(4, 2), inf, at(6, 2)); err != nil {
		t.Fatal(err)
	}
	hq := s.At(at(6, 2))
	if hq.Owner != CentralPowers || hq.Container.Slots[0][0] != inf || s.At(at(4, 2)).Unit != nil {
		t.Errorf("HQ not captured: owner %v, contents %v", hq.Owner, ids(hq.Container.Units()))
	}
	if s.Winner != CentralPowers {
		t.Errorf("Winner = %v, want CentralPowers", s.Winner)
	}
	if _, err := s.Move(at(4, 4), art, at(6, 4)); err == nil {
		t.Error("moved after the game was won")
	}
	if err := s.EndTurn(); err == nil {
		t.Error("ended a turn after the game was won")
	}
}

func TestDepotCaptureWithoutVictory(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_cavalry"), at(8, 2): building("factoryE_w", Entente)})
	if _, err := s.Move(at(4, 2), s.At(at(4, 2)).Unit, at(8, 2)); err != nil {
		t.Fatal(err)
	}
	if s.At(at(8, 2)).Owner != CentralPowers || s.Winner != None {
		t.Errorf("owner %v, winner %v", s.At(at(8, 2)).Owner, s.Winner)
	}
}

func TestFullDepotLeavesUnitOnHex(t *testing.T) {
	sappers := []string{"CP_sappers", "CP_sappers", "CP_sappers", "CP_sappers", "CP_sappers", "CP_sappers"}
	s := boardWith(t, Classic, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): building("cmpN", CentralPowers, sappers...)})
	u := s.At(at(4, 2)).Unit
	if _, err := s.Move(at(4, 2), u, at(6, 2)); err != nil {
		t.Fatal(err)
	}
	if h := s.At(at(6, 2)); h.Unit != u || len(h.Container.Units()) != 6 {
		t.Errorf("unit on hex %v, depot holds %d", h.Unit, len(h.Container.Units()))
	}
}

func TestCarrier(t *testing.T) {
	s := boardWith(t, Classic, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): unit("CP_supplyCar"), at(5, 2): unit("CP_armCar")})
	inf, car := s.At(at(4, 2)).Unit, s.At(at(6, 2)).Unit
	if _, err := s.Move(at(4, 2), inf, at(6, 2)); err != nil {
		t.Fatal(err)
	}
	if car.Content.Slots[0][0] != inf || s.At(at(4, 2)).Unit != nil || s.At(at(6, 2)).Unit != car {
		t.Fatal("infantry not in the supply car")
	}
	// Python computes the exit from the carrier's own stats (speed 7, not 3).
	inf.Moved = false
	var to hexgrid.Coord
	for to = range s.Movable(at(6, 2), inf) {
		if hexgrid.Distance(at(6, 2), to) > inf.Def.Speed {
			break
		}
	}
	if hexgrid.Distance(at(6, 2), to) <= inf.Def.Speed {
		t.Fatal("no exit beyond infantry speed")
	}
	if _, err := s.Move(at(6, 2), inf, to); err != nil {
		t.Fatalf("leave carrier: %v", err)
	}
	if len(car.Content.Units()) != 0 || s.At(to).Unit != inf {
		t.Error("infantry did not leave the car")
	}
}

func TestCarrierFull(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_hArtillery"), at(6, 2): unit("CP_supplyCar"), at(8, 2): unit("CP_infantry")})
	if _, err := s.Move(at(4, 2), s.At(at(4, 2)).Unit, at(6, 2)); err != nil {
		t.Fatalf("weight 5 into an empty car of 5: %v", err)
	}
	if _, ok := s.Movable(at(8, 2), s.At(at(8, 2)).Unit)[at(6, 2)]; ok {
		t.Error("infantry may enter a full car")
	}
}

func TestTargets(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{
		at(4, 2): unit("CP_hArtillery"), at(6, 2): unit("EC_infantry"), at(4, 3): unit("CP_armCar"),
		at(6, 3): unit("EC_infantry"), at(10, 2): unit("EC_infantry"), at(11, 4): unit("CP_infantry"),
	})
	got := s.Targets(at(4, 2))
	if want := []hexgrid.Coord{at(6, 3), at(10, 2)}; !reflect.DeepEqual(got, want) {
		t.Errorf("Targets = %v, want %v (distance 1 is below the morser's minimum)", got, want)
	}
	if got := s.Targets(at(4, 3)); !reflect.DeepEqual(got, []hexgrid.Coord{at(6, 3)}) {
		t.Errorf("armoured car Targets = %v", got)
	}
	s.vision[CentralPowers] = vision{}
	if got := s.Targets(at(4, 3)); len(got) != 0 {
		t.Errorf("Targets without vision = %v", got)
	}
}

func TestAttack(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): unit("EC_infantry")})
	a, d := s.At(at(4, 2)).Unit, s.At(at(6, 2)).Unit
	rolls := []int{5, -5}
	roll := func() int { r := rolls[0]; rolls = rolls[1:]; return r }
	// gewehr98 30 - armour 20 + 5 = 15 → 1; leeEnfield fires back: 10 - 5 = 5 → 0.
	b, err := s.Attack(at(4, 2), at(6, 2), 2, roll)
	if err != nil {
		t.Fatal(err)
	}
	if want := (Battle{Counter: 2, Damage: 1}); b != want {
		t.Errorf("Battle = %+v, want %+v", b, want)
	}
	if d.Size != 9 || a.Size != 10 || a.Experience != 1 || d.Experience != 0 || !a.Attacked || a.Moved {
		t.Errorf("after battle: attacker %d/%d exp %d, defender %d exp %d", a.Size, a.MaxSize, a.Experience, d.Size, d.Experience)
	}

	// A kill removes the defender and gives two ranks, capped at 10.
	d.Size, a.Experience, a.Attacked = 1, 9, false
	rolls = []int{0, 0}
	b, err = s.Attack(at(4, 2), at(6, 2), 2, roll)
	if err != nil {
		t.Fatal(err)
	}
	if !b.DefenderDestroyed || s.At(at(6, 2)).Unit != nil || d.Size != 0 || a.Experience != 10 {
		t.Errorf("kill: %+v, hex unit %v, size %d, exp %d", b, s.At(at(6, 2)).Unit, d.Size, a.Experience)
	}
}

func TestAttackerDies(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): unit("EC_armTrain")})
	a, d := s.At(at(4, 2)).Unit, s.At(at(6, 2)).Unit
	a.Size, d.Experience = 1, 10
	rolls := []int{0, 5}
	b, err := s.Attack(at(4, 2), at(6, 2), 0, func() int { r := rolls[0]; rolls = rolls[1:]; return r })
	if err != nil {
		t.Fatal(err)
	}
	if !b.AttackerDestroyed || b.DefenderDestroyed || s.At(at(4, 2)).Unit != nil || d.Experience != 10 {
		t.Errorf("%+v, attacker hex %v, defender exp %d", b, s.At(at(4, 2)).Unit, d.Experience)
	}
}

func TestAttackErrors(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{
		at(4, 2): unit("CP_infantry"), at(6, 2): unit("EC_infantry"), at(2, 2): unit("CP_infantry"),
		at(8, 2): unit("EC_infantry"), at(0, 0): unit("CP_supplyCar"),
	})
	roll := func() int { return 0 }
	cases := map[string]struct {
		from, to hexgrid.Coord
		slot     int
	}{
		"empty slot":     {at(4, 2), at(6, 2), 3},
		"slot range":     {at(4, 2), at(6, 2), 4},
		"own unit":       {at(4, 2), at(2, 2), 0},
		"out of range":   {at(4, 2), at(8, 2), 0},
		"enemy attacker": {at(6, 2), at(4, 2), 0},
		"no attacker":    {at(5, 2), at(6, 2), 0},
		"no weapons":     {at(0, 0), at(6, 2), 0},
	}
	for name, c := range cases {
		if _, err := s.Attack(c.from, c.to, c.slot, roll); err == nil {
			t.Errorf("%s: Attack succeeded", name)
		}
	}
}

func TestEndTurn(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(10, 2): unit("EC_infantry")})
	cp, ec := s.At(at(4, 2)).Unit, s.At(at(10, 2)).Unit
	cp.Moved, cp.Attacked, ec.Moved = true, true, true
	if err := s.EndTurn(); err != nil {
		t.Fatal(err)
	}
	if s.Turn != Entente || s.Round != 2 || !cp.Moved || ec.Moved {
		t.Errorf("after CP: turn %v round %d, cp moved %v, ec moved %v", s.Turn, s.Round, cp.Moved, ec.Moved)
	}
	if _, err := s.Move(at(10, 2), ec, at(8, 2)); err != nil {
		t.Errorf("Entente move: %v", err)
	}
	if err := s.EndTurn(); err != nil {
		t.Fatal(err)
	}
	if s.Turn != CentralPowers || s.Round != 2 || cp.Moved || cp.Attacked || !ec.Moved {
		t.Errorf("after EC: turn %v round %d, cp %v/%v", s.Turn, s.Round, cp.Moved, cp.Attacked)
	}
}

func TestTargetsSorted(t *testing.T) {
	s := board(t, map[hexgrid.Coord]level.Hex{at(5, 2): unit("CP_armCar"), at(3, 2): unit("EC_infantry"), at(7, 2): unit("EC_infantry"), at(4, 3): unit("EC_infantry")})
	got := s.Targets(at(5, 2))
	if !sort.SliceIsSorted(got, func(i, j int) bool {
		return got[i].Line < got[j].Line || got[i].Line == got[j].Line && got[i].Col < got[j].Col
	}) || len(got) != 3 {
		t.Errorf("Targets = %v", got)
	}
}

func TestCounterWeaponTie(t *testing.T) {
	set := loadSet(t)
	a, _ := NewUnit(set, "CP_infantry")
	d, _ := NewUnit(set, "EC_infantry")
	d.Weapons[1] = d.Weapons[2] // two leeEnfields: Python's stable sort keeps the first
	if counter, _, _ := resolve(Classic, a, d, 0, 1, 4, 4, func() int { return 0 }); counter != 1 {
		t.Errorf("counter = %d, want slot 1", counter)
	}
}
