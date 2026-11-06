package pixfont

import (
	"bytes"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont/bitmap"
)

func TestExportAtlasMatchesRuntime(t *testing.T) {
	for _, f := range []Face{Small, Large} {
		im, m := bitmap.Atlas(f.f.height)
		native := f.f.atlasImage()
		if im.Bounds() != native.Bounds() || !bytes.Equal(im.Pix, native.Pix) {
			t.Fatal("export changed the runtime glyphs")
		}
		for r, i := range f.f.index {
			g, ok := m.Glyphs[string(r)]
			if !ok || g.X != f.f.glyphs[i].x || g.Width != f.f.glyphs[i].w {
				t.Fatalf("metrics differ for %q", r)
			}
		}
	}
}
