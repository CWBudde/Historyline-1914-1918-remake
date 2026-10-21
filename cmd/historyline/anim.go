package main

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"

// stepFrames is how many frames a unit takes to cross one hex.
const stepFrames = 8

// walk is a unit moving along the path the rules engine chose. The engine has
// already placed the unit on the last hex; the walk only decides where to draw it.
type walk struct {
	path  []hexgrid.Coord
	unit  any // the engine's key of the walking unit; the last hex may hold its building or carrier
	frame int
}

func newWalk(path []hexgrid.Coord) *walk {
	return newWalkMotion(path, nil, false)
}

// newWalkMotion is the walk of u along path; with reduced set the unit is
// placed at once (nil walk).
func newWalkMotion(path []hexgrid.Coord, u any, reduced bool) *walk {
	if reduced || len(path) < 2 {
		return nil
	}
	return &walk{path: path, unit: u}
}

// Done reports whether the unit has reached the end of the path.
func (w *walk) Done() bool { return w == nil || w.frame >= (len(w.path)-1)*stepFrames }

// Tick advances one frame.
func (w *walk) Tick() {
	if !w.Done() {
		w.frame++
	}
}

// Heading is the map-pixel direction of the step being walked; once the walk is
// done it is the last step.
func (w *walk) Heading() (dx, dy float64) {
	i := min(w.frame/stepFrames, len(w.path)-2)
	x0, y0 := hexgrid.Center(w.path[i])
	x1, y1 := hexgrid.Center(w.path[i+1])
	return x1 - x0, y1 - y0
}

// Hex is the hex the unit is drawn over: the step's start until it is halfway
// to the next hex.
func (w *walk) Hex() hexgrid.Coord {
	last := len(w.path) - 1
	if w.Done() {
		return w.path[last]
	}
	return w.path[min((w.frame+stepFrames/2)/stepFrames, last)]
}

// Pos is the world position of the unit, interpolated between hex centres.
func (w *walk) Pos() (x, y float64) {
	last := len(w.path) - 1
	if w.Done() {
		return hexgrid.Center(w.path[last])
	}
	i, f := w.frame/stepFrames, float64(w.frame%stepFrames)/stepFrames
	x0, y0 := hexgrid.Center(w.path[i])
	x1, y1 := hexgrid.Center(w.path[i+1])
	return x0 + (x1-x0)*f, y0 + (y1-y0)*f
}
