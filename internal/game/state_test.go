package game

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func loadSet(t *testing.T) *data.Set {
	t.Helper()
	s, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func loadLevel(t *testing.T, name string) *level.Level {
	t.Helper()
	l, err := level.Load(filepath.Join("..", "level", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func ids(us []*Unit) []string {
	out := []string{}
	for _, u := range us {
		out = append(out, u.Type)
	}
	return out
}

func TestNewShippedLevels(t *testing.T) {
	set := loadSet(t)
	paths, err := filepath.Glob(filepath.Join("..", "level", "testdata", "*.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("found %d level files (%v)", len(paths), err)
	}
	for _, path := range paths {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			lv := loadLevel(t, name)
			s, err := New(lv, set)
			if err != nil {
				t.Fatal(err)
			}
			if s.Grid != lv.Grid() || s.Name != lv.MapName || s.No != lv.MapNo {
				t.Errorf("header = %q %d %+v", s.Name, s.No, s.Grid)
			}
			n, units, containers := 0, 0, 0
			for i, line := range lv.Lines {
				for col, lh := range line {
					n++
					h := s.At(hexgrid.Coord{Line: i, Col: col})
					if h == nil {
						t.Fatalf("At(%d,%d) = nil", i, col)
					}
					if h.Terrain != lh.Terrain || !reflect.DeepEqual(h.Infra, lh.Infra.IDs) {
						t.Errorf("(%d,%d) = %q %v, level %q %v", i, col, h.Terrain, h.Infra, lh.Terrain, lh.Infra.IDs)
					}
					if lh.Unit == "" {
						if h.Unit != nil {
							t.Errorf("(%d,%d) has unit %q, level none", i, col, h.Unit.Type)
						}
					} else {
						units++
						if h.Unit == nil || h.Unit.Type != lh.Unit {
							t.Errorf("(%d,%d) unit = %v, level %q", i, col, h.Unit, lh.Unit)
						} else if h.Unit.Owner != FactionOf(set.Units[lh.Unit].Country) {
							t.Errorf("(%d,%d) unit owner %v", i, col, h.Unit.Owner)
						}
					}
					if lh.Content == nil {
						continue
					}
					containers++
					c := h.Container
					if c == nil {
						t.Errorf("(%d,%d) %s has content but no container", i, col, lh.Terrain)
						continue
					}
					if h.Owner != Player(lh.Content.Owner) {
						t.Errorf("(%d,%d) owner = %v, level %d", i, col, h.Owner, lh.Content.Owner)
					}
					want := 40
					if strings.HasPrefix(lh.Terrain, "hq") {
						want = 50
					}
					if c.Capacity != want {
						t.Errorf("(%d,%d) %s capacity = %d, want %d", i, col, lh.Terrain, c.Capacity, want)
					}
					got, exp := ids(c.FillOrder()), lh.Content.Units
					if exp == nil {
						exp = []string{}
					}
					if !reflect.DeepEqual(got, exp) {
						t.Errorf("(%d,%d) contents = %v, level %v", i, col, got, exp)
					}
					for _, u := range c.Units() {
						if u.Owner != FactionOf(u.Def.Country) {
							t.Errorf("(%d,%d) contained %s owner %v", i, col, u.Type, u.Owner)
						}
					}
				}
			}
			if n != s.Grid.Size() {
				t.Errorf("level has %d hexes, Grid.Size() = %d", n, s.Grid.Size())
			}
			if units == 0 || containers == 0 {
				t.Errorf("level has %d units and %d containers", units, containers)
			}
			if s.Human != PlayerByNameMust(t, lv.Player) {
				t.Errorf("Human %v", s.Human)
			}
		})
	}

	s, err := New(loadLevel(t, "pulse.json"), set)
	if err != nil {
		t.Fatal(err)
	}
	for col, h := range s.Hexes[5] {
		if h.Terrain == "hqN" && (h.Owner != CentralPowers || h.Container == nil) {
			t.Errorf("level1 line6 HQ col %d: owner %v container %v", col, h.Owner, h.Container)
		}
	}
	if s.Human != CentralPowers {
		t.Errorf("level1 Human = %v", s.Human)
	}
}

func PlayerByNameMust(t *testing.T, name string) Player {
	t.Helper()
	p, err := PlayerByName(name)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestUnitDefaults(t *testing.T) {
	set := loadSet(t)
	inf, err := NewUnit(set, "CP_infantry")
	if err != nil {
		t.Fatal(err)
	}
	if inf.Size != 6 || inf.MaxSize != 6 || inf.Experience != 0 || inf.Fuel != 30 || inf.Moved || inf.Attacked {
		t.Errorf("infantry = %+v", inf)
	}
	if inf.Owner != CentralPowers || inf.Content != nil {
		t.Errorf("infantry owner %v content %v", inf.Owner, inf.Content)
	}
	if inf.Ammo[0] != nil || inf.Ammo[1] == nil || *inf.Ammo[1] != 5 || *inf.Ammo[2] != 3 || inf.Ammo[3] != nil {
		t.Errorf("infantry ammo = %v", inf.Ammo)
	}
	*inf.Ammo[1] = 0
	if *set.Weapons["lugerp08"].Ammo != 5 {
		t.Fatal("unit ammo shares memory with the weapon table")
	}
	other, _ := NewUnit(set, "CP_infantry")
	if *other.Ammo[1] != 5 {
		t.Fatal("two units share ammo")
	}

	car, err := NewUnit(set, "EC_supplyCar")
	if err != nil {
		t.Fatal(err)
	}
	if car.Owner != Entente || car.Content == nil || car.Content.Capacity != 9 || car.Fuel != 50 {
		t.Errorf("supply car owner %v content %+v fuel %d", car.Owner, car.Content, car.Fuel)
	}
	if _, err := NewUnit(set, "nope"); err == nil {
		t.Error("NewUnit accepted an unknown id")
	}
	for country, want := range map[string]Player{
		"Germany": CentralPowers, "Austria": CentralPowers,
		"Bulgaria": CentralPowers, "Ottoman": CentralPowers, "France": Entente, "Britain": Entente,
	} {
		if got := FactionOf(country); got != want {
			t.Errorf("FactionOf(%s) = %v, want %v", country, got, want)
		}
	}
}

func TestContainerOrder(t *testing.T) {
	set := loadSet(t)
	unit := func(id string) *Unit {
		u, err := NewUnit(set, id)
		if err != nil {
			t.Fatal(err)
		}
		return u
	}
	c := &Container{Capacity: 4} // infantry 1 + cavalry 2 + elite infantry 1
	a, b, d := unit("CP_infantry"), unit("CP_cavalry"), unit("CP_eliteInf")
	for _, u := range []*Unit{a, b, d} {
		if err := c.Add(u); err != nil {
			t.Fatal(err)
		}
	}
	if c.Slots[0][0] != a || c.Slots[1][0] != b || c.Slots[0][1] != d {
		t.Fatalf("fill order: %v", c.Slots)
	}
	if got := ids(c.Units()); !reflect.DeepEqual(got, []string{"CP_infantry", "CP_eliteInf", "CP_cavalry"}) {
		t.Errorf("Units() = %v", got)
	}
	if got := ids(c.FillOrder()); !reflect.DeepEqual(got, []string{"CP_infantry", "CP_cavalry", "CP_eliteInf"}) {
		t.Errorf("FillOrder() = %v", got)
	}

	full := &Container{Capacity: 100}
	for i := range 18 {
		if err := full.Add(unit("CP_infantry")); err != nil {
			t.Fatalf("Add %d: %v", i+1, err)
		}
	}
	if err := full.Add(unit("CP_infantry")); err == nil {
		t.Error("19th Add succeeded")
	}
}

// Any factory piece holding units is a container with the factory's
// capacity, a piece without content is none.
func TestFactoryNWithContent(t *testing.T) {
	lv := &level.Level{MapName: "m", Player: "Central Powers", Lines: [][]level.Hex{
		{{Terrain: "factoryN_w", Content: &level.Content{Owner: 1, Units: []string{"CP_infantry", "CP_cavalry"}}}, {Terrain: "factoryN"}},
		{{Terrain: "grass", Unit: "EC_infantry"}},
	}}
	s, err := New(lv, loadSet(t))
	if err != nil {
		t.Fatal(err)
	}
	if h := s.Hexes[0][0]; h.Container == nil || h.Owner != CentralPowers || h.Container.Capacity != 40 || len(h.Container.Units()) != 2 {
		t.Errorf("factoryN_w with content: owner %v container %+v", h.Owner, h.Container)
	}
	if h := s.Hexes[0][1]; h.Container != nil || h.Owner != None {
		t.Errorf("plain factoryN: owner %v container %+v", h.Owner, h.Container)
	}
}

func TestUnitsOrder(t *testing.T) {
	set := loadSet(t)
	lv := &level.Level{MapName: "m", Player: "Entente Cordial", Lines: [][]level.Hex{
		{{Terrain: "grass", Unit: "CP_supplyCar"}, {Terrain: "cmpN", Content: &level.Content{Owner: 2, Units: []string{"EC_infantry", "EC_cavalry", "EC_eliteInf"}}}},
		{{Terrain: "grass", Unit: "EC_infantry"}},
	}}
	s, err := New(lv, set)
	if err != nil {
		t.Fatal(err)
	}
	car := s.Hexes[0][0].Unit
	if err := car.Content.Add(s.Hexes[0][1].Container.Slots[0][0]); err != nil {
		t.Fatal(err)
	}
	// Map order; contents row 0 before row 1.
	want := []string{"CP_supplyCar", "EC_infantry", "EC_infantry", "EC_eliteInf", "EC_cavalry", "EC_infantry"}
	if got := ids(s.Units(None)); !reflect.DeepEqual(got, want) {
		t.Errorf("Units(None) = %v, want %v", got, want)
	}
	if got := ids(s.Units(CentralPowers)); !reflect.DeepEqual(got, []string{"CP_supplyCar"}) {
		t.Errorf("Units(CP) = %v", got)
	}
	if s.Human != Entente {
		t.Errorf("Human %v", s.Human)
	}
}

func TestNewRejects(t *testing.T) {
	set := loadSet(t)
	cases := map[string]*level.Level{
		"unknown unit": {Player: "Central Powers", Lines: [][]level.Hex{{{Terrain: "grass", Unit: "nope"}}}},
		"unknown contained unit": {Player: "Central Powers", Lines: [][]level.Hex{{{
			Terrain: "cmpN",
			Content: &level.Content{Owner: 1, Units: []string{"nope"}},
		}}}},
		"unknown terrain":  {Player: "Central Powers", Lines: [][]level.Hex{{{Terrain: "nope"}}}},
		"content on grass": {Player: "Central Powers", Lines: [][]level.Hex{{{Terrain: "grass", Content: &level.Content{Owner: 1}}}}},
		"bad owner":        {Player: "Central Powers", Lines: [][]level.Hex{{{Terrain: "cmpN", Content: &level.Content{Owner: 3}}}}},
		"unknown player":   {Player: "Nobody", Lines: [][]level.Hex{{{Terrain: "grass"}}}},
		"too many units": {Player: "Central Powers", Lines: [][]level.Hex{{{
			Terrain: "hqN",
			Content: &level.Content{Owner: 1, Units: strings.Split(strings.Repeat("CP_infantry,", 19)[:19*12-1], ",")},
		}}}},
	}
	for name, lv := range cases {
		if _, err := New(lv, set); err == nil {
			t.Errorf("%s: New accepted the level", name)
		}
	}
}
