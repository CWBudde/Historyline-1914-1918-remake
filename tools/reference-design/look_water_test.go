package referencedesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func syntheticWater() (*image.NRGBA, waterSourceAnnotation) {
	im := image.NewNRGBA(image.Rect(0, 0, 32, 20))
	for y := range 20 {
		for x := range 32 {
			v := uint8(50)
			if x >= 10 && x < 14 {
				v = 100
			}
			if x >= 16 && x < 20 {
				v = 30
			}
			im.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	a := waterSourceAnnotation{ID: "render", PixelsSHA256: fmt.Sprintf("%x", sha256.Sum256(im.Pix)), Width: 32, Height: 20, Transects: []waterTransect{{ID: "lake-north", Feature: "lake", Kind: "bank", From: [2]float64{2, 10}, To: [2]float64{30, 10}, Uncertainty: 1, Spans: []waterSpan{{Name: "bank", Role: "bank_width", Start: 20, End: 26}}, Regions: []waterRegion{{Role: "interior_water", Start: 2, End: 6, HalfWidth: 1}, {Role: "near_bank_water", Start: 8, End: 12, HalfWidth: 1}, {Role: "bank_shadow", Start: 14, End: 18, HalfWidth: 1}}}}}
	return im, a
}

func TestWaterSamplesActualImageAndNamedRegions(t *testing.T) {
	im, a := syntheticWater()
	got, err := measureAnnotatedWater(a, im)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Available || len(got.Transects) != 1 || len(got.Groups) != 1 {
		t.Fatalf("missing water evidence: %+v", got)
	}
	s := got.Transects[0]
	if len(s.Widths) != 1 || s.Widths[0].Width != 6 || s.Widths[0].Uncertainty != 2 {
		t.Fatalf("width/endpoint uncertainty: %+v", s.Widths)
	}
	if len(s.Regions) != 3 || s.Regions[0].Samples != 8 {
		t.Fatalf("strip sampling: %+v", s.Regions)
	}
	deep, _, _ := toLab(color.NRGBA{50, 50, 50, 255})
	near, _, _ := toLab(color.NRGBA{100, 100, 100, 255})
	shadow, _, _ := toLab(color.NRGBA{30, 30, 30, 255})
	if s.DepthAppearance.Value == nil || math.Abs(*s.DepthAppearance.Value-(near-deep)) > 1e-9 {
		t.Fatalf("depth appearance: %+v", s.DepthAppearance)
	}
	if s.BankShadow.Value == nil || math.Abs(*s.BankShadow.Value-(shadow-near)) > 1e-9 {
		t.Fatalf("bank shadow must compare with near water, not deep: %+v", s.BankShadow)
	}
	if got.Groups[0].Samples != 1 || got.Groups[0].Median == nil || *got.Groups[0].Median != 6 || got.Groups[0].Variation != nil {
		t.Fatalf("single width must not invent variation: %+v", got.Groups)
	}
}

func TestWaterDoesNotReuseStaleSourceAnnotations(t *testing.T) {
	im, a := syntheticWater()
	im.SetNRGBA(0, 0, color.NRGBA{1, 2, 3, 255})
	got, err := measureAnnotatedWater(a, im)
	if err != nil {
		t.Fatal(err)
	}
	if got.Available || got.Reason != "annotation_source_mismatch" || len(got.Transects) != 0 || len(got.Groups) != 0 {
		t.Fatalf("stale annotation yielded data: %+v", got)
	}
}

func TestWaterRejectsInvalidOrAmbiguousAnnotations(t *testing.T) {
	tests := []struct {
		name   string
		change func(*waterSourceAnnotation)
	}{
		{"duplicate transect", func(a *waterSourceAnnotation) { a.Transects = append(a.Transects, a.Transects[0]) }},
		{"negative span", func(a *waterSourceAnnotation) { a.Transects[0].Spans[0].Start = -1 }},
		{"reversed span", func(a *waterSourceAnnotation) { a.Transects[0].Spans[0].End = 19 }},
		{"outside span", func(a *waterSourceAnnotation) { a.Transects[0].Spans[0].End = 29 }},
		{"unknown width role", func(a *waterSourceAnnotation) { a.Transects[0].Spans[0].Role = "guessed" }},
		{"duplicate region", func(a *waterSourceAnnotation) {
			a.Transects[0].Regions = append(a.Transects[0].Regions, a.Transects[0].Regions[0])
		}},
		{"overlapping regions", func(a *waterSourceAnnotation) { a.Transects[0].Regions[1].Start = 4 }},
		{"unknown region", func(a *waterSourceAnnotation) { a.Transects[0].Regions[0].Role = "guessed" }},
		{"zero uncertainty", func(a *waterSourceAnnotation) { a.Transects[0].Uncertainty = 0 }},
		{"bad endpoint", func(a *waterSourceAnnotation) { a.Transects[0].From[0] = math.Inf(1) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			im, a := syntheticWater()
			tc.change(&a)
			if _, err := measureAnnotatedWater(a, im); err == nil {
				t.Fatal("invalid annotation measured")
			}
		})
	}
}

func TestWaterMakesMissingRegionsAndTransparentSamplesExplicit(t *testing.T) {
	im, a := syntheticWater()
	a.Transects[0].Regions = nil
	got, err := measureAnnotatedWater(a, im)
	if err != nil {
		t.Fatal(err)
	}
	if got.Transects[0].DepthAppearance.Value != nil || got.Transects[0].DepthAppearance.Reason == "" || got.Transects[0].BankShadow.Value != nil {
		t.Fatal("fabricated missing regions")
	}
	im, a = syntheticWater()
	im.SetNRGBA(4, 9, color.NRGBA{50, 50, 50, 0})
	a.PixelsSHA256 = fmt.Sprintf("%x", sha256.Sum256(im.Pix))
	got, err = measureAnnotatedWater(a, im)
	if err != nil {
		t.Fatal(err)
	}
	r := got.Transects[0].Regions[0]
	if r.L != nil || r.Reason != "nonopaque_strip" || got.Transects[0].DepthAppearance.Value != nil {
		t.Fatalf("transparent strip measured: %+v", got)
	}
}

func TestWaterVariationUsesPopulationCV(t *testing.T) {
	im, a := syntheticWater()
	a.Transects[0].Spans = append(a.Transects[0].Spans, waterSpan{Name: "other-bank", Role: "bank_width", Start: 8, End: 18})
	got, err := measureAnnotatedWater(a, im)
	if err != nil {
		t.Fatal(err)
	}
	s := got.Groups[0]
	if s.Samples != 2 || s.Median == nil || *s.Median != 8 || s.Variation == nil || math.Abs(*s.Variation-.25) > 1e-12 {
		t.Fatalf("wrong population CV: %+v", s)
	}
}

func TestWaterLoaderPinsGeometryAndWritesVerifiedPreview(t *testing.T) {
	o := comparisonOptions(t)
	p, err := load(o.root, filepath.Join(o.root, o.ref))
	if err != nil {
		t.Fatal(err)
	}
	im, source := syntheticWater()
	path := filepath.Join(filepath.Dir(p.targetsPath), "water_annotations.json")
	got, err := p.measureWater("render", im, "")
	if err != nil || got.Available || got.Reason != "independent_water_annotations_unavailable" {
		t.Fatalf("missing annotations: %+v, %v", got, err)
	}
	lv, _ := p.level.Marshal()
	reg, _ := json.Marshal(p.ref.Registration)
	a := waterAnnotations{Schema: 1, Coordinates: "world pixel centres", LevelSHA256: fmt.Sprintf("%x", sha256.Sum256(lv)), RegistrationSHA256: fmt.Sprintf("%x", sha256.Sum256(reg)), Sources: []waterSourceAnnotation{source}}
	write := func() {
		t.Helper()
		raw, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write()
	out := t.TempDir()
	got, err = p.measureWater("render", im, out)
	if err != nil || !got.Available || len(got.AnnotationSHA256) != 64 || got.Preview == "" {
		t.Fatalf("loader: %+v, %v", got, err)
	}
	encoded, err := os.ReadFile(filepath.Join(out, got.Preview))
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(out, got.Preview))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := png.Decode(f)
	_ = f.Close()
	if err != nil || preview.Bounds().Empty() || fmt.Sprintf("%x", sha256.Sum256(encoded)) != got.PreviewSHA256 {
		t.Fatalf("preview provenance: %v", err)
	}
	for _, pin := range []string{"level", "registration"} {
		t.Run(pin, func(t *testing.T) {
			bad := a
			if pin == "level" {
				bad.LevelSHA256 = "stale"
			} else {
				bad.RegistrationSHA256 = "stale"
			}
			raw, _ := json.Marshal(bad)
			if err := os.WriteFile(path, raw, 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := p.measureWater("render", im, "")
			if err != nil || got.Available || got.Reason != "annotation_registration_or_level_mismatch" || len(got.Transects) > 0 {
				t.Fatalf("stale geometry yielded data: %+v %v", got, err)
			}
		})
	}
	write()
	got, err = p.measureWater("reference", im, "")
	if err != nil || got.Available || got.Reason != "independent_water_annotations_unavailable" {
		t.Fatalf("render annotations substituted for reference: %+v, %v", got, err)
	}
	a.Sources = append(a.Sources, source)
	write()
	if _, err := p.measureWater("render", im, ""); err == nil {
		t.Fatal("duplicate source accepted")
	}
}

func TestWaterLoaderRejectsTrailingData(t *testing.T) {
	o := comparisonOptions(t)
	p, err := load(o.root, filepath.Join(o.root, o.ref))
	if err != nil {
		t.Fatal(err)
	}
	im, source := syntheticWater()
	lv, _ := p.level.Marshal()
	reg, _ := json.Marshal(p.ref.Registration)
	a := waterAnnotations{Schema: 1, LevelSHA256: fmt.Sprintf("%x", sha256.Sum256(lv)), RegistrationSHA256: fmt.Sprintf("%x", sha256.Sum256(reg)), Sources: []waterSourceAnnotation{source}}
	raw, _ := json.Marshal(a)
	path := filepath.Join(filepath.Dir(p.targetsPath), "water_annotations.json")
	for _, tail := range []string{" unexpected", " {}"} {
		if err := os.WriteFile(path, append(append([]byte{}, raw...), []byte(tail)...), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := p.measureWater("render", im, ""); err == nil {
			t.Errorf("accepted trailing data %q", tail)
		}
	}
}
