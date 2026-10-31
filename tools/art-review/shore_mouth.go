package artreview

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"slices"
)

type shoreMouthReport struct {
	ID      string       `json:"id"`
	Shore   string       `json:"shore"`
	Port    int          `json:"port"`
	Outputs []packOutput `json:"outputs"`
}

func shoreMouthStream(in shoreInputs) (connectionInputs, connectionFamily) {
	c := connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{}, grassSampling: true}
	for _, name := range []string{"grass", "water", "bank"} {
		c.materials[name] = in.Materials[name].Image
	}
	f := connectionFamily{ID: "stream", Width: 18, Border: 3.5, Material: "water", BorderMaterial: "bank", BankProfile: in.Manifest.BankProfile, Season: in.Manifest.Season}
	return c, f
}

func shoreMouthEdgeDistance(p connectionPoint, edge int) float64 {
	g := sharedConnectionGeometry()
	a, b := g.Polygon[edge], g.Polygon[(edge+1)%6]
	e, q := b.sub(a), p.sub(a)
	return (e[0]*q[1] - e[1]*q[0]) / e.length()
}

// The target is derived from this tile's water mask, never from a map position
// or a particular lake silhouette. A cubic leaves the fixed stream port normal
// and reaches the closest wet interior, where its round end merges into water.
func shoreMouthRoute(v shoreVariant, port int, curve shoreCurve) (connectionRoute, error) {
	g := sharedConnectionGeometry()
	p := g.Ports[port-1]
	n := g.Neighbors[port-1].mul(-1 / g.Neighbors[port-1].length())
	a := p.add(n.mul(12))
	target, best := connectionPoint{}, math.Inf(1)
	for y := 1.; y < 80; y++ {
		for x := 1.; x < 96; x++ {
			q := connectionPoint{x, y}
			inside := true
			for edge := range 6 {
				if shoreMouthEdgeDistance(q, edge) <= 8 {
					inside = false
					break
				}
			}
			if !inside || shoreFieldWithCurve(v, q, &curve) > -1 {
				continue
			}
			d := q.sub(a).length()
			if d < best {
				target, best = q, d
			}
		}
	}
	if math.IsInf(best, 1) {
		return connectionRoute{}, fmt.Errorf("%s has no wet interior for outlet%d", v.ID, port)
	}
	r := connectionRoute{segments: []connectionSegment{{p.sub(n.mul(20)), a}}}
	h := min(12., best/2)
	b, c := a.add(n.mul(h)), target.sub(target.sub(a).mul(h/best))
	previous := a
	for i := 1; i <= 48; i++ {
		t := float64(i) / 48
		u := 1 - t
		q := a.mul(u * u * u).add(b.mul(3 * u * u * t)).add(c.mul(3 * u * t * t)).add(target.mul(t * t * t))
		r.segments = append(r.segments, connectionSegment{previous, q})
		previous = q
	}
	return r, nil
}

func shoreMouthOver(background, feature color.NRGBA) color.NRGBA {
	a := uint32(feature.A)
	mix := func(b, f uint8) uint8 { return uint8((uint32(f)*a + uint32(b)*(255-a) + 127) / 255) }
	return color.NRGBA{mix(background.R, feature.R), mix(background.G, feature.G), mix(background.B, feature.B), 255}
}

func shoreMouthColor(in shoreInputs, v shoreVariant, port int, curve shoreCurve, c connectionInputs, f connectionFamily, route connectionRoute, p connectionPoint) color.NRGBA {
	coast := shoreFieldWithCurve(v, p, &curve)
	original := shoreMaterialColor(in, p, coast)
	other := math.Inf(1)
	for edge := range 6 {
		if edge != (port % 6) {
			other = min(other, shoreMouthEdgeDistance(p, edge))
		}
	}
	if other <= 8 {
		return original
	}
	mouth := shoreMouthEdgeDistance(p, port%6)
	stream := connectionFeature(c, f, route, p, p, .25)
	joined := shoreMouthOver(sampleGrassMaterial(in.Materials["grass"].Image, p), stream)
	if mouth <= 8 {
		return joined
	}
	d, _, _ := connectionGravelStreamProfile(c.manifest.Geometry, f, route, p, p)
	union := shoreMaterialColor(in, p, min(coast, d))
	// Only the mouth changes; every other crossing retains its original guard.
	union = connectionBlend(original, union, connectionSmoothStep((other-8)/8))
	return connectionBlend(joined, union, connectionSmoothStep((mouth-8)/8))
}

func exportShoreMouth(in shoreInputs, v shoreVariant, port int) (map[int]*image.NRGBA, error) {
	canonical, err := exportShoreTiles(in, v)
	if err != nil {
		return nil, err
	}
	return exportShoreMouthOnBase(in, v, port, canonical)
}

func exportShoreMouthOnBase(in shoreInputs, v shoreVariant, port int, canonical map[int]*image.NRGBA) (map[int]*image.NRGBA, error) {
	curve := shoreCurveFor(v)
	route, err := shoreMouthRoute(v, port, curve)
	if err != nil {
		return nil, err
	}
	c, f := shoreMouthStream(in)
	raw := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			raw.SetNRGBA(x, y, shoreMouthColor(in, v, port, curve, c, f, route, p))
		}
	}
	raw2, err := resizeLanczos(raw, 208, 176)
	if err != nil {
		return nil, err
	}
	opposite := (port+2)%6 + 1
	straight := connectionVariant{Ports: []int{port, opposite}}
	stream, err := exportConnectionTiles(c, &f, &straight)
	if err != nil {
		return nil, err
	}
	grass, err := exportGrassBaseTiles(in.Materials["grass"].Image)
	if err != nil {
		return nil, err
	}
	out := map[int]*image.NRGBA{}
	for _, scale := range []int{2, 4} {
		src := raw
		if scale == 2 {
			src = raw2
		}
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		streamGround := image.NewNRGBA(im.Bounds())
		draw.Draw(streamGround, streamGround.Bounds(), grass[scale], image.Point{}, draw.Src)
		draw.Draw(streamGround, streamGround.Bounds(), stream[scale], image.Point{}, draw.Over)
		for y := range 80 * scale {
			for x := range 96 * scale {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				if !connectionInside(c.manifest.Geometry, p) {
					continue
				}
				colour := src.NRGBAAt(x+4*scale, y+4*scale)
				other := math.Inf(1)
				for edge := range 6 {
					if edge != port%6 {
						other = min(other, shoreMouthEdgeDistance(p, edge))
					}
				}
				if other <= 8 {
					colour = canonical[scale].NRGBAAt(x, y)
				} else if shoreMouthEdgeDistance(p, port%6) <= 8 {
					// Compose the independently filtered runtime layers here, too:
					// filtering a precomposite differs at antialiased banks at2x.
					colour = streamGround.NRGBAAt(x, y)
				}
				im.SetNRGBA(x, y, colour)
			}
		}
		out[scale] = im
	}
	return out, nil
}

func (b *shoreBundle) addMouths(in shoreInputs) error {
	for _, v := range shoreRoster() {
		for port := 1; port <= 6; port++ {
			if !slices.Contains(v.LandVertices, port%6) || !slices.Contains(v.LandVertices, (port+1)%6) {
				continue
			}
			m := shoreMouthReport{ID: fmt.Sprintf("%s-outlet%d", v.ID, port), Shore: v.ID, Port: port}
			canonical := map[int]*image.NRGBA{}
			for _, scale := range []int{2, 4} {
				canonical[scale] = b.pixels[shoreTile(v.ID, scale)]
			}
			tiles, err := exportShoreMouthOnBase(in, v, port, canonical)
			if err != nil {
				return err
			}
			for _, scale := range []int{2, 4} {
				o, e := b.addPNG(shoreTile(m.ID, scale), tiles[scale])
				if e != nil {
					return e
				}
				m.Outputs = append(m.Outputs, o)
			}
			b.report.Mouths = append(b.report.Mouths, m)
		}
	}
	return b.addMouthProof(in)
}

func (b *shoreBundle) addMouthProof(in shoreInputs) error {
	c, f := shoreMouthStream(in)
	for port := 1; port <= 3; port++ {
		v := connectionVariant{ID: fmt.Sprintf("stream%d%d", port, port+3), Ports: []int{port, port + 3}}
		tiles, err := exportConnectionTiles(c, &f, &v)
		if err != nil {
			return err
		}
		for _, scale := range []int{2, 4} {
			if _, err = b.addPNG(fmt.Sprintf("mouth-proof/%s-%dx.png", v.ID, scale), tiles[scale]); err != nil {
				return err
			}
		}
	}
	for _, scale := range []int{2, 4} {
		im := image.NewNRGBA(image.Rect(0, 0, 786*scale, 220*scale))
		draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{27, 34, 32, 255}), image.Point{}, draw.Src)
		for i, m := range []struct {
			id   string
			port int
		}{{"lakeside12-outlet1", 1}, {"lakeside23-outlet4", 4}, {"lakeside3456-outlet3", 3}} {
			shoreProofLabel(im, scale, 12+i*254, 6, m.id)
			n := c.manifest.Geometry.Neighbors[m.port-1]
			minX, minY := min(0., n[0]), min(0., n[1])
			at := image.Pt((12+i*254-int(minX))*scale, (24-int(minY))*scale)
			shore := b.pixels[shoreTile(m.id, scale)]
			draw.Draw(im, shore.Bounds().Add(at), shore, shore.Bounds().Min, draw.Over)
			at = at.Add(image.Pt(int(n[0])*scale, int(n[1])*scale))
			grass := b.pixels[shoreTile("grass", scale)]
			draw.Draw(im, grass.Bounds().Add(at), grass, grass.Bounds().Min, draw.Over)
			streamPort := (m.port-1)%3 + 1
			stream := b.pixels[fmt.Sprintf("mouth-proof/stream%d%d-%dx.png", streamPort, streamPort+3, scale)]
			draw.Draw(im, stream.Bounds().Add(at), stream, stream.Bounds().Min, draw.Over)
		}
		o, err := b.addPNG(fmt.Sprintf("shore-mouths-%dx.png", scale), im)
		if err != nil {
			return err
		}
		b.report.MouthProofs = append(b.report.MouthProofs, o)
	}
	return nil
}
