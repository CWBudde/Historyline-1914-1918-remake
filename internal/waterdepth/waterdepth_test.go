package waterdepth

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// sea is a 9 by 15 level of terrain with the hexes in each of the rest set to
// waterDeep, waterSea and grass, and the field of it as composed.
func sea(terrain string, deep, open, land []hexgrid.Coord) (*level.Level, *Field) {
	l := &level.Level{Lines: make([][]level.Hex, 15)}
	for y := range l.Lines {
		l.Lines[y] = make([]level.Hex, 9-y&1)
		for x := range l.Lines[y] {
			l.Lines[y][x].Terrain = terrain
		}
	}
	for t, cells := range map[string][]hexgrid.Coord{"waterDeep": deep, "waterSea": open, "grass": land} {
		for _, c := range cells {
			l.Lines[c.Line][c.Col].Terrain = t
		}
	}
	s := mapcompose.Compose(l.Grid(), mapcompose.LevelHex(l), func(string) bool { return true })
	return l, New(l, &s)
}

// Across the edges between shallow water, the middle band and the open sea the
// depth changes smoothly: no step along a line through their hexes.
func TestNoStepAtHexEdges(t *testing.T) {
	_, f := sea("water", []hexgrid.Coord{{Line: 7, Col: 3}}, []hexgrid.Coord{{Line: 7, Col: 4}, {Line: 6, Col: 4}, {Line: 8, Col: 4}}, nil)
	x0, y := hexgrid.Center(hexgrid.Coord{Line: 7, Col: 2})
	x1, _ := hexgrid.Center(hexgrid.Coord{Line: 7, Col: 5})
	last, top := f.At(x0, y), 0.0
	for x := x0; x <= x1; x += 0.25 {
		d := f.At(x, y)
		if math.Abs(d-last) > 0.02 {
			t.Fatalf("a step of %.3f at x %.2f", d-last, x)
		}
		last, top = d, max(top, d)
	}
	if top < 0.99 {
		t.Errorf("the open sea reaches depth %.2f, want 1", top)
	}
}

// A coast is shallow: depth 0 wherever land meets the water, even beside the
// open sea, and nothing on land.
func TestCoastIsShallow(t *testing.T) {
	land := hexgrid.Coord{Line: 7, Col: 4}
	var open []hexgrid.Coord
	for _, n := range hexgrid.Around(land) {
		open = append(open, n, hexgrid.Around(n)[0], hexgrid.Around(n)[3])
	}
	_, f := sea("waterSea", nil, open, []hexgrid.Coord{land})
	cx, cy := hexgrid.Center(land)
	for k, v := range vertices {
		// Unwarp: find the world point whose warped position is the vertex.
		x, y := cx+v[0], cy+v[1]
		if d := f.depthAt(x, y); d != 0 {
			t.Errorf("land vertex %d: depth %.2f", k, d)
		}
	}
	if s := f.Signature(land); s != "" {
		t.Errorf("land hex signature %q, want none", s)
	}
	far, _ := hexgrid.Center(hexgrid.Coord{Line: 1, Col: 1})
	if d := f.At(far, 60); d < 0.99 {
		t.Errorf("open sea away from the coast: depth %.2f", d)
	}
}

// A sea of one depth gives every hex the same signature, so its shaded tiles
// are shared; shallow water alone needs no shade.
func TestSignature(t *testing.T) {
	_, f := sea("waterSea", nil, nil, nil)
	a, b := f.Signature(hexgrid.Coord{Line: 6, Col: 3}), f.Signature(hexgrid.Coord{Line: 8, Col: 5})
	if a == "" || a != b {
		t.Errorf("open sea signatures %q, %q", a, b)
	}
	if _, f := sea("water", nil, nil, nil); f.Signature(hexgrid.Coord{Line: 6, Col: 3}) != "" {
		t.Error("shallow water has a signature")
	}
}

// Shade darkens water by depth and leaves light pixels and shallow water as
// they are.
func TestShade(t *testing.T) {
	_, f := sea("waterSea", nil, nil, nil)
	im := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	water := color.NRGBA{40, 90, 110, 255}
	im.Set(0, 0, water)
	im.Set(1, 0, color.NRGBA{235, 240, 240, 255})
	im.Set(2, 0, color.NRGBA{0, 0, 0, 0})
	x, y := hexgrid.Center(hexgrid.Coord{Line: 7, Col: 4})
	out := f.Shade(im, image.Pt(int(2*x), int(2*y)), 2)
	if got := out.NRGBAAt(0, 0); got.R >= water.R || got.G >= water.G || got.B >= water.B {
		t.Errorf("open sea water %v, want darker than %v", got, water)
	}
	if got := out.NRGBAAt(1, 0); got != (color.NRGBA{235, 240, 240, 255}) {
		t.Errorf("foam %v changed", got)
	}
	if got := out.NRGBAAt(2, 0); got.A != 0 {
		t.Errorf("transparent pixel %v", got)
	}
	_, shallow := sea("water", nil, nil, nil)
	if got := shallow.Shade(im, image.Pt(int(2*x), int(2*y)), 2).NRGBAAt(0, 0); got != water {
		t.Errorf("shallow water %v, want %v", got, water)
	}
}
