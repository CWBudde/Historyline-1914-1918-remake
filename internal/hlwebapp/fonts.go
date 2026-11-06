package hlwebapp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/andybalholm/brotli"
)

// woff2 writes a single SFNT with null transforms (WOFF2 section 4.1), keeping
// all outline/metric tables intact. This avoids external font conversion tools.
// https://www.w3.org/TR/WOFF2/#table_dir_format
func woff2(src []byte) ([]byte, error) {
	be := binary.BigEndian
	if len(src) < 12 || be.Uint32(src[:4]) != 0x00010000 {
		return nil, fmt.Errorf("expected TrueType SFNT")
	}
	n := int(be.Uint16(src[4:6]))
	if n == 0 || 12+16*n > len(src) {
		return nil, fmt.Errorf("invalid SFNT directory")
	}
	type record struct {
		tag  string
		data []byte
	}
	tables := make([]record, n)
	for i := range n {
		r := src[12+16*i : 12+16*(i+1)]
		off, length := uint64(be.Uint32(r[8:12])), uint64(be.Uint32(r[12:16]))
		if off < uint64(12+16*n) || off+length > uint64(len(src)) {
			return nil, fmt.Errorf("SFNT table outside file")
		}
		tables[i] = record{string(r[:4]), src[off : off+length]}
	}
	// A WOFF2 reconstruction changes offsets/checksums, invalidating signatures.
	// The format also requires head.flags bit 11 even with null table transforms.
	tables = slices.DeleteFunc(tables, func(r record) bool { return r.tag == "DSIG" })
	n = len(tables)
	head := false
	for i, r := range tables {
		if r.tag == "head" {
			if len(r.data) < 18 {
				return nil, fmt.Errorf("short SFNT head table")
			}
			data := bytes.Clone(r.data)
			be.PutUint16(data[16:18], be.Uint16(data[16:18])|1<<11)
			tables[i].data = data
			head = true
		}
	}
	if !head {
		return nil, fmt.Errorf("missing SFNT head table")
	}
	slices.SortFunc(tables, func(a, b record) int { return bytes.Compare([]byte(a.tag), []byte(b.tag)) })
	var dir, stream bytes.Buffer
	sfntSize := 12 + 16*n
	base128 := func(v uint32) {
		var buf [5]byte
		i := 4
		buf[i] = byte(v & 127)
		for v >>= 7; v > 0; v >>= 7 {
			i--
			buf[i] = byte(v&127) | 128
		}
		dir.Write(buf[i:])
	}
	for i, r := range tables {
		if i > 0 && tables[i-1].tag == r.tag {
			return nil, fmt.Errorf("duplicate SFNT table")
		}
		flag := byte(63)
		if r.tag == "glyf" || r.tag == "loca" {
			flag |= 3 << 6
		}
		dir.WriteByte(flag)
		dir.WriteString(r.tag)
		base128(uint32(len(r.data)))
		stream.Write(r.data)
		sfntSize += (len(r.data) + 3) &^ 3
	}
	var compressed bytes.Buffer
	w := brotli.NewWriterLevel(&compressed, 9)
	if _, err := w.Write(stream.Bytes()); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	header := make([]byte, 48)
	copy(header, "wOF2")
	copy(header[4:8], src[:4])
	length := (48 + dir.Len() + compressed.Len() + 3) &^ 3
	be.PutUint32(header[8:12], uint32(length))
	be.PutUint16(header[12:14], uint16(n))
	be.PutUint32(header[16:20], uint32(sfntSize))
	be.PutUint32(header[20:24], uint32(compressed.Len()))
	be.PutUint16(header[24:26], 1)
	out := make([]byte, 0, length)
	out = append(out, header...)
	out = append(out, dir.Bytes()...)
	out = append(out, compressed.Bytes()...)
	return append(out, make([]byte, length-len(out))...), nil
}
