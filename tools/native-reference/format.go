package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math/bits"
	"sort"
	"strings"
)

const maxRawBytes = 32 << 20
const maxPixels = 4 << 20

type graphic struct {
	Image                                                    *image.NRGBA
	Width, Height, RowBytes, PlaneMask, DataOffset, RawBytes int
}
type resource struct {
	Name   string
	Offset int
	Data   []byte
}

// unpackBE reads the Amiga TPWM variant. DOS stores the size little-endian
// and is deliberately unsupported. used excludes any archive alignment byte.
func unpackBE(b []byte) ([]byte, int, error) {
	if !bytes.HasPrefix(b, []byte("TPWM")) {
		return b, len(b), nil
	}
	if len(b) < 8 {
		return nil, 0, errors.New("truncated TPWM header")
	}
	size := uint64(binary.BigEndian.Uint32(b[4:8]))
	if size == 0 || size > maxRawBytes {
		return nil, 0, errors.New("invalid Amiga TPWM output size; DOS unsupported")
	}
	out := make([]byte, 0, int(size))
	pos := 8
	for len(out) < int(size) {
		if pos >= len(b) {
			return nil, 0, errors.New("truncated TPWM flag")
		}
		flag := b[pos]
		pos++
		for bit := byte(128); bit != 0 && len(out) < int(size); bit >>= 1 {
			if flag&bit == 0 {
				if pos >= len(b) {
					return nil, 0, errors.New("truncated TPWM literal")
				}
				out = append(out, b[pos])
				pos++
				continue
			}
			if len(b)-pos < 2 {
				return nil, 0, errors.New("truncated TPWM backreference")
			}
			a, c := b[pos], b[pos+1]
			pos += 2
			distance := (int(a&0xf0) << 4) | int(c)
			if distance == 0 || distance > len(out) {
				return nil, 0, errors.New("TPWM backreference outside produced data")
			}
			count := min(int(a&15)+3, int(size)-len(out))
			for i := 0; i < count; i++ {
				out = append(out, out[len(out)-distance])
			}
		}
	}
	return out, pos, nil
}

func parsePalette(b []byte) ([32]color.NRGBA, error) {
	var pal [32]color.NRGBA
	if len(b) != 64 {
		return pal, errors.New("amiga palette must contain 32 raw 0RGB words (64 bytes); DOS unsupported")
	}
	for i := range pal {
		v := binary.BigEndian.Uint16(b[2*i : 2*i+2])
		if v&0xf000 != 0 {
			return pal, errors.New("palette word exceeds native 12-bit RGB")
		}
		pal[i] = color.NRGBA{R: byte(v>>8&15) * 17, G: byte(v>>4&15) * 17, B: byte(v&15) * 17, A: 255}
	}
	return pal, nil
}

func parseLibrary(input []byte) ([]resource, error) {
	b, used, err := unpackBE(input)
	if err != nil {
		return nil, err
	}
	if used != len(input) {
		return nil, errors.New("trailing bytes after whole-library TPWM stream")
	}
	return parseLibraryData(b, binary.BigEndian)
}

func parseLibraryData(b []byte, order binary.ByteOrder) ([]resource, error) {
	if len(b) < 4 || len(b) > maxRawBytes {
		return nil, errors.New("invalid native library size")
	}
	fat := uint64(order.Uint32(b[:4]))
	if fat < 4 || fat >= uint64(len(b)) || (uint64(len(b))-fat)%12 != 0 {
		return nil, errors.New("invalid library FAT")
	}
	count := (len(b) - int(fat)) / 12
	if count > 10000 {
		return nil, errors.New("native library exceeds 10000 resource limit")
	}
	resources := make([]resource, 0, count)
	names := make(map[string]bool)
	for pos := int(fat); pos < len(b); pos += 12 {
		name := strings.TrimRight(string(b[pos:pos+8]), "\x00 ")
		if name == "" || names[name] {
			return nil, errors.New("empty or duplicate resource name")
		}
		for _, c := range name {
			validName := c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-'
			if !validName {
				return nil, errors.New("unsafe resource name")
			}
		}
		names[name] = true
		off := uint64(order.Uint32(b[pos+8 : pos+12]))
		if off < 4 || off >= fat {
			return nil, fmt.Errorf("resource %s offset outside data section", name)
		}
		resources = append(resources, resource{Name: name, Offset: int(off)})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].Offset < resources[j].Offset })
	if resources[0].Offset != 4 {
		return nil, errors.New("native resource section must start at byte 4")
	}
	for i := range resources {
		end := int(fat)
		if i+1 < len(resources) {
			end = resources[i+1].Offset
		}
		if end <= resources[i].Offset {
			return nil, errors.New("duplicate or overlapping resource offsets")
		}
		resources[i].Data = b[resources[i].Offset:end]
	}
	return resources, nil
}

func decodeGraphic(input []byte, pal [32]color.NRGBA, externalMask []byte) (graphic, error) {
	var g graphic
	b, used, err := unpackBE(input)
	if err != nil {
		return g, err
	}
	// Individually packed Amiga resources are word-aligned. The single
	// alignment byte can contain arbitrary data; it is not another TPWM token.
	alignedPadding := bytes.HasPrefix(input, []byte("TPWM")) && used%2 == 1 && len(input) == used+1
	if used != len(input) && !alignedPadding {
		return g, errors.New("unexpected bytes after packed graphic")
	}
	if len(b) < 24 || string(b[:8]) != "ILBM    " {
		return g, errors.New("expected native 24-byte ILBM record header")
	}
	word := func(off int) int { return int(binary.BigEndian.Uint16(b[off : off+2])) }
	g.PlaneMask, g.DataOffset, g.RowBytes = word(8), word(10), word(18)*2
	g.Width, g.Height, g.RawBytes = word(20), word(22), len(b)
	if g.Width < 1 || g.Height < 1 || int64(g.Width)*int64(g.Height) > maxPixels {
		return g, errors.New("invalid graphic dimensions")
	}
	if g.PlaneMask&^31 != 0 || g.RowBytes != ((g.Width+15)/16)*2 {
		return g, errors.New("invalid plane mask or word-padded stride")
	}
	planeBytes := g.RowBytes * g.Height
	if g.DataOffset != 24 && g.DataOffset != 24+planeBytes {
		return g, errors.New("invalid pixel-data offset or embedded coverage plane")
	}
	if g.DataOffset+planeBytes*bits.OnesCount(uint(g.PlaneMask)) != len(b) {
		return g, errors.New("graphic plane data does not match resource length")
	}
	mask := externalMask
	if mask != nil && (g.Width != 24 || g.Height != 24 || g.RowBytes != 4 || len(mask) != 96) {
		return g, errors.New("external native hex mask requires a 24x24 graphic and 96 bytes")
	}
	if g.DataOffset > 24 {
		if mask != nil {
			return g, errors.New("cannot combine embedded and external masks")
		}
		mask = b[24:g.DataOffset]
	}
	g.Image = image.NewNRGBA(image.Rect(0, 0, g.Width, g.Height))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			bit := byte(128 >> uint(x%8))
			offset := y*g.RowBytes + x/8
			value, storedPlane := 0, 0
			for plane := 0; plane < 5; plane++ {
				if g.PlaneMask&(1<<plane) == 0 {
					continue
				}
				if b[g.DataOffset+storedPlane*planeBytes+offset]&bit != 0 {
					value |= 1 << plane
				}
				storedPlane++
			}
			pixel := pal[value]
			if mask != nil && mask[offset]&bit == 0 {
				pixel = color.NRGBA{}
			}
			g.Image.SetNRGBA(x, y, pixel)
		}
	}
	return g, nil
}
