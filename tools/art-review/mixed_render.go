package artreview

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

type mixedProfile struct {
	Span       float64 `json:"bridge_span_world_px"`
	FootWidth  float64 `json:"footbridge_deck_width_world_px"`
	RoadWidth  float64 `json:"road_bridge_deck_width_world_px"`
	GuardOuter float64 `json:"road_bridge_support_outer_halfwidth_world_px"`
	Protected  float64 `json:"protected_port_zone_world_px"`
	Method     string  `json:"method"`
}

// mixedFootbridgeHalf is the half-width of the procedural footbridge deck over
// a path of the given width: 6 world px over the 7-pixel path, as wide as the
// path itself over a wider one. Its side curbs run 1.5 world px beyond it.
func mixedFootbridgeHalf(pathWidth float64) float64 {
	if pathWidth <= connectionPathBaseWidth {
		return 6
	}
	return pathWidth / 2
}

// mixedRailHalf is the half-width of a crossing's plank infill across the
// track: 4.5 world px for the 11-pixel rail, scaled with the rail family.
func mixedRailHalf(rail connectionFamily) float64 {
	return 4.5 * connectionRailDimsFor(rail.Width).scale
}

// mixedPathApproach names the width of the footbridge's path approaches.
func mixedPathApproach(pathWidth float64) string {
	if pathWidth == connectionPathBaseWidth {
		return "seven-pixel"
	}
	return fmt.Sprintf("%g-pixel", pathWidth)
}

// mixedProfileFor is the mixed profile with the footbridge built for the path
// family of the given width.
func mixedProfileFor(pathWidth float64) mixedProfile {
	p := fixedMixedProfile()
	if half := mixedFootbridgeHalf(pathWidth); half != 6 {
		p.FootWidth = 2 * half
		p.Method = strings.Replace(p.Method, "Footbridge 48x12 timber deck, side curbs at transverse 6..7.5, path-width 7 approaches retained", fmt.Sprintf("Footbridge 48x%g timber deck, side curbs at transverse %g..%g, path-width %g approaches retained", 2*half, half, half+1.5, pathWidth), 1)
	}
	return p
}

func fixedMixedProfile() mixedProfile {
	return mixedProfile{48, 12, 22, 12.5, 8, "Summer materials provisional. Footbridge 48x12 timber deck, side curbs at transverse 6..7.5, path-width 7 approaches retained; earth abutment ramps extend 4 world px beyond deck. Road bridge 48x22 timber deck with side beams at transverse 10..12.5, ramps 4 px; finite structure only, independent upper route required. Round1 wood repair: authored timber grain averaged at world offsets (0,0),(+/-1,0),(0,+/-1); RGB tone=0.35*source+(112,83,47)+8*sin(board_index*2.399963229728653). Cross-planks pitch3 world px, seam0.3 at shade0.48 and bevel0.35 at shade1.18. Side beams inner30% highlighted1.3, outer22% shaded0.48, middle0.72; seven paired posts pitch8 with 1.6-world-pixel along width and1.4 caps/shaded side. Transverse end beams occupy final1.5px of each span end; rounded deck ends radius0.7. All spans/widths/contacts unchanged. Crossings retain approach then rail, replace only interior intersection ballast/ties with flush fitted pitch3 timber, rounded intersection corners radius0.6, then restore continuous steel above infill. Infill excluded within 12 world px of polygon edges (8 px protected plus filter support). Raw source-over at4x with 4-world-pixel overscan, alpha-aware Lanczos3 to2x, crop then complementary final ownership; no hard masks before filtering."}
}

func mixedOver(a, b color.NRGBA) color.NRGBA {
	// Straight-alpha source over, including transparent approach verges.
	alpha := float64(b.A) / 255
	under := float64(a.A) / 255 * (1 - alpha)
	out := alpha + under
	if out == 0 {
		return color.NRGBA{}
	}
	ch := func(x, y uint8) uint8 { return uint8(math.Floor((float64(y)*alpha+float64(x)*under)/out + .5)) }
	return color.NRGBA{ch(a.R, b.R), ch(a.G, b.G), ch(a.B, b.B), uint8(math.Floor(out*255 + .5))}
}

func mixedEdgeDistance(g connectionGeometry, p connectionPoint) float64 {
	d := math.Inf(1)
	for i, a := range g.Polygon {
		b := g.Polygon[(i+1)%len(g.Polygon)]
		e := b.sub(a)
		q := p.sub(a)
		d = math.Min(d, (e[0]*q[1]-e[1]*q[0])/e.length())
	}
	return d
}

func mixedBoxSDF(x, y, hx, hy, r float64) float64 {
	a, b := math.Abs(x)-hx+r, math.Abs(y)-hy+r
	return math.Hypot(math.Max(a, 0), math.Max(b, 0)) + math.Min(math.Max(a, b), 0) - r
}

// mixedWood keeps the authored timber grain but suppresses fine speckle before
// lifting its dark source values. Board tone and bevels are geometry, not new
// source pixels; world-phase material sampling is unchanged.
func mixedWood(src *image.NRGBA, p connectionPoint, along float64) color.NRGBA {
	var channels [3]float64
	for _, delta := range []connectionPoint{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		c := sampleConnectionMaterial(src, p.add(delta))
		channels[0] += float64(c.R) / 5
		channels[1] += float64(c.G) / 5
		channels[2] += float64(c.B) / 5
	}
	board := math.Floor((along + 24) / 3)
	variation := 8 * math.Sin(board*2.399963229728653)
	phase := along + 24 - 3*board
	shade := 1.
	if phase < .3 {
		shade = .48
	} else if phase < .65 {
		shade = 1.18
	}
	channel := func(value, offset float64) uint8 {
		return uint8(math.Floor(math.Max(0, math.Min(255, (value*.35+offset+variation)*shade)) + .5))
	}
	return color.NRGBA{channel(channels[0], 112), channel(channels[1], 83), channel(channels[2], 47), 255}
}

func renderMixedRaw(in mixedInputs, v mixedVariant) *image.NRGBA {
	g := in.merged.manifest.Geometry
	var base *image.NRGBA
	var rail connectionRailGeometry
	var roadRoute connectionRoute
	var roadFamily connectionFamily
	var railHalf float64
	footHalf := 6.
	switch v.Kind {
	case "crossing":
		road, track := in.components[v.Layers[0]], in.components[v.Layers[1]]
		base = renderConnectionRaw(road.input, &road.family, &road.variant)
		top := renderConnectionRaw(track.input, &track.family, &track.variant)
		for y := 0; y < base.Bounds().Dy(); y++ {
			for x := 0; x < base.Bounds().Dx(); x++ {
				base.SetNRGBA(x, y, mixedOver(base.NRGBAAt(x, y), top.NRGBAAt(x, y)))
			}
		}
		rail = connectionRailRoutes(g, track.family, track.variant)
		roadRoute = connectionRoutes(g, road.family, road.variant)
		roadFamily = road.family
		railHalf = mixedRailHalf(track.family)
	case "footbridge":
		id := "path" + string(rune('0'+v.Ports[0])) + string(rune('0'+v.Ports[1]))
		c := in.components[id]
		base = renderConnectionRaw(c.input, &c.family, &c.variant)
		footHalf = mixedFootbridgeHalf(c.family.Width)
	default:
		base = image.NewNRGBA(image.Rect(0, 0, 416, 352))
	}
	if in.manifest.BridgeArtwork == "standing" && v.Kind != "crossing" {
		if v.Kind == "footbridge" {
			axis := connectionRailUnit(g.Neighbors[v.Ports[0]-1])
			// Keep the original path outside the painted deck. Removing the span
			// before density filtering preserves approaches without old timber,
			// rails or earth abutments beneath the separate standing painting.
			for y := range base.Bounds().Dy() {
				for x := range base.Bounds().Dx() {
					q := connectionPoint{(float64(x)+.5)/4 - 52, (float64(y)+.5)/4 - 44}
					along := q[0]*axis[0] + q[1]*axis[1]
					c := base.NRGBAAt(x, y)
					c.A = uint8(math.Floor(float64(c.A)*connectionRamp(math.Abs(along)-24, .25) + .5))
					if c.A == 0 {
						c = color.NRGBA{}
					}
					base.SetNRGBA(x, y, c)
				}
			}
		}
		return base
	}
	for y := 0; y < 352; y++ {
		for x := 0; x < 416; x++ {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			c := base.NRGBAAt(x, y)
			if v.Kind == "crossing" {
				// Flush infill is limited to the intersection, leaving the complete rail
				// (gauge, sleeper phase and ballast) unchanged near every boundary.
				dRoad := connectionRouteDistance(roadRoute, p)
				dRail := connectionRouteDistance(rail.ballast, p)
				if mixedEdgeDistance(g, p) > 12 && dRoad < roadFamily.Width/2+1+.25 && dRail < railHalf+.25 {
					track := in.components[v.Layers[1]]
					tangent := connectionRailUnit(g.Neighbors[track.variant.Ports[0]-1])
					q := p.sub(connectionPoint{48, 40})
					board := q[0]*tangent[0] + q[1]*tangent[1]
					infill := mixedWood(in.merged.materials["timber"], p, board)
					if track.family.Verge == "snow" {
						// Winter crossings: the planks lie under the same
						// packed snow as the way crossing the track.
						way := in.components[v.Layers[0]]
						infill = connectionBlend(infill, sampleConnectionMaterial(way.input.materials[way.family.Material], p), mixedWinterInfillSnow)
						infill = connectionBlend(infill, connectionRutColor, mixedWinterInfillSlush)
					}
					// Rounded approach/track intersection produces fitted flush
					// boards without bridge curbs, posts, or raised shadows.
					rx, ry := dRoad-roadFamily.Width/2-1+.6, dRail-railHalf+.6
					rounded := math.Hypot(math.Max(rx, 0), math.Max(ry, 0)) + math.Min(math.Max(rx, ry), 0) - .6
					a := connectionRamp(-rounded, .25)
					infill.A = uint8(math.Floor(255*a + .5))
					c = mixedOver(c, infill)
					steel := connectionRailSteel(track.input, track.family, p)
					steel.A = uint8(math.Floor(255*connectionRamp(rail.dims.steel/2-connectionRouteDistance(rail.rails, p), .25) + .5))
					c = mixedOver(c, steel)
				}
			} else {
				t := connectionRailUnit(g.Neighbors[v.Ports[0]-1])
				n := connectionRailNormal(t)
				q := p.sub(connectionPoint{48, 40})
				along, across := q[0]*t[0]+q[1]*t[1], q[0]*n[0]+q[1]*n[1]
				half := 11.
				if v.Kind == "footbridge" {
					half = footHalf
				}
				// Abutments use opaque earth only within the finite span ends.
				ramp := connectionRamp(-mixedBoxSDF(along, across, 28, half-.5, 1.5), .25)
				if math.Abs(along) > 23 {
					earth := sampleConnectionMaterial(in.merged.materials["earth"], p)
					earth.A = uint8(math.Floor(255*ramp + .5))
					c = mixedOver(c, earth)
				}
				deck := connectionRamp(-mixedBoxSDF(along, across, 24, half, .7), .25)
				if deck > 0 {
					timber := mixedWood(in.merged.materials["timber"], p, along)
					// Full-width transverse end beams visibly land on earth ramps.
					if math.Abs(along) > 22.5 {
						timber = connectionTrenchShade(timber, .72)
					}
					timber.A = uint8(math.Floor(255*deck + .5))
					c = mixedOver(c, timber)
				}
				inner, outer := 10., 12.5
				if v.Kind == "footbridge" {
					inner, outer = footHalf, footHalf+1.5
				}
				beam := connectionRamp(24-math.Abs(along), .25) * connectionRamp(math.Abs(across)-inner, .25) * connectionRamp(outer-math.Abs(across), .25)
				if beam > 0 {
					wood := mixedWood(in.merged.materials["timber"], p, 4.5)
					// Light inner arris, darker outer face: depth remains legible
					// at2x while the complete structure stays within its footprint.
					fraction := (math.Abs(across) - inner) / (outer - inner)
					shade := .72
					if fraction < .3 {
						shade = 1.3
					} else if fraction > .78 {
						shade = .48
					}
					wood = connectionTrenchShade(wood, shade)
					// Seven paired square posts, pitch8, with bright top caps and
					// a narrow shaded side. They sit on the finite side beams.
					nearest := math.Round(along/8) * 8
					postAlong := along - nearest
					postAcross := math.Abs(across) - (inner+outer)/2
					post := connectionRamp(.8-math.Abs(postAlong), .25) * connectionRamp((outer-inner)/2-math.Abs(postAcross), .25)
					if math.Abs(nearest) <= 24 && post > 0 {
						cap := mixedWood(in.merged.materials["timber"], p, 4.5)
						cap = connectionTrenchShade(cap, 1.4)
						if postAlong > .45 {
							cap = connectionTrenchShade(cap, .5)
						}
						wood = connectionBlend(wood, cap, post)
					}
					wood.A = uint8(math.Floor(255*beam + .5))
					c = mixedOver(c, wood)
				}
			}
			base.SetNRGBA(x, y, c)
		}
	}
	return base
}

func exportMixedTiles(in mixedInputs, v mixedVariant) (map[int]*image.NRGBA, error) {
	raw4 := renderMixedRaw(in, v)
	raw2, err := resizeLanczos(raw4, 208, 176)
	if err != nil {
		return nil, err
	}
	out := map[int]*image.NRGBA{}
	for scale, raw := range map[int]*image.NRGBA{2: raw2, 4: raw4} {
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		for y := 0; y < 80*scale; y++ {
			for x := 0; x < 96*scale; x++ {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				if connectionInside(in.merged.manifest.Geometry, p) {
					c := raw.NRGBAAt(x+4*scale, y+4*scale)
					if c.A == 0 {
						c = color.NRGBA{}
					}
					im.SetNRGBA(x, y, c)
				}
			}
		}
		out[scale] = im
	}
	return out, nil
}
