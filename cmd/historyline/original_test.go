package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

// An imported original map loads from the embedded files, plays on the
// original rules only, and shows units the remake has no unit for by type.
func TestOriginalMap(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lvl, b, err := loadMap(t.TempDir(), "original/06", set)
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
	unnamed := false
	for at, u := range gm.dos.own().Units {
		if u.ID != "" {
			continue
		}
		gm.hover, gm.hoverOnMap = at, true
		if lines := strings.Join(gm.infoLines(), "\n"); !strings.Contains(lines, fmt.Sprintf("Type %d", u.Type)) {
			t.Errorf("unit of type %d without a remake unit: %q", u.Type, lines)
		}
		unnamed = true
		break
	}
	if !unnamed {
		t.Error("map 06 has no unit without a remake unit")
	}
	for _, rules := range []string{"standard", "classic"} {
		c.rules = rules
		if _, err := newMatch(c, nil); err == nil {
			t.Errorf("%s rules accepted an original map", rules)
		}
	}
	if _, _, err := loadMap(t.TempDir(), "original/99", set); err == nil {
		t.Error("original/99 loaded")
	}
}
