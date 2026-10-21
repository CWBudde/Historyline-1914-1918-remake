package main

import (
	"image"
	"image/color"
	"math"
	"slices"
)

type shoreSegment struct {
	a, delta             connectionPoint
	inverseLengthSquared float64
}
type shoreCurve struct {
	segments []shoreSegment
	landSign float64
}

// All bounded profiles have one land arc and exactly two mixed edges. Equal,
// collinear first two quintic handles give zero curvature at both anchors.
func shoreCurveFor(v shoreVariant) shoreCurve {
	g := sharedConnectionGeometry()
	var ports, normals []connectionPoint
	landSign := 1.
	for i, a := range g.Polygon {
		j := (i + 1) % 6
		landA, landB := slices.Contains(v.LandVertices, i), slices.Contains(v.LandVertices, j)
		if landA == landB {
			continue
		}
		if len(ports) == 0 && !landA {
			landSign = -1
		}
		e := g.Polygon[j].sub(a)
		length := e.length()
		ports = append(ports, a.add(g.Polygon[j]).mul(.5))
		normals = append(normals, connectionPoint{-e[1] / length, e[0] / length})
	}
	curve := shoreCurve{landSign: landSign}
	if len(ports) != 2 {
		return curve
	}
	a, b := ports[0].add(normals[0].mul(8)), ports[1].add(normals[1].mul(8))
	chord := b.sub(a)
	handle := .24 * chord.length()
	controls := [6]connectionPoint{a, a.add(normals[0].mul(handle)), a.add(normals[0].mul(2 * handle)), b.add(normals[1].mul(2 * handle)), b.add(normals[1].mul(handle)), b}
	normal := connectionPoint{-chord[1], chord[0]}.mul(1 / chord.length())
	mask := 0
	for _, vertex := range v.LandVertices {
		mask |= 1 << vertex
	}
	// Opposing approaches have room for a broad bow. Existing turning profiles
	// already carry their large-scale bend in the Bezier and need less offset.
	opposition := -(normals[0][0]*normals[1][0] + normals[0][1]*normals[1][1])
	straight := connectionSmoothStep((opposition - .6) / .4)
	amplitude := 6 + .5*float64((mask*7)%5)
	if mask%4 < 2 {
		amplitude = -amplitude
	}
	skew := 5 + .75*float64((mask*11)%5)
	if chord[0] >= 0 || chord[1] < 0 {
		skew = -skew
	}
	amplitude = landSign + straight*(amplitude-landSign)
	skew *= straight
	add := func(a, b connectionPoint) {
		delta := b.sub(a)
		curve.segments = append(curve.segments, shoreSegment{a, delta, 1 / (delta[0]*delta[0] + delta[1]*delta[1])})
	}
	previous := ports[0].sub(normals[0].mul(160))
	add(previous, a)
	previous = a
	for step := 1; step <= 256; step++ {
		t := float64(step) / 256
		u := 1 - t
		weights := [6]float64{u * u * u * u * u, 5 * u * u * u * u * t, 10 * u * u * u * t * t, 10 * u * u * t * t * t, 5 * u * t * t * t * t, t * t * t * t * t}
		p := connectionPoint{}
		for i, w := range weights {
			p = p.add(controls[i].mul(w))
		}
		// A single asymmetric polynomial bow gives the tile an irregular coast,
		// rather than a repeating wiggle. Its first two endpoint derivatives vanish.
		envelope := 64 * t * t * t * u * u * u
		broad := envelope * (amplitude + skew*(2*t-1))
		// Retain the established approach beyond the 8px guard and Lanczos halo.
		// The quintic handover is C2, like the endpoints themselves.
		blend := math.Max(0, math.Min(1, (math.Min(t, u)-.08)/.12))
		blend = blend * blend * blend * (10 + blend*(-15+6*blend))
		old := .7 * envelope * math.Sin(2*math.Pi*t+.4)
		organic := old + blend*(broad-old)
		p = p.add(normal.mul(organic))
		add(previous, p)
		previous = p
	}
	add(previous, ports[1].sub(normals[1].mul(160)))
	return curve
}
func (c shoreCurve) signedDistance(p connectionPoint) float64 {
	best, side := math.Inf(1), 1.
	for _, s := range c.segments {
		q := p.sub(s.a)
		t := math.Max(0, math.Min(1, (q[0]*s.delta[0]+q[1]*s.delta[1])*s.inverseLengthSquared))
		dx, dy := q[0]-t*s.delta[0], q[1]-t*s.delta[1]
		distance := dx*dx + dy*dy
		if distance < best {
			best = distance
			side = 1
			if s.delta[0]*q[1]-s.delta[1]*q[0] < 0 {
				side = -1
			}
		}
	}
	return math.Sqrt(best) * side * c.landSign
}

// Exact reciprocal guard through 8 world pixels, including exterior overscan. Beyond
// it the sole field is coast distance; no nearest-edge blend enters interior.
func shoreField(v shoreVariant, p connectionPoint) float64 { return shoreFieldWithCurve(v, p, nil) }
func shoreFieldWithCurve(v shoreVariant, p connectionPoint, curve *shoreCurve) float64 {
	g := sharedConnectionGeometry()
	edge, minDistance := 0, math.Inf(1)
	for i, a := range g.Polygon {
		b := g.Polygon[(i+1)%6]
		e := b.sub(a)
		q := p.sub(a)
		d := (e[0]*q[1] - e[1]*q[0]) / e.length()
		if d < minDistance {
			minDistance, edge = d, i
		}
	}
	if minDistance <= 8 {
		a, b := g.Polygon[edge], g.Polygon[(edge+1)%6]
		landA, landB := slices.Contains(v.LandVertices, edge), slices.Contains(v.LandVertices, (edge+1)%6)
		if landA && landB {
			return 32
		}
		if !landA && !landB {
			return -32
		}
		e := b.sub(a)
		q := p.sub(a.add(b).mul(.5))
		guard := (e[0]*q[0] + e[1]*q[1]) / e.length()
		if landA {
			guard = -guard
		}
		return guard
	}
	if len(v.LandVertices) == 0 {
		return -32
	}
	if len(v.LandVertices) == 6 {
		return 32
	}
	if curve == nil {
		built := shoreCurveFor(v)
		curve = &built
	}
	return curve.signedDistance(p)
}
func shoreColor(in shoreInputs, v shoreVariant, p connectionPoint) color.NRGBA {
	return shoreColorWithCurve(in, v, p, nil)
}
func shoreColorWithCurve(in shoreInputs, v shoreVariant, p connectionPoint, curve *shoreCurve) color.NRGBA {
	return shoreMaterialColor(in, p, shoreFieldWithCurve(v, p, curve))
}

// The bank's broken outline uses only lattice-periodic coordinates and material
// samples, so neighboring variants paint identical crossing profiles. Geometry
// keeps the established land/water endpoints, including the lakeside23 mask.
func shoreMaterialColor(in shoreInputs, p connectionPoint, d float64) color.NRGBA {
	if in.Manifest.BankProfile == authoredBankProfile {
		return authoredShoreColor(in, p, d)
	}
	grass := sampleGrassMaterial(in.Materials["grass"].Image, p)
	water := sampleConnectionMaterial(in.Materials["water"].Image, p)
	if d >= 8 {
		return grass
	}
	if d <= -7 {
		return water
	}
	bank := sampleConnectionMaterial(in.Materials["bank"].Image, p)
	u, v := 2*math.Pi*p[0]/71, 2*math.Pi*p[1]/40
	broken := .65*math.Sin(3*u+2*v) + .35*math.Sin(7*u-5*v)
	stone := (float64(bank.R)+float64(bank.G)+float64(bank.B))/765 - .5
	d += broken + .65*stone
	// DOS-inspired pale gravel above a narrow dark wet lip, with a local
	// blue-green submerged shelf. Fine material detail survives at 4x.
	shallow := connectionBlend(water, color.NRGBA{97, 159, 152, 255}, .28)
	wet := connectionBlend(bank, color.NRGBA{38, 65, 58, 255}, .62)
	c := connectionBlend(water, shallow, connectionSmoothStep((d+5.5)/2.5))
	c = connectionBlend(c, wet, connectionSmoothStep((d+2.2)/.9))
	c = connectionBlend(c, bank, connectionSmoothStep((d+.9)/1.4))
	return connectionBlend(c, grass, connectionSmoothStep((d-3.8)/2.4))
}
func renderShoreRaw(in shoreInputs, v shoreVariant) *image.NRGBA {
	curve := shoreCurveFor(v)
	im := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			im.SetNRGBA(x, y, shoreColorWithCurve(in, v, p, &curve))
		}
	}
	return im
}
func exportShoreTiles(in shoreInputs, v shoreVariant) (map[int]*image.NRGBA, error) {
	if v.ID == "grass" {
		return exportGrassBaseTiles(in.Materials[v.ID].Image)
	}
	if v.ID == "water" && in.Manifest.BankProfile == "" {
		return exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": in.Materials[v.ID].Image}}, nil, nil)
	}
	raw4 := renderShoreRaw(in, v)
	raw2, err := resizeLanczos(raw4, 208, 176)
	if err != nil {
		return nil, err
	}
	out := map[int]*image.NRGBA{}
	for _, scale := range []int{2, 4} {
		raw := raw4
		if scale == 2 {
			raw = raw2
		}
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		for y := range 80 * scale {
			for x := range 96 * scale {
				if connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}) {
					im.SetNRGBA(x, y, raw.NRGBAAt(x+4*scale, y+4*scale))
				}
			}
		}
		out[scale] = im
	}
	return out, nil
}
