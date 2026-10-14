package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"testing"
)

// Expectations are hand-derived: three literals plus a distance-three,
// six-byte copy must produce ABCABCABC, including an overlapping copy.
func TestUnpackBE(t *testing.T) {
	tests := []struct {
		name        string
		input, want []byte
		bad         bool
	}{
		{"literal", append([]byte{'T', 'P', 'W', 'M', 0, 0, 0, 3, 0}, []byte("ABC")...), []byte("ABC"), false},
		{"backreference", []byte{'T', 'P', 'W', 'M', 0, 0, 0, 9, 0x10, 'A', 'B', 'C', 3, 3}, []byte("ABCABCABC"), false},
		{"overlapping", []byte{'T', 'P', 'W', 'M', 0, 0, 0, 7, 0x40, 'x', 3, 1}, []byte("xxxxxxx"), false},
		{"uncompressed", []byte("ILBM    "), []byte("ILBM    "), false},
		{"header truncated", []byte("TPWM\x00"), nil, true},
		{"literal truncated", []byte{'T', 'P', 'W', 'M', 0, 0, 0, 1, 0}, nil, true},
		{"reference truncated", []byte{'T', 'P', 'W', 'M', 0, 0, 0, 3, 128, 0}, nil, true},
		{"distance zero", []byte{'T', 'P', 'W', 'M', 0, 0, 0, 3, 128, 0, 0}, nil, true},
		{"distance before output", []byte{'T', 'P', 'W', 'M', 0, 0, 0, 3, 128, 0, 1}, nil, true},
		{"oversized declaration", []byte{'T', 'P', 'W', 'M', 0x7f, 0xff, 0xff, 0xff}, nil, true},
		{"DOS little endian unsupported", []byte{'T', 'P', 'W', 'M', 3, 0, 0, 0, 0, 'A', 'B', 'C'}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, used, err := unpackBE(tt.input)
			if tt.bad {
				if err == nil {
					t.Fatal("accepted malformed stream")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) || used != len(tt.input) {
				t.Fatalf("got %q consumed %d; want %q consumed %d", got, used, tt.want, len(tt.input))
			}
		})
	}
}

func paletteFixture() []byte {
	p := make([]byte, 64)
	binary.BigEndian.PutUint16(p[2:], 0x0a53)
	binary.BigEndian.PutUint16(p[32:], 0x08a0)
	return p
}
func TestPaletteHardwareWords(t *testing.T) {
	p, err := parsePalette(paletteFixture())
	if err != nil {
		t.Fatal(err)
	}
	if p[1] != (color.NRGBA{170, 85, 51, 255}) || p[16] != (color.NRGBA{136, 170, 0, 255}) {
		t.Fatalf("wrong palette %v %v", p[1], p[16])
	}
	for _, b := range [][]byte{make([]byte, 768), make([]byte, 63), append([]byte{0xf0}, make([]byte, 63)...)} {
		if _, err := parsePalette(b); err == nil {
			t.Fatal("accepted non-native palette")
		}
	}
}

// A sparse p0+p4 image checks plane significance, left-to-right bits,
// word padding, and opaque black. Padding bits must not become pixels.
func graphicFixture(w, h int, pick uint16, coverage, pixels []byte) []byte {
	b := make([]byte, 24)
	copy(b, "ILBM    ")
	binary.BigEndian.PutUint16(b[8:], pick)
	binary.BigEndian.PutUint16(b[10:], uint16(24+len(coverage)))
	binary.BigEndian.PutUint16(b[18:], uint16((w+15)/16))
	binary.BigEndian.PutUint16(b[20:], uint16(w))
	binary.BigEndian.PutUint16(b[22:], uint16(h))
	b = append(b, coverage...)
	return append(b, pixels...)
}
func TestSparsePlanesAndRowPadding(t *testing.T) {
	pal, _ := parsePalette(paletteFixture())
	b := graphicFixture(17, 1, 0x11, nil, []byte{0x80, 0, 0x80, 0xff, 0x40, 0, 0, 0xff})
	g, err := decodeGraphic(b, pal, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.Image.Bounds().Dx() != 17 {
		t.Fatal("padding increased width")
	}
	for _, tt := range []struct {
		x    int
		want color.NRGBA
	}{{0, pal[1]}, {1, pal[16]}, {2, pal[0]}, {16, pal[1]}} {
		if got := g.Image.NRGBAAt(tt.x, 0); got != tt.want {
			t.Fatalf("pixel %d=%v want %v", tt.x, got, tt.want)
		}
	}
}
func TestFivePlanesPreserveOrder(t *testing.T) {
	pal, _ := parsePalette(paletteFixture())
	pal[31] = color.NRGBA{255, 255, 255, 255}
	b := graphicFixture(1, 1, 31, nil, []byte{128, 0, 128, 0, 128, 0, 128, 0, 128, 0})
	g, err := decodeGraphic(b, pal, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.Image.NRGBAAt(0, 0) != pal[31] {
		t.Fatal("five planes did not select color31")
	}
}
func TestEmbeddedMaskPreservesBlack(t *testing.T) {
	pal, _ := parsePalette(paletteFixture())
	g, err := decodeGraphic(graphicFixture(3, 1, 1, []byte{0xc0, 0}, []byte{0x40, 0}), pal, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.Image.NRGBAAt(0, 0).A != 255 || g.Image.NRGBAAt(0, 0).R != 0 || g.Image.NRGBAAt(2, 0).A != 0 {
		t.Fatal("black coverage or transparent outside lost")
	}
}
func TestExternalHexMaskPreservesBlack(t *testing.T) {
	pal, _ := parsePalette(paletteFixture())
	mask := make([]byte, 96)
	mask[0] = 0x80
	g, err := decodeGraphic(graphicFixture(24, 24, 1, nil, make([]byte, 96)), pal, mask)
	if err != nil {
		t.Fatal(err)
	}
	if g.Image.NRGBAAt(0, 0) != (color.NRGBA{0, 0, 0, 255}) || g.Image.NRGBAAt(1, 0).A != 0 {
		t.Fatal("mask must define alpha independently of color")
	}
}
func TestGraphicRejectsMalformed(t *testing.T) {
	pal, _ := parsePalette(paletteFixture())
	good := graphicFixture(1, 1, 1, nil, []byte{128, 0})
	tests := []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"short header", func(b []byte) []byte { return b[:23] }},
		{"signature", func(b []byte) []byte { b[0] = 'X'; return b }},
		{"truncated planes", func(b []byte) []byte { return b[:len(b)-1] }},
		{"extra planes", func(b []byte) []byte { return append(b, 0) }},
		{"zero width", func(b []byte) []byte { b[20] = 0; b[21] = 0; return b }},
		{"bad stride", func(b []byte) []byte { b[19] = 2; return b }},
		{"bad data offset", func(b []byte) []byte { b[11] = 25; return b }},
		{"unsupported plane", func(b []byte) []byte { b[8] = 1; return b }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := decodeGraphic(tt.mutate(bytes.Clone(good)), pal, nil); err == nil {
				t.Fatal("accepted malformed graphic")
			}
		})
	}
	if _, err := decodeGraphic(good, pal, []byte{1}); err == nil {
		t.Fatal("accepted external mask for wrong dimensions")
	}
}

func libraryFixture(name string, record []byte) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(4+len(record)))
	b = append(b, record...)
	entry := make([]byte, 12)
	copy(entry, name)
	binary.BigEndian.PutUint32(entry[8:], 4)
	return append(b, entry...)
}
func TestNativeArchiveIndex(t *testing.T) {
	record := graphicFixture(1, 1, 1, nil, []byte{128, 0})
	got, err := parseLibrary(libraryFixture("SYNTH001", record))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "SYNTH001" || got[0].Offset != 4 || !bytes.Equal(got[0].Data, record) {
		t.Fatalf("wrong archive result %+v", got)
	}
}
func TestNativeArchiveRejectsBadOffsetsNamesAndLengths(t *testing.T) {
	good := libraryFixture("SYNTH001", []byte("record"))
	tests := []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"short", func(b []byte) []byte { return b[:3] }},
		{"FAT outside", func(b []byte) []byte { binary.BigEndian.PutUint32(b, 999); return b }},
		{"partial FAT", func(b []byte) []byte { return b[:len(b)-1] }},
		{"resource points into FAT", func(b []byte) []byte { binary.BigEndian.PutUint32(b[len(b)-4:], 10); return b }},
		{"resource points into header", func(b []byte) []byte { binary.BigEndian.PutUint32(b[len(b)-4:], 0); return b }},
		{"path traversal", func(b []byte) []byte { copy(b[10:], "../evil!"); return b }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseLibrary(tt.mutate(bytes.Clone(good))); err == nil {
				t.Fatal("accepted malformed archive")
			}
		})
	}
}

// Index-size validation prevents a small malformed archive from expanding
// into an unbounded collection of names and slices before record validation.
func TestArchiveRejectsExcessiveResourceCount(t *testing.T) {
	const count = 10001
	b := make([]byte, 4+count+12*count)
	binary.BigEndian.PutUint32(b, uint32(4+count))
	for i := 0; i < count; i++ {
		pos := 4 + count + i*12
		copy(b[pos:pos+8], fmt.Sprintf("%08d", i))
		binary.BigEndian.PutUint32(b[pos+8:pos+12], uint32(4+i))
	}
	if _, err := parseLibrary(b); err == nil {
		t.Fatal("accepted excessive resource count")
	}
}

func FuzzNativeFormats(f *testing.F) {
	f.Add([]byte("TPWM\x00\x00\x00\x03\x00ABC"))
	f.Add(libraryFixture("SYNTH001", graphicFixture(1, 1, 1, nil, []byte{128, 0})))
	f.Add(graphicFixture(1, 1, 1, []byte{128, 0}, []byte{0, 0}))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 4096 {
			t.Skip()
		}
		_, _, _ = unpackBE(b)
		_, _ = parseLibrary(b)
		_, _ = parsePalette(b)
		_, _ = decodeGraphic(b, [32]color.NRGBA{}, nil)
	})
}
