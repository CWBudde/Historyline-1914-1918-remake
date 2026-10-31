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
	// In this order, so land stays land where a list of water names it too.
	for _, set := range []struct {
		terrain string
		cells   []hexgrid.Coord
	}{{"waterDeep", deep}, {"waterSea", open}, {"grass", land}} {
		for _, c := range set.cells {
			l.Lines[c.Line][c.Col].Terrain = set.terrain
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

// A sea of one depth gives every hex of one sea phase the same signature, so
// its shaded tiles are shared; shallow water alone needs no shade.
func TestSignature(t *testing.T) {
	_, f := sea("waterSea", nil, nil, nil)
	a, b := f.Signature(hexgrid.Coord{Line: 6, Col: 3}), f.Signature(hexgrid.Coord{Line: 10, Col: 5})
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
	out := f.Shade(im, Tiles{}, image.Pt(int(2*x), int(2*y)), 2)
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
	if got := shallow.Shade(im, Tiles{}, image.Pt(int(2*x), int(2*y)), 2).NRGBAAt(0, 0); got != water {
		t.Errorf("shallow water %v, want %v", got, water)
	}
}

// A hex's sea tile is the phase of the 284 by 160 world pixel sea period at its
// tile's top left, so neighbouring tiles continue the one material.
func TestSeaTilePhase(t *testing.T) {
	seen := map[string][2]float64{}
	for line := range 12 {
		for col := range 6 {
			c := hexgrid.Coord{Line: line, Col: col}
			x, y := hexgrid.Center(c)
			phase := [2]float64{math.Mod(x-48, 284), math.Mod(y-40, 160)}
			id := SeaTile(c, "tile/water")
			if p, ok := seen[id]; ok && p != phase {
				t.Errorf("%s at %+v has phase %v, earlier %v", id, c, phase, p)
			}
			seen[id] = phase
			if w := SeaTile(c, "tile/water_w"); w != id+"_w" {
				t.Errorf("winter sea tile %s, want %s_w", w, id)
			}
		}
	}
	if len(seen) != 8 {
		t.Errorf("%d sea phases, want 8", len(seen))
	}
}

// In the open sea the sea's material replaces the water, the open water tile's
// light ripple crests included; land pixels and the shallows keep theirs.
func TestShadeBlendsTheSea(t *testing.T) {
	_, f := sea("waterSea", nil, nil, nil)
	im := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	im.Set(0, 0, color.NRGBA{40, 90, 110, 255})
	im.Set(1, 0, color.NRGBA{100, 107, 40, 255})
	im.Set(2, 0, color.NRGBA{190, 230, 235, 255})
	water := image.NewNRGBA(im.Rect)
	water.Set(0, 0, im.NRGBAAt(0, 0))
	water.Set(2, 0, im.NRGBAAt(2, 0))
	s := image.NewNRGBA(im.Rect)
	s.Set(0, 0, color.NRGBA{30, 80, 95, 255})
	s.Set(1, 0, color.NRGBA{34, 84, 99, 255})
	s.Set(2, 0, color.NRGBA{32, 82, 97, 255})
	x, y := hexgrid.Center(hexgrid.Coord{Line: 7, Col: 4})
	at := image.Pt(int(2*x), int(2*y))
	plain, mixed := f.Shade(im, Tiles{}, at, 2), f.Shade(im, Tiles{Sea: s, Water: water}, at, 2)
	if plain.NRGBAAt(0, 0) == mixed.NRGBAAt(0, 0) {
		t.Errorf("open sea water %v did not take the sea", mixed.NRGBAAt(0, 0))
	}
	if got := mixed.NRGBAAt(1, 0); got != plain.NRGBAAt(1, 0) {
		t.Errorf("meadow %v took the sea, want %v", got, plain.NRGBAAt(1, 0))
	}
	if got := mixed.NRGBAAt(2, 0); got.R > 100 {
		t.Errorf("a ripple crest of the water tile stays %v on the open sea", got)
	}
	_, shallow := sea("water", nil, nil, nil)
	if got := shallow.Shade(im, Tiles{Sea: s}, at, 2).NRGBAAt(0, 0); got != im.NRGBAAt(0, 0) {
		t.Errorf("shallow water %v took the sea", got)
	}
}

// Surf lies on the water beside land where the sea is near, and nowhere by a
// lake or on the land itself.
func TestSurf(t *testing.T) {
	water, meadow := color.NRGBA{30, 80, 100, 255}, color.NRGBA{100, 107, 40, 255}
	im := image.NewNRGBA(image.Rect(0, 0, 96, 80))
	for y := range 80 {
		for x := range 96 {
			im.Set(x, y, water)
			if x < 10 {
				im.Set(x, y, meadow)
			}
		}
	}
	surf := image.NewNRGBA(image.Rect(0, 0, 96, 80))
	for y := range 80 {
		for x := range 96 {
			surf.Set(x, y, color.NRGBA{220, 228, 226, 255})
		}
	}
	coast := hexgrid.Coord{Line: 7, Col: 4}
	x, y := hexgrid.Center(coast)
	at := image.Pt(int(x)-48, int(y)-40)
	_, open := sea("waterSea", nil, nil, nil)
	out, plain := open.Shade(im, Tiles{Surf: surf}, at, 1), open.Shade(im, Tiles{}, at, 1)
	if got := out.NRGBAAt(12, 10); got.R <= 150 {
		t.Errorf("water beside land by the sea %v, want foam", got)
	}
	if got := out.NRGBAAt(5, 10); got != plain.NRGBAAt(5, 10) {
		t.Errorf("land %v took foam", got)
	}
	_, lake := sea("water", []hexgrid.Coord{}, nil, nil)
	if got := lake.Shade(im, Tiles{Surf: surf}, at, 1).NRGBAAt(12, 10); got != water {
		t.Errorf("water by a lake %v, want no foam", got)
	}
}

// The water tiles follow the season of the background, and an autumn or spring
// sea has the depths of its summer one.
func TestSeasonalWater(t *testing.T) {
	c := hexgrid.Coord{Line: 3, Col: 2}
	for _, sfx := range []string{"", "_w", "_au", "_sp"} {
		bg := "tile/shore-12" + sfx
		if SeaTile(c, bg) != SeaTile(c, "tile/water")+sfx || WaterTile(bg) != "tile/water"+sfx || SurfTile(bg) != "tile/surf"+sfx {
			t.Errorf("%s: %s, %s, %s", bg, SeaTile(c, bg), WaterTile(bg), SurfTile(bg))
		}
	}
	l, summer := sea("waterSea", nil, nil, []hexgrid.Coord{{Line: 7, Col: 4}})
	s := mapcompose.ComposeSeason(l.Grid(), mapcompose.LevelHex(l), func(string) bool { return true }, mapcompose.Autumn)
	autumn := New(l, &s)
	for _, c := range []hexgrid.Coord{{Line: 6, Col: 3}, {Line: 7, Col: 3}, {Line: 2, Col: 2}} {
		if a, b := autumn.Signature(c), summer.Signature(c); a != b || a == "" {
			t.Errorf("%v: autumn signature %q, summer %q", c, a, b)
		}
	}
}
