package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

// This is the first Look0 measurement batch, not the full acceptance gate.
// Missing feature measurements and absent private originals stay explicit.
type lookReport struct {
	Schema       int               `json:"schema_version"`
	Accepted     bool              `json:"accepted"`
	Gate         *lookGate         `json:"acceptance_gate"`
	Definitions  map[string]string `json:"definitions"`
	Level        string            `json:"level"`
	LevelSHA256  string            `json:"canonical_level_sha256"`
	AssetsSHA256 string            `json:"assets_index_sha256"`
	Sources      []lookSource      `json:"sources"`
	Remaining    []string          `json:"remaining_acceptance_work"`
}

type lookSource struct {
	ID           string        `json:"id"`
	Available    bool          `json:"available"`
	Reason       string        `json:"reason,omitempty"`
	Width        int           `json:"width,omitempty"`
	Height       int           `json:"height,omitempty"`
	PixelsSHA256 string        `json:"rgba_pixels_sha256,omitempty"`
	FileSHA256   string        `json:"source_file_sha256,omitempty"`
	Whole        *lookMetrics  `json:"whole_map,omitempty"`
	Families     []lookFamily  `json:"families,omitempty"`
	Canopy       *lookCanopy   `json:"canopy"`
	Units        *lookUnits    `json:"unit_edge_contrast"`
	Water        *lookWater    `json:"water"`
	Woodland     *lookWoodland `json:"woodland_edge"`
	Light        *lookLight    `json:"cast_shadow_direction"`
}

type lookFamily struct {
	ID           string      `json:"id"`
	Cells        []string    `json:"clean_cells"`
	Colour       *metrics    `json:"colour"`
	ColourReason string      `json:"colour_unavailable_reason,omitempty"`
	Statistics   lookMetrics `json:"statistics"`
}

func (p *project) lookCells(f family) []cell {
	if f.Kind != "terrain" {
		return nil
	}
	var cells []cell
	for _, c := range p.occurrences(f) {
		h, _ := p.hex(c)
		if len(h.Infra.IDs) == 0 && h.Content == nil {
			cells = append(cells, c)
		}
	}
	return cells
}

func lookMask(plane *labPlane, cells []cell) []bool {
	mask := make([]bool, plane.w*plane.h)
	for _, c := range cells {
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		for y := range hexgrid.HexHeight {
			for x := range hexgrid.HexWidth {
				px, py := ox+x, oy+y
				if px >= 0 && py >= 0 && px < plane.w && py < plane.h && insideTile(x, y, 1) {
					mask[py*plane.w+px] = true
				}
			}
		}
	}
	return mask
}

func (p *project) measureLookSource(id string, im *image.NRGBA) lookSource {
	plane := newLabPlane(im)
	source := lookSource{ID: id, Available: true, Width: plane.w, Height: plane.h, PixelsSHA256: fmt.Sprintf("%x", sha256.Sum256(im.Pix))}
	var cells []cell
	for _, c := range p.cells() {
		if !p.excluded[c.Row] {
			cells = append(cells, c)
		}
	}
	whole := measureLook(plane, lookMask(plane, cells))
	source.Whole = &whole
	for _, f := range p.ref.Families {
		if f.Kind != "terrain" {
			continue
		}
		clean := p.lookCells(f)
		family := lookFamily{ID: f.ID, Cells: []string{}, Statistics: measureLook(plane, lookMask(plane, clean))}
		if family.Statistics.Samples > 0 {
			colour := plane.measure(clean)
			family.Colour = &colour
		} else {
			family.ColourReason = "insufficient_samples"
		}
		for _, c := range clean {
			family.Cells = append(family.Cells, c.String())
		}
		source.Families = append(source.Families, family)
	}
	return source
}

func (p *project) writeLookReport(out string) error {
	render, err := p.renderMap()
	if err != nil {
		return err
	}
	levelBytes, err := p.level.Marshal()
	if err != nil {
		return err
	}
	catalog, err := fs.ReadFile(p.assets, "index.json")
	if err != nil {
		return fmt.Errorf("look catalogue provenance: %w", err)
	}
	report := lookReport{Schema: 1, Level: p.ref.Level, LevelSHA256: fmt.Sprintf("%x", sha256.Sum256(levelBytes)), AssetsSHA256: fmt.Sprintf("%x", sha256.Sum256(catalog)), Definitions: map[string]string{
		"scale":           "1 world pixel per output pixel; runtime 2x render reduced by opaque-only 2x2 area average; reference sampled through reference.json registration at 1x",
		"clean_cells":     "every nonexcluded terrain-family cell without infrastructure or building content; unit layers are absent from the render; standing terrain remains composed",
		"percentiles":     "linear interpolation on sorted opaque masked CIELAB D65 L* values at p*(n-1), p=.05/.95",
		"bands":           "RMS difference between square box means at radii 1 and 3, and 3 and 10 world pixels; every pixel of the larger window must be opaque and within the mask; these definitions are new, not imported from the absent look-audit evidence",
		"autocorrelation": "Pearson coefficient over paired opaque masked L* endpoints at each specified world lag; at least 64 pairs; unavailable for zero variance; no full-period exemptions",
		"acceptance":      "exact render targets use unrounded measurements; missing policies, estimator calibration, populations and coverage are INCOMPLETE; no period exemptions or substitute baselines; -look-check exits nonzero unless every required rule passes",
	}, Remaining: []string{
		"complete target/baseline policies and strict reference-check CI binding",
		"original screenshot registration and private numbers-only baseline measurements",
		"complete woodland-boundary coverage and independently annotated reference/original canopy coverage",
		"global painted light skew, non-crown cast shadows and reference/original light baselines",
		"water sampling across other features, independently identifiable shallow/shadow regions, original baselines and width acceptance",
		"unit contrast acceptance aggregation, additional facings/states and reference/original unit baselines",
		"resolve exact material period versus low full-period autocorrelation requirement",
	}}
	render1x := halve(render)
	renderSource := p.measureLookSource("render", render1x)
	water, err := p.measureWater("render", render1x, out)
	if err != nil {
		return err
	}
	renderSource.Water = &water
	report.Definitions["water"] = water.Definition
	woodland, err := p.measureWoodland("render", render1x, out)
	if err != nil {
		return err
	}
	renderSource.Woodland = &woodland
	report.Definitions["woodland_edge"] = woodland.Definition
	renderer, err := maprender.New(p.assets)
	if err != nil {
		return err
	}
	light, err := p.measureLight(renderer, compareDensity)
	if err != nil {
		return err
	}
	renderSource.Light = &light
	report.Definitions["cast_shadow_direction"] = light.Definition
	crowns, err := renderer.VisibleCanopy(p.level, compareDensity)
	if err != nil {
		return err
	}
	coverage, err := averageCanopy(crowns.Coverage, compareDensity)
	if err != nil {
		return err
	}
	canopy := p.measureCanopy(coverage, crowns.Unsupported)
	canopy.Gaps, canopy.Issues = len(crowns.Gaps), len(crowns.Issues)
	renderSource.Canopy = &canopy
	report.Definitions["canopy"] = canopy.Definition + "; diagnostic PNG rounds the exact alpha means to 8-bit grayscale"
	report.Sources = append(report.Sources, renderSource)
	src, present, err := p.source()
	if err != nil {
		return err
	}
	if present {
		reference1x := p.registered(src, 1)
		source := p.measureLookSource("reference", reference1x)
		water, err := p.measureWater("reference", reference1x, out)
		if err != nil {
			return err
		}
		source.Water = &water
		woodland, err := p.measureWoodland("reference", reference1x, out)
		if err != nil {
			return err
		}
		source.Woodland = &woodland
		source.Canopy = &lookCanopy{Reason: "independent reference canopy annotations unavailable"}
		source.Units = &lookUnits{Reason: "independent reference unit baseline unavailable"}
		source.Light = &lookLight{Reason: "independent reference object/shadow correspondences unavailable", GlobalReason: "independent reference light annotations unavailable"}
		source.FileSHA256 = p.ref.Image.SHA256
		report.Sources = append(report.Sources, source)
	} else {
		report.Sources = append(report.Sources, lookSource{ID: "reference", Reason: "reference image absent; no measurements substituted", Canopy: &lookCanopy{Reason: "source_unavailable"}})
	}
	report.Sources = append(report.Sources, lookSource{ID: "original", Reason: "no registered original screenshot baseline supplied in this checkout", Canopy: &lookCanopy{Reason: "source_unavailable"}})
	canopy.MaskFile = "canopy-render.png"
	canopy.MaskSHA256, err = writeCanopyPlane(out, canopy.MaskFile, coverage)
	if err != nil {
		return err
	}
	units, err := p.measureUnits(out)
	if err != nil {
		return err
	}
	report.Sources[0].Units = &units
	for i := range report.Sources {
		if report.Sources[i].Light == nil {
			report.Sources[i].Light = &lookLight{Reason: "source_unavailable", GlobalReason: "source_unavailable"}
		}
		if report.Sources[i].Woodland == nil {
			report.Sources[i].Woodland = &lookWoodland{Reason: "source_unavailable"}
		}
		if report.Sources[i].Water == nil {
			report.Sources[i].Water = &lookWater{Reason: "source_unavailable"}
		}
		if report.Sources[i].Units == nil {
			report.Sources[i].Units = &lookUnits{Reason: "source_unavailable"}
		}
	}
	report.Definitions["unit_edge_contrast"] = units.Definition
	gate := evaluateLookGate(report)
	report.Gate, report.Accepted = &gate, gate.Accepted
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "look.json"), append(raw, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("wrote %s/look.json; partial Look 0 evidence, acceptance remains open\n", out)
	return nil
}
