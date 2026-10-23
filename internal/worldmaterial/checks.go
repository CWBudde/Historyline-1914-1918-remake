package worldmaterial

import (
	"fmt"
	"image"
	"math"
)

// Meadow targets of art/direction.md (grass row and the G1 checks), at 1×.
var (
	meadowRGB      = [3]float64{100, 107, 40}
	latticeOffsets = []image.Point{{71, 40}, {142, 0}, {0, 80}, {0, 160}, {284, 0}}
)

const (
	meadowRGBTolerance = 8
	sobelEdge          = 6    // normalised Sobel L* gradient that counts as an edge, as in tools/reference-design
	meadowSigma        = 11.5 // σL* the suggested gain aims for, the middle of 10–13
	meadowBandCap      = 4.0  // RMS of the 3/10 band
	darkDepth          = 8    // L* below the radius-10 box mean that counts as dark
	bandWindow         = 50   // world px of the windows for the band CV
)

// Check is one measured rule of a material at 1×.
type Check struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Want  string  `json:"want"`
	Pass  bool    `json:"pass"`
}

func round4(v float64) float64 { return math.Round(v*1e4) / 1e4 }

func within(name string, v, lo, hi float64) Check {
	return Check{name, round4(v), fmt.Sprintf("%g to %g", lo, hi), v >= lo && v <= hi}
}

func below(name string, v, limit float64) Check {
	return Check{name, round4(v), fmt.Sprintf("< %g", limit), v < limit}
}

func atMost(name string, v, limit float64) Check {
	return Check{name, round4(v), fmt.Sprintf("<= %g", limit), v <= limit}
}

// Passed reports whether every check passed.
func Passed(cs []Check) bool {
	for _, c := range cs {
		if !c.Pass {
			return false
		}
	}
	return true
}

func meanRGB(im *image.NRGBA) [3]float64 {
	r := im.Bounds()
	var s [3]float64
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := im.NRGBAAt(x, y)
			s[0] += float64(c.R)
			s[1] += float64(c.G)
			s[2] += float64(c.B)
		}
	}
	n := float64(r.Dx() * r.Dy())
	return [3]float64{s[0] / n, s[1] / n, s[2] / n}
}

func rgbChecks(im *image.NRGBA) []Check {
	m := meanRGB(im)
	cs := make([]Check, 3)
	for i, name := range []string{"mean R", "mean G", "mean B"} {
		cs[i] = within(name, m[i], meadowRGB[i]-meadowRGBTolerance, meadowRGB[i]+meadowRGBTolerance)
	}
	return cs
}

func stats(v []float64) (mean, sigma float64) {
	for _, x := range v {
		mean += x
	}
	mean /= float64(len(v))
	for _, x := range v {
		sigma += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(sigma / float64(len(v)))
}

// band is the difference of the box means at the two radii, the band field of
// tools/reference-design, wrapped over the period.
func band(p *plane, small, large int) []float64 {
	a, b := boxMean(p.l, p.w, p.h, small), boxMean(p.l, p.w, p.h, large)
	for i := range a {
		a[i] -= b[i]
	}
	return a
}

func rms(v []float64) float64 {
	s := 0.
	for _, x := range v {
		s += x * x
	}
	return math.Sqrt(s / float64(len(v)))
}

func edgeDensity(p *plane) float64 {
	at := func(x, y int) float64 { return p.l[mod(y, p.h)*p.w+mod(x, p.w)] }
	edges := 0
	for y := range p.h {
		for x := range p.w {
			gx := (at(x+1, y-1) + 2*at(x+1, y) + at(x+1, y+1) - at(x-1, y-1) - 2*at(x-1, y) - at(x-1, y+1)) / 8
			gy := (at(x-1, y+1) + 2*at(x, y+1) + at(x+1, y+1) - at(x-1, y-1) - 2*at(x, y-1) - at(x+1, y-1)) / 8
			if math.Hypot(gx, gy) > sobelEdge {
				edges++
			}
		}
	}
	return float64(edges) / float64(p.w*p.h)
}

// pearson correlates L* with itself shifted by off, wrapped over the period.
func pearson(p *plane, off image.Point) float64 {
	mean, sigma := stats(p.l)
	if sigma < 1e-9 {
		return 0 // a flat field has nothing to repeat
	}
	s := 0.
	for y := range p.h {
		for x := range p.w {
			s += (p.l[y*p.w+x] - mean) * (p.l[mod(y+off.Y, p.h)*p.w+mod(x+off.X, p.w)] - mean)
		}
	}
	return s / float64(p.w*p.h) / (sigma * sigma)
}

// darkWidth is the widest side of the bounding boxes of four-connected pixels
// at least darkDepth below their radius-10 box mean.
func darkWidth(p *plane) int {
	local := boxMean(p.l, p.w, p.h, 10)
	dark := make([]bool, len(p.l))
	for i := range dark {
		dark[i] = p.l[i] <= local[i]-darkDepth
	}
	widest := 0
	stack := []int{}
	for start := range dark {
		if !dark[start] {
			continue
		}
		dark[start] = false
		stack = append(stack[:0], start)
		x0, y0, x1, y1 := p.w, p.h, -1, -1
		for len(stack) > 0 {
			i := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := i%p.w, i/p.w
			x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
			for _, n := range [4][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
				if n[0] >= 0 && n[1] >= 0 && n[0] < p.w && n[1] < p.h && dark[n[1]*p.w+n[0]] {
					dark[n[1]*p.w+n[0]] = false
					stack = append(stack, n[1]*p.w+n[0])
				}
			}
		}
		widest = max(widest, x1-x0+1, y1-y0+1)
	}
	return widest
}

// bandCV is the coefficient of variation of the 3/10 band RMS over whole
// bandWindow squares.
func bandCV(p *plane) float64 {
	b := band(p, 3, 10)
	var values []float64
	for wy := 0; wy+bandWindow <= p.h; wy += bandWindow {
		for wx := 0; wx+bandWindow <= p.w; wx += bandWindow {
			s := 0.
			for y := wy; y < wy+bandWindow; y++ {
				for x := wx; x < wx+bandWindow; x++ {
					s += b[y*p.w+x] * b[y*p.w+x]
				}
			}
			values = append(values, math.Sqrt(s/(bandWindow*bandWindow)))
		}
	}
	if len(values) == 0 {
		return math.Inf(1)
	}
	mean, sigma := stats(values)
	if mean == 0 {
		return math.Inf(1)
	}
	return sigma / mean
}

// ToneChecks measures a tone layer (G1a) exported at 1×.
func ToneChecks(tone *image.NRGBA) []Check {
	p := newPlane(tone)
	_, sigma := stats(boxMean(p.l, p.w, p.h, 10))
	return append(rgbChecks(tone), within("sigma L* of the radius-10 box mean", sigma, 8, 12))
}

// DetailChecks measures a detail layer (G1b) wrapped over the period at 1×.
// Its amplitude does not matter, because the gain sets it.
func DetailChecks(detail *image.NRGBA) []Check {
	p := newPlane(detail)
	ratio := math.Inf(1)
	if r := rms(band(p, 3, 10)); r > 0 {
		ratio = rms(band(p, 1, 3)) / r
	}
	return []Check{
		within("band ratio RMS(1/3) / RMS(3/10)", ratio, 0.85, 1.25),
		atMost("widest dark component, world px", float64(darkWidth(p)), 12),
		atMost("CV of the 3/10 band RMS over 50 px windows", bandCV(p), 0.25),
	}
}

// ResultChecks measures a meadow material at 1× against the grass row, the
// lattice and, for repeats > 1, the repeat of its detail.
func ResultChecks(result *image.NRGBA, period image.Point, repeats int) []Check {
	p := newPlane(result)
	mean, sigma := stats(p.l)
	cs := []Check{within("mean L*", mean, 42, 45), within("sigma L*", sigma, 10, 13)}
	cs = append(cs, rgbChecks(result)...)
	cs = append(cs,
		within("Sobel edge density", edgeDensity(p), 0.28, 0.40),
		atMost("RMS of the 3/10 band", rms(band(p, 3, 10)), meadowBandCap))
	for _, off := range latticeOffsets {
		cs = append(cs, below(fmt.Sprintf("autocorrelation at (%d,%d)", off.X, off.Y), pearson(p, off), 0.15))
	}
	if repeats > 1 {
		for _, off := range []image.Point{{int(math.Round(float64(period.X) / float64(repeats))), 0}, {0, int(math.Round(float64(period.Y) / float64(repeats)))}} {
			cs = append(cs, below(fmt.Sprintf("autocorrelation at the detail repeat (%d,%d)", off.X, off.Y), pearson(p, off), 0.3))
		}
	}
	return cs
}

// SuggestGain returns the detail gain, at 1×, that brings the combined σL* to
// meadowSigma without the 3/10 band exceeding meadowBandCap, assuming tone and
// detail are uncorrelated. It is 0 when the tone alone reaches either.
func SuggestGain(tone, detail *image.NRGBA) float64 {
	t, d := newPlane(tone), newPlane(detail)
	hp := &plane{w: d.w, h: d.h, l: highPass(d.l, d.w, d.h, HighPassRadius)}
	_, sigmaT := stats(t.l)
	_, sigmaD := stats(hp.l)
	bandT, bandD := rms(band(t, 3, 10)), rms(band(hp, 3, 10))
	if sigmaD == 0 || bandD == 0 {
		return 0
	}
	return math.Min(math.Sqrt(math.Max(0, meadowSigma*meadowSigma-sigmaT*sigmaT))/sigmaD,
		math.Sqrt(math.Max(0, meadowBandCap*meadowBandCap-bandT*bandT))/bandD)
}
