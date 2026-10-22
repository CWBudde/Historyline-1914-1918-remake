package pixfont

import (
	"image"
	"image/color"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
)

const required = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 :-./!?()%',+&"

var (
	faces = map[string]Face{"Small": Small, "Large": Large}
	pale  = color.RGBA{0xBE, 0xD3, 0x92, 0xFF}
	amber = color.RGBA{0xE8, 0xA8, 0x30, 0xFF}
)

// raster draws s the way Draw does, from the same layout and atlas, onto a
// CPU image: Ebiten cannot read pixels back before a game runs.
func raster(f Face, s string, x, y, scale int, text, numbers color.Color, dst *image.RGBA) {
	atlas := f.f.atlasImage()
	f.f.layout(s, text, numbers, func(g, gx int, c color.Color) {
		r := f.f.rect(g)
		cr, cg, cb, _ := c.RGBA()
		for py := r.Min.Y; py < r.Max.Y; py++ {
			for px := r.Min.X; px < r.Max.X; px++ {
				if atlas.RGBAAt(px, py).A == 0 {
					continue
				}
				for sy := range scale {
					for sx := range scale {
						dst.SetRGBA(x+(gx+px-r.Min.X)*scale+sx, y+(py-r.Min.Y)*scale+sy,
							color.RGBA{uint8(cr >> 8), uint8(cg >> 8), uint8(cb >> 8), 0xFF})
					}
				}
			}
		}
	})
}

// ink returns the bounds of the drawn pixels of img.
func ink(img *image.RGBA) image.Rectangle {
	var r image.Rectangle
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A != 0 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}

func TestCoverage(t *testing.T) {
	for name, f := range faces {
		for _, r := range required {
			if f.f.lookup(r) == box {
				t.Errorf("%s: no glyph for %q", name, r)
			}
		}
		for r := 'a'; r <= 'z'; r++ {
			if f.f.lookup(r) != f.f.lookup(r-'a'+'A') {
				t.Errorf("%s: %q does not draw as upper case", name, r)
			}
		}
		for _, r := range "#€\n\x00~" {
			if f.f.lookup(r) != box {
				t.Errorf("%s: unknown rune %q does not draw as the box", name, r)
			}
		}
		for _, u := range "ÄäÖöÜü" {
			if f.f.lookup(u) == box {
				t.Errorf("%s: %q draws as the box", name, u)
			}
		}
		w1, _ := f.Measure("ß", 1)
		w2, _ := f.Measure("SS", 1)
		if w1 != w2 {
			t.Errorf("%s: ß does not draw as SS", name)
		}
	}
	// Small has no room above its capitals: its umlauts are the plain letters.
	for _, p := range []string{"ÄA", "ÖO", "ÜU"} {
		u, n := utf8.DecodeRuneInString(p)
		base := rune(p[n])
		if Small.f.lookup(u) != Small.f.lookup(base) {
			t.Errorf("Small: %q does not draw as %q", u, base)
		}
		if Large.f.lookup(u) == Large.f.lookup(base) {
			t.Errorf("Large: %q has no dotted glyph", u)
		}
	}
}

// Every glyph is a full grid of the face's height, and every inked glyph
// touches its first and last column, so the spacing is even and Measure tight.
// No two glyphs look alike.
func TestGlyphShapes(t *testing.T) {
	for name, f := range faces {
		seen := map[string]rune{strings.Join(f.f.glyphs[box].rows, "/"): utf8.RuneError}
		for r, g := range f.f.index {
			gl := f.f.glyphs[g]
			if len(gl.rows) != f.f.height {
				t.Errorf("%s %q: %d rows, want %d", name, r, len(gl.rows), f.f.height)
				continue
			}
			first, last := false, false
			for _, row := range gl.rows {
				if len(row) != gl.w || strings.Trim(row, ".#") != "" {
					t.Errorf("%s %q: bad row %q (width %d)", name, r, row, gl.w)
				}
				first = first || row[0] == '#'
				last = last || row[len(row)-1] == '#'
			}
			if r != ' ' && (!first || !last) {
				t.Errorf("%s %q: no ink in the first or last column", name, r)
			}
			key := strings.Join(gl.rows, "/")
			if o, dup := seen[key]; dup {
				t.Errorf("%s: %q and %q look alike", name, r, o)
			}
			seen[key] = r
		}
		if f.f.glyphs[box].w == 0 {
			t.Errorf("%s: empty box glyph", name)
		}
	}
	if Small.f.height != 5 || Large.f.height <= Small.f.height {
		t.Errorf("cap heights %d and %d", Small.f.height, Large.f.height)
	}
}

// The drawn pixels lie inside the measured box and touch its edges.
func TestMeasureMatchesDraw(t *testing.T) {
	texts := []string{"A", "I", "6  LIGHT ARTILLERY", "Waiting for movement-mode", required[1:len(required)-14] + "&",
		"(100%) 3/4, it's: ok!?", "Ärger über Öl", "WWWMMM", "1", "€#"}
	for name, f := range faces {
		for _, scale := range []int{1, 2, 3} {
			for _, s := range texts {
				w, h := f.Measure(s, scale)
				if h != f.f.height*scale {
					t.Errorf("%s ×%d %q: height %d", name, scale, s, h)
				}
				dst := image.NewRGBA(image.Rect(0, 0, w+40, h+40))
				raster(f, s, 20, 20, scale, pale, pale, dst)
				if got, want := ink(dst), image.Rect(20, 20, 20+w, 20+h); got != want {
					t.Errorf("%s ×%d %q: ink %v, measured %v", name, scale, s, got, want)
				}
			}
		}
		if w, h := f.Measure("", 2); w != 0 || h != 2*f.f.height {
			t.Errorf("%s: empty string measures %d×%d", name, w, h)
		}
		w0, h0 := f.Measure("A", 0)
		w1, h1 := f.Measure("A", 1)
		if w0 != w1 || h0 != h1 {
			t.Errorf("%s: scale 0 is not drawn at scale 1", name)
		}
	}
}

// Adjacent glyphs never share a pixel and keep one empty font pixel between them.
func TestNoOverlap(t *testing.T) {
	s := required + "ÄÖÜ" + required
	for name, f := range faces {
		const scale = 2
		w, h := f.Measure(s, scale)
		owner := make([]int, w*h)
		atlas := f.f.atlasImage()
		i := 0
		prevEnd := -1
		f.f.layout(s, pale, pale, func(g, gx int, _ color.Color) {
			i++
			r := f.f.rect(g)
			if prevEnd >= 0 && gx*scale < prevEnd+scale {
				t.Errorf("%s: glyph %d starts at %d, previous ends at %d", name, i, gx*scale, prevEnd)
			}
			prevEnd = (gx + r.Dx()) * scale
			for py := r.Min.Y; py < r.Max.Y; py++ {
				for px := r.Min.X; px < r.Max.X; px++ {
					if atlas.RGBAAt(px, py).A == 0 {
						continue
					}
					for sy := range scale {
						for sx := range scale {
							j := ((py-r.Min.Y)*scale+sy)*w + (gx+px-r.Min.X)*scale + sx
							if owner[j] != 0 {
								t.Fatalf("%s: glyphs %d and %d overlap", name, owner[j], i)
							}
							owner[j] = i
						}
					}
				}
			}
		})
	}
}

func TestFit(t *testing.T) {
	for name, f := range faces {
		const scale = 2
		s := "Waiting for movement-mode"
		full, _ := f.Measure(s, scale)
		if got := f.Fit(s, scale, full); got != s {
			t.Errorf("%s: a fitting text became %q", name, got)
		}
		for maxW := full - 1; maxW >= 0; maxW -= 7 {
			got := f.Fit(s, scale, maxW)
			if w, _ := f.Measure(got, scale); w > maxW {
				t.Errorf("%s: Fit(%d) = %q, %d wide", name, maxW, got, w)
			}
			dots, _ := f.Measure("...", scale)
			if maxW < dots {
				if got != "" {
					t.Errorf("%s: Fit(%d) = %q, want empty", name, maxW, got)
				}
				continue
			}
			head, ok := strings.CutSuffix(got, "...")
			if !ok || !strings.HasPrefix(s, head) || strings.HasSuffix(head, " ") {
				t.Errorf("%s: Fit(%d) = %q", name, maxW, got)
				continue
			}
			// One rune more would not fit.
			if n := len(head); n < len(s) {
				longer := strings.TrimRight(s[:n+1], " ") + "..."
				if w, _ := f.Measure(longer, scale); w <= maxW && len(longer) > len(got) {
					t.Errorf("%s: Fit(%d) = %q, but %q fits", name, maxW, got, longer)
				}
			}
		}
		if got := f.Fit("Öl über", 1, 1000); got != "Öl über" {
			t.Errorf("%s: Fit changed %q", name, got)
		}
	}
}

// DrawNumbers colours the digits, and only the digits.
func TestDrawNumbers(t *testing.T) {
	for name, f := range faces {
		s := "HP 10/12 ok"
		w, h := f.Measure(s, 1)
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		raster(f, s, 0, 0, 1, pale, amber, dst)
		x := 0
		for _, r := range s {
			gw, _ := f.Measure(string(r), 1)
			want := pale
			if r >= '0' && r <= '9' {
				want = amber
			}
			n := 0
			for py := range h {
				for px := x; px < x+gw; px++ {
					c := dst.RGBAAt(px, py)
					if c.A == 0 {
						continue
					}
					n++
					if c != want {
						t.Errorf("%s %q: pixel %v, want %v", name, r, c, want)
					}
				}
			}
			if r != ' ' && n == 0 {
				t.Errorf("%s %q: nothing drawn", name, r)
			}
			x += gw + 1
		}
	}
}

// The atlas is one image per face, built once, and every glyph's sub-image
// has the glyph's size. Drawing queues on offscreen images without a game.
func TestDrawEbiten(t *testing.T) {
	dst := ebiten.NewImage(320, 200)
	for name, f := range faces {
		f.Draw(dst, "6 LIGHT ARTILLERY", 4, 4, 1, pale)
		atlas := f.f.atlas
		f.DrawNumbers(dst, "Round 12: 3 units", 4, 20, 2, pale, amber)
		f.DrawShadow(dst, "Waiting!", 4, 60, 3, pale, color.Black)
		f.Draw(dst, "", 0, 0, 1, pale)
		if atlas == nil || f.f.atlas != atlas {
			t.Errorf("%s: atlas not cached", name)
		}
		for g, sub := range f.f.subs {
			if sub.Bounds().Size() != f.f.rect(g).Size() {
				t.Errorf("%s: glyph %d sub-image %v, want %v", name, g, sub.Bounds(), f.f.rect(g))
			}
		}
	}
}
