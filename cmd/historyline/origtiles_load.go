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

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// loadOrigTiles reads the original terrain tiles, summer (LIB/PARTS) and
// winter (LIB/PARTW), one per terrain record, from the local original/ folder
// above the working directory. Only builds with -tags origtiles have it.
func loadOrigTiles() (*origTiles, error) {
	dir, err := findDosFull(".")
	if err != nil {
		return nil, err
	}
	sheets, err := readOrigTiles(dir)
	if err != nil {
		return nil, err
	}
	t := &origTiles{}
	for season, sheet := range sheets {
		for _, im := range sheet {
			t.tiles[season] = append(t.tiles[season], ebiten.NewImageFromImage(im))
		}
	}
	return t, nil
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

// decodeOrigTiles cuts the tiles of a library out in the order of names
// (8-byte names, one per terrain record). The library ends in a directory of
// 8-byte names and 32-bit offsets; each entry holds its transparent colour,
// its size and one palette index per pixel, in four planes of every fourth
// column.
func decodeOrigTiles(names, lib, pal []byte) ([]*image.NRGBA, error) {
	if len(names) != 8*origview.Records || len(pal) < 768 || len(lib) < 4 {
		return nil, errors.New("unexpected file sizes")
	}
	at := map[string]int{}
	for i := int(binary.LittleEndian.Uint32(lib)); i+12 <= len(lib); i += 12 {
		at[string(lib[i:i+8])] = int(binary.LittleEndian.Uint32(lib[i+8:]))
	}
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
