package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// railDeckRegrade is one rail deck master whose rail heads are regraded onto
// the dark steel of the track it carries: from (v1) is the painted master, to
// (v2) the regraded one, track the 4x tile/rail36 of the same season, rails
// the native x of the deck's two rail heads.
type railDeckRegrade struct {
	from, to, track string
	rails           [2]int
}

var railDeckRegrades = []railDeckRegrade{
	{"art/terrain/rBridge/rbridge-36-rail-summer-v1.png", "art/terrain/rBridge/rbridge-36-rail-summer-v2.png", "assets/terrain/rail/rail36-4x.png", [2]int{652, 720}},
	{"art/terrain/rBridge/rbridge-36-rail-winter-v1.png", "art/terrain/rBridge/rbridge-36-rail-winter-v2.png", "assets/terrain/rail/rail36_w-4x.png", [2]int{653, 723}},
}

// trackRails are the columns of tile/rail36's two rail cores at 4x: its rails
// lie 2.375 logical left and 2.125 right of the hex centre (x 48).
var trackRails = [2][2]int{{182, 183}, {200, 201}}

// The rail heads of a deck master are the columns within headReach of a rail's
// x whose mean L* (over the opaque rows) is headL or more: the lit steel
// between the rail's dark flanks. A head pixel takes the regrade to the share
// clamp((L*-headFloor)/headRamp), so the flanks and the sleepers' shadows
// between them stay as painted.
const (
	headReach             = 6
	headL                 = 35.0
	headFloor, headRamp   = 20.0, 15.0
	headSpread            = 0.6 // the share of the head's L*, a*, b* variation kept
	railDeckOpaqueMinimum = 128
)

// steelStats is the mean L*, a*, b* of the track's rail cores.
func steelStats(track *image.NRGBA) (l, a, b float64) {
	n := 0
	for y := track.Rect.Min.Y; y < track.Rect.Max.Y; y++ {
		for _, rail := range trackRails {
			for _, x := range rail {
				c := track.NRGBAAt(x, y)
				if c.A < railDeckOpaqueMinimum {
					continue
				}
				pl, pa, pb := srgbToLab(c)
				l, a, b, n = l+pl, a+pa, b+pb, n+1
			}
		}
	}
	return l / float64(n), a / float64(n), b / float64(n)
}

// railHeadColumns returns, per rail of x near rails, the first and last column
// of its head in im.
func railHeadColumns(im *image.NRGBA, rails [2]int) [2][2]int {
	var heads [2][2]int
	for r, cx := range rails {
		first, last := -1, -1
		for x := cx - headReach; x <= cx+headReach; x++ {
			s, n := 0.0, 0
			for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
				c := im.NRGBAAt(x, y)
				if c.A < railDeckOpaqueMinimum {
					continue
				}
				l, _, _ := srgbToLab(c)
				s, n = s+l, n+1
			}
			if n > 0 && s/float64(n) >= headL {
				if first < 0 {
					first = x
				}
				last = x
			}
		}
		heads[r] = [2]int{first, last}
	}
	return heads
}

// headWeight is the share of the regrade a head pixel of L* l takes.
func headWeight(l float64) float64 {
	return math.Max(0, math.Min(1, (l-headFloor)/headRamp))
}

// regradeRailHeads returns a copy of the deck master im with its rail heads
// moved onto the dark steel of track: each head pixel's L*, a* and b* keep
// headSpread of their deviation from the heads' mean around the steel's mean,
// blended in by headWeight. Every other pixel, and the alpha, stay as painted.
func regradeRailHeads(im, track *image.NRGBA, rails [2]int) *image.NRGBA {
	out := image.NewNRGBA(im.Rect)
	copy(out.Pix, im.Pix)
	heads := railHeadColumns(im, rails)
	tl, ta, tb := steelStats(track)
	var hl, ha, hb, hw float64
	for _, h := range heads {
		for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
			for x := h[0]; x <= h[1]; x++ {
				c := im.NRGBAAt(x, y)
				if c.A == 0 {
					continue
				}
				l, a, b := srgbToLab(c)
				w := headWeight(l)
				hl, ha, hb, hw = hl+w*l, ha+w*a, hb+w*b, hw+w
			}
		}
	}
	hl, ha, hb = hl/hw, ha/hw, hb/hw
	for _, h := range heads {
		for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
			for x := h[0]; x <= h[1]; x++ {
				c := im.NRGBAAt(x, y)
				if c.A == 0 {
					continue
				}
				l, a, b := srgbToLab(c)
				w := headWeight(l)
				if w == 0 {
					continue
				}
				s := labToSRGB(tl+headSpread*(l-hl), ta+headSpread*(a-ha), tb+headSpread*(b-hb))
				mix := func(p, q uint8) uint8 { return uint8(math.Round(float64(p) + w*(float64(q)-float64(p)))) }
				out.SetNRGBA(x, y, color.NRGBA{mix(c.R, s.R), mix(c.G, s.G), mix(c.B, s.B), c.A})
			}
		}
	}
	return out
}

// The v2 rail deck masters are the v1 paintings with their rail heads regraded
// onto tile/rail36's dark steel: the heads come to the steel's mean L* and
// every pixel outside the head columns is unchanged.
func TestRailDeckMastersRegradeRailHeads(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, r := range railDeckRegrades {
		from := decodeRailDeckPNG(t, filepath.Join(root, r.from))
		track := decodeRailDeckPNG(t, filepath.Join(root, r.track))
		want := regradeRailHeads(from, track, r.rails)
		got := decodeRailDeckPNG(t, filepath.Join(root, r.to))
		if !bytes.Equal(got.Pix, want.Pix) {
			t.Errorf("%s is not the regrade of %s; run the bridgegfx generator", r.to, r.from)
		}
		heads := railHeadColumns(from, r.rails)
		inHead := func(x int) bool {
			for _, h := range heads {
				if x >= h[0] && x <= h[1] {
					return true
				}
			}
			return false
		}
		changed := 0
		var sl, sw float64
		for y := got.Rect.Min.Y; y < got.Rect.Max.Y; y++ {
			for x := got.Rect.Min.X; x < got.Rect.Max.X; x++ {
				i := got.PixOffset(x, y)
				if !bytes.Equal(got.Pix[i:i+4], from.Pix[i:i+4]) && !inHead(x) {
					changed++
				}
				if inHead(x) && from.Pix[i+3] >= railDeckOpaqueMinimum {
					l0, _, _ := srgbToLab(from.NRGBAAt(x, y))
					if w := headWeight(l0); w == 1 {
						l, _, _ := srgbToLab(got.NRGBAAt(x, y))
						sl, sw = sl+l, sw+1
					}
				}
			}
		}
		tl, _, _ := steelStats(track)
		t.Logf("%s: heads %v, head L* %.1f, steel L* %.1f", r.to, heads, sl/sw, tl)
		if changed > 0 {
			t.Errorf("%s: %d pixels outside the rail heads changed", r.to, changed)
		}
		for _, h := range heads {
			if w := h[1] - h[0] + 1; h[0] < 0 || w < 5 || w > 9 {
				t.Errorf("%s: rail head columns %v, want 5 to 9", r.from, h)
			}
		}
		if d := sl/sw - tl; math.Abs(d) > 2 {
			t.Errorf("%s: head L* %.1f, want within 2 of the steel's %.1f", r.to, sl/sw, tl)
		}
	}
}

func decodeRailDeckPNG(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(src.Bounds())
	draw.Draw(im, im.Rect, src, src.Bounds().Min, draw.Src)
	return im
}

// srgbToLab converts an sRGB colour to CIE L*a*b* (D65).
func srgbToLab(c color.NRGBA) (l, a, b float64) {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	x := (0.4124564*r + 0.3575761*g + 0.1804375*bl) / 0.95047
	y := 0.2126729*r + 0.7151522*g + 0.0721750*bl
	z := (0.0193339*r + 0.1191920*g + 0.9503041*bl) / 1.08883
	f := func(t float64) float64 {
		if t > 216.0/24389 {
			return math.Cbrt(t)
		}
		return (24389.0/27*t + 16) / 116
	}
	fx, fy, fz := f(x), f(y), f(z)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

// labToSRGB converts CIE L*a*b* (D65) to sRGB, clipping out-of-gamut channels.
func labToSRGB(l, a, b float64) color.NRGBA {
	inv := func(f float64) float64 {
		if f*f*f > 216.0/24389 {
			return f * f * f
		}
		return (116*f - 16) * 27 / 24389
	}
	fy := (l + 16) / 116
	x := inv(fy+a/500) * 0.95047
	y := inv(fy)
	z := inv(fy-b/200) * 1.08883
	enc := func(v float64) uint8 {
		v = math.Max(0, math.Min(1, v))
		if v <= 0.0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		return uint8(math.Round(v * 255))
	}
	return color.NRGBA{
		enc(3.2404542*x - 1.5371385*y - 0.4985314*z),
		enc(-0.9692660*x + 1.8760108*y + 0.0415560*z),
		enc(0.0556434*x - 0.2040259*y + 1.0572252*z),
		255,
	}
}
