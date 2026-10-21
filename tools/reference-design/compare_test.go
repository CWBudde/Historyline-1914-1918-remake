package main

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestToLab(t *testing.T) {
	for _, tc := range []struct {
		c       color.NRGBA
		l, a, b float64
	}{
		{color.NRGBA{255, 255, 255, 255}, 100, 0, 0},
		{color.NRGBA{0, 0, 0, 255}, 0, 0, 0},
		{color.NRGBA{255, 0, 0, 255}, 53.24, 80.09, 67.20},
	} {
		l, a, b := toLab(tc.c)
		if math.Abs(l-tc.l) > 0.05 || math.Abs(a-tc.a) > 0.05 || math.Abs(b-tc.b) > 0.05 {
			t.Errorf("toLab(%v) = %.2f %.2f %.2f, want %.2f %.2f %.2f", tc.c, l, a, b, tc.l, tc.a, tc.b)
		}
	}
}

func filled(w, h int, at func(x, y int) color.NRGBA) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			im.SetNRGBA(x, y, at(x, y))
		}
	}
	return im
}

func TestMeasure(t *testing.T) {
	cells := []cell{{2, 1}}
	flat := newLabPlane(filled(400, 200, func(int, int) color.NRGBA { return color.NRGBA{119, 119, 119, 255} }))
	if m := flat.measure(cells); math.Abs(m.L-50) > 0.1 || m.Contrast != 0 || m.Chroma != 0 || m.Edges != 0 {
		t.Errorf("flat grey: %+v, want L 50 and no contrast, chroma or edges", m)
	}
	stripes := newLabPlane(filled(400, 200, func(x, _ int) color.NRGBA {
		if x%4 < 2 {
			return color.NRGBA{A: 255}
		}
		return color.NRGBA{255, 255, 255, 255}
	}))
	if m := stripes.measure(cells); math.Abs(m.L-50) > 1 || math.Abs(m.Contrast-50) > 1 || m.Edges != 1 {
		t.Errorf("stripes: %+v, want L 50, contrast 50, edges 1", m)
	}
	hole := newLabPlane(image.NewNRGBA(image.Rect(0, 0, 400, 200)))
	if m := hole.measure(cells); m != (metrics{}) {
		t.Errorf("transparent: %+v, want zero", m)
	}
}

func TestHalve(t *testing.T) {
	im := filled(4, 2, func(x, _ int) color.NRGBA { return color.NRGBA{uint8(x * 10), 0, 0, 255} })
	im.SetNRGBA(3, 1, color.NRGBA{})
	h := halve(im)
	if got := h.NRGBAAt(0, 0); got != (color.NRGBA{5, 0, 0, 255}) {
		t.Errorf("block 0 = %v", got)
	}
	if got := h.NRGBAAt(1, 0); got.A != 0 {
		t.Errorf("block with a hole = %v, want transparent", got)
	}
}

func TestToleranceOutside(t *testing.T) {
	tol := tolerance{DeltaE: 5, Contrast: 0.3, Chroma: 0.3, Edges: 0.3}
	want := metrics{L: 50, A: 0, B: 30, Contrast: 10, Chroma: 30, Edges: 0.4}
	if got := tol.outside(want, want); len(got) != 0 {
		t.Errorf("identical: %v", got)
	}
	got := tol.outside(want, metrics{L: 60, A: 0, B: 30, Contrast: 5, Chroma: 35, Edges: 0.1})
	if !slices.Equal(got, []string{"colour", "contrast", "edges"}) {
		t.Errorf("outside = %v", got)
	}
}

func TestValidateTolerance(t *testing.T) {
	p := loadFixture(t)
	p.ref.Tolerance.Edges = 0
	if err := errors.Join(p.validate()...); err == nil || !strings.Contains(err.Error(), "tolerance") {
		t.Fatalf("err %v, want a tolerance error", err)
	}
}

func TestCompareAndTargets(t *testing.T) {
	p := loadFixture(t)
	p.ref.Image.Width, p.ref.Image.Height = 300, 160
	writeImage(t, p)
	src, present, err := p.source()
	if err != nil || !present {
		t.Fatalf("source: present=%t err=%v", present, err)
	}
	out := filepath.Join(t.TempDir(), "compare")
	cmp, err := p.compare(src, out, formatPNG)
	if err != nil {
		t.Fatal(err)
	}
	if len(cmp.Bands) != 1 || len(cmp.Families) != len(p.ref.Families) || len(cmp.Features) != 1 {
		t.Fatalf("comparison %d bands, %d families, %d features", len(cmp.Bands), len(cmp.Families), len(cmp.Features))
	}
	meadow := cmp.Families[0]
	if meadow.ID != "meadow" || len(meadow.Crops) != 3 || meadow.Crops[0] != "meadow-r1c3.png" {
		t.Fatalf("meadow %+v", meadow)
	}
	// The fixture image is flat grey: no contrast in the reference, some in the render.
	if meadow.Reference.Contrast != 0 || meadow.Render.Contrast == 0 || !slices.Contains(meadow.Outside, "colour") {
		t.Errorf("meadow reference %+v render %+v outside %v", meadow.Reference, meadow.Render, meadow.Outside)
	}
	for _, name := range append([]string{"compare.json", "band-1.png", "feature-crossing.png"}, meadow.Crops...) {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Error(err)
		}
	}

	if err := p.writeTargets(cmp); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(p.targetsPath)
	tg, err := p.loadTargets()
	if err != nil || tg == nil {
		t.Fatalf("loadTargets: %v %v", tg, err)
	}
	if errs := p.checkTargets(tg); len(errs) > 0 {
		t.Fatal(errors.Join(errs...))
	}
	if tg.Families[0].metrics != meadow.Reference {
		t.Errorf("target %+v, want %+v", tg.Families[0].metrics, meadow.Reference)
	}
	rows, err := p.drift(tg)
	if err != nil || len(rows) != len(p.ref.Families) || rows[0].Got != meadow.Render {
		t.Fatalf("drift %+v %v", rows, err)
	}
	if !strings.Contains(driftTable(rows), "meadow") {
		t.Error("drift table lacks meadow")
	}

	// A second measurement writes the same bytes.
	if err := p.writeTargets(cmp); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.ReadFile(p.targetsPath); !bytes.Equal(first, again) {
		t.Error("targets.json differs between two writes")
	}

	p.ref.Families[0].Exemplars = []string{"r1c3", "r3c1"}
	p.ref.Families = p.ref.Families[:len(p.ref.Families)-1]
	p.ref.Image.SHA256 = strings.Repeat("1", 64)
	err = errors.Join(p.checkTargets(tg)...)
	for _, want := range []string{"measured on image", "family meadow was measured on exemplars", "family stream is not in reference.json"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("checkTargets %v, want %q", err, want)
		}
	}
}

func TestMissingTargets(t *testing.T) {
	tg, err := loadFixture(t).loadTargets()
	if tg != nil || err != nil {
		t.Fatalf("loadTargets = %v, %v; want nothing", tg, err)
	}
}
