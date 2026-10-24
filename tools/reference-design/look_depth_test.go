package main

import (
	"io/fs"
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// The sea darkens by the original's steps (waterdepth): the open sea, away
// from any other depth, at 0.82 of the L* it has drawn as shallow water, the
// middle band between that and the shallows. Measured on a winter and a
// summer campaign map against the same map with every depth drawn as water.
func TestSeaDepthSteps(t *testing.T) {
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"campaign/level10_cp.json", "campaign/level7_cp.json"} {
		raw, err := fs.ReadFile(levels.Campaign, name)
		if err != nil {
			t.Fatal(err)
		}
		lvl, err := level.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		flat, err := level.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		g := lvl.Grid()
		families := map[string][]cell{}
		for l, line := range lvl.Lines {
			for c, h := range line {
				if mapcompose.OpenWater(h.Terrain) {
					flat.Lines[l][c].Terrain = "water"
				}
				at := hexgrid.Coord{Line: l, Col: c}
				switch {
				case h.Terrain == "waterDeep":
					families["waterDeep"] = append(families["waterDeep"], cell{Row: l + 1, Col: c + 1})
				case h.Terrain == "waterSea" && within(g, at, 2, func(n hexgrid.Coord) bool { return lvl.Lines[n.Line][n.Col].Terrain == "waterSea" }):
					families["waterSea"] = append(families["waterSea"], cell{Row: l + 1, Col: c + 1})
				}
			}
		}
		measure := func(l *level.Level) map[string]metrics {
			m, err := r.Render(l, compareDensity, false)
			if err != nil {
				t.Fatal(err)
			}
			plane := newLabPlane(halve(m.Image))
			out := map[string]metrics{}
			for f, cells := range families {
				out[f] = plane.measure(cells)
			}
			return out
		}
		deep, shallow := measure(lvl), measure(flat)
		sea, band := deep["waterSea"].L/shallow["waterSea"].L, deep["waterDeep"].L/shallow["waterDeep"].L
		t.Logf("%s: open sea (%d cells) L* %.2f of %.2f (%.3f), middle band (%d cells) L* %.2f of %.2f (%.3f)", lvl.MapName,
			len(families["waterSea"]), deep["waterSea"].L, shallow["waterSea"].L, sea, len(families["waterDeep"]), deep["waterDeep"].L, shallow["waterDeep"].L, band)
		if sea < 0.80 || sea > 0.85 {
			t.Errorf("%s: the open sea at %.3f of shallow water, want 0.80–0.85", lvl.MapName, sea)
		}
		if band <= sea || band >= 0.98 {
			t.Errorf("%s: the middle band at %.3f, want between the open sea's %.3f and 0.98", lvl.MapName, band, sea)
		}
	}
}

// within reports whether every hex of g within rings of c holds (off-map hexes
// do not count).
func within(g hexgrid.Grid, c hexgrid.Coord, rings int, holds func(hexgrid.Coord) bool) bool {
	ring := []hexgrid.Coord{c}
	for range rings {
		var next []hexgrid.Coord
		for _, h := range ring {
			for _, n := range hexgrid.Around(h) {
				if g.Contains(n) && !holds(n) {
					return false
				}
				next = append(next, n)
			}
		}
		ring = next
	}
	return true
}
