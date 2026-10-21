package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

const (
	compareDensity = 2  // crops show the render at its runtime density
	bandLines      = 12 // hex lines per full-map band
	cropMargin     = 24 // world pixels around a feature
	pairGap        = 8  // pixels between reference and render
	// edgeThreshold is the L* gradient per 1× pixel (Sobel, normalised) above which a pixel counts as an edge.
	edgeThreshold = 6
)

// metrics are colour statistics of a set of hexes, measured at the game's 1× scale,
// which is close to the reference painting's own resolution.
type metrics struct {
	L        float64 `json:"l"`        // mean CIELAB L*
	A        float64 `json:"a"`        // mean a*
	B        float64 `json:"b"`        // mean b*
	Contrast float64 `json:"contrast"` // standard deviation of L*
	Chroma   float64 `json:"chroma"`   // mean C*ab
	Edges    float64 `json:"edges"`    // share of pixels with an L* gradient above edgeThreshold
}

// deltaE is the CIE76 distance between the mean colours.
func (m metrics) deltaE(o metrics) float64 {
	return math.Sqrt((m.L-o.L)*(m.L-o.L) + (m.A-o.A)*(m.A-o.A) + (m.B-o.B)*(m.B-o.B))
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func (m metrics) rounded() metrics {
	return metrics{round2(m.L), round2(m.A), round2(m.B), round2(m.Contrast), round2(m.Chroma), round2(m.Edges)}
}

// labPlane is an image converted to CIELAB once; ok marks opaque pixels.
type labPlane struct {
	w, h    int
	l, a, b []float64
	ok      []bool
}

func srgbLinear(v uint8) float64 {
	c := float64(v) / 255
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func labF(t float64) float64 {
	if t > 216.0/24389 {
		return math.Cbrt(t)
	}
	return (24389.0/27*t + 16) / 116
}

// toLab converts sRGB to CIELAB (D65).
func toLab(c color.NRGBA) (float64, float64, float64) {
	r, g, b := srgbLinear(c.R), srgbLinear(c.G), srgbLinear(c.B)
	x := (0.4124564*r + 0.3575761*g + 0.1804375*b) / 0.95047
	y := 0.2126729*r + 0.7151522*g + 0.0721750*b
	z := (0.0193339*r + 0.1191920*g + 0.9503041*b) / 1.08883
	fx, fy, fz := labF(x), labF(y), labF(z)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

func newLabPlane(im *image.NRGBA) *labPlane {
	b := im.Bounds()
	p := &labPlane{w: b.Dx(), h: b.Dy()}
	n := p.w * p.h
	p.l, p.a, p.b, p.ok = make([]float64, n), make([]float64, n), make([]float64, n), make([]bool, n)
	for y := range p.h {
		for x := range p.w {
			c := im.NRGBAAt(b.Min.X+x, b.Min.Y+y)
			i := y*p.w + x
			if c.A == 255 {
				p.l[i], p.a[i], p.b[i] = toLab(c)
				p.ok[i] = true
			}
		}
	}
	return p
}

// gradient is the normalised Sobel magnitude of L* at (x, y); ok is false when a
// neighbour is missing.
func (p *labPlane) gradient(x, y int) (float64, bool) {
	if x < 1 || y < 1 || x >= p.w-1 || y >= p.h-1 {
		return 0, false
	}
	var v [3][3]float64
	for j := range 3 {
		for i := range 3 {
			k := (y+j-1)*p.w + x + i - 1
			if !p.ok[k] {
				return 0, false
			}
			v[j][i] = p.l[k]
		}
	}
	gx := (v[0][2] + 2*v[1][2] + v[2][2] - v[0][0] - 2*v[1][0] - v[2][0]) / 8
	gy := (v[2][0] + 2*v[2][1] + v[2][2] - v[0][0] - 2*v[0][1] - v[0][2]) / 8
	return math.Hypot(gx, gy), true
}

// measure collects the metrics over the given hexes of a 1× plane.
func (p *labPlane) measure(cells []cell) metrics {
	var n, edges, edgeN int
	var sl, sa, sb, sll, sc float64
	for _, c := range cells {
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		for y := range hexgrid.HexHeight {
			for x := range hexgrid.HexWidth {
				px, py := ox+x, oy+y
				if !insideTile(x, y, 1) || px < 0 || py < 0 || px >= p.w || py >= p.h {
					continue
				}
				i := py*p.w + px
				if !p.ok[i] {
					continue
				}
				n++
				sl += p.l[i]
				sa += p.a[i]
				sb += p.b[i]
				sll += p.l[i] * p.l[i]
				sc += math.Hypot(p.a[i], p.b[i])
				if g, ok := p.gradient(px, py); ok {
					edgeN++
					if g > edgeThreshold {
						edges++
					}
				}
			}
		}
	}
	if n == 0 {
		return metrics{}
	}
	fn := float64(n)
	m := metrics{L: sl / fn, A: sa / fn, B: sb / fn, Chroma: sc / fn}
	m.Contrast = math.Sqrt(math.Max(0, sll/fn-m.L*m.L))
	if edgeN > 0 {
		m.Edges = float64(edges) / float64(edgeN)
	}
	return m.rounded()
}

// registered resamples the reference onto the game map at the given density.
func (p *project) registered(src image.Image, density int) *image.NRGBA {
	w, h := maprender.Size(p.level.Grid())
	im := image.NewNRGBA(image.Rect(0, 0, w*density, h*density))
	d := float64(density)
	for y := range h * density {
		for x := range w * density {
			im.SetNRGBA(x, y, sampleSource(src, p.ref.Registration, (float64(x)+.5)/d, (float64(y)+.5)/d))
		}
	}
	return im
}

// renderMap draws the level with the runtime art at compareDensity.
func (p *project) renderMap() (*image.NRGBA, error) {
	r, err := maprender.New(p.assets)
	if err != nil {
		return nil, err
	}
	m, err := r.Render(p.level, compareDensity, false)
	if err != nil {
		return nil, err
	}
	return m.Image, nil
}

// halve averages 2×2 blocks; a block with a transparent pixel stays transparent.
func halve(im *image.NRGBA) *image.NRGBA {
	b := im.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx()/2, b.Dy()/2))
	for y := range out.Bounds().Dy() {
		for x := range out.Bounds().Dx() {
			var r, g, bl int
			opaque := true
			for j := range 2 {
				for i := range 2 {
					c := im.NRGBAAt(b.Min.X+2*x+i, b.Min.Y+2*y+j)
					opaque = opaque && c.A == 255
					r, g, bl = r+int(c.R), g+int(c.G), bl+int(c.B)
				}
			}
			if opaque {
				out.SetNRGBA(x, y, color.NRGBA{uint8((r + 2) / 4), uint8((g + 2) / 4), uint8((bl + 2) / 4), 255})
			}
		}
	}
	return out
}

// renderMetrics measures every family's exemplars on the 2× render.
func (p *project) renderMetrics(render *image.NRGBA) map[string]metrics {
	plane := newLabPlane(halve(render))
	out := map[string]metrics{}
	for _, f := range p.ref.Families {
		out[f.ID] = plane.measure(p.exemplars(f))
	}
	return out
}

func (p *project) exemplars(f family) []cell {
	var out []cell
	for _, s := range f.Exemplars {
		c, _ := parseCell(s)
		out = append(out, c)
	}
	return out
}

// comparison is compare.json: what the review page and the art book show beside
// the images.
type comparison struct {
	Level    string          `json:"level"`
	Density  int             `json:"density"`
	Bands    []string        `json:"bands"`
	Families []familyCompare `json:"families"`
	Features []featureCrop   `json:"features"`
}

type familyCompare struct {
	ID        string   `json:"id"`
	Crops     []string `json:"crops"` // one per exemplar, in exemplar order
	Exemplars []string `json:"exemplars"`
	Reference metrics  `json:"reference"`
	Render    metrics  `json:"render"`
	DeltaE    float64  `json:"delta_e"`
	Outside   []string `json:"outside"` // metrics outside the tolerance
}

type featureCrop struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Crop  string `json:"crop"`
}

// compare writes full-map bands, one crop per family exemplar and per feature
// (reference | render) and compare.json into out, which it replaces.
func (p *project) compare(src image.Image, out string, format imageFormat) (*comparison, error) {
	render, err := p.renderMap()
	if err != nil {
		return nil, err
	}
	ref := masked(p.registered(src, compareDensity), render)
	if err := os.RemoveAll(out); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, err
	}
	cmp := &comparison{Level: p.ref.Level, Density: compareDensity}

	// Bands at 1×, the scale of the painting.
	render1 := halve(render)
	ref1 := masked(p.registered(src, 1), render1)
	lines := len(p.level.Lines)
	for first := 0; first < lines; first += bandLines {
		last := min(first+bandLines, lines) - 1
		r := image.Rect(0, first*hexgrid.HexHeight/2, ref1.Bounds().Dx(), last*hexgrid.HexHeight/2+hexgrid.HexHeight)
		name := fmt.Sprintf("band-%d.%s", len(cmp.Bands)+1, format)
		if err := format.write(filepath.Join(out, name), pair(ref1, render1, r, 1, nil)); err != nil {
			return nil, err
		}
		cmp.Bands = append(cmp.Bands, name)
	}

	refPlane := newLabPlane(ref1)
	renderMetrics := p.renderMetrics(render)
	for _, f := range p.ref.Families {
		fc := familyCompare{ID: f.ID, Exemplars: f.Exemplars, Reference: refPlane.measure(p.exemplars(f)), Render: renderMetrics[f.ID]}
		fc.DeltaE = round2(fc.Render.deltaE(fc.Reference))
		fc.Outside = p.ref.Tolerance.outside(fc.Reference, fc.Render)
		for _, c := range p.exemplars(f) {
			ox, oy := tileOrigin(c.Row-1, c.Col-1)
			r := image.Rect(ox-hexgrid.HexWidth/2, oy-hexgrid.HexHeight/2, ox+hexgrid.HexWidth*3/2, oy+hexgrid.HexHeight*3/2)
			name := fmt.Sprintf("%s-%s.%s", f.ID, c, format)
			if err := format.write(filepath.Join(out, name), pair(ref, render, scaleRect(r, compareDensity), compareDensity, []cell{c})); err != nil {
				return nil, err
			}
			fc.Crops = append(fc.Crops, name)
		}
		cmp.Families = append(cmp.Families, fc)
	}

	for _, ft := range p.ref.Features {
		var cells []cell
		r := image.Rectangle{}
		for _, s := range ft.Cells {
			c, _ := parseCell(s)
			cells = append(cells, c)
			ox, oy := tileOrigin(c.Row-1, c.Col-1)
			r = r.Union(image.Rect(ox, oy, ox+hexgrid.HexWidth, oy+hexgrid.HexHeight))
		}
		name := "feature-" + ft.ID + "." + string(format)
		if err := format.write(filepath.Join(out, name), pair(ref, render, scaleRect(r.Inset(-cropMargin), compareDensity), compareDensity, cells)); err != nil {
			return nil, err
		}
		cmp.Features = append(cmp.Features, featureCrop{ID: ft.ID, Title: ft.Title, Crop: name})
	}

	raw, err := json.MarshalIndent(cmp, "", "  ")
	if err != nil {
		return nil, err
	}
	return cmp, os.WriteFile(filepath.Join(out, "compare.json"), append(raw, '\n'), 0o644)
}

func scaleRect(r image.Rectangle, k int) image.Rectangle {
	return image.Rect(r.Min.X*k, r.Min.Y*k, r.Max.X*k, r.Max.Y*k)
}

var outline = color.NRGBA{255, 212, 0, 255}

// pair places the area r of the reference and of the render side by side, with
// the given hexes outlined in both; r is in pixels at density. Parts outside the
// map stay transparent.
func pair(ref, render *image.NRGBA, r image.Rectangle, density int, marked []cell) *image.NRGBA {
	w, h := r.Dx(), r.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, 2*w+pairGap, h))
	draw.Draw(out, image.Rect(0, 0, w, h), ref, r.Min, draw.Src)
	draw.Draw(out, image.Rect(w+pairGap, 0, 2*w+pairGap, h), render, r.Min, draw.Src)
	for _, c := range marked {
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		inside := func(x, y int) bool {
			return x >= 0 && y >= 0 && x < hexgrid.HexWidth*density && y < hexgrid.HexHeight*density && insideTile(x, y, density)
		}
		for y := range hexgrid.HexHeight * density {
			for x := range hexgrid.HexWidth * density {
				if !inside(x, y) || (inside(x-1, y) && inside(x+1, y) && inside(x, y-1) && inside(x, y+1)) {
					continue
				}
				px, py := ox*density+x-r.Min.X, oy*density+y-r.Min.Y
				if px >= 0 && py >= 0 && px < w && py < h {
					out.SetNRGBA(px, py, outline)
					out.SetNRGBA(px+w+pairGap, py, outline)
				}
			}
		}
	}
	return out
}

// imageFormat is the file type of the comparison images: PNG for the review
// page, JPEG (flattened onto white) to keep the art book small.
type imageFormat string

const (
	formatPNG  imageFormat = "png"
	formatJPEG imageFormat = "jpg"
)

func (f imageFormat) write(path string, im *image.NRGBA) error {
	if f == formatPNG {
		return writePNG(path, im)
	}
	flat := image.NewRGBA(im.Bounds())
	draw.Draw(flat, flat.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(flat, flat.Bounds(), im, im.Bounds().Min, draw.Over)
	var b bytes.Buffer
	if err := jpeg.Encode(&b, flat, &jpeg.Options{Quality: 88}); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// masked clears the reference wherever the render has no map, so the painting's
// border outside the grid stays out of the crops.
func masked(ref, render *image.NRGBA) *image.NRGBA {
	for i := 3; i < len(ref.Pix); i += 4 {
		if render.Pix[i] == 0 {
			ref.Pix[i-3], ref.Pix[i-2], ref.Pix[i-1], ref.Pix[i] = 0, 0, 0, 0
		}
	}
	return ref
}
