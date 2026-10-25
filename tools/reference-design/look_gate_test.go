package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func gateRule(t *testing.T, g lookGate, id string) lookGateRule {
	t.Helper()
	for _, r := range g.Rules {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("missing required rule %s", id)
	return lookGateRule{}
}

func gateFixture() lookReport {
	return lookReport{Sources: []lookSource{{ID: "render", Available: true, Whole: &lookMetrics{MeanL: lookMeasured(40, 100)}, Families: []lookFamily{{ID: "grass", Statistics: lookMetrics{Samples: 100, MeanL: lookMeasured(43, 100), SigmaL: lookMeasured(11, 100), EdgeDensity: lookMeasured(.3, 100), Autocorrelation: []lookLag{{Offset: [2]int{568, 0}, Pearson: lookMeasured(1, 100)}, {Offset: [2]int{142, 0}, Pearson: lookMeasured(-.5, 100)}}}}, {ID: "hills", Statistics: lookMetrics{Samples: 100, MeanL: lookMeasured(38, 100)}}}}}}
}

func TestLookGateChecksExactTargetsAndRetainsIncompleteRequirements(t *testing.T) {
	g := evaluateLookGate(gateFixture())
	for _, id := range []string{"render.whole_mean_l", "render.grass_mean_l", "render.grass_sigma_l", "render.grass_edge_density", "render.hills_delta_l", "render.grass.correlation.142.0"} {
		if r := gateRule(t, g, id); r.Status != "PASS" {
			t.Fatalf("exact target %s: %+v", id, r)
		}
	}
	if r := gateRule(t, g, "render.grass.correlation.568.0"); r.Status != "FAIL" {
		t.Fatal("full period lag silently exempted")
	}
	for _, id := range []string{"policy.material_periods", "policy.grass_band_estimator", "policy.value_tolerances", "render.unit_edge_contrast", "reference.baseline_policy", "original.baseline"} {
		if r := gateRule(t, g, id); r.Status != "INCOMPLETE" || r.Reason == "" {
			t.Fatalf("invented missing policy: %+v", r)
		}
	}
	if g.Accepted || g.Failed == 0 || g.Incomplete == 0 || g.Passed == 0 {
		t.Fatalf("false acceptance: %+v", g)
	}
}

func TestLookGateBoundsAreInclusiveAndMissingIsIncomplete(t *testing.T) {
	for _, v := range []float64{10, 13} {
		r := gateFixture()
		r.Sources[0].Families[0].Statistics.SigmaL = lookMeasured(v, 100)
		if g := gateRule(t, evaluateLookGate(r), "render.grass_sigma_l"); g.Status != "PASS" {
			t.Fatalf("inclusive boundary %g: %+v", v, g)
		}
	}
	for _, v := range []float64{9.99, 13.01} {
		r := gateFixture()
		r.Sources[0].Families[0].Statistics.SigmaL = lookMeasured(v, 100)
		if g := gateRule(t, evaluateLookGate(r), "render.grass_sigma_l"); g.Status != "FAIL" {
			t.Fatalf("outside boundary %g: %+v", v, g)
		}
	}
	r := gateFixture()
	r.Sources[0].Families[0].Statistics.SigmaL = lookValue{}
	g := evaluateLookGate(r)
	if gateRule(t, g, "render.grass_sigma_l").Status != "INCOMPLETE" {
		t.Fatal("missing sample became pass/fail")
	}
	r = gateFixture()
	r.Sources[0].Whole.MeanL.Value = new(float64)
	*r.Sources[0].Whole.MeanL.Value = math.NaN()
	if gateRule(t, evaluateLookGate(r), "render.whole_mean_l").Status != "INCOMPLETE" {
		t.Fatal("nonfinite value measured")
	}
}

func TestLookGateCannotPassPartialWoodlandOrAssetLightAsGlobal(t *testing.T) {
	r := gateFixture()
	v := .1
	m := 12.
	a := 135.
	e := 0.
	r.Sources[0].Woodland = &lookWoodland{Available: true, Statistics: woodlandDistanceStats{Median: &m, WithinSix: &v}}
	r.Sources[0].Light = &lookLight{Available: true, Instances: 76, Summary: lookLightSummary{Measured: 76, MeanBearing: &a, MaxTargetError: &e, WithinTen: 76}, GlobalReason: "painted light unavailable"}
	g := evaluateLookGate(r)
	if gateRule(t, g, "render.woodland_edge_fraction").Status != "INCOMPLETE" || gateRule(t, g, "render.global_light").Status != "INCOMPLETE" || gateRule(t, g, "render.separate_crown_shadow_direction").Status != "PASS" {
		t.Fatal("subscope became whole-map acceptance")
	}
}

func TestLookStrictCLIRequiresOutputAndPreservesFailedReport(t *testing.T) {
	if err := run(options{lookCheck: true}); err == nil || !strings.Contains(err.Error(), "-look-out") {
		t.Fatalf("strict flag ignored: %v", err)
	}
	o := comparisonOptions(t)
	o.compare = ""
	o.lookOut = filepath.Join(t.TempDir(), "strict")
	o.lookCheck = true
	err := run(o)
	if err == nil || !strings.Contains(err.Error(), "look 0 acceptance") {
		t.Fatalf("incomplete strict check passed: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(o.lookOut, "look.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r lookReport
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if r.Accepted || r.Gate == nil || r.Gate.Incomplete == 0 {
		t.Fatal("missing failing report")
	}
}
