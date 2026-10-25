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
// (v2) the regraded one, steel the mean L*, a*, b* of the rail cores of the
// same season's 4x tile/rail36 when v2 was made (the narrow track, its rails
// at 4x columns 182-183 and 200-201), rails the native x of the deck's two
// rail heads. lifted (v3) is v2 with the shadow columns beside each head
// lifted to the deck next to them.
type railDeckRegrade struct {
	from, to, lifted string
	steel            [3]float64
	rails            [2]int
}

var railDeckRegrades = []railDeckRegrade{
	{"art/terrain/rBridge/rbridge-36-rail-summer-v1.png", "art/terrain/rBridge/rbridge-36-rail-summer-v2.png", "art/terrain/rBridge/rbridge-36-rail-summer-v3.png", [3]float64{30.63103831436162, 3.265265501484798, 3.7455122748144603}, [2]int{652, 720}},
	{"art/terrain/rBridge/rbridge-36-rail-winter-v1.png", "art/terrain/rBridge/rbridge-36-rail-winter-v2.png", "art/terrain/rBridge/rbridge-36-rail-winter-v3.png", [3]float64{24.20138726815569, 2.03139208535672, 0.6373898691468278}, [2]int{653, 723}},
}

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
// moved onto the dark steel: each head pixel's L*, a* and b* keep headSpread
// of their deviation from the heads' mean around the steel's mean, blended in
// by headWeight. Every other pixel, and the alpha, stay as painted.
func regradeRailHeads(im *image.NRGBA, steel [3]float64, rails [2]int) *image.NRGBA {
	out := image.NewNRGBA(im.Rect)
	copy(out.Pix, im.Pix)
	heads := railHeadColumns(im, rails)
	tl, ta, tb := steel[0], steel[1], steel[2]
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
		want := regradeRailHeads(from, r.steel, r.rails)
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
		tl := r.steel[0]
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

// The shadow columns beside a rail head are the columns next to it, on each
// side, whose mean L* (over the opaque rows) stays more than flankMargin
// below the deck's: the mean of the flankRef columns flankRef0 to flankRef1
// px further out. A flank runs at most flankMax columns.
const (
	flankRef0, flankRef1 = 20, 28
	flankMargin          = 3.0
	flankMax             = 24
)

// columnL is the mean L* of column x of im over its opaque rows.
func columnL(im *image.NRGBA, x int) float64 {
	s, n := 0.0, 0
	for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
		c := im.NRGBAAt(x, y)
		if c.A < railDeckOpaqueMinimum {
			continue
		}
		l, _, _ := srgbToLab(c)
		s, n = s+l, n+1
	}
	return s / float64(n)
}

// railFlanks returns, per rail head of heads, its two shadow columns spans in
// im: [0] the left one, [1] the right one, each as first and last column.
func railFlanks(im *image.NRGBA, heads [2][2]int) [2][2][2]int {
	var flanks [2][2][2]int
	for r, h := range heads {
		for side, dir := range [2]int{-1, 1} {
			edge := h[0]
			if dir > 0 {
				edge = h[1]
			}
			ref := 0.0
			for d := flankRef0; d < flankRef1; d++ {
				ref += columnL(im, edge+dir*d)
			}
			ref /= flankRef1 - flankRef0
			n := 0
			for n < flankMax && columnL(im, edge+dir*(n+1)) < ref-flankMargin {
				n++
			}
			if dir < 0 {
				flanks[r][side] = [2]int{edge - n, edge - 1}
			} else {
				flanks[r][side] = [2]int{edge + 1, edge + n}
			}
		}
	}
	return flanks
}

// liftRailFlanks returns a copy of im with its shadow columns flanks lifted to
// the deck beside them: each row of a flank takes the colour of the deck just
// outside it, mirrored about the flank's outer edge, so sleepers, plates and
// ballast run on to the rail head. Alpha and every other pixel stay as they are.
func liftRailFlanks(im *image.NRGBA, flanks [2][2][2]int) *image.NRGBA {
	out := image.NewNRGBA(im.Rect)
	copy(out.Pix, im.Pix)
	for _, rail := range flanks {
		for side, f := range rail {
			for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
				for k := 0; k <= f[1]-f[0]; k++ {
					dst, src := f[0]+k, f[0]-1-k // left: mirrored about f[0]-0.5
					if side == 1 {
						dst, src = f[1]-k, f[1]+1+k // right: about f[1]+0.5
					}
					d, s := out.PixOffset(dst, y), im.PixOffset(src, y)
					if im.Pix[d+3] == 0 {
						continue
					}
					copy(out.Pix[d:d+3], im.Pix[s:s+3])
				}
			}
		}
	}
	return out
}

// inFlank reports whether column x lies in one of flanks.
func inFlank(flanks [2][2][2]int, x int) bool {
	for _, rail := range flanks {
		for _, f := range rail {
			if x >= f[0] && x <= f[1] {
				return true
			}
		}
	}
	return false
}

// The v3 rail deck masters are the v2 ones with the shadow columns beside each
// rail head lifted to the deck next to them, so the dark run across a rail is
// its head alone (tile/rail36's rails are 0.5 logical wide): every pixel
// outside the shadow columns is within MAD 1/255 of v2, the heads keep their
// dark steel, and each shadow column comes within flankMargin of its deck.
func TestRailDeckMastersLiftRailFlanks(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, r := range railDeckRegrades {
		from := decodeRailDeckPNG(t, filepath.Join(root, r.from))
		v2 := decodeRailDeckPNG(t, filepath.Join(root, r.to))
		heads := railHeadColumns(from, r.rails)
		flanks := railFlanks(from, heads)
		got := decodeRailDeckPNG(t, filepath.Join(root, r.lifted))
		if want := liftRailFlanks(v2, flanks); !bytes.Equal(got.Pix, want.Pix) {
			t.Errorf("%s is not the lift of %s; run the bridgegfx generator", r.lifted, r.to)
		}
		var mad float64
		n := 0
		for y := got.Rect.Min.Y; y < got.Rect.Max.Y; y++ {
			for x := got.Rect.Min.X; x < got.Rect.Max.X; x++ {
				if inFlank(flanks, x) {
					continue
				}
				i := got.PixOffset(x, y)
				for c := range 4 {
					mad += math.Abs(float64(got.Pix[i+c]) - float64(v2.Pix[i+c]))
				}
				n += 4
			}
		}
		mad /= float64(n)
		t.Logf("%s: heads %v, shadow columns %v, MAD outside them %.4f/255", r.lifted, heads, flanks, mad)
		if mad > 1 {
			t.Errorf("%s: MAD %.3f/255 outside the shadow columns, want at most 1", r.lifted, mad)
		}
		for i, rail := range flanks {
			for _, f := range rail {
				if w := f[1] - f[0] + 1; w < 4 || w > flankMax {
					t.Errorf("%s: rail %d shadow columns %v, want 4 to %d", r.from, i, f, flankMax)
				}
				// The lifted columns come to the deck's tone: their mean L* is
				// within flankMargin of the columns they mirror.
				var lf, ls float64
				for x := f[0]; x <= f[1]; x++ {
					lf += columnL(got, x)
				}
				w := f[1] - f[0] + 1
				src := f[0] - w
				if f[0] > heads[i][1] {
					src = f[1] + 1
				}
				for x := src; x < src+w; x++ {
					ls += columnL(got, x)
				}
				if d := (lf - ls) / float64(w); math.Abs(d) > flankMargin {
					t.Errorf("%s: shadow columns %v L* %.1f off their deck's", r.lifted, f, d)
				}
			}
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
