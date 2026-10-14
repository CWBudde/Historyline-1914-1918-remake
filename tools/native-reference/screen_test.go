package main

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"testing"
)

func iffChunk(id string, b []byte) []byte {
	out := append([]byte(id), 0, 0, 0, 0)
	binary.BigEndian.PutUint32(out[4:8], uint32(len(b)))
	out = append(out, b...)
	if len(b)%2 != 0 {
		out = append(out, 0)
	}
	return out
}
func iffForm(chunks ...[]byte) []byte {
	b := []byte("ILBM")
	for _, c := range chunks {
		b = append(b, c...)
	}
	return iffChunk("FORM", b)
}
func screenHeader(w, h int, planes, mask, compression byte) []byte {
	b := make([]byte, 20)
	binary.BigEndian.PutUint16(b[0:2], uint16(w))
	binary.BigEndian.PutUint16(b[2:4], uint16(h))
	b[8], b[9], b[10], b[14], b[15] = planes, mask, compression, 10, 11
	binary.BigEndian.PutUint16(b[16:18], 320)
	binary.BigEndian.PutUint16(b[18:20], 200)
	return b
}
func screenFixture(hdr, body []byte) []byte {
	// Deliberately preserve low CMAP nibbles rather than reusing 00.PAL.
	return iffForm(iffChunk("BMHD", hdr), iffChunk("ANNO", []byte("odd")), iffChunk("CMAP", []byte{0, 0, 0, 19, 37, 55, 99, 77, 66, 200, 210, 220}), iffChunk("BODY", body))
}
func TestScreenInterleavedPlanesAndIntrinsicAspect(t *testing.T) {
	// Row 0 indices 1,2,3; row 1 indices 2,1,0. Plane 0 is low bit.
	s, err := decodeScreen(screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0}))
	if err != nil {
		t.Fatal(err)
	}
	if s.Width != 3 || s.Height != 2 || s.RowBytes != 2 || s.AspectX != 10 || s.AspectY != 11 || s.PlanarBytes != 8 || s.BodyBytes != 8 || s.PaletteColors != 4 {
		t.Fatalf("bad metadata: %+v", s)
	}
	want := []color.NRGBA{{19, 37, 55, 255}, {99, 77, 66, 255}, {200, 210, 220, 255}, {99, 77, 66, 255}, {19, 37, 55, 255}, {0, 0, 0, 255}}
	for i, c := range want {
		if got := s.Image.NRGBAAt(i%3, i/3); got != c {
			t.Fatalf("pixel %d=%v want %v", i, got, c)
		}
	}
	if s.Image.Bounds().Dx() != 3 || s.Image.Bounds().Dy() != 2 {
		t.Fatal("aspect resized intrinsic image")
	}
}
func TestScreenByteRun1RowsAndCoverage(t *testing.T) {
	for _, tt := range []struct {
		name        string
		mask        byte
		body        []byte
		transparent uint16
		alpha       []byte
	}{
		{"literal repeat noop and mask", 1, []byte{128, 1, 0x80, 0, 255, 0, 1, 0x80, 0}, 0, []byte{255, 0}},
		{"transparent color", 2, []byte{1, 0x80, 0, 255, 0}, 1, []byte{0, 255}},
		{"opaque zero", 0, []byte{1, 0x80, 0, 255, 0}, 0, []byte{255, 255}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hdr := screenHeader(2, 1, 2, tt.mask, 1)
			binary.BigEndian.PutUint16(hdr[12:14], tt.transparent)
			s, err := decodeScreen(screenFixture(hdr, tt.body))
			if err != nil {
				t.Fatal(err)
			}
			for x, a := range tt.alpha {
				if got := s.Image.NRGBAAt(x, 0).A; got != a {
					t.Fatalf("alpha %d=%d want %d", x, got, a)
				}
			}
		})
	}
}
func TestScreenRejectsMalformedAndUnsupported(t *testing.T) {
	hdr := screenHeader(2, 1, 2, 0, 0)
	good := screenFixture(hdr, []byte{0, 0, 0, 0})
	replaceHeader := func(off int, v byte) []byte {
		h := bytes.Clone(hdr)
		h[off] = v
		return screenFixture(h, []byte{0, 0, 0, 0})
	}
	badLen := bytes.Clone(good)
	binary.BigEndian.PutUint32(badLen[4:8], 0xffffffff)
	chunkLen := bytes.Clone(good)
	binary.BigEndian.PutUint32(chunkLen[16:20], 0xffffffff)
	cases := map[string][]byte{
		"custom record": graphicFixture(1, 1, 1, nil, []byte{128, 0}), "short form": good[:11], "form size": badLen, "chunk bounds": chunkLen, "trailing form": append(bytes.Clone(good), 0), "missing pad": iffForm(append([]byte("ANNO\x00\x00\x00\x01"), 'a')),
		"missing BMHD": iffForm(iffChunk("BODY", nil)), "missing CMAP": iffForm(iffChunk("BMHD", hdr), iffChunk("BODY", []byte{0, 0, 0, 0})), "duplicate BMHD": iffForm(iffChunk("BMHD", hdr), good[12:]), "short BMHD": screenFixture(hdr[:19], nil), "lasso": replaceHeader(9, 3), "compression": replaceHeader(10, 2), "too many planes": replaceHeader(8, 9), "zero planes": replaceHeader(8, 0), "short body": screenFixture(hdr, []byte{0}), "excess body": screenFixture(hdr, []byte{0, 0, 0, 0, 0}), "pixel cap": screenFixture(screenHeader(65535, 65535, 2, 0, 0), nil),
		"short literal": screenFixture(screenHeader(2, 1, 2, 0, 1), []byte{1, 0}), "short repeat": screenFixture(screenHeader(2, 1, 2, 0, 1), []byte{255}), "run crosses row": screenFixture(screenHeader(2, 1, 2, 0, 1), []byte{253, 0}), "noop truncation": screenFixture(screenHeader(2, 1, 2, 0, 1), []byte{128}),
		"HAM":                   iffForm(iffChunk("BMHD", hdr), iffChunk("CMAP", []byte{0, 0, 0}), iffChunk("CAMG", []byte{0, 0, 8, 0}), iffChunk("BODY", []byte{0, 0, 0, 0})),
		"EHB":                   iffForm(iffChunk("BMHD", hdr), iffChunk("CMAP", []byte{0, 0, 0}), iffChunk("CAMG", []byte{0, 0, 0, 128}), iffChunk("BODY", []byte{0, 0, 0, 0})),
		"DEST":                  iffForm(iffChunk("DEST", make([]byte, 8)), good[12:]),
		"bad CMAP":              iffForm(iffChunk("BMHD", hdr), iffChunk("CMAP", []byte{0}), iffChunk("BODY", []byte{0, 0, 0, 0})),
		"index outside palette": iffForm(iffChunk("BMHD", hdr), iffChunk("CMAP", []byte{0, 0, 0}), iffChunk("BODY", []byte{128, 0, 0, 0})),
	}
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeScreen(b); err == nil {
				t.Fatal("accepted invalid/unsupported screen")
			}
		})
	}
}

func TestScreenOuterAmigaTPWM(t *testing.T) {
	plain := screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0})
	packed := append([]byte("TPWM"), 0, 0, 0, 0)
	binary.BigEndian.PutUint32(packed[4:8], uint32(len(plain)))
	for i := 0; i < len(plain); i += 8 {
		packed = append(packed, 0)
		packed = append(packed, plain[i:min(i+8, len(plain))]...)
	}
	s, err := decodeScreen(packed)
	if err != nil {
		t.Fatal(err)
	}
	if s.Image.NRGBAAt(0, 0) != (color.NRGBA{19, 37, 55, 255}) {
		t.Fatal("outer compression changed screen")
	}
	if _, err := decodeScreen(append(packed, 0)); err == nil {
		t.Fatal("trailing outer TPWM bytes accepted")
	}
	little := bytes.Clone(packed)
	binary.LittleEndian.PutUint32(little[4:8], uint32(len(plain)))
	if _, err := decodeScreen(little); err == nil {
		t.Fatal("DOS TPWM accepted")
	}
}

func FuzzScreen(f *testing.F) {
	f.Add(screenFixture(screenHeader(3, 2, 2, 0, 0), []byte{0xa0, 0, 0x60, 0, 0x40, 0, 0x80, 0}))
	f.Add(screenFixture(screenHeader(2, 1, 2, 1, 1), []byte{128, 1, 128, 0, 255, 0, 1, 128, 0}))
	f.Add([]byte("TPWM\x00\x00\x00\x03\x00ABC"))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 4096 {
			t.Skip()
		}
		_, _ = decodeScreen(b)
	})
}
