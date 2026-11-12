package hlwebapp

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/webp"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlasset"
)

// decodeLossless reads an exported full-tier image back to pixels. The decoder of
// golang.org/x/image reads VP8L exactly (the webp package used for encoding returns
// YUV even for lossless data).
func decodeLossless(raw []byte) (image.Image, error) {
	return webp.Decode(bytes.NewReader(raw))
}

// chunks lists the RIFF chunk types of a WebP file.
func chunks(t *testing.T, data []byte) []string {
	t.Helper()
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		t.Fatalf("not a WebP file: % x", data[:min(len(data), 12)])
	}
	var out []string
	for p := 12; p+8 <= len(data); {
		n := int(binary.LittleEndian.Uint32(data[p+4:]))
		out = append(out, string(data[p:p+4]))
		p += 8 + n + n&1
	}
	return out
}

// sample has hard alpha edges, partial alpha and colour under alpha 0, which a
// default (inexact) lossless encode throws away.
func sample() *image.NRGBA {
	rng := rand.New(rand.NewSource(7))
	im := image.NewNRGBA(image.Rect(0, 0, 37, 29))
	for y := range 29 {
		for x := range 37 {
			a := uint8(255)
			switch {
			case (x+y)%5 == 0:
				a = 0
			case x%7 == 0:
				a = uint8(rng.Intn(256))
			}
			im.SetNRGBA(x, y, color.NRGBA{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), a})
		}
	}
	return im
}

func TestLosslessWebPKeepsEveryPixel(t *testing.T) {
	src := sample()
	data, err := hlasset.EncodeLossless(src)
	if err != nil {
		t.Fatal(err)
	}
	var vp8l bool
	for _, c := range chunks(t, data) {
		if c == "VP8 " {
			t.Fatal("lossy VP8 chunk")
		}
		vp8l = vp8l || c == "VP8L"
	}
	if !vp8l {
		t.Fatal("no VP8L chunk")
	}
	got, err := decodeLossless(data)
	if err != nil {
		t.Fatal(err)
	}
	n := image.NewNRGBA(got.Bounds())
	for y := range n.Rect.Dy() {
		for x := range n.Rect.Dx() {
			n.Set(x, y, got.At(x, y))
		}
	}
	if !bytes.Equal(n.Pix, src.Pix) {
		t.Error("decoded pixels differ from the source, including the colour under alpha 0")
	}
}

func TestFullArtIsCachedLosslessWebP(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	src := sample()
	b := blobWriter{root: t.TempDir()}
	ref, err := b.raster(src)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(ref.Path) != ".webp" {
		t.Fatalf("path %s", ref.Path)
	}
	data, err := os.ReadFile(filepath.Join(b.root, ref.Path))
	if err != nil {
		t.Fatal(err)
	}
	chunks(t, data)
	cached, err := filepath.Glob(filepath.Join(os.Getenv("XDG_CACHE_HOME"), "historyline", "webapp-lossless-v1", "*.webp"))
	if err != nil || len(cached) != 1 {
		t.Fatalf("cache entries %v (%v)", cached, err)
	}
	// A second export reads the cache: a damaged-free entry gives the same blob.
	again, err := b.raster(src)
	if err != nil || again != ref {
		t.Fatalf("second export %v, %v", again, err)
	}
	// Other formats pass through; pixel fonts stay PNG.
	if r, err := b.fullArt("jpg", src, []byte("x")); err != nil || filepath.Ext(r.Path) != ".jpg" {
		t.Fatalf("%v %v", r, err)
	}
	if r, err := b.png(src); err != nil || filepath.Ext(r.Path) != ".png" {
		t.Fatalf("%v %v", r, err)
	}
}
