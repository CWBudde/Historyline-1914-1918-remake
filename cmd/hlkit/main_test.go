package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestPNG(t *testing.T, path string, im image.Image) string {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
}

// testSheet is a 2×1 sheet: a 40 px square with a hazy rim in the left cell,
// a 20×10 px bar in the right cell, both on an alpha-20 matte.
func testSheet() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			im.SetNRGBA(x, y, color.NRGBA{120, 100, 60, 20})
		}
	}
	for y := 30; y < 70; y++ {
		for x := 30; x < 70; x++ {
			im.SetNRGBA(x, y, color.NRGBA{120, 100, 60, 255})
		}
	}
	for y := 45; y < 55; y++ {
		for x := 140; x < 160; x++ {
			im.SetNRGBA(x, y, color.NRGBA{60, 60, 60, 255})
		}
	}
	return im
}

func testProfile(t *testing.T, dir string, cells string) string {
	t.Helper()
	sheetSHA := writeTestPNG(t, filepath.Join(dir, "art/terrain/test/test-kit-shared-v1.png"), testSheet())
	ground := image.NewNRGBA(image.Rect(0, 0, 300, 200))
	for i := range ground.Pix {
		ground.Pix[i] = [4]uint8{100, 107, 40, 255}[i%4]
	}
	groundSHA := writeTestPNG(t, filepath.Join(dir, "ground.png"), ground)
	t.Chdir(dir)
	p := fmt.Sprintf(`{"schema_version":1,"densities":[2,4],"meadow":{"source":"ground.png","source_sha256":"%s"},
 "sheets":[{"id":"kit/test/kit","art":"terrain/test","source":"art/terrain/test/test-kit-shared-v1.png","source_sha256":"%s",
  "grid":[2,1],"cut":32,"min_area":4,"scale":10,"grade_rgb":[140,132,118],"cells":[%s]}]}`, groundSHA, sheetSHA, cells)
	if err := os.WriteFile("profile.json", []byte(p), 0644); err != nil {
		t.Fatal(err)
	}
	return "profile.json"
}

type testReport struct {
	Sheets []sheetReport `json:"sheets"`
}

func readReport(t *testing.T) testReport {
	t.Helper()
	b, err := os.ReadFile("out/kits.json")
	if err != nil {
		t.Fatal(err)
	}
	var r testReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRunPreparesSizedGradedSprites(t *testing.T) {
	p := testProfile(t, t.TempDir(), `{"cell":"r1c1","size":5},{"cell":"r1c2"}`)
	if err := run(p, "out", true); err != nil {
		t.Fatal(err)
	}
	r := readReport(t)
	s := r.Sheets[0]
	if !s.Pass || s.HazeFrom == 0 || s.Haze != 0 || len(s.Items) != 2 || len(s.Contact) != 2 {
		t.Fatalf("sheet report %+v", s)
	}
	square, bar := s.Items[0], s.Items[1]
	if square.ExtentFrom != 4 || square.Extent != 5 {
		t.Fatalf("square extent %v -> %v, want 4 -> 5", square.ExtentFrom, square.Extent)
	}
	// Without a size the bar keeps the painting scale: 20 px at 10 px per world px.
	if bar.Extent != 2 || bar.Outputs[1].Width != 8 || bar.Outputs[1].Height != 4 {
		t.Fatalf("bar %+v", bar)
	}
	if square.Outputs[0].Width != 10 || square.Outputs[1].Width != 20 || square.Outputs[0].Anchor != [2]float64{2.5, 2.5} {
		t.Fatalf("square outputs %+v", square.Outputs)
	}
	f, err := os.Open(filepath.Join("out", filepath.FromSlash(square.Outputs[1].Path)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, a := im.At(0, 0).RGBA(); a != 0xffff {
		t.Fatal("the sprite lost its opaque corner")
	}
}

func TestRunFailsAnEmptyOrUnlistedCell(t *testing.T) {
	p := testProfile(t, t.TempDir(), `{"cell":"r1c1"}`)
	err := run(p, "out", true)
	if err == nil || !strings.Contains(err.Error(), "kit/test/kit") {
		t.Fatalf("an unlisted item passed: %v", err)
	}
	if err := run(p, "out", false); err != nil {
		t.Fatalf("report mode failed: %v", err)
	}
	s := readReport(t).Sheets[0]
	if s.Pass || !strings.Contains(s.Checks[len(s.Checks)-1].Name, "r1c2") {
		t.Fatalf("checks %+v", s.Checks)
	}
}

func TestValidateRejectsBadCells(t *testing.T) {
	for _, cells := range []string{`{"cell":"r2c1"}`, `{"cell":"r1c1"},{"cell":"r1c1"}`, `{"cell":"x1"}`, `{"cell":"r01c1"}`} {
		p := testProfile(t, t.TempDir(), cells)
		if err := run(p, "out", false); err == nil || !strings.Contains(err.Error(), "cell") {
			t.Errorf("cells %s accepted: %v", cells, err)
		}
	}
}
