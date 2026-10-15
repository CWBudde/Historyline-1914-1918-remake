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
	case "hq-summer", "hq-winter":
		l.Layout = "hq"
		l.Height = 240
	case "factory-winter":
		l.Layout = "factory"
		l.EntryCell = "E"
	case "mountain-summer", "mountain-winter":
		l.Layout = "mountain"
		l.EntryCell = ""
	case "mountain-single-summer", "mountain-single-winter":
		l.Layout = "mountain-single"
		l.Width, l.Height, l.EntryCell = 96, 80, ""
	case "hamlet-summer", "hamlet-winter":
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
		l.Tiles = []compoundTile{{Cell: "C", RuntimeID: "house" + suffix, Origin: connectionPoint{0, 0}}}
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

func compoundAxis(count, source int, scale, offset float64) []compoundWeight {
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
		for i := range clipped {
			clipped[i] = w[lo-first+i] / sum
		}
		out[d] = compoundWeight{lo, clipped}
	}
	return out
}

// Uniform source-edge affine transform; separable float64 premultiplied Lanczos3.
// Horizontal values remain unquantized, including signed lobes, until final output.
func compoundAffine(src *image.NRGBA, l compoundLayout, inset float64) (*image.NRGBA, compoundTransform, error) {
	tr := compoundTransform{Source: dimensions{src.Bounds().Dx(), src.Bounds().Dy()}}
	if math.IsNaN(inset) || math.IsInf(inset, 0) || inset < .5 || inset > 1 {
		return nil, tr, errors.New("inset must be finite in [0.5,1]")
	}
	tr.Scale = min(float64(l.Width)/float64(tr.Source.Width), float64(l.Height)/float64(tr.Source.Height)) * inset
	tr.Offset = connectionPoint{(float64(l.Width) - float64(tr.Source.Width)*tr.Scale) / 2, (float64(l.Height) - float64(tr.Source.Height)*tr.Scale) / 2}
	if tr.Scale*4 > 1 {
		return nil, tr, errors.New("compound artwork would upscale at 4x")
	}
	width, height := l.Width*4, l.Height*4
	if err := compoundCheckTapBudget(width, height, tr.Scale*4); err != nil {
		return nil, tr, err
	}
	wx := compoundAxis(width, tr.Source.Width, tr.Scale*4, tr.Offset[0]*4)
	wy := compoundAxis(height, tr.Source.Height, tr.Scale*4, tr.Offset[1]*4)
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
	MaxBoundaryAlpha      uint8    `json:"max_boundary_alpha"`
	BoundaryPixels        int      `json:"boundary_pixels"`
	BoundaryAlphaMass     uint64   `json:"boundary_alpha_mass"`
	ClippedPixels         int      `json:"clipped_pixels"`
	ClippedAlphaMass      uint64   `json:"clipped_alpha_mass"`
	GuardedOpaque         int      `json:"guarded_opaque_count"`
	MinimumOpaqueDistance *float64 `json:"minimum_opaque_distance_world,omitempty"`
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
