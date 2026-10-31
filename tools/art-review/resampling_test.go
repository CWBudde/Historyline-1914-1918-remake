package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// A widened tent samples beyond the box footprint, including at boundaries.
// For 8→2 the first sample has weights (5,7,7,5,3,1)/28.
func TestBilinearReductionUsesAntialiasedTent(t *testing.T) {
	src := image.NewNRGBA(image.Rect(4, 7, 12, 11))
	for y := 7; y < 11; y++ {
		for x := 4; x < 12; x++ {
			p := color.NRGBA{A: 255}
			if x == 7 {
				p.R = 255
			}
			src.SetNRGBA(x, y, p)
		}
	}
	got, err := resizeBilinear(src, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	for x, want := range []color.NRGBA{{R: 46, A: 255}, {R: 27, A: 255}} {
		if p := got.NRGBAAt(x, 0); p != want {
			t.Errorf("pixel %d = %+v; want %+v", x, p, want)
		}
	}
}

func TestBilinearReductionIgnoresHiddenRGB(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.SetNRGBA(x, y, color.NRGBA{B: 255})
		}
	}
	src.SetNRGBA(3, 3, color.NRGBA{R: 255, A: 255})
	got, err := resizeBilinear(src, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range []uint8{8, 5, 5, 3} {
		if p := got.NRGBAAt(i%2, i/2); p != (color.NRGBA{R: 255, A: a}) {
			t.Errorf("pixel %d: hidden blue contaminated red: %+v", i, p)
		}
	}
	src.SetNRGBA(3, 3, color.NRGBA{R: 255})
	got, err = resizeBilinear(src, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Pix, make([]byte, len(got.Pix))) {
		t.Fatal("transparent output retains hidden RGB")
	}
}

func TestBilinearRejectsInvalidSizes(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for _, dims := range [][2]int{{0, 1}, {1, -1}, {4, 2}, {3, 3}} {
		if _, err := resizeBilinear(src, dims[0], dims[1]); err == nil {
			t.Errorf("accepted dimensions %v", dims)
		}
	}
}

func TestCLIBilinearDerivesUnmaskedLevelsDirectlyFrom4x(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{A: 128})
	src := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			p := color.NRGBA{A: 255}
			if x == 3 {
				p.R = 255
			}
			src.SetNRGBA(x, y, p)
		}
	}
	rp, cp := writeFixture(t, dir, "ref.png", ref), writeFixture(t, dir, "src.png", src)
	out := filepath.Join(dir, "review")
	if err := run([]string{"-reference", rp, "-candidate", cp, "-out", out, "-resampling", "bilinear"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		x, y int
		want color.NRGBA
	}{
		{"candidate-4x.png", 3, 0, color.NRGBA{R: 255, A: 128}},
		{"candidate-2x.png", 1, 0, color.NRGBA{R: 96, A: 128}},
		{"candidate-2x.png", 2, 0, color.NRGBA{}},
		{"downsample.png", 0, 0, color.NRGBA{R: 46, A: 128}},
		{"downsample.png", 1, 0, color.NRGBA{}},
	} {
		got, err := loadPNG(filepath.Join(out, tt.name))
		if err != nil {
			t.Fatal(err)
		}
		if p := got.NRGBAAt(tt.x, tt.y); p != tt.want {
			t.Errorf("%s (%d,%d)=%+v; want %+v", tt.name, tt.x, tt.y, p, tt.want)
		}
	}
	data, err := os.ReadFile(filepath.Join(out, "diagnostics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report["resampling"] != "bilinear" {
		t.Fatalf("missing method: %s", data)
	}
	wantChain := []any{"source -> raw4x: bilinear", "raw4x -> raw2x: bilinear", "raw4x -> rawNative: bilinear", "raw4x/raw2x/rawNative -> outputs: apply reference footprint independently"}
	chain, ok := report["resampling_chain"].([]any)
	if !ok || len(chain) != len(wantChain) {
		t.Fatalf("missing chain: %s", data)
	}
	for i, want := range wantChain {
		if chain[i] != want {
			t.Errorf("chain[%d]=%v; want %v", i, chain[i], want)
		}
	}
	if report["unmasked_native_coverage"].(map[string]any)["extra_visible_pixels"] != float64(1) {
		t.Fatalf("unmasked coverage lost: %s", data)
	}
}

func TestCLIResamplingRejectsUnknownMethodAndNeighborhoodOverride(t *testing.T) {
	for _, method := range []string{"", "nearest", "Bilinear"} {
		err := run([]string{"-reference", "unused", "-candidate", "unused", "-out", filepath.Join(t.TempDir(), "review"), "-resampling", method}, &bytes.Buffer{})
		if err == nil || !bytes.Contains([]byte(err.Error()), []byte("-resampling")) {
			t.Fatalf("method %q: %v", method, err)
		}
	}
	for _, method := range []string{"area", "bilinear", "lanczos"} {
		dir, config, _ := neighborhoodFixture(t, false)
		out := filepath.Join(dir, "review")
		err := run([]string{"-neighborhood", config, "-out", out, "-resampling", method}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("neighborhood accepted explicit %s", method)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatal("created output before rejection")
		}
	}
}

// Golden hashes were captured before adding each option. They catch rounding,
// mask-order and PNG changes in area and bilinear mode.
func TestCLIDefaultAreaPreservesExistingPNGBytes(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for i, a := range []uint8{0, 128, 255, 255, 50, 0} {
		ref.SetNRGBA(i%3, i/3, color.NRGBA{R: uint8(30 * i), G: 60, A: a})
	}
	src := image.NewNRGBA(image.Rect(0, 0, 15, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 15; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(17 * x), G: uint8(25 * y), B: uint8(11 * (x + y)), A: uint8((x*31 + y*19) % 256)})
		}
	}
	rp, cp := writeFixture(t, dir, "ref.png", ref), writeFixture(t, dir, "src.png", src)
	for _, mode := range []string{"default", "area", "bilinear"} {
		out := filepath.Join(dir, mode)
		args := []string{"-reference", rp, "-candidate", cp, "-out", out}
		if mode != "default" {
			args = append(args, "-resampling", mode)
		}
		if err := run(args, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		for name, want := range map[string]string{
			"candidate-4x.png": "2595a97fab963b06dfb8dc4a652349d48b77b785fe286e6d78b8867469afafc2",
			"candidate-2x.png": "6e297053b8e4f21b4bcc8570d6a0b59b3072060b0da5f42f831dbc09bcc46dea",
			"downsample.png":   "ac0bfbeea13a787334e41b73a7e6640c1deb15a0a07eb7f1fd89628b33660f9e",
			"comparison.png":   "e7946902434a0d2f7f2ed60380af65890c33369a2776289ff7fe1abc862ce4b4",
		} {
			data, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data)
			got := hex.EncodeToString(sum[:])
			if mode == "bilinear" {
				want = map[string]string{
					"candidate-4x.png": "04fac0e5370dbfa11da16165454c7799803279e7cc06b497980ca643ba16beb4",
					"candidate-2x.png": "34a54079717700a88556bf0ba8040e49bbad449eb5e2b07057a26023978bc9d3",
					"downsample.png":   "1f1e31ee7949e08b4abd3c4aa8839930226ac295ffc0cd518bacc484e6aef831",
					"comparison.png":   "ed9e5abb3099e1ce0e23ae07fd52766c06ef9d09a4fac1b1919cde8015a36a3b",
				}[name]
			}
			if got != want {
				t.Errorf("%s/%s sha256 = %s; want %s", mode, name, got, want)
			}
		}
	}
}

// The source is 8x native: this catches accidentally using area for source→4x,
// or taking native from the source or the already reduced 2x output.
func TestCLIBilinearReducesSourceTo4xBeforeNative(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{A: 255})
	src := image.NewNRGBA(image.Rect(0, 0, 16, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 16; x++ {
			p := color.NRGBA{A: 255}
			if x == 7 {
				p.R = 255
			}
			src.SetNRGBA(x, y, p)
		}
	}
	rp, cp := writeFixture(t, dir, "ref.png", ref), writeFixture(t, dir, "src.png", src)
	out := filepath.Join(dir, "review")
	if err := run([]string{"-reference", rp, "-candidate", cp, "-out", out, "-resampling", "bilinear"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		x    int
		r    uint8
	}{
		{"candidate-4x.png", 3, 96},
		{"candidate-2x.png", 1, 40},
		{"downsample.png", 0, 21},
	} {
		got, err := loadPNG(filepath.Join(out, tt.name))
		if err != nil {
			t.Fatal(err)
		}
		if p := got.NRGBAAt(tt.x, 0); p != (color.NRGBA{R: tt.r, A: 255}) {
			t.Errorf("%s pixel %d = %+v; want red %d, alpha 255", tt.name, tt.x, p, tt.r)
		}
	}
}

func TestBilinearFractionalReductionWeights(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	src.SetNRGBA(2, 0, color.NRGBA{B: 255, A: 255})
	got, err := resizeBilinear(src, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The left sample has normalized weights (5,3,0)/8; right is mirrored.
	for x, want := range []color.NRGBA{{R: 159, G: 96, A: 255}, {G: 96, B: 159, A: 255}} {
		if p := got.NRGBAAt(x, 0); p != want {
			t.Errorf("pixel %d = %+v; want %+v", x, p, want)
		}
	}
}
