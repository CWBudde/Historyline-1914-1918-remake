// Package dosscreen lays out and frames the original's split battle screen for
// `-rules dos`: two equal map panels side by side, a divider post between them
// and a status bar under each, inside a riveted steel frame whose inner edge
// follows the outline of the visible hex cells.
//
// Layout turns a window size and UI scale into a Screen (device-pixel rects and
// the cell scale). A View is one panel's camera: the panel, the scale and the
// first visible column pair and row, so the view always scrolls in whole cells
// and the frame's teeth meet the cell edges. Mask, Outline and Rivets describe
// the frame of a View as pure geometry; PaintPanel and PaintChrome paint it with
// the frame art (SetArt: painted hex plates and steel) or a placeholder steel
// texture; Frame caches the painted images as Ebitengine images and draws them
// over the maps.
//
// A "column" is a vertical stack of hexes at one x (the even lines of one hexgrid
// column, or the odd lines of one); every other column is shifted down by half a
// cell. A "row" is one cell height, two hexgrid lines.
package dosscreen

import (
	"image"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Cell counts and the smallest cell scale.
const (
	// MinCols is the fewest columns a panel shows down to the smallest window.
	MinCols = 7
	// TargetRows is the cells per column the panel height is fitted to; with the
	// half-cell shift of every other column that is 8½ rows, so a hex covers
	// about 7.5 % of a 16:9 screen's width, as in the original.
	TargetRows = 8
	// MinRows is the fewest cells per column a panel shows in the smallest window.
	MinRows = 6
	// MinScale is the smallest cell scale per step of UI scale (device pixels per
	// world pixel), so cells stay readable; the smallest window holds MinCols ×
	// MinRows cells at it.
	MinScale = 0.5
)

// The fixed pieces in reference pixels (the original's 320×200 screen). The remake
// draws them at its art scale of 2 logical pixels per reference pixel, times the
// UI scale.
const (
	artScale      = 2
	dividerNative = 8  // the divider post between the panels
	gapNative     = 4  // steel between the viewports and the status bars
	statusNative  = 14 // a status bar well
	footNative    = 2  // steel under the status bars
	boltNative    = 8  // the bolt at the foot of the divider
	bevelNative   = 1  // the status well's bevel
)

// Rect is a rectangle in device pixels.
type Rect struct{ X, Y, W, H int }

// Right and Bottom are the first column and row past r.
func (r Rect) Right() int  { return r.X + r.W }
func (r Rect) Bottom() int { return r.Y + r.H }

// Overlaps reports whether r and o share a pixel.
func (r Rect) Overlaps(o Rect) bool {
	return r.X < o.Right() && o.X < r.Right() && r.Y < o.Bottom() && o.Y < r.Bottom()
}

func (r Rect) image() image.Rectangle { return image.Rect(r.X, r.Y, r.Right(), r.Bottom()) }

// Screen is the split battle screen for one window size and UI scale, in device
// pixels. Panels[0] is player A's (left), Panels[1] player B's (right).
type Screen struct {
	W, H      int
	UI        float64
	Panels    [2]Rect
	Divider   Rect // the post between the panels, down to the status bars
	Bolt      Rect // the bolt between the status bars, under the divider
	Status    [2]Rect
	CellScale float64 // device pixels per world pixel
	Cols      int     // columns per panel
	Rows      int     // cells per column (the panel shows Rows+½ rows)
	Zoom      int     // the index in Zooms of CellScale
	FitScale  float64 // the cell scale Layout fitted, at DefaultZoom
}

// Layout splits a physW × physH window at UI scale ui (1 = 100 %; ≤ 0 counts as
// 1). The panels, the divider and the status band fill the window; the panel
// height sets the cell scale for 8½ rows, extra width shows more columns, and a
// window too narrow for MinCols columns at that scale shrinks the cells to keep
// them (showing more rows). The scale never drops below MinScale × ui. The
// screen starts at DefaultZoom; Zoomed picks another level.
func Layout(physW, physH int, ui float64) Screen {
	if ui <= 0 {
		ui = 1
	}
	physW, physH = max(physW, 1), max(physH, 1)
	s := Screen{W: physW, H: physH, UI: ui}
	gap, status, foot := scaled(gapNative, ui), scaled(statusNative, ui), scaled(footNative, ui)
	panelW := max((physW-scaled(dividerNative, ui))/2, 1)
	panelH := max(physH-gap-status-foot, 1)
	divX := panelW
	s.Panels = [2]Rect{{0, 0, panelW, panelH}, {physW - panelW, 0, panelW, panelH}}
	s.Divider = Rect{divX, 0, max(physW-2*panelW, 0), panelH + gap}
	statusY := panelH + gap
	s.Status = [2]Rect{{0, statusY, panelW, status}, {physW - panelW, statusY, panelW, status}}
	bolt := min(scaled(boltNative, ui), s.Divider.W, status)
	s.Bolt = Rect{divX + (s.Divider.W-bolt)/2, statusY + (status-bolt)/2, bolt, bolt}

	tw, th := BlockSize(MinCols, TargetRows)
	scale := min(float64(panelH)/th, float64(panelW)/tw)
	s.CellScale = max(scale, MinScale*ui)
	s.FitScale, s.Zoom = s.CellScale, DefaultZoom
	s.Cols, s.Rows = Fit(s.Panels[0], s.CellScale)
	return s
}

// View returns the camera of panel side (0 or 1) at the layout's cell scale with
// its first column pair and row at pair, row.
func (s Screen) View(side, pair, row int) View {
	return NewView(s.Panels[side], s.CellScale, pair, row)
}

// MinWindow is the smallest window at UI scale ui that still shows MinCols ×
// MinRows cells per panel. Callers use it as the window's size limit.
func MinWindow(ui float64) (w, h int) {
	if ui <= 0 {
		ui = 1
	}
	bw, bh := BlockSize(MinCols, MinRows)
	sc := MinScale * ui
	pw, ph := int(math.Ceil(bw*sc-eps)), int(math.Ceil(bh*sc-eps))
	w = 2*pw + scaled(dividerNative, ui)
	h = ph + scaled(gapNative, ui) + scaled(statusNative, ui) + scaled(footNative, ui)
	return w, h
}

// BlockSize is the bounding box in world pixels of cols columns of rows cells:
// the columns overlap by a slanted edge and every other one is shifted down by
// half a cell.
func BlockSize(cols, rows int) (w, h float64) {
	w = float64((cols-1)*hexgrid.HalfPitch + hexgrid.HexWidth)
	h = float64(rows*hexgrid.HexHeight + hexgrid.HexHeight/2)
	if cols <= 1 {
		h = float64(rows * hexgrid.HexHeight)
	}
	return w, h
}

// Fit returns how many columns and cells per column fit in panel at scale (at
// least one each).
func Fit(panel Rect, scale float64) (cols, rows int) {
	w, h := float64(panel.W)/scale, float64(panel.H)/scale
	cols = int(math.Floor((w-hexgrid.HexWidth)/hexgrid.HalfPitch+eps)) + 1
	rows = int(math.Floor((h-hexgrid.HexHeight/2)/hexgrid.HexHeight + eps))
	if cols <= 1 {
		rows = int(math.Floor(h/hexgrid.HexHeight + eps))
	}
	return max(cols, 1), max(rows, 1)
}

// eps absorbs rounding when a block fits a panel exactly.
const eps = 1e-6

// scaled is n reference pixels at the art scale and UI scale ui, in whole device
// pixels (at least one).
func scaled(n int, ui float64) int {
	return max(int(math.Round(float64(n*artScale)*ui)), 1)
}
