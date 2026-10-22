// Package finmap decodes the original DOS level files and assembles them
// into the versioned level JSON (Level).
//
// A level is up to four files: MAP/NN.FIN (the map), MAP/NN.SHP (building
// and transporter inventories, energy and the types factories may produce),
// MAP/NN.COM (per-type parameters of the computer player, single-player maps
// only) and the map's 10-byte record in CODES.DAT (password, players, season,
// successor). All of them are TPWM-packed on disk; the decoders here take the
// unpacked bytes.
//
// Sides are 0 (Central Powers, player A), 1 (Allies, player B) and 2
// (neutral). Cells are [row, col] in rows of Width cells; odd columns sit half
// a cell lower (Coord converts to the hexgrid lines). Unit types index the 51
// UNIT.DAT records, terrain values the WW1GNDS.DAT records.
package finmap

import (
	"errors"
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// TypeCount is the number of unit types (UNIT.DAT records).
const TypeCount = 51

// Sides.
const (
	Central = 0
	Allied  = 1
	Neutral = 2
)

// Map is a decoded FIN file: a 4-byte header of two big-endian words (width,
// height; the game reads only their low bytes, the high bytes are 0), then two
// bytes per cell in rows of Width cells: the terrain record and the occupant.
// An occupant below 0xf4 is a unit, 2*type+side; 0xf4 and above is empty
// (the files use 0xff).
type Map struct {
	Width, Height int
	Terrain       []byte // row-major
	Occupant      []byte // row-major
}

// EmptyFrom is the lowest occupant value that holds no unit.
const EmptyFrom = 0xf4

// DecodeMap decodes an unpacked FIN file.
func DecodeMap(b []byte) (*Map, error) {
	if len(b) < 4 {
		return nil, errors.New("finmap: FIN header truncated")
	}
	if b[0] != 0 || b[2] != 0 {
		return nil, fmt.Errorf("finmap: FIN size words %#04x %#04x exceed a byte", int(b[0])<<8|int(b[1]), int(b[2])<<8|int(b[3]))
	}
	m := &Map{Width: int(b[1]), Height: int(b[3])}
	if m.Width == 0 || m.Height == 0 {
		return nil, fmt.Errorf("finmap: FIN size %dx%d", m.Width, m.Height)
	}
	n := m.Width * m.Height
	if len(b) != 4+2*n {
		return nil, fmt.Errorf("finmap: FIN has %d bytes, %dx%d needs %d", len(b), m.Width, m.Height, 4+2*n)
	}
	m.Terrain, m.Occupant = make([]byte, n), make([]byte, n)
	for i := range n {
		m.Terrain[i], m.Occupant[i] = b[4+2*i], b[5+2*i]
	}
	return m, nil
}

// TerrainAt is the terrain record of a cell.
func (m *Map) TerrainAt(row, col int) int { return int(m.Terrain[row*m.Width+col]) }

// OccupantAt is the occupant byte of a cell.
func (m *Map) OccupantAt(row, col int) int { return int(m.Occupant[row*m.Width+col]) }

// Coord is the hexgrid coordinate of cell [row, col]: the line is the half
// row (odd columns sit half a cell lower), Col the index within that line.
// The last column and the odd cells of the last row are the map edge
// (terrain record 174) and fall outside hexgrid.Grid{Width: Width/2,
// Lines: 2*Height-1}.
func Coord(row, col int) hexgrid.Coord { return hexgrid.Coord{Line: 2*row + col%2, Col: col / 2} }
