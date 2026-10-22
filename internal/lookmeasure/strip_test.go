package lookmeasure_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/lookmeasure"
)

func TestStripPixelsGeometry(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		width, height         int
		from, to              [2]float64
		start, end, halfWidth float64
		want                  []int
	}{
		{
			name: "horizontal half-open endpoints and sides", width: 6, height: 5,
			from: [2]float64{0, 2}, to: [2]float64{6, 2}, start: 1.5, end: 4.5, halfWidth: 0.5,
			want: []int{7, 8, 9},
		},
		{
			name: "vertical negative side included", width: 6, height: 5,
			from: [2]float64{3, 0}, to: [2]float64{3, 5}, start: 1.5, end: 4.5, halfWidth: 0.5,
			want: []int{9, 15, 21},
		},
		{
			name: "reversed direction flips half-open sides", width: 6, height: 5,
			from: [2]float64{6, 2}, to: [2]float64{0, 2}, start: 1.5, end: 4.5, halfWidth: 0.5,
			want: []int{14, 15, 16},
		},
		{
			name: "diagonal", width: 6, height: 6,
			from: [2]float64{1, 1}, to: [2]float64{5, 5}, start: 0, end: 4 * math.Sqrt2, halfWidth: 0.25,
			want: []int{7, 14, 21, 28},
		},
		{
			name: "oblique three four five direction", width: 6, height: 6,
			from: [2]float64{1, 1}, to: [2]float64{4, 5}, start: 0, end: 5, halfWidth: 0.4,
			want: []int{7, 14, 20, 27},
		},
		{
			name: "rectangle may meet canvas boundaries", width: 4, height: 2,
			from: [2]float64{0, 1}, to: [2]float64{4, 1}, start: 0, end: 4, halfWidth: 1,
			want: []int{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "only selected strip must be inside canvas", width: 6, height: 5,
			from: [2]float64{-1, 2}, to: [2]float64{7, 2}, start: 2, end: 6, halfWidth: 0.5,
			want: []int{7, 8, 9, 10},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lookmeasure.StripPixels(tc.width, tc.height, tc.from, tc.to, tc.start, tc.end, tc.halfWidth)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("pixels = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStripPixelsSubpixelStripCanBeEmpty(t *testing.T) {
	got, err := lookmeasure.StripPixels(6, 5, [2]float64{1, 2}, [2]float64{4, 2}, 0, 3, 0.1)
	if err != nil || len(got) != 0 {
		t.Fatalf("subpixel strip = %v, %v; want empty pixels without error", got, err)
	}
}

func TestStripPixelsRejectsInvalidGeometry(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		width, height         int
		from, to              [2]float64
		start, end, halfWidth float64
	}{
		{"zero width", 0, 5, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, 0.5},
		{"negative width", -6, 5, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, 0.5},
		{"zero height", 6, 0, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, 0.5},
		{"negative height", 6, -5, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, 0.5},
		{"overflow dimensions", int(^uint(0) >> 1), 2, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, 0.5},
		{"zero segment", 6, 5, [2]float64{1, 2}, [2]float64{1, 2}, 0, 1, 0.5},
		{"overflow direction", 6, 5, [2]float64{-math.MaxFloat64, 2}, [2]float64{math.MaxFloat64, 2}, 0, 4, 0.5},
		{"overflow length", 6, 5, [2]float64{0, 0}, [2]float64{math.MaxFloat64, math.MaxFloat64}, 0, 4, 0.5},
		{"negative start", 6, 5, [2]float64{1, 2}, [2]float64{5, 2}, -1, 4, 0.5},
		{"equal interval", 6, 5, [2]float64{1, 2}, [2]float64{5, 2}, 2, 2, 0.5},
		{"reversed interval", 6, 5, [2]float64{1, 2}, [2]float64{5, 2}, 3, 2, 0.5},
		{"past segment", 6, 5, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4.01, 0.5},
		{"zero half width", 6, 5, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, 0},
		{"negative half width", 6, 5, [2]float64{1, 2}, [2]float64{5, 2}, 0, 4, -0.5},
		{"left corner clipped", 6, 5, [2]float64{0, 1}, [2]float64{1, 4}, 0, math.Sqrt(10), 0.5},
		{"right corner clipped", 6, 5, [2]float64{6, 1}, [2]float64{5, 4}, 0, math.Sqrt(10), 0.5},
		{"top corner clipped", 6, 5, [2]float64{1, 0.25}, [2]float64{5, 0.25}, 0, 4, 0.5},
		{"bottom corner clipped", 6, 5, [2]float64{1, 4.75}, [2]float64{5, 4.75}, 0, 4, 0.5},
		{"end clipped", 6, 5, [2]float64{1, 2}, [2]float64{7, 2}, 0, 6, 0.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lookmeasure.StripPixels(tc.width, tc.height, tc.from, tc.to, tc.start, tc.end, tc.halfWidth)
			if err == nil || len(got) != 0 {
				t.Fatalf("invalid geometry = %v, %v; want error without samples", got, err)
			}
		})
	}
}

func TestStripPixelsRejectsNonfiniteInputs(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(-1), math.Inf(1)} {
		for field := range 7 {
			params := [7]float64{1, 2, 5, 2, 0, 4, 0.5}
			params[field] = value
			got, err := lookmeasure.StripPixels(6, 5, [2]float64{params[0], params[1]}, [2]float64{params[2], params[3]}, params[4], params[5], params[6])
			if err == nil || len(got) != 0 {
				t.Errorf("nonfinite field %d (%v) = %v, %v; want error without samples", field, value, got, err)
			}
		}
	}
}

func TestStripPixelsSmallStripOnLargeCanvas(t *testing.T) {
	// These dimensions describe 10^12 pixels, but only the strip's small box
	// needs visiting. A full-canvas allocation or scan is unnecessary.
	got, err := lookmeasure.StripPixels(1_000_000, 1_000_000, [2]float64{1, 2}, [2]float64{4, 2}, 0, 3, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{1_000_001, 1_000_002, 1_000_003}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pixels = %v, want %v", got, want)
	}
}
