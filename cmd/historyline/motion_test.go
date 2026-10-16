package main

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestReportDismiss(t *testing.T) {
	r := &report{frames: reportFrames}
	if !r.Active() {
		t.Fatal("report not active")
	}
	r.Dismiss()
	if r.Active() {
		t.Error("report still active after dismiss")
	}
	var nilReport *report
	nilReport.Dismiss() // must not panic
	c := &control{report: &report{frames: 5}}
	c.DismissReport()
	if c.report.Active() {
		t.Error("control.DismissReport left the panel up")
	}
}

func TestReducedMotionWalkAndEffects(t *testing.T) {
	path := []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 0, Col: 1}, {Line: 0, Col: 2}}
	if w := newWalkMotion(path, nil, false); w.Done() {
		t.Error("normal walk is instant")
	}
	if w := newWalkMotion(path, nil, true); !w.Done() {
		t.Error("reduced walk is not instant")
	}

	var fx effects
	fx.Victory(state.CentralPowers)
	long := fx.Banner().Frames
	fx.Reduced = true
	fx.Victory(state.CentralPowers)
	if got := fx.Banner().Frames; got > reducedFrames || got >= long {
		t.Errorf("reduced banner frames = %d (normal %d)", got, long)
	}
	fx.AddBattle(hexgrid.Coord{}, hexgrid.Coord{Col: 1}, state.Battle{Damage: 1, DefenderDestroyed: true})
	for _, m := range fx.Marks() {
		if m.Total > reducedFrames {
			t.Errorf("reduced mark lasts %d frames", m.Total)
		}
	}

	c := newTestControl(t)
	c.SetReducedMotion(true)
	if !c.fx.Reduced || c.reportTime() != reportFramesReduced {
		t.Error("SetReducedMotion did not reach effects and report")
	}
	from := ownMovers(c)[0]
	c.Click(from)
	for at := range c.moves {
		c.Click(at)
		break
	}
	if !c.walk.Done() {
		t.Error("walk animates with reduced motion")
	}
	c.SetReducedMotion(false)
	if c.fx.Reduced {
		t.Error("reduced motion not switched off")
	}
}

func TestMotionKeyBinding(t *testing.T) {
	got := actionsFor(func(k ebiten.Key) bool { return k == ebiten.KeyM })
	if len(got) != 1 || got[0] != actToggleMotion {
		t.Errorf("M -> %v", got)
	}
}

// Entering an enemy depot with a unit that may capture reports the capture.
func TestCaptureBanner(t *testing.T) {
	c := newTestControl(t)
	other := state.Entente
	if c.player == other {
		other = state.CentralPowers
	}
	found := false
	for _, from := range ownMovers(c) {
		u := c.st.At(from).Unit
		for to := range c.st.Movable(from, u) {
			h := c.st.At(to)
			if h.Container != nil || h.Unit != nil {
				continue
			}
			h.Container, h.Owner = &state.Container{Capacity: 99}, other
			if _, ok := c.st.Movable(from, u)[to]; !ok {
				h.Container = nil
				continue
			}
			c.Click(from)
			c.Click(to)
			found = true
			break
		}
		if found {
			break
		}
	}
	if !found {
		t.Skip("no unit in level1_cp can enter a depot")
	}
	b := c.fx.Banner()
	if b == nil || !strings.Contains(b.Text, "CAPTURING") {
		t.Fatalf("no capture banner: %+v, message %q", b, c.Message)
	}
	if !strings.Contains(c.Message, "capturing") {
		t.Errorf("message = %q", c.Message)
	}
}
