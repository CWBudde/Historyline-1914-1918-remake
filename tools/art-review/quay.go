package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// A quay pier (art/terrain/quay) is painted once as a horizontal strip and
// turned to the three port axes. The strip is made periodic over the distance
// between two hex centres on the axis, with every hex centre at phase zero, so
// a chain of quay hexes runs on without a seam.
const (
	quayWidth        = 26.  // world px across the whole strip, fenders included
	quayShadowOffset = 1.6  // world px toward the lower right, along each axis
	quayShadowAlpha  = 0.42 // opacity of the shadow on the water
)

var quayShadowColor = color.NRGBA{16, 30, 32, 255}

// quayAxes are the pier axes: the step between two hex centres along ports
// 3–6, 1–4 and 2–5, in world px.
var quayAxes = map[string]connectionPoint{"quay36": {0, 80}, "quay14": {71, -40}, "quay25": {71, 40}}

// quayStrip is the painted strip: its opaque rows and source px per world px.
type quayStrip struct {
	im         *image.NRGBA
	centre, k  float64 // centre row; source px per world px
	start, end float64 // the usable columns
}

func newQuayStrip(im *image.NRGBA) (quayStrip, error) {
	b := im.Bounds()
	top, bottom := -1, -1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		solid := 0
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A > 128 {
				solid++
			}
		}
		if solid*2 > b.Dx() {
			if top < 0 {
				top = y
			}
			bottom = y
		}
	}
	if top < 0 || bottom-top < 32 {
		return quayStrip{}, fmt.Errorf("quay strip: no opaque band")
	}
	return quayStrip{im: im, centre: float64(top+bottom+1) / 2, k: float64(bottom-top+1) / quayWidth, start: float64(b.Min.X) + 4, end: float64(b.Max.X) - 4}, nil
}

// sample is the premultiplied colour at strip column x and row y, bilinear,
// with the alpha haze of the painting cleared.
func (s quayStrip) sample(x, y float64) [4]float64 {
	x0, y0 := math.Floor(x-.5), math.Floor(y-.5)
	fx, fy := x-.5-x0, y-.5-y0
	var out [4]float64
	for _, c := range [4]struct {
		dx, dy int
		w      float64
	}{{0, 0, (1 - fx) * (1 - fy)}, {1, 0, fx * (1 - fy)}, {0, 1, (1 - fx) * fy}, {1, 1, fx * fy}} {
		p := image.Pt(int(x0)+c.dx, int(y0)+c.dy)
		if !p.In(s.im.Bounds()) {
			continue
		}
		px := s.im.NRGBAAt(p.X, p.Y)
		a := float64(px.A) / 255
		if px.A < 32 {
			a = 0
		}
		out[0] += c.w * a * float64(px.R)
		out[1] += c.w * a * float64(px.G)
		out[2] += c.w * a * float64(px.B)
		out[3] += c.w * a
	}
	return out
}

// periodic samples the strip at phase s (source px, any value) of a period of
// p source px: the painting from the left of its usable span, with the last
// quarter period crossfaded into the start, so phase p meets phase 0.
func (s quayStrip) periodic(phase, p, y float64) [4]float64 {
	fade := p / 4
	x0 := s.start + (s.end-s.start-p-fade)/2
	phase = math.Mod(phase, p)
	if phase < 0 {
		phase += p
	}
	a := s.sample(x0+phase, y)
	if phase >= fade {
		return a
	}
	b := s.sample(x0+phase+p, y)
	t := phase / fade
	for i := range a {
		a[i] = b[i]*(1-t) + a[i]*t
	}
	return a
}

// quayTiles renders the quay on axis id as 2x and 4x hex tiles: the pier over
// its soft shadow, clipped to the hex, transparent elsewhere.
func quayTiles(s quayStrip, id string) (map[int]*image.NRGBA, error) {
	step, ok := quayAxes[id]
	if !ok {
		return nil, fmt.Errorf("unknown quay %s", id)
	}
	if s.end-s.start < 1.3*step.length()*s.k {
		return nil, fmt.Errorf("quay strip too short for %s", id)
	}
	u := step.mul(1 / step.length())
	n := connectionPoint{-u[1], u[0]}
	period := step.length() * s.k
	out := map[int]*image.NRGBA{}
	for _, d := range []int{2, 4} {
		// Supersample to the strip's resolution.
		sub := int(math.Ceil(s.k / float64(d)))
		im := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		for y := range 80 * d {
			for x := range 96 * d {
				if !connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}) {
					continue
				}
				var acc [4]float64
				var shadow float64
				for sy := range sub {
					for sx := range sub {
						q := connectionPoint{(float64(x) + (float64(sx)+.5)/float64(sub)) / float64(d), (float64(y) + (float64(sy)+.5)/float64(sub)) / float64(d)}.sub(connectionPoint{48, 40})
						c := s.periodic((q[0]*u[0]+q[1]*u[1])*s.k, period, s.centre+(q[0]*n[0]+q[1]*n[1])*s.k)
						for i := range acc {
							acc[i] += c[i]
						}
						r := q.sub(connectionPoint{quayShadowOffset, quayShadowOffset})
						across := math.Abs(r[0]*n[0] + r[1]*n[1])
						shadow += quayShadowAlpha * (1 - connectionSmoothStep((across-quayWidth/2+1.5)/3))
					}
				}
				m := float64(sub * sub)
				a, sa := acc[3]/m, shadow/m
				// The pier over its shadow, straight alpha.
				oa := a + sa*(1-a)
				if oa <= 0 {
					continue
				}
				var rgb [3]float64
				for i := range rgb {
					rgb[i] = (acc[i]/m + float64([3]uint8{quayShadowColor.R, quayShadowColor.G, quayShadowColor.B}[i])*sa*(1-a)) / oa
				}
				im.SetNRGBA(x, y, color.NRGBA{uint8(math.Round(rgb[0])), uint8(math.Round(rgb[1])), uint8(math.Round(rgb[2])), uint8(math.Round(oa * 255))})
			}
		}
		out[d] = im
	}
	return out, nil
}
