//go:build !origtiles

package main

import (
	"strings"
	"testing"
)

// TestOrigTilesOff: a build without -tags origtiles answers F8 with how to
// get the original tiles and keeps drawing the remake's terrain.
func TestOrigTilesOff(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	g := newMapGame(d.s.Level, false)
	g.dos = d
	if msg := g.toggleOrigTiles(); !strings.Contains(msg, "-tags origtiles") {
		t.Errorf("F8 says %q", msg)
	}
	if g.drawOrigTiles(nil, nil) {
		t.Error("the original tiles are drawn")
	}
}
