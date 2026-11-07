package dosscreen

import (
	"image"
	"image/color"
	"math"
	"sync"
)

// Shading of the steel texture under the frame's gunmetal.
const (
	steelBase = 0.46 // the texture's mean ramp position
	steelTile = 128  // the texture repeats every steelTile device pixels
)

// lightX, lightY is the direction towards the light on the screen: the top left.
var lightX, lightY = -math.Sqrt2 / 2, -math.Sqrt2 / 2

var (
	steelOnce sync.Once
	steelTex  []float64
)

// steelAt is the steel texture at screen pixel (x, y) as a ramp position:
// plates of soft mottling with fine horizontal brushing, repeating every
// steelTile pixels in both directions.
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

// blend mixes the opaque colour c over the opaque pixel (x, y) by f.
func blend(img *image.RGBA, x, y int, c color.RGBA, f float64) {
	img.SetRGBA(x, y, mixRGBA(img.RGBAAt(x, y), c, f))
}

// mixRGBA is opaque a mixed towards opaque b by f.
func mixRGBA(a, b color.RGBA, f float64) color.RGBA {
	f = math.Max(0, math.Min(1, f))
	mix := func(p, q uint8) uint8 { return uint8(math.Round(float64(p)*(1-f) + float64(q)*f)) }
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}
