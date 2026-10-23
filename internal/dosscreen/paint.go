package dosscreen

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The reference palette: olive-grey steel from dark to light, and the status
// wells' near-black olive fill with a brown bevel.
var (
	steelDark  = color.RGBA{0x24, 0x24, 0x1C, 0xFF}
	steelLight = color.RGBA{0xB2, 0xB2, 0xA6, 0xFF}
	wellFill   = color.RGBA{0x18, 0x1C, 0x00, 0xFF}
	wellDark   = color.RGBA{0x24, 0x18, 0x0C, 0xFF} // top and left: in shadow
	wellLight  = color.RGBA{0x45, 0x38, 0x28, 0xFF} // bottom and right: lit
)

// Shading of the placeholder steel, as positions on the dark-to-light ramp.
const (
	steelBase  = 0.46 // a plain plate
	lipShade   = 0.06 // the dark lip along the openings
	bevelLight = 0.42 // how far the bevel moves towards light or dark
	steelTile  = 128  // the texture repeats every steelTile device pixels
)

// lightX, lightY is the direction towards the light on the screen: the top left.
var lightX, lightY = -math.Sqrt2 / 2, -math.Sqrt2 / 2

// PaintPanel paints the frame of one panel with the view's openings cut out
// (transparent) and a dark lip along the outline. With a plate painting
// (SetArt) every other cell is one of its plates; without one the frame is drawn
// steel with a bevel lit from the top left and a rivet at every tooth vertex,
// its texture following screen pixels so it runs on across the panels and
// chrome. The image covers the panel's rect in screen pixels.
func PaintPanel(v View, g hexgrid.Grid) *image.RGBA {
	a, _ := currentArt()
	return paintPanel(v, g, a)
}

func paintPanel(v View, g hexgrid.Grid, pa *plateArt) *image.RGBA {
	m := Mask(v, g)
	img := image.NewRGBA(m.Rect)
	fm := metrics(v.Scale)
	if pa != nil {
		fm.bevel = 0
	}
	near := newBuckets(m.Rect, Outline(v, g), fm.lip+fm.bevel+1)
	for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
		for x := m.Rect.Min.X; x < m.Rect.Max.X; x++ {
			open := m.Pix[m.PixOffset(x, y)] != 0
			d, n, ok := near.nearest(float64(x)+0.5, float64(y)+0.5)
			if open && (!ok || d >= 0.5) {
				continue // clear
			}
			if pa != nil && !open {
				wx, wy := v.ScreenToWorld(float64(x)+0.5, float64(y)+0.5)
				c, a := pa.plateAt(v, wx, wy), 1.0
				if ok && d < fm.lip+0.5 {
					// The lip, its inner edge smoothed into the plate.
					c = mixRGBA(ramp(lipShade), c, math.Max(0, d-fm.lip+0.5))
					a = math.Min(1, 0.5+d)
				}
				setRGBA(img, x, y, c, a)
				continue
			}
			t, a := steelAt(x, y), 1.0
			if ok {
				switch {
				case open:
					// Smooth the cut: a pixel inside the opening but on the edge keeps
					// some lip.
					t, a = lipShade, math.Max(0, 0.5-d)
				case d < fm.lip:
					t, a = lipShade, math.Min(1, 0.5+d)
				case d < fm.lip+fm.bevel:
					f := 1 - (d-fm.lip)/fm.bevel
					t += bevelLight * (n.X*lightX + n.Y*lightY) * f
				}
			}
			setRamp(img, x, y, t, a)
		}
	}
	if pa == nil {
		for _, r := range Rivets(v, g) {
			paintRivet(img, r, fm.rivetR, nil)
		}
	}
	return img
}

// PaintChrome paints the fixed pieces of the screen: steel around the panels, the
// divider post with its highlight line, the bolt at its foot and the two recessed
// status wells. The panels are left transparent for PaintPanel.
func PaintChrome(s Screen) *image.RGBA {
	a, _ := currentArt()
	return paintChrome(s, a)
}

func paintChrome(s Screen, a *plateArt) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	p0, p1 := s.Panels[0].image(), s.Panels[1].image()
	for y := range s.H {
		for x := range s.W {
			if pt := (image.Point{x, y}); !pt.In(p0) && !pt.In(p1) {
				img.SetRGBA(x, y, a.chromeSteel(x, y, 0, s.UI))
			}
		}
	}
	paintPost(img, s.Divider, s.UI, a)
	b := s.Bolt
	paintRivet(img, Point{float64(b.X) + float64(b.W)/2, float64(b.Y) + float64(b.H)/2}, float64(b.W)/2-0.5, a)
	for _, w := range s.Status { // after the bolt, which casts its shadow to the right
		paintWell(img, w, scaled(bevelNative, s.UI))
	}
	return img
}

// paintPost shades the divider as a round vertical post: dark outer edges, a body
// that rises towards the middle and a light line on the side facing the light.
func paintPost(img *image.RGBA, r Rect, ui float64, a *plateArt) {
	line := max(int(math.Round(ui)), 1)
	hx := r.X + r.W*3/10
	for y := r.Y; y < r.Bottom(); y++ {
		for x := r.X; x < r.Right(); x++ {
			u := (float64(x-r.X) + 0.5) / float64(r.W)
			t := 0.30 + 0.25*math.Sin(math.Pi*u)
			switch {
			case x == r.X || x == r.Right()-1:
				t = 0.12
			case x >= hx && x < hx+line:
				t = 0.92
			}
			img.SetRGBA(x, y, a.chromeSteel(x, y, t-steelBase, ui))
		}
	}
}

// paintWell fills a status well with a bevel of width bw: dark along the top and
// left, light along the bottom and right, as a recess lit from the top left.
func paintWell(img *image.RGBA, r Rect, bw int) {
	for y := r.Y; y < r.Bottom(); y++ {
		for x := r.X; x < r.Right(); x++ {
			dl := min(y-r.Y, x-r.X)
			dr := min(r.Bottom()-1-y, r.Right()-1-x)
			c := wellFill
			switch {
			case dl < bw && dl <= dr:
				c = wellDark
			case dr < bw:
				c = wellLight
			}
			img.SetRGBA(x, y, c)
		}
	}
}

// paintRivet draws a round rivet of radius rad at c over opaque steel: a dome lit
// from the top left with a soft shadow to the bottom right, in the tones of a's
// steel.
func paintRivet(img *image.RGBA, c Point, rad float64, a *plateArt) {
	lz := 1.5
	l := math.Sqrt(lightX*lightX + lightY*lightY + lz*lz)
	lx, ly, lz := lightX/l, lightY/l, lz/l
	x0, x1 := int(math.Floor(c.X-rad-2)), int(math.Ceil(c.X+rad*1.5+2))
	y0, y1 := int(math.Floor(c.Y-rad-2)), int(math.Ceil(c.Y+rad*1.5+2))
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !(image.Point{x, y}).In(img.Rect) || img.RGBAAt(x, y).A != 255 {
				continue
			}
			dx, dy := float64(x)+0.5-c.X, float64(y)+0.5-c.Y
			d := math.Hypot(dx, dy)
			if sd := math.Hypot(dx-rad*0.4, dy-rad*0.4); sd < rad+0.5 && d >= rad {
				blend(img, x, y, a.tone(0.1), 0.5*math.Min(1, rad+0.5-sd))
			}
			if cover := math.Min(1, rad+0.5-d); cover > 0 {
				nx, ny := dx/rad, dy/rad
				nz := math.Sqrt(math.Max(0, 1-nx*nx-ny*ny))
				blend(img, x, y, a.tone(0.12+0.8*math.Max(0, nx*lx+ny*ly+nz*lz)), cover)
			}
		}
	}
}

// setRamp sets (x, y) to position t on the steel ramp with coverage a.
func setRamp(img *image.RGBA, x, y int, t, a float64) { setRGBA(img, x, y, ramp(t), a) }

// setRGBA sets (x, y) to the opaque colour c with coverage a.
func setRGBA(img *image.RGBA, x, y int, c color.RGBA, a float64) {
	if a >= 1 {
		img.SetRGBA(x, y, c)
		return
	}
	a = math.Max(0, math.Min(1, a))
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(math.Round(float64(c.R) * a)), G: uint8(math.Round(float64(c.G) * a)),
		B: uint8(math.Round(float64(c.B) * a)), A: uint8(math.Round(255 * a)),
	})
}

// blend mixes the opaque colour c over the opaque pixel (x, y) by f.
func blend(img *image.RGBA, x, y int, c color.RGBA, f float64) {
	img.SetRGBA(x, y, mixRGBA(img.RGBAAt(x, y), c, f))
}

// ramp is position t (clamped to 0..1) between steelDark and steelLight.
func ramp(t float64) color.RGBA {
	return rampLUT[int(math.Max(0, math.Min(1, t))*(rampSteps-1)+0.5)]
}

const rampSteps = 1024

// rampLUT holds rampSteps colours evenly along the ramp.
var rampLUT = func() (lut [rampSteps]color.RGBA) {
	for i := range lut {
		t := float64(i) / (rampSteps - 1)
		lerp := func(a, b uint8) uint8 { return uint8(math.Round(float64(a) + t*(float64(b)-float64(a)))) }
		lut[i] = color.RGBA{lerp(steelDark.R, steelLight.R), lerp(steelDark.G, steelLight.G), lerp(steelDark.B, steelLight.B), 255}
	}
	return lut
}()

var (
	steelOnce sync.Once
	steelTex  []float64
)

// steelAt is the placeholder steel texture at screen pixel (x, y) as a ramp
// position: plates of soft mottling with fine horizontal brushing, repeating
// every steelTile pixels in both directions.
func steelAt(x, y int) float64 {
	steelOnce.Do(func() {
		steelTex = make([]float64, steelTile*steelTile)
		for ty := range steelTile {
			for tx := range steelTile {
				u, w := float64(tx)/steelTile, float64(ty)/steelTile
				t := steelBase
				for k, amp := range []float64{0.10, 0.06, 0.04, 0.025} {
					p := 4 << k
					t += amp * (valueNoise(u*float64(p), w*float64(p), p, p, uint32(k)) - 0.5)
				}
				t += 0.05 * (valueNoise(u*2, w*64, 2, 64, 7) - 0.5) // brushing
				steelTex[ty*steelTile+tx] = t
			}
		}
	})
	return steelTex[mod(y, steelTile)*steelTile+mod(x, steelTile)]
}

// valueNoise is smooth noise in 0..1 on a lattice that wraps every px × py cells.
func valueNoise(x, y float64, px, py int, seed uint32) float64 {
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(ix), y-float64(iy)
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	at := func(i, j int) float64 { return lattice(mod(i, px), mod(j, py), seed) }
	top := at(ix, iy) + fx*(at(ix+1, iy)-at(ix, iy))
	bot := at(ix, iy+1) + fx*(at(ix+1, iy+1)-at(ix, iy+1))
	return top + fy*(bot-top)
}

// lattice is a hashed value in 0..1 for lattice point (i, j).
func lattice(i, j int, seed uint32) float64 {
	h := uint32(i)*0x8da6b343 ^ uint32(j)*0xd8163841 ^ seed*0xcb1ab31f
	h ^= h >> 13
	h *= 0x5bd1e995
	h ^= h >> 15
	return float64(h&0xffff) / 0xffff
}

func mod(a, n int) int { return ((a % n) + n) % n }

// buckets index outline segments by screen area, so a pixel only measures its
// distance to the few segments within reach.
type buckets struct {
	rect  image.Rectangle
	size  int
	cols  int
	cells [][]Segment
	reach float64
}

func newBuckets(r image.Rectangle, segs []Segment, reach float64) *buckets {
	const size = 32
	b := &buckets{rect: r, size: size, reach: reach,
		cols: (r.Dx() + size - 1) / size}
	rows := (r.Dy() + size - 1) / size
	b.cells = make([][]Segment, b.cols*rows)
	for _, s := range segs {
		x0 := int(math.Floor((math.Min(s.A.X, s.B.X)-reach)-float64(r.Min.X))) / size
		x1 := int(math.Floor((math.Max(s.A.X, s.B.X)+reach)-float64(r.Min.X))) / size
		y0 := int(math.Floor((math.Min(s.A.Y, s.B.Y)-reach)-float64(r.Min.Y))) / size
		y1 := int(math.Floor((math.Max(s.A.Y, s.B.Y)+reach)-float64(r.Min.Y))) / size
		for by := max(y0, 0); by <= min(y1, rows-1); by++ {
			for bx := max(x0, 0); bx <= min(x1, b.cols-1); bx++ {
				b.cells[by*b.cols+bx] = append(b.cells[by*b.cols+bx], s)
			}
		}
	}
	return b
}

// nearest returns the distance from (x, y) to the closest segment within reach
// and that segment's normal.
func (b *buckets) nearest(x, y float64) (d float64, n Point, ok bool) {
	bx := (int(x) - b.rect.Min.X) / b.size
	by := (int(y) - b.rect.Min.Y) / b.size
	d2 := b.reach * b.reach
	for _, s := range b.cells[by*b.cols+bx] {
		if sd := segDist2(s, x, y); sd < d2 {
			d2, n, ok = sd, s.N, true
		}
	}
	return math.Sqrt(d2), n, ok
}

// segDist2 is the squared distance from (x, y) to segment s.
func segDist2(s Segment, x, y float64) float64 {
	ex, ey := s.B.X-s.A.X, s.B.Y-s.A.Y
	f := ((x-s.A.X)*ex + (y-s.A.Y)*ey) / (ex*ex + ey*ey)
	f = max(0, min(1, f))
	dx, dy := x-s.A.X-f*ex, y-s.A.Y-f*ey
	return dx*dx + dy*dy
}

// mixRGBA is opaque a mixed towards opaque b by f.
func mixRGBA(a, b color.RGBA, f float64) color.RGBA {
	f = math.Max(0, math.Min(1, f))
	mix := func(p, q uint8) uint8 { return uint8(math.Round(float64(p)*(1-f) + float64(q)*f)) }
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}
