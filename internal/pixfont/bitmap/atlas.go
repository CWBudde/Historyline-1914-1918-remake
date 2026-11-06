// Package bitmap holds the pixel font's glyphs without a graphics dependency.
package bitmap

import (
	"image"
	"image/color"
	"slices"
)

// Rows returns immutable glyph rows and the unknown-rune box for a face.
func Rows(height int) (map[rune][]string, []string) {
	if height == 5 {
		return smallGlyphs, boxSmall
	}
	if height == 7 {
		return largeGlyphs, boxLarge
	}
	panic("unsupported pixel font height")
}

type Glyph struct {
	X       int `json:"x"`
	Width   int `json:"width"`
	Height  int `json:"height"`
	Advance int `json:"advance"`
}
type Metadata struct {
	Width    int               `json:"width"`
	Height   int               `json:"height"`
	Glyphs   map[string]Glyph  `json:"glyphs"`
	Fallback Glyph             `json:"fallback"`
	Plain    map[string]string `json:"plain"`
	Expand   map[string]string `json:"expand"`
}

// Atlas is white ink on transparent, with one empty column between glyphs.
func Atlas(height int) (*image.RGBA, Metadata) {
	table, box := Rows(height)
	keys := make([]rune, 0, len(table))
	for r := range table {
		keys = append(keys, r)
	}
	slices.Sort(keys)
	m := Metadata{Height: height, Glyphs: map[string]Glyph{}, Plain: map[string]string{"Ä": "A", "Ö": "O", "Ü": "U"}, Expand: map[string]string{"ß": "SS", "ẞ": "SS", "…": "..."}}
	m.Fallback = Glyph{0, len(box[0]), height, len(box[0]) + 1}
	m.Width = m.Fallback.Advance
	for _, r := range keys {
		w := len(table[r][0])
		m.Glyphs[string(r)] = Glyph{m.Width, w, height, w + 1}
		m.Width += w + 1
	}
	m.Width--
	im := image.NewRGBA(image.Rect(0, 0, m.Width, height))
	paint := func(x int, rows []string) {
		for y, row := range rows {
			for px, ink := range row {
				if ink == '#' {
					im.SetRGBA(x+px, y, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
	paint(0, box)
	for _, r := range keys {
		paint(m.Glyphs[string(r)].X, table[r])
	}
	return im, m
}
