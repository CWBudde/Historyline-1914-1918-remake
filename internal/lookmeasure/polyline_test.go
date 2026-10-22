package lookmeasure_test

import (
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/lookmeasure"
)

func TestPolylineDistancesAnalytic(t *testing.T) {
	for _, tc := range []struct {
		name    string
		points  []lookmeasure.Point
		edges   []lookmeasure.Segment
		maxStep float64
		want    []lookmeasure.EdgeDistance
	}{
		{
			name:    "parallel equal pieces below maximum step",
			points:  []lookmeasure.Point{{X: 0, Y: 3}, {X: 4, Y: 3}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: -1, Y: 0}, To: lookmeasure.Point{X: 5, Y: 0}}},
			maxStep: 1.5,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: 2.0 / 3, Y: 3}, Distance: 3, Weight: 4.0 / 3},
				{Point: lookmeasure.Point{X: 2, Y: 3}, Distance: 3, Weight: 4.0 / 3},
				{Point: lookmeasure.Point{X: 10.0 / 3, Y: 3}, Distance: 3, Weight: 4.0 / 3},
			},
		},
		{
			name:    "crossing reference",
			points:  []lookmeasure.Point{{X: -2, Y: 0}, {X: 2, Y: 0}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: -1}, To: lookmeasure.Point{X: 0, Y: 1}}},
			maxStep: 1,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: -1.5, Y: 0}, Distance: 1.5, Weight: 1},
				{Point: lookmeasure.Point{X: -0.5, Y: 0}, Distance: 0.5, Weight: 1},
				{Point: lookmeasure.Point{X: 0.5, Y: 0}, Distance: 0.5, Weight: 1},
				{Point: lookmeasure.Point{X: 1.5, Y: 0}, Distance: 1.5, Weight: 1},
			},
		},
		{
			name:    "finite reference endpoint distances",
			points:  []lookmeasure.Point{{X: -2, Y: 2}, {X: 4, Y: 2}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: 0}, To: lookmeasure.Point{X: 2, Y: 0}}},
			maxStep: 2,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: -1, Y: 2}, Distance: math.Sqrt(5), Weight: 2},
				{Point: lookmeasure.Point{X: 1, Y: 2}, Distance: 2, Weight: 2},
				{Point: lookmeasure.Point{X: 3, Y: 2}, Distance: math.Sqrt(5), Weight: 2},
			},
		},
		{
			name:   "nearest of multiple finite edges",
			points: []lookmeasure.Point{{X: 0, Y: 1}, {X: 4, Y: 1}},
			edges: []lookmeasure.Segment{
				{From: lookmeasure.Point{X: -1, Y: 0}, To: lookmeasure.Point{X: 1, Y: 0}},
				{From: lookmeasure.Point{X: 3, Y: 2}, To: lookmeasure.Point{X: 5, Y: 2}},
			},
			maxStep: 1,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: 0.5, Y: 1}, Distance: 1, Weight: 1},
				{Point: lookmeasure.Point{X: 1.5, Y: 1}, Distance: math.Sqrt(1.25), Weight: 1},
				{Point: lookmeasure.Point{X: 2.5, Y: 1}, Distance: math.Sqrt(1.25), Weight: 1},
				{Point: lookmeasure.Point{X: 3.5, Y: 1}, Distance: 1, Weight: 1},
			},
		},
		{
			name:    "open path segment order without vertex samples",
			points:  []lookmeasure.Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: -1, Y: -2}, To: lookmeasure.Point{X: 3, Y: -2}}},
			maxStep: 1,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: 0.5, Y: 0}, Distance: 2, Weight: 1},
				{Point: lookmeasure.Point{X: 1.5, Y: 0}, Distance: 2, Weight: 1},
				{Point: lookmeasure.Point{X: 2, Y: 0.5}, Distance: 2.5, Weight: 1},
				{Point: lookmeasure.Point{X: 2, Y: 1.5}, Distance: 3.5, Weight: 1},
			},
		},
		{
			name:    "diagonal Euclidean arclength weights",
			points:  []lookmeasure.Point{{X: 0, Y: 0}, {X: 3, Y: 4}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: -2}, To: lookmeasure.Point{X: 3, Y: -2}}},
			maxStep: 2,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: 0.5, Y: 2.0 / 3}, Distance: 8.0 / 3, Weight: 5.0 / 3},
				{Point: lookmeasure.Point{X: 1.5, Y: 2}, Distance: 4, Weight: 5.0 / 3},
				{Point: lookmeasure.Point{X: 2.5, Y: 10.0 / 3}, Distance: 16.0 / 3, Weight: 5.0 / 3},
			},
		},
		{
			name:    "one midpoint for short segment",
			points:  []lookmeasure.Point{{X: 0, Y: 3}, {X: 4, Y: 3}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: 0}, To: lookmeasure.Point{X: 4, Y: 0}}},
			maxStep: 10,
			want:    []lookmeasure.EdgeDistance{{Point: lookmeasure.Point{X: 2, Y: 3}, Distance: 3, Weight: 4}},
		},
		{
			name:    "large finite segment without squared length overflow",
			points:  []lookmeasure.Point{{X: 0, Y: 2e200}, {X: 4e200, Y: 2e200}},
			edges:   []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: 0}, To: lookmeasure.Point{X: 4e200, Y: 0}}},
			maxStep: 2e200,
			want: []lookmeasure.EdgeDistance{
				{Point: lookmeasure.Point{X: 1e200, Y: 2e200}, Distance: 2e200, Weight: 2e200},
				{Point: lookmeasure.Point{X: 3e200, Y: 2e200}, Distance: 2e200, Weight: 2e200},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lookmeasure.PolylineDistances(tc.points, tc.edges, tc.maxStep)
			if err != nil {
				t.Fatal(err)
			}
			assertEdgeDistances(t, got, tc.want)
		})
	}
}

func TestPolylineDistancesReversedPath(t *testing.T) {
	edges := []lookmeasure.Segment{{From: lookmeasure.Point{X: -1, Y: -2}, To: lookmeasure.Point{X: 3, Y: -2}}}
	got, err := lookmeasure.PolylineDistances([]lookmeasure.Point{{X: 2, Y: 2}, {X: 2, Y: 0}, {X: 0, Y: 0}}, edges, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertEdgeDistances(t, got, []lookmeasure.EdgeDistance{
		{Point: lookmeasure.Point{X: 2, Y: 1.5}, Distance: 3.5, Weight: 1},
		{Point: lookmeasure.Point{X: 2, Y: 0.5}, Distance: 2.5, Weight: 1},
		{Point: lookmeasure.Point{X: 1.5, Y: 0}, Distance: 2, Weight: 1},
		{Point: lookmeasure.Point{X: 0.5, Y: 0}, Distance: 2, Weight: 1},
	})
}

func TestPolylineDistancesResegmentationPreservesLengthWeights(t *testing.T) {
	edges := []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: 0}, To: lookmeasure.Point{X: 4, Y: 0}}}
	for _, points := range [][]lookmeasure.Point{
		{{X: 0, Y: 3}, {X: 4, Y: 3}},
		{{X: 0, Y: 3}, {X: 2, Y: 3}, {X: 4, Y: 3}},
		{{X: 0, Y: 3}, {X: 1.3, Y: 3}, {X: 4, Y: 3}},
	} {
		got, err := lookmeasure.PolylineDistances(points, edges, 1)
		if err != nil {
			t.Fatal(err)
		}
		var weight float64
		for _, sample := range got {
			if !nearPolylineFloat(sample.Distance, 3) || sample.Weight <= 0 || sample.Weight > 1 {
				t.Errorf("sample = %+v; want distance 3 and weight in (0,1]", sample)
			}
			weight += sample.Weight
		}
		if !nearPolylineFloat(weight, 4) {
			t.Errorf("total weight = %v, want polyline length 4", weight)
		}
	}
}

func TestPolylineDistancesRejectsInvalidGeometry(t *testing.T) {
	points := []lookmeasure.Point{{X: 0, Y: 2}, {X: 4, Y: 2}}
	edges := []lookmeasure.Segment{{From: lookmeasure.Point{X: 0, Y: 0}, To: lookmeasure.Point{X: 4, Y: 0}}}
	for _, tc := range []struct {
		name    string
		points  []lookmeasure.Point
		edges   []lookmeasure.Segment
		maxStep float64
	}{
		{"no points", nil, edges, 1},
		{"one point", points[:1], edges, 1},
		{"adjacent duplicate", []lookmeasure.Point{{X: 0, Y: 2}, {X: 4, Y: 2}, {X: 4, Y: 2}}, edges, 1},
		{"no edges", points, nil, 1},
		{"degenerate reference", points, []lookmeasure.Segment{edges[0], {From: points[0], To: points[0]}}, 1},
		{"zero step", points, edges, 0},
		{"negative step", points, edges, -1},
		{"nonfinite count", points, edges, math.SmallestNonzeroFloat64},
		{"count beyond int", []lookmeasure.Point{{X: 0, Y: 2}, {X: 1e20, Y: 2}}, edges, 1},
		{"sample allocation byte overflow", []lookmeasure.Point{{X: 0, Y: 2}, {X: 1e18, Y: 2}}, edges, 1},
		{"total sample allocation overflow", []lookmeasure.Point{{X: 0, Y: 2}, {X: 2e17, Y: 2}, {X: 0, Y: 2}}, edges, 1},
		{"source delta overflow", []lookmeasure.Point{{X: -math.MaxFloat64}, {X: math.MaxFloat64}}, edges, 1},
		{"source length overflow", []lookmeasure.Point{{}, {X: math.MaxFloat64, Y: math.MaxFloat64}}, edges, 1},
		{"total length overflow", []lookmeasure.Point{{}, {X: math.MaxFloat64}, {}}, edges, math.MaxFloat64},
		{"reference delta overflow", points, []lookmeasure.Segment{{From: lookmeasure.Point{X: -math.MaxFloat64}, To: lookmeasure.Point{X: math.MaxFloat64}}}, 1},
		{"reference length overflow", points, []lookmeasure.Segment{{To: lookmeasure.Point{X: math.MaxFloat64, Y: math.MaxFloat64}}}, 1},
		{"sample to reference delta overflow", []lookmeasure.Point{{X: math.MaxFloat64, Y: 0}, {X: math.MaxFloat64, Y: 1}}, []lookmeasure.Segment{{From: lookmeasure.Point{X: -math.MaxFloat64, Y: 0}, To: lookmeasure.Point{X: -math.MaxFloat64, Y: 1}}}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lookmeasure.PolylineDistances(tc.points, tc.edges, tc.maxStep)
			if err == nil || got != nil {
				t.Fatalf("invalid geometry = %v, %v; want nil samples and error", got, err)
			}
		})
	}
}

func TestPolylineDistancesRejectsNonfiniteInputs(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(-1), math.Inf(1)} {
		for field := range 9 {
			params := [9]float64{0, 2, 4, 2, 0, 0, 4, 0, 1}
			params[field] = value
			got, err := lookmeasure.PolylineDistances(
				[]lookmeasure.Point{{X: params[0], Y: params[1]}, {X: params[2], Y: params[3]}},
				[]lookmeasure.Segment{{From: lookmeasure.Point{X: params[4], Y: params[5]}, To: lookmeasure.Point{X: params[6], Y: params[7]}}}, params[8])
			if err == nil || got != nil {
				t.Errorf("nonfinite field %d (%v) = %v, %v; want nil samples and error", field, value, got, err)
			}
		}
	}
}

func assertEdgeDistances(t *testing.T, got, want []lookmeasure.EdgeDistance) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("samples = %+v, want %+v", got, want)
	}
	for i := range want {
		if !nearPolylineFloat(got[i].Point.X, want[i].Point.X) || !nearPolylineFloat(got[i].Point.Y, want[i].Point.Y) || !nearPolylineFloat(got[i].Distance, want[i].Distance) || !nearPolylineFloat(got[i].Weight, want[i].Weight) {
			t.Errorf("sample %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func nearPolylineFloat(got, want float64) bool {
	return math.Abs(got-want) <= 1e-12*math.Max(1, math.Abs(want))
}
