package main

import (
	"image"
	"image/color"
	"math"
	"testing"
)

var groundTestDeltas = [6][2]int{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
var groundTestPolygon = [6][2]float64{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}
var groundTestColors = map[string]color.NRGBA{
	"grass": {11, 43, 23, 255}, "mud": {211, 31, 53, 255}, "stone": {37, 149, 229, 255},
	"grass_w": {239, 241, 251, 255}, "rough_w": {41, 63, 79, 255}, "stoney_w": {131, 157, 181, 255}, "shelled_w": {83, 71, 59, 255},
}

func groundTestConstants() map[string]*image.NRGBA {
	m := map[string]*image.NRGBA{}
	for id, c := range groundTestColors {
		im := image.NewNRGBA(image.Rect(0, 0, 2, 2))
		for y := range 2 {
			for x := range 2 {
				im.SetNRGBA(x, y, c)
			}
		}
		m[id] = im
	}
	return m
}
func groundTestContext(t *testing.T, id string, ids [6]string) groundContext {
	t.Helper()
	c, err := newGroundContext(id, ids[:])
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func groundTestDistance(x, y float64) float64 {
	inside := true
	best := math.Inf(1)
	for i, a := range groundTestPolygon {
		b := groundTestPolygon[(i+1)%6]
		dx, dy := b[0]-a[0], b[1]-a[1]
		qx, qy := x-a[0], y-a[1]
		if dx*qy-dy*qx < 0 {
			inside = false
		}
		u := math.Max(0, math.Min(1, (qx*dx+qy*dy)/(dx*dx+dy*dy)))
		best = math.Min(best, math.Hypot(qx-u*dx, qy-u*dy))
	}
	if inside {
		return 0
	}
	return best
}
func groundTestS(d float64) float64 {
	u := math.Max(0, math.Min(1, (d-4)/8))
	return u * u * u * (10 + u*(-15+6*u))
}
func groundTestGlobalRGB(world map[[2]int]string, p [2]float64, base string) [3]float64 {
	var sum [3]float64
	total, e := 0., 1.
	for origin, id := range world {
		x, y := p[0]-float64(origin[0]), p[1]-float64(origin[1])
		if id == base {
			e *= groundTestS(groundTestDistance(x, y))
			continue
		}
		u := math.Max(0, 1-((x-48)*(x-48)+(y-40)*(y-40))/6400)
		w := u * u * u
		c := groundTestColors[id]
		total += w
		sum[0] += w * float64(c.R)
		sum[1] += w * float64(c.G)
		sum[2] += w * float64(c.B)
	}
	b := groundTestColors[base]
	bc := [3]float64{float64(b.R), float64(b.G), float64(b.B)}
	for i := range sum {
		if total == 0 {
			sum[i] = bc[i]
		} else {
			sum[i] = bc[i]*(1-e) + sum[i]/total*e
		}
	}
	return sum
}
func groundTestNearRGB(t *testing.T, got, want [3]float64) {
	t.Helper()
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-8 {
			t.Fatalf("RGB got %v want %v", got, want)
		}
	}
}
func TestGroundFieldWeightedProductAndFilledDistance(t *testing.T) {
	materials := groundTestConstants()
	cases := []struct {
		name, id string
		ids      [6]string
		p        [2]float64
	}{
		{"single-center-weight", "mud", [6]string{"grass", "grass", "grass", "grass", "grass", "grass"}, [2]float64{48, 40}},
		{"pair", "mud", [6]string{"stone", "grass", "grass", "grass", "grass", "grass"}, [2]float64{83.5, 20}},
		{"triple", "rough_w", [6]string{"stoney_w", "shelled_w", "rough_w", "stoney_w", "shelled_w", "rough_w"}, [2]float64{96, 40}},
		{"NE-SE-product-tie", "mud", [6]string{"grass", "grass", "mud", "mud", "mud", "mud"}, [2]float64{86, 40}},
		{"filled-base-interior", "mud", [6]string{"grass", "grass", "mud", "mud", "mud", "mud"}, [2]float64{119, 0}},
		{"homogeneous-no-grass", "stone", [6]string{"stone", "stone", "stone", "stone", "stone", "stone"}, [2]float64{48, 1}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c := groundTestContext(t, tt.id, tt.ids)
			world := map[[2]int]string{{0, 0}: tt.id}
			for i, d := range groundTestDeltas {
				world[d] = tt.ids[i]
			}
			got := groundContextRGB(c, connectionPoint(tt.p), materials)
			base := "grass"
			if tt.id == "rough_w" {
				base = "grass_w"
			}
			groundTestNearRGB(t, got, groundTestGlobalRGB(world, tt.p, base))
			if tt.name == "NE-SE-product-tie" {
				a := groundTestS(groundTestDistance(15, 80))
				b := groundTestS(groundTestDistance(15, 0))
				if a <= 0 || a >= 1 || math.Abs(a-b) > 1e-12 {
					t.Fatalf("bad product witness %g %g", a, b)
				}
				want := 11 + (211-11)*a*b
				if math.Abs(got[0]-want) > 1e-9 || math.Abs(got[0]-(11+200*math.Min(a, b))) < 1 {
					t.Fatalf("product not minimum: %v", got)
				}
			}
		})
	}
	if groundFilledDistance(connectionPoint{48, 40}) != 0 || groundFilledDistance(connectionPoint{96, 40}) != 0 {
		t.Fatal("filled polygon interior/boundary distance must be zero")
	}
}
func TestGroundContextRejectsInvalidIDsSeasonsAndDirections(t *testing.T) {
	for _, tt := range []struct {
		id  string
		ids []string
	}{
		{"bogus", []string{"grass", "grass", "grass", "grass", "grass", "grass"}},
		{"mud", []string{"grass", "grass", "grass", "grass", "grass"}},
		{"mud", []string{"grass", "grass", "grass", "grass", "grass", "grass", "grass"}},
		{"mud", []string{"grass_w", "grass", "grass", "grass", "grass", "grass"}},
		{"rough_w", []string{"grass_w", "mud", "grass_w", "grass_w", "grass_w", "grass_w"}},
		{"mud", []string{"grass", "grass", "grass", "grass", "grass", "Mud"}},
	} {
		if _, err := newGroundContext(tt.id, tt.ids); err == nil {
			t.Fatalf("accepted invalid center/neighbors %q %v", tt.id, tt.ids)
		}
	}
}
func groundTestWorld() map[[2]int]string {
	world := map[[2]int]string{{0, 0}: ""}
	front := [][2]int{{0, 0}}
	for range 2 {
		var next [][2]int
		for _, o := range front {
			for _, d := range groundTestDeltas {
				q := [2]int{o[0] + d[0], o[1] + d[1]}
				if _, ok := world[q]; !ok {
					world[q] = ""
					next = append(next, q)
				}
			}
		}
		front = next
	}
	return world
}
func groundTestAt(t *testing.T, world map[[2]int]string, o [2]int) groundContext {
	t.Helper()
	var ids [6]string
	for i, d := range groundTestDeltas {
		var ok bool
		ids[i], ok = world[[2]int{o[0] + d[0], o[1] + d[1]}]
		if !ok {
			t.Fatal("incomplete test world")
		}
	}
	return groundTestContext(t, world[o], ids)
}
func TestGroundCompatibleContextsSharedEdgesTriplesAndFilterStrip(t *testing.T) {
	materials := groundTestConstants()
	for permutation := range 6 {
		world := groundTestWorld()
		for o := range world {
			ids := []string{"rough_w", "stoney_w", "shelled_w"}
			index := (o[0]/71 + o[1]/40 + permutation) % 3
			if index < 0 {
				index += 3
			}
			world[o] = ids[index]
		}
		// Change terrain order and include real base boundaries, while retaining a nonbase center.
		if permutation%2 == 0 {
			world[[2]int{-71, -40}] = "grass_w"
			world[[2]int{0, -80}] = "grass_w"
		}
		c := groundTestAt(t, world, [2]int{0, 0})
		for dir, d := range groundTestDeltas {
			n := groundTestAt(t, world, d)
			// Direction NE corresponds to polygon edge 1, SE edge 2, ..., N edge 0.
			a, b := groundTestPolygon[(dir+1)%6], groundTestPolygon[(dir+2)%6]
			dx, dy := b[0]-a[0], b[1]-a[1]
			length := math.Hypot(dx, dy)
			for _, u := range []float64{0, .2, .5, .8, 1} {
				for _, offset := range []float64{-3.9, -2, 0, 2, 3.9} {
					p := [2]float64{a[0] + u*dx - offset*dy/length, a[1] + u*dy + offset*dx/length}
					if groundTestDistance(p[0], p[1]) > 4 || groundTestDistance(p[0]-float64(d[0]), p[1]-float64(d[1])) > 4 {
						continue
					}
					want := groundTestGlobalRGB(world, p, "grass_w")
					groundTestNearRGB(t, groundContextRGB(c, connectionPoint(p), materials), want)
					groundTestNearRGB(t, groundContextRGB(n, connectionPoint{p[0] - float64(d[0]), p[1] - float64(d[1])}, materials), want)
				}
			}
		}
	}
	// An incorrect actual neighbor must visibly fail the independent comparison.
	world := groundTestWorld()
	for o := range world {
		world[o] = "rough_w"
	}
	world[[2]int{71, -40}] = "stoney_w"
	correct := groundTestAt(t, world, [2]int{})
	wrong := correct
	wrong.NeighborIDs[0] = "rough_w"
	p := connectionPoint{83.5, 20}
	a := groundContextRGB(correct, p, materials)
	b := groundContextRGB(wrong, p, materials)
	if math.Abs(a[0]-b[0]) < 10 {
		t.Fatal("wrong-neighbor fixture does not detect missing context")
	}
}
func TestGroundProductMedialTieHasContinuousFirstDerivative(t *testing.T) {
	c := groundTestContext(t, "mud", [6]string{"grass", "grass", "mud", "mud", "mud", "mud"})
	m := groundTestConstants()
	f := func(y float64) float64 { return groundContextRGB(c, connectionPoint{86, y}, m)[0] }
	h := 1e-4
	left := (f(40) - f(40-h)) / h
	right := (f(40+h) - f(40)) / h
	if math.Abs(left-right) > .003 {
		t.Fatalf("medial tie gradient cusp: left %g right %g", left, right)
	}
	if math.Abs(f(40)-11) < 1 || math.Abs(f(40)-211) < 1 {
		t.Fatal("tie fixture must be in transition")
	}
}
