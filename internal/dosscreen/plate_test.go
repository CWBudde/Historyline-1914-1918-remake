package dosscreen

import (
	"image"
	"image/color"
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
	img := image.NewRGBA(image.Rect(0, 0, 2*96+40, 2*80+40))
	for y := 20; y < 20+2*80; y++ {
		for x := 20; x < 20+2*96; x++ {
			c := color.RGBA{200, 40, 40, 255}
			if x >= 20+96 {
				c = color.RGBA{40, 40, 200, 255}
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
	for _, s := range []float64{0.5, 1, 2.25} {
		v := testView(s, -1, -1) // at the map corner: some cells are off the map
		img := paintPanel(v, g, pl)
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

// Setting a plate painting repaints the panel frames, not the chrome.
func TestFrameRepaintsOnPlate(t *testing.T) {
	defer SetPlate(nil)
	g := hexgrid.Grid{Width: 40, Lines: 40}
	s := Layout(640, 400, 1)
	dst := ebiten.NewImage(640, 400)
	var f Frame
	defer f.Release()
	views := [2]View{s.View(0, 5, 5), s.View(1, 9, 9)}
	f.Draw(dst, s, views, g)
	n := f.painted
	SetPlate(halfPlate())
	f.Draw(dst, s, views, g)
	if f.painted != n+2 {
		t.Errorf("a new plate painted %d images, want both panels again", f.painted-n)
	}
}
