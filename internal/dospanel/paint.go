package dospanel

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/texture"
)

// The placeholder palette: olive-grey steel from dark to light (as the
// frame's), brown name plates with brass edges and rivets.
var (
	steelDark  = color.RGBA{0x24, 0x24, 0x1C, 0xFF}
	steelLight = color.RGBA{0xB2, 0xB2, 0xA6, 0xFF}
	plateBrown = color.RGBA{0x42, 0x34, 0x26, 0xFF}
	plateEdge  = color.RGBA{0x1C, 0x14, 0x0C, 0xFF}
	brassLight = color.RGBA{0xE0, 0xBC, 0x6C, 0xFF}
	brass      = color.RGBA{0xA8, 0x80, 0x3C, 0xFF}
	brassDark  = color.RGBA{0x58, 0x40, 0x1C, 0xFF}
	ringLight  = color.RGBA{0xD8, 0xD8, 0xC8, 0xFF}
)

// Steel tones as positions on the dark-to-light ramp.
const (
	toneBack  = 0.38 // the panel behind the plates
	tonePlate = 0.46 // a raised steel plate
	toneWell  = 0.04 // the floor of a recess
	toneBlank = 0.52 // a photo without a portrait
	toneHex   = 0.58 // the sprite's hex plate
	bevelTone = 0.30 // how far a bevel moves towards light or dark
)

// painter paints placeholder steel and brass on a CPU image; k is device
// pixels per reference pixel.
type painter struct {
	img *image.RGBA
	k   float64
}

// bw is the width of a bevel or edge line.
func (p painter) bw() int { return max(1, int(math.Round(p.k))) }

func lerp(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	m := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 0xFF}
}

// ramp is the steel colour at t from dark (0) to light (1).
func ramp(t float64) color.RGBA { return lerp(steelDark, steelLight, t) }

// grain is a fixed noise in [-0.5, 0.5] per cell.
func grain(x, y int) float64 {
	h := uint32(x)*0x9E3779B1 ^ uint32(y)*0x85EBCA77
	h ^= h >> 15
	h *= 0x2C1B3C6D
	h ^= h >> 12
	return float64(h&0xFFFF)/0xFFFF - 0.5
}

// steel fills r with steel of tone t.
func (p painter) steel(r image.Rectangle, t float64) {
	r = r.Intersect(p.img.Rect)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p.img.SetRGBA(x, y, p.steelColour(x, y, t))
		}
	}
}

// edges calls fn for the pixels within w of r's border: lit is true on the
// top and left sides, false on the bottom and right.
func edges(r image.Rectangle, w int, fn func(x, y int, lit bool)) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			dl := min(y-r.Min.Y, x-r.Min.X)
			dr := min(r.Max.Y-1-y, r.Max.X-1-x)
			switch {
			case dl < w && dl <= dr:
				fn(x, y, true)
			case dr < w:
				fn(x, y, false)
			}
		}
	}
}

// bevel shades r's border: light on the top and left when raised, dark when
// recessed, the other sides the other way.
func (p painter) bevel(r image.Rectangle, t float64, raised bool) {
	r = r.Intersect(p.img.Rect)
	edges(r, p.bw(), func(x, y int, lit bool) {
		d := bevelTone
		if lit != raised {
			d = -bevelTone
		}
		p.img.SetRGBA(x, y, p.steelColour(x, y, t+d))
	})
}

// plate is a raised steel plate.
func (p painter) plate(r image.Rectangle, t float64) {
	p.steel(r, t)
	p.bevel(r, t, true)
}

// well is a dark recess with a lip lit from the top left.
func (p painter) well(r image.Rectangle) {
	p.steel(r, toneWell)
	p.bevel(r, toneBack, false)
}

// namePlate is a brown plate with a dark edge, a brass line and a brass
// rivet in each corner, one at each end on a low plate.
func (p painter) namePlate(r image.Rectangle) {
	r = r.Intersect(p.img.Rect)
	w := p.bw()
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := lerp(plateBrown, plateEdge, 0.12+0.24*grain(x/w, y/w))
			if s := surfaces.steel; s != nil {
				dr, dg, db := s.At(x, y, p.k*steelSpan)
				c = texture.Modulate(lerp(plateBrown, plateEdge, 0.12), 1.5*steelGain, dr, dg, db)
			}
			p.img.SetRGBA(x, y, c)
		}
	}
	edges(r, w, func(x, y int, _ bool) { p.img.SetRGBA(x, y, plateEdge) })
	edges(r.Inset(w), w, func(x, y int, lit bool) {
		if lit {
			p.img.SetRGBA(x, y, brass)
		} else {
			p.img.SetRGBA(x, y, brassDark)
		}
	})
	in := 3.5 * p.k
	rad := math.Max(1.2, 1.3*p.k)
	x0, x1 := float64(r.Min.X)+in, float64(r.Max.X)-in
	ys := []float64{float64(r.Min.Y) + in, float64(r.Max.Y) - in}
	if ys[1]-ys[0] < 4*rad { // too low for two: one at mid height
		ys = []float64{float64(r.Min.Y+r.Max.Y) / 2}
	}
	for _, y := range ys {
		p.rivet(x0, y, rad)
		p.rivet(x1, y, rad)
	}
}

// rivet is a brass head of radius rad at (cx, cy), lit from the top left
// with a soft shadow to the bottom right: a painted head of the rivet sheet
// picked by the place, or a drawn dome without one.
func (p painter) rivet(cx, cy, rad float64) {
	sheet := surfaces.rivets
	var head rivetHead
	if sheet != nil {
		head = sheet.heads[min(len(sheet.heads)-1, int((grain(int(cx), int(cy))+0.5)*float64(len(sheet.heads))))]
	}
	for y := int(cy - rad - 2); y <= int(cy+rad*1.5+2); y++ {
		for x := int(cx - rad - 2); x <= int(cx+rad*1.5+2); x++ {
			if !(image.Point{x, y}).In(p.img.Rect) {
				continue
			}
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			d := math.Hypot(dx, dy)
			if sd := math.Hypot(dx-rad*0.4, dy-rad*0.4); sd < rad+0.5 && d >= rad {
				p.blend(x, y, ramp(0.1), 0.5*math.Min(1, rad+0.5-sd))
			}
			cov := math.Min(1, rad+0.5-d) // an antialiased rim
			if cov <= 0 {
				continue
			}
			if sheet != nil {
				p.blend(x, y, sheet.at(head, dx/rad, dy/rad, 1/rad, brass), cov)
				continue
			}
			nx, ny := dx/rad, dy/rad
			l := -0.5*nx - 0.5*ny + 0.707*math.Sqrt(math.Max(0, 1-nx*nx-ny*ny))
			c := lerp(brassDark, brass, l*2)
			if l > 0.5 {
				c = lerp(brass, brassLight, (l-0.5)*2)
			}
			p.blend(x, y, c, cov)
		}
	}
}

// hexNorm is how far (dx, dy) lies towards the edge of a flat-topped hex of
// half width a and half height b around the origin: at most 1 inside. The
// top and bottom edges are a third of the width long, half as wide as the
// slants, as in the map's cells.
func hexNorm(dx, dy, a, b float64) float64 {
	dx, dy = math.Abs(dx), math.Abs(dy)
	return math.Max(dy/b, (dx+2*a/3*dy/b)/a)
}

// hexEdge is the distance of (dx, dy) from the edge of the hex, negative
// outside it, and the light falling on that edge (-1 to 1, from the top left).
func hexEdge(dx, dy, a, b float64) (d, light float64) {
	ax, ay := math.Abs(dx), math.Abs(dy)
	sx, sy := math.Copysign(1, dx), math.Copysign(1, dy)
	top := b - ay
	s := math.Hypot(b, 2*a/3)
	slant := (a - ax - 2*a/3*ay/b) * b / s
	if top <= slant {
		return top, -sy * math.Sqrt2 / 2
	}
	nx, ny := sx*b/s, sy*(2*a/3)/s
	return slant, -(nx + ny) * math.Sqrt2 / 2
}

// hex fills the hex of half size a×b around (cx, cy) with steel of tone t,
// its edge lit from the top left when raised and the other way when
// recessed.
func (p painter) hex(cx, cy, a, b, t float64, raised bool) {
	bw := float64(p.bw())
	sign := 1.0
	if !raised {
		sign = -1
	}
	for y := int(cy - b - 1); y <= int(cy+b+1); y++ {
		for x := int(cx - a - 1); x <= int(cx+a+1); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			d, l := hexEdge(dx, dy, a, b)
			cov := math.Min(1, d+0.5) // an antialiased edge
			if cov <= 0 || !(image.Point{x, y}).In(p.img.Rect) {
				continue
			}
			bevel := math.Max(0, math.Min(1, bw-d+0.5))
			p.blend(x, y, p.steelColour(x, y, t+bevel*sign*bevelTone*l), cov)
		}
	}
}

// hexOf is the centre and half size of the hex inscribed in r.
func hexOf(r image.Rectangle) (cx, cy, a, b float64) {
	return float64(r.Min.X+r.Max.X) / 2, float64(r.Min.Y+r.Max.Y) / 2, float64(r.Dx()) / 2, float64(r.Dy()) / 2
}

// The plates are painted once per panel size and scale.
type plateKey struct {
	kind  int
	size  image.Point
	scale float64
}

const (
	unitInfoPlate = iota
	contentsPlate
	battlePlate
)

var plates = struct {
	sync.Mutex
	m map[plateKey]*ebiten.Image
}{m: map[plateKey]*ebiten.Image{}}

// cachedPlate is the plate of kind for a panel of size at scale, painted by
// paint on first use.
func cachedPlate(kind int, size image.Point, scale float64, paint func() *image.RGBA) *ebiten.Image {
	key := plateKey{kind, size, scale}
	plates.Lock()
	defer plates.Unlock()
	if img, ok := plates.m[key]; ok {
		return img
	}
	if len(plates.m) >= 8 { // window sizes come and go; keep a few
		for k, img := range plates.m {
			img.Deallocate()
			delete(plates.m, k)
		}
	}
	img := ebiten.NewImageFromImage(paint())
	plates.m[key] = img
	return img
}

// drawPlate draws a cached plate at the panel's top left.
func drawPlate(dst *ebiten.Image, panel image.Rectangle, img *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panel.Min.X), float64(panel.Min.Y))
	clip(dst, panel).DrawImage(img, op)
}

// sepia is the tone of the photos: a classic sepia matrix.
func sepia() colorm.ColorM {
	var m colorm.ColorM
	rows := [3][3]float64{{0.393, 0.769, 0.189}, {0.349, 0.686, 0.168}, {0.272, 0.534, 0.131}}
	for i, row := range rows {
		for j, v := range row {
			m.SetElement(i, j, v)
		}
	}
	return m
}

// drawPhoto draws img over r, scaled to cover it, in sepia; nothing for nil.
func drawPhoto(dst *ebiten.Image, r image.Rectangle, img *ebiten.Image) {
	if img == nil || r.Empty() {
		return
	}
	b := img.Bounds()
	if b.Empty() {
		return
	}
	s := math.Max(float64(r.Dx())/float64(b.Dx()), float64(r.Dy())/float64(b.Dy()))
	op := &colorm.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-float64(b.Min.X), -float64(b.Min.Y))
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(r.Min.X)+(float64(r.Dx())-float64(b.Dx())*s)/2,
		float64(r.Min.Y)+(float64(r.Dy())-float64(b.Dy())*s)/2)
	colorm.DrawImage(clip(dst, r), img, sepia(), op)
}

// drawSprite draws img centred in r, scaled to fit share of it.
func drawSprite(dst *ebiten.Image, r image.Rectangle, img *ebiten.Image, share float64) {
	if img == nil || r.Empty() {
		return
	}
	b := img.Bounds()
	if b.Empty() {
		return
	}
	s := share * math.Min(float64(r.Dx())/float64(b.Dx()), float64(r.Dy())/float64(b.Dy()))
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-float64(b.Min.X), -float64(b.Min.Y))
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(r.Min.X)+(float64(r.Dx())-float64(b.Dx())*s)/2,
		float64(r.Min.Y)+(float64(r.Dy())-float64(b.Dy())*s)/2)
	clip(dst, r).DrawImage(img, op)
}
