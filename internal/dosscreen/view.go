package dosscreen

import (
	"image"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// View is one panel's camera. The panel shows Cols columns of Rows cells, its
// first column being the even lines of hexgrid column Pair and its first row
// hexgrid lines 2·Row and 2·Row+1. Because the origin moves in whole column pairs
// and rows, the teeth of the frame keep their phase and always meet cell edges.
// The block of cells is centred in the panel; the margin is steel.
//
// ShiftX and ShiftY move the origin by world pixels. The plate frame keeps them
// 0, so its teeth meet cell edges; the slim frame shifts a view at the map's
// ends so that its opening ends on the map (Screen.SlimFit).
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

// shows reports whether c is one of the view's cells and on the map: an opening
// in the frame.
func (v View) shows(g hexgrid.Grid, c hexgrid.Coord) bool {
	j := 2*(c.Col-v.Pair) + c.Line&1
	i := c.Line>>1 - v.Row
	return j >= 0 && j < v.Cols && i >= 0 && i < v.Rows && g.Contains(c)
}

// Mask returns the openings of the panel frame: 255 where a pixel centre lies in
// one of the view's on-map cells (their hexgrid polygons), 0 on steel. Its bounds
// are the panel's rect in screen pixels. Cells off the map stay steel, so the
// frame closes in at the map edge.
func Mask(v View, g hexgrid.Grid) *image.Alpha {
	m := image.NewAlpha(v.Panel.image())
	hw, hh := float64(hexgrid.HexWidth)/2*v.Scale, float64(hexgrid.HexHeight)/2*v.Scale
	run := float64(hexgrid.HexWidth-hexgrid.HalfPitch) * v.Scale // the slanted edges' run
	for _, c := range v.Cells() {
		if !g.Contains(c) {
			continue
		}
		cx, cy := v.WorldToScreen(hexgrid.Center(c))
		y0 := max(int(math.Ceil(cy-hh-0.5)), m.Rect.Min.Y)
		y1 := min(int(math.Ceil(cy+hh-0.5)), m.Rect.Max.Y)
		for y := y0; y < y1; y++ {
			half := hw - run*math.Abs(float64(y)+0.5-cy)/hh
			x0 := max(int(math.Ceil(cx-half-0.5)), m.Rect.Min.X)
			x1 := min(int(math.Ceil(cx+half-0.5)), m.Rect.Max.X)
			for x := x0; x < x1; x++ {
				m.Pix[m.PixOffset(x, y)] = 255
			}
		}
	}
	return m
}

// Point is a position in device pixels.
type Point struct{ X, Y float64 }

// Segment is one edge of the frame's inner outline, from A to B in device pixels,
// between an opening and steel. N is its unit normal pointing into the opening.
type Segment struct {
	A, B, N Point
}

// edgeNeighbour[k] is the index in hexgrid.Around of the hex across the edge from
// corner k to corner k+1 of hexgrid.Corners.
var edgeNeighbour = [6]int{4, 5, 0, 1, 2, 3}

// Outline returns the inner outline of the panel frame: every cell edge between
// an opening and a steel-covered hex.
func Outline(v View, g hexgrid.Grid) []Segment {
	var out []Segment
	for _, c := range v.Cells() {
		if !g.Contains(c) {
			continue
		}
		cx, cy := v.WorldToScreen(hexgrid.Center(c))
		corners := hexgrid.Corners(c)
		around := hexgrid.Around(c)
		for k := range 6 {
			if v.shows(g, around[edgeNeighbour[k]]) {
				continue
			}
			ax, ay := v.WorldToScreen(corners[k][0], corners[k][1])
			bx, by := v.WorldToScreen(corners[(k+1)%6][0], corners[(k+1)%6][1])
			nx, ny := -(by - ay), bx-ax
			l := math.Hypot(nx, ny)
			nx, ny = nx/l, ny/l
			if (cx-ax)*nx+(cy-ay)*ny < 0 {
				nx, ny = -nx, -ny
			}
			out = append(out, Segment{Point{ax, ay}, Point{bx, by}, Point{nx, ny}})
		}
	}
	return out
}

// Rivets returns the rivet centres: one on the steel beside every tooth vertex of
// the outline, set back from the vertex against the openings by the rivet offset.
// Rivets near the panel edge may lie partly outside it.
func Rivets(v View, g hexgrid.Grid) []Point {
	type acc struct{ p, n Point } // a vertex and the sum of its segments' normals
	var order []int64
	verts := map[int64]*acc{}
	key := func(p Point) int64 {
		return int64(math.Round(p.X*16))<<32 ^ int64(math.Round(p.Y*16))&0xffffffff
	}
	add := func(p, n Point) {
		k := key(p)
		a := verts[k]
		if a == nil {
			a = &acc{p: p}
			verts[k] = a
			order = append(order, k)
		}
		a.n.X, a.n.Y = a.n.X+n.X, a.n.Y+n.Y
	}
	segs := Outline(v, g)
	for _, s := range segs {
		add(s.A, s.N)
		add(s.B, s.N)
	}
	off := metrics(v.Scale).rivetOff
	out := make([]Point, 0, len(order))
	for _, k := range order {
		a := verts[k]
		l := math.Hypot(a.n.X, a.n.Y)
		if l < 1e-9 {
			continue // two openings meet at a point: no steel to rivet
		}
		out = append(out, Point{a.p.X - a.n.X/l*off, a.p.Y - a.n.Y/l*off})
	}
	return out
}

// frameMetrics are the frame's feature sizes in device pixels at one cell scale.
type frameMetrics struct {
	lip      float64 // the dark lip along the outline
	bevel    float64 // the bevel inside the lip
	rivetR   float64 // rivet radius
	rivetOff float64 // rivet centre to its tooth vertex
}

// metrics sizes the frame for cell scale s from world-pixel sizes, with a floor so
// the features stay visible in small windows.
func metrics(s float64) frameMetrics {
	m := frameMetrics{
		lip:    math.Max(2*s, 1),
		bevel:  math.Max(6*s, 2),
		rivetR: math.Max(4.5*s, 1.5),
	}
	// The narrowest steel corner at a vertex is a hex angle of about 116°: set the
	// rivet back far enough that it clears the lip on both edges.
	m.rivetOff = (m.rivetR + m.lip + 1) / math.Sin(56*math.Pi/180)
	return m
}
