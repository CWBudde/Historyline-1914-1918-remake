package terrainstamps

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"path/filepath"
	"testing"
)

// railDeckWiden is a v3 rail deck master and its v4: the same deck with its two
// rails moved apart to the widened track's gauge (9 logical px, 108 export px,
// from 4.5). rails are the native x of the heads in v1 (railDeckRegrades);
// each rail moves shift native px outward. On the sleeper rows deck each rail
// carries its assembly along (the native columns of its tie plates, its head
// and a few px of deck either side, [first, last+1)); the deck between the two
// assemblies takes the columns the assemblies moved over outside them, and the
// seams cross-fade over widenFade px. On the end slabs (every other row) only
// the head columns move: their old place takes the slab beside them, mirrored.
// Last, on every row, each moved head becomes head native px of flat steel
// about its middle (the row's mean head colour), so that on the map its dark
// band is as wide as the wide track's rails (1.4 logical steel with soft
// edges). Alpha stays as it is everywhere.
type railDeckWiden struct {
	from, to   string
	rails      [2]int
	shift      int
	assemblies [2][2]int
	deck       [2]int
	head       int
}

// The shifts set the native gauge to 108 export px at each master's
// registration scale (357/404 summer, 357/429 winter): 68+54 = 122 native px
// (107.8 export), 70+60 = 130 (108.2). The heads come to 13 native px (11.5
// export, 0.96 logical) and 15 (12.5, 1.04): on the composed map their dark
// run is 1.0 logical wide like the track's 0.75-1.0 (TestStandingRailDeck).
var railDeckWidens = []railDeckWiden{
	{"art/terrain/rBridge/rbridge-36-rail-summer-v3.png", "art/terrain/rBridge/rbridge-36-rail-summer-v4.png", [2]int{652, 720}, 27, [2][2]int{{622, 682}, {692, 753}}, [2]int{208, 892}, 13},
	{"art/terrain/rBridge/rbridge-36-rail-winter-v3.png", "art/terrain/rBridge/rbridge-36-rail-winter-v4.png", [2]int{653, 723}, 30, [2][2]int{{621, 683}, {695, 757}}, [2]int{215, 910}, 15},
}

// wideHead returns, per rail of w, the first native column of its stretched
// head in v4.
func wideHead(h [2]int, r int, w railDeckWiden) int {
	c := float64(h[0]+h[1]+1)/2 + float64((2*r-1)*w.shift)
	return int(math.Round(c - float64(w.head)/2))
}

const widenFade = 6

// widenSegment maps the dest columns from from on (up to the next segment's
// from) to the source columns x+off.
type widenSegment struct{ from, off int }

// widenSegments returns the column map of the deck rows of w.
func widenSegments(w railDeckWiden) []widenSegment {
	l0, l1 := w.assemblies[0][0], w.assemblies[0][1]
	r0, r1 := w.assemblies[1][0], w.assemblies[1][1]
	d, h := w.shift, (r0-l1)/2
	return []widenSegment{
		{math.MinInt32, 0},
		{l0 - d, d},               // the left assembly and the deck right of it
		{l1 + h - d, l0 - l1 - h}, // the columns left of the left assembly it moved over
		{l1 + h, r1 - l1 - h},     // the columns right of the right assembly it moved over
		{l1 + h + d, -d},          // the deck left of the right assembly and the assembly
		{r1 + d, 0},
	}
}

// widenRails returns v4 of the v3 master im; v1 locates the rail heads.
func widenRails(im, v1 *image.NRGBA, w railDeckWiden) *image.NRGBA {
	out := image.NewNRGBA(im.Rect)
	copy(out.Pix, im.Pix)
	segs := widenSegments(w)
	b := im.Rect
	lo, hi := w.assemblies[0][0]-w.shift-widenFade, w.assemblies[1][1]+w.shift+widenFade
	at := func(x, y int) [3]float64 {
		x = max(b.Min.X, min(b.Max.X-1, x))
		i := im.PixOffset(x, y)
		return [3]float64{float64(im.Pix[i]), float64(im.Pix[i+1]), float64(im.Pix[i+2])}
	}
	for y := w.deck[0]; y < w.deck[1]; y++ {
		for x := lo; x < hi; x++ {
			k := 0
			for k+1 < len(segs) && x >= segs[k+1].from {
				k++
			}
			c := at(x+segs[k].off, y)
			// Cross-fade over widenFade px centred on each junction.
			for j := 1; j < len(segs); j++ {
				t := (float64(x-segs[j].from) + 0.5 + widenFade/2.0) / widenFade
				if t <= 0 || t >= 1 {
					continue
				}
				p, q := at(x+segs[j-1].off, y), at(x+segs[j].off, y)
				for ch := range 3 {
					c[ch] = p[ch] + t*(q[ch]-p[ch])
				}
			}
			i := out.PixOffset(x, y)
			for ch := range 3 {
				out.Pix[i+ch] = uint8(math.Round(c[ch]))
			}
		}
	}
	heads := railHeadColumns(v1, w.rails)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if y >= w.deck[0] && y < w.deck[1] {
			continue
		}
		for r, h := range heads {
			dir := 2*r - 1
			mid := (h[0] + h[1]) / 2
			for x := h[0]; x <= h[1]; x++ {
				src := h[0] - 1 - (x - h[0])
				if x > mid {
					src = h[1] + 1 + (h[1] - x)
				}
				d, s := out.PixOffset(x, y), im.PixOffset(src, y)
				if im.Pix[d+3] != 0 {
					copy(out.Pix[d:d+3], im.Pix[s:s+3])
				}
			}
			for x := h[0]; x <= h[1]; x++ {
				d, s := out.PixOffset(x+dir*w.shift, y), im.PixOffset(x, y)
				if out.Pix[d+3] != 0 && im.Pix[s+3] != 0 {
					copy(out.Pix[d:d+3], im.Pix[s:s+3])
				}
			}
		}
	}
	// Widen each head to w.head columns of flat steel: per row the mean colour
	// of the painted head's columns, so the rail's dark band is as even across
	// as the track's and centred where the head was.
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for r, h := range heads {
			var c [3]float64
			n := 0.0
			for x := h[0]; x <= h[1]; x++ {
				p := im.PixOffset(x, y)
				if im.Pix[p+3] == 0 {
					continue
				}
				for ch := range 3 {
					c[ch] += float64(im.Pix[p+ch])
				}
				n++
			}
			if n == 0 {
				continue
			}
			x0 := wideHead(h, r, w)
			for k := range w.head {
				d := out.PixOffset(x0+k, y)
				if out.Pix[d+3] == 0 {
					continue
				}
				for ch := range 3 {
					out.Pix[d+ch] = uint8(math.Round(c[ch] / n))
				}
			}
		}
	}
	return out
}

// v1Of returns the painted v1 master of the v3 master path.
func v1Of(t *testing.T, root, v3 string) *image.NRGBA {
	t.Helper()
	for _, r := range railDeckRegrades {
		if r.lifted == v3 {
			return decodeRailDeckPNG(t, filepath.Join(root, r.from))
		}
	}
	t.Fatalf("%s: no v1", v3)
	return nil
}

// The v4 rail deck masters are the v3 ones with the rails moved apart to the
// widened track's gauge: the alpha is unchanged and the heads sit shift px
// further out than v1's, as wide as before.
func TestRailDeckMastersWidenRails(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, w := range railDeckWidens {
		from := decodeRailDeckPNG(t, filepath.Join(root, w.from))
		v1 := v1Of(t, root, w.from)
		got := decodeRailDeckPNG(t, filepath.Join(root, w.to))
		if want := widenRails(from, v1, w); !bytes.Equal(got.Pix, want.Pix) {
			t.Errorf("%s is not the widening of %s; run the bridgegfx generator", w.to, w.from)
		}
		for i := 3; i < len(got.Pix); i += 4 {
			if got.Pix[i] != from.Pix[i] {
				t.Fatalf("%s: alpha changed", w.to)
			}
		}
		// Each head sits shift px further out, w.head columns wide, its mean
		// L* within 1 of the painted head's.
		heads := railHeadColumns(v1, w.rails)
		for r, h := range heads {
			x0 := wideHead(h, r, w)
			c := float64(x0) + float64(w.head)/2 - float64(h[0]+h[1]+1)/2
			var lo, ln, wo, wn float64
			for y := got.Rect.Min.Y; y < got.Rect.Max.Y; y++ {
				for x := h[0]; x <= h[1]; x++ {
					if from.Pix[from.PixOffset(x, y)+3] >= railDeckOpaqueMinimum {
						l, _, _ := srgbToLab(from.NRGBAAt(x, y))
						lo, ln = lo+l, ln+1
					}
				}
				for x := x0; x < x0+w.head; x++ {
					if got.Pix[got.PixOffset(x, y)+3] >= railDeckOpaqueMinimum {
						l, _, _ := srgbToLab(got.NRGBAAt(x, y))
						wo, wn = wo+l, wn+1
					}
				}
			}
			msg := fmt.Sprintf("%s: rail %d head %v moved %+.1f px to %d columns from %d, L* %.1f from %.1f", w.to, r, h, c, w.head, x0, wo/wn, lo/ln)
			if math.Abs(c-float64((2*r-1)*w.shift)) > 0.5 || math.Abs(wo/wn-lo/ln) > 1 {
				t.Error(msg)
			} else {
				t.Log(msg)
			}
		}
	}
}
