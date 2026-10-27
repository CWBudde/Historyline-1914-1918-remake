package main

import (
	"testing"
)

// The browser build has no levels directory; the embedded copies must load.
func TestLoadLevelEmbeddedFallback(t *testing.T) {
	for _, name := range []string{"monthly/1914-08:cp", "monthly/1914-10:cp", "monthly/1914-12:cp"} {
		if _, err := loadLevel(t.TempDir(), name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := loadLevel(t.TempDir(), "nope"); err == nil {
		t.Error("unknown level loaded")
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
	// Unapproved unit art is labelled too, so screenshots are never mistaken for it.
	g.reveal, g.hexNumbers, g.art = false, false, &artPacks{candidate: true}
	if got := g.devIndicators(); got != "DEV: CANDIDATE-ART" {
		t.Errorf("got %q", got)
	}
	g.art.candidate = false
	if got := g.devIndicators(); got != "" {
		t.Errorf("approved art: got %q", got)
	}
}
