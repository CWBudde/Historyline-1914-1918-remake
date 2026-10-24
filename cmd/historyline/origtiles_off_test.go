//go:build !origtiles

package main

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// TestOrigTilesOff: a build without -tags origtiles answers F8 with how to
// get the original art and keeps drawing the remake's terrain and units.
func TestOrigTilesOff(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	g := newMapGame(d.s.Level, false)
	g.dos = d
	if msg := g.toggleOrigArt(); !strings.Contains(msg, "-tags origtiles") {
		t.Errorf("F8 says %q", msg)
	}
	if g.drawOrigTiles(nil, nil) {
		t.Error("the original tiles are drawn")
	}
	if g.drawOrigUnit(nil, d.s.Unit(0), 0, 0, unitsprite.N, false) {
		t.Error("an original unit is drawn")
	}
}
