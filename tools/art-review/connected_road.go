package main

import (
	"encoding/json"
	"errors"
	"image/color"
	"math"
	"strings"
)

func validateConnectionRoadFields(data []byte) error {
	var raw struct {
		Families []map[string]json.RawMessage `json:"families"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, family := range raw.Families {
		for key, value := range family {
			var id, s string
			switch {
			case strings.EqualFold(key, "markings"):
				if json.Unmarshal(family["id"], &id) != nil || json.Unmarshal(value, &s) != nil || key != "markings" || id != "road" || s != "dashed" && s != "ruts" {
					return errors.New("markings requires the canonical road-only value dashed or ruts")
				}
			case strings.EqualFold(key, "verge"):
				if json.Unmarshal(family["id"], &id) != nil || json.Unmarshal(value, &s) != nil || key != "verge" || id != "road" || s != "snow" {
					return errors.New("verge requires the canonical road-only value snow")
				}
				var markings string
				if json.Unmarshal(family["markings"], &markings) != nil || markings == "" {
					return errors.New("verge requires road markings")
				}
			}
		}
	}
	return nil
}

// Winter roads carry two darker slush ruts instead of painted lines, and their
// shoulders are dirty snow instead of dusty gravel.
const (
	connectionRutOffset   = 4.5 // world px from the centre line
	connectionRutCore     = .6  // fully dark half-width
	connectionRutFade     = 1.2 // soft edge beyond the core
	connectionRutStrength = .55
	// connectionSnowShoulderMix blends the earth border toward snow; the dusty
	// summer shoulder blends at .35.
	connectionSnowShoulderMix = .85
)

var (
	connectionRutColor      = color.NRGBA{88, 86, 90, 255}
	connectionDustyShoulder = color.NRGBA{177, 168, 145, 255}
	connectionSnowShoulder  = color.NRGBA{214, 217, 222, 255}
)

// connectionRoadRutAt is the weight of the slush ruts at p: two tracks along
// every course, connectionRutOffset either side of its centre line.
func connectionRoadRutAt(paint connectionRoadPaint, p connectionPoint) float64 {
	best := math.Inf(1)
	for _, c := range paint.courses {
		for i := 1; i < len(c.points); i++ {
			a, b := c.points[i-1], c.points[i]
			delta := b.sub(a)
			l2 := delta[0]*delta[0] + delta[1]*delta[1]
			if l2 == 0 {
				continue
			}
			q := p.sub(a)
			t := max(0, min(1, (q[0]*delta[0]+q[1]*delta[1])/l2))
			v := p.sub(a.add(delta.mul(t)))
			best = min(best, v[0]*v[0]+v[1]*v[1])
		}
	}
	off := math.Abs(math.Sqrt(best) - connectionRutOffset)
	return connectionRutStrength * (1 - connectionSmoothStep((off-connectionRutCore)/connectionRutFade))
}

type connectionRoadCourse struct {
	points  []connectionPoint
	lengths []float64
	length  float64
}
type connectionRoadPaint struct {
	courses   []connectionRoadCourse
	junction  bool
	halfWidth float64
	ruts      bool // slush ruts instead of the dashed centre and edge lines
	snow      bool // dirty snow shoulders instead of dusty gravel
}

// Markings follow the same curves as the surface. A fixed phase at each port
// makes a dash cross a shared edge continuously, independently of tile origin.
// Only the interior dash spacing adjusts to fit a curved two-port course.
func connectionRoadPaintFor(g connectionGeometry, v connectionVariant) connectionRoadPaint {
	r := connectionRoadPaint{junction: len(v.Ports) >= 3}
	course := func(points []connectionPoint) {
		c := connectionRoadCourse{points: points, lengths: make([]float64, len(points))}
		for i := 1; i < len(points); i++ {
			c.length += points[i].sub(points[i-1]).length()
			c.lengths[i] = c.length
		}
		r.courses = append(r.courses, c)
	}
	curve := func(points []connectionPoint, a, b, c, d connectionPoint) []connectionPoint {
		for i := 1; i <= 64; i++ {
			t := float64(i) / 64
			s := 1 - t
			points = append(points, a.mul(s*s*s).add(b.mul(3*s*s*t)).add(c.mul(3*s*t*t)).add(d.mul(t*t*t)))
		}
		return points
	}
	port := func(p int) (connectionPoint, connectionPoint, connectionPoint) {
		edge := g.Ports[p-1]
		in := g.Neighbors[p-1].mul(-1 / g.Neighbors[p-1].length())
		return edge.sub(in.mul(16)), edge.add(in.mul(8)), in
	}
	if len(v.Ports) == 2 {
		a0, a, na := port(v.Ports[0])
		b0, b, nb := port(v.Ports[1])
		points := []connectionPoint{a0, a}
		if (v.Ports[0]+2)%6+1 == v.Ports[1] {
			points = append(points, b)
		} else {
			points = curve(points, a, a.add(na.mul(20)), b.add(nb.mul(20)), b)
		}
		course(append(points, b0))
	} else {
		center := connectionPoint{48, 40}
		for _, p := range v.Ports {
			a0, a, n := port(p)
			course(curve([]connectionPoint{a0, a}, a, a.add(n.mul(16)), center.sub(n.mul(8)), center))
		}
	}
	return r
}

func connectionRoadPaintAt(paint connectionRoadPaint, p connectionPoint, aa float64) float64 {
	best, along, length := math.Inf(1), 0., 0.
	for _, c := range paint.courses {
		for i := 1; i < len(c.points); i++ {
			a, b := c.points[i-1], c.points[i]
			delta := b.sub(a)
			l := c.lengths[i] - c.lengths[i-1]
			if l == 0 {
				continue
			}
			q := p.sub(a)
			t := max(0, min(1, (q[0]*delta[0]+q[1]*delta[1])/(l*l)))
			v := p.sub(a.add(delta.mul(t)))
			d := v[0]*v[0] + v[1]*v[1]
			if d < best {
				best = d
				along = c.lengths[i-1] + t*l - 16
				length = c.length - 32
			}
		}
	}
	phase := along
	if paint.junction {
		if p.sub(connectionPoint{48, 40}).length() < 9 {
			return 0
		}
	} else {
		if along >= length-8 {
			phase = along - length
		} else if along > 8 {
			cycles := max(2, math.Round(length/12))
			phase = 8 + (along-8)*(cycles*12-16)/(length-16)
		}
	}
	distance := math.Sqrt(best)
	stripe := connectionRamp(.45-distance, aa)
	dash := connectionRamp(4-math.Abs(math.Remainder(phase, 12)), aa)
	mark := min(stripe, dash)
	if paint.halfWidth > 0 {
		edge := connectionRamp(.3-math.Abs(distance-(paint.halfWidth-1)), aa)
		mark = max(mark, edge)
	}
	return mark
}
