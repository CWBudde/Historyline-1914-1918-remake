package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func packFixture(t *testing.T, width, height int, transparent bool) (string, map[string]any, *image.NRGBA) {
	t.Helper()
	dir := t.TempDir()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x % 251), uint8(y % 251), uint8((x + y) % 251), 255})
		}
	}
	if transparent {
		img.SetNRGBA(0, 0, color.NRGBA{})
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "source.png"), b.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b.Bytes())
	manifest := map[string]any{"schema_version": 1, "art_direction": "Painterly terrain", "assets": []any{map[string]any{"id": "plain", "source": "source.png", "source_sha256": hex.EncodeToString(sum[:]), "description": "A terrain candidate"}}}
	return dir, manifest, img
}

func writePackManifest(t *testing.T, dir string, manifest map[string]any) string {
	t.Helper()
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "pack.json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPackCenterCropRegistrationAndRawRectangles(t *testing.T) {
	for _, size := range [][2]int{{401, 320}, {384, 341}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			dir, m, src := packFixture(t, size[0], size[1], false)
			out := filepath.Join(dir, "out")
			if err := run([]string{"-pack", writePackManifest(t, dir, m), "-out", out}, io.Discard); err != nil {
				t.Fatal(err)
			}
			got, err := loadPNG(filepath.Join(out, "plain-4x.png"))
			if err != nil {
				t.Fatal(err)
			}
			if got.Bounds() != image.Rect(0, 0, 384, 320) {
				t.Fatalf("4x bounds %v", got.Bounds())
			}
			dx, dy := (size[0]-384)/2, (size[1]-320)/2
			for y := 0; y < 320; y++ {
				for x := 0; x < 384; x++ {
					if got.NRGBAAt(x, y) != src.NRGBAAt(x+dx, y+dy) {
						t.Fatalf("crop registration at %d,%d: got %v want %v", x, y, got.NRGBAAt(x, y), src.NRGBAAt(x+dx, y+dy))
					}
				}
			}
			two, err := loadPNG(filepath.Join(out, "plain-2x.png"))
			if err != nil {
				t.Fatal(err)
			}
			if two.Bounds() != image.Rect(0, 0, 192, 160) {
				t.Fatalf("2x bounds %v", two.Bounds())
			}
			// A rectangle preserves overscan at every corner; applying a hex footprint would erase these.
			for _, p := range []image.Point{{0, 0}, {191, 0}, {0, 159}, {191, 159}} {
				if two.NRGBAAt(p.X, p.Y).A != 255 {
					t.Fatalf("corner %v was clipped", p)
				}
			}
		})
	}
}

func TestPackRejectsInvalidInputsWithoutOutput(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		transparent   bool
		mutate        func(map[string]any)
	}{
		{name: "hash mismatch", width: 384, height: 320, mutate: func(m map[string]any) {
			m["assets"].([]any)[0].(map[string]any)["source_sha256"] = strings.Repeat("0", 64)
		}},
		{name: "malformed hash", width: 384, height: 320, mutate: func(m map[string]any) { m["assets"].([]any)[0].(map[string]any)["source_sha256"] = "abc" }},
		{name: "too narrow", width: 383, height: 320}, {name: "too short", width: 384, height: 319}, {name: "transparent", width: 384, height: 320, transparent: true},
		{name: "unsafe id", width: 384, height: 320, mutate: func(m map[string]any) { m["assets"].([]any)[0].(map[string]any)["id"] = "../escape" }},
		{name: "duplicate id", width: 384, height: 320, mutate: func(m map[string]any) { a := m["assets"].([]any); m["assets"] = append(a, a[0]) }},
		{name: "case-insensitive duplicate id", width: 384, height: 320, mutate: func(m map[string]any) {
			a := m["assets"].([]any)
			first := a[0].(map[string]any)
			m["assets"] = append(a, map[string]any{"id": "PLAIN", "source": first["source"], "source_sha256": first["source_sha256"], "description": "Duplicate with different case"})
		}},
		{name: "schema", width: 384, height: 320, mutate: func(m map[string]any) { m["schema_version"] = 2 }},
		{name: "unknown field", width: 384, height: 320, mutate: func(m map[string]any) { m["surprise"] = true }},
		{name: "empty assets", width: 384, height: 320, mutate: func(m map[string]any) { m["assets"] = []any{} }},
		{name: "missing direction", width: 384, height: 320, mutate: func(m map[string]any) { delete(m, "art_direction") }},
		{name: "missing description", width: 384, height: 320, mutate: func(m map[string]any) { delete(m["assets"].([]any)[0].(map[string]any), "description") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, m, _ := packFixture(t, tt.width, tt.height, tt.transparent)
			if tt.mutate != nil {
				tt.mutate(m)
			}
			out := filepath.Join(dir, "out")
			if err := run([]string{"-pack", writePackManifest(t, dir, m), "-out", out}, io.Discard); err == nil {
				t.Fatal("invalid input accepted")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatalf("invalid input left output: %v", err)
			}
		})
	}
}

func TestPackReproducibleReportAndGallery(t *testing.T) {
	dir, m, _ := packFixture(t, 780, 650, false)
	manifest := writePackManifest(t, dir, m)
	for _, name := range []string{"first", "second"} {
		if err := run([]string{"-pack", manifest, "-out", filepath.Join(dir, name)}, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(dir, "first"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("artifacts: %v", entries)
	}
	for _, entry := range entries {
		first, err := os.ReadFile(filepath.Join(dir, "first", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		second, err := os.ReadFile(filepath.Join(dir, "second", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Errorf("%s is not reproducible", entry.Name())
		}
	}
	reportBytes, err := os.ReadFile(filepath.Join(dir, "first", "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Status       string   `json:"status"`
		ManifestHash string   `json:"manifest_sha256"`
		Resampling   string   `json:"resampling"`
		Chain        []string `json:"resampling_chain"`
		Assets       []struct {
			SourceHash string                            `json:"source_sha256"`
			SourceSize dimensions                        `json:"source_dimensions"`
			Crop       struct{ X, Y, Width, Height int } `json:"crop_rectangle"`
			Outputs    []struct {
				Path string     `json:"path"`
				Hash string     `json:"sha256"`
				Size dimensions `json:"dimensions"`
			} `json:"outputs"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "candidate" || report.Resampling != "lanczos" {
		t.Fatalf("report lacks candidate Lanczos provenance: %s", reportBytes)
	}
	manifestBytes, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestHash := sha256.Sum256(manifestBytes)
	if report.ManifestHash != hex.EncodeToString(manifestHash[:]) {
		t.Fatal("missing manifest input hash")
	}
	wantChain := []string{"source -> centered 6:5 crop (no resampling)", "crop -> raw4x (384x320): lanczos3", "raw4x -> raw2x (192x160): lanczos3"}
	if !reflect.DeepEqual(report.Chain, wantChain) {
		t.Fatalf("chain %v", report.Chain)
	}
	if len(report.Assets) != 1 {
		t.Fatalf("assets %d", len(report.Assets))
	}
	a := report.Assets[0]
	if a.SourceHash != m["assets"].([]any)[0].(map[string]any)["source_sha256"] || a.SourceSize != (dimensions{780, 650}) || a.Crop.X != 0 || a.Crop.Y != 0 || a.Crop.Width != 780 || a.Crop.Height != 650 {
		t.Fatalf("source provenance %+v", a)
	}
	if len(a.Outputs) != 2 {
		t.Fatalf("outputs %+v", a.Outputs)
	}
	for _, o := range a.Outputs {
		b, err := os.ReadFile(filepath.Join(dir, "first", o.Path))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if o.Hash != hex.EncodeToString(sum[:]) {
			t.Errorf("wrong hash for %s", o.Path)
		}
	}
	html, err := os.ReadFile(filepath.Join(dir, "first", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, word := range []string{"plain-source.png", "plain-2x.png", "plain-4x.png", "clipping", "joins", "pending"} {
		if !bytes.Contains(html, []byte(word)) {
			t.Errorf("gallery missing %q", word)
		}
	}
	for _, b := range [][]byte{html, reportBytes} {
		for _, bad := range []string{"reference_path", "original/", "amiga-native", "apply reference footprint"} {
			if bytes.Contains(b, []byte(bad)) {
				t.Errorf("pack includes forbidden reference %q", bad)
			}
		}
	}
	// Exported 2x pixels must come from the exported raw4x stage.
	four, err := loadPNG(filepath.Join(dir, "first", "plain-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	two, err := loadPNG(filepath.Join(dir, "first", "plain-2x.png"))
	if err != nil {
		t.Fatal(err)
	}
	wantTwo, err := resizeLanczos(four, 192, 160)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(two.Pix, wantTwo.Pix) {
		t.Fatal("2x bypassed raw4x")
	}
}

func TestPackRejectsExistingDirectoryAndConflictingFlags(t *testing.T) {
	dir, m, _ := packFixture(t, 384, 320, false)
	manifest := writePackManifest(t, dir, m)
	existing := filepath.Join(dir, "existing")
	if err := os.Mkdir(existing, 0755); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-pack", manifest, "-out", existing}, io.Discard); err == nil {
		t.Fatal("existing empty directory accepted")
	}
	for _, flag := range []string{"-reference", "-candidate", "-reference-platform", "-neighborhood", "-resampling"} {
		t.Run(flag, func(t *testing.T) {
			out := filepath.Join(dir, "out")
			if err := run([]string{"-pack", manifest, "-out", out, flag, "unused"}, io.Discard); err == nil {
				t.Fatal("conflicting flag accepted")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatalf("conflict left output: %v", err)
			}
		})
	}
}

func TestPackEmptyFlagCannotFallThroughToReferenceMode(t *testing.T) {
	err := run([]string{"-pack", "", "-reference", "missing-reference.png", "-candidate", "missing-candidate.png", "-out", filepath.Join(t.TempDir(), "out")}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "-pack requires") {
		t.Fatalf("empty pack flag fell through to another mode: %v", err)
	}
}

func TestPackExcludesOriginalSourcesAndResolvedSymlinks(t *testing.T) {
	for _, symlink := range []bool{false, true} {
		t.Run(fmt.Sprintf("symlink=%t", symlink), func(t *testing.T) {
			dir, m, _ := packFixture(t, 384, 320, false)
			original := filepath.Join(dir, "original")
			if err := os.Mkdir(original, 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(original, "source.png")
			if err := os.Rename(filepath.Join(dir, "source.png"), target); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join("original", "source.png")
			if symlink {
				source = "authored.png"
				if err := os.Symlink(target, filepath.Join(dir, source)); err != nil {
					t.Fatal(err)
				}
			}
			m["assets"].([]any)[0].(map[string]any)["source"] = source
			out := filepath.Join(dir, "out")
			if err := run([]string{"-pack", writePackManifest(t, dir, m), "-out", out}, io.Discard); err == nil {
				t.Fatal("excluded source accepted")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatalf("excluded source left output: %v", err)
			}
			b, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b)
			if hex.EncodeToString(sum[:]) != m["assets"].([]any)[0].(map[string]any)["source_sha256"] {
				t.Fatal("excluded source was modified")
			}
		})
	}
}

func TestPackCropsBeforeReductionWithoutOutsideColorBleeding(t *testing.T) {
	dir, m, src := packFixture(t, 803, 652, false)
	// Largest exact 6:5 crop is 780x650 at (11,1), leaving asymmetric red overscan.
	blue := color.NRGBA{B: 255, A: 255}
	for y := 0; y < 652; y++ {
		for x := 0; x < 803; x++ {
			p := color.NRGBA{R: 255, A: 255}
			if x >= 11 && x < 791 && y >= 1 && y < 651 {
				p = blue
			}
			src.SetNRGBA(x, y, p)
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, src); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "source.png"), encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(encoded.Bytes())
	m["assets"].([]any)[0].(map[string]any)["source_sha256"] = hex.EncodeToString(sum[:])
	out := filepath.Join(dir, "out")
	if err := run([]string{"-pack", writePackManifest(t, dir, m), "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, scale := range []string{"2x", "4x"} {
		got, err := loadPNG(filepath.Join(out, "plain-"+scale+".png"))
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < got.Bounds().Dy(); y++ {
			for x := 0; x < got.Bounds().Dx(); x++ {
				if p := got.NRGBAAt(x, y); p != blue {
					t.Fatalf("%s pixel %d,%d = %v; outside crop contaminated solid blue", scale, x, y, p)
				}
			}
		}
	}
}

func TestPackDoesNotOverwriteExistingOutputFileOrSymlink(t *testing.T) {
	for _, symlink := range []bool{false, true} {
		t.Run(fmt.Sprintf("symlink=%t", symlink), func(t *testing.T) {
			dir, m, _ := packFixture(t, 384, 320, false)
			existing := filepath.Join(dir, "existing")
			if err := os.Mkdir(existing, 0755); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(existing, "sentinel")
			if err := os.WriteFile(sentinel, []byte("keep me"), 0644); err != nil {
				t.Fatal(err)
			}
			out := sentinel
			if symlink {
				out = filepath.Join(dir, "out")
				if err := os.Symlink(existing, out); err != nil {
					t.Fatal(err)
				}
			}
			if err := run([]string{"-pack", writePackManifest(t, dir, m), "-out", out}, io.Discard); err == nil {
				t.Fatal("existing output accepted")
			}
			b, err := os.ReadFile(sentinel)
			if err != nil || string(b) != "keep me" {
				t.Fatalf("sentinel changed: %q, %v", b, err)
			}
			entries, err := os.ReadDir(existing)
			if err != nil || len(entries) != 1 {
				t.Fatalf("existing directory changed: %v, %v", entries, err)
			}
			if symlink {
				if target, err := os.Readlink(out); err != nil || target != existing {
					t.Fatalf("output symlink changed: %q, %v", target, err)
				}
			}
		})
	}
}
