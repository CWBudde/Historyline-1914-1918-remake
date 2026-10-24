package main

import (
	"image"
	"math"
	"path/filepath"
	"testing"
)

// TestCompoundFeatherIntactScenes: the installed HQ, depot and factory scenes
// (both seasons) fade their ground apron out at the outer boundary along an
// irregular line, while their keep polygons (walls, towers, fences and
// buildings that reach the boundary) keep the artwork exactly and the entry
// corridor stays open.
func TestCompoundFeatherIntactScenes(t *testing.T) {
	var pooled []float64
	for _, name := range []string{"compound-hq-v4.json", "compound-depot-v5.json", "compound-factory-v5.json", "compound-factory-summer-v3.json"} {
		in, err := loadCompoundInputs(filepath.Join("..", "..", "art", "pipeline", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range in.Manifest.Scenes {
			t.Run(s.ID, func(t *testing.T) {
				if s.EdgeFeather == 0 || len(s.EdgeKeep) == 0 {
					t.Fatalf("%s: no edge feather with keep polygons", name)
				}
				l, err := compoundLayoutFor(s.ID)
				if err != nil {
					t.Fatal(err)
				}
				framed, art4, _, _, err := compoundSceneArtwork(in, s, l)
				if err != nil {
					t.Fatal(err)
				}
				f, _, err := compoundEntryRoad(in.EntryRoads[s.ID], l)
				if err != nil {
					t.Fatal(err)
				}
				corridor := f.Width/2 + f.Border + 2
				kept := 0
				for y := range l.Height * 4 {
					for x := range l.Width * 4 {
						p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
						if compoundKeepDistance(s.EdgeKeep, p) == 0 {
							kept++
							if art4.NRGBAAt(x, y) != framed.NRGBAAt(x, y) {
								t.Fatalf("keep polygon changed at (%.2f,%.2f)", p[0], p[1])
							}
						}
					}
				}
				art2, err := resizeLanczos(art4, l.Width*2, l.Height*2)
				if err != nil {
					t.Fatal(err)
				}
				before2, err := resizeLanczos(framed, l.Width*2, l.Height*2)
				if err != nil {
					t.Fatal(err)
				}
				// Ground: away from the keep polygons by their ramp and the 2x
				// Lanczos reach (1.5 world px), and outside the entry corridor.
				ground := func(p connectionPoint) bool {
					lat, depth := compoundEntryLateral(t, l, p)
					return compoundKeepDistance(s.EdgeKeep, p) > compoundKeepRamp+1.5 && !(lat <= corridor+1.5 && depth < s.EdgeFeather+2)
				}
				edges := compoundBoundary(l)
				band, opaqueBefore, worst := 0, 0, uint8(0)
				for y := range l.Height * 2 {
					for x := range l.Width * 2 {
						p := connectionPoint{(float64(x) + .5) / 2, (float64(y) + .5) / 2}
						if d := compoundSignedDistance(l, edges, p); d < 0 || d > 2 || !ground(p) {
							continue
						}
						band++
						if before2.NRGBAAt(x, y).A > 64 {
							opaqueBefore++
						}
						worst = max(worst, art2.NRGBAAt(x, y).A)
					}
				}
				if worst > 64 {
					t.Errorf("ground within 2 world px of the boundary keeps alpha %d", worst)
				}
				// The 50 % contour, where the apron was cut opaque at the
				// boundary and stays opaque up to the contour.
				var depths []float64
				for _, c := range compoundFeatherContourSamples(l, s.EdgeFeather, before2, art2, ground) {
					depths = append(depths, c.depth)
				}
				mean, sd, lo, hi := compoundContourStats(t, depths)
				t.Logf("%s: %d 4x px kept; %d 2x ground px within 2 world px of the boundary (%d above alpha 64 before), max alpha %d; 50 %% contour over %d samples: mean %.2f, sd %.2f, range %.2f..%.2f world px", name, kept, band, opaqueBefore, worst, len(depths), mean, sd, lo, hi)
				if kept < 1000 || band < 500 || opaqueBefore < 50 || len(depths) < 5 {
					t.Fatalf("too few samples")
				}
				if lo < 2 || hi > s.EdgeFeather+1 {
					t.Errorf("50 %% contour outside the band: %.2f..%.2f", lo, hi)
				}
				pooled = append(pooled, depths...)
			})
		}
	}
	// Few stretches of these scenes reach the boundary opaque: the contour's
	// irregularity is judged over all of them.
	_, sd, lo, hi := compoundContourStats(t, pooled)
	t.Logf("50 %% contour over all %d samples: sd %.2f, range %.2f..%.2f world px", len(pooled), sd, lo, hi)
	if len(pooled) < 100 || sd < 1 || hi-lo < 3.5 {
		t.Errorf("50 %% contour too regular: %d samples, sd %.2f, range %.2f..%.2f", len(pooled), sd, lo, hi)
	}
}

// compoundContourSample is a point on the outer boundary and the depth (world
// px) at which the feathered 2x alpha first reaches 50 % inward from it (NaN
// when it never does).
type compoundContourSample struct {
	at    connectionPoint
	depth float64
}

// compoundFeatherContourSamples marches inward from points along the outer
// boundary where the unfeathered 2x apron was cut opaque at the boundary and
// stays ground (not kept, not the entry corridor) up to the contour.
func compoundFeatherContourSamples(l compoundLayout, width float64, before2, art2 *image.NRGBA, ground func(connectionPoint) bool) []compoundContourSample {
	var out []compoundContourSample
	edges := compoundBoundary(l)
	for _, e := range edges {
		v := e.b.sub(e.a)
		n := v.length()
		u := v.mul(1 / n)
		inward := connectionPoint{-u[1], u[0]}
		if mid := e.a.add(v.mul(.5)); compoundSignedDistance(l, edges, mid.add(inward.mul(1))) < 0 {
			inward = inward.mul(-1)
		}
	sample:
		for along := 3.; along <= n-3; along++ {
			p := e.a.add(u.mul(along))
			found := math.NaN()
			for d := 0.; d <= width+2; d += .25 {
				q := p.add(inward.mul(d))
				x, y := int(q[0]*2), int(q[1]*2)
				if !ground(q) || !image.Pt(x, y).In(before2.Bounds()) || (d >= .5 && before2.NRGBAAt(x, y).A < 128) {
					continue sample
				}
				if art2.NRGBAAt(x, y).A >= 128 {
					// Its distance to the boundary, which near a corner
					// is less than the depth along this normal.
					found = compoundSignedDistance(l, edges, q)
					break
				}
			}
			out = append(out, compoundContourSample{p, found})
		}
	}
	return out
}

// compoundContourStats returns mean, standard deviation and range of contour
// depths (world px), failing on a missing contour.
func compoundContourStats(t *testing.T, depths []float64) (mean, sd, lo, hi float64) {
	t.Helper()
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, d := range depths {
		if math.IsNaN(d) {
			t.Fatal("50 % contour missing")
		}
		mean += d
		lo, hi = min(lo, d), max(hi, d)
	}
	mean /= float64(len(depths))
	for _, d := range depths {
		sd += (d - mean) * (d - mean)
	}
	return mean, math.Sqrt(sd / float64(len(depths))), lo, hi
}
