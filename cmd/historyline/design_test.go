package main

import (
	"strings"
	"testing"
)

// TestCycleDesignNoDos: without a dos session F8 cycles the painted backdrop
// and the terrain tiles, and the original (DOS) design is skipped.
func TestCycleDesignNoDos(t *testing.T) {
	lvl, err := loadLevel(t.TempDir(), "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	g := newMapGame(lvl, false)
	ww, wh := lvl.Grid().Bounds()
	g.backdrop = &backdrop{ww: float64(ww), wh: float64(wh)}

	if msg := g.cycleDesign(); !strings.Contains(msg, "terrain") {
		t.Fatalf("backdrop -> %q, want the terrain tiles", msg)
	}
	if g.backdrop != nil || g.backdropHidden == nil {
		t.Fatalf("backdrop still shown: %v, hidden %v", g.backdrop != nil, g.backdropHidden != nil)
	}
	if msg := g.cycleDesign(); !strings.Contains(msg, "backdrop") {
		t.Fatalf("tiles -> %q, want the backdrop again", msg)
	}
	if g.backdrop == nil || g.backdropHidden != nil {
		t.Fatalf("backdrop not restored: %v, hidden %v", g.backdrop != nil, g.backdropHidden != nil)
	}
	if g.origOn() {
		t.Error("the original art is drawn without a dos session")
	}
}

// TestCycleDesignNoBackdrop: a level without a painting stays on the terrain
// tiles.
func TestCycleDesignNoBackdrop(t *testing.T) {
	lvl, err := loadLevel(t.TempDir(), "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	g := newMapGame(lvl, false)
	if msg := g.cycleDesign(); !strings.Contains(msg, "terrain") {
		t.Fatalf("no backdrop -> %q, want the terrain tiles", msg)
	}
	if g.backdrop != nil || g.backdropHidden != nil {
		t.Error("a backdrop appeared for a level without one")
	}
}

// TestCycleDesignDosUnavailable: under -rules dos with the original art
// unavailable, F8 still cycles the backdrop and the terrain tiles.
func TestCycleDesignDosUnavailable(t *testing.T) {
	d := newDosTest(t, "monthly/1914-08:cp", passPlanner{})
	g := newMapGame(d.s.Level, false)
	g.dos = d
	ww, wh := d.s.Level.Grid().Bounds()
	g.backdrop = &backdrop{ww: float64(ww), wh: float64(wh)}
	g.orig = &origArt{err: "build with -tags origtiles (just play-original)"}

	if msg := g.cycleDesign(); !strings.Contains(msg, "terrain") {
		t.Fatalf("backdrop -> %q, want the terrain tiles", msg)
	}
	if msg := g.cycleDesign(); !strings.Contains(msg, "backdrop") || !strings.Contains(msg, "origtiles") {
		t.Fatalf("tiles -> %q, want the backdrop, noting the unavailable original art", msg)
	}
	if g.origOn() {
		t.Error("the unavailable original art is drawn")
	}
}

// TestCycleDesignDos: under -rules dos with the original art loaded, F8 cycles
// backdrop, terrain tiles and the original tiles and units.
func TestCycleDesignDos(t *testing.T) {
	d := newDosTest(t, "monthly/1914-08:cp", passPlanner{})
	g := newMapGame(d.s.Level, false)
	g.dos = d
	ww, wh := d.s.Level.Grid().Bounds()
	g.backdrop = &backdrop{ww: float64(ww), wh: float64(wh)}
	g.orig = &origArt{} // already loaded, so no local original files are needed
	g.applyLayout(1800, 1000, 1)
	g.layoutSplit()
	if !g.split.frame.Slim {
		t.Fatal("the backdrop does not lay out the slim frame")
	}

	if msg := g.cycleDesign(); !strings.Contains(msg, "terrain") {
		t.Fatalf("backdrop -> %q, want the terrain tiles", msg)
	}
	if g.split.frame.Slim {
		t.Error("the terrain tiles keep the backdrop's slim frame")
	}
	if msg := g.cycleDesign(); !strings.Contains(msg, "original") {
		t.Fatalf("tiles -> %q, want the original (DOS) design", msg)
	}
	if !g.origOn() || g.backdrop != nil {
		t.Fatalf("DOS design: orig on %v, backdrop shown %v", g.origOn(), g.backdrop != nil)
	}
	if msg := g.cycleDesign(); !strings.Contains(msg, "backdrop") {
		t.Fatalf("DOS -> %q, want the backdrop again", msg)
	}
	if !g.split.frame.Slim {
		t.Error("the backdrop again does not lay out the slim frame")
	}
	if g.origOn() || g.backdrop == nil {
		t.Fatalf("backdrop again: orig on %v, backdrop shown %v", g.origOn(), g.backdrop != nil)
	}
}
