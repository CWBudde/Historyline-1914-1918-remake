package waterdepth

import (
	"image"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// Surf is foam along a coast by the sea: the surf tile (SurfTile) holds four
// strips of 20 world px, each from the coast side outwards, 96 world px along
// the coast. A water pixel of a tile within surfReach world px of its land
// takes the strip at its distance from the coast, read along the coast in
// world coordinates and passing from strip to strip slowly over the map, as
// strong as the sea is near: none by a lake, all of it where the depth within
// exposureReach world px reaches exposureFull.
const (
	surfReach      = 20.0
	surfLength     = 96.0
	exposureReach  = 80.0
	exposureFrom   = 0.05
	exposureFull   = 0.35
	surfStripPitch = 20.0
	surfStrength   = 0.85
)

// landOf is the land of tile im, true where a pixel is opaque and not open
// water (neither the open water tile's pixel, water, nor wet by the grade's
// test), or nil where the tile has no land.
func landOf(im, water *image.NRGBA) []bool {
	b := im.Rect
	land := make([]bool, b.Dx()*b.Dy())
	some := false
	for y := range b.Dy() {
		for x := range b.Dx() {
			i := im.PixOffset(b.Min.X+x, b.Min.Y+y)
			p := im.Pix[i : i+4 : i+4]
			if p[3] < 128 {
				continue
			}
			if water != nil {
				w := water.PixOffset(water.Rect.Min.X+x, water.Rect.Min.Y+y)
				if [4]uint8(p) == [4]uint8(water.Pix[w:w+4]) {
					continue
				}
			}
			if terraingrade.Wet(p[0], p[1], p[2]) < 0.5 {
				land[y*b.Dx()+x] = true
				some = true
			}
		}
	}
	if !some {
		return nil
	}
	return land
}

// surf lays the foam of surf over the water of im, a tile whose top left lies
// at origin in pixels of density, land its land (landOf).
func (f *Field) surf(im *image.NRGBA, land []bool, surf *image.NRGBA, origin image.Point, density int) {
	w, h := im.Rect.Dx(), im.Rect.Dy()
	d := float64(density)
	// Chamfer distance from land in thirds of a pixel.
	const far = math.MaxInt32 / 2
	dist := make([]int, w*h)
	for i := range dist {
		if !land[i] {
			dist[i] = far
		}
	}
	at := func(x, y int) int {
		if x < 0 || y < 0 || x >= w || y >= h {
			return far
		}
		return dist[y*w+x]
	}
	for y := range h {
		for x := range w {
			i := y*w + x
			dist[i] = min(dist[i], at(x-1, y)+3, at(x, y-1)+3, at(x-1, y-1)+4, at(x+1, y-1)+4)
		}
	}
	for y := h - 1; y >= 0; y-- {
		for x := w - 1; x >= 0; x-- {
			i := y*w + x
			dist[i] = min(dist[i], at(x+1, y)+3, at(x, y+1)+3, at(x+1, y+1)+4, at(x-1, y+1)+4)
		}
	}
	reach := int(surfReach * d * 3)
	exposure := map[image.Point]float64{}
	for y := range h {
		for x := range w {
			i := y*w + x
			p := im.Pix[im.PixOffset(im.Rect.Min.X+x, im.Rect.Min.Y+y):]
			if land[i] || dist[i] >= reach || p[3] == 0 {
				continue
			}
			wx, wy := (float64(origin.X+x)+0.5)/d, (float64(origin.Y+y)+0.5)/d
			// The exposure changes slowly: once per 4 world px.
			cell := image.Pt(int(math.Floor(wx/4)), int(math.Floor(wy/4)))
			e, ok := exposure[cell]
			if !ok {
				e = f.exposure(float64(cell.X)*4+2, float64(cell.Y)*4+2)
				exposure[cell] = e
			}
			if e <= 0 {
				continue
			}
			// Along the coast: the distance's gradient points away from land.
			gx, gy := float64(at(x+1, y)-at(x-1, y)), float64(at(x, y+1)-at(x, y-1))
			if math.Abs(gx) > far/2 || math.Abs(gy) > far/2 {
				gx, gy = 0, 0
			}
			ax, ay := math.Abs(gy), math.Abs(gx) // the tangent (-gy, gx), unsigned
			if ax+ay == 0 {
				ax = 1
			}
			v := float64(dist[i]) / 3 / d
			s := wx/337 + wy/211
			k := math.Floor(s)
			t := s - k
			t = t * t * (3 - 2*t)
			var c [4]float64
			for _, row := range []struct {
				strip  int
				weight float64
			}{{int(k), 1 - t}, {int(k) + 1, t}} {
				for _, along := range []struct {
					u, weight float64
				}{{wx, ax / (ax + ay)}, {wy, ay / (ax + ay)}} {
					q := surfAt(surf, along.u, (float64(((row.strip%4)+4)%4))*surfStripPitch+v, d)
					share := row.weight * along.weight
					a := float64(q[3]) / 255 * share
					for ch := range 3 {
						c[ch] += float64(q[ch]) * a
					}
					c[3] += a
				}
			}
			a := c[3] * e * surfStrength
			if a <= 0 {
				continue
			}
			for ch := range 3 {
				p[ch] = uint8(math.Round(float64(p[ch])*(1-a) + c[ch]/c[3]*a))
			}
		}
	}
}

// surfAt is the pixel of the surf tile at u world px along the coast and v
// world px into the strips, at density d.
func surfAt(surf *image.NRGBA, u, v, d float64) [4]uint8 {
	w, h := surf.Rect.Dx(), surf.Rect.Dy()
	x := int(math.Floor(math.Mod(math.Mod(u, surfLength)+surfLength, surfLength) * d))
	y := int(math.Floor(v * d))
	if x >= w {
		x = w - 1
	}
	if y < 0 || y >= h {
		return [4]uint8{}
	}
	i := surf.PixOffset(surf.Rect.Min.X+x, surf.Rect.Min.Y+y)
	return [4]uint8(surf.Pix[i : i+4])
}

// exposure is how open to the sea world point (x, y) is, 0 by a lake to 1: the
// greatest depth within exposureReach world px, from exposureFrom to
// exposureFull.
func (f *Field) exposure(x, y float64) float64 {
	m := f.At(x, y)
	for _, r := range []float64{exposureReach / 2, exposureReach} {
		for k := range 8 {
			a := float64(k) * math.Pi / 4
			m = math.Max(m, f.At(x+r*math.Cos(a), y+r*math.Sin(a)))
		}
	}
	t := math.Max(0, math.Min(1, (m-exposureFrom)/(exposureFull-exposureFrom)))
	return t * t * (3 - 2*t)
}
