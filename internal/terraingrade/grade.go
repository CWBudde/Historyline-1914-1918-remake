// Package terraingrade grades the terrain art to the value targets of
// art/direction.md, as the pack build and the map renderer load it.
//
// The grade works on each pixel's colour, so a colour comes out the same in every
// tile and the tiles keep meeting without a step:
//   - land follows one L* curve that stretches the map's 5–95 % range to about
//     8–64 and lifts the meadow to 42–45, and loses 4.5 b* where it is yellow;
//   - open water (blue pixels) follows its own curve to the deep-water L* 31–36
//     and keeps its b*; in a winter tile (ID suffix _w) it follows a darker
//     curve to L* 24–32 and turns to a cold slate teal of chroma about 8;
//     the renderers darken it further by the sea's depth (Deepen);
//   - snow and highlights above L* 85 stay as painted, so winter art is kept.
//
// A few families grade on their own where they differ from the meadow baked
// into their tiles (the snow meadow in winter tiles, so their snow is kept):
// relief is darker than grass, and buildings come to b* 16 instead of losing
// the land's b*. Crowns follow their own curve, so the forest
// stays at L* 26–31 with dark gaps. Winter forest (a forest image that is mostly
// grey: snow and bare branches) follows a darker curve instead, so it sits 35
// or more under the snow. Cast shadows, fog and units are left alone.
// The source art in assets/ is unchanged; Version names the grade.
package terraingrade

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"runtime"
	"strings"
	"sync"
)

// Version names the grade, so a pack built with an older one can be told apart.
const Version = "terraingrade-v4"

// curve maps L* through knots of increasing input; between knots it is linear.
type curve [][2]float64

func (c curve) at(v float64) float64 {
	for i := 1; i < len(c); i++ {
		if v <= c[i][0] {
			a, b := c[i-1], c[i]
			return a[1] + (v-a[0])*(b[1]-a[1])/(b[0]-a[0])
		}
	}
	return c[len(c)-1][1]
}

var (
	// landL takes the PULSE render's 5 %, meadow mean and 95 % L* (15, 41, 53)
	// to 5, 43.5 and 66, and returns to identity at 85.
	landL = curve{{0, 0}, {15, 5}, {41, 43.5}, {53, 66}, {85, 85}, {100, 100}}
	// crownL takes the crowns' shade, mean and 95 % L* (12, 25.7, 46) to 4,
	// 28.5 and 56: the forest at 26–31, its gaps dark, its lit tops at most 58.
	crownL = curve{{0, 0}, {12, 4}, {25.7, 28.5}, {46, 56}, {80, 80}, {100, 100}}
	// waterL takes open water's 5 %, mean and 95 % L* (32, 40.7, 57.3) to 25,
	// 33.5 and 50.
	waterL = curve{{0, 0}, {32, 25}, {40.7, 33.5}, {57.3, 50}, {75, 75}, {100, 100}}
	// winterForestL takes the winter forest's trees (mean L* 44 under half a
	// tile of snow) to about 24, so a winter forest cell sits 35 under the snow,
	// and leaves the snow as painted.
	winterForestL = curve{{0, 0}, {30, 15}, {45, 25}, {60, 38}, {75, 65}, {85, 85}, {100, 100}}
	// winterWaterL takes open water's 5 %, mean and 95 % L* (32, 40.7, 57.3)
	// to 21, 28 and 40: the direction's cold winter water at L* 24–32.
	winterWaterL = curve{{0, 0}, {32, 21}, {40.7, 28}, {57.3, 40}, {75, 75}, {100, 100}}
)

// The winter water colour (icy): the ripples keep 0.3 of their colour around
// a*, b* of −3, +2.8, so the painted blue (a* −11.8, b* −24.4) comes to a slate
// teal of chroma about 8, near RGB 40,62,66.
const (
	winterChroma     = 0.3
	winterA, winterB = -3.0, 2.8
)

// winterShare is the share of grey in a forest image (opaque pixels of chroma
// 12 or less: snow and bare branches) from which it is winter forest: the
// summer forest art has at most 16 %, the winter art 85 % or more.
const winterShare = 0.5

// landB is the b* added to yellow land.
const landB = -4.5

// row is a family's own grade, weighted by how far a pixel differs from the
// meadow baked into its tile (fully where there is no meadow to compare): an L*
// offset that fades out from L* 70 to 85, so snow stays as painted, and for
// buildings the b* offset that replaces the land's.
type row struct {
	l, b     float64
	building bool
}

// rows are the families with a grade of their own, by art key. Relief is darker
// than grass. The summer buildings are painted at b* 17.1 (hq) to 19.8 (factory)
// where they differ from the meadow; each is brought to the direction's 16. The
// house variants (tile/house2 …) take the hut's offset: their gardens and yards
// would turn grey at 16.
var rows = map[string]row{
	"terrain/hills":    {l: -3.5},
	"terrain/mount":    {l: -12},
	"terrain/mountain": {l: -9},
	"terrain/house":    {b: -1.0, building: true},
	"terrain/hq":       {b: -0.3, building: true},
	"terrain/cmp":      {b: -3.3, building: true},
	"terrain/factory":  {b: -3.5, building: true},
}

// Meadow is the asset id of the meadow tile baked into the tile id: the snow
// meadow for a winter tile, so its snow counts as meadow and keeps the land's
// grade instead of its family's.
func Meadow(id string) string {
	if strings.HasSuffix(id, "_w") {
		return "tile/grass_w"
	}
	return "tile/grass"
}

// Applies reports whether the asset id of art family art is graded.
func Applies(id, art string) bool {
	if !strings.HasPrefix(art, "terrain/") || art == "terrain/fog" {
		return false
	}
	return !strings.HasPrefix(id, "standing/shadow/") && id != "tile/unseen"
}

// Grade returns im graded as the asset id of art family art, as a new image.
// meadow is the tile Meadow(id) names at the same density when im is a hex tile
// of that size, else nil.
func Grade(im image.Image, id, art string, meadow image.Image) *image.NRGBA {
	b := im.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Rect, im, b.Min, draw.Src)
	if !Applies(id, art) {
		return out
	}
	r := rows[art]
	land := landL
	if strings.HasPrefix(id, "standing/scatter/crown/") || strings.HasPrefix(id, "standing/context/forest/") || id == "standing/forest" {
		land = crownL
	}
	if art == "terrain/forest" && grey(out) >= winterShare {
		land = winterForestL
	}
	water, winter := waterL, strings.HasSuffix(id, "_w")
	if winter {
		water = winterWaterL
	}
	var m *image.NRGBA
	if meadow != nil && meadow.Bounds().Size() == b.Size() && r != (row{}) {
		mb := meadow.Bounds()
		m = image.NewNRGBA(image.Rect(0, 0, mb.Dx(), mb.Dy()))
		draw.Draw(m, m.Rect, meadow, mb.Min, draw.Src)
	}
	rowsOf(out.Rect.Dy(), func(y int) {
		for x := range out.Rect.Dx() {
			i := out.PixOffset(x, y)
			p := out.Pix[i : i+4 : i+4]
			if p[3] == 0 {
				continue
			}
			l0, a, bb := toLab(p[0], p[1], p[2])
			l, db := base(land, water, l0, bb)
			if winter {
				a, bb = icy(a, bb, wet(l0, bb))
			}
			if r != (row{}) {
				w := 1.0
				if m != nil {
					w = smooth(diff(out, i, m.Pix[i:i+4:i+4]), 2, 8)
				}
				l += w * r.l * clamp((85-l0)/15)
				if r.building {
					db = (1-w)*db + w*r.b
				}
			}
			c := fromLab(l, a, bb+db, p[3])
			p[0], p[1], p[2] = c.R, c.G, c.B
		}
	})
	return out
}

// wet is how much a pixel of L* l and b* b is open water: blue and not light.
func wet(l, b float64) float64 { return clamp((-b-2)/8) * clamp((70-l)/15) }

// base is the grade of one pixel's L* and the b* it adds: land follows the
// curve land, open water the curve water.
func base(land, water curve, l, b float64) (float64, float64) {
	w := wet(l, b)
	yellow := clamp(b / 8)
	return land.at(l)*(1-w) + water.at(l)*w, landB * yellow * (1 - w)
}

// icy is a's and b's winter water colour where the pixel is open water to the
// share w: the ripples keep winterChroma of their colour around the cold slate
// teal winterA, winterB.
func icy(a, b, w float64) (float64, float64) {
	k := 1 - w*(1-winterChroma)
	return a*k + w*winterA, b*k + w*winterB
}

// grey is the share of im's opaque pixels (alpha at least half) of chroma 12
// or less.
func grey(im *image.NRGBA) float64 {
	var n, all int
	for i := 0; i < len(im.Pix); i += 4 {
		if im.Pix[i+3] < 128 {
			continue
		}
		all++
		if _, a, b := toLab(im.Pix[i], im.Pix[i+1], im.Pix[i+2]); math.Hypot(a, b) <= 12 {
			n++
		}
	}
	if all == 0 {
		return 0
	}
	return float64(n) / float64(all)
}

// diff is the CIE76 distance between pixel i of im (still ungraded) and the
// meadow pixel q.
func diff(im *image.NRGBA, i int, q []uint8) float64 {
	p := im.Pix[i : i+3 : i+3]
	if p[0] == q[0] && p[1] == q[1] && p[2] == q[2] {
		return 0
	}
	l0, a0, b0 := toLab(p[0], p[1], p[2])
	l1, a1, b1 := toLab(q[0], q[1], q[2])
	return math.Sqrt((l0-l1)*(l0-l1) + (a0-a1)*(a0-a1) + (b0-b1)*(b0-b1))
}

// rowsOf calls f for every row 0..n-1, spread over the processors.
func rowsOf(n int, f func(y int)) {
	workers := min(runtime.GOMAXPROCS(0), max(n/16, 1))
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for y := w; y < n; y += workers {
				f(y)
			}
		})
	}
	wg.Wait()
}

func clamp(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// smooth is 0 below lo, 1 above hi and a smoothstep between.
func smooth(v, lo, hi float64) float64 {
	t := clamp((v - lo) / (hi - lo))
	return t * t * (3 - 2*t)
}

// D65 white of the sRGB primaries, as in internal/worldmaterial.
const whiteX, whiteZ = 0.95047, 1.08883

var linear = func() (t [256]float64) {
	for i := range t {
		c := float64(i) / 255
		if c <= 0.04045 {
			t[i] = c / 12.92
		} else {
			t[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return t
}()

func encode(v float64) uint8 {
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
func toLab(r8, g8, b8 uint8) (l, a, b float64) {
	r, g, bl := linear[r8], linear[g8], linear[b8]
	x := (0.4124564*r + 0.3575761*g + 0.1804375*bl) / whiteX
	y := 0.2126729*r + 0.7151522*g + 0.0721750*bl
	z := (0.0193339*r + 0.1191920*g + 0.9503041*bl) / whiteZ
	fx, fy, fz := labF(x), labF(y), labF(z)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

// fromLab converts CIELAB (D65) to sRGB with alpha, clipping out-of-gamut channels.
func fromLab(l, a, b float64, alpha uint8) color.NRGBA {
	fy := (l + 16) / 116
	x := labInverse(fy+a/500) * whiteX
	y := labInverse(fy)
	z := labInverse(fy-b/200) * whiteZ
	return color.NRGBA{
		encode(3.2404542*x - 1.5371385*y - 0.4985314*z),
		encode(-0.9692660*x + 1.8760108*y + 0.0415560*z),
		encode(0.0556434*x - 0.2040259*y + 1.0572252*z),
		alpha,
	}
}
