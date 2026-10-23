// Package dospanel draws the views that open inside one map panel of the
// original's battle screen (`-rules dos`): the unit info, the game info, a
// container's contents and the battle scene: attacker above defender, each
// with photo, flag, count and a hex holding a figure per man, the losses
// falling one by one; and the capture scene, a hand planting the capturer's
// flag on a model of the building. The plates are procedural steel and
// brass, the capture scene painted; only layout, proportions and information
// follow the original. The finale of a battle won by taking the HQ, drawn
// over the whole battle screen, lives here too.
//
// Every view is laid out inside a panel rectangle in device pixels at a
// scale in device pixels per reference pixel (the original's 320×200 screen).
// The layout keeps the reference proportions, centred in the panel, and
// shrinks below the scale when the panel is too small; the battle scene
// takes the largest scale that fits. Drawing never paints outside the panel.
//
// The views take plain data (UnitInfo, GameInfo, Contents, Battle) and draw their
// text through Text, so the caller binds the game session and the font.
package dospanel

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Size is a text size of the pixel font.
type Size int

// The two text sizes: Small for the overlays, Large for headings.
const (
	Small Size = iota
	Large
)

// Text measures and draws one line of text in device pixels; (x, y) is the
// top left of the line.
type Text interface {
	Measure(s string, size Size) (w, h int)
	Draw(dst *ebiten.Image, s string, x, y int, size Size, c color.Color)
}

// Dir is a cursor direction; it mirrors gesture.Dir.
type Dir int

// Cursor directions, in screen terms; Centre is no direction.
const (
	Centre Dir = iota
	Up
	Left
	Right
	Down
)

var dirNames = [...]string{"centre", "up", "left", "right", "down"}

func (d Dir) String() string {
	if d < 0 || int(d) >= len(dirNames) {
		return fmt.Sprintf("Dir(%d)", int(d))
	}
	return dirNames[d]
}

// grid places a block of w×h reference pixels in a panel: centred, at the
// scale k (device pixels per reference pixel).
type grid struct {
	ox, oy, k float64
}

// fit centres a w×h block in panel at scale, or at the largest scale that
// fits when the panel is too small.
func fit(panel image.Rectangle, scale, w, h float64) grid {
	k := math.Min(scale, math.Min(float64(panel.Dx())/w, float64(panel.Dy())/h))
	k = math.Max(k, 1e-3)
	return grid{
		ox: float64(panel.Min.X) + (float64(panel.Dx())-w*k)/2,
		oy: float64(panel.Min.Y) + (float64(panel.Dy())-h*k)/2,
		k:  k,
	}
}

// x and y are the device pixel of the reference coordinate.
func (g grid) x(v float64) int { return int(math.Round(g.ox + v*g.k)) }
func (g grid) y(v float64) int { return int(math.Round(g.oy + v*g.k)) }

// rect is the device rectangle of the reference rectangle (x0, y0)-(x1, y1).
func (g grid) rect(x0, y0, x1, y1 float64) image.Rectangle {
	return image.Rect(g.x(x0), g.y(y0), g.x(x1), g.y(y1))
}

// clip is dst cut to panel, so nothing drawn on it lands outside.
func clip(dst *ebiten.Image, panel image.Rectangle) *ebiten.Image {
	return dst.SubImage(panel.Intersect(dst.Bounds())).(*ebiten.Image)
}

// writer draws text on a clipped image; a nil Text draws nothing.
type writer struct {
	dst *ebiten.Image
	t   Text
}

// left draws s with its left edge at x, centred on the line at cy.
func (w writer) left(s string, x, cy int, c color.Color) {
	if w.t == nil || s == "" {
		return
	}
	_, h := w.t.Measure(s, Small)
	w.t.Draw(w.dst, s, x, cy-h/2, Small, c)
}

// right draws s with its right edge at x.
func (w writer) right(s string, x, cy int, c color.Color) {
	if w.t == nil || s == "" {
		return
	}
	tw, _ := w.t.Measure(s, Small)
	w.left(s, x-tw, cy, c)
}

// centre draws s centred on cx.
func (w writer) centre(s string, cx, cy int, c color.Color) {
	if w.t == nil || s == "" {
		return
	}
	tw, _ := w.t.Measure(s, Small)
	w.left(s, cx-tw/2, cy, c)
}

// Text colours: cream on the plates, amber numbers, pale labels on the
// darkened map.
var (
	cream = color.RGBA{0xFB, 0xE3, 0xC7, 0xFF}
	amber = color.RGBA{0xF0, 0xB8, 0x58, 0xFF}
	pale  = color.RGBA{0xEE, 0xEE, 0xE2, 0xFF}
)
