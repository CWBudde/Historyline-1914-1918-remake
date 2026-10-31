// Package terrainlight is the map's one light (art/direction.md, "One light"):
// it comes from the upper left at azimuth 315°, and every standing item casts
// a shadow of RGB 18,26,16 that is 0.25 × the item's diameter long, at an
// opacity by kind: crowns 0.6, rocks 0.45, buildings 0.35. Paintings carry
// only form shading; the cast shadows come from code, here for the painted
// shadows of the standing items and in internal/hlkit for the scattered ones.
package terrainlight

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"slices"
	"strings"
)

// Version names the light, so a pack built with an older one can be told
// apart.
const Version = "terrainlight-v1"

// Azimuth is the direction the light comes from, in degrees clockwise from
// north; shadows fall the opposite way.
const Azimuth = 315.0

// Length is a shadow's length as a fraction of its item's diameter.
const Length = 0.25

// Shadow opacities by kind of item.
const (
	Crown    = 0.6
	Rock     = 0.45
	Building = 0.35
)

// Colour is the shadow colour.
var Colour = color.NRGBA{18, 26, 16, 255}

// softness is the σ of the shadow's soft edge, in world px.
const softness = 0.5

// opacities are the art families whose standing items cast a shadow, by
// their kind.
var opacities = map[string]float64{"terrain/forest": Crown}

// Recast reports whether the asset id of art family art is a painted cast
// shadow to be drawn by Cast instead, and of which standing item at which
// opacity. The shadows of scattered items (standing/shadow/scatter/…) are
// made by internal/hlkit with this model and kept.
func Recast(id, art string) (standing string, opacity float64, ok bool) {
	rest, shadow := strings.CutPrefix(id, "standing/shadow/")
	if !shadow || strings.HasPrefix(rest, "scatter/") {
		return "", 0, false
	}
	if opacity, ok = Opacity(art); !ok {
		return "", 0, false
	}
	return "standing/" + rest, opacity, true
}

// Opacity is the shadow opacity of the standing items of art family art, if
// they cast one.
func Opacity(art string) (float64, bool) {
	o, ok := opacities[art]
	return o, ok
}

// Diameter is the median equivalent diameter of the opaque blobs of im (alpha
// above half), in world px at density px per world px; 0 without any. Blobs
// smaller than a world px square are left out. It is measured at 2 px per world
// px, the alpha box-reduced from higher densities, so every density of an item
// gives it the same shadow.
func Diameter(im image.Image, density int) float64 {
	a := alphaOf(im)
	if k := density / measureDensity; k > 1 {
		a = reduce(a, k)
	}
	density = min(density, measureDensity)
	w, h := a.Rect.Dx(), a.Rect.Dy()
	seen := make([]bool, w*h)
	var sizes []float64
	for i := range seen {
		if seen[i] || a.Pix[i] <= 128 {
			continue
		}
		n, stack := 0, []int{i}
		seen[i] = true
		for len(stack) > 0 {
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			n++
			x, y := j%w, j/w
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if X, Y := x+dx, y+dy; X >= 0 && Y >= 0 && X < w && Y < h && !seen[Y*w+X] && a.Pix[Y*w+X] > 128 {
						seen[Y*w+X] = true
						stack = append(stack, Y*w+X)
					}
				}
			}
		}
		if n >= density*density {
			sizes = append(sizes, 2*math.Sqrt(float64(n)/math.Pi)/float64(density))
		}
	}
	if len(sizes) == 0 {
		return 0
	}
	slices.Sort(sizes)
	return sizes[len(sizes)/2]
}

// Cast is the cast shadow of item im on the same canvas: its alpha moved
// Length × diameter world px away from the light, with a soft edge, in Colour
// at opacity. density is im's px per world px.
func Cast(im image.Image, density int, diameter, opacity float64) *image.NRGBA {
	a := alphaOf(im)
	w, h := a.Rect.Dx(), a.Rect.Dy()
	// The light comes from Azimuth; the shadow falls along Azimuth−180°.
	rad := (Azimuth - 180) * math.Pi / 180
	d := Length * diameter * float64(density)
	dx, dy := d*math.Sin(rad), -d*math.Cos(rad)
	moved := make([]float64, w*h)
	for y := range h {
		for x := range w {
			moved[y*w+x] = bilinear(a, float64(x)-dx, float64(y)-dy)
		}
	}
	soft := blur(moved, w, h, softness*float64(density))
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i, v := range soft {
		if v := math.Round(v * opacity); v > 0 {
			out.Pix[4*i], out.Pix[4*i+1], out.Pix[4*i+2], out.Pix[4*i+3] = Colour.R, Colour.G, Colour.B, uint8(min(v, 255))
		}
	}
	return out
}

// measureDensity is the px per world px Diameter measures at.
const measureDensity = 2

// reduce is the k×k box mean of a.
func reduce(a *image.Alpha, k int) *image.Alpha {
	out := image.NewAlpha(image.Rect(0, 0, a.Rect.Dx()/k, a.Rect.Dy()/k))
	for y := range out.Rect.Dy() {
		for x := range out.Rect.Dx() {
			s := 0
			for dy := range k {
				for dx := range k {
					s += int(a.Pix[(k*y+dy)*a.Stride+k*x+dx])
				}
			}
			out.Pix[y*out.Stride+x] = uint8((s + k*k/2) / (k * k))
		}
	}
	return out
}

func alphaOf(im image.Image) *image.Alpha {
	b := im.Bounds()
	a := image.NewAlpha(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(a, a.Rect, im, b.Min, draw.Src)
	return a
}

// bilinear is a's alpha at (x, y) in px, 0 outside.
func bilinear(a *image.Alpha, x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := x-x0, y-y0
	at := func(x, y int) float64 {
		if x < 0 || y < 0 || x >= a.Rect.Dx() || y >= a.Rect.Dy() {
			return 0
		}
		return float64(a.Pix[y*a.Stride+x])
	}
	ix, iy := int(x0), int(y0)
	return (at(ix, iy)*(1-fx)+at(ix+1, iy)*fx)*(1-fy) + (at(ix, iy+1)*(1-fx)+at(ix+1, iy+1)*fx)*fy
}

// blur is a separable Gaussian of σ px over the w×h plane v, keeping its mass
// inside the canvas.
func blur(v []float64, w, h int, sigma float64) []float64 {
	r := int(math.Ceil(3 * sigma))
	k := make([]float64, 2*r+1)
	sum := 0.0
	for i := range k {
		k[i] = math.Exp(-float64((i-r)*(i-r)) / (2 * sigma * sigma))
		sum += k[i]
	}
	for i := range k {
		k[i] /= sum
	}
	pass := func(src []float64, step, n, lines, stride int) []float64 {
		out := make([]float64, len(src))
		for l := range lines {
			for i := range n {
				s := 0.0
				for j, kv := range k {
					if p := i + j - r; p >= 0 && p < n {
						s += kv * src[l*stride+p*step]
					}
				}
				out[l*stride+i*step] = s
			}
		}
		return out
	}
	return pass(pass(v, 1, w, h, w), w, h, w, 1)
}
