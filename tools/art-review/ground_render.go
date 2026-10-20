package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

// The six entries always mean NE, SE, S, SW, NW, N. Runtime IDs remain stable;
// the context key identifies a rendering profile, never a new runtime terrain ID.
type groundContext struct {
	RuntimeID   string
	NeighborIDs [6]string
}

func groundSeason(id string) (season, base string) {
	switch id {
	case "grass", "mud", "stone":
		return "summer", "grass"
	case "grass_w", "rough_w", "stoney_w", "shelled_w":
		return "winter", "grass_w"
	default:
		return "", ""
	}
}
func newGroundContext(centerRuntimeID string, neighborIDs []string) (groundContext, error) {
	c := groundContext{RuntimeID: centerRuntimeID}
	season, _ := groundSeason(centerRuntimeID)
	if season == "" || len(neighborIDs) != 6 {
		return c, errors.New("ground context requires a known center and exactly six NE,SE,S,SW,NW,N neighbors")
	}
	for i, id := range neighborIDs {
		other, _ := groundSeason(id)
		if other == "" || other != season {
			return c, fmt.Errorf("ground neighbor %d has unknown or incompatible seasonal ID %q", i+1, id)
		}
		c.NeighborIDs[i] = id
	}
	return c, nil
}
func (c groundContext) effective() groundContext {
	_, base := groundSeason(c.RuntimeID)
	if c.RuntimeID == base {
		for i := range c.NeighborIDs {
			c.NeighborIDs[i] = base
		}
	}
	return c
}
func (c groundContext) descriptor() string {
	c = c.effective()
	season, _ := groundSeason(c.RuntimeID)
	var b strings.Builder
	fmt.Fprintf(&b, "ground-v1\nseason=%s\ncenter=%s\n", season, c.RuntimeID)
	for i, label := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
		fmt.Fprintf(&b, "%s=%s\n", label, c.NeighborIDs[i])
	}
	return b.String()
}
func (c groundContext) key() string {
	sum := sha256.Sum256([]byte(c.descriptor()))
	return hex.EncodeToString(sum[:])
}

// Closed filled-polygon distance is zero throughout the interior. Ownership is
// a separate half-open operation applied only after overscan filtering.
func groundFilledDistance(p connectionPoint) float64 {
	polygon := sharedConnectionGeometry().Polygon
	inside := true
	best := math.Inf(1)
	for i, a := range polygon {
		edge := polygon[(i+1)%len(polygon)].sub(a)
		q := p.sub(a)
		if edge[0]*q[1]-edge[1]*q[0] < 0 {
			inside = false
		}
		t := max(0, min(1, (q[0]*edge[0]+q[1]*edge[1])/(edge[0]*edge[0]+edge[1]*edge[1])))
		distance := q.sub(edge.mul(t))
		best = min(best, distance[0]*distance[0]+distance[1]*distance[1])
	}
	if inside {
		return 0
	}
	return math.Sqrt(best)
}
func groundQuinticEnvelope(distance float64) float64 {
	t := max(0, min(1, (distance-4)/8))
	return t * t * t * (10 + t*(-15+6*t))
}

type groundGeometrySample struct {
	weights   [7]float64
	envelopes [7]float64
}

func groundGeometryAt(p connectionPoint) groundGeometrySample {
	var s groundGeometrySample
	origins := [7]connectionPoint{}
	copy(origins[1:], sharedConnectionGeometry().Neighbors)
	for i, o := range origins {
		local := p.sub(o)
		q := local.sub(connectionPoint{48, 40})
		t := max(0, 1-(q[0]*q[0]+q[1]*q[1])/6400)
		s.weights[i] = t * t * t
		s.envelopes[i] = groundQuinticEnvelope(groundFilledDistance(local))
	}
	return s
}
func groundMaterialIndex(id string) int {
	for i, name := range groundMaterialIDs {
		if id == name {
			return i
		}
	}
	panic("unvalidated ground material ID")
}
func groundMix(c groundContext, g groundGeometrySample, samples [7]color.NRGBA) [3]float64 {
	_, base := groundSeason(c.RuntimeID)
	bc := samples[groundMaterialIndex(base)]
	out := [3]float64{float64(bc.R), float64(bc.G), float64(bc.B)}
	if c.RuntimeID == base {
		return out
	}
	ids := [7]string{c.RuntimeID}
	copy(ids[1:], c.NeighborIDs[:])
	var sum [3]float64
	total, envelope := 0., 1.
	for i, id := range ids {
		if id == base {
			envelope *= g.envelopes[i]
			continue
		}
		w := g.weights[i]
		color := samples[groundMaterialIndex(id)]
		total += w
		sum[0] += w * float64(color.R)
		sum[1] += w * float64(color.G)
		sum[2] += w * float64(color.B)
	}
	if total > 0 {
		for i := range out {
			out[i] = out[i]*(1-envelope) + sum[i]/total*envelope
		}
	}
	return out
}

// groundContextRGB is a pure context field. Only P expanded by four world
// pixels is the promised shared field: omitted sites cannot contribute there.
// Each scalar quintic is C2; polygon distance makes the spatial envelope C1 and
// piecewise C2. The existing mirrored byte texture sampler has its own rounding.
func groundContextRGB(c groundContext, p connectionPoint, materials map[string]*image.NRGBA) [3]float64 {
	var samples [7]color.NRGBA
	for i, id := range groundMaterialIDs {
		if id == "grass" {
			samples[i] = sampleGrassMaterial(materials[id], p)
		} else {
			samples[i] = sampleConnectionMaterial(materials[id], p)
		}
	}
	return groundMix(c, groundGeometryAt(p), samples)
}
func groundRounded(rgb [3]float64) color.NRGBA {
	round := func(v float64) uint8 { return uint8(math.Floor(max(0, min(255, v)) + .5)) }
	return color.NRGBA{round(rgb[0]), round(rgb[1]), round(rgb[2]), 255}
}

type groundRawSample struct {
	geometry  groundGeometrySample
	materials [7]color.NRGBA
}
type groundRenderer struct {
	materials  map[string]*image.NRGBA
	samples    []groundRawSample
	mudFeature string
}

func newGroundRenderer(materials map[string]*image.NRGBA) *groundRenderer {
	r := &groundRenderer{materials: materials, samples: make([]groundRawSample, 416*352)}
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			s := &r.samples[y*416+x]
			s.geometry = groundGeometryAt(p)
			for i, id := range groundMaterialIDs {
				if id == "grass" {
					s.materials[i] = sampleGrassMaterial(materials[id], p)
				} else {
					s.materials[i] = sampleConnectionMaterial(materials[id], p)
				}
			}
		}
	}
	return r
}
func (r *groundRenderer) tiles(c groundContext) (map[int]*image.NRGBA, error) {
	_, base := groundSeason(c.RuntimeID)
	if c.RuntimeID == base {
		if base == "grass" {
			return exportGrassBaseTiles(r.materials[base])
		}
		// Preserve both existing base pixels and the resulting encoded PNG bytes.
		return exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": r.materials[base]}}, nil, nil)
	}
	raw4 := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	for y := range 352 {
		for x := range 416 {
			s := r.samples[y*416+x]
			if r.mudFeature == "circular-earthwork" {
				// Mud sites become grass outside their compact feature. Keep the
				// same continuous context field for neighboring stone and grass.
				s.materials[groundMaterialIndex("mud")] = s.materials[groundMaterialIndex("grass")]
			}
			paint := groundRounded(groundMix(c, s.geometry, s.materials))
			if c.RuntimeID == "mud" && r.mudFeature == "circular-earthwork" {
				p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
				paint = groundEarthwork(r.materials["mud"], p, paint)
			}
			raw4.SetNRGBA(x, y, paint)
		}
	}
	raw2, err := resizeLanczos(raw4, 208, 176)
	if err != nil {
		return nil, err
	}
	tiles := map[int]*image.NRGBA{}
	for scale, raw := range map[int]*image.NRGBA{2: raw2, 4: raw4} {
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		for y := range 80 * scale {
			for x := range 96 * scale {
				if connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}) {
					im.SetNRGBA(x, y, raw.NRGBAAt(x+4*scale, y+4*scale))
				}
			}
		}
		tiles[scale] = im
	}
	return tiles, nil
}

// The full source is one authored feature, never mirrored or averaged. Its
// elliptical footprint lies inside the hex with grass on the complete filter
// guard. Thus a neighboring tile cannot clip the ring or receive half a crater.
func groundEarthwork(source *image.NRGBA, p connectionPoint, base color.NRGBA) color.NRGBA {
	q := connectionPoint{(p[0] - 48) / 34, (p[1] - 40) / 28}
	alpha := 1 - connectionSmoothStep((q.length()-.88)/.12)
	if alpha <= 0 {
		return base
	}
	x := (q[0] + 1) / 2 * float64(source.Bounds().Dx()-1)
	y := (q[1] + 1) / 2 * float64(source.Bounds().Dy()-1)
	ix, iy := int(x), int(y)
	fx, fy := x-float64(ix), y-float64(iy)
	var rgb [3]float64
	for dy := range 2 {
		for dx := range 2 {
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
			c := source.NRGBAAt(ix+dx, iy+dy)
			rgb[0] += weight * float64(c.R)
			rgb[1] += weight * float64(c.G)
			rgb[2] += weight * float64(c.B)
		}
	}
	return connectionBlend(base, groundRounded(rgb), alpha)
}
