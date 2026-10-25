package main

import (
	"math"
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

// The sea darkens by the original's steps (waterdepth): the open sea, away
// from any other depth, at 0.82 of the L* it has drawn as shallow water, the
// middle band between that and the shallows (away from the coast, whose surf
// lightens it). Measured on a winter and a summer campaign map against the same
// map with every depth drawn as water.
func TestSeaDepthSteps(t *testing.T) {
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"monthly/1916-02-parts:cp", "monthly/1915-08-orkan:cp"} {
		lvl, flat := testLevel(t, name), testLevel(t, name)
		g := lvl.Grid()
		families := map[string][]cell{}
		for l, line := range lvl.Lines {
			for c, h := range line {
				if mapcompose.OpenWater(h.Terrain) {
					flat.Lines[l][c].Terrain = "water"
				}
				at := hexgrid.Coord{Line: l, Col: c}
				switch {
				case h.Terrain == "waterDeep" && within(g, at, 1, func(n hexgrid.Coord) bool { return mapcompose.OpenWater(lvl.Lines[n.Line][n.Col].Terrain) }):
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

// The open sea does not repeat with the hex lattice: the L* of open-sea hexes
// (two rings of sea around them) correlates with the same place one lattice
// step away, (142, 0) or (71, 40) world px, by less than 0.5. Drawn with the
// water tile alone every hex is the same and the correlation is about 1.
func TestSeaRepeat(t *testing.T) {
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	lvl := testLevel(t, "monthly/1916-02-parts:cp")
	g := lvl.Grid()
	open := map[hexgrid.Coord]bool{}
	for l, line := range lvl.Lines {
		for c, h := range line {
			at := hexgrid.Coord{Line: l, Col: c}
			if h.Terrain == "waterSea" && within(g, at, 2, func(n hexgrid.Coord) bool { return lvl.Lines[n.Line][n.Col].Terrain == "waterSea" }) {
				open[at] = true
			}
		}
	}
	m, err := r.Render(lvl, compareDensity, false)
	if err != nil {
		t.Fatal(err)
	}
	plane := newLabPlane(halve(m.Image))
	for _, step := range []struct {
		name   string
		dx, dy int
		next   func(hexgrid.Coord) hexgrid.Coord
	}{
		{"(142, 0)", 142, 0, func(c hexgrid.Coord) hexgrid.Coord { return hexgrid.Coord{Line: c.Line, Col: c.Col + 1} }},
		{"(71, 40)", 71, 40, func(c hexgrid.Coord) hexgrid.Coord { return hexgrid.Coord{Line: c.Line + 1, Col: c.Col + (c.Line & 1)} }},
	} {
		var n, sa, sb, saa, sbb, sab float64
		for c := range open {
			if !open[step.next(c)] {
				continue
			}
			ox, oy := tileOrigin(c.Line, c.Col)
			for y := 24; y < 56; y++ {
				for x := 28; x < 68; x++ {
					i := (oy+y)*plane.w + ox + x
					j := (oy+y+step.dy)*plane.w + ox + x + step.dx
					if !plane.ok[i] || !plane.ok[j] {
						continue
					}
					a, b := plane.l[i], plane.l[j]
					n++
					sa, sb, saa, sbb, sab = sa+a, sb+b, saa+a*a, sbb+b*b, sab+a*b
				}
			}
		}
		if n < 1000 {
			t.Fatalf("%s: only %.0f pixel pairs of open sea", step.name, n)
		}
		corr := (sab/n - sa/n*sb/n) / math.Sqrt((saa/n-sa/n*sa/n)*(sbb/n-sb/n*sb/n))
		t.Logf("%s: L* correlation %.3f over %.0f pixel pairs in %d open-sea hexes", step.name, corr, n, len(open))
		if corr >= 0.5 {
			t.Errorf("%s: open sea repeats one lattice step away, L* correlation %.3f, want below 0.5", step.name, corr)
		}
	}
}
