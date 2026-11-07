package dosscreen

import (
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
)

// maxCached is how many painted panel frames a Frame keeps: both panels in the
// map's interior plus a few edge patterns while they scroll.
const maxCached = 6

// Frame draws the steel frame over both panels' maps and caches what it painted:
// the chrome per window size, UI scale and frame art, a panel frame per panel
// rect, cell scale, pattern of on-map cells and frame art. Inside the map that
// pattern does not change with the origin, so scrolling repaints only when a view
// reaches the map edge, and zooming repaints the panel frames but not the chrome. The
// zero Frame is ready to use.
//
// With Slim set it draws the slim frame of a painted map instead (slim.go): no
// panel frames, only its chrome, cached per window size and UI scale.
type Frame struct {
	Slim bool

	screen  Screen // the chrome's key
	art     int    // the frame art's generation the chrome was painted with
	slim    bool   // the chrome is the slim frame's
	chrome  *ebiten.Image
	metal   *ebiten.Image // the slim frame's gunmetal tile (DrawBeyond)
	panels  []cachedPanel // least recently used first
	painted int           // images painted so far, for tests
}

type cachedPanel struct {
	key panelKey
	img *ebiten.Image
}

type panelKey struct {
	panel      Rect
	scale      float64
	cols, rows int
	open       string // which of the view's cells are on the map, column by column
	art        int    // the frame art's generation
}

// Draw draws the frame of s with the panels' views on dst, over the maps already
// drawn there through each view's WorldToScreen. The slim frame draws only its
// chrome: views and g do not shape it.
func (f *Frame) Draw(dst *ebiten.Image, s Screen, views [2]View, g hexgrid.Grid) {
	_, art := currentArt()
	if key := s.chromeKey(); f.chrome == nil || f.screen != key || f.art != art || f.slim != f.Slim {
		if f.chrome != nil {
			f.chrome.Deallocate()
		}
		if f.Slim {
			f.chrome = ebiten.NewImageFromImage(PaintSlimChrome(s))
		} else {
			f.chrome = ebiten.NewImageFromImage(PaintChrome(s))
		}
		f.screen, f.art, f.slim = key, art, f.Slim
		f.painted++
	}
	if f.Slim {
		dst.DrawImage(f.chrome, nil)
		return
	}
	for _, v := range views {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(v.Panel.X), float64(v.Panel.Y))
		dst.DrawImage(f.panel(v, g), op)
	}
	dst.DrawImage(f.chrome, nil)
}

// Release frees the cached images.
func (f *Frame) Release() {
	if f.chrome != nil {
		f.chrome.Deallocate()
		f.chrome = nil
	}
	if f.metal != nil {
		f.metal.Deallocate()
		f.metal = nil
	}
	for _, p := range f.panels {
		p.img.Deallocate()
	}
	f.panels = nil
}

// panel returns the painted frame of v, from the cache when it has one.
func (f *Frame) panel(v View, g hexgrid.Grid) *ebiten.Image {
	var open strings.Builder
	for _, c := range v.Cells() {
		if g.Contains(c) {
			open.WriteByte('1')
		} else {
			open.WriteByte('0')
		}
	}
	_, art := currentArt()
	key := panelKey{v.Panel, v.Scale, v.Cols, v.Rows, open.String(), art}
	for i, p := range f.panels {
		if p.key == key {
			f.panels = append(append(f.panels[:i:i], f.panels[i+1:]...), p)
			return p.img
		}
	}
	if len(f.panels) >= maxCached {
		f.panels[0].img.Deallocate()
		f.panels = f.panels[1:]
	}
	img := ebiten.NewImageFromImage(PaintPanel(v, g))
	f.painted++
	f.panels = append(f.panels, cachedPanel{key, img})
	return img
}

// chromeKey is s without the cell scale and zoom, which the chrome does not
// depend on.
func (s Screen) chromeKey() Screen {
	s.CellScale, s.FitScale, s.Cols, s.Rows, s.Zoom = 0, 0, 0, 0, 0
	return s
}
