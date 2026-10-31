package artreview

import (
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// Derived trenches (mapcompose.TrenchShape) join the levels' trench pieces
// across hexes: linked ports run through their edges, blind spurs end inside
// the hex, and a dug-in position takes its links as courses that end at its
// ring. A derived run keeps the dugout and blind spurs of the roster piece
// with the same ports, where they do not take a port of the shape.
func derivedTrenchVariant(roster []connectionVariant, s mapcompose.TrenchShape) connectionVariant {
	v := connectionVariant{ID: s.ID(), Ports: maskPorts(s.Links), ClosedRing: s.Ring}
	if s.Ring {
		return v
	}
	v.BlindSpurs = maskPorts(s.Spurs)
	taken := s.Links | s.Spurs
	for _, r := range roster {
		if r.ClosedRing || portsMask(r.Ports) != taken {
			continue
		}
		for _, p := range r.BlindSpurs {
			if taken&(1<<(p-1)) == 0 {
				v.BlindSpurs = append(v.BlindSpurs, p)
				taken |= 1 << (p - 1)
			}
		}
		if d := r.DugoutDirection; d != 0 && taken&(1<<(d-1)) == 0 {
			v.DugoutDirection = d
		}
		break
	}
	return v
}

func maskPorts(m uint8) []int {
	ports := []int{}
	for p := 1; p <= 6; p++ {
		if m&(1<<(p-1)) != 0 {
			ports = append(ports, p)
		}
	}
	return ports
}

func portsMask(ports []int) (m uint8) {
	for _, p := range ports {
		m |= 1 << (p - 1)
	}
	return m
}

// trenchRingDistance is the signed distance from the ring's centerline: the
// rounded rectangle of connectionTrenchRoutes, negative inside.
func trenchRingDistance(p connectionPoint) float64 {
	x, y := math.Abs(p[0]-48)-11, math.Abs(p[1]-40)-6
	return math.Hypot(math.Max(x, 0), math.Max(y, 0)) + math.Min(math.Max(x, y), 0) - 7
}

// trenchRingEntry is where the course from a outside the ring toward the
// centre meets the ring's centerline.
func trenchRingEntry(a connectionPoint) connectionPoint {
	lo, hi := 0., 1.
	centre := connectionPoint{48, 40}
	for range 60 {
		t := (lo + hi) / 2
		if trenchRingDistance(a.add(centre.sub(a).mul(t))) > 0 {
			lo = t
		} else {
			hi = t
		}
	}
	return a.add(centre.sub(a).mul(hi))
}
