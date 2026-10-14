// Package data holds the terrain, unit and weapon tables ported from hlrData.py.
// The tables live in embedded JSON files whose field names follow the Python
// dictionaries.
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

// Unit is an unitsParameters entry without its graphics.
type Unit struct {
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
	Weapons      [4]string `json:"weapons"` // weapon ids, "" for an empty slot
}

// Weapon is a weaponsParameters entry without its graphics.
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
		for _, w := range u.Weapons {
			if _, ok := s.Weapons[w]; w != "" && !ok {
				return nil, fmt.Errorf("data: unit %s: unknown weapon %q", id, w)
			}
		}
	}
	return s, nil
}
