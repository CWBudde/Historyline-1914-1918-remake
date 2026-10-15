package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestWalkInterpolatesAlongPath(t *testing.T) {
	path := []hexgrid.Coord{{Line: 2, Col: 2}, {Line: 2, Col: 3}, {Line: 2, Col: 4}}
	w := newWalk(path)
	x0, y0 := hexgrid.Center(path[0])
	if x, y := w.Pos(); x != x0 || y != y0 {
		t.Errorf("start %v,%v want %v,%v", x, y, x0, y0)
	}
	for range stepFrames {
		w.Tick()
	}
	x1, y1 := hexgrid.Center(path[1])
	if x, y := w.Pos(); x != x1 || y != y1 {
		t.Errorf("after one step %v,%v want %v,%v", x, y, x1, y1)
	}
	if w.Done() {
		t.Fatal("done too early")
	}
	for range stepFrames {
		w.Tick()
	}
	x2, y2 := hexgrid.Center(path[2])
	if x, y := w.Pos(); !w.Done() || x != x2 || y != y2 {
		t.Errorf("end %v,%v done=%v", x, y, w.Done())
	}
	if newWalk(path[:1]) != nil || !newWalk(nil).Done() {
		t.Error("a path without steps must not animate")
	}
}

func TestClickWaitsForWalk(t *testing.T) {
	c := newTestControl(t)
	from := ownMovers(c)[0]
	c.Click(from)
	for at := range c.moves {
		c.Click(at)
		break
	}
	if c.walk.Done() {
		t.Skip("one-step move, nothing to wait for")
	}
	c.Click(from)
	if c.Message != "" {
		t.Errorf("click during the walk was handled: %q", c.Message)
	}
}

func TestBattleReportFromEngineResult(t *testing.T) {
	c := newTestControl(t)
	for l, line := range c.st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			u := c.st.At(at).Unit
			if u == nil || u.Owner != c.player {
				continue
			}
			c.Select(at)
			for tgt := range c.targets {
				c.OpenWeaponMenu(tgt)
				for i, o := range c.menu {
					if !o.Usable {
						continue
					}
					def := c.st.At(tgt).Unit
					before := def.Size
					c.ChooseWeapon(i)
					r := c.report
					if !r.Active() || r.Defender.Before != before || r.Defender.After != def.Size || r.Attacker.Weapon == "" {
						t.Fatalf("report %+v does not match engine state (defender %d -> %d)", r, before, def.Size)
					}
					for range reportFrames {
						r.Tick()
					}
					if r.Active() {
						t.Error("report never expires")
					}
					return
				}
			}
		}
	}
	t.Skip("no attack available at the start of level1_cp")
}
