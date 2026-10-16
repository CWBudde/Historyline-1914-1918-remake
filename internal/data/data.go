// Package data holds the terrain, unit and weapon tables. Terrain and the weapon
// catalogue are ported from hlrData.py; unit stats and attacks come from the
// original UNIT.DAT (see testdata/unitdat.json). The tables live in embedded JSON
// files whose field names follow the Python dictionaries.
package data

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed terrain.json units.json weapons.json
var files embed.FS

// Terrain holds the modifiers of a bgTilesModifiers entry. A nil value is the
// Python None, which is different from 0 (water has movement 0, a house None).
type Terrain struct {
	Movement *int `json:"movement"` // movement cost, 0 = not accessible (0-10)
	Battle   *int `json:"battle"`   // battle advantage (0-100)
	Sight    *int `json:"sight"`    // sight hindrance (0-10)
}

// Movement types of a unit, also the values of Hex.MovementAllowed.
const (
	MoveNormal = 0
	MoveTracks = 1 // trains: only on rail
	MoveWater  = 2 // boats: only on water
)

// Unit is a unit type. Armour, speed (the speed the original shows), weight,
// storageMax, group, cost and the slot attacks are the UNIT.DAT record Dos;
// sight, fuel, exp, skills and movementType are hlrData.py's unitsParameters.
type Unit struct {
	Name         string   `json:"name"`
	Country      string   `json:"country"`
	Armour       int      `json:"armour"`
	Speed        int      `json:"speed"`
	Weight       int      `json:"weight"`
	StorageMax   int      `json:"storageMax"`
	Group        int      `json:"group"` // original maximum group strength; not used by the rules yet
	Cost         *int     `json:"cost"`  // production cost, nil when the original sells none
	Dos          int      `json:"dos"`   // UNIT.DAT record index
	Sight        int      `json:"sight"`
	Fuel         int      `json:"fuel"`
	Exp          int      `json:"exp"`
	Skills       []int    `json:"skills"`
	MovementType int      `json:"movementType"`
	Weapons      [4]*Slot `json:"weapons"` // nil for an empty slot
}

// Slot is a weapon slot of a unit type: a catalogue weapon for name, art and
// ammo, and the original attack that replaces the catalogue's power, ranges and
// target classes. Slots without an attack (sidearms) keep the catalogue values.
type Slot struct {
	Weapon string  `json:"weapon"`
	Attack *Attack `json:"attack,omitempty"`
}

// Attack is what a slot does against air, ground and water targets.
type Attack struct {
	Power    int  `json:"power"`
	RangeMin int  `json:"rangeMin"`
	RangeMax int  `json:"rangeMax"`
	Air      bool `json:"air"`
	Ground   bool `json:"ground"`
	Water    bool `json:"water"`
}

// Weapon is a weaponsParameters entry without its graphics. The catalogue is
// bound to the weapon art (art/interface-v1.json); units use it through Slot.
type Weapon struct {
	Name     string `json:"name"`
	RangeMin int    `json:"rangeMin"`
	RangeMax int    `json:"rangeMax"`
	Power    int    `json:"power"`
	Air      bool   `json:"air"`
	Ground   bool   `json:"ground"`
	Water    bool   `json:"water"`
	Ammo     *int   `json:"ammo"` // nil = unlimited
}

// Set is the complete game data, keyed by the ids the levels use.
type Set struct {
	Terrain map[string]Terrain
	Units   map[string]Unit
	Weapons map[string]Weapon
}

// Load parses the embedded tables and checks that every weapon a unit carries
// exists.
func Load() (*Set, error) {
	s := &Set{}
	for name, dst := range map[string]any{
		"terrain.json": &s.Terrain,
		"units.json":   &s.Units,
		"weapons.json": &s.Weapons,
	} {
		raw, err := files.ReadFile(name)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, dst); err != nil {
			return nil, fmt.Errorf("data: %s: %w", name, err)
		}
	}
	for id, u := range s.Units {
		for _, sl := range u.Weapons {
			if _, ok := s.Weapons[sl.weapon()]; sl != nil && !ok {
				return nil, fmt.Errorf("data: unit %s: unknown weapon %q", id, sl.Weapon)
			}
		}
	}
	return s, nil
}

func (sl *Slot) weapon() string {
	if sl == nil {
		return ""
	}
	return sl.Weapon
}

// SlotWeapon is the weapon a unit fires from slot sl: the catalogue entry with
// the slot's attack applied.
func (s *Set) SlotWeapon(sl *Slot) Weapon {
	w := s.Weapons[sl.Weapon]
	if a := sl.Attack; a != nil {
		w.Power, w.RangeMin, w.RangeMax = a.Power, a.RangeMin, a.RangeMax
		w.Air, w.Ground, w.Water = a.Air, a.Ground, a.Water
	}
	return w
}
