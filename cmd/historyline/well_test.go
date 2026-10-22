package main

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestWellLines(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1920, 1080)
	at := ownUnits(d)[0]
	u := d.own().Units[at]
	empty := hexgrid.Coord{Line: -1}
	for _, c := range g.splitView(d.side).Cells() {
		if _, ok := d.own().Units[c]; !ok && g.grid.Contains(c) {
			empty = c
			break
		}
	}
	mode := modeWord(d.s.Phase(d.side))
	tests := []struct {
		name       string
		setup      func()
		side       int
		want       string // a part of the line
		unit       bool
		progressed bool
	}{
		{"cursor on an own unit", func() { g.setCursor(at) }, d.side, strings.ToUpper(u.Name), true, false},
		{"cursor on an empty cell", func() { g.setCursor(empty) }, d.side, mode + " PHASE  0 ORDERS", false, false},
		{"a fresh message beats the unit line", func() {
			g.setCursor(at)
			d.Message = "Nothing to attack there"
			g.tickWells()
		}, d.side, "NOTHING TO ATTACK THERE", false, false},
		{"an old message gives way", func() {
			g.setCursor(at)
			d.Message = "old"
			for range msgFrames + 1 {
				g.tickWells()
			}
		}, d.side, strings.ToUpper(u.Name), true, false},
		{"the other side's phase", nil, 1 - d.side, modeWord(d.s.Phase(1-d.side)) + " PHASE", false, false},
		{"orders playing", func() { d.queue = []dosplay.Event{{Kind: dosplay.Skipped}} }, d.side, "ORDERS UNDER WAY", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d.Message, d.queue = "", nil
			g.split.lastMsg, g.split.msgAge = "", msgFrames
			if tt.setup != nil {
				tt.setup()
			}
			w := g.wellOf(tt.side)
			got := w.text
			if w.unit {
				got = w.name
			}
			if w.unit != tt.unit || !strings.Contains(got, tt.want) || (w.progress >= 0) != tt.progressed {
				t.Errorf("well %+v, want %q (unit line %v)", w, tt.want, tt.unit)
			}
			if w.unit && (w.size != u.Size || w.owner != u.Owner) {
				t.Errorf("unit line %+v for %+v", w, u)
			}
		})
	}
	d.Message, d.queue = "", nil

	// The other side confirms: it waits for the human side's phase to end.
	d.s.Confirm(1 - d.side)
	if w := g.wellOf(1 - d.side); !strings.HasPrefix(w.text, "WAITING: ENEMY") {
		t.Errorf("ready side's well %q", w.text)
	}
	// Both wells draw at every UI scale without leaving them.
	for _, ui := range []int{1, 2} {
		g.applyLayout(1920*ui, 1080*ui, float64(ui))
		g.layoutSplit()
		img := ebiten.NewImage(1920*ui, 1080*ui)
		g.drawWells(img)
		for side := range 2 {
			r, k := g.split.screen.WellText(side), wellScale(g.split.screen, side)
			if k < ui || 5*k > r.H {
				t.Errorf("ui %d side %d: font scale %d in a well %d high", ui, side, k, r.H)
			}
		}
	}
}

func TestWellEnd(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1280, 720)
	d.over = true
	if w := g.wellOf(d.side); w.text != strings.ToUpper(d.endText()) {
		t.Errorf("end well %q, want %q", w.text, d.endText())
	}
}
