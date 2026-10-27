package main

import (
	"fmt"
	"strings"
	"testing"

	"errors"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
)

// An original map loads from the local original files with its campaign
// level, plays its exact board under both rule sets, and names every unit, a
// ship's rear by its ship.
func TestOriginalMap(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lvl, b := originalMap(t, "original/06", set)
	if b == nil || lvl.MapName != "06" || b.Width != 2*len(lvl.Lines[0]) || len(lvl.Lines) != 2*b.Height-1 {
		t.Fatalf("map %s: board %v, %d lines of %d", lvl.MapName, b != nil, len(lvl.Lines), len(lvl.Lines[0]))
	}
	player, err := state.PlayerByName(lvl.Player)
	if err != nil {
		t.Fatal(err)
	}
	c := matchConfig{lvl: lvl, board: b, set: set, player: player, rules: rulesClassic, opponent: "pass"}
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
	c.rules = rulesModern
	gm, err = newMatch(c, nil)
	if err != nil {
		t.Fatalf("modern: %v", err)
	}
	if gm.dos == nil || gm.dos.s.Level != c.lvl {
		t.Error("modern: not a session on map 06")
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
		{"monthly/1914-08:ec", "BATLE", state.Entente.String()},
		{"monthly/1914-08", "PULSE", state.CentralPowers.String()},
		{"monthly/1914-09", "KLUCK", state.CentralPowers.String()},
		{"skirmish/01-track:2p", "TRACK", state.CentralPowers.String()},
	} {
		lvl, b, err := loadMap(t.TempDir(), c.name, set)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if b != nil || lvl.MapName != c.mapName || lvl.Player != c.player || lvl.Extra == nil {
			t.Errorf("%s: %s as %s, board %v", c.name, lvl.MapName, lvl.Player, b != nil)
		}
	}
	if _, _, err := loadMap(t.TempDir(), "monthly/1914-08:xx", set); err == nil {
		t.Error("an unknown setup loaded")
	}
}

// originalMap loads a map of the local original files; the test is skipped
// without them.
func originalMap(t *testing.T, name string, set *data.Set) (*level.Level, *dosbind.Board) {
	t.Helper()
	lvl, b, err := loadMap(t.TempDir(), name, set)
	if errors.Is(err, origmaps.ErrNotFound) {
		t.Skip("no local original files")
	}
	if err != nil {
		t.Fatal(err)
	}
	return lvl, b
}
