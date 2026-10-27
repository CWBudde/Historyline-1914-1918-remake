package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestWalkInterpolatesAlongPath(t *testing.T) {
	path := []hexgrid.Coord{{Line: 2, Col: 2}, {Line: 2, Col: 3}, {Line: 2, Col: 4}}
	w := newWalkMotion(path, nil, false)
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
	if newWalkMotion(path[:1], nil, false) != nil || !newWalkMotion(nil, nil, false).Done() {
		t.Error("a path without steps must not animate")
	}
}
