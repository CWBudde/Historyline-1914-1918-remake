// Package tpwm unpacks Turbo Packer ("TPWM") data. The DOS files store the
// unpacked size little-endian, the Amiga files big-endian; callers choose the
// variant explicitly because the stream header does not say.
package tpwm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// MaxUnpacked bounds the declared output size.
const MaxUnpacked = 32 * 1024 * 1024

var magic = []byte("TPWM")

// HasMagic reports whether b starts with a TPWM header.
func HasMagic(b []byte) bool { return bytes.HasPrefix(b, magic) }

// Unpack decodes a TPWM stream whose size field uses the given byte order.
// The stream must be consumed exactly: trailing bytes are an error. Data
// without the TPWM magic is returned unchanged.
func Unpack(b []byte, order binary.ByteOrder) ([]byte, error) {
	if !HasMagic(b) {
		return b, nil
	}
	if len(b) < 8 {
		return nil, errors.New("truncated TPWM header")
	}
	size := int(order.Uint32(b[4:8]))
	if size <= 0 || size > MaxUnpacked {
		return nil, fmt.Errorf("invalid TPWM size %d", order.Uint32(b[4:8]))
	}
	out := make([]byte, 0, size)
	pos := 8
	for len(out) < size {
		if pos >= len(b) {
			return nil, errors.New("truncated TPWM flag")
		}
		flags := b[pos]
		pos++
		for bit := 7; bit >= 0 && len(out) < size; bit-- {
			if flags&(1<<bit) == 0 {
				if pos >= len(b) {
					return nil, errors.New("truncated TPWM literal")
				}
				out = append(out, b[pos])
				pos++
				continue
			}
			if pos+2 > len(b) {
				return nil, errors.New("truncated TPWM reference")
			}
			hi, lo := b[pos], b[pos+1]
			pos += 2
			distance := int(hi&0xF0)<<4 | int(lo)
			if distance == 0 || distance > len(out) {
				return nil, errors.New("invalid TPWM distance")
			}
			for n := min(int(hi&15)+3, size-len(out)); n > 0; n-- {
				out = append(out, out[len(out)-distance])
			}
		}
	}
	if pos != len(b) {
		return nil, errors.New("trailing bytes after TPWM stream")
	}
	return out, nil
}
