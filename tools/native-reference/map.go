package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"strconv"
)

// This is a private ground-reference reconstruction, not the game compositor.
// Placement and FIN ordering are traced in docs/art/original-placement.md.
type groundMap struct {
	Image *image.NRGBA
	Cells [][]string
}

func reconstructGround(packed, index []byte, artifacts []artifact, library string) (*groundMap, error) {
	b, used, err := unpackBE(packed)
	if err != nil {
		return nil, err
	}
	if used != len(packed) || len(b) < 4 {
		return nil, errors.New("truncated FIN or trailing packed bytes")
	}
	w, h := int(binary.BigEndian.Uint16(b)), int(binary.BigEndian.Uint16(b[2:]))
	width, height := 16*(w-1)+24, 24*h
	if w > 1 {
		height += 12
	}
	if w < 1 || h < 1 || w > 255 || h > 255 || int64(width)*int64(height) > maxPixels || int64(len(b)) != 4+int64(w)*int64(h)*2 {
		return nil, errors.New("invalid FIN dimensions, size or pixel limit")
	}
	names, used, err := unpackBE(index)
	if err != nil {
		return nil, err
	}
	if used != len(index) || len(names) != 175*8 {
		return nil, errors.New("terrain index must contain exactly 175 eight-byte names")
	}
	images := make(map[string]*image.NRGBA)
	for _, a := range artifacts {
		if a.Library == library {
			images[a.Name] = a.Graphic.Image
		}
	}
	tiles := make([]*image.NRGBA, 175)
	for i := range tiles {
		name := string(names[i*8 : (i+1)*8])
		tiles[i] = images[name]
		if tiles[i] == nil || tiles[i].Bounds() != image.Rect(0, 0, 24, 24) {
			return nil, fmt.Errorf("terrain index %d names missing/non-24x24 resource %q", i, name)
		}
	}
	// Validate all IDs before allocating and drawing. The second byte is retained,
	// never interpreted as a ground index or silently installed as unit state.
	for i := 4; i < len(b); i += 2 {
		if int(b[i]) >= len(tiles) {
			return nil, fmt.Errorf("terrain ID %d at FIN offset %d exceeds index", b[i], i)
		}
	}
	g := &groundMap{Image: image.NewNRGBA(image.Rect(0, 0, width, height)), Cells: [][]string{{"column", "row", "fin_offset", "terrain_index", "resource", "second_byte_raw", "pixel_x", "pixel_y"}}}
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			offset := 4 + 2*(row*w+col)
			id := int(b[offset])
			x, y := 16*col, 24*row+12*(col%2)
			draw.Draw(g.Image, image.Rect(x, y, x+24, y+24), tiles[id], image.Point{}, draw.Over)
			g.Cells = append(g.Cells, []string{strconv.Itoa(col), strconv.Itoa(row), strconv.Itoa(offset), strconv.Itoa(id), string(names[id*8 : (id+1)*8]), strconv.Itoa(int(b[offset+1])), strconv.Itoa(x), strconv.Itoa(y)})
		}
	}
	return g, nil
}

func exportGround(output, stem string, g *groundMap) error {
	if err := os.Mkdir(filepath.Join(output, "maps"), 0755); err != nil {
		return err
	}
	png, err := encodePNG(g.Image)
	if err != nil {
		return err
	}
	if err = writeNew(filepath.Join(output, "maps", stem+".png"), png); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = csv.NewWriter(&buf).WriteAll(g.Cells); err != nil {
		return err
	}
	return writeNew(filepath.Join(output, "maps", stem+"-cells.csv"), buf.Bytes())
}
