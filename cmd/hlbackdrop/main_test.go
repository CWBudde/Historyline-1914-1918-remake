package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuality(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for _, q := range []int{1, 49, 50, 75, 90, 95, 100} {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, im, &jpeg.Options{Quality: q}); err != nil {
			t.Fatal(err)
		}
		got, err := quality(buf.Bytes())
		if err != nil || got != q {
			t.Errorf("quality of a q%d JPEG = %d, %v", q, got, err)
		}
	}
	if _, err := quality([]byte{0xFF, 0xD8}); err == nil {
		t.Error("a JPEG without a table gave a quality")
	}
}

// TestCheck walks a levels directory through a missing source, a written
// JPEG, a stale JPEG and a source no scenario names.
func TestCheck(t *testing.T) {
	levels := t.TempDir()
	write := func(name string, data []byte) {
		t.Helper()
		p := filepath.Join(levels, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	source := func(name string, c uint8) {
		t.Helper()
		im := image.NewRGBA(image.Rect(0, 0, 24, 16))
		for i := range im.Pix {
			im.Pix[i] = c + uint8(i%7)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, im); err != nil {
			t.Fatal(err)
		}
		write(name, buf.Bytes())
	}
	expect := func(want ...string) {
		t.Helper()
		errs := check(levels)
		if len(errs) != len(want) {
			t.Fatalf("check = %v, want %d errors %q", errs, len(want), want)
		}
		for i, err := range errs {
			if !strings.Contains(err.Error(), want[i]) {
				t.Errorf("check error %q, want it to say %q", err, want[i])
			}
		}
	}

	write("monthly/1914-08.json", []byte(`{"month": "1914-08", "backdrop": "1914-08.jpg"}`))
	write("monthly/1914-09.json", []byte(`{"month": "1914-09"}`))
	write("skirmish/01-track.json", []byte(`{"name": "track"}`))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 24, 16)), &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	write("monthly/1914-08.jpg", buf.Bytes())
	expect("no source")

	source("backdrops/monthly/1914-08.png", 40)
	expect("is not")
	b, err := find(levels, "monthly/1914-08")
	if err != nil {
		t.Fatal(err)
	}
	if _, q, err := writeJPEG(b, 0); err != nil || q != 90 {
		t.Fatalf("writeJPEG = quality %d, %v; want the JPEG's own 90", q, err)
	}
	expect()

	source("backdrops/monthly/1914-08.png", 80)
	expect("is not")
	if _, _, err := writeJPEG(b, 0); err != nil {
		t.Fatal(err)
	}
	source("backdrops/monthly/1914-09.png", 80)
	expect("no scenario names")

	if _, err := find(levels, "monthly/1914-09"); err == nil {
		t.Error("find gave a backdrop for a scenario without one")
	}
}

// TestShipped checks the repository's backdrops against their sources.
func TestShipped(t *testing.T) {
	if testing.Short() {
		t.Skip("decodes every backdrop source")
	}
	for _, err := range check(filepath.Join("..", "..", "levels")) {
		t.Error(err)
	}
}
