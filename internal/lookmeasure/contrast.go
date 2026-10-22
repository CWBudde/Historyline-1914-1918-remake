package lookmeasure

import (
	"fmt"
	"math"
)

// EdgeSample records the visible L* contribution at one body boundary pixel.
type EdgeSample struct {
	X, Y  int
	Delta float64
}

// EdgeContrast measures abs(composite-ground) at each four-connected body
// boundary pixel, in row-major order. The planes contain CIELAB L* in [0,100];
// body is the caller's silhouette mask. Nil valid means every pixel is valid.
// Invalid pixels may contain arbitrary values, but each boundary pixel and its
// adjacent nonbody pixels must be valid. A body touching the canvas perimeter
// is clipped and returns an error. Errors never return partial samples.
func EdgeContrast(ground, composite []float64, body, valid []bool, width, height int) ([]EdgeSample, error) {
	if width <= 0 || height <= 0 || width > int(^uint(0)>>1)/height {
		return nil, fmt.Errorf("invalid dimensions %dx%d", width, height)
	}
	size := width * height
	if len(ground) != size || len(composite) != size || len(body) != size || (valid != nil && len(valid) != size) {
		return nil, fmt.Errorf("invalid plane shapes for %dx%d: ground=%d composite=%d body=%d valid=%d", width, height, len(ground), len(composite), len(body), len(valid))
	}
	for i := range ground {
		if valid != nil && !valid[i] {
			continue
		}
		if math.IsNaN(ground[i]) || ground[i] < 0 || ground[i] > 100 {
			return nil, fmt.Errorf("invalid ground L* at (%d,%d): %v", i%width, i/width, ground[i])
		}
		if math.IsNaN(composite[i]) || composite[i] < 0 || composite[i] > 100 {
			return nil, fmt.Errorf("invalid composite L* at (%d,%d): %v", i%width, i/width, composite[i])
		}
	}
	var samples []EdgeSample
	for i, inside := range body {
		if !inside {
			continue
		}
		x, y := i%width, i/width
		if x == 0 || y == 0 || x == width-1 || y == height-1 {
			return nil, fmt.Errorf("clipped body perimeter at (%d,%d)", x, y)
		}
		boundary := false
		for _, neighbor := range [4]int{i - width, i - 1, i + 1, i + width} {
			if body[neighbor] {
				continue
			}
			boundary = true
			if valid != nil && !valid[neighbor] {
				return nil, fmt.Errorf("invalid background at (%d,%d) beside body boundary (%d,%d)", neighbor%width, neighbor/width, x, y)
			}
		}
		if !boundary {
			continue
		}
		if valid != nil && !valid[i] {
			return nil, fmt.Errorf("invalid body boundary at (%d,%d)", x, y)
		}
		samples = append(samples, EdgeSample{X: x, Y: y, Delta: math.Abs(composite[i] - ground[i])})
	}
	return samples, nil
}
