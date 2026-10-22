// Package pixfont is a capitals pixel font in two sizes, defined in code:
// Small for the status bars, Large for the overlays. It stands in until the
// font art is bound. Text is drawn at an integer scale, device pixels per font
// pixel, from one cached glyph atlas per face.
package pixfont

import (
	"image"
	"image/color"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
)

// Face is one size of the font. Lower case draws as upper case; runes without
// a glyph draw as a box.
type Face struct{ f *font }

var (
	// Small has capitals 5 font pixels high, for the status bars.
	Small = Face{newFont(5, boxSmall, smallGlyphs)}
	// Large has capitals 7 font pixels high, for the overlays.
	Large = Face{newFont(7, boxLarge, largeGlyphs)}
)

// box is the glyph index of the box for unknown runes.
const box = 0

// plain maps letters a face may lack to the letters they draw as.
var plain = map[rune]rune{'Ä': 'A', 'Ö': 'O', 'Ü': 'U'}

// expand spells out runes that draw as several glyphs.
var expand = strings.NewReplacer("ß", "SS", "ẞ", "SS", "…", "...")

type glyph struct {
	rows []string
	w    int
	x    int // left edge in the atlas
}

type font struct {
	height int
	glyphs []glyph
	index  map[rune]int

	once  sync.Once
	atlas *ebiten.Image
	subs  []*ebiten.Image
}

func newFont(height int, boxRows []string, table map[rune][]string) *font {
	f := &font{height: height, index: make(map[rune]int, len(table))}
	runes := make([]rune, 0, len(table))
	for r := range table {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	add := func(rows []string) int {
		x := 0
		if n := len(f.glyphs); n > 0 {
			x = f.glyphs[n-1].x + f.glyphs[n-1].w + 1
		}
		f.glyphs = append(f.glyphs, glyph{rows: rows, w: len(rows[0]), x: x})
		return len(f.glyphs) - 1
	}
	add(boxRows)
	for _, r := range runes {
		f.index[r] = add(table[r])
	}
	return f
}

// lookup returns the glyph index of r.
func (f *font) lookup(r rune) int {
	r = unicode.ToUpper(r)
	if g, ok := f.index[r]; ok {
		return g
	}
	if g, ok := f.index[plain[r]]; ok {
		return g
	}
	return box
}

// rect is glyph g's place in the atlas.
func (f *font) rect(g int) image.Rectangle {
	gl := f.glyphs[g]
	return image.Rect(gl.x, 0, gl.x+gl.w, f.height)
}

// atlasImage paints every glyph white on transparent, side by side with one
// empty column between them.
func (f *font) atlasImage() *image.RGBA {
	last := f.glyphs[len(f.glyphs)-1]
	img := image.NewRGBA(image.Rect(0, 0, last.x+last.w, f.height))
	for _, gl := range f.glyphs {
		for y, row := range gl.rows {
			for x := range row {
				if row[x] == '#' {
					img.SetRGBA(gl.x+x, y, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
				}
			}
		}
	}
	return img
}

// images returns the atlas sub-image of every glyph, building the atlas once.
func (f *font) images() []*ebiten.Image {
	f.once.Do(func() {
		f.atlas = ebiten.NewImageFromImage(f.atlasImage())
		f.subs = make([]*ebiten.Image, len(f.glyphs))
		for g := range f.glyphs {
			f.subs[g] = f.atlas.SubImage(f.rect(g)).(*ebiten.Image)
		}
	})
	return f.subs
}

// layout calls fn with every glyph of s, its left edge in font pixels and its
// colour: numbers for digits, text otherwise. It returns the width of s in
// font pixels.
func (f *font) layout(s string, text, numbers color.Color, fn func(g, x int, c color.Color)) int {
	if strings.ContainsAny(s, "ßẞ…") {
		s = expand.Replace(s)
	}
	x := 0
	for _, r := range s {
		g := f.lookup(r)
		if fn != nil {
			c := text
			if r >= '0' && r <= '9' {
				c = numbers
			}
			fn(g, x, c)
		}
		x += f.glyphs[g].w + 1
	}
	return max(x-1, 0)
}

// Measure returns the size of s drawn at scale in device pixels: the width up
// to the last glyph's last column and the cap height. Scales below 1 count as 1.
func (f Face) Measure(s string, scale int) (w, h int) {
	scale = max(scale, 1)
	return f.f.layout(s, nil, nil, nil) * scale, f.f.height * scale
}

// Draw draws s with its top-left corner at (x, y).
func (f Face) Draw(dst *ebiten.Image, s string, x, y, scale int, c color.Color) {
	f.draw(dst, s, x, y, scale, c, c)
}

// DrawNumbers draws s like Draw, with its digits in the numbers colour.
func (f Face) DrawNumbers(dst *ebiten.Image, s string, x, y, scale int, text, numbers color.Color) {
	f.draw(dst, s, x, y, scale, text, numbers)
}

// DrawShadow draws s like Draw over a copy in the shadow colour one font pixel
// to the right and down, so the ink reaches one font pixel past Measure.
func (f Face) DrawShadow(dst *ebiten.Image, s string, x, y, scale int, c, shadow color.Color) {
	scale = max(scale, 1)
	f.draw(dst, s, x+scale, y+scale, scale, shadow, shadow)
	f.draw(dst, s, x, y, scale, c, c)
}

func (f Face) draw(dst *ebiten.Image, s string, x, y, scale int, text, numbers color.Color) {
	scale = max(scale, 1)
	subs := f.f.images()
	var op ebiten.DrawImageOptions
	f.f.layout(s, text, numbers, func(g, gx int, c color.Color) {
		op.GeoM.Reset()
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x+gx*scale), float64(y))
		op.ColorScale.Reset()
		op.ColorScale.ScaleWithColor(c)
		dst.DrawImage(subs[g], &op)
	})
}

// Fit returns s if it fits into maxW device pixels at scale, otherwise its
// longest start that fits followed by "...", with trailing spaces dropped
// before the dots. It returns "" when not even the dots fit.
func (f Face) Fit(s string, scale, maxW int) string {
	if w, _ := f.Measure(s, scale); w <= maxW {
		return s
	}
	runes := []rune(s)
	for n := len(runes) - 1; n >= 0; n-- {
		t := strings.TrimRight(string(runes[:n]), " ") + "..."
		if w, _ := f.Measure(t, scale); w <= maxW {
			return t
		}
	}
	return ""
}
