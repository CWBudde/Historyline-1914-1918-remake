package dosscreen

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The panel frame is opaque steel in the reference palette with the visible cells
// cut out.
func TestPaintPanelCutsOutCells(t *testing.T) {
	g := hexgrid.Grid{Width: 6, Lines: 12}
	for _, s := range []float64{0.5, 1, 2.25} {
		v := testView(s, 1, 2)
		img := PaintPanel(v, g)
		m := Mask(v, g)
		if img.Rect != m.Rect {
			t.Fatalf("scale %v: image %v, mask %v", s, img.Rect, m.Rect)
		}
		for _, c := range v.Cells() {
			cx, cy := v.WorldToScreen(hexgrid.Center(c))
			a := img.RGBAAt(int(cx), int(cy)).A
			if g.Contains(c) && a != 0 || !g.Contains(c) && a != 255 {
				t.Errorf("scale %v: cell %v centre alpha %d", s, c, a)
			}
		}
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
				p := img.RGBAAt(x, y)
				if p.A == 255 && !inPalette(p) {
					t.Fatalf("scale %v: steel %v at %d,%d outside the palette", s, p, x, y)
				}
				if m.AlphaAt(x, y).A == 0 && p.A < 128 {
					t.Fatalf("scale %v: steel pixel %d,%d is transparent", s, x, y)
				}
			}
		}
	}
}

// Light falls from the top left: the bevel facing it (the steel below and right of
// an opening) is brighter than the bevel facing away, and the lip along the
// opening is darker than both.
func TestPaintPanelBevelAndLip(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	for _, s := range []float64{0.6, 1, 1.7333, 3} {
		v := testView(s, 3, 3)
		img := PaintPanel(v, g)
		fm := metrics(s)
		var lit, shade, lip []float64
		for _, sg := range Outline(v, g) {
			mx, my := (sg.A.X+sg.B.X)/2, (sg.A.Y+sg.B.Y)/2
			at := func(d float64) (float64, bool) {
				x, y := int(math.Floor(mx-d*sg.N.X)), int(math.Floor(my-d*sg.N.Y))
				if !(image.Point{x, y}).In(img.Rect) {
					return 0, false
				}
				return luma(img.RGBAAt(x, y)), true
			}
			facing := -(sg.N.X + sg.N.Y) / math.Sqrt2 // the bevel slopes down towards N

			if l, ok := at(fm.lip + fm.bevel*0.4); ok {
				switch {
				case facing > 0.5:
					lit = append(lit, l)
				case facing < -0.5:
					shade = append(shade, l)
				}
			}
			if l, ok := at(fm.lip * 0.5); ok {
				lip = append(lip, l)
			}
		}
		if len(lit) == 0 || len(shade) == 0 || len(lip) == 0 {
			t.Fatalf("scale %v: no samples (%d lit, %d shaded, %d lip)", s, len(lit), len(shade), len(lip))
		}
		ml, ms, mp := mean(lit), mean(shade), mean(lip)
		if ml < ms+30 || mp > ms-5 {
			t.Errorf("scale %v: lit bevel %.0f, shaded bevel %.0f, lip %.0f", s, ml, ms, mp)
		}
	}
}

// The status wells are recessed: near-black olive inside, dark brown along the top
// and left, light brown along the bottom and right. The divider post carries a
// highlight line and the bolt sits between the status bars; the panels are left
// clear.
func TestPaintChrome(t *testing.T) {
	for _, ui := range []float64{1, 2} {
		s := Layout(1280, 720, ui)
		img := PaintChrome(s)
		if img.Rect != image.Rect(0, 0, 1280, 720) {
			t.Fatalf("chrome %v", img.Rect)
		}
		for _, p := range s.Panels {
			if a := img.RGBAAt(p.X+p.W/2, p.Y+p.H/2).A; a != 0 {
				t.Errorf("ui %v: panel covered (alpha %d)", ui, a)
			}
		}
		for _, w := range s.Status {
			cx, cy := w.X+w.W/2, w.Y+w.H/2
			for _, tc := range []struct {
				x, y int
				want color.RGBA
			}{
				{cx, cy, wellFill},
				{cx, w.Y, wellDark},
				{w.X, cy, wellDark},
				{cx, w.Bottom() - 1, wellLight},
				{w.Right() - 1, cy, wellLight},
			} {
				if got := img.RGBAAt(tc.x, tc.y); got != tc.want {
					t.Errorf("ui %v: well %+v at %d,%d is %v, want %v", ui, w, tc.x, tc.y, got, tc.want)
				}
			}
		}
		d := s.Divider
		y := d.H / 2
		best, edge := 0.0, math.Min(luma(img.RGBAAt(d.X, y)), luma(img.RGBAAt(d.Right()-1, y)))
		for x := d.X; x < d.Right(); x++ {
			best = math.Max(best, luma(img.RGBAAt(x, y)))
		}
		if best < edge+60 {
			t.Errorf("ui %v: divider highlight %.0f against edges %.0f", ui, best, edge)
		}
		b := s.Bolt
		if c := img.RGBAAt(b.X+b.W/2, b.Y+b.H/2); c.A != 255 || !inPalette(c) {
			t.Errorf("ui %v: bolt centre %v", ui, c)
		}
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
				pt := image.Point{x, y}
				if c := img.RGBAAt(x, y); c.A == 255 && !inPalette(c) && !pt.In(s.Status[0].image()) && !pt.In(s.Status[1].image()) {
					t.Fatalf("ui %v: chrome %v at %d,%d outside the palette", ui, c, x, y)
				}
			}
		}
	}
}

// The steel texture tiles seamlessly: across the wrap it changes no more than
// between any two neighbouring pixels inside the tile.
func TestSteelTiles(t *testing.T) {
	inside := 0.0
	for y := range steelTile - 1 {
		for x := range steelTile - 1 {
			inside = math.Max(inside, math.Max(math.Abs(steelAt(x+1, y)-steelAt(x, y)), math.Abs(steelAt(x, y+1)-steelAt(x, y))))
		}
	}
	for k := range steelTile {
		if d := math.Abs(steelAt(steelTile-1, k) - steelAt(steelTile, k)); d > inside+1e-12 {
			t.Errorf("row %d jumps by %.4f at the wrap (inside at most %.4f)", k, d, inside)
		}
		if d := math.Abs(steelAt(k, -1) - steelAt(k, 0)); d > inside+1e-12 {
			t.Errorf("column %d jumps by %.4f at the wrap (inside at most %.4f)", k, d, inside)
		}
	}
}

func inPalette(c color.RGBA) bool {
	return c.R >= steelDark.R && c.R <= steelLight.R && c.G >= steelDark.G && c.G <= steelLight.G &&
		c.B >= steelDark.B && c.B <= steelLight.B
}

func luma(c color.RGBA) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }

func mean(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}
