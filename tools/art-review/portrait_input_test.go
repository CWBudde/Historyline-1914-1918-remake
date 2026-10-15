package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPortraitStrictManifest(t *testing.T) {
	path, manifest, _ := portraitTestFixture(t, 840, 752)
	for _, tc := range []struct{ name, old, replacement string }{
		{"family", `"infantry"`, `"eliteInfantry"`},
		{"case family", `"infantry"`, `"Infantry"`},
		{"version", `"schema_version":1`, `"schema_version":2`},
		{"fraction version", `"schema_version":1`, `"schema_version":1.0`},
		{"exponent version", `"schema_version":1`, `"schema_version":1e0`},
		{"case field", `"family"`, `"Family"`},
		{"unknown", `"family":`, `"unknown":0,"family":`},
		{"duplicate", `"family":`, `"family":"infantry","family":`},
		{"null family", `"family":"infantry"`, `"family":null`},
		{"null source", `"source":{`, `"source":null,"ignored":{`},
		{"null pin", `"source_sha256":`, `"source_sha256":null,"ignored":`},
		{"source case", `"source_sha256"`, `"Source_SHA256"`},
		{"nested duplicate", `"source":"source.png"`, `"source":"source.png","source":"source.png"`},
		{"missing source", `"source":"source.png",`, ``},
		{"null version", `"schema_version":1`, `"schema_version":null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := strings.Replace(string(manifest), tc.old, tc.replacement, 1)
			if bad == string(manifest) {
				t.Fatal("fixture replacement did not apply")
			}
			unitTestWrite(t, path, []byte(bad))
			out := filepath.Join(t.TempDir(), "absent")
			if run([]string{"-portrait", path, "-out", out}, io.Discard) == nil {
				t.Fatal("invalid manifest accepted")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("validation created output")
			}
		})
	}
	for _, bad := range [][]byte{[]byte("null"), append(append([]byte(nil), manifest...), manifest...), []byte(`{"schema_version":1,"family":"infantry"}`)} {
		unitTestWrite(t, path, bad)
		if _, err := loadPortraitInputs(path); err == nil {
			t.Fatal("invalid top-level document accepted")
		}
	}
}

func portraitTestReplaceSource(t *testing.T, path string, source []byte) {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(unitTestRead(t, path), &m); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(source)
	m["source"] = map[string]string{"source": "source.png", "source_sha256": hex.EncodeToString(sum[:])}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	unitTestWrite(t, path, b)
	unitTestWrite(t, filepath.Join(filepath.Dir(path), "source.png"), source)
}

func TestPortraitRejectsUnsafeSourceAndPins(t *testing.T) {
	path, manifest, source := portraitTestFixture(t, 840, 752)
	for _, name := range []string{"../source.png", "/tmp/source.png", "./source.png", "nested/../source.png", "original/source.png", "pilot/source.png", `nested\source.png`} {
		unitTestWrite(t, path, bytes.Replace(manifest, []byte(`"source.png"`), []byte(`"`+strings.ReplaceAll(name, `\`, `\\`)+`"`), 1))
		if _, err := loadPortraitInputs(path); err == nil {
			t.Fatalf("unsafe path accepted: %s", name)
		}
	}
	unitTestWrite(t, path, manifest)
	unitTestWrite(t, filepath.Join(filepath.Dir(path), "source.png"), append(source, '\n'))
	if _, err := loadPortraitInputs(path); err == nil {
		t.Fatal("changed source pin accepted")
	}
	unitTestWrite(t, filepath.Join(filepath.Dir(path), "source.png"), source)
	sum := sha256.Sum256(source)
	unitTestWrite(t, path, bytes.Replace(manifest, []byte(hex.EncodeToString(sum[:])), []byte(strings.ToUpper(hex.EncodeToString(sum[:]))), 1))
	if _, err := loadPortraitInputs(path); err == nil {
		t.Fatal("noncanonical pin accepted")
	}
	unitTestWrite(t, path, manifest)
	outside := filepath.Join(t.TempDir(), "external.png")
	unitTestWrite(t, outside, source)
	if err := os.Remove(filepath.Join(filepath.Dir(path), "source.png")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(filepath.Dir(path), "source.png")); err != nil {
		t.Fatal(err)
	}
	if _, err := loadPortraitInputs(path); err == nil {
		t.Fatal("escaping source symlink accepted")
	}
}

func TestPortraitRejectsSmallCropAndDiscardedTransparency(t *testing.T) {
	for _, size := range [][2]int{{839, 1000}, {1000, 751}, {839, 751}, {105, 94}} {
		path, _, _ := portraitTestFixture(t, size[0], size[1])
		if _, err := loadPortraitInputs(path); err == nil {
			t.Fatalf("undersized exact crop accepted: %v", size)
		}
	}
	path, _, _ := portraitTestFixture(t, 1024, 1024)
	im := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	for y := range 1024 {
		for x := range 1024 {
			im.SetNRGBA(x, y, color.NRGBA{40, 80, 120, 255})
		}
	}
	for _, alpha := range []uint8{0, 254} {
		im.SetNRGBA(0, 0, color.NRGBA{40, 80, 120, alpha})
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, im); err != nil {
			t.Fatal(err)
		}
		portraitTestReplaceSource(t, path, encoded.Bytes())
		if _, err := loadPortraitInputs(path); err == nil {
			t.Fatalf("discarded overscan alpha %d accepted", alpha)
		}
	}
}

func TestPortraitRejectsPNGPixelBudgetBeforeDecode(t *testing.T) {
	path, _, source := portraitTestFixture(t, 840, 752)
	patched := append([]byte(nil), source...)
	binary.BigEndian.PutUint32(patched[16:20], 8193)
	binary.BigEndian.PutUint32(patched[20:24], 8192)
	binary.BigEndian.PutUint32(patched[29:33], crc32.ChecksumIEEE(patched[12:29]))
	portraitTestReplaceSource(t, path, patched)
	if _, err := loadPortraitInputs(path); err == nil || !strings.Contains(err.Error(), "64 megapixels") {
		t.Fatalf("oversized PNG must fail config budget before allocation: %v", err)
	}
}

func TestPortraitLargestExactCropOddCenterAndMinimum(t *testing.T) {
	for _, tc := range []struct {
		width, height int
		want          packCrop
	}{
		{840, 752, packCrop{Width: 840, Height: 752}},
		{1024, 1025, packCrop{X: 39, Y: 89, Width: 945, Height: 846}},
		{1261, 1129, packCrop{Width: 1260, Height: 1128}},
		{1000, 800, packCrop{X: 80, Y: 24, Width: 840, Height: 752}},
	} {
		path, _, _ := portraitTestFixture(t, tc.width, tc.height)
		in, err := loadPortraitInputs(path)
		if err != nil {
			t.Fatal(err)
		}
		if in.Crop != tc.want {
			t.Fatalf("%dx%d: got %+v want %+v", tc.width, tc.height, in.Crop, tc.want)
		}
	}
}

func TestPortraitRejectsNonopaque16BitOverscan(t *testing.T) {
	path, _, _ := portraitTestFixture(t, 1024, 1024)
	im := portraitTestOpaque16BitImage()
	// This discarded overscan pixel rounds to alpha 255 if converted to
	// NRGBA8 first, but is not fully opaque in the source PNG.
	im.SetNRGBA64(0, 0, color.NRGBA64{R: 0x2f2f, G: 0x7a7a, B: 0xcfcf, A: 65534})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, im); err != nil {
		t.Fatal(err)
	}
	if encoded.Bytes()[24] != 16 || encoded.Bytes()[25] != 6 {
		t.Fatal("regression fixture requires 16-bit RGBA PNG")
	}
	portraitTestReplaceSource(t, path, encoded.Bytes())
	out := filepath.Join(t.TempDir(), "absent")
	if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err == nil {
		t.Fatal("16-bit nonopaque source alpha 65534 was rounded to opaque and accepted")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("rejected 16-bit source created output")
	}
}

func portraitTestOpaque16BitImage() *image.NRGBA64 {
	im := image.NewNRGBA64(image.Rect(0, 0, 1024, 1024))
	for y := range 1024 {
		for x := range 1024 {
			im.SetNRGBA64(x, y, color.NRGBA64{R: 0x2f2f, G: 0x7a7a, B: 0xcfcf, A: 65535})
		}
	}
	return im
}

func TestPortraitAcceptsFullyOpaque16BitSource(t *testing.T) {
	path, _, _ := portraitTestFixture(t, 1024, 1024)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, portraitTestOpaque16BitImage()); err != nil {
		t.Fatal(err)
	}
	if encoded.Bytes()[24] != 16 {
		t.Fatal("fixture requires 16-bit PNG")
	}
	portraitTestReplaceSource(t, path, encoded.Bytes())
	out := filepath.Join(t.TempDir(), "bundle")
	if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err != nil {
		t.Fatalf("fully opaque 16-bit source rejected: %v", err)
	}
	if !bytes.Equal(unitTestRead(t, filepath.Join(out, "infantry-source.png")), encoded.Bytes()) {
		t.Fatal("16-bit source snapshot changed")
	}
	for _, name := range []string{"infantry-1x.png", "infantry-2x.png", "infantry-4x.png"} {
		im := unitTestLoad(t, filepath.Join(out, name))
		for _, p := range []image.Point{{0, 0}, {im.Bounds().Dx() / 2, im.Bounds().Dy() / 2}, {im.Bounds().Dx() - 1, im.Bounds().Dy() - 1}} {
			if got := im.NRGBAAt(p.X, p.Y); got != (color.NRGBA{R: 0x2f, G: 0x7a, B: 0xcf, A: 255}) {
				t.Fatalf("opaque 16-bit conversion/resampling changed at %s %v: %v", name, p, got)
			}
		}
	}
}
