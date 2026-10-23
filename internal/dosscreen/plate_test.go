package dosscreen

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// cellCentre finds the hex polygon around a point, off the map as well: the
// centre itself and points just inside each corner.
func TestCellCentre(t *testing.T) {
	for line := -3; line <= 3; line++ {
		for col := -2; col <= 2; col++ {
			c := hexgrid.Coord{Line: line, Col: col}
			cx, cy := hexgrid.Center(c)
			pts := [][2]float64{{cx, cy}}
			for _, v := range hexgrid.Corners(c) {
				pts = append(pts, [2]float64{cx + (v[0]-cx)*0.95, cy + (v[1]-cy)*0.95})
			}
			for _, p := range pts {
				if x, y := cellCentre(p[0], p[1]); x != cx || y != cy {
					t.Errorf("point %v of %v: centre (%v, %v), want (%v, %v)", p, c, x, y, cx, cy)
				}
			}
		}
	}
}

// halfPlate is a plate painting on a clear margin: its 96×80 box left half
// one colour, right half another.
func halfPlate() *image.RGBA {
	return twoTone(color.RGBA{200, 40, 40, 255}, color.RGBA{40, 40, 200, 255})
}

// flatPlate is a plate painting of one colour on a clear margin.
func flatPlate(c color.RGBA) *image.RGBA { return twoTone(c, c) }

func twoTone(l, r color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2*96+40, 2*80+40))
	for y := 20; y < 20+2*80; y++ {
		for x := 20; x < 20+2*96; x++ {
			c := l
			if x >= 20+96 {
				c = r
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// With a plate painting every steel cell shows the plate: left of a covered
// cell's centre its left half, right of it its right half; the openings stay
// clear and the lip along them dark.
func TestPaintPanelPlates(t *testing.T) {
	g := hexgrid.Grid{Width: 6, Lines: 12}
	pl := newPlateTex(halfPlate())
	if pl == nil || pl.mips[0].w != 2*96 || pl.mips[0].h != 2*80 {
		t.Fatalf("plate box not found: %+v", pl)
	}
	pa := &plateArt{plates: []*plateTex{pl}}
	for _, s := range []float64{0.5, 1, 2.25} {
		v := testView(s, -1, -1) // at the map corner: some cells are off the map
		img := paintPanel(v, g, pa)
		covered := 0
		for _, c := range v.Cells() {
			cx, cy := v.WorldToScreen(hexgrid.Center(c))
			if g.Contains(c) {
				if a := img.RGBAAt(int(cx), int(cy)).A; a != 0 {
					t.Errorf("scale %v: opening %v alpha %d", s, c, a)
				}
				continue
			}
			covered++
			dx := 24 * s
			l, r := img.RGBAAt(int(cx-dx), int(cy)), img.RGBAAt(int(cx+dx), int(cy))
			if l.R <= l.B || r.B <= r.R || l.A != 255 || r.A != 255 {
				t.Errorf("scale %v: covered cell %v shows %v left, %v right", s, c, l, r)
			}
		}
		if covered == 0 {
			t.Fatalf("scale %v: no covered cell in the view", s)
		}
		m, fm := Mask(v, g), metrics(s)
		near := newBuckets(img.Rect, Outline(v, g), fm.lip+1)
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
				d, _, ok := near.nearest(float64(x)+0.5, float64(y)+0.5)
				if p := img.RGBAAt(x, y); ok && d < fm.lip-0.5 && m.AlphaAt(x, y).A == 0 && (p.R > 60 || p.B > 60) {
					t.Fatalf("scale %v: lip pixel %d,%d is %v", s, x, y, p)
				}
			}
		}
	}
}

// With several plates each cell picks one by its place in the view: all of
// them show, and a view scrolled elsewhere shows the same plate in the same
// place, so scrolling needs no repaint.
func TestPlateVariants(t *testing.T) {
	cols := []color.RGBA{{200, 40, 40, 255}, {40, 200, 40, 255}, {40, 40, 200, 255}}
	pa := &plateArt{}
	for _, c := range cols {
		pa.plates = append(pa.plates, newPlateTex(flatPlate(c)))
	}
	a, b := testView(1, 0, 0), testView(1, 7, 5)
	seen := map[color.RGBA]bool{}
	for j := -1; j <= a.Cols; j++ {
		for i := -1; i <= a.Rows; i++ {
			ax, ay := hexgrid.Center(a.Cell(j, i))
			bx, by := hexgrid.Center(b.Cell(j, i))
			got := pa.plateAt(a, ax, ay)
			if other := pa.plateAt(b, bx, by); got != other {
				t.Errorf("cell %d,%d: %v here, %v scrolled", j, i, got, other)
			}
			seen[got] = true
		}
	}
	for _, c := range cols {
		if !seen[c] {
			t.Errorf("plate %v never shows", c)
		}
	}
}

// With art the chrome is steel of the plates' colour, and the divider keeps
// its dark edges and highlight line.
func TestPaintChromeArt(t *testing.T) {
	face := color.RGBA{90, 110, 70, 255}
	pa := &plateArt{plates: []*plateTex{newPlateTex(flatPlate(face))}}
	pa.face = meanRGBA(pa.plates[0].mean())
	if pa.face != face {
		t.Fatalf("plate mean %v, want %v", pa.face, face)
	}
	s := Layout(1280, 720, 1)
	img := paintChrome(s, pa)
	p := s.Panels[0]
	strip := img.RGBAAt(p.X+p.W/2, p.Bottom()+1)
	if math.Abs(luma(strip)-luma(face)) > 25 || strip.G <= strip.R || strip.G <= strip.B {
		t.Errorf("steel strip %v, plates %v", strip, face)
	}
	d := s.Divider
	y := d.H / 2
	best, edge := 0.0, math.Min(luma(img.RGBAAt(d.X, y)), luma(img.RGBAAt(d.Right()-1, y)))
	for x := d.X; x < d.Right(); x++ {
		best = math.Max(best, luma(img.RGBAAt(x, y)))
	}
	if best < edge+60 {
		t.Errorf("divider highlight %.0f against edges %.0f", best, edge)
	}
}

func meanRGBA(m [3]float64) color.RGBA {
	return color.RGBA{uint8(math.Round(m[0])), uint8(math.Round(m[1])), uint8(math.Round(m[2])), 255}
}

// Setting the frame art repaints the panel frames and the chrome.
func TestFrameRepaintsOnArt(t *testing.T) {
	defer SetArt(Art{})
	g := hexgrid.Grid{Width: 40, Lines: 40}
	s := Layout(640, 400, 1)
	dst := ebiten.NewImage(640, 400)
	var f Frame
	defer f.Release()
	views := [2]View{s.View(0, 5, 5), s.View(1, 9, 9)}
	f.Draw(dst, s, views, g)
	n := f.painted
	SetArt(Art{Plates: []image.Image{halfPlate()}})
	f.Draw(dst, s, views, g)
	if f.painted != n+3 {
		t.Errorf("new art painted %d images, want the chrome and both panels again", f.painted-n)
	}
}
