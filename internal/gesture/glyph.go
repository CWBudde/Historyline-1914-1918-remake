package gesture

import "math"

// polygon is one part of a glyph in its unit box: x and y in [-0.5, 0.5], y
// down. Solid parts wind with positive shoelace area, holes negative, so a
// non-zero fill of all parts cuts the holes; a hole must lie inside a part
// covered once.
type polygon struct {
	pts  [][2]float64
	hole bool
}

// glyph returns the embossed symbol of a. Every symbol is newly drawn from
// bars, triangles and discs.
func glyph(a Action) []polygon {
	switch a {
	case Cancel: // X
		return solid(bar(-0.36, -0.36, 0.36, 0.36, 0.16), bar(-0.36, 0.36, 0.36, -0.36, 0.16))
	case Move: // four arrows
		return solid(
			bar(-0.3, 0, 0.3, 0, 0.12), bar(0, -0.3, 0, 0.3, 0.12),
			tri(0, -0.5, 0.16, -0.28, -0.16, -0.28), tri(0, 0.5, -0.16, 0.28, 0.16, 0.28),
			tri(-0.5, 0, -0.28, 0.16, -0.28, -0.16), tri(0.5, 0, 0.28, -0.16, 0.28, 0.16),
		)
	case Attack: // fist: four knuckles over the hand, thumb, wrist
		return solid(
			rect(-0.28, -0.12, 0.26, 0.3),
			rect(-0.28, -0.34, -0.16, -0.08), rect(-0.14, -0.36, -0.02, -0.08),
			rect(0, -0.36, 0.12, -0.08), rect(0.14, -0.32, 0.26, -0.08),
			rect(-0.4, -0.06, -0.24, 0.18), rect(-0.18, 0.3, 0.18, 0.46),
		)
	case ModeChange: // two arrows, opposite ways
		return solid(
			bar(-0.38, -0.16, 0.18, -0.16, 0.1), tri(0.44, -0.16, 0.16, 0.02, 0.16, -0.34),
			bar(-0.18, 0.16, 0.38, 0.16, 0.1), tri(-0.44, 0.16, -0.16, -0.02, -0.16, 0.34),
		)
	case Open: // box with its lid lifted, a handle slot
		return append(solid(rect(-0.36, -0.06, 0.36, 0.38), bar(-0.38, -0.16, 0.34, -0.38, 0.12)),
			hole(rect(-0.12, 0.06, 0.12, 0.14)))
	case Map: // a sheet between two rolls, two lines of writing
		return append(solid(rect(-0.3, -0.3, 0.3, 0.3), rect(-0.4, -0.42, 0.4, -0.28), rect(-0.4, 0.28, 0.4, 0.42)),
			hole(rect(-0.18, -0.14, 0.18, -0.08)), hole(rect(-0.18, 0.04, 0.1, 0.1)))
	case Info: // question mark
		return question()
	case MoveOut: // arrow up out of an open box
		return solid(
			bar(0, 0.14, 0, -0.2, 0.14), tri(0, -0.48, 0.24, -0.2, -0.24, -0.2),
			rect(-0.4, 0, -0.3, 0.42), rect(-0.4, 0.32, 0.4, 0.42), rect(0.3, 0, 0.4, 0.42),
		)
	case Produce: // hammer
		return solid(bar(-0.3, 0.4, 0.13, -0.13, 0.12), bar(-0.05, -0.35, 0.35, 0.05, 0.2))
	case Repair: // ring spanner
		return append(solid(bar(-0.34, 0.34, 0.1, -0.1, 0.14), disc(0.18, -0.18, 0.2, 16)),
			hole(disc(0.18, -0.18, 0.09, 6)))
	}
	return nil
}

func question() []polygon {
	const w, cx, cy, r = 0.13, 0, -0.16, 0.2
	var pts [][2]float64
	for deg := 190.0; deg <= 420; deg += 23 {
		a := deg * math.Pi / 180
		pts = append(pts, [2]float64{cx + r*math.Cos(a), cy + r*math.Sin(a)})
	}
	pts = append(pts, [2]float64{0, 0.08}, [2]float64{0, 0.17})
	var parts [][][2]float64
	for i := 1; i < len(pts); i++ {
		parts = append(parts, bar(pts[i-1][0], pts[i-1][1], pts[i][0], pts[i][1], w))
		if i < len(pts)-1 {
			parts = append(parts, disc(pts[i][0], pts[i][1], w/2, 8))
		}
	}
	return solid(append(parts, disc(0, 0.36, 0.08, 12))...)
}

func rect(x0, y0, x1, y1 float64) [][2]float64 {
	return [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}}
}

func tri(x0, y0, x1, y1, x2, y2 float64) [][2]float64 {
	return [][2]float64{{x0, y0}, {x1, y1}, {x2, y2}}
}

// bar is a straight stroke of width w from (x0, y0) to (x1, y1), square ends
// flush with the end points.
func bar(x0, y0, x1, y1, w float64) [][2]float64 {
	dx, dy := x1-x0, y1-y0
	l := math.Hypot(dx, dy)
	nx, ny := -dy/l*w/2, dx/l*w/2
	return [][2]float64{{x0 + nx, y0 + ny}, {x1 + nx, y1 + ny}, {x1 - nx, y1 - ny}, {x0 - nx, y0 - ny}}
}

func disc(cx, cy, r float64, n int) [][2]float64 {
	out := make([][2]float64, n)
	for i := range out {
		a := 2 * math.Pi * float64(i) / float64(n)
		out[i] = [2]float64{cx + r*math.Cos(a), cy + r*math.Sin(a)}
	}
	return out
}

func solid(parts ...[][2]float64) []polygon {
	out := make([]polygon, len(parts))
	for i, p := range parts {
		out[i] = wind(p, false)
	}
	return out
}

func hole(p [][2]float64) polygon { return wind(p, true) }

// wind orients p: positive shoelace area for a solid part, negative for a hole.
func wind(p [][2]float64, isHole bool) polygon {
	if (shoelace(p) < 0) != isHole {
		for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
	}
	return polygon{pts: p, hole: isHole}
}

func shoelace(p [][2]float64) float64 {
	var a float64
	for i := range p {
		j := (i + 1) % len(p)
		a += p[i][0]*p[j][1] - p[j][0]*p[i][1]
	}
	return a / 2
}

// inset returns the convex polygon p (positive shoelace area) with every edge
// moved inwards by d.
func inset(p [][2]float64, d float64) [][2]float64 {
	n := len(p)
	type line struct{ px, py, ex, ey float64 }
	lines := make([]line, n)
	for i := range p {
		a, b := p[i], p[(i+1)%n]
		ex, ey := b[0]-a[0], b[1]-a[1]
		l := math.Hypot(ex, ey)
		lines[i] = line{a[0] - ey/l*d, a[1] + ex/l*d, ex, ey}
	}
	out := make([][2]float64, n)
	for i := range p {
		l1, l2 := lines[(i+n-1)%n], lines[i]
		t := ((l2.px-l1.px)*l2.ey - (l2.py-l1.py)*l2.ex) / (l1.ex*l2.ey - l1.ey*l2.ex)
		out[i] = [2]float64{l1.px + t*l1.ex, l1.py + t*l1.ey}
	}
	return out
}
