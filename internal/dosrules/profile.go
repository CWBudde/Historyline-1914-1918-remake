// Package dosrules is the pure Go resolver of the original DOS rules
// (profile dos-0.1): the casualty routine, the pincer and blocking support,
// the battle distance, the experience awards and the order handler that
// composes them; the movement queries, the move order and its steps, loading,
// unloading and exits, and the phase change that brings both sides' views
// into line, as executed in the identified builds. It holds no original
// bytes, only the numeric fields the rules read. The research oracle
// internal/dosoracle checks it against the executed routines and records the
// fixtures under testdata/.
package dosrules

import (
	_ "embed"
	"encoding/json"
)

// Type is the part of a UNIT.DAT type record that combat reads.
type Type struct {
	Armour        int `json:"armour"`        // +2
	Group         int `json:"group"`         // +3, full group strength
	TargetFlags   int `json:"targetFlags"`   // +6 word: classes 0x3c, long range 0x40, never returns fire 0x80
	AirRange      int `json:"airRange"`      // +8, encoded (one more than shown)
	SurfaceRange  int `json:"surfaceRange"`  // +9, encoded
	AirPower      int `json:"airPower"`      // +10
	GroundPower   int `json:"groundPower"`   // +11
	WaterPower    int `json:"waterPower"`    // +12
	AttackMask    int `json:"attackMask"`    // +0x0d word: target classes it attacks
	StrengthFlags int `json:"strengthFlags"` // +0x0f word: bit 4 counts the alternate strength
}

// alternate reports whether the unit's strength is its alternate byte.
func (t Type) alternate() bool { return t.StrengthFlags&4 != 0 }

// Terms are the signed attack and defence terms of one terrain class.
type Terms struct{ Attack, Defence int }

// Tables are the numbers one build supplies: its 51 type records and the
// terms of the nine terrain classes, and the fields of the type and terrain
// records the movement queries read (nil where the build's movement code is
// not located).
type Tables struct {
	Build string      `json:"build"` // SHA-256 prefix of the executable
	Types []Type      `json:"types"`
	Terms [9]Terms    `json:"terms"`
	Move  *MoveTables `json:"move,omitempty"`
}

// UnmarshalJSON reads the terms as [attack, defence] pairs.
func (t *Terms) UnmarshalJSON(b []byte) error {
	var p [2]int
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	t.Attack, t.Defence = p[0], p[1]
	return nil
}

// MarshalJSON writes the terms as an [attack, defence] pair.
func (t Terms) MarshalJSON() ([]byte, error) { return json.Marshal([2]int{t.Attack, t.Defence}) }

// Profile is a rules profile bound to one build's tables. Builds that turn
// out to differ in a rule get a field here, not a hidden constant.
type Profile struct {
	Name   string
	Tables *Tables
	// TransposedDistance passes (row, column) to the hex distance, as the
	// casualty routine does: one of the six adjacent directions then counts
	// as distance 2 and ranged distances are off by up to two. Without it
	// the distance is called with (column, row), as the movement code does,
	// and is the layout distance.
	TransposedDistance bool
}

//go:embed demo.json
var demoJSON []byte

// DemoTables are the tables of HL14-18.EXE of the free DOS demo, the build
// of record until the full HL.EXE is verified.
var DemoTables = mustTables(demoJSON)

// Demo is profile dos-0.1 on the demo build, as executed.
var Demo = Profile{Name: "dos-0.1", Tables: DemoTables, TransposedDistance: true}

func mustTables(b []byte) *Tables {
	var t Tables
	if err := json.Unmarshal(b, &t); err != nil {
		panic("dosrules: " + err.Error())
	}
	return &t
}
