package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

type compoundTile struct {
	Cell      string          `json:"cell"`
	RuntimeID string          `json:"runtime_id"`
	Origin    connectionPoint `json:"origin"`
	Outputs   []packOutput    `json:"outputs"`
}
type compoundLayout struct {
	Layout    string         `json:"layout"`
	Season    string         `json:"season"`
	Width     int            `json:"width"`
	Height    int            `json:"height"`
	EntryCell string         `json:"entry_cell"`
	Tiles     []compoundTile `json:"tiles"`
}

func compoundLayoutFor(id string) (compoundLayout, error) {
	l := compoundLayout{Width: 238, Height: 160, EntryCell: "N"}
	switch id {
	case "depot-summer", "depot-winter":
		l.Layout = "depot"
	case "ruin-summer", "ruin-winter":
		l.Layout = "ruin"
	case "hq-summer", "hq-winter":
		l.Layout = "hq"
		l.Height = 240
	case "factory-summer", "factory-winter":
		l.Layout = "factory"
		l.EntryCell = "E"
	case "mountain-summer", "mountain-winter":
		l.Layout = "mountain"
		l.EntryCell = ""
	case "mountain-single-summer", "mountain-single-winter":
		l.Layout = "mountain-single"
		l.Width, l.Height, l.EntryCell = 96, 80, ""
	case "hamlet-summer", "hamlet-winter", "hamlet2-summer", "hamlet2-winter", "hamlet3-summer", "hamlet3-winter",
		"hamlet4-summer", "hamlet4-winter", "hamlet5-summer", "hamlet5-winter", "hamlet6-summer", "hamlet6-winter",
		"hamlet7-summer", "hamlet7-winter", "hamlet8-summer", "hamlet8-winter":
		// hamletN is the house variant houseN (mapcompose.MaxHouseVariants).
		l.Layout = "hamlet"
		l.Width, l.Height, l.EntryCell = 96, 80, ""
	case "water-stones":
		l.Layout, l.Season = "water-stones", "shared"
		l.Width, l.Height, l.EntryCell = 96, 80, ""
		l.Tiles = []compoundTile{{Cell: "C", RuntimeID: "waterStones", Origin: connectionPoint{0, 0}}}
		return l, nil
	default:
		return l, fmt.Errorf("unsupported compound scene %q", id)
	}
	l.Season = "summer"
	suffix := ""
	if strings.HasSuffix(id, "-winter") {
		l.Season = "winter"
		suffix = "_w"
	}
	if l.Layout == "mountain-single" {
		l.Tiles = []compoundTile{{Cell: "C", RuntimeID: "mountain" + suffix, Origin: connectionPoint{0, 0}}}
		return l, nil
	}
	if l.Layout == "hamlet" {
		variant := strings.TrimPrefix(strings.TrimSuffix(strings.TrimSuffix(id, "-summer"), "-winter"), "hamlet")
		l.Tiles = []compoundTile{{Cell: "C", RuntimeID: "house" + variant + suffix, Origin: connectionPoint{0, 0}}}
		return l, nil
	}
	cells := []string{"N", "W", "E", "S"}
	origins := []connectionPoint{{71, 0}, {0, 40}, {142, 40}, {71, 80}}
	if l.Layout == "hq" {
		cells = []string{"N", "NW", "NE", "C", "SW", "SE", "S"}
		origins = []connectionPoint{{71, 0}, {0, 40}, {142, 40}, {71, 80}, {0, 120}, {142, 120}, {71, 160}}
	}
	prefix := l.Layout
	if prefix == "depot" {
		prefix = "cmp"
	}
	if prefix == "mountain" {
		prefix = "mount"
	}
	for i, c := range cells {
		l.Tiles = append(l.Tiles, compoundTile{Cell: c, RuntimeID: prefix + c + suffix, Origin: origins[i]})
	}
	return l, nil
}

// Cancel reciprocal edges, retaining only the concave union's outer segments.
func compoundBoundary(l compoundLayout) []connectionSegment {
	type edge struct{ A, B connectionPoint }
	edges := map[edge]bool{}
	g := sharedConnectionGeometry()
	for _, t := range l.Tiles {
		for i, a := range g.Polygon {
			b := g.Polygon[(i+1)%len(g.Polygon)]
			e := edge{a.add(t.Origin), b.add(t.Origin)}
			reverse := edge{e.B, e.A}
			if edges[reverse] {
				delete(edges, reverse)
			} else {
				edges[e] = true
			}
		}
	}
	// Preserve deterministic order independently of map iteration.
	var out []connectionSegment
	for _, t := range l.Tiles {
		for i, a := range g.Polygon {
			e := edge{a.add(t.Origin), g.Polygon[(i+1)%len(g.Polygon)].add(t.Origin)}
			if edges[e] {
				out = append(out, connectionSegment{e.A, e.B})
			}
		}
	}
	return out
}

func compoundSignedDistance(l compoundLayout, edges []connectionSegment, p connectionPoint) float64 {
	d := math.Inf(1)
	for _, e := range edges {
		v := e.b.sub(e.a)
		q := p.sub(e.a)
		u := max(0., min(1., (q[0]*v[0]+q[1]*v[1])/(v[0]*v[0]+v[1]*v[1])))
		d = min(d, p.sub(e.a.add(v.mul(u))).length())
	}
	for _, t := range l.Tiles {
		if connectionInside(sharedConnectionGeometry(), p.sub(t.Origin)) {
			return d
		}
	}
	return -d
}

func compoundDistance(l compoundLayout, p connectionPoint) float64 {
	return compoundSignedDistance(l, compoundBoundary(l), p)
}

type compoundTransform struct {
	Scale  float64         `json:"scale"`
	Offset connectionPoint `json:"offset"`
	Source dimensions      `json:"source_dimensions"`
}
type compoundWeight struct {
	first   int
	weights []float64
}

const compoundMaxAxisTaps = 4_000_000

// Budget full kernel evaluation, including taps outside the source that are
// evaluated for transparent-extension normalization. Each output axis sample
// has at most ceil(6*max(1,1/scale))+1 integer taps. This conservative bound
// also caps retained float64 weights at 32,000,000 bytes, plus one temporary
// kernel of at most the same size. Compare in float64 before integer conversion
// or allocation, so tiny scales/overflowing reciprocals reject safely.
func compoundCheckTapBudget(width, height int, scale float64) error {
	if width <= 0 || height <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) || scale <= 0 {
		return errors.New("compound tap budget requires positive dimensions and finite positive scale")
	}
	support := math.Ceil(6*max(1., 1/scale)) + 1
	taps := (float64(width) + float64(height)) * support
	if math.IsInf(taps, 0) || taps > compoundMaxAxisTaps {
		return fmt.Errorf("compound tap budget exceeded: conservative full-support axis taps %.0f > %d", taps, compoundMaxAxisTaps)
	}
	return nil
}

// compoundAxis returns the Lanczos3 taps of every output sample. Out-of-source
// taps are transparent; with clamp they are dropped and the in-source taps
// renormalized, so an opaque source edge stays opaque in the output.
func compoundAxis(count, source int, scale, offset float64, clamp bool) []compoundWeight {
	out := make([]compoundWeight, count)
	radius := max(1., 1/scale)
	for d := range count {
		center := (float64(d)+.5-offset)/scale - .5
		first, end := int(math.Ceil(center-3*radius)), int(math.Floor(center+3*radius))+1
		w := make([]float64, end-first)
		sum := 0.
		for i := range w {
			u := math.Abs((float64(first+i) - center) / radius)
			v := math.Pi * u
			if u == 0 {
				w[i] = 1
			} else if u < 3 && u != 1 && u != 2 {
				w[i] = math.Sin(v) / v * math.Sin(v/3) / (v / 3)
			}
			sum += w[i]
		}
		// Full support normalization; out-of-source taps are transparent, not edge-clamped.
		lo, hi := max(first, 0), min(end, source)
		if hi <= lo {
			continue
		}
		clipped := make([]float64, hi-lo)
		if clamp {
			sum = 0
			for i := range clipped {
				sum += w[lo-first+i]
			}
		}
		for i := range clipped {
			clipped[i] = w[lo-first+i] / sum
		}
		out[d] = compoundWeight{lo, clipped}
	}
	return out
}

// compoundFit says how the artwork maps onto the scene box. Fill scales it to
// cover the box, crops the excess and keeps its opaque edges opaque, so streets
// painted to the canvas edge reach the box edge; shift then moves it by world
// pixels (to put a street on its port) as long as it still covers the box.
// Without fill the artwork is centred inside the box (letterboxed). Zoom
// (fill only, 0 meaning 1) enlarges the filled artwork for more shift slack.
type compoundFit struct {
	Fill  bool
	Shift connectionPoint
	Zoom  float64
}

// Uniform source-edge affine transform; separable float64 premultiplied Lanczos3.
// Horizontal values remain unquantized, including signed lobes, until final output.
func compoundAffine(src *image.NRGBA, l compoundLayout, inset float64, fit compoundFit) (*image.NRGBA, compoundTransform, error) {
	tr := compoundTransform{Source: dimensions{src.Bounds().Dx(), src.Bounds().Dy()}}
	if math.IsNaN(inset) || math.IsInf(inset, 0) || inset < .5 || inset > 1 {
		return nil, tr, errors.New("inset must be finite in [0.5,1]")
	}
	sx, sy := float64(l.Width)/float64(tr.Source.Width), float64(l.Height)/float64(tr.Source.Height)
	tr.Scale = min(sx, sy) * inset
	if fit.Fill {
		tr.Scale = max(sx, sy) * inset
		if fit.Zoom != 0 {
			tr.Scale *= fit.Zoom
		}
	} else if fit.Shift != (connectionPoint{}) || fit.Zoom != 0 {
		return nil, tr, errors.New("compound shift and zoom require fill")
	}
	tr.Offset = connectionPoint{(float64(l.Width)-float64(tr.Source.Width)*tr.Scale)/2 + fit.Shift[0], (float64(l.Height)-float64(tr.Source.Height)*tr.Scale)/2 + fit.Shift[1]}
	if fit.Fill {
		const eps = 1e-9
		if tr.Offset[0] > eps || tr.Offset[1] > eps || tr.Offset[0]+float64(tr.Source.Width)*tr.Scale < float64(l.Width)-eps || tr.Offset[1]+float64(tr.Source.Height)*tr.Scale < float64(l.Height)-eps {
			return nil, tr, errors.New("compound fill artwork does not cover the scene box")
		}
	}
	if tr.Scale*4 > 1 {
		return nil, tr, errors.New("compound artwork would upscale at 4x")
	}
	width, height := l.Width*4, l.Height*4
	if err := compoundCheckTapBudget(width, height, tr.Scale*4); err != nil {
		return nil, tr, err
	}
	wx := compoundAxis(width, tr.Source.Width, tr.Scale*4, tr.Offset[0]*4, fit.Fill)
	wy := compoundAxis(height, tr.Source.Height, tr.Scale*4, tr.Offset[1]*4, fit.Fill)
	// Choose the smaller intermediate so a legal tall/narrow source does not
	// multiply its long dimension by the full output width. Actual 1536x1024
	// artwork retains horizontal-first arithmetic and byte-identical exports.
	sourceHeight := tr.Source.Height
	transposed := width*tr.Source.Height > height*tr.Source.Width
	if transposed {
		width = height
		wx, wy = wy, wx
		sourceHeight = tr.Source.Width
	}
	tmp := make([][4]float64, width*sourceHeight)
	for y := range sourceHeight {
		for x, weights := range wx {
			var q [4]float64
			for i, w := range weights.weights {
				sx, sy := weights.first+i, y
				if transposed {
					sx, sy = sy, sx
				}
				p := src.NRGBAAt(sx+src.Bounds().Min.X, sy+src.Bounds().Min.Y)
				a := float64(p.A) * w
				q[0] += a * float64(p.R)
				q[1] += a * float64(p.G)
				q[2] += a * float64(p.B)
				q[3] += a
			}
			tmp[y*width+x] = q
		}
	}
	dst := image.NewNRGBA(image.Rect(0, 0, l.Width*4, l.Height*4))
	for y, weights := range wy {
		for x := range width {
			var q [4]float64
			for i, w := range weights.weights {
				p := tmp[(weights.first+i)*width+x]
				for k := range q {
					q[k] += p[k] * w
				}
			}
			if q[3] <= 0 {
				continue
			}
			a := uint8(math.Floor(min(255., q[3]) + .5))
			if a == 0 {
				continue
			}
			channel := func(v float64) uint8 { return uint8(math.Floor(min(max(v, 0), 255*q[3])/q[3] + .5)) }
			dx, dy := x, y
			if transposed {
				dx, dy = dy, dx
			}
			dst.SetNRGBA(dx, dy, color.NRGBA{channel(q[0]), channel(q[1]), channel(q[2]), a})
		}
	}
	return dst, tr, nil
}

type compoundGuardReport struct {
	Policy                string   `json:"policy,omitempty"`
	MaxBoundaryAlpha      uint8    `json:"max_boundary_alpha"`
	BoundaryPixels        int      `json:"boundary_pixels"`
	BoundaryAlphaMass     uint64   `json:"boundary_alpha_mass"`
	ClippedPixels         int      `json:"clipped_pixels"`
	ClippedAlphaMass      uint64   `json:"clipped_alpha_mass"`
	GuardedOpaque         int      `json:"guarded_opaque_count"`
	MinimumOpaqueDistance *float64 `json:"minimum_opaque_distance_world,omitempty"`
	// FeatherWidth and FeatherCorridor (world px) describe an edge feather;
	// its alpha reduction is counted in ClippedPixels and ClippedAlphaMass.
	FeatherWidth    float64 `json:"feather_width_world,omitempty"`
	FeatherCorridor float64 `json:"feather_corridor_world,omitempty"`
}

// Edge feather (world px): the band at the outer boundary of a scene in which
// its artwork fades out, the hard cut of alpha zero next to the boundary, the
// width of the smoothstep that ends the fade, and how far outside a keep
// polygon the feather eases back in.
const (
	compoundFeatherMin, compoundFeatherMax  = 4., 12.
	compoundFeatherCut, compoundFeatherRamp = 2.5, 1.25
	compoundKeepRamp                        = 2.
)

// Limits of the keep polygons of an edge feather.
const compoundMaxKeep, compoundMaxKeepPoints = 24, 32

// compoundValidateFeather allows an edge feather on the reference-layout
// scenes whose entry road keeps the corridor to the port (HQ, depot, ruined
// depot, factory), so the ground apron ends in the ground instead of a
// straight cut along the hex edges. Keep polygons (world px of the scene box)
// protect what must keep its straight edge at the boundary: walls, towers,
// fences and buildings.
func compoundValidateFeather(l compoundLayout, framing string, width float64, entryRoad bool, keep [][]connectionPoint) error {
	if width == 0 {
		if len(keep) != 0 {
			return errors.New("edge_keep requires edge_feather")
		}
		return nil
	}
	switch l.Layout {
	case "hq", "depot", "ruin", "factory":
	default:
		return errors.New("edge_feather is restricted to the HQ, depot, ruined depot and factory")
	}
	if framing != "reference-layout" || !entryRoad {
		return errors.New("edge_feather requires reference-layout framing and an entry_road")
	}
	if !(width >= compoundFeatherMin && width <= compoundFeatherMax) {
		return fmt.Errorf("edge_feather must lie in [%g,%g] world px", compoundFeatherMin, compoundFeatherMax)
	}
	if len(keep) > compoundMaxKeep {
		return fmt.Errorf("edge_keep allows at most %d polygons", compoundMaxKeep)
	}
	for _, poly := range keep {
		if len(poly) < 3 || len(poly) > compoundMaxKeepPoints {
			return fmt.Errorf("edge_keep polygons need 3 to %d points", compoundMaxKeepPoints)
		}
		for _, p := range poly {
			if !(p[0] >= -8 && p[0] <= float64(l.Width)+8 && p[1] >= -8 && p[1] <= float64(l.Height)+8) {
				return errors.New("edge_keep points must lie within 8 world px of the scene box")
			}
		}
	}
	return nil
}

// compoundKeepDistance is how far p lies outside the nearest keep polygon
// (zero inside one, +Inf without polygons).
func compoundKeepDistance(keep [][]connectionPoint, p connectionPoint) float64 {
	best := math.Inf(1)
	for _, poly := range keep {
		inside := false
		d := math.Inf(1)
		for i, a := range poly {
			b := poly[(i+1)%len(poly)]
			if (a[1] > p[1]) != (b[1] > p[1]) && p[0] < a[0]+(p[1]-a[1])*(b[0]-a[0])/(b[1]-a[1]) {
				inside = !inside
			}
			v, q := b.sub(a), p.sub(a)
			u := max(0., min(1., (q[0]*v[0]+q[1]*v[1])/(v[0]*v[0]+v[1]*v[1])))
			d = min(d, q.sub(v.mul(u)).length())
		}
		if inside {
			return 0
		}
		best = min(best, d)
	}
	return best
}

// compoundHash maps a lattice point to [0,1) (integer finaliser of murmur3).
func compoundHash(x, y int, seed uint32) float64 {
	h := uint32(x)*0x8da6b343 ^ uint32(y)*0xd8163841 ^ seed*0xcb1ab31f
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return float64(h) / (1 << 32)
}

// compoundValueNoise is smooth value noise in [0,1) on a lattice of cell world px.
func compoundValueNoise(p connectionPoint, cell float64, seed uint32) float64 {
	x, y := p[0]/cell, p[1]/cell
	x0, y0 := math.Floor(x), math.Floor(y)
	sx, sy := connectionSmoothStep(x-x0), connectionSmoothStep(y-y0)
	i, j := int(x0), int(y0)
	top := compoundHash(i, j, seed)*(1-sx) + compoundHash(i+1, j, seed)*sx
	bottom := compoundHash(i, j+1, seed)*(1-sx) + compoundHash(i+1, j+1, seed)*sx
	return top*(1-sy) + bottom*sy
}

// compoundFeatherEdge fades 4x scene artwork out towards the outer union
// boundary along an irregular line, so rubble or a ground apron ends in the
// ground instead of a straight cut along the hex edges. At signed distance d the
// alpha is scaled by smoothstep((d-e)/ramp), where the edge e wanders between
// the cut and width-ramp with two octaves of value noise in world coordinates
// (the same in every season). Within corridor world px beside the entry port
// (fully within corridor-2) the artwork is kept, so the entry street stays
// open to the port. Inside a keep polygon the artwork is kept exactly; within
// compoundKeepRamp world px outside it the feather eases back in by smoothstep.
func compoundFeatherEdge(art *image.NRGBA, l compoundLayout, stats compoundGuardReport, width, corridor float64, keep [][]connectionPoint) (*image.NRGBA, compoundGuardReport, error) {
	if !(width >= compoundFeatherMin && width <= compoundFeatherMax) || !(corridor >= 2 && corridor <= 24) {
		return nil, stats, errors.New("edge feather width or corridor out of range")
	}
	origin, port, err := compoundEntryPort(l)
	if err != nil {
		return nil, stats, err
	}
	g := sharedConnectionGeometry()
	at := origin.add(g.Ports[port-1])
	inward := g.Neighbors[port-1].mul(-1 / g.Neighbors[port-1].length())
	stats.FeatherWidth, stats.FeatherCorridor = width, corridor
	stats.Policy += fmt.Sprintf("; edge feather: alpha scaled by smoothstep((d-e)/%g) with e between %g and %g world px by value noise, entry corridor %g world px", compoundFeatherRamp, compoundFeatherCut, width-compoundFeatherRamp, corridor)
	if len(keep) > 0 {
		stats.Policy += fmt.Sprintf(", %d keep polygons eased out over %g world px", len(keep), compoundKeepRamp)
	}
	edges := compoundBoundary(l)
	out := image.NewNRGBA(art.Bounds())
	copy(out.Pix, art.Pix)
	for y := range art.Bounds().Dy() {
		for x := range art.Bounds().Dx() {
			c := art.NRGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
			d := compoundSignedDistance(l, edges, p)
			if d >= width {
				continue
			}
			n := .55*compoundValueNoise(p, 9, 1) + .45*compoundValueNoise(p, 3, 2)
			n = connectionSmoothStep((n - .2) / .6)
			e := compoundFeatherCut + (width-compoundFeatherRamp-compoundFeatherCut)*n
			f := connectionSmoothStep((d - e) / compoundFeatherRamp)
			q := p.sub(at)
			if depth := q[0]*inward[0] + q[1]*inward[1]; depth < width+2 {
				lateral := math.Abs(q[0]*inward[1] - q[1]*inward[0])
				f = max(f, 1-connectionSmoothStep((lateral-(corridor-2))/2))
			}
			if f < 1 && len(keep) > 0 {
				f = max(f, 1-connectionSmoothStep(compoundKeepDistance(keep, p)/compoundKeepRamp))
			}
			a := uint8(math.Floor(float64(c.A)*f + .5))
			if a == c.A {
				continue
			}
			stats.ClippedPixels++
			stats.ClippedAlphaMass += uint64(c.A - a)
			c.A = a
			if a == 0 {
				c = color.NRGBA{}
			}
			out.SetNRGBA(x, y, c)
		}
	}
	return out, stats, nil
}

// Reference-layout is explicit because original compound fences and entrance
// roads intentionally reach the union boundary. Other artwork keeps the strict
// eight-world-pixel exclusion and fade, including existing manifests.
func compoundValidateFraming(l compoundLayout, framing string) error {
	switch framing {
	case "":
		return nil
	case "reference-layout":
		if l.Layout == "hq" || l.Layout == "depot" || l.Layout == "ruin" || l.Layout == "factory" {
			return nil
		}
		return errors.New("reference-layout framing is restricted to HQ, depot, ruined depot and factory")
	default:
		return fmt.Errorf("unsupported compound framing %q", framing)
	}
}

func compoundFrame(raw *image.NRGBA, l compoundLayout, framing string) (*image.NRGBA, compoundGuardReport, error) {
	if err := compoundValidateFraming(l, framing); err != nil {
		return nil, compoundGuardReport{}, err
	}
	if framing == "" {
		return compoundGuard(raw, l)
	}
	stats := compoundGuardReport{Policy: "reference-layout: eight-world-pixel band is diagnostic only; no rejection or fade; final compound union mask clips outside after density resampling"}
	edges := compoundBoundary(l)
	for y := range raw.Bounds().Dy() {
		for x := range raw.Bounds().Dx() {
			c := raw.NRGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			d := compoundSignedDistance(l, edges, connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4})
			if c.A >= 250 && (stats.MinimumOpaqueDistance == nil || d < *stats.MinimumOpaqueDistance) {
				dd := d
				stats.MinimumOpaqueDistance = &dd
			}
			if d <= 8 {
				stats.BoundaryPixels++
				stats.BoundaryAlphaMass += uint64(c.A)
				stats.MaxBoundaryAlpha = max(stats.MaxBoundaryAlpha, c.A)
				if c.A >= 250 {
					stats.GuardedOpaque++
				}
			}
		}
	}
	// Keep the complete source through 2x Lanczos filtering. Union masking
	// happens in compoundCompose, before complementary tile crops.
	return raw, stats, nil
}

func compoundGuard(raw *image.NRGBA, l compoundLayout) (*image.NRGBA, compoundGuardReport, error) {
	stats := compoundGuardReport{}
	edges := compoundBoundary(l)
	dst := image.NewNRGBA(raw.Bounds())
	unsafe := false
	for y := range raw.Bounds().Dy() {
		for x := range raw.Bounds().Dx() {
			c := raw.NRGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			d := compoundSignedDistance(l, edges, connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4})
			if c.A >= 250 {
				if stats.MinimumOpaqueDistance == nil || d < *stats.MinimumOpaqueDistance {
					dd := d
					stats.MinimumOpaqueDistance = &dd
				}
			}
			if d <= 8 {
				stats.BoundaryPixels++
				stats.BoundaryAlphaMass += uint64(c.A)
				stats.MaxBoundaryAlpha = max(stats.MaxBoundaryAlpha, c.A)
				if c.A >= 250 {
					stats.GuardedOpaque++
					unsafe = true
				}
			}
			before := c.A
			c.A = uint8(math.Floor(float64(c.A)*connectionSmoothStep((d-4)/4) + .5))
			if c.A < before {
				stats.ClippedPixels++
				stats.ClippedAlphaMass += uint64(before - c.A)
			}
			if c.A == 0 {
				c = color.NRGBA{}
			}
			dst.SetNRGBA(x, y, c)
		}
	}
	if unsafe {
		return nil, stats, fmt.Errorf("unsafe compound framing: %d opaque sampled pixels within 8 world pixels of outer boundary or outside", stats.GuardedOpaque)
	}
	return dst, stats, nil
}
