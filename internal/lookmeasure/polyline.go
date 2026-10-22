package lookmeasure

import (
	"fmt"
	"math"
)

// Point is a continuous position in image or world coordinates.
type Point struct{ X, Y float64 }

// Segment is a finite straight reference edge, including both endpoints.
type Segment struct{ From, To Point }

// EdgeDistance records a source midpoint's nearest reference distance and the
// source arclength represented by that sample.
type EdgeDistance struct {
	Point    Point
	Distance float64
	Weight   float64
}

// PolylineDistances samples an open source path, in path order. Each source
// segment of length L is split into ceil(L/maxStep) equal pieces, represented
// by their midpoints and weighted by piece length. Distances are to the nearest
// finite reference segment, with projection clamped to its endpoints. All input
// geometry and sample counts are validated before allocating the result.
// Nonfinite derived geometry returns an error without partial samples.
func PolylineDistances(points []Point, edges []Segment, maxStep float64) ([]EdgeDistance, error) {
	if len(points) < 2 || len(edges) == 0 {
		return nil, fmt.Errorf("need at least two source points and one reference edge")
	}
	if math.IsNaN(maxStep) || math.IsInf(maxStep, 0) || maxStep <= 0 {
		return nil, fmt.Errorf("invalid maximum sample step: %v", maxStep)
	}
	for i, edge := range edges {
		if _, _, err := polylineVector(edge.From, edge.To); err != nil {
			return nil, fmt.Errorf("reference edge %d: %w", i, err)
		}
	}
	// Each EdgeDistance contains four float64 values. Guard backing-array byte
	// overflow as well as conversion and accumulation of individual counts.
	maxSamples := int(^uint(0)>>1) / 32
	total := 0
	var totalLength float64
	for i := 1; i < len(points); i++ {
		_, length, err := polylineVector(points[i-1], points[i])
		if err != nil {
			return nil, fmt.Errorf("source segment %d: %w", i-1, err)
		}
		totalLength += length
		if math.IsInf(totalLength, 0) {
			return nil, fmt.Errorf("source polyline length overflow")
		}
		count := math.Max(1, math.Ceil(length/maxStep))
		if math.IsInf(count, 0) || count >= float64(maxSamples) || int(count) > maxSamples-total {
			return nil, fmt.Errorf("sample count overflow at source segment %d", i-1)
		}
		total += int(count)
	}
	result := make([]EdgeDistance, 0, total)
	for i := 1; i < len(points); i++ {
		from, to := points[i-1], points[i]
		vector := Point{X: to.X - from.X, Y: to.Y - from.Y}
		length := math.Hypot(vector.X, vector.Y)
		count := int(math.Max(1, math.Ceil(length/maxStep)))
		weight := length / float64(count)
		for j := range count {
			fraction := (float64(j) + 0.5) / float64(count)
			point := Point{X: from.X + fraction*vector.X, Y: from.Y + fraction*vector.Y}
			nearest := math.Inf(1)
			for _, edge := range edges {
				distance, err := polylineEdgeDistance(point, edge)
				if err != nil {
					return nil, err
				}
				nearest = math.Min(nearest, distance)
			}
			result = append(result, EdgeDistance{Point: point, Distance: nearest, Weight: weight})
		}
	}
	return result, nil
}

func polylineVector(from, to Point) (Point, float64, error) {
	for _, value := range [4]float64{from.X, from.Y, to.X, to.Y} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return Point{}, 0, fmt.Errorf("nonfinite endpoint coordinate: %v", value)
		}
	}
	vector := Point{X: to.X - from.X, Y: to.Y - from.Y}
	length := math.Hypot(vector.X, vector.Y)
	if length == 0 || math.IsInf(length, 0) {
		return Point{}, 0, fmt.Errorf("invalid segment length: %v", length)
	}
	return vector, length, nil
}

func polylineEdgeDistance(point Point, edge Segment) (float64, error) {
	dx, dy := edge.To.X-edge.From.X, edge.To.Y-edge.From.Y
	length := math.Hypot(dx, dy)
	ux, uy := dx/length, dy/length
	px, py := point.X-edge.From.X, point.Y-edge.From.Y
	projection := px*ux + py*uy
	if math.IsNaN(projection) || math.IsInf(projection, 0) {
		return 0, fmt.Errorf("nonfinite source-to-reference projection")
	}
	var distance float64
	switch {
	case projection <= 0:
		distance = math.Hypot(px, py)
	case projection >= length:
		distance = math.Hypot(point.X-edge.To.X, point.Y-edge.To.Y)
	default:
		distance = math.Hypot(px-projection*ux, py-projection*uy)
	}
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		return 0, fmt.Errorf("nonfinite source-to-reference distance")
	}
	return distance, nil
}
