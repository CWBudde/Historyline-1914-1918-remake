package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Single-sprite exports are the established Lanczos contract. The neighborhood
// must validate those exports from their source and use the same independent
// native reduction; reducing the masked 2x export changes this sharp pattern.
func TestNeighborhoodSourceBackedLanczosMatchesSingleSprite(t *testing.T) {
	dir, config, positions := neighborhoodLanczosFixture(t)
	var first map[string][]byte
	for _, repeat := range []string{"first", "repeat"} {
		out := filepath.Join(dir, repeat)
		if err := run([]string{"-neighborhood", config, "-out", out}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		for _, scale := range []int{1, 2, 4} {
			name := fmt.Sprintf("candidate-%dx.png", scale)
			if scale == 1 {
				name = "downsample.png"
			}
			tile, err := loadPNG(filepath.Join(dir, "exports", name))
			if err != nil {
				t.Fatal(err)
			}
			got, err := loadPNG(filepath.Join(out, "scene", fmt.Sprintf("mixed-%dx.png", scale)))
			if err != nil {
				t.Fatal(err)
			}
			want := image.NewNRGBA(got.Bounds())
			for i, p := range positions {
				for y := 0; y < 24*scale; y++ {
					for x := 0; x < 24*scale; x++ {
						if !fixtureHex(x/scale, y/scale) {
							continue
						}
						c := color.NRGBA{R: 40, A: 255}
						if i%2 == 0 {
							c = tile.NRGBAAt(x, y)
						}
						want.SetNRGBA(p.X*scale+x, p.Y*scale+y, c)
					}
				}
			}
			if !bytes.Equal(got.Pix, want.Pix) {
				t.Fatalf("density %d differs from independently generated single-sprite exports", scale)
			}
		}
		artifacts := neighborhoodArtifacts(t, out)
		if repeat == "first" {
			first = artifacts
		} else if !reflect.DeepEqual(first, artifacts) {
			t.Fatal("repeated neighborhood run changed artifact bytes")
		}
	}
	var report struct {
		Inputs       map[string]string `json:"input_sha256"`
		Replacements []struct {
			Resource     string   `json:"resource"`
			SourcePath   string   `json:"source_path"`
			SourceSHA256 string   `json:"source_sha256"`
			Resampling   string   `json:"resampling"`
			Algorithm    string   `json:"resampling_algorithm"`
			Chain        []string `json:"resampling_chain"`
		} `json:"replacement_provenance"`
	}
	if err := json.Unmarshal(first["diagnostics.json"], &report); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "source.png")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	digest := hex.EncodeToString(sum[:])
	if len(report.Replacements) != 1 || report.Inputs[source] != digest {
		t.Fatalf("source provenance missing: %s", first["diagnostics.json"])
	}
	r := report.Replacements[0]
	wantChain := []string{"source -> raw4x: lanczos", "raw4x -> raw2x: lanczos", "raw4x -> rawNative: lanczos", "raw4x/raw2x/rawNative -> outputs: apply reference footprint independently"}
	if r.Resource != "road0005" || r.SourcePath != source || r.SourceSHA256 != digest || r.Resampling != "lanczos" || !reflect.DeepEqual(r.Chain, wantChain) {
		t.Fatalf("incorrect source provenance: %+v", r)
	}
	for _, term := range []string{"Lanczos3", "radius 3", "premultiplied", "half up"} {
		if !strings.Contains(r.Algorithm, term) {
			t.Fatalf("algorithm description missing %q: %s", term, r.Algorithm)
		}
	}
	// Original and native images retain pixel magnification. Only the two
	// recreated high-density scene images opt into smooth browser scaling.
	if got := bytes.Count(first["index.html"], []byte(`style="image-rendering:auto"`)); got != 2 {
		t.Fatalf("smooth high-density redraw images = %d; want 2", got)
	}
	if !bytes.Contains(first["index.html"], []byte("Selected redraw resources replace every occurrence among the seven cells.")) || bytes.Contains(first["index.html"], []byte("only resources matching the center")) {
		t.Fatal("source-backed gallery does not describe replacements of selected neighbor resources")
	}
}

func TestNeighborhoodSourceBackedLanczosRejectsUnverifiedExports(t *testing.T) {
	for _, kind := range []string{"missing source", "missing resampling", "unknown resampling", "area resampling", "bilinear resampling", "missing file", "small source", "wrong aspect", "wrong source", "source alpha hole", "wrong 2x", "wrong 4x", "hidden 2x", "wrong filter"} {
		t.Run(kind, func(t *testing.T) {
			dir, config, _ := neighborhoodLanczosFixture(t)
			rewriteNeighborhoodReplacement(t, config, func(r map[string]any) {
				switch kind {
				case "missing source":
					delete(r, "source")
				case "missing resampling":
					delete(r, "resampling")
				case "unknown resampling":
					r["resampling"] = "Lanczos3"
				case "area resampling":
					r["resampling"] = "area"
				case "bilinear resampling":
					r["resampling"] = "bilinear"
				case "missing file":
					r["source"] = "absent.png"
				case "wrong filter":
					out := filepath.Join(dir, "bilinear")
					if err := run([]string{"-reference", filepath.Join(dir, "road0005.png"), "-candidate", filepath.Join(dir, "source.png"), "-out", out, "-resampling", "bilinear"}, &bytes.Buffer{}); err != nil {
						t.Fatal(err)
					}
					r["image_2x"], r["image_4x"] = "bilinear/candidate-2x.png", "bilinear/candidate-4x.png"
				}
			})
			switch kind {
			case "small source":
				writeFixture(t, dir, "source.png", image.NewNRGBA(image.Rect(0, 0, 95, 95)))
			case "wrong aspect":
				writeFixture(t, dir, "source.png", image.NewNRGBA(image.Rect(0, 0, 192, 194)))
			case "source alpha hole":
				img, err := loadPNG(filepath.Join(dir, "source.png"))
				if err != nil {
					t.Fatal(err)
				}
				for y := 64; y < 128; y++ {
					for x := 64; x < 128; x++ {
						img.SetNRGBA(x, y, color.NRGBA{})
					}
				}
				writeFixture(t, dir, "source.png", img)
			case "wrong source", "wrong 2x", "wrong 4x", "hidden 2x":
				name := "source.png"
				if kind == "wrong 2x" || kind == "hidden 2x" {
					name = "exports/candidate-2x.png"
				}
				if kind == "wrong 4x" {
					name = "exports/candidate-4x.png"
				}
				img, err := loadPNG(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				p := img.Bounds().Max.X / 2
				if kind == "hidden 2x" {
					p = 0
				}
				c := img.NRGBAAt(p, p)
				c.R ^= 255
				img.SetNRGBA(p, p, c)
				writeFixture(t, dir, name, img)
			}
			out := filepath.Join(dir, "rejected")
			if err := run([]string{"-neighborhood", config, "-out", out}, &bytes.Buffer{}); err == nil {
				t.Fatal("accepted unverified source/export/filter")
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("created output before validation")
			}
		})
	}
}

func TestNeighborhoodLegacyRemainsWithoutSourceProvenance(t *testing.T) {
	dir, config, _ := neighborhoodFixture(t, false)
	out := filepath.Join(dir, "legacy")
	if err := run([]string{"-neighborhood", config, "-out", out}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(neighborhoodArtifacts(t, out)["diagnostics.json"], &report); err != nil {
		t.Fatal(err)
	}
	if _, exists := report["replacement_provenance"]; exists {
		t.Fatal("legacy diagnostics changed schema")
	}
	if bytes.Contains(neighborhoodArtifacts(t, out)["index.html"], []byte("image-rendering:auto")) {
		t.Fatal("legacy gallery changed scaling")
	}
}

func neighborhoodLanczosFixture(t *testing.T) (string, string, []image.Point) {
	t.Helper()
	dir, config, positions := neighborhoodFixture(t, false)
	src := image.NewNRGBA(image.Rect(0, 0, 192, 192))
	for y := 0; y < 192; y++ {
		for x := 0; x < 192; x++ {
			c := color.NRGBA{R: uint8((x*17 + y*13) % 256), G: uint8((x/3 + y/5) % 2 * 255), A: 255}
			src.SetNRGBA(x, y, c)
		}
	}
	writeFixture(t, dir, "source.png", src)
	out := filepath.Join(dir, "exports")
	if err := run([]string{"-reference", filepath.Join(dir, "road0005.png"), "-candidate", filepath.Join(dir, "source.png"), "-out", out, "-resampling", "lanczos"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	rewriteNeighborhoodReplacement(t, config, func(r map[string]any) {
		r["source"], r["resampling"] = "source.png", "lanczos"
		r["image_2x"], r["image_4x"] = "exports/candidate-2x.png", "exports/candidate-4x.png"
	})
	return dir, config, positions
}

func rewriteNeighborhoodReplacement(t *testing.T, config string, edit func(map[string]any)) {
	t.Helper()
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	edit(c["replacements"].([]any)[0].(map[string]any))
	data, err = json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func neighborhoodArtifacts(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	artifacts := map[string][]byte{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		artifacts[name], err = os.ReadFile(path)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return artifacts
}
