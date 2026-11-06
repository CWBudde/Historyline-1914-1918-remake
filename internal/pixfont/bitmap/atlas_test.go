package bitmap

import "testing"

func TestAtlasGlyphCoverage(t *testing.T) {
	for _, height := range []int{5, 7} {
		im, m := Atlas(height)
		if im.Bounds().Dy() != height || m.Height != height || m.Width != im.Bounds().Dx() {
			t.Fatal("atlas geometry")
		}
		for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 :-./!?()%',+&" {
			g, ok := m.Glyphs[string(r)]
			if !ok || g.Width < 1 || g.Height != height || g.Advance != g.Width+1 || g.X+g.Width > m.Width {
				t.Fatalf("missing/bad glyph %q: %+v", r, g)
			}
		}
		if m.Expand["ß"] != "SS" || m.Expand["…"] != "..." || m.Fallback.Width == 0 {
			t.Fatal("text normalization lost")
		}
	}
}
