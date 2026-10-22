package main

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

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/lookmeasure"
)

func syntheticWoodland() (*image.NRGBA, woodlandSourceAnnotation, []lookmeasure.Segment) {
	im := image.NewNRGBA(image.Rect(0, 0, 40, 30))
	for y := range 30 {
		for x := range 40 {
			im.SetNRGBA(x, y, color.NRGBA{50, 70, 20, 255})
		}
	}
	a := woodlandSourceAnnotation{ID: "render", Width: 40, Height: 30, PixelsSHA256: fmt.Sprintf("%x", sha256.Sum256(im.Pix)), Scope: "two incomplete crown arcs", Fragments: []woodlandFragment{{ID: "near", Points: [][2]float64{{5, 5}, {15, 5}}, Uncertainty: 2}, {ID: "far", Points: [][2]float64{{5, 15}, {35, 15}}, Uncertainty: 3}}}
	edges := []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: 0}, To: lookmeasure.Point{X: 40, Y: 0}}}
	return im, a, edges
}

func TestWoodlandDistancesWeightVisibleArcLength(t *testing.T) {
	im, a, edges := syntheticWoodland()
	got, err := measureAnnotatedWoodland(a, im, edges)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Available || got.Complete || len(got.Fragments) != 2 {
		t.Fatalf("false coverage: %+v", got)
	}
	s := got.Statistics
	if s.Length != 40 || s.Median == nil || *s.Median != 15 || s.WithinSix == nil || math.Abs(*s.WithinSix-.25) > 1e-12 {
		t.Fatalf("must weight 10px near vs30px far, not average fragments: %+v", s)
	}
	if s.WithinSixLower == nil || *s.WithinSixLower != 0 || s.WithinSixUpper == nil || *s.WithinSixUpper != .25 {
		t.Fatalf("distance uncertainty not propagated: %+v", s)
	}
}

func TestWoodlandRejectsStalePixelsWithoutNumbers(t *testing.T) {
	im, a, e := syntheticWoodland()
	im.SetNRGBA(0, 0, color.NRGBA{1, 2, 3, 255})
	got, err := measureAnnotatedWoodland(a, im, e)
	if err != nil || got.Available || got.Reason != "annotation_source_mismatch" || got.Statistics.Median != nil || len(got.Fragments) != 0 {
		t.Fatalf("stale source measured: %+v %v", got, err)
	}
}

func TestWoodlandDoesNotBridgeFragments(t *testing.T) {
	im, a, e := syntheticWoodland()
	got, err := measureAnnotatedWoodland(a, im, e)
	if err != nil {
		t.Fatal(err)
	}
	if got.Statistics.Length != 40 || got.Statistics.Samples != 40 {
		t.Fatalf("gap between arcs counted: %+v", got.Statistics)
	}
}

func TestWoodlandRejectsInvalidAnnotationsAndNonopaqueSamples(t *testing.T) {
	tests := []struct {
		name   string
		change func(*image.NRGBA, *woodlandSourceAnnotation)
	}{
		{"duplicate fragment", func(_ *image.NRGBA, a *woodlandSourceAnnotation) { a.Fragments = append(a.Fragments, a.Fragments[0]) }},
		{"outside source", func(_ *image.NRGBA, a *woodlandSourceAnnotation) { a.Fragments[0].Points[0][0] = -1 }},
		{"nonfinite", func(_ *image.NRGBA, a *woodlandSourceAnnotation) { a.Fragments[0].Points[0][0] = math.NaN() }},
		{"zero uncertainty", func(_ *image.NRGBA, a *woodlandSourceAnnotation) { a.Fragments[0].Uncertainty = 0 }},
		{"nonopaque", func(im *image.NRGBA, a *woodlandSourceAnnotation) {
			im.SetNRGBA(6, 5, color.NRGBA{0, 0, 0, 0})
			a.PixelsSHA256 = fmt.Sprintf("%x", sha256.Sum256(im.Pix))
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			im, a, e := syntheticWoodland()
			tc.change(im, &a)
			if _, err := measureAnnotatedWoodland(a, im, e); err == nil {
				t.Fatal("bad outline measured")
			}
		})
	}
}

func TestWoodlandHexEdgesUseActualTilePolygon(t *testing.T) {
	p := &project{level: &level.Level{Lines: [][]level.Hex{{{Terrain: "grass"}}, {{Terrain: "forest"}}}}, excluded: map[int]bool{}}
	e := p.woodlandHexEdges()
	if len(e) != 12 {
		t.Fatalf("got %d edges, want actual 2 hexes", len(e))
	}
	if e[0].From != (lookmeasure.Point{X: 25, Y: 0}) || e[0].To != (lookmeasure.Point{X: 71, Y: 0}) || e[6].From != (lookmeasure.Point{X: 96, Y: 40}) {
		t.Fatalf("wrong polygon or odd-line shift: %+v", e)
	}
	p.excluded[2] = true
	if len(p.woodlandHexEdges()) != 6 {
		t.Fatal("excluded row included")
	}
}

func TestWoodlandLoaderPinsAndPreview(t *testing.T) {
	o := comparisonOptions(t)
	p, err := load(o.root, filepath.Join(o.root, o.ref))
	if err != nil {
		t.Fatal(err)
	}
	im, source, _ := syntheticWoodland()
	got, err := p.measureWoodland("render", im, "")
	if err != nil || got.Available || got.Reason != "independent_woodland_annotations_unavailable" {
		t.Fatalf("missing annotations: %+v %v", got, err)
	}
	lv, _ := p.level.Marshal()
	reg, _ := json.Marshal(p.ref.Registration)
	a := woodlandAnnotations{Schema: 1, LevelSHA256: fmt.Sprintf("%x", sha256.Sum256(lv)), RegistrationSHA256: fmt.Sprintf("%x", sha256.Sum256(reg)), Sources: []woodlandSourceAnnotation{source}}
	path := filepath.Join(filepath.Dir(p.targetsPath), "woodland_annotations.json")
	write := func(value woodlandAnnotations, tail string) {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		raw = append(raw, []byte(tail)...)
		if err := os.WriteFile(path, raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(a, "")
	out := t.TempDir()
	got, err = p.measureWoodland("render", im, out)
	if err != nil || !got.Available || got.Complete || len(got.AnnotationSHA256) != 64 || got.Preview == "" {
		t.Fatalf("loader: %+v %v", got, err)
	}
	data, err := os.ReadFile(filepath.Join(out, got.Preview))
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(out, got.Preview))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := png.Decode(f)
	_ = f.Close()
	if err != nil || preview.Bounds().Empty() || fmt.Sprintf("%x", sha256.Sum256(data)) != got.PreviewSHA256 {
		t.Fatalf("preview provenance: %v", err)
	}
	got, err = p.measureWoodland("reference", im, "")
	if err != nil || got.Available || got.Statistics.Median != nil {
		t.Fatalf("render substituted for reference: %+v %v", got, err)
	}
	for _, pin := range []string{"level", "registration"} {
		t.Run(pin, func(t *testing.T) {
			bad := a
			if pin == "level" {
				bad.LevelSHA256 = "stale"
			} else {
				bad.RegistrationSHA256 = "stale"
			}
			write(bad, "")
			got, err := p.measureWoodland("render", im, "")
			if err != nil || got.Available || got.Reason != "annotation_registration_or_level_mismatch" || got.Statistics.Median != nil {
				t.Fatalf("stale geometry measured: %+v %v", got, err)
			}
		})
	}
	for _, tail := range []string{" unexpected", " {}"} {
		write(a, tail)
		if _, err := p.measureWoodland("render", im, ""); err == nil {
			t.Fatal("accepted trailing JSON data")
		}
	}
	a.Sources = append(a.Sources, source)
	write(a, "")
	if _, err := p.measureWoodland("render", im, ""); err == nil {
		t.Fatal("duplicate source accepted")
	}
}
