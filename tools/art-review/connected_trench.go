package main

import (
	"encoding/json"
	"errors"
	"image/color"
	"math"
	"strings"
)

// Remember presence so even false/empty trench-only fields are rejected outside
// this family. The existing manifest decoder still checks duplicate object keys.
func validateConnectionTrenchFields(b []byte) error {
	var raw struct {
		Families []struct {
			ID       string                       `json:"id"`
			Variants []map[string]json.RawMessage `json:"variants"`
		} `json:"families"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	for _, f := range raw.Families {
		for _, v := range f.Variants {
			for key := range v {
				for _, canonical := range []string{"closed_ring", "blind_spurs", "dugout_direction"} {
					if !strings.EqualFold(key, canonical) {
						continue
					}
					if f.ID != "trench" {
						return errors.New("closed_ring, blind_spurs and dugout_direction are trench-only")
					}
					if key != canonical {
						return errors.New("trench fields require canonical lowercase closed_ring, blind_spurs and dugout_direction keys")
					}
				}
			}
		}
	}
	return nil
}

func validateConnectionTrenchVariant(family string, v connectionVariant) error {
	if family != "trench" {
		if v.ClosedRing || len(v.BlindSpurs) > 0 || v.DugoutDirection != 0 {
			return errors.New("closed_ring, blind_spurs and dugout_direction are trench-only")
		}
		return nil
	}
	if len(v.Ports) > 6 || (len(v.Ports) == 0) != v.ClosedRing {
		return errors.New("trench requires 1..6 ports or a closed ring with no ports")
	}
	if v.ClosedRing && (len(v.BlindSpurs) > 0 || v.DugoutDirection != 0) {
		return errors.New("closed trench ring cannot have blind spurs or a dugout")
	}
	occupied := map[int]bool{}
	for _, p := range v.Ports {
		occupied[p] = true
	}
	for _, p := range v.BlindSpurs {
		if p < 1 || p > 6 || occupied[p] {
			return errors.New("blind spurs must be unique 1..6 and disjoint from ports")
		}
		occupied[p] = true
	}
	if v.DugoutDirection < 0 || v.DugoutDirection > 6 || v.DugoutDirection != 0 && occupied[v.DugoutDirection] {
		return errors.New("dugout direction must be 0..6 and disjoint from ports and blind spurs")
	}
	return nil
}

type connectionTrenchProfile struct {
	PortGuard               float64         `json:"port_guard_world_px"`
	InteriorStraight        float64         `json:"interior_straight_world_px"`
	TraverseOffset          float64         `json:"traverse_offset_world_px"`
	CapBeyondCenter         float64         `json:"cap_beyond_center_world_px"`
	BlindSpurLength         float64         `json:"blind_spur_length_world_px"`
	DugoutOffset            float64         `json:"dugout_offset_world_px"`
	DugoutNeckMaxWidth      float64         `json:"dugout_neck_max_width_world_px"`
	DugoutNeckWidthFraction float64         `json:"dugout_neck_width_fraction"`
	ChamberSize             connectionPoint `json:"chamber_floor_size_world_px"`
	RingSize                connectionPoint `json:"ring_centerline_size_world_px"`
	RingCornerRadius        float64         `json:"ring_corner_radius_world_px"`
	RevetmentWidth          float64         `json:"revetment_width_world_px"`
	Method                  string          `json:"method"`
	Courses                 string          `json:"courses"`
	Guards                  string          `json:"guards"`
	Shading                 string          `json:"shading"`
}

func fixedConnectionTrenchProfile() connectionTrenchProfile {
	return connectionTrenchProfile{8, 16, 2, 14, 14, 15, 3.5, .6, connectionPoint{10, 9}, connectionPoint{36, 26}, 7, 1.1,
		"Trench only: new dark earth floor, shaded new earth banks and new timber revetment/braces. Geometry uses numeric runtime ports only; no original pixels, masks or outlines. Candidate appearance requires visual review.",
		"Each port runs straight 16 world px inward, then via a point 13 world px from center with a 2-world-pixel transverse offset to center. A single port with no explicit blind spurs extends 14 world px beyond center and ends in a round cap; explicit blind spurs supersede that implicit tail. Blind spurs run center to unit-neighbor*14 with round caps. A dugout uses an independent narrow neck from center to unit-neighbor*15 with floor width min(0.6*width,3.5), no width wear, and a separate rounded 10x9 floor chamber (corner radius 2), aligned with its direction. The neck is excluded from the full-width main course. Ring is a new 36x26 rounded rectangle centerline with radius 7, sampled 16 segments per quarter; center remains open.",
		"The first 8 world px from every polygon edge have fixed width/bank and no wear. Straight continuation to 16 px keeps angular courses and their radius outside that guard. Width wear is 0.18*E*sin(3A+B), bank fringe is 0.25*E*sin(2B-A), E=S((edge_distance-8)/8), A=2pi(x/142+y/80), B=2pi(x/142-y/80). Texture, wood phase and shading repeat on the runtime lattice; actual exported PNG joins use existing overscan/Lanczos/final ownership and ordinary decoded blits.",
		"Let N be the unit vector from the nearest main-course/neck point to the sample, or the chamber SDF gradient (central differences at +/-0.01 world px) for a chamber-owned sample; N=(0,0) at length <=1e-9. D=round(dot(N,(0.6,0.8))*1e9)/1e9. Wall=S((core_distance+2.25)/2.25). Floor samples trench at 0.72*(1-0.6*Wall*(0.5+0.5*D)) brightness, producing a consistent southeast inner-wall shadow. Banks sample earth at (0.94+0.1*sin(A+0.4)*cos(B))*(1-0.15*D) brightness. Timber is source*1.08 brightness; revetment is a 1.1-world-pixel band centered 0.1 px inside the floor edge. Braces use clamp(0.5+(sin(3A+2B)-0.97)/0.06,0,1) and reach 1.2 world px from the floor edge. RGB rounds to nearest byte and clamps to 255. Reciprocal straight guards share nearest-course world normals, textures and phases. Core remains opaque; only the exterior 0.25-world-pixel analytic AA ramp is translucent. No baked grass."}
}

type connectionTrenchGeometry struct {
	course                     connectionRoute
	neck                       connectionRoute
	chamber                    bool
	chamberCenter, chamberAxis connectionPoint
}

func connectionTrenchRoutes(g connectionGeometry, f connectionFamily, v connectionVariant) connectionTrenchGeometry {
	var r connectionTrenchGeometry
	line := func(a, b connectionPoint) { r.course.segments = append(r.course.segments, connectionSegment{a, b}) }
	center := connectionPoint{48, 40}
	if v.ClosedRing {
		corners := []connectionPoint{{59, 34}, {59, 46}, {37, 46}, {37, 34}}
		var first, previous connectionPoint
		for i, c := range corners {
			for j := 0; j <= 16; j++ {
				angle := (-math.Pi/2 + float64(i)*math.Pi/2) + float64(j)*math.Pi/32
				p := c.add(connectionPoint{7 * math.Cos(angle), 7 * math.Sin(angle)})
				if i == 0 && j == 0 {
					first = p
				} else {
					line(previous, p)
				}
				previous = p
			}
		}
		line(previous, first)
		return r
	}
	for _, port := range v.Ports {
		p := g.Ports[port-1]
		axis := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
		normal := connectionPoint{-axis[1], axis[0]}
		outside := p.add(axis.mul(f.Width/2 + f.Border + 4))
		guard := p.sub(axis.mul(16))
		traverse := center.add(axis.mul(13)).add(normal.mul(2))
		line(outside, guard)
		line(guard, traverse)
		line(traverse, center)
		if len(v.Ports) == 1 && len(v.BlindSpurs) == 0 {
			line(center, center.sub(axis.mul(14)))
		}
	}
	for _, port := range v.BlindSpurs {
		axis := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
		line(center, center.add(axis.mul(14)))
	}
	if v.DugoutDirection != 0 {
		axis := g.Neighbors[v.DugoutDirection-1].mul(1 / g.Neighbors[v.DugoutDirection-1].length())
		r.chamber = true
		r.chamberAxis = axis
		r.chamberCenter = center.add(axis.mul(15))
		r.neck.segments = append(r.neck.segments, connectionSegment{center, r.chamberCenter})
	}
	return r
}

func connectionTrenchChamberDistance(r connectionTrenchGeometry, p connectionPoint) float64 {
	q := p.sub(r.chamberCenter)
	axis := r.chamberAxis
	x := math.Abs(q[0]*axis[0]+q[1]*axis[1]) - 3
	y := math.Abs(-q[0]*axis[1]+q[1]*axis[0]) - 2.5
	return math.Hypot(math.Max(x, 0), math.Max(y, 0)) + math.Min(math.Max(x, y), 0) - 2
}

// A closest-point normal has no dependence on centerline traversal direction,
// so reciprocal guards receive the same directional light and wall shadow.
func connectionTrenchRouteSample(route connectionRoute, p connectionPoint) (float64, connectionPoint) {
	best := math.Inf(1)
	var normal connectionPoint
	for _, s := range route.segments {
		delta := s.b.sub(s.a)
		lengthSquared := delta[0]*delta[0] + delta[1]*delta[1]
		q := p.sub(s.a)
		t := 0.0
		if lengthSquared > 0 {
			t = math.Max(0, math.Min(1, (q[0]*delta[0]+q[1]*delta[1])/lengthSquared))
		}
		offset := q.sub(delta.mul(t))
		distance := offset.length()
		if distance < best {
			best = distance
			normal = connectionPoint{}
			if distance > 1e-9 {
				normal = offset.mul(1 / distance)
			}
		}
	}
	return best, normal
}
func connectionTrenchShade(c color.NRGBA, factor float64) color.NRGBA {
	return color.NRGBA{uint8(math.Min(255, math.Floor(float64(c.R)*factor+.5))), uint8(math.Min(255, math.Floor(float64(c.G)*factor+.5))), uint8(math.Min(255, math.Floor(float64(c.B)*factor+.5))), 255}
}
func connectionTrenchFeature(in connectionInputs, f connectionFamily, r connectionTrenchGeometry, p connectionPoint, aa float64) color.NRGBA {
	edgeDistance := math.Inf(1)
	g := in.manifest.Geometry
	for i, a := range g.Polygon {
		edge := g.Polygon[(i+1)%len(g.Polygon)].sub(a)
		q := p.sub(a)
		edgeDistance = math.Min(edgeDistance, (edge[0]*q[1]-edge[1]*q[0])/edge.length())
	}
	envelope := connectionSmoothStep((edgeDistance - 8) / 8)
	phaseA := 2 * math.Pi * (p[0]/142 + p[1]/80)
	phaseB := 2 * math.Pi * (p[0]/142 - p[1]/80)
	distance, normal := connectionTrenchRouteSample(r.course, p)
	coreDistance := distance - f.Width/2 - .18*envelope*math.Sin(3*phaseA+phaseB)
	if r.chamber {
		neckDistance, neckNormal := connectionTrenchRouteSample(r.neck, p)
		neckCore := neckDistance - math.Min(.6*f.Width, 3.5)/2
		if neckCore < coreDistance {
			coreDistance, normal = neckCore, neckNormal
		}
		chamber := connectionTrenchChamberDistance(r, p)
		if chamber < coreDistance {
			coreDistance = chamber
			normal = connectionPoint{
				connectionTrenchChamberDistance(r, p.add(connectionPoint{.01, 0})) - connectionTrenchChamberDistance(r, p.sub(connectionPoint{.01, 0})),
				connectionTrenchChamberDistance(r, p.add(connectionPoint{0, .01})) - connectionTrenchChamberDistance(r, p.sub(connectionPoint{0, .01})),
			}
			if normal.length() > 1e-9 {
				normal = normal.mul(1 / normal.length())
			}
		}
	}
	border := f.Border + .25*envelope*math.Sin(2*phaseB-phaseA)
	outer := connectionRamp(border-coreDistance, aa)
	if outer == 0 {
		return color.NRGBA{}
	}
	core := connectionRamp(-coreDistance, aa)
	// World-periodic lighting keeps reciprocal ports identical regardless of the
	// direction in which their independent centerlines were constructed.
	shade := .94 + .1*math.Sin(phaseA+.4)*math.Cos(phaseB)
	directional := math.Round((.6*normal[0]+.8*normal[1])*1e9) / 1e9
	wall := connectionSmoothStep((coreDistance + 2.25) / 2.25)
	bank := connectionTrenchShade(sampleConnectionMaterial(in.materials["earth"], p), shade*(1-.15*directional))
	floor := connectionTrenchShade(sampleConnectionMaterial(in.materials["trench"], p), .72*(1-.6*wall*(.5+.5*directional)))
	paint := connectionBlend(bank, floor, core)
	wood := connectionRamp(.55-math.Abs(coreDistance+.1), aa)
	brace := connectionRamp(math.Sin(3*phaseA+2*phaseB)-.97, .06) * connectionRamp(1.2-math.Abs(coreDistance), aa)
	timber := math.Max(wood, brace)
	if timber > 0 {
		paint = connectionBlend(paint, connectionTrenchShade(sampleConnectionMaterial(in.materials["timber"], p), 1.08), timber)
	}
	paint.A = uint8(math.Floor(255*outer + .5))
	return paint
}
