package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

// python is testdata/python.json, written by testdata/gen_python.py from
// hlrData.py: the data tables and the results of the infrastructure loop in
// HexSquare.__init__.
type python struct {
	Terrain map[string][3]*int `json:"terrain"`
	Units   map[string]struct {
		Name         string    `json:"name"`
		Country      string    `json:"country"`
		Armour       int       `json:"armour"`
		Speed        int       `json:"speed"`
		Weight       int       `json:"weight"`
		StorageMax   int       `json:"storageMax"`
		Sight        int       `json:"sight"`
		Fuel         int       `json:"fuel"`
		Exp          int       `json:"exp"`
		Skills       []int     `json:"skills"`
		MovementType int       `json:"movementType"`
		Weapons      []*string `json:"weapons"`
	} `json:"units"`
	Weapons map[string]struct {
		Name     string `json:"name"`
		RangeMin int    `json:"rangeMin"`
		RangeMax int    `json:"rangeMax"`
		Power    int    `json:"power"`
		Air      bool   `json:"air"`
		Ground   bool   `json:"ground"`
		Water    bool   `json:"water"`
		Ammo     *int   `json:"ammo"`
	} `json:"weapons"`
	Infra []struct {
		Terrain         string   `json:"terrain"`
		Infra           []string `json:"infra"`
		Movement        *int     `json:"movement"`
		Battle          *int     `json:"battle"`
		Sight           *int     `json:"sight"`
		MovementAllowed []int    `json:"movementAllowed"`
	} `json:"infra"`
}

func loadPython(t *testing.T) python {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "python.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p python
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func loadSet(t *testing.T) *Set {
	t.Helper()
	s, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func eqInt(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func show(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

var summerFactoryAliases = map[string]string{
	"factoryN": "factoryN_w",
	"factoryE": "factoryE_w",
	"factoryS": "factoryS_w",
	"factoryW": "factoryW_w",
}

// waterDepths are the depths of open water the original maps tell apart (the
// middle band and the open sea), which play as water.
var waterDepths = []string{"waterDeep", "waterSea"}

func TestWaterDepthsPlayAsWater(t *testing.T) {
	s := loadSet(t)
	for _, id := range waterDepths {
		if got, ok := s.Terrain[id]; !ok || !reflect.DeepEqual(got, s.Terrain["water"]) {
			t.Errorf("terrain %q = %+v, %v; want water's %+v", id, got, ok, s.Terrain["water"])
		}
	}
}

func TestSummerFactoryAliasesMatchWinterModifiers(t *testing.T) {
	s := loadSet(t)
	for summer, winter := range summerFactoryAliases {
		t.Run(summer, func(t *testing.T) {
			got, ok := s.Terrain[summer]
			if !ok {
				t.Fatalf("summer factory terrain %q missing", summer)
			}
			want, ok := s.Terrain[winter]
			if !ok {
				t.Fatalf("winter factory terrain %q missing", winter)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("summer %q modifiers %+v differ from winter %q: %+v", summer, got, winter, want)
			}
		})
	}
}

func TestTerrainUnitsWeaponsMatchPython(t *testing.T) {
	p, s := loadPython(t), loadSet(t)

	// Summer factory art adds aliases whose modifiers are checked against the
	// winter records separately, as are the water depths against water; every
	// original terrain value remains pinned.
	originalTerrainIDs := 0
	for id := range s.Terrain {
		if _, alias := summerFactoryAliases[id]; !alias && !slices.Contains(waterDepths, id) {
			originalTerrainIDs++
		}
	}
	if originalTerrainIDs != len(p.Terrain) {
		t.Errorf("terrain: %d original ids, Python has %d", originalTerrainIDs, len(p.Terrain))
	}
	for id, want := range p.Terrain {
		got, ok := s.Terrain[id]
		if !ok {
			t.Errorf("terrain %q missing", id)
			continue
		}
		if !eqInt(got.Movement, want[0]) || !eqInt(got.Battle, want[1]) || !eqInt(got.Sight, want[2]) {
			t.Errorf("terrain %q = [%v %v %v], Python [%v %v %v]", id,
				show(got.Movement), show(got.Battle), show(got.Sight), show(want[0]), show(want[1]), show(want[2]))
		}
	}

	if len(p.Units) != 27 {
		t.Errorf("units: Python has %d, want 27", len(p.Units))
	}
	// Stats and attacks come from UNIT.DAT (TestUnitsFollowUnitDat); the rest is
	// Python's for the units it has. The others are in TestRosterCoversUnitDat.
	for id, want := range p.Units {
		got, ok := s.Units[id]
		if !ok {
			t.Errorf("unit %q missing", id)
			continue
		}
		if got.Country != want.Country || got.Sight != want.Sight || got.Fuel != want.Fuel || got.Exp != want.Exp ||
			!reflect.DeepEqual(got.Skills, want.Skills) || got.MovementType != want.MovementType {
			t.Errorf("unit %q = %+v, Python %+v", id, got, want)
		}
	}

	if len(s.Weapons) != 15 || len(p.Weapons) != 15 {
		t.Errorf("weapons: Go has %d, Python has %d, want 15", len(s.Weapons), len(p.Weapons))
	}
	for id, want := range p.Weapons {
		got, ok := s.Weapons[id]
		if !ok {
			t.Errorf("weapon %q missing", id)
			continue
		}
		if got.Name != want.Name || got.RangeMin != want.RangeMin || got.RangeMax != want.RangeMax ||
			got.Power != want.Power || got.Air != want.Air || got.Ground != want.Ground ||
			got.Water != want.Water || !eqInt(got.Ammo, want.Ammo) {
			t.Errorf("weapon %q = %+v (ammo %v), Python %+v (ammo %v)", id, got, show(got.Ammo), want, show(want.Ammo))
		}
	}
}

func TestInfraRulesMatchPython(t *testing.T) {
	p, s := loadPython(t), loadSet(t)
	if len(p.Infra) == 0 {
		t.Fatal("no infrastructure cases in golden data")
	}
	for _, c := range p.Infra {
		terrain, ok := s.Terrain[c.Terrain]
		if !ok {
			t.Errorf("terrain %q missing", c.Terrain)
			continue
		}
		got := Modifiers(terrain, c.Infra)
		if !eqInt(got.Movement, c.Movement) || !eqInt(got.Battle, c.Battle) || !eqInt(got.Sight, c.Sight) ||
			!reflect.DeepEqual(got.MovementAllowed, c.MovementAllowed) {
			t.Errorf("Modifiers(%s, %v) = [%v %v %v] %v, Python [%v %v %v] %v", c.Terrain, c.Infra,
				show(got.Movement), show(got.Battle), show(got.Sight), got.MovementAllowed,
				show(c.Movement), show(c.Battle), show(c.Sight), c.MovementAllowed)
		}
	}
}

func TestUnitWeaponsExist(t *testing.T) {
	s := loadSet(t)
	for id, u := range s.Units {
		for i, sl := range u.Weapons {
			if sl == nil {
				continue
			}
			if _, ok := s.Weapons[sl.Weapon]; !ok {
				t.Errorf("unit %q slot %d: unknown weapon %q", id, i, sl.Weapon)
			}
		}
	}
}

func TestSlotWeaponAppliesAttack(t *testing.T) {
	s := loadSet(t)
	inf := s.Units["CP_infantry"]
	if w := s.SlotWeapon(inf.Weapons[0]); w != s.Weapons["bayonet"] {
		t.Errorf("a slot without attack changed the catalogue weapon: %+v", w)
	}
	rifle := s.SlotWeapon(inf.Weapons[2])
	want := s.Weapons["gewehr98"]
	a := inf.Weapons[2].Attack
	want.Power, want.RangeMin, want.RangeMax, want.Air, want.Ground, want.Water = a.Power, a.RangeMin, a.RangeMax, a.Air, a.Ground, a.Water
	if rifle != want || rifle.Name != "Gewehr 98" || rifle.Ammo != s.Weapons["gewehr98"].Ammo {
		t.Errorf("rifle slot %+v, want %+v", rifle, want)
	}
	if s.Weapons["maxim"].Power != 430 {
		t.Error("the catalogue (bound to the weapon art) must keep hlrData.py's values")
	}
}

func TestModifiersDoesNotAliasTerrain(t *testing.T) {
	s := loadSet(t)
	grass := s.Terrain["grass"]
	before := *grass.Movement
	h := Modifiers(grass, []string{"road13"})
	*h.Movement = 99
	if *s.Terrain["grass"].Movement != before {
		t.Fatal("Modifiers result shares memory with the terrain table")
	}
}

func TestFixedModifiersRoadBridge(t *testing.T) {
	set, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	g := set.Terrain["grass"]
	infra := []string{"stream25", "rBridge36"}
	if m := Modifiers(g, infra); m.Movement != nil {
		t.Fatalf("Modifiers changed: rBridge made the stream passable (%d)", *m.Movement)
	}
	m := FixedModifiers(g, infra)
	if m.Movement == nil || *m.Movement != 0 || m.Battle == nil || *m.Battle != 0 {
		t.Errorf("FixedModifiers(stream, rBridge) = %+v, want a road hex", m)
	}
	// Everything else is unchanged.
	for _, infra := range [][]string{{"road25"}, {"stream25"}, {"stream25", "rBridge36", "path36"}, {"rail25", "trench1"}} {
		if a, b := Modifiers(g, infra), FixedModifiers(g, infra); !reflect.DeepEqual(a, b) {
			t.Errorf("%v: Modifiers %+v, FixedModifiers %+v", infra, a, b)
		}
	}
}
