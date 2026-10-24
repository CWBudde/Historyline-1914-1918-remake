package main

import (
	"image"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// Derived shores (mapcompose.ShoreShape) cover any arrangement of water and
// land around a water hex. A shape with one land arc and no channel is one of
// the roster's profiles and renders through shoreField unchanged. Otherwise
// the land is the union of one coast lobe per arc, and water is carved back
// through channels (an edge whose two land vertices still lead to a water
// neighbour) and a central pool.
const (
	shoreChannelHalf = 10 // world px of water either side of a channel edge's midpoint
	shoreChannelPool = 17 // world px radius of the pool the channels meet in
	shorePondRadius  = 21 // world px radius of a pond with no channel, 15 % wider across
)

type shoreShapeField struct {
	shape   mapcompose.ShoreShape
	variant *shoreVariant // the roster path: one arc, no channel
	curve   shoreCurve
	lobes   []shoreCurve
}

// shoreArcs splits a land mask into its runs of consecutive land vertices.
func shoreArcs(land uint8) [][]int {
	if land == 0 || land == 0x3f {
		return nil
	}
	start := 0
	for land&(1<<start) == 0 || land&(1<<((start+5)%6)) != 0 {
		start++
	}
	var arcs [][]int
	for i := range 6 {
		v := (start + i) % 6
		if land&(1<<v) == 0 {
			continue
		}
		if i == 0 || land&(1<<((v+5)%6)) == 0 {
			arcs = append(arcs, nil)
		}
		arcs[len(arcs)-1] = append(arcs[len(arcs)-1], v)
	}
	return arcs
}

func newShoreShapeField(s mapcompose.ShoreShape) shoreShapeField {
	f := shoreShapeField{shape: s}
	arcs := shoreArcs(s.Land)
	if s.Channels == 0 && len(arcs) == 1 {
		v := shoreVariant{ID: s.ID(), LandVertices: arcs[0]}
		f.variant, f.curve = &v, shoreCurveFor(v)
		return f
	}
	for _, arc := range arcs {
		f.lobes = append(f.lobes, shoreCurveFor(shoreVariant{LandVertices: arc}))
	}
	return f
}

// field is the signed coast distance at p, positive on land. The band within
// eight world px of an edge depends only on that edge, so the tiles on both
// sides of it agree.
func (f shoreShapeField) field(p connectionPoint) float64 {
	if f.variant != nil {
		return shoreFieldWithCurve(*f.variant, p, &f.curve)
	}
	g := sharedConnectionGeometry()
	edge, minDistance := 0, math.Inf(1)
	for i, a := range g.Polygon {
		e := g.Polygon[(i+1)%6].sub(a)
		q := p.sub(a)
		if d := (e[0]*q[1] - e[1]*q[0]) / e.length(); d < minDistance {
			minDistance, edge = d, i
		}
	}
	if minDistance <= 8 && f.shape.Channels&(1<<edge) != 0 {
		a, b := g.Polygon[edge], g.Polygon[(edge+1)%6]
		e := b.sub(a)
		along := (e[0]*(p[0]-(a[0]+b[0])/2) + e[1]*(p[1]-(a[1]+b[1])/2)) / e.length()
		return carve(math.Abs(along), shoreChannelHalf)
	}
	if minDistance <= 8 {
		var land []int
		for v := range 6 {
			if f.shape.Land&(1<<v) != 0 {
				land = append(land, v)
			}
		}
		return shoreFieldWithCurve(shoreVariant{LandVertices: land}, p, &shoreCurve{})
	}
	land := 32.
	if f.shape.Land == 0 {
		land = -32
	}
	if len(f.lobes) > 0 {
		land = math.Inf(-1)
		for _, c := range f.lobes {
			land = max(land, c.signedDistance(p))
		}
	}
	centre := connectionPoint{48, 40}
	// Inside the tile the carved banks wander; the edge bands stay exact.
	envelope, patch := authoredShoreVariation(p)
	wobble := 2.5 * envelope * patch
	if f.shape.Land == 0x3f && f.shape.Channels == 0 {
		// A pond is an irregular oval; it stays clear of the edge bands.
		q := p.sub(centre)
		theta := math.Atan2(q[1], q[0])
		rim := shorePondRadius + 2*math.Sin(3*theta+1.3) + 1.2*math.Sin(5*theta+.4)
		return min(land, carve(math.Hypot(q[0]/1.15, q[1]), rim))
	}
	if f.shape.Channels == 0 {
		return land
	}
	water := carve(p.sub(centre).length(), shoreChannelPool+wobble)
	for edge := range 6 {
		if f.shape.Channels&(1<<edge) == 0 {
			// The pool also reaches the tile's other water: open edges and the
			// water half of mixed edges, stopping short of the edge band.
			a, b := g.Polygon[edge], g.Polygon[(edge+1)%6]
			landA, landB := f.shape.Land&(1<<edge) != 0, f.shape.Land&(1<<((edge+1)%6)) != 0
			if landA && landB {
				continue
			}
			e := b.sub(a)
			target := a.add(b).mul(.5)
			if landA != landB {
				toward := .25
				if landB {
					toward = -.25
				}
				target = target.add(e.mul(toward))
			}
			inward := connectionPoint{-e[1], e[0]}.mul(1 / e.length())
			target = target.add(inward.mul(16))
			water = min(water, carve(segmentDistance(p, centre, target), 9+wobble))
			continue
		}
		// The channel leaves its edge along the normal, so inside the band
		// its distance is the band's, then bends to the pool.
		a, b := g.Polygon[edge], g.Polygon[(edge+1)%6]
		e := b.sub(a).mul(1 / b.sub(a).length())
		mid := a.add(b).mul(.5)
		inner := mid.add(connectionPoint{-e[1], e[0]}.mul(10))
		d := min(segmentDistance(p, mid, inner), segmentDistance(p, inner, centre))
		water = min(water, carve(d, shoreChannelHalf+wobble))
	}
	return min(land, water)
}

func segmentDistance(p, a, b connectionPoint) float64 {
	ab, ap := b.sub(a), p.sub(a)
	t := max(0, min(1, (ap[0]*ab[0]+ap[1]*ab[1])/(ab[0]*ab[0]+ab[1]*ab[1])))
	return p.sub(a.add(ab.mul(t))).length()
}

// exportShoreShapeTiles renders shape s as 2x and 4x hex tiles.
func exportShoreShapeTiles(in shoreInputs, s mapcompose.ShoreShape) (map[int]*image.NRGBA, error) {
	f := newShoreShapeField(s)
	raw := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			raw.SetNRGBA(x, y, shoreMaterialColor(in, p, f.field(p)))
		}
	}
	return cropShoreTiles(raw)
}

// carve is the field at distance d from a carved channel or pond of radius r.
// It is twice the distance in the water, so the shallows match a coast's. On
// land it first follows the distance through the rock, then rises four times
// as fast, so a channel's bank has turned to meadow before its edge band meets
// a corner.
func carve(d, r float64) float64 {
	switch x := d - r; {
	case x < 0:
		return 2 * x
	case x < 4:
		return x
	default:
		return 4 + 4*(x-4)
	}
}
