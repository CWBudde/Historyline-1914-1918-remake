//go:build origtiles

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
	"github.com/hajimehoshi/ebiten/v2"
)

// loadOrigArt reads the original terrain tiles, summer (LIB/PARTS) and winter
// (LIB/PARTW), one per terrain record, and the unit sprites (LIB/UNIT) from
// the local original/ folder above the working directory. Only builds with
// -tags origtiles have it.
func loadOrigArt() (*origArt, error) {
	dir, err := findDosFull(".")
	if err != nil {
		return nil, err
	}
	sheets, err := readOrigTiles(dir)
	if err != nil {
		return nil, err
	}
	units, err := readOrigUnits(dir)
	if err != nil {
		return nil, err
	}
	a := &origArt{}
	for season, sheet := range sheets {
		for _, im := range sheet {
			a.tiles[season] = append(a.tiles[season], ebiten.NewImageFromImage(im))
		}
	}
	for side := range units {
		for moved := range units[side] {
			for _, im := range units[side][moved] {
				a.units[side][moved] = append(a.units[side][moved], ebiten.NewImageFromImage(im))
			}
		}
	}
	return a, nil
}

// readOrigTiles decodes the summer and winter tiles of the game in dir
// (original/extracted/dos-full), in terrain record order.
func readOrigTiles(dir string) ([2][]*image.NRGBA, error) {
	var out [2][]*image.NRGBA
	pal, err := unpackFile(filepath.Join(dir, "00.PAL"))
	if err != nil {
		return out, err
	}
	for season, lib := range [2]string{"PARTS", "PARTW"} {
		names, err := unpackFile(filepath.Join(dir, "LIB", lib+".DAT"))
		if err != nil {
			return out, err
		}
		data, err := unpackFile(filepath.Join(dir, "LIB", lib+".LIB"))
		if err != nil {
			return out, err
		}
		if out[season], err = decodeOrigTiles(names, data, pal); err != nil {
			return out, fmt.Errorf("%s: %w", lib, err)
		}
	}
	return out, nil
}

// origUnitTypes is the number of unit types, each with six frames.
const origUnitTypes = 51

// origSideColours are the first palette entries of each side's unit colours:
// side 0 (Central Powers) green, side 1 (Entente) tan. Neutral units keep
// their sprite's own.
var origSideColours = [2]int{0x20, 0x30}

// readOrigUnits decodes the unit frames of the game in dir by side (0, 1, 2
// neutral) and moved (0, 1), in unit type order: six frames per type,
// clockwise from north.
func readOrigUnits(dir string) ([3][2][]*image.NRGBA, error) {
	var out [3][2][]*image.NRGBA
	pal, err := unpackFile(filepath.Join(dir, "00.PAL"))
	if err != nil {
		return out, err
	}
	names, err := unpackFile(filepath.Join(dir, "LIB", "UNIT.DAT"))
	if err != nil {
		return out, err
	}
	data, err := unpackFile(filepath.Join(dir, "LIB", "UNIT.LIB"))
	if err != nil {
		return out, err
	}
	if len(names) != 8*6*origUnitTypes || len(pal) < 768 {
		return out, errors.New("UNIT: unexpected file sizes")
	}
	at := libDirectory(data)
	for k := range 6 * origUnitTypes {
		name := string(names[8*k : 8*k+8])
		off, ok := at[name]
		if !ok || off+18 > len(data) {
			return out, fmt.Errorf("UNIT frame %d: no sprite %q", k, name)
		}
		key, own := data[off+8], int(data[off+9])
		w, h := int(binary.LittleEndian.Uint16(data[off+14:])), int(binary.LittleEndian.Uint16(data[off+16:]))
		px := data[off+18:]
		if len(px) < w*h/2 {
			return out, fmt.Errorf("UNIT frame %d: sprite %q is cut short", k, name)
		}
		for side := range out {
			base := own
			if side < len(origSideColours) {
				base = origSideColours[side]
			}
			im, grey := image.NewNRGBA(image.Rect(0, 0, w, h)), image.NewNRGBA(image.Rect(0, 0, w, h))
			q := w / 4 * h // one plane: every fourth column
			for i := range w * h {
				v := px[i/2] >> 4 // two pixels a byte, the first in the high bits
				if i%2 == 1 {
					v = px[i/2] & 15
				}
				if v == key {
					continue
				}
				x, y := i/q+i%q%(w/4)*4, i%q/(w/4)
				c := 3 * (base + int(v))
				r, g, b := pal[c], pal[c+1], pal[c+2]
				im.SetNRGBA(x, y, color.NRGBA{r, g, b, 255})
				l := uint8((299*int(r) + 587*int(g) + 114*int(b)) * 3 / 4 / 1000)
				grey.SetNRGBA(x, y, color.NRGBA{l, l, l, 255})
			}
			out[side][0] = append(out[side][0], im)
			out[side][1] = append(out[side][1], grey)
		}
	}
	return out, nil
}

// libDirectory maps the entry names of a library to their offsets: the
// library ends in a directory of 8-byte names and 32-bit offsets.
func libDirectory(lib []byte) map[string]int {
	at := map[string]int{}
	if len(lib) < 4 {
		return at
	}
	for i := int(binary.LittleEndian.Uint32(lib)); i+12 <= len(lib); i += 12 {
		at[string(lib[i:i+8])] = int(binary.LittleEndian.Uint32(lib[i+8:]))
	}
	return at
}

// decodeOrigTiles cuts the tiles of a library out in the order of names
// (8-byte names, one per terrain record). Each entry holds its transparent
// colour, its size and one palette index per pixel, in four planes of every
// fourth column.
func decodeOrigTiles(names, lib, pal []byte) ([]*image.NRGBA, error) {
	if len(names) != 8*origview.Records || len(pal) < 768 || len(lib) < 4 {
		return nil, errors.New("unexpected file sizes")
	}
	at := libDirectory(lib)
	out := make([]*image.NRGBA, origview.Records)
	for r := range out {
		name := string(names[8*r : 8*r+8])
		off, ok := at[name]
		if !ok || off+18 > len(lib) {
			return nil, fmt.Errorf("record %d: no tile %q", r, name)
		}
		key := lib[off+8]
		w, h := int(binary.LittleEndian.Uint16(lib[off+14:])), int(binary.LittleEndian.Uint16(lib[off+16:]))
		px := lib[off+18:]
		if len(px) < w*h {
			return nil, fmt.Errorf("record %d: tile %q is cut short", r, name)
		}
		im := image.NewNRGBA(image.Rect(0, 0, w, h))
		q := w / 4 * h // one plane: every fourth column
		for i, p := range px[:w*h] {
			if p != key {
				x, y := i/q+i%q%(w/4)*4, i%q/(w/4)
				im.SetNRGBA(x, y, color.NRGBA{pal[3*int(p)], pal[3*int(p)+1], pal[3*int(p)+2], 255})
			}
		}
		out[r] = im
	}
	return out, nil
}

// unpackFile reads and unpacks a packed file of the game.
func unpackFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return tpwm.Unpack(b, binary.LittleEndian)
}

// findDosFull looks for original/extracted/dos-full from dir upwards.
func findDosFull(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		d := filepath.Join(dir, "original", "extracted", "dos-full")
		if st, err := os.Stat(filepath.Join(d, "LIB")); err == nil && st.IsDir() {
			return d, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no original/extracted/dos-full above the working directory")
		}
		dir = parent
	}
}
