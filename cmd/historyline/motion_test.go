package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestMotionKeyBinding(t *testing.T) {
	got := actionsFor(func(k ebiten.Key) bool { return k == ebiten.KeyM })
	if len(got) != 1 || got[0] != actToggleMotion {
		t.Errorf("M -> %v", got)
	}
}
