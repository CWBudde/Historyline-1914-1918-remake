package artreview

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
const groundManifestPath = "art/pipeline/ground-terrain-v4.json"

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
	// Shell-hole fields (mud, shelled_w) are a field material like the rough
	// winter grounds, not one crater ring pasted on the meadow per hex.
	for _, id := range []string{"mud", "rough_w", "shelled_w", "stoney_w"} {
		if _, ok := gi.Manifest.LatticeMaterials[id]; !ok {
			t.Errorf("%s is not a lattice material", id)
		}
	}
	if gi.Manifest.MudFeature != "" || gi.Manifest.ShelledFeature != "" {
		t.Errorf("mud_feature %q, shelled_feature %q: the crater fields use the lattice material", gi.Manifest.MudFeature, gi.Manifest.ShelledFeature)
	}
	for _, id := range slices.Sorted(maps.Keys(gi.Manifest.LatticeMaterials)) {
		f, err := os.Open(filepath.Join(root, "assets/terrain/ground", id+"-4x.png"))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(f)
		if closeErr := f.Close(); closeErr != nil {
			t.Error(closeErr)
		}
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

// The selected cosets are lattice-periodic, and the installed stoney_w keeps
// the lightness and contrast of its painting; the coset average came out paler
// and with half the contrast (L* 78.5 and sigma 12.4 at 2x against 73.5 and 23.4).
func TestStoneyWinterKeepsPaintedContrast(t *testing.T) {
	root := filepath.Join("..", "..")
	gi, err := loadGroundInputs(filepath.Join(root, groundManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gi.Manifest.LatticeContrast, "stoney_w") {
		t.Fatal("stoney_w does not select its cosets")
	}
	master := gi.Sources["stoney_w"].Image
	crop := latticeCrop(master, gi.Manifest.LatticeMaterials["stoney_w"])
	for _, p := range []connectionPoint{{3.25, 7.5}, {40, 33}, {90.5, 71}, {-2, 60}} {
		got := sampleSelectedMaterial(crop, p)
		for _, delta := range sharedConnectionGeometry().Neighbors {
			if shifted := sampleSelectedMaterial(crop, p.add(delta)); shifted != got {
				t.Fatalf("selected sample at %v changes under the lattice shift %v: %v != %v", p, delta, got, shifted)
			}
		}
	}
	mL, mS := imageLightness(master)
	for _, density := range []string{"2x", "4x"} {
		f, err := os.Open(filepath.Join(root, "assets/terrain/ground/stoney_w-"+density+".png"))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(f)
		if closeErr := f.Close(); closeErr != nil {
			t.Error(closeErr)
		}
		if err != nil {
			t.Fatal(err)
		}
		l, s := imageLightness(im)
		t.Logf("stoney_w %s: L* %.1f, sigma %.1f; painted L* %.1f, sigma %.1f", density, l, s, mL, mS)
		if math.Abs(l-mL) > 3 || s < .8*mS {
			t.Errorf("stoney_w %s: L* %.1f, sigma %.1f, want within 3 of %.1f and at least %.1f", density, l, s, mL, .8*mS)
		}
	}
}

// imageLightness is the mean and standard deviation of CIE L* over im's opaque pixels.
func imageLightness(im image.Image) (mean, sigma float64) {
	lin := func(v uint32) float64 {
		x := float64(v) / 65535
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	var n, sum, sq float64
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := im.At(x, y).RGBA()
			if a < 0xff00 {
				continue
			}
			l := 116*math.Cbrt(.2126*lin(r)+.7152*lin(g)+.0722*lin(bl)) - 16
			n, sum, sq = n+1, sum+l, sq+l*l
		}
	}
	mean = sum / n
	return mean, math.Sqrt(sq/n - mean*mean)
}
