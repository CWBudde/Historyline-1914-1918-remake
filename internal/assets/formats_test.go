package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gen2brain/webp"
)

// compressedFixture is a 1x and 2x pack of one grey sprite, encoded by enc
// under the extension ext; its left column is clear.
func compressedFixture(t *testing.T, ext string, enc func(io.Writer, image.Image) error) (fstest.MapFS, Manifest) {
	t.Helper()
	f := fstest.MapFS{}
	m := Manifest{Schema: 1, ID: "lite", Status: "approved"}
	for _, d := range []int{1, 2} {
		im := image.NewNRGBA(image.Rect(0, 0, 8*d, 8*d))
		for y := range 8 * d {
			for x := d; x < 8*d; x++ {
				im.SetNRGBA(x, y, color.NRGBA{R: 128, G: 128, B: 128, A: 255})
			}
		}
		var b bytes.Buffer
		if err := enc(&b, im); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		e := Entry{ID: "unit/lite/cp/n-normal", Path: "images/" + string(rune('0'+d)) + ext, SHA256: hex.EncodeToString(sum[:]),
			Width: 8 * d, Height: 8 * d, Density: d, Anchor: [2]float64{4, 4}}
		f[e.Path] = &fstest.MapFile{Data: b.Bytes()}
		m.Entries = append(m.Entries, e)
	}
	return f, m
}

func TestCompressedFormats(t *testing.T) {
	for _, c := range []struct {
		ext   string
		alpha bool
		enc   func(io.Writer, image.Image) error
	}{
		{".jpg", false, func(w io.Writer, im image.Image) error { return jpeg.Encode(w, im, &jpeg.Options{Quality: 90}) }},
		{".webp", true, func(w io.Writer, im image.Image) error { return webp.Encode(w, im, webp.Options{Quality: 90}) }},
	} {
		f, m := compressedFixture(t, c.ext, c.enc)
		p, err := Load(f, m, Options{})
		if err != nil {
			t.Fatalf("%s: %v", c.ext, err)
		}
		if err := p.Validate(); err != nil {
			t.Fatalf("%s: %v", c.ext, err)
		}
		for _, d := range []int{1, 2} {
			im, err := p.Image("unit/lite/cp/n-normal", d)
			if err != nil {
				t.Fatal(err)
			}
			if im.Bounds() != image.Rect(0, 0, 8*d, 8*d) {
				t.Fatalf("%s %dx: bounds %v", c.ext, d, im.Bounds())
			}
			clear := color.NRGBAModel.Convert(im.At(0, 0)).(color.NRGBA)
			solid := color.NRGBAModel.Convert(im.At(4*d, 4*d)).(color.NRGBA)
			if c.alpha && (clear.A != 0 || solid.A != 255) {
				t.Errorf("%s %dx: alpha %d and %d", c.ext, d, clear.A, solid.A)
			}
			if solid.R < 120 || solid.R > 136 {
				t.Errorf("%s %dx: colour %v", c.ext, d, solid)
			}
		}
		if err := p.WritePNG(io.Discard, "unit/lite/cp/n-normal", 1); err == nil {
			t.Errorf("WritePNG wrote a %s", c.ext)
		}

		// A header that does not match the entry is refused.
		bad := m
		bad.Entries = []Entry{m.Entries[0]}
		bad.Entries[0].Width = 16
		if _, err := Load(f, bad, Options{}); err == nil || !strings.Contains(err.Error(), "dimensions") {
			t.Errorf("%s: wrong size accepted: %v", c.ext, err)
		}
		bad.Entries[0] = m.Entries[0]
		f[bad.Entries[0].Path] = &fstest.MapFile{Data: []byte("not an image")}
		sum := sha256.Sum256([]byte("not an image"))
		bad.Entries[0].SHA256 = hex.EncodeToString(sum[:])
		if _, err := Load(f, bad, Options{}); err == nil || !strings.Contains(err.Error(), "header") {
			t.Errorf("%s: broken header accepted: %v", c.ext, err)
		}
	}
	f, m := compressedFixture(t, ".jpg", func(w io.Writer, im image.Image) error { return jpeg.Encode(w, im, nil) })
	m.Entries[0].Path = "images/1.gif"
	if _, err := Load(f, m, Options{}); err == nil || !strings.Contains(err.Error(), "path") {
		t.Errorf("a GIF entry: %v", err)
	}
}
