package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests catch enlargement, misplaced crops, alpha contamination, incorrect
// density metadata and accepting unverified input before publishing output.
func TestExportCropAndTransparentCompanion(t *testing.T) {
	root, m := fixture(t)
	m.Assets = append(m.Assets, asset{ID: "stamp/knoll/1/1", Transparent: true, Canvas: [2]int{8, 6}, Anchor: [2]float64{4, 3}, Layer: "ground", Art: "terrain/mountain", Path: "terrain/mountain/knoll-ground"})
	out := filepath.Join(root, "out")
	if err := run([]string{"-manifest", writeManifest(t, root, m), "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	var entries []indexEntry
	readJSON(t, filepath.Join(out, "index-entries.json"), &entries)
	if len(entries) != 4 {
		t.Fatalf("entries = %d", len(entries))
	}
	for _, e := range entries {
		if e.Anchor != [2]float64{4, 3} || e.Width != 8*e.Density || e.Height != 6*e.Density {
			t.Fatalf("wrong geometry: %+v", e)
		}
		b, err := os.ReadFile(filepath.Join(out, e.Path))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != e.SHA256 {
			t.Fatal("wrong output hash")
		}
		im, err := png.Decode(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		if e.ID == "stamp/knoll/1/1" {
			for y := 0; y < e.Height; y++ {
				for x := 0; x < e.Width; x++ {
					if _, _, _, a := im.At(x, y).RGBA(); a != 0 {
						t.Fatal("companion is not transparent")
					}
				}
			}
		} else {
			p := color.NRGBAModel.Convert(im.At(e.Width/2, e.Height/2)).(color.NRGBA)
			if p.R != 80 || p.G != 100 || p.B != 20 || p.A < 125 || p.A > 131 {
				t.Fatalf("crop or alpha changed: %+v", p)
			}
			if e.Pack != "terrain-standing-"+string(rune('0'+e.Density))+"x" || e.Layer != "standing" {
				t.Fatalf("wrong standing metadata: %+v", e)
			}
		}
	}
	if err := run([]string{"-manifest", writeManifest(t, root, m), "-out", out}, io.Discard); err == nil {
		t.Fatal("overwrote existing output")
	}
}

func TestInvalidInputDoesNotPublish(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		change     func(*manifest)
	}{
		{"pin", "SHA256", func(m *manifest) { m.Assets[0].SourceSHA256 = strings.Repeat("0", 64) }},
		{"enlarge", "enlarg", func(m *manifest) { m.Assets[0].Canvas = [2]int{16, 12} }},
		{"bounds", "bounds", func(m *manifest) { m.Assets[0].Crop = [4]int{20, 0, 32, 24} }},
		{"aspect", "aspect", func(m *manifest) { m.Assets[0].Crop = [4]int{0, 0, 31, 24} }},
		{"overhang", "overhang", func(m *manifest) { m.Assets[0].Canvas = [2]int{200, 6}; m.Assets[0].Anchor = [2]float64{100, 3} }},
		{"escape", "path", func(m *manifest) { m.Assets[0].Path = "../escape" }},
		{"duplicate", "duplicate", func(m *manifest) { m.Assets = append(m.Assets, m.Assets[0]) }},
		{"layer", "layer", func(m *manifest) { m.Assets[0].Layer = "ground" }},
		{"missing pin", "SHA256", func(m *manifest) { m.Assets[0].SourceSHA256 = "" }},
		{"mixed transparent source", "transparent", func(m *manifest) { m.Assets[0].Transparent = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, m := fixture(t)
			tc.change(&m)
			out := filepath.Join(root, "out")
			err := run([]string{"-manifest", writeManifest(t, root, m), "-out", out}, io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("invalid input published output")
			}
		})
	}
}

func TestClusterCanvasAndAlphaReduction(t *testing.T) {
	root, m := fixture(t)
	// A half-transparent silhouette beside fully transparent magenta catches
	// nearest-neighbour reduction and RGB bleeding from alpha-zero pixels.
	im := image.NewNRGBA(image.Rect(0, 0, 40, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 40; x++ {
			im.SetNRGBA(x, y, color.NRGBA{255, 0, 255, 0})
			if x < 20 {
				im.SetNRGBA(x, y, color.NRGBA{80, 100, 20, 255})
			}
		}
	}
	f, err := os.Create(m.Assets[0].Source)
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(m.Assets[0].Source)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	m.Assets[0].SourceSHA256 = hex.EncodeToString(h[:])
	m.Assets = append(m.Assets, asset{ID: "stamp/crater/4_-1.1_1.1_0.2/1", Transparent: true, Canvas: [2]int{238, 160}, Anchor: [2]float64{119, 40}, Layer: "ground", Art: "terrain/ground", Path: "terrain/ground/crater"})
	out := filepath.Join(root, "out")
	if err = run([]string{"-manifest", writeManifest(t, root, m), "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	f, err = os.Open(filepath.Join(out, "terrain/house/standing-2x.png"))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(f)
	closeErr := f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	for _, x := range []int{7, 8} {
		p := color.NRGBAModel.Convert(decoded.At(x, 6)).(color.NRGBA)
		if p.A == 0 || p.A == 255 || p.R != 80 || p.G != 100 || p.B != 20 {
			t.Fatalf("unfiltered or contaminated edge at %d: %+v", x, p)
		}
	}
	var entries []indexEntry
	readJSON(t, filepath.Join(out, "index-entries.json"), &entries)
	if entries[3].Width != 952 || entries[3].Height != 640 || entries[3].Pack != "terrain-stamps-4x" || entries[3].Anchor != [2]float64{119, 40} {
		t.Fatalf("cluster geometry: %+v", entries[3])
	}
}

func fixture(t *testing.T) (string, manifest) {
	t.Helper()
	root := t.TempDir()
	p := filepath.Join(root, "master.png")
	im := image.NewNRGBA(image.Rect(0, 0, 40, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 40; x++ {
			im.SetNRGBA(x, y, color.NRGBA{255, 0, 255, 0})
			if x >= 4 && x < 36 {
				im.SetNRGBA(x, y, color.NRGBA{80, 100, 20, 128})
			}
		}
	}
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	return root, manifest{SchemaVersion: 1, Assets: []asset{{ID: "standing/house", Source: p, SourceSHA256: hex.EncodeToString(h[:]), Crop: [4]int{4, 0, 32, 24}, Canvas: [2]int{8, 6}, Anchor: [2]float64{4, 3}, Layer: "standing", Art: "terrain/house", Path: "terrain/house/standing"}}}
}
func writeManifest(t *testing.T, root string, m manifest) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "manifest.json")
	if err = os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func readJSON(t *testing.T, p string, v any) {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}
