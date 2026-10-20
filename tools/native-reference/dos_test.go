package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func dosTerrainFixture() []byte {
	b := make([]byte, 18)
	copy(b, "INFOILBM")
	b[8], b[9] = 4, 0x55
	binary.LittleEndian.PutUint16(b[14:], 8)
	binary.LittleEndian.PutUint16(b[16:], 2)
	// Four VGA x-mod-4 lanes, each holding two rows of two pixels.
	return append(b, 0, 200, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14)
}

// FIN indices must follow PARTS.DAT rather than archive order; DOS TPWM and
// FAT words are little-endian while the FIN dimension words stay big-endian.
func TestDOSCLIReconstructsIndexedGround(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	input := filepath.Join(root, "original", "dos")
	if err := os.MkdirAll(filepath.Join(input, "LIB"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(input, "MAP"), 0755); err != nil {
		t.Fatal(err)
	}
	pal := make([]byte, 768)
	for i := range 256 {
		pal[3*i], pal[3*i+1], pal[3*i+2] = byte(i), byte(255-i), byte(i/2)
	}
	lib := make([]byte, 4)
	var fat, names []byte
	mask, _ := maskFixture()
	for i := 174; i >= 0; i-- {
		name := fmt.Sprintf("SYNTH%03d", i)
		entry := append([]byte(name), 0, 0, 0, 0)
		binary.LittleEndian.PutUint32(entry[8:], uint32(len(lib)))
		fat = append(fat, entry...)
		r := make([]byte, 594)
		copy(r, "INFOILBM")
		r[8], r[9] = 4, 0x55
		binary.LittleEndian.PutUint16(r[14:], 24)
		binary.LittleEndian.PutUint16(r[16:], 24)
		for y := range 24 {
			for x := range 24 {
				v := byte(4)
				if mask[y*4+x/8]&(128>>uint(x%8)) != 0 {
					v = byte(i + 50)
				}
				r[18+(x%4)*144+y*6+x/4] = v
			}
		}
		lib = append(lib, r...)
	}
	binary.LittleEndian.PutUint32(lib, uint32(len(lib)))
	lib = append(lib, fat...)
	for i := range 175 {
		names = append(names, fmt.Sprintf("SYNTH%03d", i)...)
	}
	fin := []byte{0, 16, 0, 24}
	for i := range 16 * 24 {
		id := byte(0)
		if i == 1 {
			id = 1
		}
		fin = append(fin, id, 255)
	}
	for name, b := range map[string][]byte{"00.PAL": pal, "LIB/PARTS.LIB": lib, "LIB/PARTS.DAT": names, "MAP/00.FIN": fin} {
		if err := os.WriteFile(filepath.Join(input, filepath.FromSlash(name)), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"-platform", "dos", "-data", input, "-libraries", "PARTS", "-map", "00"}
	if err := run(args, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(root, "build/native-reference/reference.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest dosReference
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	cellBytes, err := os.ReadFile(filepath.Join(root, "build/native-reference/maps/00-PARTS-cells.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(cellBytes)); got != manifest.CellsSHA {
		t.Fatal("exported cell CSV differs from its provenance pin")
	}
	for _, tc := range []struct {
		file       string
		w, h, x, y int
	}{{"00-PARTS.png", 264, 588, 28, 24}, {"00-PARTS-aligned.png", 1090, 1920, 119, 80}} {
		b, err := os.ReadFile(filepath.Join(root, "build/native-reference/maps", tc.file))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if im.Bounds().Dx() != tc.w || im.Bounds().Dy() != tc.h {
			t.Errorf("%s geometry %v", tc.file, im.Bounds())
		}
		if got := color.NRGBAModel.Convert(im.At(tc.x, tc.y)); got != (color.NRGBA{R: 51, G: 204, B: 25, A: 255}) {
			t.Errorf("%s wrong indexed cell: %+v", tc.file, got)
		}
		if tc.file == "00-PARTS-aligned.png" {
			// The native masked lattice tessellates. Normalization must not
			// introduce transparent seams between its interior cells.
			for y := 160; y < 1760; y++ {
				for x := 200; x < 850; x++ {
					if _, _, _, a := im.At(x, y).RGBA(); a != 0xffff {
						t.Fatalf("aligned map introduced a seam at (%d,%d)", x, y)
					}
				}
			}
		}
	}
	if err := run(args, &bytes.Buffer{}); err == nil {
		t.Fatal("existing private export overwritten")
	}
}

func TestDOSTerrainLanesPaletteAndTransparency(t *testing.T) {
	var pal [256]color.NRGBA
	for i := range pal {
		pal[i] = color.NRGBA{R: byte(i), A: 255}
	}
	pal[200] = color.NRGBA{R: 243, G: 83, B: 195, A: 255}
	im, err := decodeDOSTerrain(dosTerrainFixture(), pal)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		x, y int
		want color.NRGBA
	}{
		{0, 0, color.NRGBA{A: 255}}, {4, 0, pal[200]}, {5, 0, color.NRGBA{}}, {0, 1, pal[1]}, {7, 1, pal[14]},
	} {
		if got := im.NRGBAAt(tc.x, tc.y); got != tc.want {
			t.Errorf("(%d,%d)=%+v want %+v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestDOSRejectsUnsupportedOrTruncatedTerrain(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func([]byte) []byte
	}{
		{"truncated", func(b []byte) []byte { return b[:len(b)-1] }},
		{"packed unit", func(b []byte) []byte { b[9] = 0x50; return b }},
		{"unaligned width", func(b []byte) []byte { b[14] = 7; return b }},
		{"unknown signature", func(b []byte) []byte { b[0] = 0; return b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decodeDOSTerrain(tc.change(dosTerrainFixture()), [256]color.NRGBA{}); err == nil {
				t.Fatal("malformed terrain accepted")
			}
		})
	}
}
