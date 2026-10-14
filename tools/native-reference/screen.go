package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
)

type screen struct {
	Image                                                     *image.NRGBA
	Width, Height, RowBytes, Planes, Masking, Compression     int
	TransparentColor, AspectX, AspectY, PageWidth, PageHeight int
	PaletteColors, BodyBytes, PlanarBytes                     int
	CAMG                                                      uint32
	Coverage                                                  string
}

// decodeScreen reads standalone, indexed FORM/ILBM screens. This is separate
// from the custom ILBM records in the native resource libraries.
func decodeScreen(input []byte) (screen, error) {
	var s screen
	if len(input) > maxRawBytes {
		return s, errors.New("screen input exceeds 32 MiB limit")
	}
	b, used, err := unpackBE(input)
	if err != nil {
		return s, err
	}
	if used != len(input) {
		return s, errors.New("trailing bytes after screen TPWM stream")
	}
	if len(b) < 12 || len(b) > maxRawBytes || string(b[:4]) != "FORM" || string(b[8:12]) != "ILBM" {
		return s, errors.New("expected bounded standalone FORM/ILBM")
	}
	if uint64(binary.BigEndian.Uint32(b[4:8]))+8 != uint64(len(b)) {
		return s, errors.New("FORM length does not match file")
	}
	chunks := make(map[string][]byte)
	bodySeen := false
	for pos := 12; pos < len(b); {
		if len(b)-pos < 8 {
			return s, errors.New("truncated IFF chunk header")
		}
		id := string(b[pos : pos+4])
		size := uint64(binary.BigEndian.Uint32(b[pos+4 : pos+8]))
		end := uint64(pos) + 8 + size
		next := end + (size & 1)
		if next > uint64(len(b)) {
			return s, errors.New("IFF chunk or padding exceeds FORM")
		}
		switch id {
		case "BMHD", "CMAP", "CAMG", "BODY":
			if _, exists := chunks[id]; exists {
				return s, fmt.Errorf("duplicate %s chunk", id)
			}
			if bodySeen {
				return s, errors.New("ILBM properties must precede BODY")
			}
			chunks[id] = b[pos+8 : int(end)]
			if id == "BODY" {
				bodySeen = true
			}
		case "DEST":
			return s, errors.New("ILBM DEST plane merging unsupported")
		}
		pos = int(next)
	}
	hdr := chunks["BMHD"]
	if len(hdr) != 20 {
		return s, errors.New("BMHD must contain 20 bytes")
	}
	word := func(off int) int { return int(binary.BigEndian.Uint16(hdr[off : off+2])) }
	s.Width, s.Height = word(0), word(2)
	s.Planes, s.Masking, s.Compression = int(hdr[8]), int(hdr[9]), int(hdr[10])
	s.TransparentColor, s.AspectX, s.AspectY = word(12), int(hdr[14]), int(hdr[15])
	s.PageWidth, s.PageHeight = int(int16(word(16))), int(int16(word(18)))
	if s.Width < 1 || s.Height < 1 || int64(s.Width)*int64(s.Height) > maxPixels {
		return s, errors.New("invalid screen dimensions or four megapixel limit exceeded")
	}
	if s.Planes < 1 || s.Planes > 8 {
		return s, errors.New("only indexed ILBM with 1-8 planes supported")
	}
	if s.Masking > 2 {
		return s, errors.New("ILBM lasso/unknown masking unsupported")
	}
	if s.Compression > 1 {
		return s, errors.New("only uncompressed or ByteRun1 ILBM supported")
	}
	if s.Masking == 2 && s.TransparentColor >= 1<<s.Planes {
		return s, errors.New("transparent color outside source plane range")
	}
	if camg, ok := chunks["CAMG"]; ok {
		if len(camg) != 4 {
			return s, errors.New("CAMG must contain four bytes")
		}
		s.CAMG = binary.BigEndian.Uint32(camg)
		if s.CAMG&(0x800|0x80) != 0 {
			return s, errors.New("ILBM HAM/EHB display modes unsupported")
		}
	}
	cmap := chunks["CMAP"]
	if len(cmap) == 0 || len(cmap)%3 != 0 || len(cmap) > 256*3 {
		return s, errors.New("CMAP must contain 1-256 RGB triples")
	}
	s.PaletteColors = len(cmap) / 3
	pal := make([]color.NRGBA, s.PaletteColors)
	for i := range pal {
		pal[i] = color.NRGBA{R: cmap[3*i], G: cmap[3*i+1], B: cmap[3*i+2], A: 255}
	}
	body, ok := chunks["BODY"]
	if !ok {
		return s, errors.New("missing BODY")
	}
	s.BodyBytes = len(body)
	s.RowBytes = ((s.Width + 15) / 16) * 2
	storedPlanes := s.Planes
	s.Coverage = "none"
	if s.Masking == 1 {
		storedPlanes++
		s.Coverage = "BMHD mask plane"
	}
	if s.Masking == 2 {
		s.Coverage = "BMHD transparent color"
	}
	s.PlanarBytes = s.RowBytes * s.Height * storedPlanes
	if s.PlanarBytes > maxRawBytes {
		return s, errors.New("screen planar data exceeds 32 MiB limit")
	}
	if s.Compression == 0 && len(body) != s.PlanarBytes {
		return s, errors.New("uncompressed BODY length mismatch")
	}
	// Only a single scan line needs temporary planar storage.
	rows := make([]byte, s.RowBytes*storedPlanes)
	s.Image = image.NewNRGBA(image.Rect(0, 0, s.Width, s.Height))
	pos := 0
	for y := 0; y < s.Height; y++ {
		for plane := 0; plane < storedPlanes; plane++ {
			row := rows[plane*s.RowBytes : (plane+1)*s.RowBytes]
			if s.Compression == 0 {
				copy(row, body[pos:pos+s.RowBytes])
				pos += s.RowBytes
			} else {
				used, err := decodeByteRunRow(body[pos:], row)
				if err != nil {
					return screen{}, fmt.Errorf("BODY row %d plane %d: %w", y, plane, err)
				}
				pos += used
			}
		}
		for x := 0; x < s.Width; x++ {
			bit := byte(128 >> uint(x%8))
			off := x / 8
			index := 0
			for plane := 0; plane < s.Planes; plane++ {
				if rows[plane*s.RowBytes+off]&bit != 0 {
					index |= 1 << plane
				}
			}
			if index >= len(pal) {
				return screen{}, errors.New("screen pixel index outside CMAP")
			}
			pixel := pal[index]
			if s.Masking == 1 && rows[s.Planes*s.RowBytes+off]&bit == 0 || s.Masking == 2 && index == s.TransparentColor {
				pixel = color.NRGBA{}
			}
			s.Image.SetNRGBA(x, y, pixel)
		}
	}
	if pos != len(body) {
		return screen{}, errors.New("trailing bytes after ILBM BODY rows")
	}
	return s, nil
}

// ByteRun1 packets must fit one word-padded row; runs never cross planes or
// scan lines. Every token consumes input, including the signed -128 no-op.
func decodeByteRunRow(input, row []byte) (int, error) {
	pos, out := 0, 0
	for out < len(row) {
		if pos >= len(input) {
			return 0, errors.New("truncated ByteRun1 control")
		}
		control := int(int8(input[pos]))
		pos++
		if control == -128 {
			continue
		}
		count := control + 1
		if control < 0 {
			count = 1 - control
		}
		if count > len(row)-out {
			return 0, errors.New("ByteRun1 packet crosses row boundary")
		}
		if control >= 0 {
			if count > len(input)-pos {
				return 0, errors.New("truncated ByteRun1 literal")
			}
			copy(row[out:out+count], input[pos:pos+count])
			pos += count
		} else {
			if pos >= len(input) {
				return 0, errors.New("truncated ByteRun1 repeat")
			}
			for i := 0; i < count; i++ {
				row[out+i] = input[pos]
			}
			pos++
		}
		out += count
	}
	return pos, nil
}

// safeScreenName confines the opt-in source to one data-folder filename and
// keeps file paths and the line-based source manifest free of control bytes.
func safeScreenName(name string) bool {
	if name == "." || name == ".." || name == "" {
		return false
	}
	for _, c := range name {
		allowed := c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-'
		if !allowed {
			return false
		}
	}
	return true
}

type screenArtifact struct {
	Source, Filename, Library, SourceHash string
	Screen                                screen
}

func exportScreens(output string, screens []screenArtifact) error {
	dir := filepath.Join(output, "screens")
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	number := strconv.Itoa
	rows := [][]string{{"source", "platform", "format", "width", "height", "row_bytes", "planes", "masking", "compression", "transparent_color", "coverage", "palette", "palette_colors", "bmhd_x_aspect", "bmhd_y_aspect", "bmhd_page_width", "bmhd_page_height", "camg", "body_bytes", "planar_bytes", "source_sha256", "png_sha256"}}
	for _, a := range screens {
		s := a.Screen
		encoded, err := encodePNG(s.Image)
		if err != nil {
			return err
		}
		if err := writeNew(filepath.Join(dir, a.Filename), encoded); err != nil {
			return err
		}
		rows = append(rows, []string{a.Source, "amiga-native", "FORM/ILBM", number(s.Width), number(s.Height), number(s.RowBytes), number(s.Planes), number(s.Masking), number(s.Compression), number(s.TransparentColor), s.Coverage, "CMAP", number(s.PaletteColors), number(s.AspectX), number(s.AspectY), number(s.PageWidth), number(s.PageHeight), fmt.Sprintf("0x%08x", s.CAMG), number(s.BodyBytes), number(s.PlanarBytes), a.SourceHash, fmt.Sprintf("%x", sha256.Sum256(encoded))})
	}
	var b bytes.Buffer
	if err := csv.NewWriter(&b).WriteAll(rows); err != nil {
		return err
	}
	return writeNew(filepath.Join(output, "screen-metadata.csv"), b.Bytes())
}
