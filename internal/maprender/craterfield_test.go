package maprender

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// craterMonths are the months the lattice measure runs on.
var craterMonths = []string{"1914-08", "1914-10", "1915-12", "1916-03", "1916-05", "1916-07", "1917-11", "1918-01"}

// craterLattice are the shortest periods of the hex lattice, in world pixels.
var craterLattice = [3]image.Point{{142, 0}, {71, 40}, {0, 80}}

// craterInset keeps the measure away from a field's edge, where the tiles fade
// to their neighbours by design: world pixels from any other ground.
const craterInset = 20

// The limits of the lattice measure. A field of 10,000 pixel pairs must meet
// L1.1's lattice limit; a few hexes hold too few independent hollows, so a
// correlation of chance is allowed them. craterMinSigma is above the 13.4 to 14.7
// of the two-coset average and at about 0.8 of the masters' 22.6 (summer, graded)
// and 0.7 of 25.5 (winter).
const (
	craterMaxPooled = 0.10
	craterMaxLarge  = 0.15
	craterMaxSmall  = 0.35
	craterMinSigma  = 17.5
	// The median L* step, snow side less rough side, across a border of stamp and
	// rough ground. The notch measured 5.3 to 7.4 on line 14, column 11, and 4.0 to
	// 7.4 on the 24 sides of all months.
	craterMaxStep    = 2.5
	craterMaxStepAll = 3.6
)

func shippedLevel(t *testing.T, month string) *level.Level {
	t.Helper()
	lvl, err := scenario.LoadLevel(filepath.Join("..", "..", "levels", month), scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	return lvl
}

func shippedRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := New(os.DirFS(filepath.Join("..", "..", "assets")))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// craterMask marks the world pixels of unstamped crater ground without
// infrastructure that lie at least craterInset from any other ground.
func craterMask(r *Renderer, lvl *level.Level) (mask []bool, w, h, hexes int) {
	grid := lvl.Grid()
	scene := mapcompose.ComposeSeason(grid, mapcompose.LevelHex(lvl), func(id string) bool { return r.has[id] }, mapcompose.SeasonAt(lvl))
	field := map[hexgrid.Coord]bool{}
	for line, cells := range scene.Cells {
		for col, cell := range cells {
			h := lvl.Lines[line][col]
			if cell.Cluster == 0 || scene.Clusters[cell.Cluster-1].Stamp != "" || len(h.Infra.IDs) != 0 || (h.Terrain != "mud" && h.Terrain != "shelled_w") {
				continue
			}
			field[hexgrid.Coord{Line: line, Col: col}] = true
		}
	}
	w, h = Size(grid)
	dist := make([]int, w*h)
	for y := range h {
		for x := range w {
			if c, ok := grid.PixelToHex(float64(x)+.5, float64(y)+.5); ok && field[c] {
				dist[y*w+x] = 1 << 20
			}
		}
	}
	// A 3-4 chamfer distance from the field's complement.
	at := func(x, y int) int {
		if x < 0 || y < 0 || x >= w || y >= h {
			return 0
		}
		return dist[y*w+x]
	}
	for y := range h {
		for x := range w {
			if d := dist[y*w+x]; d > 0 {
				dist[y*w+x] = min(d, at(x-1, y)+3, at(x, y-1)+3, at(x-1, y-1)+4, at(x+1, y-1)+4)
			}
		}
	}
	for y := h - 1; y >= 0; y-- {
		for x := w - 1; x >= 0; x-- {
			if d := dist[y*w+x]; d > 0 {
				dist[y*w+x] = min(d, at(x+1, y)+3, at(x, y+1)+3, at(x+1, y+1)+4, at(x-1, y+1)+4)
			}
		}
	}
	mask = make([]bool, w*h)
	for i, d := range dist {
		mask[i] = d >= craterInset*3
	}
	return mask, w, h, len(field)
}

// craterMeasure is the crater ground of one render.
type craterMeasure struct {
	hexes, pixels int
	sigma, mean   float64
	r             [3]float64 // Pearson correlation of L* at each lattice period
	pairs         [3]int
}

func measureCraters(im *image.NRGBA, density int, mask []bool, w, h, hexes int) craterMeasure {
	m := craterMeasure{hexes: hexes}
	lum := make([]float64, w*h)
	var sum, sum2 float64
	for y := range h {
		for x := range w {
			if !mask[y*w+x] {
				continue
			}
			var v float64
			for dy := range density {
				for dx := range density {
					p := pixelLStar(im, x*density+dx, y*density+dy)
					v += p
					sum += p
					sum2 += p * p
				}
			}
			lum[y*w+x] = v / float64(density*density)
			m.pixels++
		}
	}
	if m.pixels > 0 {
		n := float64(m.pixels * density * density)
		m.mean = sum / n
		m.sigma = math.Sqrt(math.Max(0, sum2/n-m.mean*m.mean))
	}
	for i, off := range craterLattice {
		var sx, sy, sxx, syy, sxy float64
		for y := range h - off.Y {
			for x := range w - off.X {
				if !mask[y*w+x] || !mask[(y+off.Y)*w+x+off.X] {
					continue
				}
				a, b := lum[y*w+x], lum[(y+off.Y)*w+x+off.X]
				m.pairs[i]++
				sx += a
				sy += b
				sxx += a * a
				syy += b * b
				sxy += a * b
			}
		}
		m.r[i] = pearsonOf(m.pairs[i], sx, sy, sxx, syy, sxy)
	}
	return m
}

func pearsonOf(n int, sx, sy, sxx, syy, sxy float64) float64 {
	if n == 0 {
		return math.NaN()
	}
	f := float64(n)
	d := math.Sqrt((sxx - sx*sx/f) * (syy - sy*sy/f))
	if d == 0 {
		return math.NaN()
	}
	return (sxy - sx*sy/f) / d
}

// pixelLStar is the L* of a pixel of im.
func pixelLStar(im *image.NRGBA, x, y int) float64 {
	c := im.NRGBAAt(x, y)
	f := func(v uint8) float64 {
		u := float64(v) / 255
		if u <= 0.04045 {
			return u / 12.92
		}
		return math.Pow((u+0.055)/1.055, 2.4)
	}
	Y := 0.2126*f(c.R) + 0.7152*f(c.G) + 0.0722*f(c.B)
	if Y > 0.008856 {
		return 116*math.Cbrt(Y) - 16
	}
	return 903.3 * Y
}

// TestCraterFieldsDoNotRepeatPerHex measures the crater ground of the shipped
// months as rendered: its L* at one lattice period and its contrast.
func TestCraterFieldsDoNotRepeatPerHex(t *testing.T) {
	r := shippedRenderer(t)
	var pooled, pairs [3]float64
	var sigmaSum float64
	var sigmaN int
	for _, month := range craterMonths {
		lvl := shippedLevel(t, month)
		mask, w, h, hexes := craterMask(r, lvl)
		m, err := r.Render(lvl, 2, false)
		if err != nil {
			t.Fatal(err)
		}
		cm := measureCraters(m.Image, 2, mask, w, h, hexes)
		t.Logf("%s: %4d hexes %7d px  mean L* %5.1f  sigma L* %5.1f  r(142,0) %+.2f [%d]  r(71,40) %+.2f [%d]  r(0,80) %+.2f [%d]", month, cm.hexes, cm.pixels, cm.mean, cm.sigma, cm.r[0], cm.pairs[0], cm.r[1], cm.pairs[1], cm.r[2], cm.pairs[2])
		for i := range craterLattice {
			switch {
			case cm.pairs[i] >= 10000 && math.Abs(cm.r[i]) > craterMaxLarge:
				t.Errorf("%s: r%v %.2f, want at most %.2f", month, craterLattice[i], cm.r[i], craterMaxLarge)
			case cm.pairs[i] > 0 && math.Abs(cm.r[i]) > craterMaxSmall:
				t.Errorf("%s: r%v %.2f over %d pairs, want at most %.2f", month, craterLattice[i], cm.r[i], cm.pairs[i], craterMaxSmall)
			}
		}
		if cm.pixels > 0 && cm.sigma < craterMinSigma {
			t.Errorf("%s: sigma L* %.1f, want at least %.1f", month, cm.sigma, craterMinSigma)
		}
		for i := range craterLattice {
			if cm.pairs[i] > 0 {
				pooled[i] += cm.r[i] * float64(cm.pairs[i])
				pairs[i] += float64(cm.pairs[i])
			}
		}
		sigmaSum += cm.sigma * float64(cm.pixels)
		sigmaN += cm.pixels
	}
	var worst float64
	for i, off := range craterLattice {
		p := pooled[i] / pairs[i]
		t.Logf("all: r(%d,%d) %+.2f over %.0f pairs", off.X, off.Y, p, pairs[i])
		worst = math.Max(worst, math.Abs(p))
	}
	t.Logf("all: sigma L* %.1f", sigmaSum/float64(sigmaN))
	// Before the rework every lattice period measured 1.00 at sigma 13.4 to 14.7.
	if worst > craterMaxPooled {
		t.Errorf("pooled lattice correlation %.2f, want at most %.2f", worst, craterMaxPooled)
	}
}

// edgeStep is the median L* just inside hex at's side in direction port less
// the median just outside, in map pixels at density: a white notch is a large
// positive step.
func edgeStep(im *image.NRGBA, density int, at hexgrid.Coord, port int) float64 {
	verts := [6][2]float64{{-23, -40}, {23, -40}, {48, 0}, {23, 40}, {-23, 40}, {-48, 0}}
	cx, cy := hexgrid.Center(at)
	e := (port + 1) % 6
	a, b := verts[e], verts[(e+1)%6]
	ex, ey := b[0]-a[0], b[1]-a[1]
	length := math.Hypot(ex, ey)
	nx, ny := ey/length, -ex/length // outward: the corners run clockwise
	sample := func(from, to float64) []float64 {
		var v []float64
		for s := 6.; s <= length-6; s += .5 {
			for d := from; d <= to; d += .5 {
				x := cx + a[0] + ex*s/length + nx*d
				y := cy + a[1] + ey*s/length + ny*d
				px, py := int(x*float64(density)), int(y*float64(density))
				if image.Pt(px, py).In(im.Bounds()) {
					v = append(v, pixelLStar(im, px, py))
				}
			}
		}
		slices.Sort(v)
		return v
	}
	inside, outside := sample(-5, -1), sample(1, 5)
	if len(inside) == 0 || len(outside) == 0 {
		return 0
	}
	return inside[len(inside)/2] - outside[len(outside)/2]
}

// TestWinterCraterStampLeavesNoNotch measures the step across a winter crater
// stamp's sides that face rough ground: the worst on 1914-10's line 14, column
// 11, and over the months with such a stamp.
func TestWinterCraterStampLeavesNoNotch(t *testing.T) {
	r := shippedRenderer(t)
	lvl := shippedLevel(t, "1914-10")
	m, err := r.Render(lvl, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	at := hexgrid.Coord{Line: 14, Col: 11}
	if got := lvl.Lines[at.Line][at.Col].Terrain; got != "shelled_w" {
		t.Fatalf("%v is %s", at, got)
	}
	worst := 0.
	for port, n := range hexgrid.Around(at) {
		if !lvl.Grid().Contains(n) || lvl.Lines[n.Line][n.Col].Terrain != "rough_w" {
			continue
		}
		step := edgeStep(m.Image, 2, at, port)
		t.Logf("%v side %d: step %+.1f", at, port, step)
		worst = math.Max(worst, step)
	}
	if worst > craterMaxStep {
		t.Errorf("%v: step %.1f across rough ground, want at most %.1f", at, worst, craterMaxStep)
	}

	entries, err := os.ReadDir(filepath.Join("..", "..", "levels"))
	if err != nil {
		t.Fatal(err)
	}
	all, sides := 0., 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		lvl, err := scenario.LoadLevel(filepath.Join("..", "..", "levels", e.Name()), scenario.CP)
		if err != nil {
			continue
		}
		var border [][2]hexgrid.Coord
		scene := mapcompose.ComposeSeason(lvl.Grid(), mapcompose.LevelHex(lvl), func(id string) bool { return r.has[id] }, mapcompose.SeasonAt(lvl))
		for _, cl := range scene.Clusters {
			if cl.Kind != "crater_w" || cl.Stamp == "" {
				continue
			}
			for _, c := range cl.Cells {
				for _, n := range hexgrid.Around(c) {
					if lvl.Grid().Contains(n) && lvl.Lines[n.Line][n.Col].Terrain == "rough_w" {
						border = append(border, [2]hexgrid.Coord{c, n})
					}
				}
			}
		}
		if len(border) == 0 {
			continue
		}
		m, err := r.Render(lvl, 2, false)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range border {
			for port, n := range hexgrid.Around(b[0]) {
				if n != b[1] {
					continue
				}
				step := edgeStep(m.Image, 2, b[0], port)
				t.Logf("%s %v side %d: step %+.1f", e.Name(), b[0], port, step)
				all = math.Max(all, step)
				sides++
			}
		}
	}
	t.Logf("%d sides, worst %.1f", sides, all)
	if sides == 0 {
		t.Error("no winter crater stamp beside rough ground")
	}
	if all > craterMaxStepAll {
		t.Errorf("step %.1f on a border, want at most %.1f", all, craterMaxStepAll)
	}
}
