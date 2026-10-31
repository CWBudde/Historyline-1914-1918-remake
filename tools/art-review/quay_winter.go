package artreview

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// The winter pier (quay-pier-winter-v1.png) gets two finishes in the export,
// on the strip before it is turned to the axes, so the painting stays as
// selected:
//   - the deck between the kerbs (|across| <= quayWinterDeckHalf world px,
//     fading out over the next 0.6) is shifted by quayWinterDeckL in graded
//     L*: each pixel's L* goes to the source L* that the pack grade
//     (terraingrade, a neutral grey) takes to its graded L* plus the shift,
//     at the same chromaticity, so the deck comes to the graded winter road
//     and the ruts keep their graded depth below it;
//   - an ice collar runs along both waterlines: from the pier's outline it
//     covers the timber fender (its width per side the median distance from
//     the outline to the kerb's snow) up to a dark line of quayIceTimber..
//     quayIceTimber+quayIceTimberSpan world px left before the kerb, wandering
//     along the pier (value noise over 5 and 2 world px cells, the inner edge
//     jittered over 0.8 world px cells by up to quayIceJitter) and feathered
//     over quayIceFeather toward the kerb; outside the outline it goes on over
//     the water as a shelf quayIceShelfMin..+quayIceShelfSpan world px wide
//     (value noise over 1.6 and 0.5 world px cells), its broken edge feathered
//     over quayIceEdge, and beyond it thin ice of opacity quayIceGlaze fading
//     out over quayIceGlazeWidth. Bollard caps standing on the fender (light
//     pixels 0.35..1.2 world px inside the outline) keep it clear: the ice
//     fades in from 3 to 9 source px beside them. The ice is quayIceColor,
//     turning toward quayThinIceColor by up to quayIceThin out on the shelf,
//     lightened and darkened by up to 11 % by value noise over 1.2 and 0.45
//     world px cells, with a fifth of the timber's grain showing through.
//     The collar is 1.2..2.5 world px wide in all.
const (
	quayWinterDeckL    = -6.9 // graded L* added to the winter deck
	quayWinterDeckHalf = 11.4 // world px from the centre line covered by the deck shift

	quayIceTimber, quayIceTimberSpan  = .25, .35 // world px of dark fender left between the ice and the kerb
	quayIceJitter                     = .2       // world px of jitter on the collar's inner edge
	quayIceFeather                    = .4       // world px over which the ice thins out toward the kerb
	quayIceShelfMin, quayIceShelfSpan = .2, .8   // world px of ice on the water, outside the outline
	quayIceEdge                       = .3       // world px over which the shelf's edge is feathered
	quayIceGlaze, quayIceGlazeWidth   = .3, .5   // opacity and width (world px) of the thin ice beyond the shelf
	quayIceThin                       = .4       // share of the thin ice colour at the shelf's outer edge
	quayLight                         = 130      // mean RGB above which a pixel is snow (a bollard cap, the kerb)
)

var (
	quayIceColor     = color.NRGBA{186, 196, 203, 255}
	quayThinIceColor = color.NRGBA{150, 166, 174, 255}
)

// quayWinterFinish returns s with the winter deck shift and the ice collar
// painted into a copy of its image; the strip's rows and scale stay.
func quayWinterFinish(s quayStrip) quayStrip {
	src := s.im
	b := src.Bounds()
	im := image.NewNRGBA(b)
	copy(im.Pix, src.Pix)
	out := s
	out.im = im

	// graded is the pack grade's L* of a neutral grey of source L* l, tabled
	// at 0.1 steps; source inverts it.
	const steps = 1000
	greys := image.NewNRGBA(image.Rect(0, 0, steps+1, 1))
	for i := range steps + 1 {
		v := quaySRGB(quayLabY(float64(i) / 10))
		greys.SetNRGBA(i, 0, color.NRGBA{v, v, v, 255})
	}
	greys = terraingrade.Grade(greys, "tile/quay36_w", "terrain/quay", nil)
	var table [steps + 1]float64
	for i := range table {
		table[i] = quayL(greys.NRGBAAt(i, 0))
	}
	graded := func(l float64) float64 {
		f := math.Max(0, math.Min(steps, l*10))
		i := math.Min(steps-1, math.Floor(f))
		return table[int(i)] + (f-i)*(table[int(i)+1]-table[int(i)])
	}
	source := func(g float64) float64 {
		lo, hi := 0., 100.
		for range 40 {
			if mid := (lo + hi) / 2; graded(mid) < g {
				lo = mid
			} else {
				hi = mid
			}
		}
		return (lo + hi) / 2
	}

	for x := b.Min.X; x < b.Max.X; x++ {
		// The deck.
		for y := b.Min.Y; y < b.Max.Y; y++ {
			across := math.Abs((float64(y) + .5 - s.centre) / s.k)
			m := 1 - connectionSmoothStep((across-quayWinterDeckHalf)/.6)
			i := im.PixOffset(x, y)
			if m <= 0 || im.Pix[i+3] == 0 {
				continue
			}
			p := im.Pix[i : i+3 : i+3]
			l := quayL(color.NRGBA{p[0], p[1], p[2], 255})
			if l <= 0 {
				continue
			}
			l2 := l + m*(source(graded(l)+quayWinterDeckL)-l)
			f := quayLabY(l2) / quayLabY(l)
			for c := range p {
				p[c] = quaySRGB(quayLinear(p[c]) * f)
			}
		}
	}

	// The ice collar on both sides: edge is each column's outermost opaque
	// row, dir the step from it onto the water.
	light := func(x, y int) bool {
		p := src.NRGBAAt(x, y)
		return (int(p.R)+int(p.G)+int(p.B))/3 > quayLight
	}
	for side, dir := range []int{-1, 1} {
		edges := make([]int, b.Dx())
		near := make([]float64, b.Dx())
		var widths []float64
		for x := b.Min.X; x < b.Max.X; x++ {
			edge := int(math.Floor(s.centre))
			for y := edge; y+dir >= b.Min.Y && y+dir < b.Max.Y && src.NRGBAAt(x, y+dir).A > 128; y += dir {
				edge = y + dir
			}
			edges[x-b.Min.X] = edge
			near[x-b.Min.X] = math.Inf(1)
		}
		for x := b.Min.X; x < b.Max.X; x++ {
			edge := edges[x-b.Min.X]
			// A bollard's snow cap stands on the fender.
			fitted := false
			for t := .35; t <= 1.2 && !fitted; t += 1 / s.k {
				fitted = light(x, edge-dir*int(math.Round(t*s.k)))
			}
			if !fitted {
				// The fender reaches to the kerb's snow.
				for t := 1.; t <= 3; t += 1 / s.k {
					if light(x, edge-dir*int(math.Round(t*s.k))) {
						widths = append(widths, t)
						break
					}
				}
				continue
			}
			for d := -12; d <= 12; d++ {
				if i := x - b.Min.X + d; i >= 0 && i < len(near) {
					near[i] = math.Min(near[i], math.Abs(float64(d)))
				}
			}
		}
		fender := 1.5
		if len(widths) > 0 {
			sort.Float64s(widths)
			fender = widths[len(widths)/2]
		}
		for x := b.Min.X; x < b.Max.X; x++ {
			edge := edges[x-b.Min.X]
			open := connectionSmoothStep((near[x-b.Min.X] - 3) / 6)
			at := connectionPoint{(float64(x) + .5) / s.k, float64(100 * (side + 1))}
			n := .6*(valueNoise(at, 5)+1)/2 + .4*(valueNoise(at.add(connectionPoint{0, 50}), 2)+1)/2
			inner := fender - quayIceTimber - quayIceTimberSpan*connectionSmoothStep((n-.2)/.6) + quayIceJitter*valueNoise(at.add(connectionPoint{0, 20}), .8)
			m := .65*(valueNoise(at.add(connectionPoint{0, 30}), 1.6)+1)/2 + .35*(valueNoise(at.add(connectionPoint{0, 40}), .5)+1)/2
			shelf := quayIceShelfMin + quayIceShelfSpan*connectionSmoothStep((m-.2)/.6)
			// The outline lies between row edge and the next row onto the water.
			line := float64(edge) + .5 + .5*float64(dir)
			for y := edge - dir*int(math.Ceil((inner+1)*s.k)); ; y += dir {
				if y < b.Min.Y || y >= b.Max.Y {
					break
				}
				t := (line - (float64(y) + .5)) * float64(dir) / s.k // world px inside the outline
				if t < -shelf-quayIceGlazeWidth {
					break
				}
				q := connectionPoint{(float64(x) + .5) / s.k, (float64(y) + .5) / s.k}
				shade := 1 + .06*valueNoise(q, 1.2) + .05*valueNoise(q.add(connectionPoint{0, 7}), .45)
				i := im.PixOffset(x, y)
				p := im.Pix[i : i+4 : i+4]
				thin := quayIceThin * connectionSmoothStep(-t/math.Max(shelf, .3))
				ice := [3]float64{
					(float64(quayIceColor.R)*(1-thin) + float64(quayThinIceColor.R)*thin) * shade,
					(float64(quayIceColor.G)*(1-thin) + float64(quayThinIceColor.G)*thin) * shade,
					(float64(quayIceColor.B)*(1-thin) + float64(quayThinIceColor.B)*thin) * shade,
				}
				if t >= 0 {
					c := open * (1 - connectionSmoothStep((t-inner)/quayIceFeather+.5))
					for k := range 3 {
						grain := float64(p[k]) - 50
						v := ice[k] + .2*grain
						p[k] = uint8(math.Round(math.Max(0, math.Min(255, float64(p[k])*(1-c)+v*c))))
					}
					continue
				}
				// The shelf on the water, its broken edge feathered, and the
				// glaze of thin ice beyond it.
				a := math.Max(1-connectionSmoothStep((-t-shelf)/quayIceEdge+.5), quayIceGlaze*(1-connectionSmoothStep((-t-shelf)/quayIceGlazeWidth)))
				pa := float64(p[3]) / 255
				oa := a + pa*(1-a)
				if oa <= 0 {
					continue
				}
				for k := range 3 {
					p[k] = uint8(math.Round(math.Max(0, math.Min(255, (ice[k]*a+float64(p[k])*pa*(1-a))/oa))))
				}
				p[3] = uint8(math.Round(oa * 255))
			}
		}
	}
	return out
}

// quayLinear is the linear light of an sRGB value.
func quayLinear(v uint8) float64 {
	s := float64(v) / 255
	if s <= .04045 {
		return s / 12.92
	}
	return math.Pow((s+.055)/1.055, 2.4)
}

// quaySRGB is the sRGB value of linear light v.
func quaySRGB(v float64) uint8 {
	v = math.Max(0, math.Min(1, v))
	if v <= .0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - .055
	}
	return uint8(math.Round(v * 255))
}

// quayL is the CIE L* of c.
func quayL(c color.NRGBA) float64 {
	y := .2126*quayLinear(c.R) + .7152*quayLinear(c.G) + .0722*quayLinear(c.B)
	if y > 216./24389 {
		return 116*math.Cbrt(y) - 16
	}
	return y * 24389 / 27
}

// quayLabY is the relative luminance of L* l.
func quayLabY(l float64) float64 {
	if l > 8 {
		f := (l + 16) / 116
		return f * f * f
	}
	return l * 27 / 24389
}
