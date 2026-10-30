package dosplay

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// lightBridgeLevel is a stream down the odd lines of column 2 (as on
// NONNE), crossed at line 3 by a bridge with the given deck infra, and a
// unit west of the bridge at (2,2); the east bank is (2,3).
func lightBridgeLevel(unit string, deck ...string) *level.Level {
	lv := &level.Level{MapName: "lightbridge", Player: "Central Powers"}
	for l := range 11 {
		line := make([]level.Hex, 6-l%2)
		for c := range line {
			line[c] = level.Hex{Terrain: "grass"}
		}
		lv.Lines = append(lv.Lines, line)
	}
	for l := 1; l < 11; l += 2 {
		lv.Lines[l][2] = level.Hex{Terrain: "grass", Infra: level.Infra{IDs: []string{"stream36"}, List: true}}
	}
	infra := append([]string{"stream36", "rBridge25"}, deck...)
	lv.Lines[3][2] = level.Hex{Terrain: "grass", Infra: level.Infra{IDs: infra, List: true}}
	lv.Lines[2][2].Unit = unit
	return lv
}

func bridgeUnit(t *testing.T, s *Session) int {
	t.Helper()
	for idx, id := range s.Board.IDs {
		if id != "" {
			if c, ok := s.Board.Coord(s.Game.Units[idx].Cell(0)); ok && c == (hexgrid.Coord{Line: 2, Col: 2}) {
				return idx
			}
		}
	}
	t.Fatal("no unit at (2,2)")
	return -1
}

var bridgeHex = hexgrid.Coord{Line: 3, Col: 2}
var eastHex = hexgrid.Coord{Line: 2, Col: 3}

// A light bridge (a path deck over the stream, as on NONNE) carries foot
// units and light towed guns up to the light artillery, but no armoured
// vehicles, tanks or heavier guns.
func TestLightBridgeDestinations(t *testing.T) {
	for _, tc := range []struct {
		unit string
		want bool // may enter the bridge hex
		east bool // and reach the east bank within its budget
	}{
		{"CP_infantry", true, true},
		{"CP_eliteInf", true, true},
		{"CP_sappers", true, true},
		{"CP_antiTank", true, false},
		{"CP_staticAA", true, false},
		{"CP_cavalry", true, true},
		{"CP_lArtillery", true, false},
		{"CP_ft17", false, false},
		{"CP_armCar", false, false},
		{"CP_markI", false, false},
		{"CP_mArtillery", false, false},
		{"CP_supplyCar", false, false},
		{"CP_construction", false, false},
	} {
		s := session(t, lightBridgeLevel(tc.unit, "path25"), Options{})
		u := bridgeUnit(t, s)
		dst, err := s.Destinations(0, u)
		if err != nil {
			t.Fatalf("%s: %v", tc.unit, err)
		}
		if got := dst[bridgeHex]; got != tc.want {
			t.Errorf("%s on the bridge = %v, want %v", tc.unit, got, tc.want)
		}
		if tc.east && !dst[eastHex] {
			t.Errorf("%s cannot cross to the east bank", tc.unit)
		}
		if !tc.want && dst[eastHex] {
			t.Errorf("%s crossed to the east bank", tc.unit)
		}
	}
}

// A road bridge carries everything a road does.
func TestRoadBridgeDestinations(t *testing.T) {
	s := session(t, lightBridgeLevel("CP_ft17", "road25"), Options{})
	dst, err := s.Destinations(0, bridgeUnit(t, s))
	if err != nil {
		t.Fatal(err)
	}
	if !dst[bridgeHex] || !dst[eastHex] {
		t.Errorf("tank destinations %v lack the road bridge or the east bank", dst)
	}
}

func TestLightBridgeMove(t *testing.T) {
	s := session(t, lightBridgeLevel("CP_infantry", "path25"), Options{})
	u := bridgeUnit(t, s)
	if _, err := s.Move(0, u, eastHex); err != nil {
		t.Fatalf("infantry move: %v", err)
	}
	if _, ok := s.View(0).Units[eastHex]; !ok {
		t.Errorf("infantry is not on the east bank: %v", s.View(0).Units)
	}

	s = session(t, lightBridgeLevel("CP_ft17", "path25"), Options{})
	u = bridgeUnit(t, s)
	evs, err := s.Move(0, u, eastHex)
	if err != nil {
		t.Fatalf("tank move: %v", err)
	}
	for _, ev := range evs {
		if ev.Kind == Moved {
			t.Errorf("the tank crossed the light bridge: %+v", ev)
		}
	}
	if _, ok := s.View(0).Units[hexgrid.Coord{Line: 2, Col: 2}]; !ok {
		t.Errorf("the tank left its hex: %v", s.View(0).Units)
	}
}

// NONNE's path bridge (line 20, column 8) binds to a bridge record, not to
// the bare stream its key used to fall back to.
func TestNonnePathBridgeBinds(t *testing.T) {
	lv, err := scenario.LoadLevel("../../levels/monthly/1914-11.json", scenario.EC)
	if err != nil {
		t.Fatal(err)
	}
	s := session(t, lv, Options{})
	cell, ok := s.Board.Cell(hexgrid.Coord{Line: 19, Col: 8})
	if !ok {
		t.Fatal("bridge hex off the board")
	}
	rec := s.Board.Terrain[cell]
	tr := dosrules.Full.Tables.Move.Terrains[rec]
	if tr.Mask == 4 {
		t.Fatalf("path bridge bound to the bare stream record %d", rec)
	}
	if tr.Mask&16 == 0 || tr.Cost != 5 {
		t.Errorf("path bridge record %d = %+v, want a path bridge (mask 16, cost 5)", rec, tr)
	}
}
