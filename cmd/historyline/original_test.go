package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
)

// skipMissingOriginal skips local comparison tests when the unshipped
// original files are absent. Other loading errors still fail the test.
func skipMissingOriginal(t *testing.T, err error) {
	t.Helper()
	if errors.Is(err, origmaps.ErrNotFound) {
		t.Skip("no original files")
	}
}

// An imported original map loads from the embedded files with its campaign
// level, plays its exact board on the original rules and its campaign level on
// the remake rules, and names every unit, a ship's rear by its ship.
func TestOriginalMap(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lvl, b, err := loadMap(t.TempDir(), "original/06", set)
	skipMissingOriginal(t, err)
	if err != nil {
		t.Fatal(err)
	}
	if b == nil || lvl.MapName != "06" || b.Width != 2*len(lvl.Lines[0]) || len(lvl.Lines) != 2*b.Height-1 {
		t.Fatalf("map %s: board %v, %d lines of %d", lvl.MapName, b != nil, len(lvl.Lines), len(lvl.Lines[0]))
	}
	player, err := state.PlayerByName(lvl.Player)
	if err != nil {
		t.Fatal(err)
	}
	c := matchConfig{lvl: lvl, board: b, set: set, player: player, rules: "dos", opponent: "pass"}
	gm, err := newMatch(c, nil)
	if err != nil || gm.dos == nil {
		t.Fatalf("dos match: %v", err)
	}
	if gm.dos.s.Board != b {
		t.Error("the session does not play the original map's board")
	}
	// A ship's rear is its head's unit: the info box names the ship.
	rears := 0
	for at, u := range gm.dos.own().Units {
		if u.ID == "" {
			t.Errorf("unit of type %d at %v without a remake unit", u.Type, at)
		}
		if u.Head == nil {
			continue
		}
		rears++
		gm.hover, gm.hoverOnMap, gm.hexNumbers = at, true, true // the dev info box
		if lines := strings.Join(gm.infoLines(), "\n"); !strings.Contains(lines, u.Name) || strings.Contains(lines, fmt.Sprintf("Type %d", u.Type)) {
			t.Errorf("rear of type %d: %q", u.Type, lines)
		}
	}
	if rears == 0 {
		t.Error("map 06 has no ship")
	}
	c.rules = "standard"
	gm, err = newMatch(c, nil)
	if err != nil {
		t.Fatalf("standard: %v", err)
	}
	if gm.state == nil || gm.state.Name != "06" {
		t.Error("standard: not the campaign level of map 06")
	}
	if _, _, err := loadMap(t.TempDir(), "original/99", set); err == nil {
		t.Error("original/99 loaded")
	}
}

// -level plays a scenario's setup: the named one, else the first of cp,
// ec, 2p.
func TestScenarioLevel(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ name, mapName, player string }{
		{"1914-08:ec", "BATLE", state.Entente.String()},
		{"1914-08", "PULSE", state.CentralPowers.String()},
		{"1914-09", "KLUCK", state.CentralPowers.String()},
		{"1915-05:2p", "Track", state.CentralPowers.String()},
		{"extra/13-grand:2p", "GRAND", state.CentralPowers.String()},
	} {
		lvl, b, err := loadMap(t.TempDir(), c.name, set)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if b != nil || lvl.MapName != c.mapName || lvl.Player != c.player || lvl.Extra == nil {
			t.Errorf("%s: %s as %s, board %v", c.name, lvl.MapName, lvl.Player, b != nil)
		}
	}
	if _, _, err := loadMap(t.TempDir(), "1914-08:xx", set); err == nil {
		t.Error("an unknown setup loaded")
	}
}
