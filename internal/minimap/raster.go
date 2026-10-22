package minimap

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// World pixels per raster step: a block column per half column of hexes
// (HalfPitch), half a block per line (LinePitch). The raster leaves out the slanted
// tips of the outermost hexes, rasterInset world pixels on the left and right.
const (
	blockWorldW = hexgrid.HalfPitch
	blockWorldH = 2 * hexgrid.LinePitch
	rasterInset = (hexgrid.HexWidth - hexgrid.HalfPitch) / 2.0
)

// Raster is the cell-exact minimap of Grid, like the original's strategic map: hex
// (l, c) covers the CellW×CellH block at x = (2c + l&1)·CellW, y = l·CellH/2. A
// block column holds the hexes of one line parity stacked without a gap, and odd
// columns sit half a block lower, so every pixel inside the map belongs to exactly
// one hex. Only the half blocks along the jagged top and bottom edge stay empty.
type Raster struct {
	Grid         hexgrid.Grid
	CellW, CellH int // CellH is even
}

// Cells returns the raster of g with cellW×cellH blocks. cellW is at least 1;
// cellH is at least 2 and rounded up to an even number.
func Cells(g hexgrid.Grid, cellW, cellH int) Raster {
	return Raster{Grid: g, CellW: max(cellW, 1), CellH: max(cellH+cellH&1, 2)}
}

// aspectCellW is the block width that keeps a cellH-high block in the world's
// proportions (71 wide per 80 high).
func aspectCellW(cellH int) int {
	return max(1, int(math.Round(float64(cellH)*blockWorldW/blockWorldH)))
}

// FitCells returns the raster of g with the largest blocks in the world's
// proportions whose whole raster fits in fit. The smallest blocks are 2×2, even if
// they overflow.
func FitCells(g hexgrid.Grid, fit image.Point) Raster {
	for ch := max(fit.Y, 2) &^ 1; ch > 2; ch -= 2 {
		r := Cells(g, aspectCellW(ch), ch)
		if s := r.Size(); s.X <= fit.X && s.Y <= fit.Y {
			return r
		}
	}
	return Cells(g, 2, 2)
}

// FitGrid returns the cell-exact minimap of g that fits in fit, anchored to the
// bottom-right corner of a view of the given size with margin, and its raster.
// Draw the raster's image 1:1 at the minimap's Pos.
func FitGrid(g hexgrid.Grid, fit, view camera.Size, margin float64) (Minimap, Raster) {
	r := FitCells(g, image.Pt(int(fit.W), int(fit.H)))
	s := r.Size()
	return r.Minimap(camera.Point{X: view.W - margin - float64(s.X), Y: view.H - margin - float64(s.Y)}), r
}

// Size returns the raster size in pixels.
func (r Raster) Size() image.Point {
	if r.Grid.Width <= 0 || r.Grid.Lines <= 0 {
		return image.Point{}
	}
	return image.Pt((2*r.Grid.Width-1)*r.CellW, (r.Grid.Lines+1)*r.CellH/2)
}

// Block returns the pixels of hex c.
func (r Raster) Block(c hexgrid.Coord) image.Rectangle {
	x, y := (2*c.Col+c.Line&1)*r.CellW, c.Line*r.CellH/2
	return image.Rect(x, y, x+r.CellW, y+r.CellH)
}

// At returns the hex whose block holds pixel (x, y), and false for a pixel off
// the map.
func (r Raster) At(x, y int) (hexgrid.Coord, bool) {
	if x < 0 || y < 0 {
		return hexgrid.Coord{}, false
	}
	q, half := x/r.CellW, y/(r.CellH/2)
	line := half
	if half&1 != q&1 {
		line--
	}
	c := hexgrid.Coord{Line: line, Col: (q - q&1) / 2}
	if !r.Grid.Contains(c) {
		return hexgrid.Coord{}, false
	}
	return c, true
}

// ToWorld converts a raster position to a world position; FromWorld is its
// inverse. A block's centre is its hex's centre.
func (r Raster) ToWorld(px, py float64) (wx, wy float64) {
	return px*blockWorldW/float64(r.CellW) + rasterInset, py * blockWorldH / float64(r.CellH)
}

// FromWorld converts a world position to a raster position.
func (r Raster) FromWorld(wx, wy float64) (px, py float64) {
	return (wx - rasterInset) * float64(r.CellW) / blockWorldW, wy * float64(r.CellH) / blockWorldH
}

// Minimap returns the screen mapping of the raster drawn 1:1 at pos, so clicks and
// the camera rectangle line up with the blocks.
func (r Raster) Minimap(pos camera.Point) Minimap {
	w, h := r.Grid.Bounds()
	return Minimap{
		World:  camera.Size{W: w, H: h},
		Pos:    pos,
		Scale:  float64(r.CellW) / blockWorldW,
		ScaleY: float64(r.CellH) / blockWorldH,
		Inset:  camera.Point{X: rasterInset},
	}
}

// Render returns the raster with each block filled by fill; pixels off the map stay
// transparent.
func (r Raster) Render(fill func(hexgrid.Coord) color.RGBA) *image.RGBA {
	s := r.Size()
	img := image.NewRGBA(image.Rect(0, 0, s.X, s.Y))
	r.DrawCells(img, image.Point{}, fill)
	return img
}

// DrawCells fills each block of the raster placed at origin in dst; a transparent
// fill colour leaves the block alone.
func (r Raster) DrawCells(dst *image.RGBA, origin image.Point, fill func(hexgrid.Coord) color.RGBA) {
	for l := 0; l < r.Grid.Lines; l++ {
		for c := 0; c < r.Grid.Width; c++ {
			at := hexgrid.Coord{Line: l, Col: c}
			if !r.Grid.Contains(at) {
				continue
			}
			if col := fill(at); col.A != 0 {
				fillRect(dst, r.Block(at).Add(origin), col)
			}
		}
	}
}

// Shade is a side's colour as lit, mid and dark tones (light from the top left).
type Shade struct{ Lit, Mid, Dark color.RGBA }

// Side colours: the German side (player A on the shipped levels) blue, the other
// side yellow, Neutral for unowned buildings and units of no side.
var (
	Blue    = Shade{color.RGBA{85, 195, 243, 255}, color.RGBA{56, 134, 174, 255}, color.RGBA{32, 77, 109, 255}}
	Yellow  = Shade{color.RGBA{231, 211, 101, 255}, color.RGBA{166, 138, 52, 255}, color.RGBA{101, 73, 16, 255}}
	Neutral = Shade{color.RGBA{196, 196, 184, 255}, color.RGBA{140, 140, 128, 255}, color.RGBA{72, 72, 62, 255}}
)

// Unit is a unit dot on the map: a hex and its side's shade.
type Unit struct {
	At    hexgrid.Coord
	Shade Shade
}

// DrawUnit draws u as a square dot inset a quarter block into its block: the top
// and left edge lit, the bottom and right edge dark, the rest mid.
func (r Raster) DrawUnit(dst *image.RGBA, origin image.Point, u Unit) {
	b := r.Block(u.At).Add(origin)
	in := min(r.CellW, r.CellH) / 4
	shadedRect(dst, image.Rect(b.Min.X+in, b.Min.Y+in, b.Max.X-in, b.Max.Y-in), u.Shade)
}

// shadedRect fills rc with s.Mid, its top and left edge s.Lit and its bottom and
// right edge s.Dark; the two corners where the edges meet stay mid.
func shadedRect(dst *image.RGBA, rc image.Rectangle, s Shade) {
	clip := rc.Intersect(dst.Rect)
	for y := clip.Min.Y; y < clip.Max.Y; y++ {
		for x := clip.Min.X; x < clip.Max.X; x++ {
			lit := x == rc.Min.X || y == rc.Min.Y
			dark := x == rc.Max.X-1 || y == rc.Max.Y-1
			switch {
			case lit && !dark:
				dst.SetRGBA(x, y, s.Lit)
			case dark && !lit:
				dst.SetRGBA(x, y, s.Dark)
			default:
				dst.SetRGBA(x, y, s.Mid)
			}
		}
	}
}

func fillRect(dst *image.RGBA, rc image.Rectangle, c color.RGBA) {
	rc = rc.Intersect(dst.Rect)
	for y := rc.Min.Y; y < rc.Max.Y; y++ {
		for x := rc.Min.X; x < rc.Max.X; x++ {
			dst.SetRGBA(x, y, c)
		}
	}
}
