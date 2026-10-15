package tpwm

import (
	"encoding/binary"
	"testing"
)

func header(order binary.ByteOrder, size uint32) []byte {
	h := append([]byte("TPWM"), 0, 0, 0, 0)
	order.PutUint32(h[4:], size)
	return h
}

func TestUnpack(t *testing.T) {
	le, be := binary.ByteOrder(binary.LittleEndian), binary.ByteOrder(binary.BigEndian)
	cases := []struct {
		name  string
		in    []byte
		order binary.ByteOrder
		want  string
	}{
		{"overlapping reference", append(header(le, 9), []byte("\x10ABC\x03\x03")...), le, "ABCABCABC"},
		{"reference clipped to size", append(header(le, 4), []byte("\x10ABC\x0f\x03")...), le, "ABCA"},
		{"big endian size", append(header(be, 9), []byte("\x10ABC\x03\x03")...), be, "ABCABCABC"},
		{"no magic", []byte("plain"), le, "plain"},
	}
	for _, c := range cases {
		got, err := Unpack(c.in, c.order)
		if err != nil || string(got) != c.want {
			t.Errorf("%s: got %q, %v; want %q", c.name, got, err, c.want)
		}
	}
}

func TestUnpackRejects(t *testing.T) {
	le := binary.ByteOrder(binary.LittleEndian)
	cases := map[string][]byte{
		"short header":        []byte("TPWM"),
		"zero size":           append([]byte("TPWM"), 0, 0, 0, 0),
		"truncated literal":   append(header(le, 2), 0x00, 'A'),
		"distance too far":    append(header(le, 3), 0x80, 0x00, 0x01),
		"truncated reference": append(header(le, 3), 0x80, 0x00),
		"trailing bytes":      append(header(le, 1), 0x00, 'A', 'x'),
		"wrong byte order":    append(header(binary.BigEndian, 3), 0x00, 'A', 'B', 'C'),
	}
	for name, in := range cases {
		if _, err := Unpack(in, le); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
