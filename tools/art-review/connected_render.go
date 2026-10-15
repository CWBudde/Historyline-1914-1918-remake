package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"
)

// sampleConnectionMaterial derives a continuous texture on the runtime lattice.
// Mirror wrapping keeps values continuous across periods; averaging both lattice
// cosets makes the half-period (71,40) displacement invariant as well.
func sampleConnectionMaterial(src *image.NRGBA, p connectionPoint) color.NRGBA {
	mirror := func(v, period float64) float64 {
		v = math.Mod(v, period)
		if v < 0 {
			v += period
		}
		return 1 - math.Abs(2*v/period-1)
	}
	sample := func(p connectionPoint) [3]float64 {
		x := mirror(p[0], 142) * float64(src.Bounds().Dx()-1)
		y := mirror(p[1], 80) * float64(src.Bounds().Dy()-1)
		ix, iy := int(x), int(y)
		fx, fy := x-float64(ix), y-float64(iy)
		var out [3]float64
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				weight := (1 - fx) * (1 - fy)
				if dx == 1 {
					weight = fx * (1 - fy)
				}
				if dy == 1 {
					weight = (1 - fx) * fy
					if dx == 1 {
						weight = fx * fy
					}
				}
				c := src.NRGBAAt(min(ix+dx, src.Bounds().Dx()-1), min(iy+dy, src.Bounds().Dy()-1))
				out[0] += float64(c.R) * weight
				out[1] += float64(c.G) * weight
				out[2] += float64(c.B) * weight
			}
		}
		return out
	}
	a, b := sample(p), sample(p.add(connectionPoint{71, 40}))
	return color.NRGBA{uint8(math.Floor((a[0]+b[0])/2 + .5000000001)), uint8(math.Floor((a[1]+b[1])/2 + .5000000001)), uint8(math.Floor((a[2]+b[2])/2 + .5000000001)), 255}
}

type connectionSegment struct{ a, b connectionPoint }
type connectionRoute struct {
	segments []connectionSegment
	junction bool
	road     *connectionRoadPaint
}

func connectionRoutes(g connectionGeometry, f connectionFamily, v connectionVariant) connectionRoute {
	route := connectionRoute{junction: len(v.Ports) >= 3}
	line := func(a, b connectionPoint) { route.segments = append(route.segments, connectionSegment{a, b}) }
	cubic := func(a, b, c, d connectionPoint) {
		prev := a
		for i := 1; i <= 64; i++ {
			t := float64(i) / 64
			s := 1 - t
			p := a.mul(s * s * s).add(b.mul(3 * s * s * t)).add(c.mul(3 * s * t * t)).add(d.mul(t * t * t))
			line(prev, p)
			prev = p
		}
	}
	guard := func(port int) (connectionPoint, connectionPoint) {
		p := g.Ports[port-1]
		inward := g.Neighbors[port-1].mul(-1 / g.Neighbors[port-1].length())
		inside := p.add(inward.mul(8))
		outside := p.sub(inward.mul(f.Width/2 + f.Border + 4))
		line(outside, inside)
		return inside, inward
	}
	if len(v.Ports) == 2 {
		a, na := guard(v.Ports[0])
		b, nb := guard(v.Ports[1])
		if (v.Ports[0]+2)%6+1 == v.Ports[1] {
			line(a, b)
		} else {
			cubic(a, a.add(na.mul(20)), b.add(nb.mul(20)), b)
		}
	} else {
		center := connectionPoint{48, 40}
		for _, port := range v.Ports {
			a, n := guard(port)
			cubic(a, a.add(n.mul(16)), center.sub(n.mul(8)), center)
		}
	}
	if f.Markings == "dashed" {
		paint := connectionRoadPaintFor(g, v)
		route.road = &paint
	}
	return route
}
func connectionRouteDistance(route connectionRoute, p connectionPoint) float64 {
	best := math.Inf(1)
	for _, s := range route.segments {
		delta := s.b.sub(s.a)
		l2 := delta[0]*delta[0] + delta[1]*delta[1]
		q := p.sub(s.a)
		t := 0.0
		if l2 > 0 {
			t = math.Max(0, math.Min(1, (q[0]*delta[0]+q[1]*delta[1])/l2))
		}
		d := q.sub(delta.mul(t))
		ds := d[0]*d[0] + d[1]*d[1]
		if ds < best {
			best = ds
		}
	}
	return math.Sqrt(best)
}
func connectionInside(g connectionGeometry, p connectionPoint) bool {
	for i, a := range g.Polygon {
		v := g.Polygon[(i+1)%len(g.Polygon)].sub(a)
		q := p.sub(a)
		cross := v[0]*q[1] - v[1]*q[0]
		// A directed edge and its reverse have complementary ownership.
		// Include top/left edges; exclude bottom/right edges, including ties.
		if cross < 0 || (cross == 0 && !(v[1] < 0 || (v[1] == 0 && v[0] > 0))) {
			return false
		}
	}
	return true
}
func connectionRamp(signed, span float64) float64 {
	// Reciprocal segments can differ by a few floating-point ulps. Stabilize
	// half-alpha rounding at mathematically identical shared-edge samples.
	signed = math.Round(signed*1e9) / 1e9
	return math.Max(0, math.Min(1, .5+signed/span))
}
func connectionBlend(a, b color.NRGBA, t float64) color.NRGBA {
	mix := func(a, b uint8) uint8 { return uint8(math.Floor(float64(a)*(1-t) + float64(b)*t + .5)) }
	return color.NRGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}
func connectionSmoothStep(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func connectionStreamProfile(g connectionGeometry, f connectionFamily, route connectionRoute, local, world connectionPoint) (float64, float64, float64) {
	edgeDistance := math.Inf(1)
	for i, a := range g.Polygon {
		edge := g.Polygon[(i+1)%len(g.Polygon)].sub(a)
		q := local.sub(a)
		edgeDistance = math.Min(edgeDistance, (edge[0]*q[1]-edge[1]*q[0])/edge.length())
	}
	envelope := connectionSmoothStep((edgeDistance - 8) / 8)
	phaseA := 2 * math.Pi * (world[0]/142 + world[1]/80)
	phaseB := 2 * math.Pi * (world[0]/142 - world[1]/80)
	warped := local.add(connectionPoint{1.2 * envelope * math.Sin(phaseA), 1.2 * envelope * math.Sin(phaseB)})
	radius := f.Width/2 + .8*envelope*math.Sin(phaseA+.4)*math.Cos(phaseB)
	coreDistance := connectionRouteDistance(route, warped) - radius
	if route.junction && envelope > 0 {
		disc := warped.sub(connectionPoint{48, 40}).length() - .68*f.Width
		// Polynomial smooth union rounds the inner banks of a confluence.
		h := math.Max(3-math.Abs(coreDistance-disc), 0) / 3
		merged := math.Min(coreDistance, disc) - h*h*3/4
		coreDistance += envelope * (merged - coreDistance)
	}
	border := math.Max(0, f.Border-.8*envelope+.3*envelope*math.Sin(phaseB+.7))
	pocket := connectionSmoothStep((math.Sin(phaseA)*math.Sin(phaseB) - .15) / .65)
	return coreDistance, coreDistance - border, 1 - .65*envelope*pocket
}

// Paths keep straight, constant-width ports but wear down their interior verges.
// The world-space phases repeat under every runtime lattice translation.
func connectionPathProfile(g connectionGeometry, f connectionFamily, route connectionRoute, local, world connectionPoint) (float64, float64, float64) {
	edgeDistance := math.Inf(1)
	for i, a := range g.Polygon {
		edge := g.Polygon[(i+1)%len(g.Polygon)].sub(a)
		q := local.sub(a)
		edgeDistance = math.Min(edgeDistance, (edge[0]*q[1]-edge[1]*q[0])/edge.length())
	}
	envelope := connectionSmoothStep((edgeDistance - 8) / 8)
	phaseA := 2 * math.Pi * (world[0]/142 + world[1]/80)
	phaseB := 2 * math.Pi * (world[0]/142 - world[1]/80)
	warped := local.add(connectionPoint{.55 * envelope * math.Sin(phaseA), .55 * envelope * math.Sin(phaseB)})
	wear := .2 * math.Sin(5*phaseA+2*phaseB+.7) * math.Sin(3*phaseB-phaseA)
	radius := f.Width/2 + envelope*(.4*math.Sin(phaseA+.4)*math.Cos(phaseB)+wear)
	coreDistance := connectionRouteDistance(route, warped) - radius
	if route.junction && envelope > 0 {
		disc := warped.sub(connectionPoint{48, 40}).length() - .65*f.Width
		h := math.Max(1.5-math.Abs(coreDistance-disc), 0) / 1.5
		merged := math.Min(coreDistance, disc) - 1.5*h*h/4
		coreDistance += envelope * (merged - coreDistance)
	}
	border := f.Border * (1 - .45*envelope + .25*envelope*math.Sin(phaseB+.7))
	pocket := connectionSmoothStep((math.Sin(2*phaseA+.3)*math.Cos(3*phaseB-.4) + .15) / .9)
	return coreDistance, coreDistance - border, 1 - .75*envelope*pocket
}

func connectionFeature(in connectionInputs, f connectionFamily, route connectionRoute, local, world connectionPoint, aa float64) color.NRGBA {
	var outer, core float64
	bankOpacity := 1.0
	switch f.ID {
	case "stream":
		coreDistance, outerDistance, opacity := connectionStreamProfile(in.manifest.Geometry, f, route, local, world)
		outer, core, bankOpacity = connectionRamp(-outerDistance, aa), connectionRamp(-coreDistance, aa), opacity
	case "path":
		coreDistance, outerDistance, opacity := connectionPathProfile(in.manifest.Geometry, f, route, local, world)
		outer, core, bankOpacity = connectionRamp(-outerDistance, aa), connectionRamp(-coreDistance, aa), opacity
	default:
		d := connectionRouteDistance(route, local)
		outer = connectionRamp(f.Width/2+f.Border-d, aa)
		core = connectionRamp(f.Width/2-d, aa)
	}
	if outer == 0 {
		return color.NRGBA{}
	}
	bank := sampleConnectionMaterial(in.materials[f.BorderMaterial], world)
	if route.road != nil {
		// The reference paved road has a narrow dark edge, rather than a dirt verge.
		bank = sampleConnectionMaterial(in.materials[f.Material], world)
		bank.R, bank.G, bank.B = unitByte(float64(bank.R)*.65), unitByte(float64(bank.G)*.65), unitByte(float64(bank.B)*.65)
	}
	if core > 0 {
		bank = connectionBlend(bank, sampleConnectionMaterial(in.materials[f.Material], world), core)
	}
	if route.road != nil && core > 0 {
		paint := connectionRoadPaintAt(*route.road, local, aa)
		bank = connectionBlend(bank, color.NRGBA{242, 240, 237, 255}, math.Min(core, paint))
	}
	alpha := outer
	if f.ID == "stream" || f.ID == "path" {
		alpha *= core + (1-core)*bankOpacity
	}
	bank.A = uint8(math.Floor(255*alpha + .5))
	return bank
}
func renderConnectionRaw(in connectionInputs, f *connectionFamily, v *connectionVariant) *image.NRGBA {
	const scale, margin = 4, 4
	im := image.NewNRGBA(image.Rect(0, 0, (96+2*margin)*scale, (80+2*margin)*scale))
	g := in.manifest.Geometry
	aa := 1 / float64(scale)
	var route connectionRoute
	var rail connectionRailGeometry
	var trench connectionTrenchGeometry
	if f != nil {
		switch f.ID {
		case "trench":
			trench = connectionTrenchRoutes(g, *f, *v)
		case "rail":
			rail = connectionRailRoutes(g, *f, *v)
		default:
			route = connectionRoutes(g, *f, *v)
		}
	}
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			p := connectionPoint{(float64(x)+.5)/scale - margin, (float64(y)+.5)/scale - margin}
			var c color.NRGBA
			if f == nil {
				c = sampleConnectionMaterial(in.materials["grass"], p)
			} else if f.ID == "trench" {
				c = connectionTrenchFeature(in, *f, trench, p, aa)
			} else if f.ID == "rail" {
				c = connectionRailFeature(in, *f, rail, p, aa)
			} else {
				c = connectionFeature(in, *f, route, p, p, aa)
			}
			im.SetNRGBA(x, y, c)
		}
	}
	return im
}

// Filter unmasked overscan before cropping. Final hard half-open ownership makes
// ordinary source-over sprite blits cover every shared pixel exactly once.
func exportConnectionTiles(in connectionInputs, f *connectionFamily, v *connectionVariant) (map[int]*image.NRGBA, error) {
	raw4 := renderConnectionRaw(in, f, v)
	raw2, err := resizeLanczos(raw4, 208, 176)
	if err != nil {
		return nil, err
	}
	outputs := map[int]*image.NRGBA{}
	for scale, raw := range map[int]*image.NRGBA{2: raw2, 4: raw4} {
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		for y := 0; y < im.Bounds().Dy(); y++ {
			for x := 0; x < im.Bounds().Dx(); x++ {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				if !connectionInside(in.manifest.Geometry, p) {
					continue
				}
				c := raw.NRGBAAt(x+4*scale, y+4*scale)
				if c.A == 0 {
					c = color.NRGBA{}
				}
				im.SetNRGBA(x, y, c)
			}
		}
		outputs[scale] = im
	}
	return outputs, nil
}
func buildConnectionScene(g connectionGeometry, f connectionFamily, v connectionVariant, parity int) (connectionScene, error) {
	scene := connectionScene{ID: fmt.Sprintf("%s-patch-parity%d", v.ID, parity), CenterLineParity: parity}
	center := connectionPoint{0, 0}
	if parity == 1 {
		center = connectionPoint{71, 40}
	}
	scene.Cells = append(scene.Cells, connectionCell{Origin: center, Variant: v.ID, Ports: append([]int(nil), v.Ports...)})
	for i, delta := range g.Neighbors {
		cell := connectionCell{Origin: center.add(delta), Ports: []int{}}
		port := i + 1
		active := false
		for _, p := range v.Ports {
			if p == port {
				active = true
			}
		}
		if active {
			opposite := (port+2)%6 + 1
			for _, candidate := range sortedConnectionVariants(f.Variants) {
				if len(candidate.Ports) != 2 {
					continue
				}
				a, b := candidate.Ports[0], candidate.Ports[1]
				if (a == port && b == opposite) || (a == opposite && b == port) {
					cell.Variant = candidate.ID
					cell.Ports = append([]int(nil), candidate.Ports...)
					break
				}
			}
			if cell.Variant == "" {
				return scene, errors.New("patch requires declared straight continuation for each active port")
			}
		}
		scene.Cells = append(scene.Cells, cell)
	}
	sort.Slice(scene.Cells, func(i, j int) bool {
		a, b := scene.Cells[i].Origin, scene.Cells[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	return scene, nil
}
func renderConnectionScene(exports map[string]map[int]*image.NRGBA, scene connectionScene, scale int) *image.NRGBA {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, c := range scene.Cells {
		minX = math.Min(minX, c.Origin[0])
		minY = math.Min(minY, c.Origin[1])
		maxX = math.Max(maxX, c.Origin[0]+96)
		maxY = math.Max(maxY, c.Origin[1]+80)
	}
	im := image.NewNRGBA(image.Rect(0, 0, int(maxX-minX)*scale, int(maxY-minY)*scale))
	// Use the decoded output PNG pixels, first terrain then infrastructure,
	// with the same integer translations and source-over operation as runtime.
	for _, layer := range []bool{false, true} {
		for _, c := range scene.Cells {
			id := "grass"
			if layer {
				id = c.Variant
				if id == "" {
					continue
				}
			}
			tile := exports[id][scale]
			position := image.Pt(int(c.Origin[0]-minX)*scale, int(c.Origin[1]-minY)*scale)
			draw.Draw(im, tile.Bounds().Add(position), tile, image.Point{}, draw.Over)
		}
	}
	return im
}
func connectionContactSheet(panels []*image.NRGBA) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 384, 160*((len(panels)+1)/2)))
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			v := uint8(80)
			if (x/8+y/8)%2 == 0 {
				v = 110
			}
			c := color.NRGBA{v, v, v, 255}
			index := (y/160)*2 + x/192
			if index < len(panels) {
				p := panels[index].NRGBAAt(x%192, y%160)
				c = connectionBlend(c, p, float64(p.A)/255)
			}
			im.SetNRGBA(x, y, c)
		}
	}
	return im
}
