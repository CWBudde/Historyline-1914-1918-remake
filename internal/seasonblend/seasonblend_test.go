package seasonblend

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// winterEast is 1915-08 with every hex from column 8 on drawn in winter.
func winterEast(t *testing.T) *level.Level {
	t.Helper()
	l, err := scenario.ShippedLevel("1915-08", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	l.Extra.Seasons = map[hexgrid.Coord]string{}
	each(l.Grid(), func(c hexgrid.Coord) {
		if c.Col >= 8 {
			l.Extra.Seasons[c] = string(mapcompose.Winter)
		}
	})
	return l
}

func all(string) bool { return true }

func TestNewOneSeason(t *testing.T) {
	l, err := scenario.ShippedLevel("1915-08", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	if b := New(l, all); b != nil || b.Seasons(hexgrid.Coord{}) != nil {
		t.Error("a level in one season has a blend")
	}
}

// Only hexes near the border mix, their own season first; the weights sum to
// one, are their own season's deep inside a region, and agree on a point
// whichever hex asks.
func TestBorder(t *testing.T) {
	l := winterEast(t)
	b := New(l, all)
	if b == nil || len(b.Scenes) != 2 {
		t.Fatal("no blend of a summer map with a winter east")
	}
	for c, f := range b.border {
		seasons := f.seasons
		if want := mapcompose.Season(l.Extra.SeasonAt(c)); seasons[0] != want {
			t.Errorf("%v mixes %v, its own season %q not first", c, seasons, want)
		}
		if c.Col < 5 || c.Col > 10 {
			t.Errorf("%v, far from the border, mixes %v", c, seasons)
		}
	}
	for _, c := range []hexgrid.Coord{{Line: 10, Col: 7}, {Line: 10, Col: 8}} {
		if len(b.Seasons(c)) != 2 {
			t.Errorf("%v at the border mixes %v", c, b.Seasons(c))
		}
	}
	w := make([]float64, 2)
	x, y := hexgrid.Center(hexgrid.Coord{Line: 10, Col: 2})
	b.weights(x, y, []mapcompose.Season{mapcompose.Summer, mapcompose.Winter}, w)
	if w[0] != 1 {
		t.Errorf("deep in summer: %v", w)
	}
	x, y = hexgrid.Center(hexgrid.Coord{Line: 10, Col: 7})
	x += 71
	a, z := make([]float64, 2), make([]float64, 2)
	b.weights(x, y, []mapcompose.Season{mapcompose.Summer, mapcompose.Winter}, a)
	b.weights(x, y, []mapcompose.Season{mapcompose.Winter, mapcompose.Summer}, z)
	if math.Abs(a[0]+a[1]-1) > 1e-9 || a[0] != z[1] || a[1] != z[0] || a[0] == 0 || a[1] == 0 {
		t.Errorf("on the border: %v, asked the other way %v", a, z)
	}
}

// Mixing a layer that is the same in every season gives the layer back, and
// the result covers every part.
func TestMixSame(t *testing.T) {
	b := New(winterEast(t), all)
	c := hexgrid.Coord{Line: 10, Col: 8}
	if len(b.Seasons(c)) != 2 {
		t.Fatal("not a border hex")
	}
	im := image.NewNRGBA(image.Rect(0, 0, 192, 160))
	for y := range 160 {
		for x := range 192 {
			im.SetNRGBA(x, y, color.NRGBA{uint8(x), uint8(y), uint8(x ^ y), uint8(255 - y)})
		}
	}
	x, y := hexgrid.Center(c)
	at := image.Pt(int(x-48)*2, int(y-40)*2)
	small := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	layer := []Part{{im, at}}
	out := b.Mix(c, [][]Part{layer, append(layer, Part{small, at.Sub(image.Pt(4, 4))})}, 2)
	if want := image.Rect(at.X-4, at.Y-4, at.X+192, at.Y+160); out.Bounds() != want {
		t.Fatalf("bounds %v, want %v", out.Bounds(), want)
	}
	out = b.Mix(c, [][]Part{layer, layer}, 2)
	for py := range 160 {
		for px := range 192 {
			r, g, bl, a := im.At(px, py).RGBA()
			r2, g2, b2, a2 := out.At(at.X+px, at.Y+py).RGBA()
			if r>>8 != r2>>8 || g>>8 != g2>>8 || bl>>8 != b2>>8 || a>>8 != a2>>8 {
				t.Fatalf("pixel %d,%d: %v, want %v", px, py, out.At(at.X+px, at.Y+py), im.At(px, py))
			}
		}
	}
	if b.Mix(c, [][]Part{nil, nil}, 2) != nil {
		t.Error("an empty layer mixes to an image")
	}
}

// Scattered items come from the scenes of the seasons, the summer ones on the
// summer side.
func TestScattered(t *testing.T) {
	b := New(winterEast(t), all)
	if _, ok := b.Scattered(hexgrid.Coord{Line: 10, Col: 2}, nil); ok {
		t.Error("a summer hex far from the border scatters by season")
	}
	for c := range b.border {
		items := func(s *mapcompose.Scene) []mapcompose.Placed {
			return []mapcompose.Placed{{ID: s.Cells[c.Line][c.Col].Background, X: -30}, {ID: "x", X: 30}}
		}
		got, ok := b.Scattered(c, items)
		if !ok || len(got) != 2 {
			t.Fatalf("%v: %v", c, got)
		}
		for i, p := range got {
			want := items(b.Scenes[b.Seasons(c)[0]])[i]
			alt := items(b.Scenes[b.Seasons(c)[1]])[i]
			if p != want && p != alt {
				t.Errorf("%v item %d: %v", c, i, p)
			}
		}
	}
}
