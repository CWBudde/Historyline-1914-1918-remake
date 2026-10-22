package main

import (
	"math"
	"slices"
)

// These are raw measurements at 1x, not terrain acceptance decisions. In
// particular, the direction's full-period correlation thresholds remain open.
type lookValue struct {
	Value   *float64 `json:"value"`
	Samples int      `json:"samples"`
	Reason  string   `json:"reason,omitempty"`
}

type lookBand struct {
	Radii [2]int    `json:"box_radii_world_px"`
	RMS   lookValue `json:"rms_l_star"`
}

type lookLag struct {
	Offset  [2]int    `json:"offset_world_px"`
	Pearson lookValue `json:"pearson"`
}

type lookMetrics struct {
	Samples         int        `json:"samples"`
	P05             lookValue  `json:"p05_l_star"`
	P95             lookValue  `json:"p95_l_star"`
	Bands           []lookBand `json:"bands"`
	Autocorrelation []lookLag  `json:"autocorrelation"`
}

const lookMinPairs = 64

func lookMeasured(value float64, samples int) lookValue {
	return lookValue{Value: &value, Samples: samples}
}

// measureLook uses only finite, fully opaque L* samples within mask. A nil mask
// selects all opaque samples. Percentiles linearly interpolate at p*(n-1).
// Band energies are RMS(box_radius_a(L*) - box_radius_b(L*)); boxes are square,
// include their center, and require every pixel of the larger support to be
// clean. Pearson uses only clean endpoint pairs (no wrapping), separate means
// for both endpoints, at least 64 pairs, and nonzero variance on both sides.
func measureLook(p *labPlane, mask []bool) lookMetrics {
	m := lookMetrics{
		Bands: []lookBand{{Radii: [2]int{1, 3}}, {Radii: [2]int{3, 10}}},
		Autocorrelation: []lookLag{
			{Offset: [2]int{71, 40}}, {Offset: [2]int{142, 0}},
			{Offset: [2]int{0, 80}}, {Offset: [2]int{0, 160}},
			{Offset: [2]int{284, 0}}, {Offset: [2]int{568, 0}},
		},
	}
	unavailable := func(reason string) lookMetrics {
		m.P05.Reason, m.P95.Reason = reason, reason
		for i := range m.Bands {
			m.Bands[i].RMS.Reason = reason
		}
		for i := range m.Autocorrelation {
			m.Autocorrelation[i].Pearson.Reason = reason
		}
		return m
	}
	if p == nil || p.w <= 0 || p.h <= 0 {
		return unavailable("insufficient_samples")
	}
	n := p.w * p.h
	if len(p.l) != n || len(p.ok) != n {
		return unavailable("invalid_plane")
	}
	if mask != nil && len(mask) != n {
		return unavailable("invalid_mask")
	}
	clean := make([]bool, n)
	values := make([]float64, 0)
	for i, lightness := range p.l {
		if p.ok[i] && (mask == nil || mask[i]) && !math.IsNaN(lightness) && !math.IsInf(lightness, 0) {
			clean[i] = true
			values = append(values, lightness)
		}
	}
	m.Samples = len(values)
	if m.Samples == 0 {
		return unavailable("insufficient_samples")
	}
	slices.Sort(values)
	percentile := func(fraction float64) lookValue {
		at := fraction * float64(len(values)-1)
		lo := int(at)
		hi := min(lo+1, len(values)-1)
		return lookMeasured(values[lo]+(at-float64(lo))*(values[hi]-values[lo]), len(values))
	}
	m.P05, m.P95 = percentile(.05), percentile(.95)
	for i := range m.Bands {
		m.Bands[i].RMS = lookBandRMS(p, clean, m.Bands[i].Radii)
	}
	for i := range m.Autocorrelation {
		m.Autocorrelation[i].Pearson = lookPearson(p, clean, m.Autocorrelation[i].Offset)
	}
	return m
}

func lookBandRMS(p *labPlane, clean []bool, radii [2]int) lookValue {
	// Summed-area tables avoid reevaluating the same large box for every center.
	stride := p.w + 1
	sums := make([]float64, stride*(p.h+1))
	counts := make([]int, len(sums))
	for y := range p.h {
		rowSum, rowCount := 0., 0
		for x := range p.w {
			i := y*p.w + x
			if clean[i] {
				rowSum += p.l[i]
				rowCount++
			}
			j := (y+1)*stride + x + 1
			sums[j] = sums[j-stride] + rowSum
			counts[j] = counts[j-stride] + rowCount
		}
	}
	box := func(x, y, radius int) (float64, int) {
		a := (y-radius)*stride + x - radius
		b := a + 2*radius + 1
		c := (y+radius+1)*stride + x - radius
		d := c + 2*radius + 1
		return sums[d] - sums[b] - sums[c] + sums[a], counts[d] - counts[b] - counts[c] + counts[a]
	}
	n, squared := 0, 0.
	large := radii[1]
	smallArea, largeArea := (2*radii[0]+1)*(2*radii[0]+1), (2*large+1)*(2*large+1)
	for y := large; y < p.h-large; y++ {
		for x := large; x < p.w-large; x++ {
			largeSum, count := box(x, y, large)
			if count != largeArea {
				continue
			}
			smallSum, _ := box(x, y, radii[0])
			difference := smallSum/float64(smallArea) - largeSum/float64(largeArea)
			squared += difference * difference
			n++
		}
	}
	if n == 0 {
		return lookValue{Reason: "insufficient_samples"}
	}
	return lookMeasured(math.Sqrt(squared/float64(n)), n)
}

func lookPearson(p *labPlane, clean []bool, offset [2]int) lookValue {
	// Online central moments retain precision for nearly constant fields.
	n, meanA, meanB, varianceA, varianceB, covariance := 0, 0., 0., 0., 0., 0.
	for y := 0; y < p.h-offset[1]; y++ {
		for x := 0; x < p.w-offset[0]; x++ {
			a, b := y*p.w+x, (y+offset[1])*p.w+x+offset[0]
			if !clean[a] || !clean[b] {
				continue
			}
			n++
			da, db := p.l[a]-meanA, p.l[b]-meanB
			meanA += da / float64(n)
			meanB += db / float64(n)
			varianceA += da * (p.l[a] - meanA)
			varianceB += db * (p.l[b] - meanB)
			covariance += da * (p.l[b] - meanB)
		}
	}
	if n < lookMinPairs {
		return lookValue{Samples: n, Reason: "insufficient_samples"}
	}
	if varianceA <= 0 || varianceB <= 0 {
		return lookValue{Samples: n, Reason: "zero_variance"}
	}
	return lookMeasured(max(-1, min(1, covariance/math.Sqrt(varianceA*varianceB))), n)
}
