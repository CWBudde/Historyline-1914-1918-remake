// Package dosexe reads the parts of the DOS installation that the rule research
// needs: the MZ executable layout and the fixed-size unit and terrain records.
// Nothing here embeds original bytes; callers supply them from the local,
// git-ignored installation.
package dosexe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// MZ is the executable header plus the sizes derived while splitting the file.
// The JSON names match the manifest written by cmd/dosexport.
type MZ struct {
	Magic              uint16 `json:"magic"`
	LastPageBytes      uint16 `json:"last_page_bytes"`
	Pages              uint16 `json:"pages"`
	RelocationCount    uint16 `json:"relocation_count"`
	HeaderParagraphs   uint16 `json:"header_paragraphs"`
	MinExtraParagraphs uint16 `json:"min_extra_paragraphs"`
	MaxExtraParagraphs uint16 `json:"max_extra_paragraphs"`
	SS                 uint16 `json:"ss"`
	SP                 uint16 `json:"sp"`
	Checksum           uint16 `json:"checksum"`
	IP                 uint16 `json:"ip"`
	CS                 uint16 `json:"cs"`
	RelocationTableOff uint16 `json:"relocation_table_offset"`
	Overlay            uint16 `json:"overlay"`
	HeaderBytes        int    `json:"header_bytes"`
	ImageBytes         int    `json:"image_bytes"`
	TrailingFileBytes  int    `json:"trailing_file_bytes"`
}

// Relocation is one segment fix-up of the load image.
type Relocation struct {
	Segment      int
	Offset       int
	ImageOffset  int
	SegmentValue int
}

// ParseMZ splits an MZ executable into header, load image and relocations.
func ParseMZ(data []byte) (MZ, []byte, []Relocation, error) {
	var h MZ
	if len(data) < 28 || data[0] != 'M' || data[1] != 'Z' {
		return h, nil, nil, errors.New("not an MZ executable")
	}
	w := func(i int) uint16 { return binary.LittleEndian.Uint16(data[i*2:]) }
	h = MZ{w(0), w(1), w(2), w(3), w(4), w(5), w(6), w(7), w(8), w(9), w(10), w(11), w(12), w(13), 0, 0, 0}
	headerSize := int(h.HeaderParagraphs) * 16
	if h.Pages == 0 || h.LastPageBytes > 511 {
		return h, nil, nil, errors.New("invalid MZ page count")
	}
	last := int(h.LastPageBytes)
	if last == 0 {
		last = 512
	}
	declared := (int(h.Pages)-1)*512 + last
	if headerSize < 28 || headerSize > declared || declared > len(data) {
		return h, nil, nil, errors.New("invalid or truncated MZ image")
	}
	off, count := int(h.RelocationTableOff), int(h.RelocationCount)
	if off < 28 || off+count*4 > headerSize {
		return h, nil, nil, errors.New("invalid MZ relocation table")
	}
	image := data[headerSize:declared]
	relocs := make([]Relocation, 0, count)
	for i := 0; i < count; i++ {
		o := int(binary.LittleEndian.Uint16(data[off+i*4:]))
		seg := int(binary.LittleEndian.Uint16(data[off+i*4+2:]))
		linear := seg*16 + o
		if linear+2 > len(image) {
			return h, nil, nil, errors.New("relocation outside MZ image")
		}
		relocs = append(relocs, Relocation{seg, o, linear, int(binary.LittleEndian.Uint16(image[linear:]))})
	}
	h.HeaderBytes = headerSize
	h.ImageBytes = len(image)
	h.TrailingFileBytes = len(data) - declared
	return h, image, relocs, nil
}

const (
	// UnitCount and UnitLen describe UNIT.DAT.
	UnitCount = 51
	UnitLen   = 69
	// TerrainCount and TerrainLen describe WW1GNDS.DAT and WW1GNDW.DAT.
	TerrainCount = 175
	TerrainLen   = 6
)

// Unit is one UNIT.DAT record. Unknown fields stay raw and Raw keeps every byte.
type Unit struct {
	Index, Offset                                         int
	Category, Name                                        string
	MovementBudgetRaw, DisplayedSpeed, Armour, MaxGroup   int
	AlternateStrengthRaw, MovementMaskRaw, TargetFlagsRaw int
	AirRangeEncoded, SurfaceRangeEncoded                  int
	AirPower, GroundPower, WaterPower                     int
	Flags13Raw, Flags15Raw, Flags17Raw                    int
	CostRaw, CapacityRaw, WeightRaw                       int
	Raw                                                   []byte
}

func cstr(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return string(b)
}

// Units decodes UNIT.DAT.
func Units(data []byte) ([]Unit, error) {
	if len(data) != UnitCount*UnitLen {
		return nil, fmt.Errorf("expected %d DOS unit records of %d bytes", UnitCount, UnitLen)
	}
	le := binary.LittleEndian
	rows := make([]Unit, UnitCount)
	for i := range rows {
		r := data[i*UnitLen : (i+1)*UnitLen]
		rows[i] = Unit{
			Index: i, Offset: i * UnitLen,
			Category: cstr(r[27:44]),
			// The long label occupies bytes 44..60; 61..62 are runtime IDs.
			Name:              cstr(r[44:61]),
			MovementBudgetRaw: int(r[0]), DisplayedSpeed: int(r[1]),
			Armour: int(r[2]), MaxGroup: int(r[3]), AlternateStrengthRaw: int(r[4]),
			MovementMaskRaw: int(r[5]), TargetFlagsRaw: int(r[6]) | int(r[7])<<8,
			AirRangeEncoded: int(r[8]), SurfaceRangeEncoded: int(r[9]),
			AirPower: int(r[10]), GroundPower: int(r[11]), WaterPower: int(r[12]),
			Flags13Raw: int(le.Uint16(r[13:])), Flags15Raw: int(le.Uint16(r[15:])),
			Flags17Raw: int(le.Uint16(r[17:])),
			CostRaw:    int(r[63]), CapacityRaw: int(r[64]), WeightRaw: int(r[65]),
			Raw: append([]byte(nil), r...),
		}
	}
	return rows, nil
}

// Terrain is one six-byte terrain record.
type Terrain struct {
	Index, Offset, FlagsRaw, MovementMaskRaw, MovementCostRaw, Byte4Raw, CombatClass int
	Raw                                                                              []byte
}

// TerrainTable decodes WW1GNDS.DAT or WW1GNDW.DAT.
func TerrainTable(data []byte) ([]Terrain, error) {
	if len(data) != TerrainCount*TerrainLen {
		return nil, fmt.Errorf("expected %d DOS terrain records of %d bytes", TerrainCount, TerrainLen)
	}
	rows := make([]Terrain, TerrainCount)
	for i := range rows {
		r := data[i*TerrainLen : (i+1)*TerrainLen]
		rows[i] = Terrain{i, i * TerrainLen, int(binary.LittleEndian.Uint16(r)), int(r[2]), int(r[3]),
			int(r[4]), int(r[5]), append([]byte(nil), r...)}
	}
	return rows, nil
}
