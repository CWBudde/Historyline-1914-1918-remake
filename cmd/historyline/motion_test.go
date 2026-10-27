package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestReducedMotionWalkAndEffects(t *testing.T) {
	path := []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 0, Col: 1}, {Line: 0, Col: 2}}
	if w := newWalkMotion(path, nil, false); w.Done() {
		t.Error("normal walk is instant")
	}
	if w := newWalkMotion(path, nil, true); !w.Done() {
		t.Error("reduced walk is not instant")
	}

	var fx effects
	fx.Show("WON", victoryFrames)
	long := fx.Banner().Frames
	fx.Reduced = true
	fx.Show("WON", victoryFrames)
	if got := fx.Banner().Frames; got > reducedFrames || got >= long {
		t.Errorf("reduced banner frames = %d (normal %d)", got, long)
	}
	fx.Mark(hexgrid.Coord{Col: 1}, true)
	for _, m := range fx.Marks() {
		if m.Total > reducedFrames {
			t.Errorf("reduced mark lasts %d frames", m.Total)
		}
	}

	d := newDosTest(t, "monthly/1914-08:cp", passPlanner{})
	d.walk = newWalkMotion(path, nil, false)
	d.SetReducedMotion(true)
	if !d.fx.Reduced || !d.walk.Done() {
		t.Error("SetReducedMotion did not reach the effects and the walk")
	}
	d.SetReducedMotion(false)
	if d.fx.Reduced {
		t.Error("reduced motion not switched off")
	}
}

func TestMotionKeyBinding(t *testing.T) {
	got := actionsFor(func(k ebiten.Key) bool { return k == ebiten.KeyM })
	if len(got) != 1 || got[0] != actToggleMotion {
		t.Errorf("M -> %v", got)
	}
}
