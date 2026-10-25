package dosscreen

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The slim frame (Frame.Slim) is the frame of a painted map: each panel shows
// its map edge to edge in a plain rectangle, the opening, inside a thin border
// of dark gunmetal with a brass hairline along the edge facing the map. Under
// each panel a gunmetal strip holds the status line; the divider is a narrow
// post with a brass knob at its foot between the strips. Where a panel looks
// past the map's ends, DrawBeyond fills it with the same gunmetal, edged with
// the hairline.

// slimRimNative is the slim frame along the window's edges in reference
// pixels, its hairline included.
const slimRimNative = 3

// The slim frame's palette: near-black blue-grey gunmetal, brass in a base and
// a lit tone, the dark line between the map and the brass and the black line
// between the brass and the gunmetal.
var (
	gunmetalDark  = color.RGBA{0x13, 0x14, 0x17, 0xFF}
	gunmetalLight = color.RGBA{0x2E, 0x2F, 0x33, 0xFF}
	brassBase     = color.RGBA{0x8C, 0x74, 0x48, 0xFF}
	brassLight    = color.RGBA{0xC2, 0xA4, 0x66, 0xFF}
	slimShadow    = color.RGBA{0x0B, 0x0A, 0x08, 0xFF}
	slimEdge      = color.RGBA{0x05, 0x05, 0x06, 0xFF}
)

// Shading of the gunmetal, as offsets along its dark-to-light ramp.
const (
	stripLift = 0.12 // the status strips stand a little lighter than the frame
	postLift  = 0.30 // the divider post at its crest
)

// SlimOpening is where panel side shows its map in the slim frame: the panel
// without the frame's rim along the window's top and outer edge. Along the
// divider and below, the frame lies outside the panel.
func (s Screen) SlimOpening(side int) Rect {
	p, rim := s.Panels[side&1], scaled(slimRimNative, s.UI)
	left, right := 0, rim
	if side&1 == 0 {
		left, right = rim, 0
	}
	return Rect{p.X + left, p.Y + rim, max(p.W-left-right, 0), max(p.H-rim, 0)}
}

// slimStrip is the gunmetal strip under panel side in the slim frame, holding
// its status well: as wide as the opening, from below the opening's hairline
// down to the well's bottom.
func (s Screen) slimStrip(side int) Rect {
	o, st := s.SlimOpening(side), s.Status[side&1]
	y := o.Bottom() + ringWidth(s.UI)
	return Rect{o.X, y, o.W, max(st.Bottom()-y, 0)}
}

// WorldRect is a rectangle in world pixels, from X0, Y0 to X1, Y1.
type WorldRect struct{ X0, Y0, X1, Y1 float64 }

// SlimFit returns v shifted by whole device pixels so that panel side's slim
// opening lies on world on each axis: a view looking past one end moves back
// onto the map. On an axis where world is smaller than the opening it is
// centred instead.
func (s Screen) SlimFit(v View, side int, world WorldRect) View {
	v.ShiftX, v.ShiftY = 0, 0
	o := s.SlimOpening(side)
	x0, y0 := v.WorldToScreen(world.X0, world.Y0)
	x1, y1 := v.WorldToScreen(world.X1, world.Y1)
	v.ShiftX = -fitAxis(o.X, o.Right(), x0, x1) / v.Scale
	v.ShiftY = -fitAxis(o.Y, o.Bottom(), y0, y1) / v.Scale
	return v
}

// fitAxis is how many whole device pixels to move a map spanning w0..w1 on the
// screen so that it covers the opening o0..o1, or is centred on a wider one.
func fitAxis(o0, o1 int, w0, w1 float64) float64 {
	switch lo, hi := float64(o0), float64(o1); {
	case w1-w0 < hi-lo:
		return math.Round((lo + hi - w0 - w1) / 2)
	case w0 > lo:
		return -math.Ceil(w0 - lo - eps)
	case w1 < hi:
		return math.Ceil(hi - w1 - eps)
	}
	return 0
}

// FillZoom is the lowest usable zoom level, at most DefaultZoom, at which a
// world of ww × wh world pixels still covers the slim openings, so the slim
// frame never shows gunmetal past the map when zoomed out. A world too small
// at DefaultZoom gets DefaultZoom.
func (s Screen) FillZoom(ww, wh float64) int {
	lo, _ := s.zoomRange()
	o := s.SlimOpening(0)
	for i := lo; i < DefaultZoom; i++ {
		if sc := s.zoomScale(i); ww*sc >= float64(o.W)-eps && wh*sc >= float64(o.H)-eps {
			return i
		}
	}
	return DefaultZoom
}

// hair is the width of one line of the slim frame at UI scale ui.
func hair(ui float64) int { return max(int(math.Round(ui)), 1) }

// ringWidth is the width of the lines around an opening at UI scale ui: the
// shadow, the brass (two lines: lit and base) and the black edge.
func ringWidth(ui float64) int { return 4 * hair(ui) }

// pieceKind is what a piece of the slim frame is painted with.
type pieceKind int

const (
	pieceMetal pieceKind = iota
	pieceShadow
	pieceBrass
	pieceBrassLit
	pieceEdge
)

// colour is the solid colour of a piece (the gunmetal's mean for pieceMetal).
func (k pieceKind) colour() color.RGBA {
	switch k {
	case pieceShadow:
		return slimShadow
	case pieceBrass:
		return brassBase
	case pieceBrassLit:
		return brassLight
	case pieceEdge:
		return slimEdge
	}
	return gunmetal(steelBase)
}

// slimPiece is a rectangle of the slim frame and its paint.
type slimPiece struct {
	r    Rect
	kind pieceKind
}

// intersect is the part of r inside o (empty: zero width or height).
func (r Rect) intersect(o Rect) Rect {
	x0, y0 := max(r.X, o.X), max(r.Y, o.Y)
	x1, y1 := min(r.Right(), o.Right()), min(r.Bottom(), o.Bottom())
	return Rect{x0, y0, max(x1-x0, 0), max(y1-y0, 0)}
}

func (r Rect) empty() bool { return r.W <= 0 || r.H <= 0 }

// ring is the lines around r, outside it, at line width l: the shadow next to
// r, the brass lit from the top left (its upper or left line lit) and the black
// edge. The horizontal pieces span the corners.
func ring(r Rect, l int) []slimPiece {
	band := func(d0, d1 int, near, far pieceKind) []slimPiece {
		w := d1 - d0
		return []slimPiece{
			{Rect{r.X - d1, r.Y - d1, r.W + 2*d1, w}, far},         // top: the far side is up
			{Rect{r.X - d1, r.Bottom() + d0, r.W + 2*d1, w}, near}, // bottom: the near side is up
			{Rect{r.X - d1, r.Y - d0, w, r.H + 2*d0}, far},         // left: the far side is left
			{Rect{r.Right() + d0, r.Y - d0, w, r.H + 2*d0}, near},  // right: the near side is left
		}
	}
	var out []slimPiece
	out = append(out, band(0, l, pieceShadow, pieceShadow)...)
	out = append(out, band(l, 2*l, pieceBrassLit, pieceBrass)...)
	out = append(out, band(2*l, 3*l, pieceBrass, pieceBrassLit)...)
	out = append(out, band(3*l, 4*l, pieceEdge, pieceEdge)...)
	return out
}

// clipped is the pieces cut to clip, without the empty ones.
func clipped(ps []slimPiece, clip Rect) []slimPiece {
	out := ps[:0:0]
	for _, p := range ps {
		if p.r = p.r.intersect(clip); !p.r.empty() {
			out = append(out, p)
		}
	}
	return out
}

// beyondPieces are what covers opening open where it looks past world (both in
// screen pixels): gunmetal outside world, and the ring along world's edge
// inside open. Nothing when world covers open.
func beyondPieces(open, world Rect, ui float64) []slimPiece {
	in := world.intersect(open)
	if in == open {
		return nil
	}
	if in.empty() {
		return []slimPiece{{open, pieceMetal}}
	}
	out := []slimPiece{
		{Rect{open.X, open.Y, open.W, in.Y - open.Y}, pieceMetal},
		{Rect{open.X, in.Bottom(), open.W, open.Bottom() - in.Bottom()}, pieceMetal},
		{Rect{open.X, in.Y, in.X - open.X, in.H}, pieceMetal},
		{Rect{in.Right(), in.Y, open.Right() - in.Right(), in.H}, pieceMetal},
	}
	return append(clipped(out, open), clipped(ring(world, hair(ui)), open)...)
}

// gunmetal is position t (clamped to 0..1) between the dark and light gunmetal.
func gunmetal(t float64) color.RGBA {
	return mixRGBA(gunmetalDark, gunmetalLight, t)
}

// gunmetalAt is the slim frame's gunmetal at screen pixel (x, y) lifted by dt:
// the steel texture's faint mottling and brushing on the gunmetal ramp.
func gunmetalAt(x, y int, dt float64) color.RGBA { return gunmetal(steelAt(x, y) + dt) }

// PaintSlimChrome paints the slim frame of s: gunmetal everywhere but the
// panels' openings, which stay transparent, the status strips, the ring of
// lines around each opening and strip, the divider post and its brass knob.
func PaintSlimChrome(s Screen) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	all := Rect{0, 0, s.W, s.H}
	l, rw := hair(s.UI), ringWidth(s.UI)
	o0, o1 := s.SlimOpening(0), s.SlimOpening(1)
	px0, px1 := o0.Right()+rw, o1.X-rw // the post between the rings
	for y := range s.H {
		for x := range s.W {
			dt := 0.0
			if x >= px0 && x < px1 {
				// The post rounds towards the light: a crest left of its middle.
				u := (float64(x-px0) + 0.5) / float64(px1-px0)
				dt = postLift * math.Pow(math.Sin(math.Pi*math.Min(1, u*1.25)), 2)
			}
			img.SetRGBA(x, y, gunmetalAt(x, y, dt))
		}
	}
	fill := func(ps []slimPiece) {
		for _, p := range clipped(ps, all) {
			c := p.kind.colour()
			for y := p.r.Y; y < p.r.Bottom(); y++ {
				for x := p.r.X; x < p.r.Right(); x++ {
					img.SetRGBA(x, y, c)
				}
			}
		}
	}
	for side := range 2 {
		st := s.slimStrip(side)
		for y := st.Y; y < st.Bottom(); y++ {
			for x := st.X; x < st.Right(); x++ {
				img.SetRGBA(x, y, gunmetalAt(x, y, stripLift))
			}
		}
		fill(ring(st, l))
	}
	for side := range 2 {
		o := s.SlimOpening(side)
		fill(ring(o, l)) // over the strip's top ring: one hairline between map and strip
		for y := o.Y; y < o.Bottom(); y++ {
			for x := o.X; x < o.Right(); x++ {
				img.SetRGBA(x, y, color.RGBA{})
			}
		}
	}
	b := s.Bolt
	paintKnob(img, Point{float64(b.X) + float64(b.W)/2, float64(b.Y) + float64(b.H)/2}, float64(b.W)/2-0.5)
	return img
}

// paintKnob draws a round brass knob of radius rad at c over the opaque frame:
// a dome lit from the top left with a soft shadow to the bottom right.
func paintKnob(img *image.RGBA, c Point, rad float64) {
	lz := 1.5
	l := math.Sqrt(lightX*lightX + lightY*lightY + lz*lz)
	lx, ly, lz := lightX/l, lightY/l, lz/l
	dark := color.RGBA{0x2E, 0x22, 0x12, 0xFF}
	shine := color.RGBA{0xF4, 0xE2, 0xB4, 0xFF}
	x0, x1 := int(math.Floor(c.X-rad-2)), int(math.Ceil(c.X+rad*1.4+2))
	y0, y1 := int(math.Floor(c.Y-rad-2)), int(math.Ceil(c.Y+rad*1.4+2))
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !(image.Point{x, y}).In(img.Rect) {
				continue
			}
			dx, dy := float64(x)+0.5-c.X, float64(y)+0.5-c.Y
			d := math.Hypot(dx, dy)
			if sd := math.Hypot(dx-rad*0.3, dy-rad*0.3); sd < rad+0.5 && d >= rad {
				blend(img, x, y, slimEdge, 0.6*math.Min(1, rad+0.5-sd))
			}
			cover := math.Min(1, rad+0.5-d)
			if cover <= 0 {
				continue
			}
			nx, ny := dx/rad, dy/rad
			nz := math.Sqrt(math.Max(0, 1-nx*nx-ny*ny))
			lit := math.Max(0, nx*lx+ny*ly+nz*lz)
			k := mixRGBA(dark, brassBase, math.Min(1, lit*1.25))
			if lit > 0.9 {
				k = mixRGBA(brassLight, shine, (lit-0.9)/0.1)
			} else if lit > 0.7 {
				k = mixRGBA(brassBase, brassLight, (lit-0.7)/0.2)
			}
			blend(img, x, y, k, cover)
		}
	}
}

// DrawBeyond covers opening open of dst where it looks past world (both in
// dst's pixels) with the slim frame's gunmetal, edged with the brass hairline
// along world's edge: the map ends in the frame. Draw it over the map, before
// any view laid on the panel.
func (f *Frame) DrawBeyond(dst *ebiten.Image, open, world Rect, ui float64) {
	ps := beyondPieces(open, world, ui)
	if len(ps) == 0 {
		return
	}
	if f.metal == nil {
		tile := image.NewRGBA(image.Rect(0, 0, steelTile, steelTile))
		for y := range steelTile {
			for x := range steelTile {
				tile.SetRGBA(x, y, gunmetalAt(x, y, 0))
			}
		}
		f.metal = ebiten.NewImageFromImage(tile)
	}
	var vs []ebiten.Vertex
	var is []uint16
	for _, p := range ps {
		if p.kind != pieceMetal {
			continue
		}
		// The texture follows dst's pixels, as the chrome's follows the screen's.
		n := uint16(len(vs))
		for _, q := range [4][2]int{{p.r.X, p.r.Y}, {p.r.Right(), p.r.Y}, {p.r.X, p.r.Bottom()}, {p.r.Right(), p.r.Bottom()}} {
			vs = append(vs, ebiten.Vertex{
				DstX: float32(q[0]), DstY: float32(q[1]), SrcX: float32(q[0]), SrcY: float32(q[1]),
				ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
			})
		}
		is = append(is, n, n+1, n+2, n+1, n+3, n+2)
	}
	if len(vs) > 0 {
		dst.DrawTriangles(vs, is, f.metal, &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat})
	}
	for _, p := range ps { // the lines over the gunmetal
		if p.kind != pieceMetal {
			vector.FillRect(dst, float32(p.r.X), float32(p.r.Y), float32(p.r.W), float32(p.r.H), p.kind.colour(), false)
		}
	}
}
