// Command unit-squad composes a unit atlas that shows n of the soldiers of a
// source atlas: every strength of an infantry group gets its own sprite, with as
// many men as it has strength left.
//
//	go run ./tools/unit-squad -src art/units/infantry/riflemen.png -n 4 -out art/units/infantry-4/atlas.png
//
// The source is a unit atlas (1024x1536, 2x3 cells, headings N, NW, SW, S, SE,
// NE) with exactly three separate men per cell. Each man is cut out of his cell,
// scaled by -scale and set into a formation that turns with the cell's heading,
// centred on the atlas anchor (256,261) of the cell. The result goes through the
// usual `art-review -unit` export.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

const (
	cells    = 6
	men      = 3   // men per source cell
	minArea  = 800 // smallest component that counts as a man
	cutGrow  = 12  // px around a man's silhouette that stay with him (his glow)
	anchorX  = 256
	anchorY  = 261
	across   = 135.0 // distance between neighbours in a rank, atlas px at scale 1
	along    = 150.0 // distance between ranks
	maxCount = 6
)

// slot is a man's place in the formation: u to the right of the heading, v
// ahead of the middle, in atlas px at the formation's own scale.
type slot struct{ u, v float64 }

// formation returns the places of n men, ranks across the heading, the middle
// of a rank a little off its neighbours so the group does not look drilled.
func formation(n int) []slot {
	a, l := across, along
	switch n {
	case 1:
		return []slot{{0, 0}}
	case 2:
		return []slot{{-a / 2, 10}, {a / 2, -10}}
	case 3:
		return []slot{{-a, 12}, {0, -12}, {a, 12}}
	case 4:
		return []slot{{-a / 2, -l/2 - 6}, {a / 2, -l / 2}, {-a / 2, l / 2}, {a / 2, l/2 + 6}}
	case 5:
		return []slot{{-a / 2, -l / 2}, {a / 2, -l/2 - 8}, {-a, l / 2}, {0, l/2 + 8}, {a, l / 2}}
	case 6:
		return []slot{{-a, -l/2 + 6}, {0, -l/2 - 10}, {a, -l/2 + 2}, {-a, l/2 - 4}, {0, l/2 + 10}, {a, l/2 - 2}}
	}
	return nil
}

// man is one cut-out soldier: his pixels and his centroid in them.
type man struct {
	img    *image.NRGBA
	cx, cy float64
}

// headings are the compass headings of the six cells, degrees clockwise from up.
var headings = [cells]float64{0, 300, 240, 180, 120, 60}

// cutMen separates the men of one cell: the 8-connected groups of pixels with
// alpha of at least 128 (want of them), each with the pixels within cutGrow of it.
func cutMen(cell *image.NRGBA, want int) ([]man, error) {
	b := cell.Bounds()
	w, h := b.Dx(), b.Dy()
	label := make([]int, w*h)
	var groups [][]int
	for i := range label {
		if label[i] != 0 || cell.Pix[i*4+3] < 128 {
			continue
		}
		id := len(groups) + 1
		stack, group := []int{i}, []int(nil)
		label[i] = id
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			group = append(group, p)
			px, py := p%w, p/w
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					x, y := px+dx, py+dy
					if x < 0 || y < 0 || x >= w || y >= h {
						continue
					}
					q := y*w + x
					if label[q] == 0 && cell.Pix[q*4+3] >= 128 {
						label[q] = id
						stack = append(stack, q)
					}
				}
			}
		}
		groups = append(groups, group)
	}
	groups = slices.DeleteFunc(groups, func(g []int) bool { return len(g) < minArea })
	if len(groups) != want {
		return nil, fmt.Errorf("cell has %d men, want %d", len(groups), want)
	}
	var out []man
	for _, g := range groups {
		x0, y0, x1, y1 := w, h, 0, 0
		var sx, sy float64
		for _, p := range g {
			x, y := p%w, p/w
			x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x+1), max(y1, y+1)
			sx, sy = sx+float64(x), sy+float64(y)
		}
		r := image.Rect(max(x0-cutGrow, 0), max(y0-cutGrow, 0), min(x1+cutGrow, w), min(y1+cutGrow, h))
		m := image.NewNRGBA(r)
		for _, p := range g {
			px, py := p%w, p/w
			for y := max(py-cutGrow, r.Min.Y); y <= min(py+cutGrow, r.Max.Y-1); y++ {
				for x := max(px-cutGrow, r.Min.X); x <= min(px+cutGrow, r.Max.X-1); x++ {
					if (x-px)*(x-px)+(y-py)*(y-py) <= cutGrow*cutGrow {
						m.SetNRGBA(x, y, cell.NRGBAAt(b.Min.X+x, b.Min.Y+y))
					}
				}
			}
		}
		n := float64(len(g))
		out = append(out, man{m, sx / n, sy / n})
	}
	// left to right, so the pose order is the same in every run
	slices.SortFunc(out, func(a, b man) int {
		switch {
		case a.cx < b.cx:
			return -1
		case a.cx > b.cx:
			return 1
		}
		return 0
	})
	return out, nil
}

// compose draws n men of pool, scaled by scale, onto a fresh cell of the given
// size, in the formation turned to heading degrees.
func compose(pool []man, n int, heading, scale float64, size int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	th := heading * math.Pi / 180
	fwd := [2]float64{math.Sin(th), -math.Cos(th)}
	right := [2]float64{math.Cos(th), math.Sin(th)}
	type placed struct {
		m    man
		x, y float64
		z    float64
	}
	poses := [maxCount]int{1, 0, 2, 2, 1, 0}
	var ps []placed
	for i, s := range formation(n) {
		x := float64(anchorX) + scale*(s.u*right[0]+s.v*fwd[0])
		y := float64(anchorY) + scale*(s.u*right[1]+s.v*fwd[1])
		ps = append(ps, placed{pool[poses[i]%len(pool)], x, y, y})
	}
	// the men lower on the screen stand in front
	slices.SortStableFunc(ps, func(a, b placed) int {
		switch {
		case a.z < b.z:
			return -1
		case a.z > b.z:
			return 1
		}
		return 0
	})
	for _, p := range ps {
		m := p.m
		// destination = scale*(source - centroid) + position
		aff := f64.Aff3{scale, 0, p.x - scale*m.cx, 0, scale, p.y - scale*m.cy}
		layer := image.NewNRGBA(dst.Bounds())
		xdraw.CatmullRom.Transform(layer, aff, m.img, m.img.Bounds(), xdraw.Over, nil)
		draw.Draw(dst, dst.Bounds(), layer, image.Point{}, draw.Over)
	}
	return dst
}

// shrink scales every cell of src about the atlas anchor, the whole subject at once.
func shrink(src *image.NRGBA, scale float64) *image.NRGBA {
	out := image.NewNRGBA(src.Bounds())
	size := src.Bounds().Dx() / 2
	for f := range cells {
		rect := image.Rect(f%2*size, f/2*size, (f%2+1)*size, (f/2+1)*size)
		cell := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.Draw(cell, cell.Bounds(), src, rect.Min, draw.Src)
		layer := image.NewNRGBA(cell.Bounds())
		aff := f64.Aff3{scale, 0, anchorX * (1 - scale), 0, scale, anchorY * (1 - scale)}
		xdraw.CatmullRom.Transform(layer, aff, cell, cell.Bounds(), xdraw.Src, nil)
		draw.Draw(out, rect, layer, image.Point{}, draw.Src)
	}
	return out
}

// atlas builds the atlas of n men from src.
func atlas(src *image.NRGBA, n int, scale float64) (*image.NRGBA, error) {
	if n == 0 {
		return shrink(src, scale), nil
	}
	if n < 1 || n > maxCount {
		return nil, fmt.Errorf("n must be 1..%d", maxCount)
	}
	b := src.Bounds()
	if b.Dx()%2 != 0 || b.Dy() != b.Dx()/2*3 {
		return nil, errors.New("source is not a 2x3 atlas")
	}
	size := b.Dx() / 2
	out := image.NewNRGBA(b)
	for f := range cells {
		rect := image.Rect(f%2*size, f/2*size, (f%2+1)*size, (f/2+1)*size)
		cell := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.Draw(cell, cell.Bounds(), src, rect.Min, draw.Src)
		pool, err := cutMen(cell, men)
		if err != nil {
			return nil, fmt.Errorf("cell %d: %w", f, err)
		}
		draw.Draw(out, rect, compose(pool, n, headings[f], scale, size), image.Point{}, draw.Src)
	}
	return out, nil
}

func main() {
	src := flag.String("src", "", "source atlas with three men per cell")
	n := flag.Int("n", 6, "men per cell, 1..6; 0 scales the source cells as they are")
	out := flag.String("out", "", "atlas to write")
	scale := flag.Float64("scale", .52, "size of a man relative to the source")
	lone := flag.Bool("gunner", false, "first drop the loader of a two-man team (use with -n 0)")
	flag.Parse()
	if *src == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	f, err := os.Open(*src)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	im, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	nrgba := image.NewNRGBA(im.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), im, im.Bounds().Min, draw.Src)
	if *lone {
		if nrgba, err = gunners(nrgba); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	res, err := atlas(nrgba, *n, *scale)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	o, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := png.Encode(o, res); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := o.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// gunner drops the second man of a two-man team cell: the coats (the magenta
// pixels, two blobs) are found, every other pixel goes to the nearest coat, and
// the coat with the smaller share (the loader, whose share is his box and belt)
// is erased with its pixels. The gun stays with the gunner.
func gunner(cell *image.NRGBA) (*image.NRGBA, error) {
	b := cell.Bounds()
	w, h := b.Dx(), b.Dy()
	coat := func(p color.NRGBA) bool { return p.A >= 128 && int(p.R)-int(p.G) > 90 && int(p.B)-int(p.G) > 40 }
	label := make([]int, w*h)
	var blobs [][]int
	for i := range label {
		if label[i] != 0 || !coat(cell.NRGBAAt(i%w, i/w)) {
			continue
		}
		id := len(blobs) + 1
		stack, blob := []int{i}, []int(nil)
		label[i] = id
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			blob = append(blob, p)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					x, y := p%w+dx, p/w+dy
					if x >= 0 && y >= 0 && x < w && y < h && label[y*w+x] == 0 && coat(cell.NRGBAAt(x, y)) {
						label[y*w+x] = id
						stack = append(stack, y*w+x)
					}
				}
			}
		}
		blobs = append(blobs, blob)
	}
	slices.SortFunc(blobs, func(a, b []int) int { return len(b) - len(a) })
	blobs = slices.DeleteFunc(blobs, func(g []int) bool { return len(g) < minArea })
	if len(blobs) < 2 {
		return nil, fmt.Errorf("cell has %d coats, want 2", len(blobs))
	}
	blobs = blobs[:2]
	centre := func(g []int) (float64, float64) {
		var sx, sy float64
		for _, p := range g {
			sx, sy = sx+float64(p%w), sy+float64(p/w)
		}
		return sx / float64(len(g)), sy / float64(len(g))
	}
	var cx, cy [2]float64
	for i := range 2 {
		cx[i], cy[i] = centre(blobs[i])
	}
	nearest := func(x, y int) int {
		d0 := math.Hypot(float64(x)-cx[0], float64(y)-cy[0])
		d1 := math.Hypot(float64(x)-cx[1], float64(y)-cy[1])
		if d1 < d0 {
			return 1
		}
		return 0
	}
	var share [2]int
	for y := range h {
		for x := range w {
			if cell.NRGBAAt(x, y).A >= 128 {
				share[nearest(x, y)]++
			}
		}
	}
	keep := 0
	if share[1] > share[0] {
		keep = 1
	}
	out := image.NewNRGBA(b)
	for y := range h {
		for x := range w {
			if nearest(x, y) == keep {
				out.SetNRGBA(x, y, cell.NRGBAAt(x, y))
			}
		}
	}
	return out, nil
}

// gunners applies gunner to every cell of the atlas src.
func gunners(src *image.NRGBA) (*image.NRGBA, error) {
	size := src.Bounds().Dx() / 2
	out := image.NewNRGBA(src.Bounds())
	for f := range cells {
		rect := image.Rect(f%2*size, f/2*size, (f%2+1)*size, (f/2+1)*size)
		cell := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.Draw(cell, cell.Bounds(), src, rect.Min, draw.Src)
		g, err := gunner(cell)
		if err != nil {
			return nil, fmt.Errorf("cell %d: %w", f, err)
		}
		draw.Draw(out, rect, g, image.Point{}, draw.Src)
	}
	return out, nil
}
