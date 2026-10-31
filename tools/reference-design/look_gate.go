package referencedesign

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

// Rules fail closed: unknown populations, estimators, sources and policies are
// incomplete, never successes. Reference/original do not inherit remake targets.
type lookGate struct {
	Accepted   bool           `json:"accepted"`
	Passed     int            `json:"passed"`
	Failed     int            `json:"failed"`
	Incomplete int            `json:"incomplete"`
	Rules      []lookGateRule `json:"rules"`
}

type lookGateRule struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Source   string   `json:"source"`
	Value    *float64 `json:"value"`
	Target   string   `json:"target"`
	Scope    string   `json:"scope"`
	Reason   string   `json:"reason,omitempty"`
	Citation string   `json:"requirement"`
}

func evaluateLookGate(report lookReport) lookGate {
	gate := lookGate{Rules: []lookGateRule{}}
	add := func(id, source, status, target, scope, reason, citation string, value *float64) {
		if value != nil && !isFinite(*value) {
			value = nil
			status = "INCOMPLETE"
			reason = "nonfinite measurement"
		}
		gate.Rules = append(gate.Rules, lookGateRule{id, status, source, value, target, scope, reason, citation})
	}
	pending := func(id, source, target, reason, citation string) {
		add(id, source, "INCOMPLETE", target, "required acceptance coverage", reason, citation, nil)
	}
	numeric := func(id string, v lookValue, lo, hi float64, strictHi bool, target, scope, citation string) {
		if v.Value == nil || v.Samples <= 0 || v.Reason != "" {
			add(id, "render", "INCOMPLETE", target, scope, "measurement unavailable: "+v.Reason, citation, nil)
			return
		}
		status := "PASS"
		if *v.Value < lo || *v.Value > hi || (strictHi && *v.Value == hi) {
			status = "FAIL"
		}
		add(id, "render", status, target, scope, "", citation, v.Value)
	}
	var render *lookSource
	for i := range report.Sources {
		if report.Sources[i].ID == "render" && report.Sources[i].Available {
			render = &report.Sources[i]
		}
	}
	families := map[string]lookFamily{}
	if render != nil {
		for _, f := range render.Families {
			families[f.ID] = f
		}
	}
	whole := lookValue{}
	if render != nil && render.Whole != nil {
		whole = render.Whole.MeanL
	}
	numeric("render.whole_mean_l", whole, 39, 42, false, "39 <= mean L* <= 42", "all opaque nonexcluded map cells", "art/direction.md:81")
	grass := families["grass"].Statistics
	numeric("render.grass_mean_l", grass.MeanL, 42, 45, false, "42 <= mean L* <= 45", "all clean grass pixels", "art/direction.md:82")
	numeric("render.grass_sigma_l", grass.SigmaL, 10, 13, false, "10 <= population sigma L* <= 13", "all clean grass pixels; unrounded", "art/direction.md:581")
	numeric("render.grass_edge_density", grass.EdgeDensity, .28, .40, false, "0.28 <= edge density <= 0.40", "clean grass pixels with opaque finite 3x3 Sobel support; gradient >6", "art/direction.md:581")
	numeric("render.forest_mean_l", families["forest"].Statistics.MeanL, 26, 31, false, "26 <= mean L* <= 31", "all clean forest pixels", "art/direction.md:84")
	for _, spec := range []struct {
		id     string
		lo, hi float64
	}{{"hills", -6, -4}, {"mountain", -10, -6}, {"mount", -10, -6}} {
		v := lookValue{}
		source := families[spec.id].Statistics.MeanL
		if grass.MeanL.Value != nil && source.Value != nil && grass.MeanL.Reason == "" && source.Reason == "" {
			v = lookMeasured(*source.Value-*grass.MeanL.Value, min(source.Samples, grass.MeanL.Samples))
		}
		numeric("render."+spec.id+"_delta_l", v, spec.lo, spec.hi, false, fmt.Sprintf("%g <= family mean L* - grass mean L* <= %g", spec.lo, spec.hi), "all clean pixels in both families", "art/direction.md:83-85")
	}
	// Keep every literal signed lag, including full material periods. The separate
	// unresolved-policy rule prevents a signed subtest from resolving intent.
	for _, f := range report.Sources {
		if f.ID != "render" || !f.Available {
			continue
		}
		for _, family := range f.Families {
			byLag := map[[2]int]lookValue{}
			for _, lag := range family.Statistics.Autocorrelation {
				byLag[lag.Offset] = lag.Pearson
			}
			for _, lag := range [][2]int{{71, 40}, {142, 0}, {0, 80}, {0, 160}, {284, 0}, {568, 0}} {
				v := byLag[lag]
				if v.Samples < lookMinPairs {
					v = lookValue{Reason: "insufficient paired samples"}
				}
				numeric(fmt.Sprintf("render.%s.correlation.%d.%d", family.ID, lag[0], lag[1]), v, -1, .15, true, "literal signed Pearson r < 0.15; no period exemption", "all clean paired family endpoints", "art/direction.md:580")
			}
		}
	}
	if render == nil {
		pending("render.lattice_correlations", "render", "all six lags for each family", "render source unavailable", "art/direction.md:580")
	}
	woodland := (*lookWoodland)(nil)
	light := (*lookLight)(nil)
	if render != nil {
		woodland = render.Woodland
		light = render.Light
	}
	fraction, median := lookValue{}, lookValue{}
	if woodland != nil && woodland.Available && woodland.Complete {
		fraction = lookValue{Value: woodland.Statistics.WithinSix, Samples: woodland.Statistics.Samples}
		median = lookValue{Value: woodland.Statistics.Median, Samples: woodland.Statistics.Samples}
	}
	numeric("render.woodland_edge_fraction", fraction, 0, .25, false, "fraction within 6 world px <= 0.25", "complete visible woodland boundary; fragments insufficient", "art/direction.md:583")
	numeric("render.woodland_edge_median", median, 10, math.Inf(1), false, "median distance >= 10 world px", "complete visible woodland boundary; fragments insufficient", "art/direction.md:583")
	direction := lookValue{}
	if light != nil && light.Available && light.Gaps == 0 && light.Issues == 0 && light.Summary.Measured == light.Instances && light.Instances > 0 {
		direction = lookValue{Value: light.Summary.MaxTargetError, Samples: light.Summary.Measured}
	}
	numeric("render.separate_crown_shadow_direction", direction, 0, 10, false, "every selected measurable crown/shadow pair within 135 +/-10 degrees", "complete unoccluded paired source alpha only; not painted/global illumination", "art/direction.md:580")
	pending("render.global_light", "render", "one global shadow azimuth 135 +/-10 degrees", "embedded shadows and painted form shading unmeasured", "art/direction.md:580")
	pending("policy.material_periods", "render", "all six correlations <0.15 and declared material periods", "exact 568x400 meadow/hills and 284x160 water repeat at tested lags; interpretation and signed-versus-magnitude policy unresolved", "PLAN.md Look 1; art/direction.md:580")
	pending("policy.grass_band_estimator", "render", "both bands >=3.8; 3-10 band capped about4.0", "box-radius RMS is explicitly a new estimator; audit correspondence and approximate upper tolerance unspecified", "art/direction.md:82,581")
	pending("policy.value_tolerances", "render", "P05/P95 about8/64; building b* about16; land b*4-5 lower", "approximate tolerances and pinned pre-grade colour baseline unspecified", "art/direction.md:81,88")
	pending("render.canopy_population", "render", "forest coverage >=85%; open coverage4-8%", "per-cell versus pooled acceptance population unspecified; unsupported and complete-map/reference coverage remain open", "art/direction.md:583")
	pending("render.forest_structure", "render", "crown size16-30 median20-24;9-13 crowns/interior hex;prominence>=35;gaps<=18;tops<=58;no identical hexes", "crown-instance/form-shading metrics unavailable", "art/direction.md:84,583")
	pending("render.water_appearance_and_geometry", "render", "deep/shallow colour;depth/shadow;6-10px ripples;bank median8-13 CV>=.35;stream median22-28 CV.10-.20;234shorepairs/24mouths", "annotated subset and uncertainty do not cover full populations; physical shallows/shadow regions and joins acceptance incomplete", "art/direction.md:87,584")
	pending("render.unit_edge_contrast", "render", "unit edge delta L*>=30 ongrass/forest/road/water;contactshadow", "acceptance aggregation and all-facing/state population unspecified", "art/direction.md:589")
	pending("render.lines_buildings_relief_winter", "render", "remaining road/path/rail/trench/building/relief/winter rows and view checks", "required widths,geometry,markers,seasonal bindings and view/performance evidence not yet evaluated by this gate", "art/direction.md:585-590; view checks")
	pending("render.hills_identification", "render", "18 of20 random hills named correctly at1x;hills 3-10band>=1.4xgrass", "blind review and calibrated band-estimator evidence missing", "art/direction.md:83,582")
	pending("reference.baseline_policy", "reference", "independent declared reference ranges and comparisons", "reference painting has its own measurements; remake targets cannot silently become reference acceptance policy", "PLAN.md Look 0")
	pending("original.baseline", "original", "authenticated original screenshot registration and declared baseline ranges", "registered authenticated baseline/policy unavailable; legacy imagery is not substituted", "PLAN.md Look 0")
	for _, r := range gate.Rules {
		switch r.Status {
		case "PASS":
			gate.Passed++
		case "FAIL":
			gate.Failed++
		default:
			gate.Incomplete++
		}
	}
	gate.Accepted = len(gate.Rules) > 0 && gate.Failed == 0 && gate.Incomplete == 0
	return gate
}

func checkWrittenLookGate(out string, stdout io.Writer) error {
	raw, err := os.ReadFile(filepath.Join(out, "look.json"))
	if err != nil {
		return err
	}
	var report lookReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return err
	}
	if report.Gate == nil {
		return fmt.Errorf("look 0 acceptance: gate unavailable")
	}
	gate := report.Gate
	if _, err := fmt.Fprintf(stdout, "Look 0: %d PASS, %d FAIL, %d INCOMPLETE\n", gate.Passed, gate.Failed, gate.Incomplete); err != nil {
		return err
	}
	if !gate.Accepted || !report.Accepted {
		return fmt.Errorf("look 0 acceptance: %d failed, %d incomplete (details in %s/look.json)", gate.Failed, gate.Incomplete, out)
	}
	return nil
}
