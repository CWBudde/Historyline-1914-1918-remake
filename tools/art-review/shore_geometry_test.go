package artreview

import (
	"math"
	"testing"
)

// A seamless material can hide a nearly straight mask. Trace the actual zero
// crossing relative to the fixed port-to-port chord, independently of texture.
func TestShoreLongCoastsHaveBroadAsymmetricBends(t *testing.T) {
	for _, tc := range []struct {
		id   string
		land []int
		a, b connectionPoint
	}{
		{"lakeside12", []int{1, 2, 3}, connectionPoint{48, 0}, connectionPoint{48, 80}},
		{"lakeside16", []int{0, 1, 2}, connectionPoint{83.5, 60}, connectionPoint{12.5, 20}},
		{"lakeside34", []int{3, 4, 5}, connectionPoint{83.5, 60}, connectionPoint{12.5, 20}},
		{"lakeside56", []int{0, 1, 5}, connectionPoint{83.5, 20}, connectionPoint{12.5, 60}},
	} {
		t.Run(tc.id, func(t *testing.T) {
			v := shoreVariant{tc.id, tc.land}
			curve := shoreCurveFor(v)
			chord := tc.b.sub(tc.a)
			normal := connectionPoint{-chord[1], chord[0]}.mul(1 / chord.length())
			crossing := func(fraction float64) float64 {
				p := tc.a.add(chord.mul(fraction))
				lo, hi := -14., 14.
				start := shoreFieldWithCurve(v, p.add(normal.mul(lo)), &curve)
				end := shoreFieldWithCurve(v, p.add(normal.mul(hi)), &curve)
				if start*end >= 0 {
					t.Fatalf("missing single interior coast crossing at%g", fraction)
				}
				for range 45 {
					mid := (lo + hi) / 2
					if shoreFieldWithCurve(v, p.add(normal.mul(mid)), &curve)*start > 0 {
						lo = mid
					} else {
						hi = mid
					}
				}
				return (lo + hi) / 2
			}
			if bow := math.Abs(crossing(.5)); bow < 4 || bow > 10 {
				t.Fatalf("coast bow=%glogicalpx; want4..10", bow)
			}
			t.Logf("coast offsets at30/50/70%%: %.3f / %.3f / %.3f logicalpx", crossing(.3), crossing(.5), crossing(.7))
			if asymmetry := math.Abs(crossing(.3) - crossing(.7)); asymmetry < 1.2 {
				t.Fatalf("shoulders too symmetric: displacement difference%g", asymmetry)
			}
		})
	}
}

// All nine masks describe one shore arc. A stronger bend must not fold back
// into an isolated pond or island when the protected edge fields take over.
func TestShoreCurvedMasksKeepLandAndWaterConnected(t *testing.T) {
	for _, v := range shoreRoster() {
		t.Run(v.ID, func(t *testing.T) {
			curve := shoreCurveFor(v)
			var classes [80][96]int
			for y := range 80 {
				for x := range 96 {
					p := connectionPoint{float64(x) + .5, float64(y) + .5}
					if !connectionInside(sharedConnectionGeometry(), p) {
						continue
					}
					classes[y][x] = 1
					if shoreFieldWithCurve(v, p, &curve) < 0 {
						classes[y][x] = 2
					}
				}
			}
			for class := 1; class <= 2; class++ {
				var seen [80][96]bool
				components := 0
				for y := range 80 {
					for x := range 96 {
						if classes[y][x] != class || seen[y][x] {
							continue
						}
						components++
						queue := [][2]int{{x, y}}
						seen[y][x] = true
						for len(queue) > 0 {
							p := queue[0]
							queue = queue[1:]
							for dy := -1; dy <= 1; dy++ {
								for dx := -1; dx <= 1; dx++ {
									x1, y1 := p[0]+dx, p[1]+dy
									if x1 >= 0 && x1 < 96 && y1 >= 0 && y1 < 80 && !seen[y1][x1] && classes[y1][x1] == class {
										seen[y1][x1] = true
										queue = append(queue, [2]int{x1, y1})
									}
								}
							}
						}
					}
				}
				if components != 1 {
					t.Fatalf("material%d split into%d regions", class, components)
				}
			}
		})
	}
}
