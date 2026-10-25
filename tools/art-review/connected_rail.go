package main

import (
	"fmt"
	"image/color"
	"math"
)

// Rail is deliberately independent of the road/stream/path geometry and paint.
// The rail centres, rather than the merged ballast outline, define each band.
// The track dimensions below belong to the 11-world-pixel ballast core; a
// family of another width scales them by width/11 (connectionRailDimsFor).
const (
	connectionRailBaseWidth = 11.
	connectionRailGauge     = 4.5
	connectionRailWidth     = .8
	connectionSleeperLength = 7.5
	connectionSleeperWidth  = 1.05
	connectionSleeperPitch  = 4.0
	connectionRailGuard     = 8.0
	// connectionRailSteelGrowth widens the steel by this much per unit of
	// scale beyond 1, less than in proportion, so wide track keeps thin rails.
	connectionRailSteelGrowth = .6
	// Winter rails (verge snow): snow cover of the ballast between the
	// sleepers and of the sleepers; the shoulders take connectionSnowVergeMix.
	connectionRailSnowBetween = .55
	connectionRailSnowDust    = .12
	// connectionRailWinterSteelMix darkens winter steel toward
	// connectionRailWinterSteel.
	connectionRailWinterSteelMix = .5
)

var connectionRailWinterSteel = color.NRGBA{42, 42, 46, 255}

// connectionRailDims is the track of one rail family in world px: the
// dimensions of the 11-pixel core scaled by width/11, the steel widening by
// connectionRailSteelGrowth per unit of scale. The port guard stays 8 world px,
// so the route centre lines do not depend on the width.
type connectionRailDims struct {
	scale, gauge, steel, sleeperLength, sleeperWidth, pitch float64
}

func connectionRailDimsFor(width float64) connectionRailDims {
	s := width / connectionRailBaseWidth
	return connectionRailDims{s, connectionRailGauge * s, connectionRailWidth + connectionRailSteelGrowth*(s-1), connectionSleeperLength * s, connectionSleeperWidth * s, connectionSleeperPitch * s}
}

type connectionRailProfile struct {
	Gauge         float64 `json:"gauge_world_px"`
	RailWidth     float64 `json:"rail_width_world_px"`
	SleeperLength float64 `json:"sleeper_length_world_px"`
	SleeperWidth  float64 `json:"sleeper_width_world_px"`
	SleeperPitch  float64 `json:"sleeper_nominal_pitch_world_px"`
	PortGuard     float64 `json:"port_guard_world_px"`
	Method        string  `json:"method"`
	Phase         string  `json:"phase"`
	Turnouts      string  `json:"turnouts"`
}

// connectionRailProfileFor reports the track of a rail family of the given
// core width.
func connectionRailProfileFor(width float64) connectionRailProfile {
	d := connectionRailDimsFor(width)
	anchors := "at 4 and 8 world px inward"
	scaled := ""
	if width != connectionRailBaseWidth {
		anchors = fmt.Sprintf("every %g world px inward to %g", d.pitch, float64(connectionRailGuard))
		scaled = fmt.Sprintf(" This is the track of the %g-world-pixel ballast core scaled by %g, the steel widened by %g per unit of scale beyond 1.", connectionRailBaseWidth, d.scale, connectionRailSteelGrowth)
	}
	return connectionRailProfile{
		d.gauge, d.steel, d.sleeperLength, d.sleeperWidth, d.pitch, connectionRailGuard,
		fmt.Sprintf("Rail only: earth border, opaque ballast, timber rectangles, then paired steel bands on separately offset route polylines; cubic curves flatten at 64 steps. Two-port bends use 20-world-pixel inward control vectors from both fixed guards. Gauge %g, steel width %g, sleeper length %g, width %g world px. Arc-length sleeper positions use local route normals. Existing 4x analytic AA, overscan/filter/crop and half-open ownership are retained.", d.gauge, d.steel, d.sleeperLength, d.sleeperWidth) + scaled,
		fmt.Sprintf("All ports anchor a sleeper centre at the port and %s, plus pitch-%g continuation outward. Each port guard is straight for %g world px. Between fixed guards fit round(interior arc length/%g) intervals of equal adjusted pitch. Shared port anchors use the lattice world phase (0,0); reciprocal half-sleepers assemble exactly at both line parities.", anchors, d.pitch, float64(connectionRailGuard), d.pitch),
		"Three-port masks retain the complete straight 3/6 spine. A separate branch merges at (48,40), tangent north for port 2 or 4, south for port 1 or 5; branch cubic controls are 16 world px inward from its guard and 12 world px opposite the merge tangent from the centre. Paired offset branch rails meet the spine rails; no star junction or merged-ballast contour rails.",
	}
}

func validConnectionRailPorts(ports []int) bool {
	if len(ports) == 2 {
		return true
	}
	if len(ports) != 3 {
		return false
	}
	have3, have6 := false, false
	for _, p := range ports {
		have3 = have3 || p == 3
		have6 = have6 || p == 6
	}
	return have3 && have6
}

type (
	connectionRailSleeper  struct{ center, tangent connectionPoint }
	connectionRailGeometry struct {
		ballast  connectionRoute
		rails    connectionRoute
		sleepers []connectionRailSleeper
		dims     connectionRailDims
	}
)

func connectionRailUnit(p connectionPoint) connectionPoint   { return p.mul(1 / p.length()) }
func connectionRailNormal(t connectionPoint) connectionPoint { return connectionPoint{-t[1], t[0]} }

func connectionRailCubic(a, b, c, d connectionPoint) []connectionPoint {
	out := []connectionPoint{a}
	for i := 1; i <= 64; i++ {
		t := float64(i) / 64
		s := 1 - t
		out = append(out, a.mul(s*s*s).add(b.mul(3*s*s*t)).add(c.mul(3*s*t*t)).add(d.mul(t*t*t)))
	}
	return out
}

func connectionRailLength(points []connectionPoint) float64 {
	length := 0.0
	for i := 1; i < len(points); i++ {
		length += points[i].sub(points[i-1]).length()
	}
	return length
}

func connectionRailAt(points []connectionPoint, distance float64) (connectionPoint, connectionPoint) {
	for i := 1; i < len(points); i++ {
		delta := points[i].sub(points[i-1])
		length := delta.length()
		if distance <= length || i == len(points)-1 {
			t := connectionRailUnit(delta)
			return points[i-1].add(t.mul(distance)), t
		}
		distance -= length
	}
	return points[0], connectionPoint{0, 1}
}

func (r *connectionRailGeometry) addTrack(points []connectionPoint, tangents map[int]connectionPoint) {
	// Mitered vertex offsets keep each rail a continuous band through curves.
	offset := func(i int, side float64) connectionPoint {
		if tangent, ok := tangents[i]; ok {
			return points[i].add(connectionRailNormal(tangent).mul(side * r.dims.gauge / 2))
		}
		before, after := connectionPoint{}, connectionPoint{}
		if i > 0 {
			before = connectionRailUnit(points[i].sub(points[i-1]))
		}
		if i+1 < len(points) {
			after = connectionRailUnit(points[i+1].sub(points[i]))
		}
		if i == 0 {
			before = after
		}
		if i == len(points)-1 {
			after = before
		}
		normal := connectionRailUnit(connectionRailNormal(before).add(connectionRailNormal(after)))
		n := connectionRailNormal(after)
		projection := normal[0]*n[0] + normal[1]*n[1]
		return points[i].add(normal.mul(side * r.dims.gauge / 2 / projection))
	}
	for i := 1; i < len(points); i++ {
		r.ballast.segments = append(r.ballast.segments, connectionSegment{points[i-1], points[i]})
		for _, side := range []float64{-1, 1} {
			r.rails.segments = append(r.rails.segments, connectionSegment{offset(i-1, side), offset(i, side)})
		}
	}
}

func (r *connectionRailGeometry) addTie(center, tangent connectionPoint) {
	// Opposite guards and the through spine may request the same sleeper.
	for _, s := range r.sleepers {
		if s.center.sub(center).length() < 1e-8 {
			return
		}
	}
	r.sleepers = append(r.sleepers, connectionRailSleeper{center, tangent})
}

func (r *connectionRailGeometry) guard(g connectionGeometry, f connectionFamily, port int) (connectionPoint, connectionPoint, connectionPoint) {
	p := g.Ports[port-1]
	inward := connectionRailUnit(g.Neighbors[port-1]).mul(-1)
	extension := f.Width/2 + f.Border + 4
	outside, inside := p.sub(inward.mul(extension)), p.add(inward.mul(connectionRailGuard))
	pitch := r.dims.pitch
	for d := 0.0; d <= connectionRailGuard; d += pitch {
		r.addTie(p.add(inward.mul(d)), inward)
	}
	for d := pitch; d < extension+pitch; d += pitch {
		r.addTie(p.sub(inward.mul(d)), inward)
	}
	return outside, inside, inward
}

func (r *connectionRailGeometry) interiorTies(points []connectionPoint, includeEnd bool) {
	length := connectionRailLength(points)
	intervals := max(1, int(math.Round(length/r.dims.pitch)))
	end := intervals
	if includeEnd {
		end++
	}
	for i := 1; i < end; i++ {
		p, t := connectionRailAt(points, length*float64(i)/float64(intervals))
		r.addTie(p, t)
	}
}

func connectionRailRoutes(g connectionGeometry, f connectionFamily, v connectionVariant) connectionRailGeometry {
	r := connectionRailGeometry{dims: connectionRailDimsFor(f.Width)}
	through := func(aPort, bPort int) {
		outsideA, a, na := r.guard(g, f, aPort)
		outsideB, b, nb := r.guard(g, f, bPort)
		interior := []connectionPoint{a, b}
		if (aPort+2)%6+1 != bPort {
			interior = connectionRailCubic(a, a.add(na.mul(20)), b.add(nb.mul(20)), b)
		}
		points := append([]connectionPoint{outsideA}, interior...)
		points = append(points, outsideB)
		r.addTrack(points, map[int]connectionPoint{1: na, len(points) - 2: nb.mul(-1)})
		r.interiorTies(interior, false)
	}
	if len(v.Ports) == 2 {
		through(v.Ports[0], v.Ports[1])
		return r
	}
	through(3, 6)
	for _, port := range v.Ports {
		if port == 3 || port == 6 {
			continue
		}
		outside, a, n := r.guard(g, f, port)
		center := connectionPoint{48, 40}
		merge := connectionPoint{0, 1}
		if port == 2 || port == 4 {
			merge = connectionPoint{0, -1}
		}
		interior := connectionRailCubic(a, a.add(n.mul(16)), center.sub(merge.mul(12)), center)
		points := append([]connectionPoint{outside}, interior...)
		r.addTrack(points, map[int]connectionPoint{1: n, len(points) - 1: merge})
		r.interiorTies(interior, false)
	}
	return r
}

func connectionRailSleeperDistance(d connectionRailDims, s connectionRailSleeper, p connectionPoint) float64 {
	q := p.sub(s.center)
	normal := connectionRailNormal(s.tangent)
	along := math.Abs(q[0]*s.tangent[0]+q[1]*s.tangent[1]) - d.sleeperWidth/2
	across := math.Abs(q[0]*normal[0]+q[1]*normal[1]) - d.sleeperLength/2
	return math.Hypot(math.Max(along, 0), math.Max(across, 0)) + math.Min(math.Max(along, across), 0)
}

// connectionRailSteel is the rail steel at p. Winter rails (verge snow) are
// darker cold steel, so they stay dark lines on the snow.
func connectionRailSteel(in connectionInputs, f connectionFamily, p connectionPoint) color.NRGBA {
	steel := sampleConnectionMaterial(in.materials["steel"], p)
	if f.Verge == "snow" {
		steel = connectionBlend(steel, connectionRailWinterSteel, connectionRailWinterSteelMix)
	}
	return steel
}

func connectionRailFeature(in connectionInputs, f connectionFamily, r connectionRailGeometry, p connectionPoint, aa float64) color.NRGBA {
	base := connectionFeature(in, f, r.ballast, p, p, aa)
	if base.A == 0 {
		return base
	}
	d := r.dims
	sleeperDistance := math.Inf(1)
	bound := 4.5 * d.scale
	for _, s := range r.sleepers {
		// A conservative bound avoids rectangle SDF work away from each sleeper.
		if math.Abs(p[0]-s.center[0]) > bound || math.Abs(p[1]-s.center[1]) > bound {
			continue
		}
		sleeperDistance = math.Min(sleeperDistance, connectionRailSleeperDistance(d, s, p))
	}
	wood := sampleConnectionMaterial(in.materials["timber"], p)
	if f.Verge == "snow" {
		// Winter: snow lies thin between the sleepers, deep on the ballast
		// shoulders and the earth border; the sleepers are dusted, the steel
		// stays bare.
		snow := sampleGrassMaterial(in.materials["grass"], p)
		shoulder := connectionSmoothStep((connectionRouteDistance(r.ballast, p) - (d.sleeperLength/2 - .5*d.scale)) / (1.25 * d.scale))
		alpha := base.A
		base = connectionBlend(base, snow, connectionRailSnowBetween+(connectionSnowVergeMix-connectionRailSnowBetween)*shoulder)
		base.A = alpha
		wood = connectionBlend(wood, snow, connectionRailSnowDust)
	}
	timber := connectionRamp(-sleeperDistance, aa)
	if timber > 0 {
		alpha := base.A
		base = connectionBlend(base, wood, timber)
		base.A = alpha
	}
	steel := connectionRamp(d.steel/2-connectionRouteDistance(r.rails, p), aa)
	if steel > 0 {
		alpha := base.A
		base = connectionBlend(base, connectionRailSteel(in, f, p), steel)
		base.A = alpha
	}
	return base
}
