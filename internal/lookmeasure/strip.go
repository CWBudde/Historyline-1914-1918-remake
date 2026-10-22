package lookmeasure

import (
	"fmt"
	"math"
)

// StripPixels returns row-major indices of pixel centres inside an oriented
// strip. Distances start and end are measured from from towards to. Projection
// is in [start,end); signed perpendicular distance, along normal (-uy,ux), is
// in [-halfWidth,halfWidth). The complete strip rectangle must fit the canvas.
// Only its bounding rectangle is scanned; a strip may contain no pixel centres.
func StripPixels(width, height int, from, to [2]float64, start, end, halfWidth float64) ([]int, error) {
	if width <= 0 || height <= 0 || width > int(^uint(0)>>1)/height {
		return nil, fmt.Errorf("invalid dimensions %dx%d", width, height)
	}
	for _, value := range [...]float64{from[0], from[1], to[0], to[1], start, end, halfWidth} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("nonfinite strip input: %v", value)
		}
	}
	dx, dy := to[0]-from[0], to[1]-from[1]
	length := math.Hypot(dx, dy)
	if length == 0 || math.IsInf(length, 0) {
		return nil, fmt.Errorf("invalid segment length: %v", length)
	}
	if start < 0 || start >= end || end > length || halfWidth <= 0 {
		return nil, fmt.Errorf("invalid strip interval [%v,%v) or half width %v for segment length %v", start, end, halfWidth, length)
	}
	ux, uy := dx/length, dy/length
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, distance := range [2]float64{start, end} {
		for _, offset := range [2]float64{-halfWidth, halfWidth} {
			x := from[0] + distance*ux - offset*uy
			y := from[1] + distance*uy + offset*ux
			if math.IsNaN(x) || math.IsNaN(y) || x < 0 || y < 0 || x > float64(width) || y > float64(height) {
				return nil, fmt.Errorf("strip corner (%v,%v) outside %dx%d canvas", x, y, width, height)
			}
			minX, minY = math.Min(minX, x), math.Min(minY, y)
			maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
		}
	}
	// Clip scan bounds only after checking all actual corners. Comparing before
	// float-to-int conversion also avoids overflow at the largest int dimension.
	if minX >= float64(width) || minY >= float64(height) {
		return nil, nil
	}
	x0, y0 := int(math.Floor(minX)), int(math.Floor(minY))
	x1, y1 := width, height
	if maxX < float64(width) {
		x1 = int(math.Ceil(maxX))
	}
	if maxY < float64(height) {
		y1 = int(math.Ceil(maxY))
	}
	var pixels []int
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			px, py := float64(x)+0.5-from[0], float64(y)+0.5-from[1]
			projection := px*ux + py*uy
			perpendicular := px*(-uy) + py*ux
			if projection >= start && projection < end && perpendicular >= -halfWidth && perpendicular < halfWidth {
				pixels = append(pixels, y*width+x)
			}
		}
	}
	return pixels, nil
}
