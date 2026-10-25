package lookmeasure_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/lookmeasure"
)

func TestEdgeContrastSameCoordinateContribution(t *testing.T) {
	for _, tc := range []struct {
		name      string
		ground    float64
		composite float64
		want      float64
	}{
		{"lighter", 20, 65, 45},
		{"darker", 85, 25, 60},
		{"fractional", 23.25, 44.75, 21.5},
		{"full range", 0, 100, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ground, composite, body, _ := contrastPlanes(3, 3)
			ground[4], composite[4], body[4] = tc.ground, tc.composite, true
			got, err := lookmeasure.EdgeContrast(ground, composite, body, nil, 3, 3)
			if err != nil {
				t.Fatal(err)
			}
			want := []lookmeasure.EdgeSample{{X: 1, Y: 1, Delta: tc.want}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("samples = %+v, want %+v", got, want)
			}
		})
	}
}

func TestEdgeContrastIdenticalNoisyGroundIsZero(t *testing.T) {
	ground := []float64{0, 100, 7, 99, 21, 85, 43, 2, 68}
	body := []bool{false, false, false, false, true, false, false, false, false}
	got, err := lookmeasure.EdgeContrast(ground, ground, body, nil, 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []lookmeasure.EdgeSample{{X: 1, Y: 1, Delta: 0}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("samples = %+v, want zero contribution despite terrain noise", got)
	}
}

func TestEdgeContrastFourConnectedBoundaryWithHole(t *testing.T) {
	ground, composite, body, _ := contrastPlanes(7, 7)
	for y := 1; y <= 5; y++ {
		for x := 1; x <= 5; x++ {
			body[y*7+x] = true
			composite[y*7+x] = 70
		}
	}
	body[3*7+3] = false
	got, err := lookmeasure.EdgeContrast(ground, composite, body, nil, 7, 7)
	if err != nil {
		t.Fatal(err)
	}
	// The four pixels directly beside the hole join the outer perimeter.
	// Its diagonal neighbors stay interior, and each corner is counted once.
	want := []lookmeasure.EdgeSample{
		{X: 1, Y: 1, Delta: 50},
		{X: 2, Y: 1, Delta: 50},
		{X: 3, Y: 1, Delta: 50},
		{X: 4, Y: 1, Delta: 50},
		{X: 5, Y: 1, Delta: 50},
		{X: 1, Y: 2, Delta: 50},
		{X: 3, Y: 2, Delta: 50},
		{X: 5, Y: 2, Delta: 50},
		{X: 1, Y: 3, Delta: 50},
		{X: 2, Y: 3, Delta: 50},
		{X: 4, Y: 3, Delta: 50},
		{X: 5, Y: 3, Delta: 50},
		{X: 1, Y: 4, Delta: 50},
		{X: 3, Y: 4, Delta: 50},
		{X: 5, Y: 4, Delta: 50},
		{X: 1, Y: 5, Delta: 50},
		{X: 2, Y: 5, Delta: 50},
		{X: 3, Y: 5, Delta: 50},
		{X: 4, Y: 5, Delta: 50},
		{X: 5, Y: 5, Delta: 50},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("samples = %+v, want %+v", got, want)
	}
}

func TestEdgeContrastEmptyBody(t *testing.T) {
	ground, composite, body, valid := contrastPlanes(3, 3)
	for i := range valid {
		valid[i] = false
		ground[i], composite[i] = math.NaN(), math.NaN()
	}
	got, err := lookmeasure.EdgeContrast(ground, composite, body, valid, 3, 3)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty body = %+v, %v; want no samples and no error", got, err)
	}
}

func TestEdgeContrastRejectsClippedBody(t *testing.T) {
	for _, position := range []struct{ x, y int }{{1, 0}, {0, 1}, {2, 1}, {1, 2}, {0, 0}} {
		ground, composite, body, _ := contrastPlanes(3, 3)
		body[position.y*3+position.x] = true
		got, err := lookmeasure.EdgeContrast(ground, composite, body, nil, 3, 3)
		if err == nil || !strings.Contains(err.Error(), "clipped") || len(got) != 0 {
			t.Errorf("body at (%d,%d) = %+v, %v; want explicit clipping error", position.x, position.y, got, err)
		}
	}
}

func TestEdgeContrastRejectsInvalidBoundary(t *testing.T) {
	for _, index := range []int{4, 1, 3, 5, 7} {
		ground, composite, body, valid := contrastPlanes(3, 3)
		body[4], valid[index] = true, false
		ground[index], composite[index] = math.NaN(), math.NaN()
		got, err := lookmeasure.EdgeContrast(ground, composite, body, valid, 3, 3)
		if err == nil || !strings.Contains(err.Error(), "invalid") || len(got) != 0 {
			t.Errorf("invalid boundary index %d = %+v, %v; want explicit invalid boundary error", index, got, err)
		}
	}
}

func TestEdgeContrastIgnoresInvalidDiagonalBackground(t *testing.T) {
	ground, composite, body, valid := contrastPlanes(3, 3)
	body[4], composite[4] = true, 75
	valid[0], ground[0], composite[0] = false, math.NaN(), math.Inf(1)
	got, err := lookmeasure.EdgeContrast(ground, composite, body, valid, 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []lookmeasure.EdgeSample{{X: 1, Y: 1, Delta: 55}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("samples = %+v, want %+v", got, want)
	}
}

func TestEdgeContrastRejectsMalformedShapes(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int
		ground        int
		composite     int
		body          int
		valid         int
	}{
		{"zero width", 0, 3, 0, 0, 0, -1},
		{"zero height", 3, 0, 0, 0, 0, -1},
		{"negative width", -3, 3, 9, 9, 9, -1},
		{"negative height", 3, -3, 9, 9, 9, -1},
		{"overflow", int(^uint(0) >> 1), 2, 0, 0, 0, -1},
		{"short ground", 3, 3, 8, 9, 9, -1},
		{"long ground", 3, 3, 10, 9, 9, -1},
		{"short composite", 3, 3, 9, 8, 9, -1},
		{"long composite", 3, 3, 9, 10, 9, -1},
		{"short body", 3, 3, 9, 9, 8, -1},
		{"long body", 3, 3, 9, 9, 10, -1},
		{"short valid", 3, 3, 9, 9, 9, 8},
		{"long valid", 3, 3, 9, 9, 9, 10},
		{"empty valid is not nil", 3, 3, 9, 9, 9, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var valid []bool
			if tc.valid >= 0 {
				valid = make([]bool, tc.valid)
			}
			got, err := lookmeasure.EdgeContrast(make([]float64, tc.ground), make([]float64, tc.composite), make([]bool, tc.body), valid, tc.width, tc.height)
			if err == nil || len(got) != 0 {
				t.Fatalf("malformed shape = %+v, %v; want error", got, err)
			}
		})
	}
}

func TestEdgeContrastRejectsInvalidLuminance(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), -0.01, 100.01} {
		for _, plane := range []string{"ground", "composite"} {
			for _, index := range []int{0, 4} {
				ground, composite, body, _ := contrastPlanes(3, 3)
				body[4] = true
				if plane == "ground" {
					ground[index] = value
				} else {
					composite[index] = value
				}
				got, err := lookmeasure.EdgeContrast(ground, composite, body, nil, 3, 3)
				if err == nil || len(got) != 0 {
					t.Errorf("%s[%d]=%v = %+v, %v; want invalid luminance error", plane, index, value, got, err)
				}
			}
		}
	}
}

func contrastPlanes(width, height int) ([]float64, []float64, []bool, []bool) {
	ground, composite := make([]float64, width*height), make([]float64, width*height)
	body, valid := make([]bool, width*height), make([]bool, width*height)
	for i := range ground {
		ground[i], composite[i], valid[i] = 20, 20, true
	}
	return ground, composite, body, valid
}
