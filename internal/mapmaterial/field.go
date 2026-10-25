// Package mapmaterial supplies a continuous world-space scrub blend over the
// union of the level's summer hills. Internal cell borders never enter the field.
package mapmaterial

import (
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

const (
	bucketSize = 64
	maxFalloff = 16
)

type (
	point   struct{ x, y float64 }
	segment struct{ a, b point }
	bucket  struct{ x, y int }
)

// Field is immutable after construction and can be shared by renderer caches.
type Field struct {
	hills map[hexgrid.Coord]bool
	cells map[bucket][]point
	edges map[bucket][]segment
}

// New builds only exposed edges of the actual 96 by 80 artwork polygons.
func New(lvl *level.Level) *Field {
	f := &Field{hills: make(map[hexgrid.Coord]bool), cells: make(map[bucket][]point), edges: make(map[bucket][]segment)}
	if lvl == nil {
		return f
	}
	for line, row := range lvl.Lines {
		for col, h := range row {
			if h.Terrain == "hills" {
				f.hills[hexgrid.Coord{Line: line, Col: col}] = true
			}
		}
	}
	for c := range f.hills {
		x, y := hexgrid.Center(c)
		center := point{x, y}
		buckets(x-48, y-40, x+48, y+40, func(b bucket) { f.cells[b] = append(f.cells[b], center) })
		ns := hexgrid.Around(c)
		for port, n := range ns {
			if f.hills[n] {
				continue
			}
			e := (port + 1) % 6
			a, b := vertices[e], vertices[(e+1)%6]
			a.x += x
			a.y += y
			b.x += x
			b.y += y
			s := segment{a, b}
			buckets(math.Min(a.x, b.x)-maxFalloff, math.Min(a.y, b.y)-maxFalloff, math.Max(a.x, b.x)+maxFalloff, math.Max(a.y, b.y)+maxFalloff, func(k bucket) { f.edges[k] = append(f.edges[k], s) })
		}
	}
	return f
}

var vertices = [6]point{{-23, -40}, {23, -40}, {48, 0}, {23, 40}, {-23, 40}, {-48, 0}}

func buckets(left, top, right, bottom float64, visit func(bucket)) {
	for y := int(math.Floor(top / bucketSize)); y <= int(math.Floor(bottom/bucketSize)); y++ {
		for x := int(math.Floor(left / bucketSize)); x <= int(math.Floor(right/bucketSize)); x++ {
			visit(bucket{x, y})
		}
	}
}

// Weight is zero outside the hills union and grows smoothly from its exposed
// boundary. The same world point always has the same weight, regardless of the
// hex through which it is rendered. There is no flat meadow guard inside hills.
func (f *Field) Weight(worldX, worldY float64) float64 {
	if f == nil || math.IsNaN(worldX) || math.IsNaN(worldY) || math.IsInf(worldX, 0) || math.IsInf(worldY, 0) {
		return 0
	}
	key := bucket{int(math.Floor(worldX / bucketSize)), int(math.Floor(worldY / bucketSize))}
	inside := false
	for _, c := range f.cells[key] {
		x, y := math.Abs(worldX-c.x), math.Abs(worldY-c.y)
		if y <= 40 && x <= 48-25*y/40 {
			inside = true
			break
		}
	}
	if !inside {
		return 0
	}
	distance := float64(maxFalloff)
	for _, s := range f.edges[key] {
		dx, dy := s.b.x-s.a.x, s.b.y-s.a.y
		t := ((worldX-s.a.x)*dx + (worldY-s.a.y)*dy) / (dx*dx + dy*dy)
		t = math.Max(0, math.Min(1, t))
		d := math.Hypot(worldX-s.a.x-t*dx, worldY-s.a.y-t*dy)
		distance = math.Min(distance, d)
	}
	// Smooth, repeatable variation shares the material's declared world periods.
	width := 12 + 2*math.Sin(2*math.Pi*worldX/568) + 2*math.Sin(2*math.Pi*worldY/400)
	t := math.Min(1, distance/width)
	return t * t * (3 - 2*t)
}

// Signature records a conservative local neighborhood sufficient for every
// polygon and exposed edge within the tile's maximum 16-pixel blend distance.
// The material phase separately captures the world-periodic falloff variation.
func (f *Field) Signature(at hexgrid.Coord) string {
	var b strings.Builder
	b.Grow(46)
	b.WriteByte(byte('0' + at.Line&1))
	for dy := -4; dy <= 4; dy++ {
		for dx := -2; dx <= 2; dx++ {
			on := f != nil && f.hills[hexgrid.Coord{Line: at.Line + dy, Col: at.Col + dx}]
			if on {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		}
	}
	return b.String()
}
