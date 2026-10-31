package referencedesign

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func lookTestPlane(w, h int, value func(x, y int) float64) *labPlane {
	p := &labPlane{w: w, h: h, l: make([]float64, w*h), ok: make([]bool, w*h)}
	for y := range h {
		for x := range w {
			p.l[y*w+x], p.ok[y*w+x] = value(x, y), true
		}
	}
	return p
}

func lookTestValue(t *testing.T, got lookValue, want float64) {
	t.Helper()
	if got.Value == nil || got.Reason != "" || math.Abs(*got.Value-want) > 1e-8 {
		t.Fatalf("got %+v, want value %g", got, want)
	}
}

func TestLookPercentilesAndCleanMask(t *testing.T) {
	p := lookTestPlane(100, 1, func(x, _ int) float64 { return float64(x) })
	got := measureLook(p, nil)
	if got.Samples != 100 || got.P05.Samples != 100 || got.P95.Samples != 100 {
		t.Fatalf("sample counts: %+v", got)
	}
	lookTestValue(t, got.P05, 4.95)
	lookTestValue(t, got.P95, 94.05)
	mask := make([]bool, 100)
	for i := 20; i < 40; i++ {
		mask[i] = true
	}
	p.ok[20] = false
	p.l[20] = 10000  // Nonopaque pixels cannot affect percentiles.
	p.l[80] = -10000 // Pixels outside the clean mask cannot affect them either.
	got = measureLook(p, mask)
	if got.Samples != 19 {
		t.Fatalf("clean sample count %d, want 19", got.Samples)
	}
	lookTestValue(t, got.P05, 21.9)
	lookTestValue(t, got.P95, 38.1)
}

func TestLookBandFullCleanSupport(t *testing.T) {
	p := lookTestPlane(25, 25, func(x, y int) float64 {
		if x == 12 && y == 12 {
			return 100
		}
		return 0
	})
	got := measureLook(p, nil)
	if len(got.Bands) != 2 || got.Bands[0].Radii != [2]int{1, 3} || got.Bands[1].Radii != [2]int{3, 10} {
		t.Fatalf("unexpected bands: %+v", got.Bands)
	}
	// The impulse contributes to nine radius-1 boxes and 49 radius-3 boxes.
	inner, outer := 100.0/9-100.0/49, 100.0/49
	want := math.Sqrt((9*inner*inner + 40*outer*outer) / (19 * 19))
	lookTestValue(t, got.Bands[0].RMS, want)
	if got.Bands[0].RMS.Samples != 19*19 || got.Bands[1].RMS.Samples != 5*5 {
		t.Fatalf("full-support counts: %+v", got.Bands)
	}
	mask := make([]bool, 25*25)
	for i := range mask {
		mask[i] = true
	}
	mask[12*25+12] = false
	got = measureLook(p, mask)
	if got.Bands[0].RMS.Samples != 19*19-7*7 || got.Bands[1].RMS.Value != nil || got.Bands[1].RMS.Reason != "insufficient_samples" {
		t.Fatalf("mask must exclude the full box support: %+v", got.Bands)
	}
	mask[12*25+12], p.ok[12*25+12] = true, false
	alphaExcluded := measureLook(p, mask)
	if alphaExcluded.Bands[0].RMS.Samples != got.Bands[0].RMS.Samples || alphaExcluded.Bands[1].RMS.Samples != 0 {
		t.Fatalf("nonopaque support must be excluded like masked support: %+v", alphaExcluded.Bands)
	}
}

func TestLookBandScaleResponse(t *testing.T) {
	for _, wavelength := range []float64{5, 16} {
		p := lookTestPlane(160, 100, func(x, _ int) float64 {
			return 50 + 20*math.Sin(2*math.Pi*float64(x)/wavelength)
		})
		got := measureLook(p, nil)
		a, b := *got.Bands[0].RMS.Value, *got.Bands[1].RMS.Value
		if wavelength == 5 && a <= b || wavelength == 16 && b <= a {
			t.Errorf("wavelength %g: small band %g, large band %g", wavelength, a, b)
		}
	}
}

func TestLookAutocorrelationRepeatAndNoise(t *testing.T) {
	offsets := [][2]int{{71, 40}, {142, 0}, {0, 80}, {0, 160}, {284, 0}, {568, 0}}
	p := lookTestPlane(720, 240, func(x, y int) float64 {
		return 50 + 10*math.Sin(2*math.Pi*float64(x)/71) + 5*math.Cos(2*math.Pi*float64(y)/40)
	})
	got := measureLook(p, nil)
	if len(got.Autocorrelation) != len(offsets) {
		t.Fatalf("lag count %d, want %d", len(got.Autocorrelation), len(offsets))
	}
	for i, lag := range got.Autocorrelation {
		if lag.Offset != offsets[i] || lag.Pearson.Samples != (p.w-lag.Offset[0])*(p.h-lag.Offset[1]) {
			t.Fatalf("lag/pair count: %+v", lag)
		}
		lookTestValue(t, lag.Pearson, 1)
	}
	state := uint64(0x123456789abcdef)
	noise := lookTestPlane(720, 240, func(_, _ int) float64 {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		return float64(state>>11) / (1 << 53) * 100
	})
	for _, lag := range measureLook(noise, nil).Autocorrelation {
		if lag.Pearson.Value == nil || math.Abs(*lag.Pearson.Value) > .025 {
			t.Errorf("uncorrelated field: %+v", lag)
		}
	}
}

func TestLookAutocorrelationUsesCleanEndpoints(t *testing.T) {
	p := lookTestPlane(600, 2, func(x, y int) float64 { return float64(x + y) })
	mask := make([]bool, p.w*p.h)
	for _, x := range []int{0, 1, 142, 143} {
		mask[x] = true
	}
	p.ok[0] = false
	got := measureLook(p, mask).Autocorrelation[1]
	if got.Pearson.Samples != 1 || got.Pearson.Value != nil || got.Pearson.Reason != "insufficient_samples" {
		t.Fatalf("nonclean endpoint counted: %+v", got)
	}
}

func TestLookAutocorrelationRequires64Pairs(t *testing.T) {
	p := lookTestPlane(400, 1, func(x, _ int) float64 { return float64(x) })
	mask := make([]bool, p.w)
	for x := range 64 {
		mask[x], mask[x+142] = true, true
	}
	got := measureLook(p, mask).Autocorrelation[1]
	if got.Pearson.Samples != 64 {
		t.Fatalf("pair count %d, want 64", got.Pearson.Samples)
	}
	lookTestValue(t, got.Pearson, 1)
	// Opposite ramps must retain the sign of the correlation.
	for x := range 64 {
		p.l[x+142] = -float64(x)
	}
	lookTestValue(t, measureLook(p, mask).Autocorrelation[1].Pearson, -1)
	mask[0] = false
	got = measureLook(p, mask).Autocorrelation[1]
	if got.Pearson.Samples != 63 || got.Pearson.Value != nil || got.Pearson.Reason != "insufficient_samples" {
		t.Fatalf("sparse Pearson must be unavailable: %+v", got)
	}
}

func TestLookUnavailableValuesRemainNull(t *testing.T) {
	p := lookTestPlane(720, 240, func(_, _ int) float64 { return 40 })
	got := measureLook(p, nil)
	for _, lag := range got.Autocorrelation {
		if lag.Pearson.Value != nil || lag.Pearson.Reason != "zero_variance" {
			t.Errorf("constant field correlation: %+v", lag)
		}
	}
	for _, band := range got.Bands {
		lookTestValue(t, band.RMS, 0)
	}
	for _, mask := range [][]bool{make([]bool, p.w*p.h), {true}} {
		got = measureLook(p, mask)
		if got.Samples != 0 || got.P05.Value != nil || got.P95.Value != nil || got.P05.Reason == "" {
			t.Fatalf("unavailable mask: %+v", got)
		}
		b, err := json.Marshal(got)
		if err != nil || !strings.Contains(string(b), `"value":null`) {
			t.Fatalf("null unavailable metric: %s, %v", b, err)
		}
	}
	if got := measureLook(nil, nil); got.P05.Value != nil || got.P05.Reason != "insufficient_samples" {
		t.Fatalf("nil plane: %+v", got)
	}
	single := measureLook(lookTestPlane(1, 1, func(_, _ int) float64 { return 12 }), nil)
	lookTestValue(t, single.P05, 12)
	lookTestValue(t, single.P95, 12)
}

func TestLookMeanUsesFiniteOpaqueMaskedSamples(t *testing.T) {
	p := lookTestPlane(4, 1, func(x, _ int) float64 { return []float64{10, 20, 100, 40}[x] })
	p.ok[2] = false
	got := measureLook(p, []bool{true, true, true, false})
	lookTestValue(t, got.MeanL, 15)
	if got.MeanL.Samples != 2 {
		t.Fatalf("mean count %+v", got.MeanL)
	}
	p.l[0] = math.NaN()
	got = measureLook(p, []bool{true, false, false, false})
	if got.MeanL.Value != nil || got.MeanL.Reason != "insufficient_samples" {
		t.Fatalf("missing mean fabricated: %+v", got.MeanL)
	}
}

func TestLookRawSigmaAndEdgesKeepUnroundedValues(t *testing.T) {
	p := lookTestPlane(5, 5, func(x, _ int) float64 { return float64(x) * 20 })
	got := measureLook(p, nil)
	lookTestValue(t, got.SigmaL, math.Sqrt(800))
	lookTestValue(t, got.EdgeDensity, 1)
	if got.EdgeDensity.Samples != 9 {
		t.Fatalf("Sobel support count %+v", got.EdgeDensity)
	}
	p = lookTestPlane(3, 3, func(x, y int) float64 { return float64(y*3 + x) })
	got = measureLook(p, nil)
	lookTestValue(t, got.SigmaL, math.Sqrt(60.0/9))
	if *got.SigmaL.Value == math.Round(*got.SigmaL.Value*100)/100 {
		t.Fatal("gate inputs rounded to two decimals")
	}
}
