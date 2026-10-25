package main

import (
	"bytes"
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

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

func TestExportMaterial(t *testing.T) {
	src := image.NewNRGBA(image.Rect(3, 7, 13, 15))
	for y := 7; y < 15; y++ {
		for x := 3; x < 13; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x * 10), uint8(y * 5), 80, 255})
		}
	}
	got, c, err := exportMaterial(src, image.Pt(6, 4), 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds() != image.Rect(0, 0, 12, 8) {
		t.Fatal(got.Bounds())
	}
	if math.Abs(c.Top-2.0/3) > 1e-12 || c.Left != 0 {
		t.Fatalf("crop %+v", c)
	}
	if _, err := worldmaterial.New(got, got, image.Pt(6, 4), 2); err != nil {
		t.Fatal(err)
	}
}

func writeNoisePNG(t *testing.T, path string, w, h int, base color.NRGBA, seed int) string {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			d := uint8((x*7 + y*13 + seed) % 23)
			im.SetNRGBA(x, y, color.NRGBA{base.R + d, base.G + d, base.B + d, 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
}

func TestRunWithDetail(t *testing.T) {
	dir := t.TempDir()
	tone, det, scrub := filepath.Join(dir, "tone.png"), filepath.Join(dir, "detail.png"), filepath.Join(dir, "scrub.png")
	p := profile{
		Schema: 2, Period: [2]int{568, 400}, Densities: []int{2, 4}, Band: 4,
		Materials: []material{
			{
				ID: worldmaterial.MeadowID, Art: "terrain/grass", Source: tone, SHA: writeNoisePNG(t, tone, 284, 200, color.NRGBA{90, 97, 30, 255}, 1),
				Detail: &detail{Source: det, SHA: writeNoisePNG(t, det, 213, 150, color.NRGBA{90, 97, 30, 255}, 5), Repeats: 3, Gain: 0.5},
			},
			{ID: worldmaterial.ScrubID, Art: "terrain/hills", Source: scrub, SHA: writeNoisePNG(t, scrub, 284, 200, color.NRGBA{80, 88, 26, 255}, 2)},
		},
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	path, out := filepath.Join(dir, "profile.json"), filepath.Join(dir, "out")
	if err = os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if err = run(path, out, false); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "materials.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Outputs []output `json:"outputs"`
		Checks  []checks `json:"checks"`
	}
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Outputs) != 4 || report.Outputs[0].Detail == nil || report.Outputs[0].Detail.Repeats != 3 || report.Outputs[0].Detail.ToneContrast != 1 || report.Outputs[2].Detail != nil {
		t.Fatalf("outputs %+v", report.Outputs)
	}
	if len(report.Checks) != 1 || report.Checks[0].SuggestedGain == nil || len(report.Checks[0].Tone) == 0 || len(report.Checks[0].Detail) == 0 || len(report.Checks[0].Result) == 0 {
		t.Fatalf("checks %+v", report.Checks)
	}
	// Every export loads as a material.
	for _, d := range []int{2, 4} {
		var ims [2]image.Image
		for i, o := range report.Outputs {
			if o.Density != d {
				continue
			}
			f, err := os.Open(filepath.Join(out, filepath.FromSlash(o.Path)))
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(f)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				t.Fatal(err)
			}
			ims[i/2] = im
		}
		if _, err := worldmaterial.New(ims[0], ims[1], image.Pt(568, 400), d); err != nil {
			t.Fatal(d, err)
		}
	}
	// Synthetic stripes are no meadow: the strict run fails after writing the report.
	if err = run(path, out, true); err == nil || report.Checks[0].Pass {
		t.Fatal("strict run passed")
	}
	for _, k := range []float64{-1, 0.1, 4, math.NaN()} {
		p.Materials[0].Detail.ToneContrast = k
		if validate(p) == nil {
			t.Fatal("tone contrast", k)
		}
	}
	p.Materials[0].Detail.ToneContrast = 1.3
	if validate(p) != nil {
		t.Fatal("tone contrast 1.3 rejected")
	}
	p.Materials[0].Detail.ToneOffset = [3]float64{0, 0, 25}
	if validate(p) == nil {
		t.Fatal("tone offset 25")
	}
	p.Materials[0].Detail.ToneOffset = [3]float64{0, 1, -4}
	if validate(p) != nil {
		t.Fatal("tone offset rejected")
	}
	p.Schema = 1
	if validate(p) == nil {
		t.Fatal("detail in schema 1")
	}
	p.Schema, p.Materials[1].Detail = 2, p.Materials[0].Detail
	if validate(p) == nil {
		t.Fatal("detail on the scrub")
	}
}
