package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func maskFixture() ([]byte, []byte) {
	widths := []int{8, 10, 10, 12, 14, 16, 16, 18, 20, 22, 22, 24, 24, 22, 22, 20, 18, 16, 16, 14, 12, 10, 10, 8}
	m := make([]byte, 96)
	for y, w := range widths {
		for x := (24 - w) / 2; x < (24+w)/2; x++ {
			m[y*4+x/8] |= 128 >> uint(x%8)
		}
	}
	inv := make([]byte, 96)
	for i := range m {
		inv[i] = ^m[i]
	}
	return m, inv
}
func TestNativeHexShapeAndComplement(t *testing.T) {
	m, inv := maskFixture()
	got, err := validateMasks(m, inv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, m) {
		t.Fatal("mask altered")
	}
	for _, tt := range []struct {
		name   string
		m, inv []byte
	}{{"truncated", m[:95], inv}, {"color mask instead of native hex", make([]byte, 96), inv}, {"wrong inverse", m, m}} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := validateMasks(tt.m, tt.inv); err == nil {
				t.Fatal("accepted invalid native mask")
			}
		})
	}
}
func testRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/CWBudde/Historyline-1914-1918-remake\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	return root
}
func TestOutputPolicy(t *testing.T) {
	root := testRoot(t)
	for _, tt := range []struct {
		name, path string
		good       bool
	}{{"default", "build/native-reference", true}, {"private originals", "original/reference/new", true}, {"tracked source", "tools/pngs", false}, {"root original", "original", false}, {"escape", "../outside", false}} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := privateOutput(root, tt.path)
			if (err == nil) != tt.good {
				t.Fatalf("path %s error=%v", tt.path, err)
			}
		})
	}
	if err := os.Mkdir(filepath.Join(root, "build"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "build", "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := privateOutput(root, "build/escape/new"); err == nil {
		t.Fatal("symlink escapes private root")
	}
}

func TestCLIExportsSyntheticNativeData(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	data := filepath.Join(root, "original", "input")
	if err := os.MkdirAll(filepath.Join(data, "LIB"), 0700); err != nil {
		t.Fatal(err)
	}
	m, inv := maskFixture()
	files := map[string][]byte{"00.PAL": paletteFixture(), "HEXMSK00.DAT": m, "HEXMSK20.DAT": inv, "LIB/PARTS.LIB": libraryFixture("SYNTH001", graphicFixture(24, 24, 1, nil, make([]byte, 96)))}
	for name, b := range files {
		if err := os.WriteFile(filepath.Join(data, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var log bytes.Buffer
	if err := run([]string{"-data", data, "-libraries", "PARTS"}, &log); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "build", "native-reference")
	f, err := os.Open(filepath.Join(out, "PARTS", "SYNTH001.png"))
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(f)
	if closeErr := f.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 24 || im.Bounds().Dy() != 24 {
		t.Fatal("wrong PNG dimensions")
	}
	_, _, _, inside := im.At(8, 0).RGBA()
	_, _, _, outside := im.At(0, 0).RGBA()
	if inside != 65535 || outside != 0 {
		t.Fatal("native hex alpha not exported")
	}
	f, err = os.Open(filepath.Join(out, "metadata.csv"))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(f).ReadAll()
	if closeErr := f.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][0] != "PARTS" || rows[1][1] != "SYNTH001" {
		t.Fatalf("invalid metadata %v", rows)
	}
	manifest, err := os.ReadFile(filepath.Join(out, "source-sha256.txt"))
	if err != nil {
		t.Fatal(err)
	}
	sha := sha256.Sum256(files["LIB/PARTS.LIB"])
	if !bytes.Contains(manifest, []byte(hex.EncodeToString(sha[:])+"  LIB/PARTS.LIB")) {
		t.Fatalf("manifest misses original source hash: %s", manifest)
	}
	before, err := os.ReadFile(filepath.Join(out, "PARTS", "SYNTH001.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-data", data, "-libraries", "PARTS"}, &log); err == nil {
		t.Fatal("existing output was overwritten")
	}
	after, err := os.ReadFile(filepath.Join(out, "PARTS", "SYNTH001.png"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("existing artifact changed")
	}
}
func TestCLIRejectsDOSBeforeCreatingOutput(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	data := filepath.Join(root, "original", "dos")
	if err := os.MkdirAll(data, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(data, "00.PAL"), make([]byte, 768), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-data", data}, &bytes.Buffer{}); err == nil {
		t.Fatal("DOS input accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "build", "native-reference")); !os.IsNotExist(err) {
		t.Fatal("invalid input created output")
	}
}

func TestPackedGraphicAlignment(t *testing.T) {
	r := graphicFixture(1, 1, 1, nil, []byte{128, 0})
	stream := append([]byte("TPWM"), 0, 0, 0, byte(len(r)))
	for i := 0; i < len(r); i += 8 {
		stream = append(stream, 0)
		stream = append(stream, r[i:min(i+8, len(r))]...)
	}
	if len(stream)%2 == 1 {
		stream = append(stream, 0x91)
	}
	pal, _ := parsePalette(paletteFixture())
	if _, err := decodeGraphic(stream, pal, nil); err != nil {
		t.Fatal(err)
	}
	stream = append(stream, 0)
	if _, err := decodeGraphic(stream, pal, nil); err == nil {
		t.Fatal("extra packed bytes accepted")
	}
}
func TestArchiveDuplicateOffsetsAndNames(t *testing.T) {
	b := libraryFixture("SYNTH001", []byte("record"))
	b = append(b, b[len(b)-12:]...)
	if _, err := parseLibrary(b); err == nil {
		t.Fatal("duplicate names accepted")
	}
	copy(b[len(b)-12:], "SYNTH002")
	if _, err := parseLibrary(b); err == nil {
		t.Fatal("duplicate offsets accepted")
	}
	binary.BigEndian.PutUint32(b[:4], 0)
	if _, err := parseLibrary(b); err == nil {
		t.Fatal("zero FAT accepted")
	}
}
