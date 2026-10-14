package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func syntheticData(t *testing.T, root string) string {
	t.Helper()
	data := filepath.Join(root, "original", "input")
	if err := os.MkdirAll(filepath.Join(data, "LIB"), 0700); err != nil {
		t.Fatal(err)
	}
	m, inv := maskFixture()
	for name, b := range map[string][]byte{"00.PAL": paletteFixture(), "HEXMSK00.DAT": m, "HEXMSK20.DAT": inv, "LIB/PARTS.LIB": libraryFixture("SYNTH001", graphicFixture(24, 24, 1, nil, make([]byte, 96)))} {
		if err := os.WriteFile(filepath.Join(data, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return data
}
func TestCLIScreenOwnPaletteMetadataHashesAndGallery(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	data := syntheticData(t, root)
	screenBytes := screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0})
	if err := os.WriteFile(filepath.Join(data, "GAME.IFF"), screenBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-data", data, "-libraries", "PARTS", "-screen", "GAME.IFF"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "build", "native-reference")
	b, err := os.ReadFile(filepath.Join(out, "screens", "GAME.IFF.png"))
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 3 || im.Bounds().Dy() != 2 {
		t.Fatal("screen resized")
	}
	r, g, blue, a := im.At(0, 0).RGBA()
	if r != 19*257 || g != 37*257 || blue != 55*257 || a != 65535 {
		t.Fatalf("screen ignored CMAP: %d %d %d %d", r, g, blue, a)
	}
	metadata, err := os.ReadFile(filepath.Join(out, "screen-metadata.csv"))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(metadata)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("screen rows %v", rows)
	}
	got := map[string]string{}
	for i, key := range rows[0] {
		got[key] = rows[1][i]
	}
	for key, want := range map[string]string{"source": "GAME.IFF", "platform": "amiga-native", "format": "FORM/ILBM", "width": "3", "height": "2", "row_bytes": "2", "planes": "2", "masking": "0", "compression": "0", "coverage": "none", "palette": "CMAP", "palette_colors": "4", "bmhd_x_aspect": "10", "bmhd_y_aspect": "11", "bmhd_page_width": "320", "bmhd_page_height": "200", "body_bytes": "8", "planar_bytes": "8", "source_sha256": fmt.Sprintf("%x", sha256.Sum256(screenBytes)), "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b))} {
		if got[key] != want {
			t.Errorf("%s=%q want %q", key, got[key], want)
		}
	}
	manifest, err := os.ReadFile(filepath.Join(out, "source-sha256.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), fmt.Sprintf("%x  GAME.IFF\n", sha256.Sum256(screenBytes))) {
		t.Fatalf("screen source missing in manifest: %s", manifest)
	}
	gallery, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"PARTS/SYNTH001.png", "screens/GAME.IFF.png", "GAME.IFF", "CMAP", "00.PAL", "24 × 24", "3 × 2", "10:11", "screen-metadata.csv", "source-sha256.txt", `id="PARTS-SYNTH001"`} {
		if !strings.Contains(string(gallery), want) {
			t.Errorf("gallery missing %q", want)
		}
	}
	// A default run still exports libraries and gallery without requiring GAME.IFF.
	if err := os.Remove(filepath.Join(data, "GAME.IFF")); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-data", data, "-libraries", "PARTS", "-out", "build/default"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "build", "default", "screen-metadata.csv")); !os.IsNotExist(err) {
		t.Fatal("default wrote screen metadata")
	}
}
func TestCLIRejectsScreenBeforeWriting(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	data := syntheticData(t, root)
	if err := os.WriteFile(filepath.Join(data, "GAME.IFF"), []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"GAME.IFF", "missing.IFF", "../GAME.IFF", "/tmp/GAME.IFF", "bad\nname.IFF"} {
		if err := run([]string{"-data", data, "-libraries", "PARTS", "-screen", name}, &bytes.Buffer{}); err == nil {
			t.Errorf("accepted %q", name)
		}
		if _, err := os.Stat(filepath.Join(root, "build", "native-reference")); !os.IsNotExist(err) {
			t.Fatal("invalid screen created output")
		}
	}
}
func TestGalleryEscapesAllReferenceValues(t *testing.T) {
	evil := `<script>alert("x")</script>&'`
	b, err := renderGallery([]galleryItem{{ID: evil, Library: evil, Name: evil, Coverage: evil, Palette: evil, Platform: evil, Path: `javascript:alert(1)`, Width: 3, Height: 2, Aspect: evil}}, true)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, evil) || strings.Contains(s, `href="javascript:`) || strings.Contains(s, `src="javascript:`) {
		t.Fatal("unescaped reference value or unsafe link")
	}
	if !strings.Contains(s, "&lt;script&gt;") {
		t.Fatal("reference name missing")
	}
	for _, want := range []string{`id="search"`, `id="library"`, `id="scale"`, `value="1"`, `value="2"`, `value="4"`, `value="8"`, `id="background"`, `value="checker"`, `value="dark"`, `value="light"`, `data-width="3"`, `data-height="2"`} {
		if !strings.Contains(s, want) {
			t.Errorf("gallery missing control/size %q", want)
		}
	}
	if strings.Contains(s, "fetch(") || strings.Contains(s, `src="http`) || strings.Contains(s, `href="http`) {
		t.Fatal("gallery requires network")
	}
}

func briefingData(t *testing.T, data string, b []byte) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(data, "TXT"), 0700); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 46; i++ {
		if err := os.WriteFile(filepath.Join(data, "TXT", fmt.Sprintf("H%02d.LBM", i)), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
func TestCLIBriefingsFixedCorpusAndCombinedMetadata(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	data := syntheticData(t, root)
	b := screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0})
	briefingData(t, data, b)
	if err := os.WriteFile(filepath.Join(data, "GAME.IFF"), b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(data, "TXT", "H46.LBM"), []byte("not part of corpus"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-data", data, "-libraries", "PARTS", "-screen", "GAME.IFF", "-briefings"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "build", "native-reference")
	metadata, err := os.ReadFile(filepath.Join(out, "screen-metadata.csv"))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(metadata)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 48 {
		t.Fatalf("got %d rows want header + screen + 46 briefings", len(rows))
	}
	for i := 0; i < 46; i++ {
		name := fmt.Sprintf("TXT/H%02d.LBM", i)
		if rows[i+2][0] != name {
			t.Fatalf("missing source %s in row %v", name, rows[i+2])
		}
		if _, err := os.Stat(filepath.Join(out, "screens", fmt.Sprintf("TXT-H%02d.LBM.png", i))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "screens", "TXT-H46.LBM.png")); !os.IsNotExist(err) {
		t.Fatal("scanned beyond fixed briefing corpus")
	}
	manifest, err := os.ReadFile(filepath.Join(out, "source-sha256.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), fmt.Sprintf("%x  TXT/H45.LBM\n", sha256.Sum256(b))) {
		t.Fatal("briefing source hash missing")
	}
	gallery, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`value="briefings"`, "screens/TXT-H45.LBM.png", "TXT/H45.LBM", "CMAP", `id="briefings-TXT-H45.LBM"`} {
		if !strings.Contains(string(gallery), want) {
			t.Errorf("briefing gallery missing %q", want)
		}
	}
}
func TestCLIBriefingsPrevalidateLateInputAndPixelLimit(t *testing.T) {
	for _, name := range []string{"late malformed", "pixel cap"} {
		t.Run(name, func(t *testing.T) {
			root := testRoot(t)
			t.Chdir(root)
			data := syntheticData(t, root)
			b := screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0})
			if name == "pixel cap" {
				// Four 128-byte zero repeats per plane row, 1024 rows, four megapixels.
				body := bytes.Repeat([]byte{129, 0, 129, 0, 129, 0, 129, 0}, 1024)
				b = screenFixture(screenHeader(4096, 1024, 1, 0, 1), body)
			}
			briefingData(t, data, b)
			if name == "late malformed" {
				if err := os.WriteFile(filepath.Join(data, "TXT", "H45.LBM"), []byte("bad"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := run([]string{"-data", data, "-libraries", "PARTS", "-briefings"}, &bytes.Buffer{})
			if err == nil {
				t.Fatal("accepted invalid briefing selection")
			}
			if name == "late malformed" && !strings.Contains(err.Error(), "TXT/H45.LBM") {
				t.Fatalf("wrong input rejected: %v", err)
			}
			if name == "pixel cap" && !strings.Contains(err.Error(), "16 megapixel") {
				t.Fatalf("wrong cap failure: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "build", "native-reference")); !os.IsNotExist(err) {
				t.Fatal("invalid late input created output")
			}
		})
	}
}

func TestCLIRejectsScreenOutputCollisionBeforeWriting(t *testing.T) {
	for _, name := range []string{"TXT-H00.LBM", "txt-h00.lbm"} {
		t.Run(name, func(t *testing.T) {
			root := testRoot(t)
			t.Chdir(root)
			data := syntheticData(t, root)
			b := screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0})
			briefingData(t, data, b)
			if err := os.WriteFile(filepath.Join(data, name), b, 0600); err != nil {
				t.Fatal(err)
			}
			err := run([]string{"-data", data, "-libraries", "PARTS", "-screen", name, "-briefings"}, &bytes.Buffer{})
			if err == nil {
				t.Fatal("accepted colliding output names")
			}
			if !strings.Contains(err.Error(), "filename collision") {
				t.Fatalf("rejected for wrong reason: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "build", "native-reference")); !os.IsNotExist(err) {
				t.Fatal("filename collision created partial output")
			}
		})
	}
}
