package main

import (
	"image"
	"image/color"
	"math"
)

const hillsProfileText = "hills-v1: bounded summer hills/grass exporter; context keys are full lowercase SHA256 of hills-v1\\nseason=summer\\ncenter=<id>\\nphase=<x>,<y>\\nNE=<id>\\nSE=<id>\\nS=<id>\\nSW=<id>\\nNW=<id>\\nN=<id>\\n with literal LF after every line. Normalize actual integer lattice origins modulo (568,320); grass normalizes phase and neighbors. Global height-derived relief normal with slope attenuation near grass: scalar H=32*Rec709(bilinear source at cosine world coordinates)/255, offsets (113,79), periods (568,320), central derivative h=2. Product of per-grass filled-polygon distance quintic gates on [7,19] attenuates slopes AFTER differentiation; no derivative of the grass gate. Only isolated hills (all six grass neighbors) add a compact rotated cubic ellipse height cue before differentiation, centered (48,40), radii (28,22); phase-derived angle and amplitude fade against nine-sample unattenuated global shade RMS. Connected and mixed hill contexts have no cue. Fixed world light (-.55,-.4,.75), ambient .35/diffuse .65, flat-normal normalization, shade clamp [.55,1.25]; unmirrored lattice-periodic sampleGrassMaterial grass sampling and half-up RGB rounding. No authored height bytes change or per-pixel art rotation. Whole opaque 416x352 4x overscan -> whole unmasked Lanczos3 208x176 2x -> crop four world pixels -> independent complementary half-open ownership; no pre-mask. Final four-world-pixel grass quiet band is byte exact; maximum filter displacement 1.375 per axis (1.944543648 radial), 4+1.944543648<7. Omitted grass polygons have distance >=42 on P expanded by four, greater than 19; global gradient is independent of context. Isolated cue radius+derivative reach+filter support <=28+2+1.944543648<40, so it cannot reach another cell's owned filter domain. No global C2 claim: polygon distance and bilinear luminance have their own regularity. Export all64 phase0 masks, all32 isolated phases and actual fixed eight fixture contexts; other phases/masks and arbitrary-map runtime selection remain OPEN." + " " + grassSamplingProfile

type hillsParameters struct {
	HeightAmplitude       float64            `json:"height_amplitude"`
	HeightPeriod          connectionPoint    `json:"height_period"`
	HeightOffset          connectionPoint    `json:"height_offset"`
	Luminance             [3]float64         `json:"luminance_coefficients"`
	DerivativeStep        float64            `json:"derivative_step"`
	GateStart             float64            `json:"gate_start"`
	GateEnd               float64            `json:"gate_end"`
	Light                 [3]float64         `json:"light"`
	Ambient               float64            `json:"ambient"`
	Diffuse               float64            `json:"diffuse"`
	ShadeClamp            [2]float64         `json:"shade_clamp"`
	OverscanWorld         float64            `json:"overscan_world"`
	FinalQuietBand        float64            `json:"final_quiet_band"`
	FilterMaxDisplacement connectionPoint    `json:"filter_max_displacement"`
	IsolatedCue           hillsCueParameters `json:"isolated_cue"`
}
type hillsCueParameters struct {
	Center       connectionPoint   `json:"center"`
	Radii        connectionPoint   `json:"radii"`
	MaxAmplitude float64           `json:"max_amplitude"`
	FadeRMS      [2]float64        `json:"fade_rms"`
	ThetaStep    float64           `json:"theta_step"`
	RMSOffsets   []connectionPoint `json:"rms_offsets"`
}

func hillsNumericParameters() hillsParameters {
	return hillsParameters{HeightAmplitude: 32, HeightPeriod: connectionPoint{568, 320}, HeightOffset: connectionPoint{113, 79}, Luminance: [3]float64{.2126, .7152, .0722}, DerivativeStep: 2, GateStart: 7, GateEnd: 19, Light: [3]float64{-.55, -.4, .75}, Ambient: .35, Diffuse: .65, ShadeClamp: [2]float64{.55, 1.25}, OverscanWorld: 4, FinalQuietBand: 4, FilterMaxDisplacement: connectionPoint{1.375, 1.375}, IsolatedCue: hillsCueParameters{Center: connectionPoint{48, 40}, Radii: connectionPoint{28, 22}, MaxAmplitude: 6, FadeRMS: [2]float64{.08, .2}, ThetaStep: math.Pi * (3 - math.Sqrt(5)), RMSOffsets: hillsRMSOffsets()}}
}
func hillsRMSOffsets() []connectionPoint {
	return []connectionPoint{{0, 0}, {12, 0}, {-12, 0}, {0, 12}, {0, -12}, {8, 8}, {8, -8}, {-8, 8}, {-8, -8}}
}
func hillsQuintic(t float64) float64 { t = max(0, min(1, t)); return t * t * t * (10 + t*(-15+6*t)) }

// The height is global and unlit. Neither grass gates nor terrain context enter this scalar.
func hillsHeightAt(src *image.NRGBA, p connectionPoint) float64 {
	u := .5 - .5*math.Cos(2*math.Pi*(p[0]+113)/568)
	v := .5 - .5*math.Cos(2*math.Pi*(p[1]+79)/320)
	x, y := u*float64(src.Bounds().Dx()-1), v*float64(src.Bounds().Dy()-1)
	ix, iy := int(x), int(y)
	fx, fy := x-float64(ix), y-float64(iy)
	lum := 0.
	for dy := range 2 {
		for dx := range 2 {
			wx, wy := 1-fx, 1-fy
			if dx == 1 {
				wx = fx
			}
			if dy == 1 {
				wy = fy
			}
			c := src.NRGBAAt(min(ix+dx, src.Bounds().Dx()-1), min(iy+dy, src.Bounds().Dy()-1))
			lum += (.2126*float64(c.R) + .7152*float64(c.G) + .0722*float64(c.B)) * wx * wy
		}
	}
	return 32 * lum / 255
}
func hillsGradientAt(src *image.NRGBA, p connectionPoint) connectionPoint {
	return connectionPoint{(hillsHeightAt(src, p.add(connectionPoint{2, 0})) - hillsHeightAt(src, p.add(connectionPoint{-2, 0}))) / 4, (hillsHeightAt(src, p.add(connectionPoint{0, 2})) - hillsHeightAt(src, p.add(connectionPoint{0, -2}))) / 4}
}
func hillsShade(g connectionPoint) float64 {
	if g == (connectionPoint{}) {
		return 1
	}
	lightLen := math.Sqrt(.55*.55 + .4*.4 + .75*.75)
	lx, ly, lz := -.55/lightLen, -.4/lightLen, .75/lightLen
	diffuse := max(0, (-g[0]*lx-g[1]*ly+lz)/math.Sqrt(g[0]*g[0]+g[1]*g[1]+1))
	return max(.55, min(1.25, (.35+.65*diffuse)/(.35+.65*lz)))
}

type hillsCue struct {
	Enabled bool    `json:"enabled"`
	A       float64 `json:"A"`
	Theta   float64 `json:"theta"`
	S       float64 `json:"S"`
}

func hillsIsolated(c hillsContext) bool {
	if c.RuntimeID != "hills" {
		return false
	}
	for _, id := range c.NeighborIDs {
		if id != "grass" {
			return false
		}
	}
	return true
}
func hillsCueFor(src *image.NRGBA, c hillsContext) hillsCue {
	if !hillsIsolated(c) {
		return hillsCue{}
	}
	sum := 0.
	for _, d := range hillsRMSOffsets() {
		g := hillsGradientAt(src, c.Phase.add(connectionPoint{48, 40}).add(d))
		s := hillsShade(g) - 1
		sum += s * s
	}
	s := math.Sqrt(sum / 9)
	line := int(c.Phase[1]) / 40
	col := (int(c.Phase[0])/71 - line%2) / 2
	index := 4*line + col
	return hillsCue{Enabled: true, A: 6 * (1 - hillsQuintic((s-.08)/.12)), Theta: float64(index) * math.Pi * (3 - math.Sqrt(5)), S: s}
}
func hillsCueHeight(c hillsCue, p connectionPoint) float64 {
	if !c.Enabled || c.A == 0 {
		return 0
	}
	sin, cos := math.Sincos(c.Theta)
	return hillsCueHeightRotated(c.A, sin, cos, p)
}
func hillsCueHeightRotated(a, sin, cos float64, p connectionPoint) float64 {
	dx, dy := p[0]-48, p[1]-40
	rx, ry := cos*dx+sin*dy, -sin*dx+cos*dy
	q := max(0, 1-(rx/28)*(rx/28)-(ry/22)*(ry/22))
	return a * q * q * q
}

type hillsRawSample struct {
	grass color.NRGBA
	gates [7]float64
}
type hillsRenderer struct {
	grass, height *image.NRGBA
	scrub         *image.NRGBA
	samples       []hillsRawSample
	// At most the 32 validated lattice phases; no cache of all 32*64 rendered contexts.
	normals map[connectionPoint][]connectionPoint
	cues    map[connectionPoint]hillsCue
}

func newHillsRenderer(grass, height *image.NRGBA) *hillsRenderer {
	r := &hillsRenderer{grass: grass, height: height, samples: make([]hillsRawSample, 416*352), normals: map[connectionPoint][]connectionPoint{}, cues: map[connectionPoint]hillsCue{}}
	origins := [7]connectionPoint{}
	copy(origins[1:], sharedConnectionGeometry().Neighbors)
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			s := &r.samples[y*416+x]
			s.grass = sampleGrassMaterial(grass, p)
			for i, o := range origins {
				s.gates[i] = hillsQuintic((groundFilledDistance(p.sub(o)) - 7) / 12)
			}
		}
	}
	return r
}
func (r *hillsRenderer) cueFor(c hillsContext) hillsCue {
	if r.scrub != nil {
		return hillsCue{}
	}
	if !hillsIsolated(c) {
		return hillsCue{}
	}
	if cue, ok := r.cues[c.Phase]; ok {
		return cue
	}
	cue := hillsCueFor(r.height, c)
	r.cues[c.Phase] = cue
	return cue
}
func (r *hillsRenderer) normalsFor(phase connectionPoint) []connectionPoint {
	if normals, ok := r.normals[phase]; ok {
		return normals
	}
	normals := make([]connectionPoint, 416*352)
	for y := range 352 {
		for x := range 416 {
			p := phase.add(connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4})
			normals[y*416+x] = hillsGradientAt(r.height, p)
		}
	}
	r.normals[phase] = normals
	return normals
}
func (r *hillsRenderer) raw(c hillsContext) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	var normals []connectionPoint
	if r.scrub == nil {
		normals = r.normalsFor(c.Phase)
	}
	cue := r.cueFor(c)
	sin, cos := math.Sincos(cue.Theta)
	ids := [7]string{c.RuntimeID}
	copy(ids[1:], c.NeighborIDs[:])
	for y := range 352 {
		for x := range 416 {
			i := y*416 + x
			s := r.samples[i]
			gate := 1.
			for j, id := range ids {
				if id == "grass" {
					gate *= s.gates[j]
				}
			}
			pixel := s.grass
			if r.scrub != nil {
				if gate != 0 {
					p := c.Phase.add(connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4})
					pixel = connectionBlend(pixel, sampleGrassMaterial(r.scrub, p), gate)
				}
				im.SetNRGBA(x, y, pixel)
				continue
			}
			if gate != 0 {
				g := normals[i]
				if cue.Enabled && cue.A != 0 {
					p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
					g[0] += (hillsCueHeightRotated(cue.A, sin, cos, p.add(connectionPoint{2, 0})) - hillsCueHeightRotated(cue.A, sin, cos, p.add(connectionPoint{-2, 0}))) / 4
					g[1] += (hillsCueHeightRotated(cue.A, sin, cos, p.add(connectionPoint{0, 2})) - hillsCueHeightRotated(cue.A, sin, cos, p.add(connectionPoint{0, -2}))) / 4
				}
				if g != (connectionPoint{}) {
					shade := hillsShade(g.mul(gate))
					pixel = groundRounded([3]float64{float64(pixel.R) * shade, float64(pixel.G) * shade, float64(pixel.B) * shade})
				}
			}
			im.SetNRGBA(x, y, pixel)
		}
	}
	return im
}
func (r *hillsRenderer) tiles(c hillsContext) (map[int]*image.NRGBA, error) {
	validated, err := newHillsContext(c.RuntimeID, c.NeighborIDs[:], c.Phase)
	if err != nil {
		return nil, err
	}
	c = validated.effective()
	if c.RuntimeID == "grass" {
		return exportGrassBaseTiles(r.grass)
	}
	raw4 := r.raw(c)
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
