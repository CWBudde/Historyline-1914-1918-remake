package artreview

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Independent evaluation of sinc(t)*sinc(t/3), t=(sx-center)/(source/dest),
// gives normalized impulse responses (51.311586..., 29.384980...) for 8→2.
// A tent gives (46,27); fixed-width Lanczos misses these widened samples.
func TestLanczosReductionUsesWidenedRadiusThree(t *testing.T) {
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
	got, err := resizeLanczos(src, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	for x, r := range []uint8{51, 29} {
		if p := got.NRGBAAt(x, 0); p != (color.NRGBA{R: r, A: 255}) {
			t.Errorf("impulse pixel %d = %+v; want red %d alpha 255", x, p, r)
		}
	}
}

func TestLanczosStepIsSharperThanBilinearAndClampsRinging(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 16, 1))
	for x := 0; x < 16; x++ {
		p := color.NRGBA{A: 255}
		if x >= 8 {
			p.R, p.G, p.B = 255, 255, 255
		}
		src.SetNRGBA(x, 0, p)
	}
	got, err := resizeLanczos(src, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The unclamped response includes -3.904007... and 258.904007....
	for x, v := range []uint8{0, 1, 0, 14, 241, 255, 254, 255} {
		if p := got.NRGBAAt(x, 0); p != (color.NRGBA{R: v, G: v, B: v, A: 255}) {
			t.Errorf("step pixel %d = %+v; want gray %d alpha 255", x, p, v)
		}
	}
	tent, err := resizeBilinear(src, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.NRGBAAt(3, 0).R >= tent.NRGBAAt(3, 0).R || got.NRGBAAt(4, 0).R <= tent.NRGBAAt(4, 0).R {
		t.Fatal("Lanczos did not preserve a steeper step than bilinear")
	}
}

func TestLanczosFlatColorOddSizesAndIdentity(t *testing.T) {
	for _, p := range []color.NRGBA{{R: 81, G: 139, B: 203, A: 255}, {R: 81, G: 139, B: 203, A: 37}, {R: 255, G: 128, B: 90}} {
		src := image.NewNRGBA(image.Rect(2, 3, 15, 12))
		for y := 3; y < 12; y++ {
			for x := 2; x < 15; x++ {
				src.SetNRGBA(x, y, p)
			}
		}
		for _, size := range [][2]int{{13, 9}, {7, 5}, {3, 2}, {1, 1}} {
			got, err := resizeLanczos(src, size[0], size[1])
			if err != nil {
				t.Fatal(err)
			}
			want := p
			if p.A == 0 {
				want = color.NRGBA{}
			}
			for y := 0; y < size[1]; y++ {
				for x := 0; x < size[0]; x++ {
					if q := got.NRGBAAt(x, y); q != want {
						t.Fatalf("flat %+v size %v pixel (%d,%d) = %+v", p, size, x, y, q)
					}
				}
			}
		}
	}
}

func TestLanczosTransparencyClampsAlphaWithoutColorFringes(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 16, 1))
	for x := 0; x < 16; x++ {
		p := color.NRGBA{B: 255} // Hidden blue must never reach visible RGB.
		if x >= 8 {
			p = color.NRGBA{R: 81, G: 139, B: 203, A: 255}
		}
		src.SetNRGBA(x, 0, p)
	}
	got, err := resizeLanczos(src, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	for x, a := range []uint8{0, 1, 0, 14, 241, 255, 254, 255} {
		want := color.NRGBA{}
		if a != 0 {
			want = color.NRGBA{R: 81, G: 139, B: 203, A: a}
		}
		if p := got.NRGBAAt(x, 0); p != want {
			t.Errorf("alpha edge pixel %d = %+v; want %+v", x, p, want)
		}
	}
}

func TestLanczosRejectsInvalidSizes(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for _, dims := range [][2]int{{0, 1}, {1, -1}, {4, 2}, {3, 3}} {
		if _, err := resizeLanczos(src, dims[0], dims[1]); err == nil {
			t.Errorf("accepted dimensions %v", dims)
		}
	}
}

func TestLanczosFractionalReductionClampsSignedColorLobes(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	src.SetNRGBA(2, 0, color.NRGBA{B: 255, A: 255})
	got, err := resizeLanczos(src, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The left normalized weights are (0.653840436, 0.418368556,
	// -0.072208992); the right weights mirror them. Clamp negative blue/red.
	for x, want := range []color.NRGBA{{R: 167, G: 107, A: 255}, {G: 107, B: 167, A: 255}} {
		if p := got.NRGBAAt(x, 0); p != want {
			t.Errorf("fractional pixel %d = %+v; want %+v", x, p, want)
		}
	}
}

func TestLanczosTwoDimensionalAlphaImpulse(t *testing.T) {
	for _, a := range []uint8{255, 1} {
		src := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				src.SetNRGBA(x, y, color.NRGBA{B: 255})
			}
		}
		src.SetNRGBA(3, 3, color.NRGBA{R: 81, G: 139, B: 203, A: a})
		got, err := resizeLanczos(src, 2, 2)
		if err != nil {
			t.Fatal(err)
		}
		// Products of the independently evaluated 1D impulse weights give
		// raw alpha (10.3250152, 5.9129018, 5.9129018, 3.3861846).
		for i, alpha := range []uint8{10, 6, 6, 3} {
			want := color.NRGBA{R: 81, G: 139, B: 203, A: alpha}
			if a == 1 {
				want = color.NRGBA{} // Positive raw alpha rounds to zero.
			}
			if p := got.NRGBAAt(i%2, i/2); p != want {
				t.Errorf("source alpha %d pixel %d = %+v; want %+v", a, i, p, want)
			}
		}
	}
}

func TestLanczosRoundsHalfUpAtFinalOutput(t *testing.T) {
	for _, tt := range []struct {
		left, right, want color.NRGBA
	}{
		{color.NRGBA{A: 255}, color.NRGBA{R: 1, G: 255, A: 255}, color.NRGBA{R: 1, G: 128, A: 255}},
		{color.NRGBA{B: 255}, color.NRGBA{R: 81, G: 139, B: 203, A: 1}, color.NRGBA{R: 81, G: 139, B: 203, A: 1}},
	} {
		src := image.NewNRGBA(image.Rect(0, 0, 2, 1))
		src.SetNRGBA(0, 0, tt.left)
		src.SetNRGBA(1, 0, tt.right)
		got, err := resizeLanczos(src, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if p := got.NRGBAAt(0, 0); p != tt.want {
			t.Errorf("half-up pixel = %+v; want %+v", p, tt.want)
		}
	}
}

func TestCLILanczosFullChainAndIndependentFootprints(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{A: 128})
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
	var firstArtifacts map[string][]byte
	for _, repeat := range []string{"first", "repeat"} {
		out := filepath.Join(dir, repeat)
		if err := run([]string{"-reference", rp, "-candidate", cp, "-out", out, "-resampling", "lanczos"}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		// Independently computed 16→8 impulse after clamp/round:
		// [0,4,0,114,35,0,1,0]; its 8→4 pixel 1 is 55.49686...,
		// its 8→2 pixel 0 is 27.97360.... Masking 4x first changes both.
		for _, tt := range []struct {
			name string
			x    int
			want color.NRGBA
		}{
			{"candidate-4x.png", 3, color.NRGBA{R: 114, A: 128}},
			{"candidate-2x.png", 1, color.NRGBA{R: 55, A: 128}},
			{"candidate-2x.png", 2, color.NRGBA{}},
			{"downsample.png", 0, color.NRGBA{R: 28, A: 128}},
			{"downsample.png", 1, color.NRGBA{}},
		} {
			got, err := loadPNG(filepath.Join(out, tt.name))
			if err != nil {
				t.Fatal(err)
			}
			if p := got.NRGBAAt(tt.x, 0); p != tt.want {
				t.Errorf("%s pixel %d = %+v; want %+v", tt.name, tt.x, p, tt.want)
			}
		}
		artifacts := make(map[string][]byte)
		for _, name := range []string{"candidate-4x.png", "candidate-2x.png", "downsample.png", "comparison.png", "diagnostics.json"} {
			data, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			artifacts[name] = data
		}
		if repeat == "first" {
			firstArtifacts = artifacts
		} else if !reflect.DeepEqual(artifacts, firstArtifacts) {
			t.Error("repeated Lanczos run changed artifact bytes")
		}
	}
	var report map[string]any
	if err := json.Unmarshal(firstArtifacts["diagnostics.json"], &report); err != nil {
		t.Fatal(err)
	}
	wantChain := []any{"source -> raw4x: lanczos", "raw4x -> raw2x: lanczos", "raw4x -> rawNative: lanczos", "raw4x/raw2x/rawNative -> outputs: apply reference footprint independently"}
	if report["resampling"] != "lanczos" || !reflect.DeepEqual(report["resampling_chain"], wantChain) {
		t.Errorf("wrong resampling chain: %s", firstArtifacts["diagnostics.json"])
	}
	algorithm, _ := report["resampling_algorithm"].(string)
	for _, term := range []string{"Lanczos3", "radius 3", "source/destination", "premultiplied", "clamp", "half up"} {
		if !strings.Contains(algorithm, term) {
			t.Errorf("algorithm description missing %q: %s", term, algorithm)
		}
	}
	if report["unmasked_native_coverage"].(map[string]any)["extra_visible_pixels"] != float64(1) {
		t.Errorf("unmasked coverage lost: %s", firstArtifacts["diagnostics.json"])
	}
}
