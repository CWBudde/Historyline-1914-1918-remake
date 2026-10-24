package main

import (
	"image"
	"image/png"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// groundManifestPath is the ground manifest the campaign exports use.
const groundManifestPath = "art/pipeline/ground-terrain-v3.json"

// The installed tile of every lattice-sampled ground material shows no mirror
// axis: inside the hex its luminance hardly correlates with its own left-right
// or top-bottom mirror image (the mirrored sampler made both about 1).
func TestLatticeGroundTilesAreNotMirrored(t *testing.T) {
	root := filepath.Join("..", "..")
	gi, err := loadGroundInputs(filepath.Join(root, groundManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(gi.Manifest.LatticeMaterials) == 0 {
		t.Fatal("no lattice materials")
	}
	for _, id := range slices.Sorted(maps.Keys(gi.Manifest.LatticeMaterials)) {
		f, err := os.Open(filepath.Join(root, "assets/terrain/ground", id+"-4x.png"))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
		lr := mirrorCorrelation(im, func(x, y, w, h int) (int, int) { return w - 1 - x, y })
		tb := mirrorCorrelation(im, func(x, y, w, h int) (int, int) { return x, h - 1 - y })
		t.Logf("%s: correlation with its mirror image left-right %.2f, top-bottom %.2f", id, lr, tb)
		if lr > 0.3 || tb > 0.3 {
			t.Errorf("%s mirrors: left-right %.2f, top-bottom %.2f, want at most 0.3", id, lr, tb)
		}
	}
}

// mirrorCorrelation is the Pearson correlation of im's luminance with its
// mirror image (m maps a pixel to its mirror) over the disc inside the hex.
func mirrorCorrelation(im image.Image, m func(x, y, w, h int) (int, int)) float64 {
	b := im.Bounds()
	w, h := b.Dx(), b.Dy()
	luma := func(x, y int) float64 {
		r, g, bl, _ := im.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(bl)
	}
	radius := 0.45 * float64(h)
	var n, sa, sb, saa, sbb, sab float64
	for y := range h {
		for x := range w {
			if math.Hypot(float64(x)-float64(w)/2, float64(y)-float64(h)/2) > radius {
				continue
			}
			mx, my := m(x, y, w, h)
			a, c := luma(x, y), luma(mx, my)
			n, sa, sb, saa, sbb, sab = n+1, sa+a, sb+c, saa+a*a, sbb+c*c, sab+a*c
		}
	}
	cov := sab/n - sa/n*sb/n
	return cov / math.Sqrt((saa/n-sa*sa/n/n)*(sbb/n-sb*sb/n/n))
}
