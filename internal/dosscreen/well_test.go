package dosscreen

import (
	"fmt"
	"image/color"
	"testing"
)

// The text area and the progress bar lie inside the status well's bevel: the
// text centred with a small inset, the bar along the bottom under it, neither
// touching the bevel or each other.
func TestWellGeometry(t *testing.T) {
	for _, w := range [][2]int{{1280, 720}, {1920, 1080}, {3840, 2160}} {
		for _, ui := range []float64{1, 1.5, 2} {
			s := Layout(w[0], w[1], ui)
			bw := scaled(bevelNative, ui)
			for side, st := range s.Status {
				name := fmt.Sprintf("%dx%d@%v side %d", w[0], w[1], ui, side)
				inner := Rect{st.X + bw, st.Y + bw, st.W - 2*bw, st.H - 2*bw}
				text, bar := s.WellText(side), s.WellBar(side)
				for _, r := range []struct {
					name string
					r    Rect
				}{{"text", text}, {"bar", bar}} {
					if r.r.W <= 0 || r.r.H <= 0 || r.r.X < inner.X || r.r.Y < inner.Y || r.r.Right() > inner.Right() || r.r.Bottom() > inner.Bottom() {
						t.Errorf("%s: %s %+v not inside the bevel of %+v", name, r.name, r.r, st)
					}
				}
				if text.X == inner.X || text.Y == inner.Y || text.Right() == inner.Right() {
					t.Errorf("%s: text %+v has no inset in %+v", name, text, inner)
				}
				if text.X-st.X != st.Right()-text.Right() {
					t.Errorf("%s: text %+v not centred in %+v", name, text, st)
				}
				if text.Overlaps(bar) || text.Bottom() > bar.Y {
					t.Errorf("%s: text %+v and bar %+v overlap", name, text, bar)
				}
				if bar.Bottom() != inner.Bottom() || bar.H != scaled(barNative, ui) || bar.X-st.X != st.Right()-bar.Right() {
					t.Errorf("%s: bar %+v not along the bottom of %+v", name, bar, inner)
				}
				if text.H < scaled(glyphNative, ui) {
					t.Errorf("%s: text %+v lower than a line of capitals", name, text)
				}
			}
		}
	}
}

// The text colours are the reference's pale green (left) and cream (right); the
// numbers are amber and the bar grey, all well off the dark fill.
func TestWellColours(t *testing.T) {
	dark := color.RGBA{0x18, 0x1C, 0x00, 0xFF}
	if TextColour(0) != (color.RGBA{0xBE, 0xD3, 0x92, 0xFF}) || TextColour(1) != (color.RGBA{0xFB, 0xE3, 0xC7, 0xFF}) {
		t.Errorf("text colours %v %v", TextColour(0), TextColour(1))
	}
	for _, c := range []color.RGBA{TextColour(0), TextColour(1), NumberColour, BarColour} {
		if c.A != 255 || luma(c) < luma(dark)+90 {
			t.Errorf("colour %v does not stand off the dark strip %v", c, dark)
		}
	}
	if NumberColour == TextColour(0) || NumberColour == TextColour(1) {
		t.Errorf("numbers %v share a text colour", NumberColour)
	}
}

func luma(c color.RGBA) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }
