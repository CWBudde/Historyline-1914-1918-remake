package referencedesign

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

func lightPair(at hexgrid.Coord, angle float64, hash string) maprender.ShadowPair {
	return maprender.ShadowPair{At: at, Standing: maprender.Entry{ID: "standing/forest", SHA: "crown-" + hash, Anchor: [2]float64{64, 56}}, Shadow: maprender.Entry{ID: "standing/shadow/forest", SHA: "shadow-" + hash, Anchor: [2]float64{64, 56}}, StandingCentroid: [2]float64{10, 20}, ShadowCentroid: [2]float64{13, 24}, Offset: [2]float64{3, 4}, Distance: 5, Bearing: &angle}
}

func TestLightReportFiltersRowsAndKeepsDuplicateEvidenceExplicit(t *testing.T) {
	p := &project{excluded: map[int]bool{3: true}, level: &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}, {Terrain: "forest", Infra: level.Infra{IDs: []string{"path36"}}}}, {{Terrain: "forest"}}, {{Terrain: "forest"}}}}}
	pairs := maprender.ShadowPairs{Density: 2, Pairs: []maprender.ShadowPair{lightPair(hexgrid.Coord{}, 135, "a"), lightPair(hexgrid.Coord{Col: 1}, 135, "a"), lightPair(hexgrid.Coord{Line: 1}, 140, "b"), lightPair(hexgrid.Coord{Line: 2}, 20, "c")}}
	got := p.lightFromPairs(pairs)
	if !got.Available || got.GlobalAvailable || got.GlobalReason == "" || got.AllSelected != 4 || got.Excluded != 1 || got.Instances != 3 || got.Clean != 2 || got.UniquePairs != 2 || len(got.Pairs) != 3 {
		t.Fatalf("scope/counters: %+v", got)
	}
	if got.Summary.Measured != 3 || got.UniqueSummary.Measured != 2 || got.Summary.WithinTen != 3 {
		t.Fatalf("duplicate bias not explicit: %+v %+v", got.Summary, got.UniqueSummary)
	}
	if got.Pairs[1].Clean || !got.Pairs[0].Clean || got.Pairs[0].Cell != "r1c1" {
		t.Fatal("clean cell and coordinate provenance")
	}
}

func TestLightCircularMeanAndWrappedError(t *testing.T) {
	a, b := 359., 1.
	pairs := []lookLightPair{{Bearing: &a}, {Bearing: &b}}
	got := summarizeLight(pairs)
	if got.MeanBearing == nil || math.Min(math.Abs(*got.MeanBearing), math.Abs(*got.MeanBearing-360)) > 1e-10 {
		t.Fatalf("mean across north: %+v", got)
	}
	if math.Abs(lightAngleDifference(359, 1)-2) > 1e-12 || math.Abs(lightAngleDifference(1, 359)-2) > 1e-12 {
		t.Fatal("wrapped angle error")
	}
	a, b = 0, 180
	got = summarizeLight(pairs)
	if got.MeanBearing != nil || got.Reason == "" {
		t.Fatal("cancelled circular directions must not invent mean")
	}
}

func TestLightUnavailableDoesNotBecomeZeroMeasurement(t *testing.T) {
	p := &project{excluded: map[int]bool{}, level: &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}}}}}
	pair := lightPair(hexgrid.Coord{}, 135, "a")
	pair.Reason = "empty_shadow_alpha"
	pair.Bearing = nil
	got := p.lightFromPairs(maprender.ShadowPairs{Density: 2, Pairs: []maprender.ShadowPair{pair}})
	if got.Available || got.Reason == "" || got.Summary.MeanBearing != nil || got.Summary.Measured != 0 || len(got.Pairs) != 1 || got.Pairs[0].Offset != nil || got.Pairs[0].Distance != nil || got.Pairs[0].StandingCentroid != nil || got.Pairs[0].ShadowCentroid != nil {
		t.Fatalf("unavailable pair became numeric evidence: %+v", got)
	}
}

func TestLightPulseSelectedPairsAndDensityAgreement(t *testing.T) {
	p, err := load("../..", "../../art/reference/pulse/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	var previous lookLight
	for _, density := range []int{2, 4} {
		t.Run(fmt.Sprint(density), func(t *testing.T) {
			got, err := p.measureLight(r, density)
			if err != nil {
				t.Fatal(err)
			}
			if got.AllSelected != 1898 || got.Excluded != 117 || got.Instances != 1781 || got.Clean != 1779 || got.Gaps != 0 || got.Issues != 0 || got.GlobalAvailable || got.GlobalReason == "" {
				t.Fatalf("runtime selected scope: %+v", got)
			}
			// Nine scattered crowns and the two baked contexts of forest with
			// infrastructure.
			if len(got.Pairs) != 1781 || got.UniquePairs != 11 {
				t.Fatalf("duplicate context accounting: %+v", got)
			}
			for i, pair := range got.Pairs {
				if len(pair.Standing.SHA) != 64 || pair.Standing.Density != density || pair.Shadow.ID == "" || pair.Cell == "" {
					t.Fatalf("missing pair provenance: %+v", pair)
				}
				if pair.Reason == "" {
					if pair.Bearing == nil || pair.Offset == nil || pair.Distance == nil || len(pair.Shadow.SHA) != 64 {
						t.Fatalf("missing measured values: %+v", pair)
					}
				} else if pair.Bearing != nil || pair.Offset != nil || pair.Distance != nil {
					t.Fatal("unavailable pair has values")
				}
				if density == 4 && pair.Reason == "" && previous.Pairs[i].Reason == "" {
					dx, dy := pair.Offset[0]-previous.Pairs[i].Offset[0], pair.Offset[1]-previous.Pairs[i].Offset[1]
					if pair.Site != ([2]float64{}) {
						// A scattered crown and its shadow have their own anchors, each
						// rounded to the density grid: up to 1/density apart per axis.
						if math.Abs(dx) > 1./2+1./4+.05 || math.Abs(dy) > 1./2+1./4+.05 {
							t.Fatalf("density drift at %s %v: %v vs %v", pair.Cell, pair.Site, pair.Offset, previous.Pairs[i].Offset)
						}
					} else if lightAngleDifference(*pair.Bearing, *previous.Pairs[i].Bearing) > .5 || math.Hypot(dx, dy) > .05 {
						t.Fatalf("density drift at %s: %v vs %v", pair.Cell, pair.Offset, previous.Pairs[i].Offset)
					}
				}
			}
			if s := got.Summary; s.Measured != 1781 || s.WithinTen != 1781 || lightAngleDifference(*s.MeanBearing, 135) > 1 {
				t.Fatalf("crown shadows leave the 135 degree light: %+v", s)
			}
			if density == 4 && lightAngleDifference(*got.Summary.MeanBearing, *previous.Summary.MeanBearing) > .5 {
				t.Fatalf("mean bearing drift: %.2f vs %.2f", *got.Summary.MeanBearing, *previous.Summary.MeanBearing)
			}
			previous = got
		})
	}
}
