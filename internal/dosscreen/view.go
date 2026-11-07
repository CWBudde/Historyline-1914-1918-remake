package dosscreen

import (
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// View is one panel's camera. The panel shows Cols columns of Rows cells, its
// first column being the even lines of hexgrid column Pair and its first row
// hexgrid lines 2·Row and 2·Row+1; the origin moves in whole column pairs and
// rows. The block of cells is centred in the panel.
//
// ShiftX and ShiftY move the origin by world pixels: the frame shifts a view at
// the map's ends so that its opening ends on the map (Screen.SlimFit).
type View struct {
	Panel          Rect
	Scale          float64 // device pixels per world pixel
	Cols, Rows     int
	Pair, Row      int
	ShiftX, ShiftY float64
}

// NewView fits as many whole cells as panel holds at scale, with the first column
// pair and row at pair, row (any values, on the map or not).
func NewView(panel Rect, scale float64, pair, row int) View {
	v := View{Panel: panel, Scale: scale, Pair: pair, Row: row}
	v.Cols, v.Rows = Fit(panel, scale)
	return v
}

// Origin is the world point at the top-left corner of the block of cells: the
// left vertex column of the first column and the top of its first cell, moved
// by the shift.
func (v View) Origin() (wx, wy float64) {
	return float64(v.Pair*hexgrid.ColPitch) + v.ShiftX, float64(v.Row*hexgrid.HexHeight) + v.ShiftY
}

// Offset is where the top-left corner of the block of cells lies on the screen.
func (v View) Offset() (sx, sy int) {
	bw, bh := BlockSize(v.Cols, v.Rows)
	return v.Panel.X + int(math.Round((float64(v.Panel.W)-bw*v.Scale)/2)),
		v.Panel.Y + int(math.Round((float64(v.Panel.H)-bh*v.Scale)/2))
}

// WorldToScreen maps a world point to device pixels. Draw the map with this
// transform (translate by -Origin, scale by Scale, translate by Offset).
func (v View) WorldToScreen(wx, wy float64) (sx, sy float64) {
	ox, oy := v.Origin()
	bx, by := v.Offset()
	return float64(bx) + (wx-ox)*v.Scale, float64(by) + (wy-oy)*v.Scale
}

// ScreenToWorld is the inverse of WorldToScreen.
func (v View) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	ox, oy := v.Origin()
	bx, by := v.Offset()
	return (sx-float64(bx))/v.Scale + ox, (sy-float64(by))/v.Scale + oy
}

// Cell returns the hex in column j (from the left) and cell i (from the top).
func (v View) Cell(j, i int) hexgrid.Coord {
	line := 2*(v.Row+i) + j&1
	return hexgrid.Coord{Line: line, Col: v.Pair + j>>1}
}

// Cells returns every cell of the view, column by column, on the map or not.
func (v View) Cells() []hexgrid.Coord {
	out := make([]hexgrid.Coord, 0, v.Cols*v.Rows)
	for j := range v.Cols {
		for i := range v.Rows {
			out = append(out, v.Cell(j, i))
		}
	}
	return out
}

// Point is a position in device pixels.
type Point struct{ X, Y float64 }
