package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path/filepath"
	"sort"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// The winter quays (tile/quay*_w) are the summer quays under snow: the same
// outline, a packed-snow deck in the winter road's colour with the two cart
// tracks kept as darker ruts, no summer green and no paint-white crown, and
// chains that join without a seam. They are judged at 4x as the pack grades
// them (terraingrade), over the graded winter water.

// quayDeck is the half-width of the sett deck between the kerbs, and of the
// road's carriageway, in world px.
const quayDeck = 11.

func loadQuayAsset(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	im, err := loadPNG(filepath.Join("..", "..", "assets", path))
	if err != nil {
		t.Fatal(err)
	}
	return im
}

// quayAcross is the signed distance of 4x pixel (x, y) from the centre line
// of axis id, in world px.
func quayAcross(id string, x, y int) float64 {
	u := quayAxes[id].mul(1 / quayAxes[id].length())
	q := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}.sub(connectionPoint{48, 40})
	return q[0]*-u[1] + q[1]*u[0]
}

// over composites a over the opaque b.
func over(a, b *image.NRGBA) *image.NRGBA {
	out := image.NewNRGBA(a.Rect)
	for i := 0; i < len(a.Pix); i += 4 {
		k := float64(a.Pix[i+3]) / 255
		for c := range 3 {
			out.Pix[i+c] = uint8(math.Round(float64(a.Pix[i+c])*k + float64(b.Pix[i+c])*(1-k)))
		}
		out.Pix[i+3] = 255
	}
	return out
}

// quayLabMean is the mean CIELAB colour of im's pixels where keep holds.
func quayLabMean(im *image.NRGBA, keep func(x, y int) bool) (l, a, b float64, n int) {
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			if !keep(x, y) {
				continue
			}
			pl, pa, pb := winterLab(im.NRGBAAt(x, y))
			l, a, b, n = l+pl, a+pa, b+pb, n+1
		}
	}
	if n > 0 {
		l, a, b = l/float64(n), a/float64(n), b/float64(n)
	}
	return l, a, b, n
}

// quayRuts finds the two cart tracks in the across-deck L* profile of the
// opaque deck of tile im on axis id: the profile is binned per 4x px across
// the axis and smoothed over 5 px, and each track is the trough around the
// lowest bin between 2 and 10 world px from the centre line. It returns the
// two positions (4x px from the centre line) and their depth below the median
// of their window.
func quayRuts(id string, im *image.NRGBA, alpha *image.NRGBA) (pos [2]int, depth [2]float64) {
	const half = 44
	var sum, cnt [2*half + 1]float64
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			if alpha.NRGBAAt(x, y).A < 250 {
				continue
			}
			bin := int(math.Round(quayAcross(id, x, y) * 4))
			if bin < -half || bin > half {
				continue
			}
			l, _, _ := winterLab(im.NRGBAAt(x, y))
			sum[bin+half] += l
			cnt[bin+half]++
		}
	}
	prof := func(bin int) float64 {
		var s, n float64
		for k := bin - 2; k <= bin+2; k++ {
			if k >= -half && k <= half && cnt[k+half] > 0 {
				s += sum[k+half]
				n += cnt[k+half]
			}
		}
		return s / n
	}
	for side, sign := range []int{-1, 1} {
		best, at := math.Inf(1), 0
		var window []float64
		for d := 8; d <= 40; d++ {
			v := prof(sign * d)
			window = append(window, v)
			if v < best {
				best, at = v, sign*d
			}
		}
		sort.Float64s(window)
		median := window[len(window)/2]
		// The tracks are flat-bottomed and uneven: the trough's position is
		// the middle of its full width at half depth below the deck beside
		// it, not its noisy lowest bin.
		cut := (best + window[len(window)-1]) / 2
		lo, hi := at, at
		for lo-1 >= -half && prof(lo-1) < cut {
			lo--
		}
		for hi+1 <= half && prof(hi+1) < cut {
			hi++
		}
		pos[side], depth[side] = (lo+hi)/2, median-best
	}
	return pos, depth
}

// quaySeam is the mean L* step where a chain of two quay tiles on axis id
// passes from one hex to the next, and the mean step along the chain
// elsewhere, on lines along the axis across the deck and the kerbs.
func quaySeam(id string, im *image.NRGBA) (seam, inner float64) {
	step := quayAxes[id]
	u := step.mul(1 / step.length())
	n := connectionPoint{-u[1], u[0]}
	at := func(p connectionPoint) (color.NRGBA, int) {
		for k, off := range []connectionPoint{{}, step} {
			q := p.sub(off).mul(4)
			x, y := int(math.Floor(q[0])), int(math.Floor(q[1]))
			if image.Pt(x, y).In(im.Rect) && im.NRGBAAt(x, y).A > 0 {
				return im.NRGBAAt(x, y), k
			}
		}
		return color.NRGBA{}, -1
	}
	var ss, si float64
	var ns, ni int
	for across := -16.; across <= 16; across++ {
		var prevL float64
		prev := -1
		for d := 0.; d <= step.length(); d += .25 {
			c, k := at(connectionPoint{48, 40}.add(u.mul(d)).add(n.mul(across)))
			if k < 0 {
				prev = -1
				continue
			}
			l, _, _ := winterLab(c)
			if prev >= 0 {
				if k != prev {
					ss, ns = ss+math.Abs(l-prevL), ns+1
				} else {
					si, ni = si+math.Abs(l-prevL), ni+1
				}
			}
			prevL, prev = l, k
		}
	}
	return ss / float64(ns), si / float64(ni)
}

func TestQuayWinterTiles(t *testing.T) {
	water := terraingrade.Grade(loadQuayAsset(t, "terrain/water/water_w-4x.png"), "tile/water_w", "terrain/water", nil)
	road := terraingrade.Grade(loadQuayAsset(t, "terrain/road/road36_w-4x.png"), "tile/road36_w", "terrain/road", nil)
	rl, ra, rb, _ := quayLabMean(road, func(x, y int) bool {
		return road.NRGBAAt(x, y).A >= 250 && math.Abs(quayAcross("quay36", x, y)) <= quayDeck
	})
	for _, id := range []string{"quay14", "quay25", "quay36"} {
		t.Run(id, func(t *testing.T) {
			summer := loadQuayAsset(t, fmt.Sprintf("terrain/quay/%s-4x.png", id))
			winter := loadQuayAsset(t, fmt.Sprintf("terrain/quay/%s_w-4x.png", id))

			// The outline is the summer quay's.
			var and, or int
			for i := 3; i < len(summer.Pix); i += 4 {
				s, w := summer.Pix[i] > 128, winter.Pix[i] > 128
				if s && w {
					and++
				}
				if s || w {
					or++
				}
			}
			iou := float64(and) / float64(or)

			ws := over(terraingrade.Grade(winter, "tile/"+id+"_w", "terrain/quay", nil), water)
			ss := over(terraingrade.Grade(summer, "tile/"+id, "terrain/quay", nil), water)

			// The deck is in the winter road's colour.
			dl, da, db, _ := quayLabMean(ws, func(x, y int) bool {
				return winter.NRGBAAt(x, y).A >= 250 && math.Abs(quayAcross(id, x, y)) <= quayDeck
			})
			dE := math.Sqrt((dl-rl)*(dl-rl) + (da-ra)*(da-ra) + (db-rb)*(db-rb))

			// The cart tracks stay where the summer tracks are, as ruts.
			wpos, wdepth := quayRuts(id, ws, winter)
			spos, sdepth := quayRuts(id, ss, summer)

			// No summer green, no paint-white crown.
			var opaque, green, middle, white int
			for y := range ws.Rect.Dy() {
				for x := range ws.Rect.Dx() {
					if winter.NRGBAAt(x, y).A <= 128 {
						continue
					}
					c := ws.NRGBAAt(x, y)
					opaque++
					if summerGreen(c) {
						green++
					}
					if math.Abs(quayAcross(id, x, y)) <= 4 {
						middle++
						if l, a, b := winterLab(c); l > 90 && math.Hypot(a, b) < 6 {
							white++
						}
					}
				}
			}

			seam, inner := quaySeam(id, winter)
			sseam, sinner := quaySeam(id, summer)

			t.Logf("IoU %.4f; deck L*a*b* %.1f,%.1f,%.1f vs road %.1f,%.1f,%.1f: dE %.1f; ruts %v depth %.1f/%.1f (summer %v depth %.1f/%.1f); green %.2f%%; white %d of %d; seam %.2f inner %.2f (summer %.2f, %.2f)",
				iou, dl, da, db, rl, ra, rb, dE, wpos, wdepth[0], wdepth[1], spos, sdepth[0], sdepth[1],
				100*float64(green)/float64(opaque), white, middle, seam, inner, sseam, sinner)
			if iou < .98 {
				t.Errorf("alpha IoU %.4f, want >= 0.98", iou)
			}
			if dE > 8 {
				t.Errorf("deck dE %.1f from the winter road, want <= 8", dE)
			}
			for i := range wpos {
				if math.Abs(float64(wpos[i]-spos[i])) > 2 || wdepth[i] < 2 {
					t.Errorf("rut %d at %d px depth %.1f, summer track at %d px", i, wpos[i], wdepth[i], spos[i])
				}
			}
			if float64(green) >= .01*float64(opaque) {
				t.Errorf("summer green %d of %d opaque px", green, opaque)
			}
			if white > 0 {
				t.Errorf("%d paint-white px in the middle 8 world px", white)
			}
			if seam > 2*inner+1 {
				t.Errorf("seam step %.2f against %.2f inside the chain", seam, inner)
			}
		})
	}
}

// The selected winter pier is a strip the export can make periodic, like the
// summer one.
func TestQuayWinterStripIsPeriodic(t *testing.T) {
	im, err := loadPNG(filepath.Join("..", "..", "art", "terrain", "quay", "quay-pier-winter-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := newQuayStrip(im)
	if err != nil {
		t.Fatal(err)
	}
	for id, step := range quayAxes {
		if _, err := quayTiles(s, id); err != nil {
			t.Fatal(err)
		}
		p := step.length() * s.k
		for _, phase := range []float64{0, p / 8, p / 2, p - .5} {
			a, b := s.periodic(phase, p, s.centre), s.periodic(phase+p, p, s.centre)
			for i := range a {
				if math.Abs(a[i]-b[i]) > 1e-9 {
					t.Fatalf("%s phase %g: %v vs %v", id, phase, a, b)
				}
			}
		}
	}
}
