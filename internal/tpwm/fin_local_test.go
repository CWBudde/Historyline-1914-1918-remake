package tpwm

import (
	"encoding/binary"
	"os"
	"testing"
)

// testdata/00.FIN is a local original level (never committed); skipped when absent.
func TestLocalFIN(t *testing.T) {
	b, err := os.ReadFile("testdata/00.FIN")
	if err != nil {
		t.Skip("no local 00.FIN")
	}
	out, err := Unpack(b, binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if want := int(binary.LittleEndian.Uint32(b[4:8])); len(out) != want {
		t.Fatalf("unpacked %d bytes, header says %d", len(out), want)
	}
}
