package main

import (
	"image/color"
	"testing"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

// Fog must not leak terrain colour or units of hexes the player cannot see.
func TestFogPolicy(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	if got := fogShade(state.Visible, c); got != c {
		t.Errorf("visible shade = %v, want unchanged", got)
	}
	u1, u2 := fogShade(state.Unseen, c), fogShade(state.Unseen, color.RGBA{R: 10, G: 240, B: 90, A: 255})
	if u1 != u2 {
		t.Errorf("unseen shade depends on terrain: %v vs %v", u1, u2)
	}
	if s := fogShade(state.Seen, c); s.R >= c.R || s.R > s.B {
		t.Errorf("seen shade %v should be dimmed and grey", s)
	}
	for f, want := range map[state.Fog]bool{state.Unseen: false, state.Seen: false, state.Visible: true} {
		if showsUnit(f) != want {
			t.Errorf("showsUnit(%v) = %v, want %v", f, !want, want)
		}
	}
}
