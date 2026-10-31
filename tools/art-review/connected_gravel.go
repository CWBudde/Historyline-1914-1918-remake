package artreview

import (
	"image/color"
	"math"
)

// The same scalar is zero throughout every eight-world-pixel port guard.
// All nonconstant detail uses lattice-periodic world coordinates, so unrelated
// bends paint an identical crossing profile at a reciprocal edge.
func connectionVergeEnvelope(g connectionGeometry, p connectionPoint) float64 {
	distance := math.Inf(1)
	for i, a := range g.Polygon {
		edge := g.Polygon[(i+1)%len(g.Polygon)].sub(a)
		q := p.sub(a)
		distance = math.Min(distance, (edge[0]*q[1]-edge[1]*q[0])/edge.length())
	}
	return connectionSmoothStep((distance - 8) / 8)
}

func connectionRoadVergeWear(g connectionGeometry, local, world connectionPoint) float64 {
	e := connectionVergeEnvelope(g, local)
	a, b := 2*math.Pi*(world[0]/142+world[1]/80), 2*math.Pi*(world[0]/142-world[1]/80)
	return .65 * e * (.65*math.Sin(3*a+2*b) + .35*math.Sin(7*a-5*b))
}

func connectionGravelStreamProfile(g connectionGeometry, f connectionFamily, route connectionRoute, local, world connectionPoint) (float64, float64, float64) {
	core, _, _ := connectionStreamProfile(g, f, route, local, world)
	e := connectionVergeEnvelope(g, local)
	a, b := 2*math.Pi*(world[0]/142+world[1]/80), 2*math.Pi*(world[0]/142-world[1]/80)
	border := math.Max(0, f.Border+1.1*e*math.Sin(3*a+2*b))
	pocket := connectionSmoothStep((math.Sin(a)*math.Sin(b) - .15) / .65)
	return core, core - border, 1 - .2*e*pocket
}

func connectionGravelStreamColor(in connectionInputs, f connectionFamily, p connectionPoint, distance float64) color.NRGBA {
	distance = math.Round(distance*1e9) / 1e9
	water := sampleConnectionMaterial(in.materials[f.Material], p)
	gravel := sampleConnectionMaterial(in.materials[f.BorderMaterial], p)
	shallow := connectionBlend(water, color.NRGBA{84, 135, 139, 255}, .24)
	wet := connectionBlend(gravel, color.NRGBA{35, 50, 43, 255}, .72)
	paint := connectionBlend(water, shallow, connectionSmoothStep((distance+4)/3))
	a, b := 2*math.Pi*(p[0]/142+p[1]/80), 2*math.Pi*(p[0]/142-p[1]/80)
	// Broken current flecks are localized to the water side of the bank.
	ripple := connectionSmoothStep((math.Sin(7*a-5*b+.3*math.Sin(3*a+4*b)) - .86) / .14)
	near := connectionSmoothStep((distance+4.5)/2) * (1 - connectionSmoothStep((distance+.8)/.8))
	paint = connectionBlend(paint, color.NRGBA{153, 192, 193, 255}, .24*ripple*near)
	crest := connectionSmoothStep((math.Sin(6*a+4*b+.8*math.Sin(3*b)) - .15) / .6)
	foamDistance := (distance + 1.2) / .45
	foam := math.Exp(-foamDistance*foamDistance) * crest
	paint = connectionBlend(paint, color.NRGBA{215, 231, 226, 255}, .72*foam)
	paint = connectionBlend(paint, wet, connectionSmoothStep((distance+.45)/.7))
	return connectionBlend(paint, gravel, connectionSmoothStep((distance-.25)/.8))
}
