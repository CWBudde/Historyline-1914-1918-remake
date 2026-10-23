package terrainlight

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// disc is a canvas of w×h px with opaque discs of diameter d px at centres.
func disc(w, h int, d float64, centres ...[2]float64) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			for _, c := range centres {
				if math.Hypot(float64(x)+0.5-c[0], float64(y)+0.5-c[1]) <= d/2 {
					im.SetNRGBA(x, y, color.NRGBA{60, 70, 30, 255})
				}
			}
		}
	}
	return im
}

func centroid(im *image.NRGBA) (x, y, w float64) {
	for py := range im.Rect.Dy() {
		for px := range im.Rect.Dx() {
			a := float64(im.NRGBAAt(px, py).A)
			x, y, w = x+a*(float64(px)+0.5), y+a*(float64(py)+0.5), w+a
		}
	}
	return x / w, y / w, w
}

// The diameter is the median equivalent diameter of the opaque blobs, in
// world px.
func TestDiameter(t *testing.T) {
	for _, density := range []int{2, 4} {
		f := float64(density)
		im := disc(64*density, 64*density, 12*f, [2]float64{16 * f, 16 * f}, [2]float64{40 * f, 20 * f})
		if d := Diameter(im, density); math.Abs(d-12) > 0.3 {
			t.Errorf("%dx: diameter %.2f, want 12", density, d)
		}
	}
	if d := Diameter(image.NewNRGBA(image.Rect(0, 0, 8, 8)), 2); d != 0 {
		t.Errorf("empty canvas: diameter %v", d)
	}
}

// The shadow falls away from the light at 315°, so along 135° (down and
// right), 0.25 × the diameter long, in the shadow colour at the opacity; the
// canvas stays the same.
func TestCast(t *testing.T) {
	for _, density := range []int{2, 4} {
		f := float64(density)
		src := disc(64*density, 56*density, 16*f, [2]float64{32 * f, 28 * f})
		sh := Cast(src, density, 16, Crown)
		if sh.Rect != src.Rect {
			t.Fatalf("%dx: canvas %v, want %v", density, sh.Rect, src.Rect)
		}
		x0, y0, w0 := centroid(src)
		x1, y1, w1 := centroid(sh)
		dx, dy := (x1-x0)/f, (y1-y0)/f
		if l := math.Hypot(dx, dy); math.Abs(l-Length*16) > 0.1 {
			t.Errorf("%dx: length %.2f world px, want %.2f", density, l, Length*16)
		}
		if bearing := math.Atan2(dx, -dy) * 180 / math.Pi; math.Abs(bearing-(Azimuth-180)) > 0.5 {
			t.Errorf("%dx: bearing %.2f°, want %v", density, bearing, Azimuth-180)
		}
		if r := w1 / w0; math.Abs(r-Crown) > 0.01 {
			t.Errorf("%dx: alpha mass %.3f of the item, want %v", density, r, Crown)
		}
		c := sh.NRGBAAt(int(x1), int(y1))
		if c.R != Colour.R || c.G != Colour.G || c.B != Colour.B || math.Abs(float64(c.A)-255*Crown) > 1 {
			t.Errorf("%dx: shadow core %v, want %v at %v", density, c, Colour, Crown)
		}
	}
}

// Only painted shadows of a family with a light class are cast again; the
// scattered items' shadows come from hlkit with the same model.
func TestRecast(t *testing.T) {
	for _, tc := range []struct {
		id, art, standing string
		opacity           float64
		ok                bool
	}{
		{"standing/shadow/props/trees/1", "terrain/forest", "standing/props/trees/1", Crown, true},
		{"standing/shadow/context/forest/ab12", "terrain/forest", "standing/context/forest/ab12", Crown, true},
		{"standing/shadow/forest", "terrain/forest", "standing/forest", Crown, true},
		{"standing/shadow/scatter/crown/oak-16", "terrain/forest", "", 0, false},
		{"standing/props/trees/1", "terrain/forest", "", 0, false},
		{"standing/shadow/bridge14", "terrain/bridge", "", 0, false},
	} {
		standing, opacity, ok := Recast(tc.id, tc.art)
		if standing != tc.standing || opacity != tc.opacity || ok != tc.ok {
			t.Errorf("Recast(%s) = %s %v %v", tc.id, standing, opacity, ok)
		}
	}
}
