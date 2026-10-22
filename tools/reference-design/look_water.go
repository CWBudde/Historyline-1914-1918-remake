package main

import (
	"bytes"
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
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type waterAnnotations struct {
	Schema             int                     `json:"schema_version"`
	LevelSHA256        string                  `json:"canonical_level_sha256"`
	RegistrationSHA256 string                  `json:"registration_sha256"`
	Coordinates        string                  `json:"coordinates"`
	Sources            []waterSourceAnnotation `json:"sources"`
}

type waterSourceAnnotation struct {
	ID           string          `json:"id"`
	PixelsSHA256 string          `json:"rgba_pixels_sha256"`
	Width        int             `json:"width"`
	Height       int             `json:"height"`
	Review       string          `json:"review_status"`
	Transects    []waterTransect `json:"transects"`
}

type waterTransect struct {
	ID          string        `json:"id"`
	Feature     string        `json:"feature"`
	Cell        string        `json:"cell,omitempty"`
	Kind        string        `json:"kind"`
	From        [2]float64    `json:"from"`
	To          [2]float64    `json:"to"`
	Uncertainty float64       `json:"boundary_uncertainty_world_px"`
	Spans       []waterSpan   `json:"spans"`
	Regions     []waterRegion `json:"regions"`
	Notes       string        `json:"notes,omitempty"`
}

type waterSpan struct {
	Name  string  `json:"name"`
	Role  string  `json:"role"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type waterRegion struct {
	Role      string  `json:"role"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
	HalfWidth float64 `json:"half_width"`
}

type lookWater struct {
	Available          bool                  `json:"available"`
	Reason             string                `json:"reason,omitempty"`
	AnnotationSHA256   string                `json:"annotations_sha256,omitempty"`
	PinnedPixelsSHA256 string                `json:"annotated_rgba_pixels_sha256,omitempty"`
	Review             string                `json:"review_status,omitempty"`
	Definition         string                `json:"definition,omitempty"`
	Transects          []waterTransectResult `json:"transects,omitempty"`
	Groups             []waterWidthGroup     `json:"width_groups,omitempty"`
	Preview            string                `json:"preview,omitempty"`
	PreviewSHA256      string                `json:"preview_sha256,omitempty"`
}

type waterTransectResult struct {
	Annotation      waterTransect       `json:"annotation"`
	Widths          []waterWidth        `json:"widths"`
	Regions         []waterRegionResult `json:"regions"`
	DepthAppearance waterDelta          `json:"near_bank_minus_interior_l_star"`
	BankShadow      waterDelta          `json:"shadow_minus_near_bank_l_star"`
}

type waterWidth struct {
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	Width       float64 `json:"width_world_px"`
	Uncertainty float64 `json:"worst_case_width_uncertainty_world_px"`
}

type waterRegionResult struct {
	Role    string   `json:"role"`
	Samples int      `json:"samples"`
	L       *float64 `json:"median_l_star"`
	Chroma  *float64 `json:"median_chroma"`
	Reason  string   `json:"reason,omitempty"`
}

type waterDelta struct {
	Value  *float64 `json:"value"`
	Reason string   `json:"reason,omitempty"`
}

type waterWidthGroup struct {
	Feature         string   `json:"feature"`
	Cell            string   `json:"cell,omitempty"`
	Role            string   `json:"role"`
	Samples         int      `json:"samples"`
	Median          *float64 `json:"median_world_px"`
	Variation       *float64 `json:"population_cv"`
	VariationReason string   `json:"variation_unavailable_reason,omitempty"`
}

func waterMedian(v []float64) *float64 {
	if len(v) == 0 {
		return nil
	}
	slices.Sort(v)
	n := len(v)
	m := (v[(n-1)/2] + v[n/2]) / 2
	return &m
}

func waterDifference(a, b waterRegionResult) waterDelta {
	if a.L == nil || b.L == nil {
		return waterDelta{Reason: "independently_identifiable_regions_unavailable"}
	}
	v := *a.L - *b.L
	return waterDelta{Value: &v}
}

// Annotations are independently inspected visual estimates, not exporter masks.
// Pin mismatches yield no measurements, while malformed annotations are errors.
func measureAnnotatedWater(a waterSourceAnnotation, im *image.NRGBA) (lookWater, error) {
	result := lookWater{PinnedPixelsSHA256: a.PixelsSHA256, Review: a.Review, Definition: "independent visual boundary estimates at continuous world coordinates; widths are annotated endpoint distances, with worst-case uncertainty twice boundary uncertainty; variation is population standard deviation divided by mean, requiring >=2 widths; bank widths group by feature, stream widths by feature/cell; pixel-centre strips sample actual opaque registered 1x imagery, reporting median CIELAB D65 L* and chroma; near-bank minus interior is a depth-appearance proxy, not physical shallows; shadow compares only with nearby unshadowed water; no acceptance is asserted"}
	if im == nil || im.Bounds().Min != (image.Point{}) || im.Bounds().Dx() != a.Width || im.Bounds().Dy() != a.Height || fmt.Sprintf("%x", sha256.Sum256(im.Pix)) != a.PixelsSHA256 {
		result.Reason = "annotation_source_mismatch"
		return result, nil
	}
	plane := newLabPlane(im)
	seen := map[string]bool{}
	type groupKey struct{ feature, cell, role string }
	groups := map[groupKey][]float64{}
	var order []groupKey
	for _, tr := range a.Transects {
		if tr.ID == "" || seen[tr.ID] || tr.Feature == "" || (tr.Kind != "stream" && tr.Kind != "bank") || !isFinite(tr.Uncertainty) || tr.Uncertainty <= 0 {
			return result, fmt.Errorf("invalid or duplicate water transect %q", tr.ID)
		}
		seen[tr.ID] = true
		length := math.Hypot(tr.To[0]-tr.From[0], tr.To[1]-tr.From[1])
		if !isFinite(length) || length <= 0 {
			return result, fmt.Errorf("invalid water transect endpoints %s", tr.ID)
		}
		for _, point := range [][2]float64{tr.From, tr.To} {
			if !isFinite(point[0]) || !isFinite(point[1]) || point[0] < 0 || point[0] > float64(a.Width) || point[1] < 0 || point[1] > float64(a.Height) {
				return result, fmt.Errorf("water transect %s outside source", tr.ID)
			}
		}
		section := waterTransectResult{Annotation: tr, Widths: []waterWidth{}, Regions: []waterRegionResult{}}
		names := map[string]bool{}
		for _, span := range tr.Spans {
			if span.Name == "" || names[span.Name] || (span.Role != "bank_width" && span.Role != "stream_width") || !validWaterInterval(span.Start, span.End, length) {
				return result, fmt.Errorf("invalid water span %s/%s", tr.ID, span.Name)
			}
			names[span.Name] = true
			width := span.End - span.Start
			section.Widths = append(section.Widths, waterWidth{span.Name, span.Role, width, 2 * tr.Uncertainty})
			cell := tr.Cell
			if span.Role == "bank_width" {
				cell = ""
			}
			key := groupKey{tr.Feature, cell, span.Role}
			if _, exists := groups[key]; !exists {
				order = append(order, key)
			}
			groups[key] = append(groups[key], width)
		}
		regions := map[string]waterRegionResult{}
		for ri, region := range tr.Regions {
			if region.Role != "interior_water" && region.Role != "near_bank_water" && region.Role != "bank_shadow" {
				return result, fmt.Errorf("unknown water region %s/%s", tr.ID, region.Role)
			}
			if _, exists := regions[region.Role]; exists {
				return result, fmt.Errorf("duplicate water region %s/%s", tr.ID, region.Role)
			}
			for _, previous := range tr.Regions[:ri] {
				if max(previous.Start, region.Start) < min(previous.End, region.End) {
					return result, fmt.Errorf("overlapping water regions %s", tr.ID)
				}
			}
			pixels, err := lookmeasure.StripPixels(a.Width, a.Height, tr.From, tr.To, region.Start, region.End, region.HalfWidth)
			if err != nil {
				return result, fmt.Errorf("%s/%s: %w", tr.ID, region.Role, err)
			}
			measured := waterRegionResult{Role: region.Role, Samples: len(pixels)}
			var light, chroma []float64
			for _, i := range pixels {
				if !plane.ok[i] {
					measured.Reason = "nonopaque_strip"
					break
				}
				light = append(light, plane.l[i])
				chroma = append(chroma, math.Hypot(plane.a[i], plane.b[i]))
			}
			if measured.Reason == "" {
				if len(pixels) < 3 {
					measured.Reason = "insufficient_samples"
				} else {
					measured.L, measured.Chroma = waterMedian(light), waterMedian(chroma)
				}
			}
			regions[region.Role] = measured
			section.Regions = append(section.Regions, measured)
		}
		section.DepthAppearance = waterDifference(regions["near_bank_water"], regions["interior_water"])
		section.BankShadow = waterDifference(regions["bank_shadow"], regions["near_bank_water"])
		result.Transects = append(result.Transects, section)
	}
	for _, key := range order {
		values := groups[key]
		group := waterWidthGroup{Feature: key.feature, Cell: key.cell, Role: key.role, Samples: len(values), Median: waterMedian(slices.Clone(values))}
		if len(values) < 2 {
			group.VariationReason = "insufficient_widths"
		} else {
			mean := 0.
			for _, v := range values {
				mean += v
			}
			mean /= float64(len(values))
			variance := 0.
			for _, v := range values {
				variance += (v - mean) * (v - mean)
			}
			cv := math.Sqrt(variance/float64(len(values))) / mean
			group.Variation = &cv
		}
		result.Groups = append(result.Groups, group)
	}
	result.Available = len(result.Transects) > 0
	if !result.Available {
		result.Reason = "no_annotated_transects"
	}
	return result, nil
}

func isFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func validWaterInterval(start, end, length float64) bool {
	return isFinite(start) && isFinite(end) && start >= 0 && end > start && end <= length
}

func (p *project) measureWater(id string, im *image.NRGBA, out string) (lookWater, error) {
	path := filepath.Join(filepath.Dir(p.targetsPath), "water_annotations.json")
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return lookWater{Reason: "independent_water_annotations_unavailable"}, nil
	}
	if err != nil {
		return lookWater{}, err
	}
	var annotations waterAnnotations
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&annotations); err != nil {
		return lookWater{}, fmt.Errorf("water annotations: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return lookWater{}, fmt.Errorf("water annotations contain trailing data")
	}
	if annotations.Schema != 1 {
		return lookWater{}, fmt.Errorf("unsupported water annotation schema")
	}
	pin := fmt.Sprintf("%x", sha256.Sum256(raw))
	lv, err := p.level.Marshal()
	if err != nil {
		return lookWater{}, err
	}
	registration, err := json.Marshal(p.ref.Registration)
	if err != nil {
		return lookWater{}, err
	}
	if annotations.LevelSHA256 != fmt.Sprintf("%x", sha256.Sum256(lv)) || annotations.RegistrationSHA256 != fmt.Sprintf("%x", sha256.Sum256(registration)) {
		return lookWater{Reason: "annotation_registration_or_level_mismatch", AnnotationSHA256: pin}, nil
	}
	seen := map[string]bool{}
	var selected *waterSourceAnnotation
	for i := range annotations.Sources {
		source := &annotations.Sources[i]
		if (source.ID != "render" && source.ID != "reference") || seen[source.ID] {
			return lookWater{}, fmt.Errorf("invalid or duplicate water source %q", source.ID)
		}
		seen[source.ID] = true
		if source.ID == id {
			selected = source
		}
	}
	if selected == nil {
		return lookWater{Reason: "independent_water_annotations_unavailable", AnnotationSHA256: pin}, nil
	}
	result, err := measureAnnotatedWater(*selected, im)
	if err != nil {
		return result, err
	}
	result.AnnotationSHA256 = pin
	if result.Available && out != "" {
		result.Preview = "water-" + id + "-annotations.png"
		preview := drawWaterAnnotations(im, selected.Transects)
		dest := filepath.Join(out, result.Preview)
		if err := os.MkdirAll(out, 0755); err != nil {
			return result, err
		}
		if err := writePNG(dest, preview); err != nil {
			return result, err
		}
		encoded, err := os.ReadFile(dest)
		if err != nil {
			return result, err
		}
		result.PreviewSHA256 = fmt.Sprintf("%x", sha256.Sum256(encoded))
	}
	return result, nil
}

func drawWaterAnnotations(im *image.NRGBA, transects []waterTransect) *image.NRGBA {
	// Crop to the inspected features, preserving 1 world pixel = 1 output pixel.
	bounds := image.Rectangle{}
	for _, tr := range transects {
		for _, v := range [][2]float64{tr.From, tr.To} {
			p := image.Pt(int(math.Floor(v[0])), int(math.Floor(v[1])))
			bounds = bounds.Union(image.Rectangle{Min: p, Max: p.Add(image.Pt(1, 1))})
		}
	}
	bounds = bounds.Inset(-25).Intersect(im.Bounds())
	out := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(out, out.Bounds(), im, bounds.Min, draw.Src)
	colors := map[string]color.NRGBA{"stream_width": {40, 230, 230, 255}, "bank_width": {255, 160, 30, 255}, "interior_water": {30, 230, 100, 255}, "near_bank_water": {220, 100, 255, 255}, "bank_shadow": {255, 70, 100, 255}}
	mark := func(x, y int, c color.NRGBA) {
		for k := -2; k <= 2; k++ {
			out.SetNRGBA(x+k, y, c)
			out.SetNRGBA(x, y+k, c)
		}
	}
	for _, tr := range transects {
		length := math.Hypot(tr.To[0]-tr.From[0], tr.To[1]-tr.From[1])
		ux, uy := (tr.To[0]-tr.From[0])/length, (tr.To[1]-tr.From[1])/length
		point := func(s float64) image.Point {
			return image.Pt(int(math.Round(tr.From[0]+ux*s))-bounds.Min.X, int(math.Round(tr.From[1]+uy*s))-bounds.Min.Y)
		}
		for _, span := range tr.Spans {
			c := colors[span.Role]
			for s := span.Start; s < span.End; s += .25 {
				p := point(s)
				out.SetNRGBA(p.X, p.Y, c)
			}
			for _, s := range []float64{span.Start, span.End} {
				p := point(s)
				mark(p.X, p.Y, c)
			}
		}
		for _, r := range tr.Regions {
			pixels, err := lookmeasure.StripPixels(im.Bounds().Dx(), im.Bounds().Dy(), tr.From, tr.To, r.Start, r.End, r.HalfWidth)
			if err != nil {
				continue
			}
			for _, i := range pixels {
				x, y := i%im.Bounds().Dx()-bounds.Min.X, i/im.Bounds().Dx()-bounds.Min.Y
				c := out.NRGBAAt(x, y)
				t := colors[r.Role]
				out.SetNRGBA(x, y, color.NRGBA{uint8((int(c.R) + int(t.R)) / 2), uint8((int(c.G) + int(t.G)) / 2), uint8((int(c.B) + int(t.B)) / 2), 255})
			}
		}
		p := point(0)
		labelX := max(2, min(p.X+4, out.Bounds().Dx()-font.MeasureString(basicfont.Face7x13, tr.ID).Ceil()-2))
		d := font.Drawer{Dst: out, Src: image.White, Face: basicfont.Face7x13, Dot: fixed.P(labelX, max(13, p.Y-4))}
		d.DrawString(tr.ID)
	}
	return out
}
