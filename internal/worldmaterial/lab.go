package worldmaterial

import (
	"image"
	"image/color"
	"math"
)

// D65 white of the sRGB primaries, as in tools/reference-design.
const whiteX, whiteZ = 0.95047, 1.08883

func srgbLinear(v uint8) float64 {
	c := float64(v) / 255
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func srgbEncode(v float64) uint8 {
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	return uint8(math.Round(255 * math.Max(0, math.Min(1, v))))
}

func labF(t float64) float64 {
	if t > 216.0/24389 {
		return math.Cbrt(t)
	}
	return (24389.0/27*t + 16) / 116
}

func labInverse(f float64) float64 {
	if f*f*f > 216.0/24389 {
		return f * f * f
	}
	return (116*f - 16) * 27 / 24389
}

// toLab converts an sRGB colour to CIELAB (D65).
func toLab(c color.NRGBA) (float64, float64, float64) {
	r, g, b := srgbLinear(c.R), srgbLinear(c.G), srgbLinear(c.B)
	x := (0.4124564*r + 0.3575761*g + 0.1804375*b) / whiteX
	y := 0.2126729*r + 0.7151522*g + 0.0721750*b
	z := (0.0193339*r + 0.1191920*g + 0.9503041*b) / whiteZ
	fx, fy, fz := labF(x), labF(y), labF(z)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

// fromLab converts CIELAB (D65) to an opaque sRGB colour, clipping out-of-gamut channels.
func fromLab(l, a, b float64) color.NRGBA {
	fy := (l + 16) / 116
	x := labInverse(fy+a/500) * whiteX
	y := labInverse(fy)
	z := labInverse(fy-b/200) * whiteZ
	return color.NRGBA{
		srgbEncode(3.2404542*x - 1.5371385*y - 0.4985314*z),
		srgbEncode(-0.9692660*x + 1.8760108*y + 0.0415560*z),
		srgbEncode(0.0556434*x - 0.2040259*y + 1.0572252*z),
		255,
	}
}

// plane is an opaque image in CIELAB, row-major.
type plane struct {
	w, h    int
	l, a, b []float64
}

func newPlane(im *image.NRGBA) *plane {
	r := im.Bounds()
	p := &plane{w: r.Dx(), h: r.Dy()}
	n := p.w * p.h
	p.l, p.a, p.b = make([]float64, n), make([]float64, n), make([]float64, n)
	for y := range p.h {
		for x := range p.w {
			i := y*p.w + x
			p.l[i], p.a[i], p.b[i] = toLab(im.NRGBAAt(r.Min.X+x, r.Min.Y+y))
		}
	}
	return p
}

// boxMean returns the periodic mean of v over squares of the given radius.
func boxMean(v []float64, w, h, radius int) []float64 {
	tmp, out := make([]float64, len(v)), make([]float64, len(v))
	n := float64(2*radius + 1)
	for y := range h {
		row := v[y*w : (y+1)*w]
		sum := 0.
		for dx := -radius; dx <= radius; dx++ {
			sum += row[mod(dx, w)]
		}
		for x := range w {
			tmp[y*w+x] = sum / n
			sum += row[mod(x+radius+1, w)] - row[mod(x-radius, w)]
		}
	}
	for x := range w {
		sum := 0.
		for dy := -radius; dy <= radius; dy++ {
			sum += tmp[mod(dy, h)*w+x]
		}
		for y := range h {
			out[y*w+x] = sum / n
			sum += tmp[mod(y+radius+1, h)*w+x] - tmp[mod(y-radius, h)*w+x]
		}
	}
	return out
}
