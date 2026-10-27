package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestWalkHexFollowsDrawnPosition(t *testing.T) {
	path := []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 1, Col: 0}, {Line: 2, Col: 0}}
	w := newWalkMotion(path, nil, false)
	var got []hexgrid.Coord
	for !w.Done() {
		got = append(got, w.Hex())
		w.Tick()
	}
	if got[0] != path[0] || got[stepFrames/2-1] != path[0] || got[stepFrames/2] != path[1] || got[len(got)-1] != path[2] {
		t.Errorf("walk hexes %v", got)
	}
	if w.Hex() != path[2] {
		t.Errorf("done walk at %v", w.Hex())
	}
}

func TestCount(t *testing.T) {
	if got := count(1, "attack") + ", " + count(0, "move") + ", " + count(3, "move"); got != "1 attack, 0 moves, 3 moves" {
		t.Error(got)
	}
}
