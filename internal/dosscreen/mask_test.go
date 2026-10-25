package dosscreen

import (
	"fmt"
	"image"
	"math"
	"sort"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// zooms are cell scales from the smallest window to 4K and beyond.
var zooms = []float64{0.5, 0.6, 0.77, 1, 680.0 / 600, 1.5, 1.7333, 2.25, 3.4667}

// testView is a panel of about 5×4 cells at scale s, placed off the screen origin
// and a little wider and taller than the block so the margins are steel.
func testView(s float64, pair, row int) View {
	bw, bh := BlockSize(5, 4)
	return NewView(Rect{X: 37, Y: 11, W: int(bw*s) + 30, H: int(bh*s) + 17}, s, pair, row)
}

func TestViewCells(t *testing.T) {
	v := NewView(Rect{0, 0, 632, 680}, 680.0/600, 3, 2)
	if v.Cols != 7 || v.Rows != 7 {
		t.Fatalf("view %d×%d, want 7×7", v.Cols, v.Rows)
	}
	for _, tc := range []struct {
		j, i int
		want hexgrid.Coord
	}{
		{0, 0, hexgrid.Coord{Line: 4, Col: 3}},
		{1, 0, hexgrid.Coord{Line: 5, Col: 3}},
		{2, 0, hexgrid.Coord{Line: 4, Col: 4}},
		{0, 1, hexgrid.Coord{Line: 6, Col: 3}},
		{6, 6, hexgrid.Coord{Line: 16, Col: 6}},
		{5, 6, hexgrid.Coord{Line: 17, Col: 5}},
	} {
		if got := v.Cell(tc.j, tc.i); got != tc.want {
			t.Errorf("Cell(%d, %d) = %v, want %v", tc.j, tc.i, got, tc.want)
		}
	}
	if n := len(v.Cells()); n != 49 {
		t.Errorf("%d cells, want 49", n)
	}
	// The block's top-left corner is the left vertex column of the first cell and
	// the top of the first line, drawn at the block offset.
	ox, oy := v.Origin()
	if ox != 3*hexgrid.ColPitch || oy != 2*hexgrid.HexHeight {
		t.Errorf("origin %v,%v", ox, oy)
	}
	bx, by := v.Offset()
	if sx, sy := v.WorldToScreen(ox, oy); sx != float64(bx) || sy != float64(by) {
		t.Errorf("origin drawn at %v,%v, offset %d,%d", sx, sy, bx, by)
	}
	for _, p := range [][2]float64{{0, 0}, {431.5, 217.25}, {1000, 900}} {
		wx, wy := v.ScreenToWorld(v.WorldToScreen(p[0], p[1]))
		if math.Abs(wx-p[0]) > 1e-9 || math.Abs(wy-p[1]) > 1e-9 {
			t.Errorf("round trip of %v = %v,%v", p, wx, wy)
		}
	}
}

// The mask is the union of the visible on-map cells' hexgrid polygons: every pixel
// more than one device pixel away from a polygon edge is set exactly when its
// centre lies inside a visible cell, at every zoom and origin.
func TestMaskIsUnionOfCellPolygons(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	for _, s := range zooms {
		for _, o := range [][2]int{{0, 0}, {3, 5}, {-1, -1}, {18, 17}} {
			v := testView(s, o[0], o[1])
			m := Mask(v, g)
			if m.Rect != v.Panel.image() {
				t.Fatalf("mask bounds %v, panel %+v", m.Rect, v.Panel)
			}
			var polys [][6][2]float64
			for _, c := range v.Cells() {
				if g.Contains(c) {
					polys = append(polys, screenCorners(v, c))
				}
			}
			bad, checked := 0, 0
			for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
				for x := m.Rect.Min.X; x < m.Rect.Max.X; x++ {
					px, py := float64(x)+0.5, float64(y)+0.5
					inside, near := false, false
					for _, p := range polys {
						if px < p[0][0]-1 || px > p[3][0]+1 || py < p[1][1]-1 || py > p[4][1]+1 {
							continue
						}
						d := polyDist(p, px, py)
						inside = inside || d > 0
						near = near || math.Abs(d) <= 1
					}
					if near {
						continue
					}
					checked++
					if got := m.AlphaAt(x, y).A == 255; got != inside {
						bad++
					}
				}
			}
			if bad > 0 || checked < m.Rect.Dx()*m.Rect.Dy()/2 {
				t.Errorf("scale %.4f origin %v: %d of %d pixels disagree with the cell polygons", s, o, bad, checked)
			}
		}
	}
}

// The teeth meet the cell edges: along every pixel row and column of the panel,
// every switch between steel and opening lies within a pixel of where that line
// crosses the outline of the visible cells' polygons, and no crossing of an
// opening at least two pixels wide goes missing, at every zoom. Along the top row
// the openings are exactly the even columns' top edges.
func TestTeethMeetCellEdges(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	for _, s := range zooms {
		for _, o := range [][2]int{{2, 3}, {-1, 0}} {
			checkTeeth(t, fmt.Sprintf("scale %.4f origin %v", s, o), testView(s, o[0], o[1]), g)
		}
	}
}

// checkTeeth checks the mask of v against its cells' polygons along every pixel
// row and column, and the top row's openings against the even columns' top edges
// when the first row is on the map.
func checkTeeth(t *testing.T, name string, v View, g hexgrid.Grid) {
	t.Helper()
	m := Mask(v, g)
	var polys [][6][2]float64
	for _, c := range v.Cells() {
		if g.Contains(c) {
			polys = append(polys, screenCorners(v, c))
		}
	}
	for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
		checkCrossings(t, fmt.Sprintf("%s row %d", name, y), rowTransitions(m, y), crossings(polys, 1, float64(y)+0.5), m.Rect.Min.X, m.Rect.Max.X)
	}
	for x := m.Rect.Min.X; x < m.Rect.Max.X; x++ {
		checkCrossings(t, fmt.Sprintf("%s column %d", name, x), colTransitions(m, x), crossings(polys, 0, float64(x)+0.5), m.Rect.Min.Y, m.Rect.Max.Y)
	}
	var corners []float64
	for j := 0; j < v.Cols; j += 2 {
		c := v.Cell(j, 0)
		if !g.Contains(c) {
			return
		}
		p := screenCorners(v, c)
		corners = append(corners, p[1][0], p[2][0])
	}
	_, by := v.Offset()
	got := rowTransitions(m, by)
	if len(got) != len(corners) {
		t.Errorf("%s: top transitions %v, want the corners %v", name, got, corners)
		return
	}
	for k, x := range got {
		if math.Abs(float64(x)-corners[k]) > 1 {
			t.Errorf("%s: top transition %d at %d, corner at %.2f", name, k, x, corners[k])
		}
	}
}

// checkCrossings matches the mask transitions of a scan line against its
// openings, the [lo, hi] intervals where it crosses the visible cells. The line
// runs from pixel from to pixel to; an edge within a pixel of its ends has no
// pixel beyond it to switch.
func checkCrossings(t *testing.T, name string, trans []int, spans [][2]float64, from, to int) {
	t.Helper()
	for _, x := range trans {
		ok := false
		for _, sp := range spans {
			ok = ok || math.Abs(float64(x)-sp[0]) <= 1 || math.Abs(float64(x)-sp[1]) <= 1
		}
		if !ok {
			t.Errorf("%s: transition at %d crosses no cell edge %v", name, x, spans)
		}
	}
	for k, sp := range spans {
		if sp[1]-sp[0] < 2 {
			continue
		}
		for side, e := range sp {
			// A steel sliver under two pixels wide (a tooth tip) may fall between pixels.
			if side == 0 && k > 0 && e-spans[k-1][1] < 2 || side == 1 && k+1 < len(spans) && spans[k+1][0]-e < 2 {
				continue
			}
			if e <= float64(from)+1 || e >= float64(to)-1 {
				continue
			}
			if !nearAny(e, ints(trans), 1) {
				t.Errorf("%s: cell edge at %.2f has no transition %v", name, e, trans)
			}
		}
	}
}

// crossings intersects the scan line coord[axis] == v with the polygons and
// returns the merged opening intervals along the other axis.
func crossings(polys [][6][2]float64, axis int, v float64) [][2]float64 {
	var spans [][2]float64
	for _, p := range polys {
		lo, hi, ok := math.Inf(1), math.Inf(-1), false
		for k := range 6 {
			a, b := p[k], p[(k+1)%6]
			if (a[axis] <= v) == (b[axis] <= v) {
				continue
			}
			f := (v - a[axis]) / (b[axis] - a[axis])
			u := a[1-axis] + f*(b[1-axis]-a[1-axis])
			lo, hi, ok = math.Min(lo, u), math.Max(hi, u), true
		}
		if ok {
			spans = append(spans, [2]float64{lo, hi})
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i][0] < spans[j][0] })
	var out [][2]float64
	for _, sp := range spans {
		if n := len(out); n > 0 && sp[0] <= out[n-1][1]+1e-6 {
			out[n-1][1] = math.Max(out[n-1][1], sp[1])
			continue
		}
		out = append(out, sp)
	}
	return out
}

func screenCorners(v View, c hexgrid.Coord) [6][2]float64 {
	var p [6][2]float64
	for k, q := range hexgrid.Corners(c) {
		p[k][0], p[k][1] = v.WorldToScreen(q[0], q[1])
	}
	return p
}

func ints(v []int) []float64 {
	out := make([]float64, len(v))
	for i, x := range v {
		out[i] = float64(x)
	}
	return out
}

// Cells beyond the map edge stay steel, so the frame closes in at the edge.
func TestMaskOffMapIsSteel(t *testing.T) {
	g := hexgrid.Grid{Width: 4, Lines: 6}
	for _, s := range zooms {
		for _, o := range [][2]int{{-1, -1}, {2, 2}, {0, 0}} {
			v := testView(s, o[0], o[1])
			m := Mask(v, g)
			for _, c := range v.Cells() {
				cx, cy := v.WorldToScreen(hexgrid.Center(c))
				if a := m.AlphaAt(int(cx), int(cy)).A; (a == 255) != g.Contains(c) {
					t.Errorf("scale %.4f origin %v: cell %v (on map %v) has mask %d", s, o, c, g.Contains(c), a)
				}
			}
		}
		if m := Mask(testView(s, 10, 10), g); !allZero(m.Pix) {
			t.Errorf("scale %.4f: a view off the map shows openings", s)
		}
	}
}

// Each outline segment has an opening on its normal side and steel on the other.
func TestOutlineSeparatesOpeningFromSteel(t *testing.T) {
	for _, g := range []hexgrid.Grid{{Width: 40, Lines: 40}, {Width: 4, Lines: 6}} {
		for _, s := range zooms {
			v := testView(s, 1, 1)
			m := Mask(v, g)
			segs := Outline(v, g)
			if len(segs) == 0 {
				t.Fatalf("scale %.4f grid %v: no outline", s, g)
			}
			for _, sg := range segs {
				if math.Abs(math.Hypot(sg.N.X, sg.N.Y)-1) > 1e-9 {
					t.Fatalf("normal %v not unit", sg.N)
				}
				mx, my := (sg.A.X+sg.B.X)/2, (sg.A.Y+sg.B.Y)/2
				for _, side := range []struct {
					d    float64
					want uint8
				}{{2, 255}, {-2, 0}} {
					x, y := int(math.Floor(mx+side.d*sg.N.X)), int(math.Floor(my+side.d*sg.N.Y))
					if !inRect(m.Rect.Min.X, m.Rect.Min.Y, m.Rect.Max.X, m.Rect.Max.Y, x, y) {
						continue // the steel side of the block's top row lies above the panel
					}
					if a := m.AlphaAt(x, y).A; a != side.want {
						t.Errorf("scale %.4f grid %v: segment %+v side %v has mask %d", s, g, sg, side.d, a)
					}
				}
			}
		}
	}
}

// A rivet sits on the steel next to every tooth vertex of the outline.
func TestRivetsOnToothVertices(t *testing.T) {
	for _, g := range []hexgrid.Grid{{Width: 40, Lines: 40}, {Width: 4, Lines: 6}} {
		for _, s := range zooms {
			v := testView(s, 1, 1)
			m := Mask(v, g)
			fm := metrics(s)
			rivets := Rivets(v, g)
			var verts []Point
			for _, sg := range Outline(v, g) {
				verts = append(verts, sg.A, sg.B)
			}
			name := fmt.Sprintf("scale %.4f grid %v", s, g)
			for _, r := range rivets {
				if !nearPoint(r, verts, fm.rivetOff+1e-6) {
					t.Errorf("%s: rivet %v far from every tooth vertex", name, r)
				}
				for _, d := range [][2]float64{{0, 0}, {fm.rivetR, 0}, {-fm.rivetR, 0}, {0, fm.rivetR}, {0, -fm.rivetR}} {
					x, y := int(math.Floor(r.X+d[0])), int(math.Floor(r.Y+d[1]))
					if inRect(m.Rect.Min.X, m.Rect.Min.Y, m.Rect.Max.X, m.Rect.Max.Y, x, y) && m.AlphaAt(x, y).A != 0 {
						t.Errorf("%s: rivet %v reaches into an opening", name, r)
					}
				}
			}
			for _, p := range verts {
				in := p.X > float64(v.Panel.X)+fm.rivetOff+fm.rivetR && p.X < float64(v.Panel.Right())-fm.rivetOff-fm.rivetR &&
					p.Y > float64(v.Panel.Y)+fm.rivetOff+fm.rivetR && p.Y < float64(v.Panel.Bottom())-fm.rivetOff-fm.rivetR
				if in && !nearPoint(p, rivets, fm.rivetOff+1e-6) {
					t.Errorf("%s: tooth vertex %v has no rivet", name, p)
				}
			}
		}
	}
}

// polyDist is the signed distance from (x, y) to the convex polygon p: positive
// inside, negative outside (the distance to the nearest edge line either way).
func polyDist(p [6][2]float64, x, y float64) float64 {
	cx, cy := 0.0, 0.0
	for _, q := range p {
		cx, cy = cx+q[0]/6, cy+q[1]/6
	}
	d := math.Inf(1)
	for k := range 6 {
		a, b := p[k], p[(k+1)%6]
		ex, ey := b[0]-a[0], b[1]-a[1]
		nx, ny := -ey, ex // towards the centre for a clockwise polygon on a y-down screen
		l := math.Hypot(nx, ny)
		nx, ny = nx/l, ny/l
		if (cx-a[0])*nx+(cy-a[1])*ny < 0 {
			nx, ny = -nx, -ny
		}
		d = math.Min(d, (x-a[0])*nx+(y-a[1])*ny)
	}
	return d
}

// rowTransitions returns the x of every switch between steel and opening along
// pixel row y: the left edge of the first pixel of each run.
func rowTransitions(m *image.Alpha, y int) []int {
	var out []int
	for x := m.Rect.Min.X + 1; x < m.Rect.Max.X; x++ {
		if (m.AlphaAt(x, y).A == 0) != (m.AlphaAt(x-1, y).A == 0) {
			out = append(out, x)
		}
	}
	return out
}

// colTransitions is rowTransitions down pixel column x.
func colTransitions(m *image.Alpha, x int) []int {
	var out []int
	for y := m.Rect.Min.Y + 1; y < m.Rect.Max.Y; y++ {
		if (m.AlphaAt(x, y).A == 0) != (m.AlphaAt(x, y-1).A == 0) {
			out = append(out, y)
		}
	}
	return out
}

func nearAny(p float64, pos []float64, tol float64) bool {
	for _, q := range pos {
		if math.Abs(p-q) <= tol {
			return true
		}
	}
	return false
}

func nearPoint(p Point, pts []Point, tol float64) bool {
	for _, q := range pts {
		if math.Hypot(p.X-q.X, p.Y-q.Y) <= tol {
			return true
		}
	}
	return false
}

func inRect(x0, y0, x1, y1, x, y int) bool { return x >= x0 && x < x1 && y >= y0 && y < y1 }

func allZero(b []uint8) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
