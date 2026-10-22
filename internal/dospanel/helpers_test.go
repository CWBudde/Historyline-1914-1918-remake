package dospanel

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// boxText is a Text of fixed-width boxes: every rune is w×h device pixels.
// It records what was drawn and the bounds of the image it was drawn on.
type boxText struct {
	draws []textDraw
}

type textDraw struct {
	s    string
	at   image.Rectangle // the text's box
	clip image.Rectangle // the bounds of the image drawn on
}

func boxSize(size Size) (w, h int) {
	if size == Large {
		return 6, 9
	}
	return 4, 6
}

func (b *boxText) Measure(s string, size Size) (int, int) {
	w, h := boxSize(size)
	return w * len([]rune(s)), h
}

func (b *boxText) Draw(dst *ebiten.Image, s string, x, y int, size Size, c color.Color) {
	w, h := b.Measure(s, size)
	b.draws = append(b.draws, textDraw{s: s, at: image.Rect(x, y, x+w, y+h), clip: dst.Bounds()})
	if w > 0 {
		dst.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image).Fill(c)
	}
}

// has reports whether s was drawn.
func (b *boxText) has(s string) bool {
	for _, d := range b.draws {
		if d.s == s {
			return true
		}
	}
	return false
}

// panelCase is a panel at a scale, from the smallest window to 4K.
type panelCase struct {
	name  string
	panel image.Rectangle
	scale float64
}

var panelCases = []panelCase{
	{"reference", image.Rect(0, 0, 152, 180), 1},
	{"reference right", image.Rect(160, 0, 312, 180), 1},
	{"smallest", image.Rect(0, 0, 300, 280), 2},
	{"720p", image.Rect(8, 8, 632, 680), 2},
	{"1080p", image.Rect(964, 12, 1904, 1000), 3},
	{"4K", image.Rect(1928, 16, 3824, 2048), 4},
	{"wide", image.Rect(0, 0, 1200, 400), 2},
	{"tall", image.Rect(40, 0, 340, 900), 4},
}

// checkInside fails for every named rect that is empty or not inside outer.
func checkInside(t *testing.T, outer image.Rectangle, rects map[string]image.Rectangle) {
	t.Helper()
	for n, r := range rects {
		if r.Empty() {
			t.Errorf("%s is empty", n)
		} else if !r.In(outer) {
			t.Errorf("%s %v not inside %v", n, r, outer)
		}
	}
}

// checkDisjoint fails for every pair of named rects that overlaps.
func checkDisjoint(t *testing.T, rects map[string]image.Rectangle) {
	t.Helper()
	for a, ra := range rects {
		for b, rb := range rects {
			if a < b && ra.Overlaps(rb) {
				t.Errorf("%s %v overlaps %s %v", a, ra, b, rb)
			}
		}
	}
}

func centre(r image.Rectangle) (int, int) { return (r.Min.X + r.Max.X) / 2, (r.Min.Y + r.Max.Y) / 2 }
