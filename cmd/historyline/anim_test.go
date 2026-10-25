package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
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
	msg := c.Message
	c.Click(from)
	if c.Message != msg || !c.walk.Done() {
		t.Errorf("click during the walk did more than finish it: %q, walk done %v", c.Message, c.walk.Done())
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
	t.Skip("no attack available at the start of PULSE")
}

func TestSkipFinishesWalk(t *testing.T) {
	c := boardControl(t, map[hexgrid.Coord]level.Hex{at(2, 2): onGrass("CP_cavalry")})
	cav := c.st.At(at(2, 2)).Unit
	c.Click(at(2, 2))
	c.Click(at(6, 2))
	if c.walk.Done() || c.st.At(at(6, 2)).Unit != cav {
		t.Fatalf("no walk to skip: %q", c.Message)
	}
	c.Click(at(2, 2)) // a click during the walk only finishes it
	if !c.walk.Done() {
		t.Error("the click did not finish the walk")
	}
	if c.st.At(at(6, 2)).Unit != cav || c.st.At(at(2, 2)).Unit != nil || !c.selected || c.from != at(6, 2) {
		t.Errorf("skipping changed the game or the selection: selected %v at %v", c.selected, c.from)
	}
	if c.SkipWalk() {
		t.Error("SkipWalk with no walk reported a skip")
	}
}

func TestWalkDrawsMovingUnitIntoCarrier(t *testing.T) {
	place := map[hexgrid.Coord]level.Hex{at(2, 2): onGrass("CP_infantry"), at(6, 2): onGrass("CP_supplyCar")}
	c := boardControl(t, place)
	g := newGame(boardLevel(place), c.st, c.player, false)
	g.ctl = c
	inf, car := c.st.At(at(2, 2)).Unit, c.st.At(at(6, 2)).Unit
	c.Click(at(2, 2))
	c.Click(at(6, 2))
	if c.walk.Done() {
		t.Fatal("no walk")
	}
	if c.walk.unit != inf {
		t.Errorf("the walk draws %v, want the infantry", c.walk.unit)
	}
	if shown, _ := g.unitShown(at(6, 2)); !shown || c.st.At(at(6, 2)).Unit != car {
		t.Error("the carrier vanished while the infantry walked into it")
	}
}
