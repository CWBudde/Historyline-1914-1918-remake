package main

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/lookmeasure"
)

type woodlandAnnotations struct {
	Schema             int                        `json:"schema_version"`
	LevelSHA256        string                     `json:"canonical_level_sha256"`
	RegistrationSHA256 string                     `json:"registration_sha256"`
	Coordinates        string                     `json:"coordinates"`
	Sources            []woodlandSourceAnnotation `json:"sources"`
}

type woodlandSourceAnnotation struct {
	ID           string             `json:"id"`
	Width        int                `json:"width"`
	Height       int                `json:"height"`
	PixelsSHA256 string             `json:"rgba_pixels_sha256"`
	Scope        string             `json:"scope"`
	Review       string             `json:"review_status"`
	Fragments    []woodlandFragment `json:"fragments"`
}

type woodlandFragment struct {
	ID          string       `json:"id"`
	Points      [][2]float64 `json:"points"`
	Uncertainty float64      `json:"position_uncertainty_world_px"`
	Notes       string       `json:"notes,omitempty"`
}

type lookWoodland struct {
	Available          bool                     `json:"available"`
	Complete           bool                     `json:"whole_forest_perimeter"`
	Reason             string                   `json:"reason,omitempty"`
	Definition         string                   `json:"definition,omitempty"`
	Scope              string                   `json:"scope,omitempty"`
	AnnotationSHA256   string                   `json:"annotations_sha256,omitempty"`
	PinnedPixelsSHA256 string                   `json:"annotated_rgba_pixels_sha256,omitempty"`
	Review             string                   `json:"review_status,omitempty"`
	Fragments          []woodlandFragmentResult `json:"fragments,omitempty"`
	Statistics         woodlandDistanceStats    `json:"statistics"`
	Preview            string                   `json:"preview,omitempty"`
	PreviewSHA256      string                   `json:"preview_sha256,omitempty"`
}

type woodlandFragmentResult struct {
	Annotation woodlandFragment      `json:"annotation"`
	Statistics woodlandDistanceStats `json:"statistics"`
}

type woodlandDistanceStats struct {
	Samples        int      `json:"samples"`
	Length         float64  `json:"sampled_arc_length_world_px"`
	Median         *float64 `json:"weighted_median_distance_world_px"`
	WithinSix      *float64 `json:"fraction_arc_within_6_world_px"`
	WithinSixLower *float64 `json:"within_6_lower_given_position_uncertainty"`
	WithinSixUpper *float64 `json:"within_6_upper_given_position_uncertainty"`
	Reason         string   `json:"reason,omitempty"`
}

type woodlandSample struct{ distance, weight, uncertainty float64 }

func summarizeWoodland(samples []woodlandSample) woodlandDistanceStats {
	result := woodlandDistanceStats{Samples: len(samples)}
	if len(samples) == 0 {
		result.Reason = "insufficient_samples"
		return result
	}
	near, lower, upper := 0., 0., 0.
	for _, s := range samples {
		result.Length += s.weight
		if s.distance <= 6 {
			near += s.weight
		}
		if s.distance+s.uncertainty <= 6 {
			lower += s.weight
		}
		if s.distance-s.uncertainty <= 6 {
			upper += s.weight
		}
	}
	samples = slices.Clone(samples)
	slices.SortFunc(samples, func(a, b woodlandSample) int { return cmp.Compare(a.distance, b.distance) })
	cumulative := 0.
	for i, s := range samples {
		cumulative += s.weight
		if cumulative >= result.Length/2 {
			v := s.distance
			if math.Abs(cumulative-result.Length/2) <= result.Length*1e-12 && i+1 < len(samples) {
				v = (v + samples[i+1].distance) / 2
			}
			result.Median = &v
			break
		}
	}
	near /= result.Length
	lower /= result.Length
	upper /= result.Length
	result.WithinSix, result.WithinSixLower, result.WithinSixUpper = &near, &lower, &upper
	return result
}

func measureAnnotatedWoodland(a woodlandSourceAnnotation, im *image.NRGBA, edges []lookmeasure.Segment) (lookWoodland, error) {
	result := lookWoodland{PinnedPixelsSHA256: a.PixelsSHA256, Scope: a.Scope, Review: a.Review, Definition: "independent open outer-crown boundary fragments, excluding cast shadows; no closure or bridging of grass gaps; each source segment sampled at equal-piece midpoints with step <=1 world px and arc-length weights; nearest Euclidean distance to finite edges of all nonexcluded actual tile polygons; median and <=6 fraction weight visible arc length, not vertices or fragment counts; lower/upper fractions assume the declared visual position uncertainty is a displacement radius, not a calibrated confidence interval; fragment evidence cannot satisfy the whole-forest gate"}
	if im == nil || im.Bounds().Min != (image.Point{}) || im.Bounds().Dx() != a.Width || im.Bounds().Dy() != a.Height || fmt.Sprintf("%x", sha256.Sum256(im.Pix)) != a.PixelsSHA256 {
		result.Reason = "annotation_source_mismatch"
		return result, nil
	}
	seen := map[string]bool{}
	var all []woodlandSample
	for _, fragment := range a.Fragments {
		if fragment.ID == "" || seen[fragment.ID] || !isFinite(fragment.Uncertainty) || fragment.Uncertainty <= 0 {
			return result, fmt.Errorf("invalid or duplicate woodland fragment %q", fragment.ID)
		}
		seen[fragment.ID] = true
		points := make([]lookmeasure.Point, len(fragment.Points))
		for i, p := range fragment.Points {
			if !isFinite(p[0]) || !isFinite(p[1]) || p[0] < 0 || p[1] < 0 || p[0] >= float64(a.Width) || p[1] >= float64(a.Height) {
				return result, fmt.Errorf("woodland fragment %s outside source", fragment.ID)
			}
			points[i] = lookmeasure.Point{X: p[0], Y: p[1]}
		}
		measured, err := lookmeasure.PolylineDistances(points, edges, 1)
		if err != nil {
			return result, fmt.Errorf("woodland fragment %s: %w", fragment.ID, err)
		}
		samples := make([]woodlandSample, len(measured))
		for i, s := range measured {
			if im.NRGBAAt(int(math.Floor(s.Point.X)), int(math.Floor(s.Point.Y))).A != 255 {
				return result, fmt.Errorf("woodland fragment %s crosses nonopaque source", fragment.ID)
			}
			samples[i] = woodlandSample{s.Distance, s.Weight, fragment.Uncertainty}
		}
		result.Fragments = append(result.Fragments, woodlandFragmentResult{fragment, summarizeWoodland(samples)})
		all = append(all, samples...)
	}
	result.Statistics = summarizeWoodland(all)
	result.Available = len(all) > 0
	if !result.Available {
		result.Reason = "no_annotated_boundary_fragments"
	}
	return result, nil
}

func (p *project) woodlandHexEdges() []lookmeasure.Segment {
	var edges []lookmeasure.Segment
	for _, c := range p.cells() {
		if p.excluded[c.Row] {
			continue
		}
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		for i, a := range hexPolygon {
			b := hexPolygon[(i+1)%len(hexPolygon)]
			edges = append(edges, lookmeasure.Segment{From: lookmeasure.Point{X: float64(ox) + a[0], Y: float64(oy) + a[1]}, To: lookmeasure.Point{X: float64(ox) + b[0], Y: float64(oy) + b[1]}})
		}
	}
	return edges
}

func (p *project) measureWoodland(id string, im *image.NRGBA, out string) (lookWoodland, error) {
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(p.targetsPath), "woodland_annotations.json"))
	if errors.Is(err, os.ErrNotExist) {
		return lookWoodland{Reason: "independent_woodland_annotations_unavailable"}, nil
	}
	if err != nil {
		return lookWoodland{}, err
	}
	var annotations woodlandAnnotations
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&annotations); err != nil {
		return lookWoodland{}, fmt.Errorf("woodland annotations: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return lookWoodland{}, fmt.Errorf("woodland annotations contain trailing data")
	}
	if annotations.Schema != 1 {
		return lookWoodland{}, fmt.Errorf("unsupported woodland annotation schema")
	}
	pin := fmt.Sprintf("%x", sha256.Sum256(raw))
	lv, err := p.level.Marshal()
	if err != nil {
		return lookWoodland{}, err
	}
	reg, err := json.Marshal(p.ref.Registration)
	if err != nil {
		return lookWoodland{}, err
	}
	if annotations.LevelSHA256 != fmt.Sprintf("%x", sha256.Sum256(lv)) || annotations.RegistrationSHA256 != fmt.Sprintf("%x", sha256.Sum256(reg)) {
		return lookWoodland{Reason: "annotation_registration_or_level_mismatch", AnnotationSHA256: pin}, nil
	}
	seen := map[string]bool{}
	var selected *woodlandSourceAnnotation
	for i := range annotations.Sources {
		source := &annotations.Sources[i]
		if (source.ID != "render" && source.ID != "reference") || seen[source.ID] {
			return lookWoodland{}, fmt.Errorf("invalid or duplicate woodland source %q", source.ID)
		}
		seen[source.ID] = true
		if source.ID == id {
			selected = source
		}
	}
	if selected == nil {
		return lookWoodland{Reason: "independent_woodland_annotations_unavailable", AnnotationSHA256: pin}, nil
	}
	edges := p.woodlandHexEdges()
	result, err := measureAnnotatedWoodland(*selected, im, edges)
	if err != nil {
		return result, err
	}
	result.AnnotationSHA256 = pin
	if result.Available && out != "" {
		result.Preview = "woodland-" + id + "-annotations-3x.png"
		if err := os.MkdirAll(out, 0o755); err != nil {
			return result, err
		}
		dest := filepath.Join(out, result.Preview)
		if err := writePNG(dest, drawWoodlandAnnotations(im, selected.Fragments, edges)); err != nil {
			return result, err
		}
		data, err := os.ReadFile(dest)
		if err != nil {
			return result, err
		}
		result.PreviewSHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	return result, nil
}

func drawWoodlandAnnotations(im *image.NRGBA, fragments []woodlandFragment, edges []lookmeasure.Segment) *image.NRGBA {
	bounds := image.Rectangle{}
	for _, f := range fragments {
		for _, p := range f.Points {
			q := image.Pt(int(p[0]), int(p[1]))
			bounds = bounds.Union(image.Rectangle{Min: q, Max: q.Add(image.Pt(1, 1))})
		}
	}
	bounds = bounds.Inset(-25).Intersect(im.Bounds())
	base := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(base, base.Bounds(), im, bounds.Min, draw.Src)
	paint := func(x, y int, c color.NRGBA) {
		x -= bounds.Min.X
		y -= bounds.Min.Y
		if !image.Pt(x, y).In(base.Bounds()) {
			return
		}
		under := base.NRGBAAt(x, y)
		base.SetNRGBA(x, y, color.NRGBA{uint8((int(under.R)*(255-int(c.A)) + int(c.R)*int(c.A)) / 255), uint8((int(under.G)*(255-int(c.A)) + int(c.G)*int(c.A)) / 255), uint8((int(under.B)*(255-int(c.A)) + int(c.B)*int(c.A)) / 255), 255})
	}
	line := func(a, b [2]float64, c color.NRGBA) {
		n := int(math.Ceil(math.Hypot(b[0]-a[0], b[1]-a[1]) * 2))
		for i := 0; i <= n; i++ {
			t := float64(i) / float64(max(1, n))
			paint(int(math.Round(a[0]+(b[0]-a[0])*t)), int(math.Round(a[1]+(b[1]-a[1])*t)), c)
		}
	}
	for _, e := range edges {
		line([2]float64{e.From.X, e.From.Y}, [2]float64{e.To.X, e.To.Y}, color.NRGBA{255, 200, 50, 70})
	}
	for _, f := range fragments {
		for i := 1; i < len(f.Points); i++ {
			line(f.Points[i-1], f.Points[i], color.NRGBA{20, 240, 240, 255})
		}
		for _, p := range f.Points {
			paint(int(math.Round(p[0])), int(math.Round(p[1])), color.NRGBA{255, 80, 100, 255})
		}
	}
	out := image.NewNRGBA(image.Rect(0, 0, base.Bounds().Dx()*3, base.Bounds().Dy()*3))
	for y := range out.Bounds().Dy() {
		for x := range out.Bounds().Dx() {
			out.SetNRGBA(x, y, base.NRGBAAt(x/3, y/3))
		}
	}
	return out
}
