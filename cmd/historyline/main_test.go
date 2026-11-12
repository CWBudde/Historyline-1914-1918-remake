package main

import (
	"image/color"
	"testing"
)

// The browser build has no levels directory; the embedded copies must load.
func TestLoadLevelEmbeddedFallback(t *testing.T) {
	for _, name := range []string{"1914-08:cp", "1914-10:cp", "1914-12:cp"} {
		if _, err := loadLevel(t.TempDir(), name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := loadLevel(t.TempDir(), "nope"); err == nil {
		t.Error("unknown level loaded")
	}
}

// Fog must not leak terrain colour or units of hexes the player cannot see.
func TestFogPolicy(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	if got := fogShade(fogVisible, c); got != c {
		t.Errorf("visible shade = %v, want unchanged", got)
	}
	u1, u2 := fogShade(fogUnseen, c), fogShade(fogUnseen, color.RGBA{R: 10, G: 240, B: 90, A: 255})
	if u1 != u2 {
		t.Errorf("unseen shade depends on terrain: %v vs %v", u1, u2)
	}
	if s := fogShade(fogSeen, c); s.R >= c.R || s.R > s.B {
		t.Errorf("seen shade %v should be dimmed and grey", s)
	}
}

// pythonPulse is the Python remake's PULSE (internal/level's fixture), a
// German unit within reach of the enemy at the start, which advancing on
// the enemy wins.
const pythonPulse = "../../internal/level/testdata/pulse.json"

func TestDevIndicators(t *testing.T) {
	g := &game{}
	if got := g.devIndicators(); got != "" {
		t.Errorf("no dev flags: got %q", got)
	}
	g.reveal, g.hexNumbers = true, true
	if got := g.devIndicators(); got != "DEV: REVEAL HEXNUMBERS" {
		t.Errorf("got %q", got)
	}
}
