package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBoxDownsampleAlphaWeighted(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{B: 255})
	src.SetNRGBA(0, 1, color.NRGBA{B: 255})
	src.SetNRGBA(1, 1, color.NRGBA{B: 255})
	got, err := boxDownsample(src, 2)
	if err != nil {
		t.Fatal(err)
	}
	if pixel := got.NRGBAAt(0, 0); pixel != (color.NRGBA{R: 255, A: 64}) {
		t.Fatalf("transparent blue contaminated red: %+v", pixel)
	}
}

func TestAreaResizeWeightsAndBounds(t *testing.T) {
	src := image.NewNRGBA(image.Rect(4, 7, 7, 8))
	src.SetNRGBA(4, 7, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(5, 7, color.NRGBA{G: 255, A: 255})
	src.SetNRGBA(6, 7, color.NRGBA{B: 255, A: 255})
	got, err := resizeArea(src, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []color.NRGBA{{R: 170, G: 85, A: 255}, {G: 85, B: 170, A: 255}}
	for x, p := range want {
		if got.NRGBAAt(x, 0) != p {
			t.Errorf("pixel %d = %+v; want %+v", x, got.NRGBAAt(x, 0), p)
		}
	}
}

func TestResamplingRejectsInvalidSizes(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for _, factor := range []int{0, -1, 2, 4} {
		if _, err := boxDownsample(src, factor); err == nil {
			t.Errorf("factor %d accepted", factor)
		}
	}
	for _, dims := range [][2]int{{0, 1}, {1, -1}, {4, 2}, {3, 3}} {
		if _, err := resizeArea(src, dims[0], dims[1]); err == nil {
			t.Errorf("dimensions %v accepted", dims)
		}
	}
}

func TestFootprintPreservesHolesAndClipsOutside(t *testing.T) {
	ref := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{A: 255})
	ref.SetNRGBA(1, 0, color.NRGBA{A: 128})
	candidate := image.NewNRGBA(ref.Bounds())
	candidate.SetNRGBA(1, 0, color.NRGBA{R: 200, A: 255})
	candidate.SetNRGBA(2, 0, color.NRGBA{B: 255, A: 255})
	got := applyFootprint(candidate, ref)
	if got.NRGBAAt(0, 0).A != 0 {
		t.Fatal("candidate hole was filled")
	}
	if got.NRGBAAt(1, 0) != (color.NRGBA{R: 200, A: 128}) {
		t.Fatal("reference boundary alpha was not capped")
	}
	if got.NRGBAAt(2, 0) != (color.NRGBA{}) {
		t.Fatal("color outside logical footprint survived")
	}
	if candidate.NRGBAAt(2, 0).A != 255 {
		t.Fatal("input was modified")
	}
}

func TestSevenHexUsesGameOffsetsInRasterOrder(t *testing.T) {
	tile := image.NewNRGBA(image.Rect(0, 0, 96, 80))
	tile.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	patch := sevenHex(tile)
	if patch.Bounds() != image.Rect(0, 0, 238, 240) {
		t.Fatalf("patch bounds %v", patch.Bounds())
	}
	for _, p := range []image.Point{{71, 0}, {0, 40}, {142, 40}, {71, 80}, {0, 120}, {142, 120}, {71, 160}} {
		if patch.NRGBAAt(p.X, p.Y).R != 255 {
			t.Errorf("missing game tile origin %v", p)
		}
	}
	// Later rows cover earlier rows where full rectangular fixtures overlap.
	for y := 0; y < 80; y++ {
		for x := 0; x < 96; x++ {
			tile.SetNRGBA(x, y, color.NRGBA{R: uint8(y), A: 255})
		}
	}
	patch = sevenHex(tile)
	if got := patch.NRGBAAt(71, 80).R; got != 0 {
		t.Errorf("later center layer did not cover upper tile: %d", got)
	}
}

func TestDiagnosticsSeparateCoverageFromRGB(t *testing.T) {
	ref := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	ref.SetNRGBA(1, 0, color.NRGBA{R: 255, A: 255})
	cand := image.NewNRGBA(ref.Bounds())
	cand.SetNRGBA(0, 0, color.NRGBA{B: 255, A: 255})
	cand.SetNRGBA(2, 0, color.NRGBA{G: 255, A: 255})
	got := measure(ref, cand)
	if got.MissingPixels != 1 || got.ExtraPixels != 1 || got.SharedPixels != 1 {
		t.Fatalf("coverage %+v", got)
	}
	if got.NormalizedRGBDifference == nil || *got.NormalizedRGBDifference != 2.0/3.0 {
		t.Fatalf("difference %+v", got.NormalizedRGBDifference)
	}
	if measure(ref, image.NewNRGBA(ref.Bounds())).NormalizedRGBDifference != nil {
		t.Fatal("empty intersection has a misleading RGB score")
	}
}

func TestSevenHexCoverageExcludesExposedBoundaryAndReportsHoles(t *testing.T) {
	ref := image.NewNRGBA(image.Rect(0, 0, 96, 80))
	for y := 0; y < 80; y++ {
		inset := 24 - y*24/40
		if y >= 40 {
			inset = (y - 40) * 24 / 40
		}
		for x := inset; x < 96-inset; x++ {
			ref.SetNRGBA(x, y, color.NRGBA{G: 100, A: 255})
		}
	}
	candidate := image.NewNRGBA(ref.Bounds())
	copy(candidate.Pix, ref.Pix)
	candidate.SetNRGBA(48, 40, color.NRGBA{})
	referencePatch, candidatePatch := sevenHex(ref), sevenHex(candidate)
	for _, row := range []struct{ y, start, end int }{{80, 0, 238}, {120, 24, 214}} {
		for x := row.start; x < row.end; x++ {
			if referencePatch.NRGBAAt(x, row.y).A == 0 {
				t.Fatalf("touching ground fixture has an internal gap at %d,%d", x, row.y)
			}
		}
	}
	if referencePatch.NRGBAAt(0, 0).A != 0 {
		t.Fatal("fixture does not expose an outer boundary")
	}
	if got := measure(referencePatch, candidatePatch); got.MissingPixels != 7 || got.ExtraPixels != 0 {
		t.Fatalf("holes or exposed outer boundary counted incorrectly: %+v", got)
	}
}

func TestCLINativeReferenceUsesItsDimensionsWithoutInventingPatchGeometry(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	raw := image.NewNRGBA(image.Rect(0, 0, 96, 96))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			ref.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}
	for y := 0; y < 96; y++ {
		for x := 0; x < 96; x++ {
			raw.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}
	rp, cp := writeFixture(t, dir, "ref.png", ref), writeFixture(t, dir, "raw.png", raw)
	out := filepath.Join(dir, "review")
	if err := run([]string{"-reference", rp, "-candidate", cp, "-reference-platform", "amiga-native", "-out", out}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]image.Point{"candidate-4x.png": {96, 96}, "candidate-2x.png": {48, 48}, "downsample.png": {24, 24}, "comparison.png": {288, 120}} {
		img, err := loadPNG(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Size() != want {
			t.Errorf("%s dimensions %v; want %v", name, img.Bounds().Size(), want)
		}
	}
	for _, name := range []string{"seven-hex-reference.png", "seven-hex-candidate.png"} {
		if _, err := os.Stat(filepath.Join(out, name)); !os.IsNotExist(err) {
			t.Fatalf("invented patch output %s: %v", name, err)
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
	if report["reference_platform"] != "amiga-native" {
		t.Fatal(report["reference_platform"])
	}
	if _, ok := report["seven_hex_coverage"]; ok {
		t.Fatal("native report invents patch coverage")
	}
	if _, ok := report["patch_placement"]; ok {
		t.Fatal("native report invents patch placement")
	}
}

func TestComparisonShowsNativePixelsNearestMagnificationAndAlphaDifference(t *testing.T) {
	ref := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	candidate := image.NewNRGBA(ref.Bounds())
	candidate.SetNRGBA(0, 0, color.NRGBA{B: 255, A: 255})
	candidate.SetNRGBA(1, 0, color.NRGBA{A: 128})
	got := comparison(ref, candidate)
	for _, test := range []struct {
		x, y int
		want color.NRGBA
	}{
		{0, 0, color.NRGBA{R: 255, A: 255}}, {2, 0, color.NRGBA{B: 255, A: 255}},
		{4, 0, color.NRGBA{R: 255, B: 255, A: 255}}, {5, 0, color.NRGBA{R: 128, G: 128, B: 128, A: 255}},
		{0, 1, color.NRGBA{R: 255, A: 255}}, {3, 4, color.NRGBA{R: 255, A: 255}}, {4, 1, color.NRGBA{}},
		{8, 1, color.NRGBA{B: 255, A: 255}}, {11, 4, color.NRGBA{B: 255, A: 255}},
	} {
		if p := got.NRGBAAt(test.x, test.y); p != test.want {
			t.Errorf("pixel %d,%d = %+v; want %+v", test.x, test.y, p, test.want)
		}
	}
}

func TestFootprintRepeatsExactLogicalMaskAtWorkingScales(t *testing.T) {
	ref := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	ref.SetNRGBA(0, 0, color.NRGBA{A: 255})
	for _, scale := range []int{2, 4} {
		src := image.NewNRGBA(image.Rect(0, 0, 2*scale, scale))
		for y := 0; y < scale; y++ {
			for x := 0; x < 2*scale; x++ {
				src.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
			}
		}
		src.SetNRGBA(0, 0, color.NRGBA{})
		got := applyFootprint(src, ref)
		for y := 0; y < scale; y++ {
			for x := 0; x < 2*scale; x++ {
				want := uint8(255)
				if x >= scale || x == 0 && y == 0 {
					want = 0
				}
				if got.NRGBAAt(x, y).A != want {
					t.Errorf("scale %d pixel %d,%d alpha mismatch", scale, x, y)
				}
			}
		}
	}
}

func TestCLIProducesReviewArtifactsWithoutChangingInputs(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 96, 80))
	ref.SetNRGBA(48, 40, color.NRGBA{R: 255, A: 255})
	raw := image.NewNRGBA(image.Rect(0, 0, 480, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 480; x++ {
			raw.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	rp, cp := writeFixture(t, dir, "reference.png", ref), writeFixture(t, dir, "candidate.png", raw)
	original, err := os.ReadFile(cp)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "review")
	if err := run([]string{"-reference", rp, "-candidate", cp, "-out", out}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]image.Point{
		"candidate-4x.png": {384, 320}, "candidate-2x.png": {192, 160}, "downsample.png": {96, 80},
		"comparison.png": {1152, 400}, "seven-hex-reference.png": {238, 240}, "seven-hex-candidate.png": {238, 240},
	} {
		f, err := os.Open(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(f)
		closeErr := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if img.Bounds().Size() != want {
			t.Errorf("%s size %v; want %v", name, img.Bounds().Size(), want)
		}
	}
	data, err := os.ReadFile(filepath.Join(out, "diagnostics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report diagnostics
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.ReferencePlatform != "repo-provisional" || report.RawCoverage.ExtraPixels != 7679 || report.FinalCoverage.MissingPixels != 0 {
		t.Fatalf("report %+v", report)
	}
	if !strings.Contains(strings.ToLower(report.Advisory), "not") {
		t.Fatal("report lacks advisory limitations")
	}
	after, err := os.ReadFile(cp)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Fatal("source input changed")
	}
	// Output folder collisions must fail instead of overwriting a previous review.
	if err := run([]string{"-reference", rp, "-candidate", cp, "-out", out}, &bytes.Buffer{}); err == nil {
		t.Fatal("existing review overwritten")
	}
}

func TestCLIRejectsMalformedInputsAndGeometry(t *testing.T) {
	dir := t.TempDir()
	ref := image.NewNRGBA(image.Rect(0, 0, 96, 80))
	ref.SetNRGBA(0, 0, color.NRGBA{A: 255})
	rp := writeFixture(t, dir, "ref.png", ref)
	valid := writeFixture(t, dir, "valid.png", image.NewNRGBA(image.Rect(0, 0, 384, 320)))
	wrongAspect := writeFixture(t, dir, "square.png", image.NewNRGBA(image.Rect(0, 0, 400, 400)))
	small := writeFixture(t, dir, "small.png", image.NewNRGBA(image.Rect(0, 0, 192, 160)))
	wrongRef := writeFixture(t, dir, "wrong-ref.png", image.NewNRGBA(image.Rect(0, 0, 24, 20)))
	broken := filepath.Join(dir, "broken.png")
	if err := os.WriteFile(broken, []byte("not a png"), 0600); err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"missing": {}, "missing out": {"-reference", rp, "-candidate", valid},
		"malformed": {"-reference", rp, "-candidate", broken}, "aspect": {"-reference", rp, "-candidate", wrongAspect},
		"upscale": {"-reference", rp, "-candidate", small}, "reference dimensions": {"-reference", wrongRef, "-candidate", valid},
		"extra args": {"-reference", rp, "-candidate", valid, "unexpected"},
	} {
		t.Run(name, func(t *testing.T) {
			if name != "missing out" {
				args = append(args, "-out", filepath.Join(dir, strings.ReplaceAll(name, " ", "-")))
			}
			if err := run(args, &bytes.Buffer{}); err == nil {
				t.Fatal("invalid invocation accepted")
			}
		})
	}
}

func TestPNGLoaderSupportsRGBAndRGBA(t *testing.T) {
	for _, src := range []image.Image{image.NewRGBA(image.Rect(0, 0, 2, 2)), image.NewNRGBA(image.Rect(0, 0, 2, 2))} {
		p := writeFixture(t, t.TempDir(), "input.png", src)
		got, err := loadPNG(p)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got.Bounds(), image.Rect(0, 0, 2, 2)) {
			t.Fatal(got.Bounds())
		}
	}
	// All-opaque Go PNG encoding uses PNG RGB rather than RGBA.
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 200, G: 40, A: 255})
		}
	}
	got, err := loadPNG(writeFixture(t, t.TempDir(), "rgb.png", src))
	if err != nil {
		t.Fatal(err)
	}
	if got.NRGBAAt(0, 0) != (color.NRGBA{R: 200, G: 40, A: 255}) {
		t.Fatal(got.NRGBAAt(0, 0))
	}
}

func writeFixture(t *testing.T, dir, name string, img image.Image) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}
