package dosexe

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func mzFixture() []byte {
	d := make([]byte, 80)
	for i, v := range []uint16{0x5A4D, 80, 1, 1, 4, 0, 0xFFFF, 0, 0x80, 0, 0, 0, 28, 0} {
		binary.LittleEndian.PutUint16(d[i*2:], v)
	}
	binary.LittleEndian.PutUint16(d[28:], 2)
	binary.LittleEndian.PutUint16(d[66:], 0x1234)
	return d
}

func TestParseMZ(t *testing.T) {
	h, image, relocs, err := ParseMZ(append(mzFixture(), "overlay"...))
	if err != nil {
		t.Fatal(err)
	}
	if h.HeaderBytes != 64 || h.TrailingFileBytes != 7 || len(image) != 16 {
		t.Fatalf("header %+v image %d", h, len(image))
	}
	if len(relocs) != 1 || relocs[0] != (Relocation{0, 2, 2, 0x1234}) {
		t.Fatalf("relocs %+v", relocs)
	}
}

func TestParseMZRejects(t *testing.T) {
	cases := map[string][]byte{
		"truncated":  mzFixture()[:79],
		"short":      []byte("MZ"),
		"no magic":   make([]byte, 80),
		"reloc away": func() []byte { d := mzFixture(); binary.LittleEndian.PutUint16(d[28:], 15); return d }(),
		"table end":  func() []byte { d := mzFixture(); binary.LittleEndian.PutUint16(d[24:], 63); return d }(),
	}
	for name, d := range cases {
		if _, _, _, err := ParseMZ(d); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestUnits(t *testing.T) {
	rec := make([]byte, UnitLen)
	copy(rec, []byte{12, 4, 20, 6})
	copy(rec[27:], "TYPE")
	copy(rec[44:], "NAME")
	copy(rec[61:], []byte{65, 66, 90, 15, 4})
	rows, err := Units(bytes.Repeat(rec, UnitCount))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != UnitCount || rows[50].Offset != 50*UnitLen || rows[0].Name != "NAME" ||
		rows[0].Category != "TYPE" {
		t.Fatalf("rows %+v", rows[0])
	}
	if rows[0].CostRaw != 90 || rows[0].CapacityRaw != 15 || rows[0].WeightRaw != 4 {
		t.Fatalf("tail fields %+v", rows[0])
	}
	if !bytes.Equal(rows[0].Raw, rec) {
		t.Fatal("raw bytes lost")
	}
	if _, err := Units(make([]byte, UnitLen)); err == nil {
		t.Fatal("want count error")
	}
}

func TestTerrainTable(t *testing.T) {
	rows, err := TerrainTable(bytes.Repeat([]byte{8, 128, 16, 6, 2, 8}, TerrainCount))
	if err != nil {
		t.Fatal(err)
	}
	r := rows[174]
	if r.Offset != 174*TerrainLen || r.FlagsRaw != 0x8008 || r.MovementCostRaw != 6 || r.CombatClass != 8 {
		t.Fatalf("row %+v", r)
	}
	if _, err := TerrainTable(make([]byte, TerrainLen)); err == nil {
		t.Fatal("want count error")
	}
}
