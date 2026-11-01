// Package seasonblend softens the borders between hexes a level draws in
// different seasons (the scenario's per-hex seasons). Every season present is
// composed on the whole map once (Scenes), and a weight field over the map
// gives each point its share of every season: the seasons' hex centres spread
// by a smooth kernel, their border warped by noise so it does not follow the
// hex outline. Near a border (Seasons) the drawing mixes a hex's layers in each
// season by the weights at each pixel (Mix), and its scattered items each take
// the season their point falls to (Scattered). A level in one season has no
// Blend and draws as composed.
package seasonblend

import (
	"image"
	"image/draw"
	"math"
	"runtime"
	"slices"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// The field, in logical map pixels. A hex centre's season spreads over radius
// (neighbouring centres lie about 80 apart), the point is warped by two noise
// octaves before the kernel is applied, and the kernel's sum is raised to the
// power 1.5 to narrow the band where the seasons mix (weights).
const (
	radius       = 112.0
	warp1, wave1 = 30.0, 110.0
	warp2, wave2 = 9.0, 30.0
	// patches scales the fine noise, of wavelength patch, that breaks the band
	// into patches.
	patches, patch = 1.6, 14.0
	// extent bounds the layers drawn for a hex around its centre: the standing
	// items' 128x112 canvas.
	extentX, extentY = 64.0, 56.0
	// least is the weight below which a season is left out of a hex's mix.
	least = 1.0 / 512
)

// order gives each season its own patch noise.
var order = []mapcompose.Season{mapcompose.Summer, mapcompose.Autumn, mapcompose.Winter, mapcompose.Spring}

// Blend is the season field of a level and its scenes per season.
type Blend struct {
	// Scenes is the level composed with every hex in each season present.
	Scenes map[mapcompose.Season]*mapcompose.Scene
	grid   hexgrid.Grid
	season func(hexgrid.Coord) mapcompose.Season
	border map[hexgrid.Coord]*field
}

// step is the spacing, in logical map pixels, of the weights sampled around a
// border hex; Mix interpolates between them.
const step = 2.0

// field is the weights around a border hex: seasons, its own first, sampled
// on nx by ny points step apart from (x0, y0) over the hex's extent, the
// weights of a point in order of seasons.
type field struct {
	seasons []mapcompose.Season
	x0, y0  float64
	nx, ny  int
	w       []float32
}

// New is the blend of l, its scenes composed with the IDs has; nil when every
// hex of l is drawn in the same season.
func New(l *level.Level, has func(string) bool) *Blend {
	g := l.Grid()
	at := mapcompose.SeasonAt(l)
	var present []mapcompose.Season
	each(g, func(c hexgrid.Coord) {
		if s := at(c); !slices.Contains(present, s) {
			present = append(present, s)
		}
	})
	if len(present) < 2 {
		return nil
	}
	b := &Blend{Scenes: map[mapcompose.Season]*mapcompose.Scene{}, grid: g, season: at, border: map[hexgrid.Coord]*field{}}
	for _, s := range present {
		scene := mapcompose.ComposeSeason(g, mapcompose.LevelHexIn(l, s), has, mapcompose.Everywhere(s))
		b.Scenes[s] = &scene
	}
	// The hexes another season may reach, then their fields, in parallel.
	reach := radius + warp1 + warp2 + math.Hypot(extentX, extentY)
	type candidate struct {
		at     hexgrid.Coord
		near   []mapcompose.Season
		result *field
	}
	var candidates []candidate
	each(g, func(c hexgrid.Coord) {
		cx, cy := hexgrid.Center(c)
		near := []mapcompose.Season{at(c)}
		b.around(cx, cy, reach, func(h hexgrid.Coord, _ float64) {
			if s := at(h); !slices.Contains(near, s) {
				near = append(near, s)
			}
		})
		if len(near) > 1 {
			candidates = append(candidates, candidate{at: c, near: near})
		}
	})
	next := make(chan int)
	var wg sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		wg.Go(func() {
			for i := range next {
				candidates[i].result = b.sample(candidates[i].at, candidates[i].near)
			}
		})
	}
	for i := range candidates {
		next <- i
	}
	close(next)
	wg.Wait()
	for _, c := range candidates {
		if c.result != nil {
			b.border[c.at] = c.result
		}
	}
	return b
}

// sample is the field of hex c among the seasons near it, its own first,
// without the seasons that reach none of its points; nil where only its own
// does.
func (b *Blend) sample(c hexgrid.Coord, near []mapcompose.Season) *field {
	cx, cy := hexgrid.Center(c)
	f := &field{seasons: near, x0: cx - extentX, y0: cy - extentY, nx: int(2*extentX/step) + 1, ny: int(2*extentY/step) + 1}
	f.w = make([]float32, f.nx*f.ny*len(near))
	w := make([]float64, len(near))
	reached := make([]bool, len(near))
	reached[0] = true
	for j := range f.ny {
		for i := range f.nx {
			b.weights(f.x0+float64(i)*step, f.y0+float64(j)*step, near, w)
			for s, v := range w {
				f.w[(j*f.nx+i)*len(near)+s] = float32(v)
				reached[s] = reached[s] || v > least
			}
		}
	}
	var seasons []mapcompose.Season
	for s, ok := range reached {
		if ok {
			seasons = append(seasons, near[s])
		}
	}
	switch {
	case len(seasons) < 2:
		return nil
	case len(seasons) < len(near):
		return b.sample(c, seasons)
	}
	return f
}

// at sets w to the weights of f's seasons at map point (x, y), interpolated
// between its samples, or computed outside them.
func (b *Blend) at(f *field, x, y float64, w []float64) {
	fx, fy := (x-f.x0)/step, (y-f.y0)/step
	i, j := int(math.Floor(fx)), int(math.Floor(fy))
	if i < 0 || j < 0 || i+1 >= f.nx || j+1 >= f.ny {
		b.weights(x, y, f.seasons, w)
		return
	}
	fx, fy = fx-float64(i), fy-float64(j)
	n := len(f.seasons)
	p := (j*f.nx + i) * n
	q := p + f.nx*n
	for s := range w {
		top := float64(f.w[p+s]) + (float64(f.w[p+n+s])-float64(f.w[p+s]))*fx
		bottom := float64(f.w[q+s]) + (float64(f.w[q+n+s])-float64(f.w[q+s]))*fx
		w[s] = top + (bottom-top)*fy
	}
}

// each visits every hex of g.
func each(g hexgrid.Grid, visit func(hexgrid.Coord)) {
	for l := range g.Lines {
		width := g.Width
		if l&1 == 1 {
			width--
		}
		for col := range width {
			visit(hexgrid.Coord{Line: l, Col: col})
		}
	}
}

// Seasons are the seasons mixed on hex c, its own first; nil where c is drawn
// in its own season only.
func (b *Blend) Seasons(c hexgrid.Coord) []mapcompose.Season {
	if b == nil || b.border[c] == nil {
		return nil
	}
	return b.border[c].seasons
}

// around visits the hexes of the grid whose centre lies within r of (x, y),
// with their squared distance.
func (b *Blend) around(x, y, r float64, visit func(hexgrid.Coord, float64)) {
	for l := max(0, int(math.Floor((y-40-r)/40))); l <= min(b.grid.Lines-1, int(math.Ceil((y-40+r)/40))); l++ {
		shift := 48.0 + float64(l&1)*71
		for col := max(0, int(math.Floor((x-shift-r)/142))); col <= int(math.Ceil((x-shift+r)/142)); col++ {
			c := hexgrid.Coord{Line: l, Col: col}
			if !b.grid.Contains(c) {
				continue
			}
			hx, hy := hexgrid.Center(c)
			if d2 := (hx-x)*(hx-x) + (hy-y)*(hy-y); d2 < r*r {
				visit(c, d2)
			}
		}
	}
}

// weights sets w to the share of each of seasons at map point (x, y), summing
// to 1; a season not listed counts for nothing, and a point no listed season
// reaches is wholly seasons[0].
func (b *Blend) weights(x, y float64, seasons []mapcompose.Season, w []float64) {
	clear(w)
	x += warp1*noise(x/wave1, y/wave1, 1) + warp2*noise(x/wave2, y/wave2, 3)
	y += warp1*noise(x/wave1, y/wave1, 2) + warp2*noise(x/wave2, y/wave2, 4)
	b.around(x, y, radius, func(h hexgrid.Coord, d2 float64) {
		if i := slices.Index(seasons, b.season(h)); i >= 0 {
			k := 1 - d2/(radius*radius)
			w[i] += k * k * k
		}
	})
	sum := 0.0
	for i, v := range w {
		if v == 0 {
			continue
		}
		// Fine noise per season breaks the band into patches.
		w[i] = v * math.Sqrt(v) * math.Exp(patches*noise(x/patch, y/patch, 10+uint32(slices.Index(order, seasons[i]))))
		sum += w[i]
	}
	if sum == 0 {
		w[0], sum = 1, 1
	}
	for i := range w {
		w[i] /= sum
	}
}

// Part is one image of a layer, its top-left corner at At in map pixels.
type Part struct {
	Image image.Image
	At    image.Point
}

// Mix draws layers, the same layer of hex c in each of Seasons(c) (each a
// stack of parts drawn in order), mixed by the seasons' weights at each pixel,
// at density. The result covers every part, in map pixels; nil without parts.
func (b *Blend) Mix(c hexgrid.Coord, layers [][]Part, density int) *image.RGBA {
	seasons := b.Seasons(c)
	var bounds image.Rectangle
	for _, layer := range layers {
		for _, p := range layer {
			bounds = bounds.Union(p.Image.Bounds().Sub(p.Image.Bounds().Min).Add(p.At))
		}
	}
	if bounds.Empty() || len(layers) != len(seasons) {
		return nil
	}
	drawn := make([]*image.RGBA, len(layers))
	for i, layer := range layers {
		drawn[i] = image.NewRGBA(bounds)
		for _, p := range layer {
			draw.Draw(drawn[i], p.Image.Bounds().Sub(p.Image.Bounds().Min).Add(p.At), p.Image, p.Image.Bounds().Min, draw.Over)
		}
	}
	out := image.NewRGBA(bounds)
	f := b.border[c]
	w := make([]float64, len(seasons))
	q := make([]int, len(seasons))
	for ly := floorDiv(bounds.Min.Y, density); ly*density < bounds.Max.Y; ly++ {
		for lx := floorDiv(bounds.Min.X, density); lx*density < bounds.Max.X; lx++ {
			b.at(f, float64(lx)+0.5, float64(ly)+0.5, w)
			fixed(w, q)
			for py := max(ly*density, bounds.Min.Y); py < min((ly+1)*density, bounds.Max.Y); py++ {
				i := out.PixOffset(max(lx*density, bounds.Min.X), py)
				j := out.PixOffset(min((lx+1)*density, bounds.Max.X), py)
				if s := slices.Index(q, 256); s >= 0 {
					copy(out.Pix[i:j], drawn[s].Pix[i:j])
					continue
				}
				for k := i; k < j; k++ {
					v := 128
					for s, im := range drawn {
						v += q[s] * int(im.Pix[k])
					}
					out.Pix[k] = uint8(v >> 8)
				}
			}
		}
	}
	return out
}

// fixed sets q to weights w in 1/256, summing to 256.
func fixed(w []float64, q []int) {
	sum, most := 0, 0
	for s, v := range w {
		q[s] = int(math.Round(v * 256))
		sum += q[s]
		if q[s] > q[most] {
			most = s
		}
	}
	q[most] += 256 - sum
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

// Scattered are the scattered items of hex c, items giving them for a scene:
// where every season scatters as many items there, each item comes from the
// season its point falls to, by the weights there and a fixed draw per point;
// otherwise the items of c's own season. ok is false where c is drawn in its
// own season only.
func (b *Blend) Scattered(c hexgrid.Coord, items func(*mapcompose.Scene) []mapcompose.Placed) (out []mapcompose.Placed, ok bool) {
	seasons := b.Seasons(c)
	if len(seasons) < 2 {
		return nil, false
	}
	per := make([][]mapcompose.Placed, len(seasons))
	for i, s := range seasons {
		per[i] = items(b.Scenes[s])
		if len(per[i]) != len(per[0]) {
			return per[0], true
		}
	}
	cx, cy := hexgrid.Center(c)
	w := make([]float64, len(seasons))
	out = make([]mapcompose.Placed, len(per[0]))
	for k, p := range per[0] {
		x, y := cx+p.X, cy+p.Y
		b.weights(x, y, seasons, w)
		u, i := unit(x, y), 0
		for ; i < len(w)-1 && u >= w[i]; i++ {
			u -= w[i]
		}
		out[k] = per[i][k]
	}
	return out, true
}

// noise is smooth value noise in [-1, 1] at (x, y), one field per seed.
func noise(x, y float64, seed uint32) float64 {
	ix, iy := math.Floor(x), math.Floor(y)
	fx, fy := x-ix, y-iy
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	at := func(dx, dy float64) float64 { return 2*hash(int64(ix+dx), int64(iy+dy), seed) - 1 }
	top := at(0, 0) + (at(1, 0)-at(0, 0))*fx
	bottom := at(0, 1) + (at(1, 1)-at(0, 1))*fx
	return top + (bottom-top)*fy
}

// unit is a fixed draw in [0, 1) for map point (x, y).
func unit(x, y float64) float64 {
	return hash(int64(math.Round(x*16)), int64(math.Round(y*16)), 5)
}

// hash is a fixed value in [0, 1) for lattice point (x, y).
func hash(x, y int64, seed uint32) float64 {
	h := uint64(x)*0x9e3779b97f4a7c15 ^ uint64(y)*0xc2b2ae3d27d4eb4f ^ uint64(seed)*0x165667b19e3779f9
	h ^= h >> 31
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 29
	return float64(h>>11) / (1 << 53)
}
