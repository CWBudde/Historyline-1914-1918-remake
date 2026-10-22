package dospanel

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/presentation"
	"github.com/hajimehoshi/ebiten/v2"
)

// Domain is the kind of battle a vignette shows.
type Domain int

// The battle domains; Domains counts them.
const (
	Air Domain = iota
	Naval
	Artillery
	Armour
	Infantry
	Domains
)

var domainNames = [...]string{"air", "naval", "artillery", "armour", "infantry"}

func (d Domain) String() string {
	if d < 0 || d >= Domains {
		return fmt.Sprintf("Domain(%d)", int(d))
	}
	return domainNames[d]
}

// The vignettes are painted at the size of a reference panel and scaled to
// cover the panel.
const (
	VignetteW, VignetteH = 152, 180
	vignetteFrames       = 6
	vignetteMS           = 110
)

// Vignette is the animation of domain d's battle: its frames and frame
// times, with the first frame alone for reduced motion. An unknown domain
// has no frames.
func Vignette(d Domain) presentation.Animation {
	if d < 0 || d >= Domains {
		return presentation.Animation{}
	}
	id := "battle-" + d.String()
	a := presentation.Animation{ID: id, Skippable: true, Notes: "procedural placeholder"}
	for f := range vignetteFrames {
		a.Frames = append(a.Frames, fmt.Sprintf("%s-%d", id, f))
		a.FrameMS = append(a.FrameMS, vignetteMS)
	}
	a.ReducedFrames, a.ReducedMS = a.Frames[:1], []int{4 * vignetteMS}
	return a
}

// VignetteFrame paints frame of domain d's vignette (the index wraps) at
// VignetteW×VignetteH in sepia.
func VignetteFrame(d Domain, frame int) *image.RGBA {
	c := canvas{image.NewRGBA(image.Rect(0, 0, VignetteW, VignetteH))}
	n := len(Vignette(d).Frames)
	if n == 0 {
		c.backdrop(VignetteH)
		return c.img
	}
	f := ((frame % n) + n) % n
	switch d {
	case Air:
		airScene(c, f)
	case Naval:
		navalScene(c, f)
	case Artillery:
		artilleryScene(c, f)
	case Armour:
		armourScene(c, f)
	case Infantry:
		infantryScene(c, f)
	}
	return c.img
}

var vignettes = struct {
	sync.Mutex
	m map[[2]int]*ebiten.Image
}{m: map[[2]int]*ebiten.Image{}}

// DrawVignette draws frame of domain d's vignette over panel, scaled to
// cover it.
func DrawVignette(dst *ebiten.Image, panel image.Rectangle, d Domain, frame int) {
	if n := len(Vignette(d).Frames); n > 0 {
		frame = ((frame % n) + n) % n
	}
	key := [2]int{int(d), frame}
	vignettes.Lock()
	img, ok := vignettes.m[key]
	if !ok {
		img = ebiten.NewImageFromImage(VignetteFrame(d, frame))
		vignettes.m[key] = img
	}
	vignettes.Unlock()

	s := math.Max(float64(panel.Dx())/VignetteW, float64(panel.Dy())/VignetteH)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(panel.Min.X)+(float64(panel.Dx())-VignetteW*s)/2,
		float64(panel.Min.Y)+(float64(panel.Dy())-VignetteH*s)/2)
	clip(dst, panel).DrawImage(img, op)
}

// The sepia ramp from shadow through mid tone to highlight; every colour on
// it has red ≥ green ≥ blue.
var (
	sepiaDark  = color.RGBA{0x26, 0x1A, 0x0E, 0xFF}
	sepiaMid   = color.RGBA{0x8C, 0x6C, 0x46, 0xFF}
	sepiaLight = color.RGBA{0xF2, 0xDE, 0xB6, 0xFF}
)

func tone(t float64) color.RGBA {
	if t < 0.5 {
		return lerp(sepiaDark, sepiaMid, t*2)
	}
	return lerp(sepiaMid, sepiaLight, (t-0.5)*2)
}

// canvas paints sepia shapes in reference pixels; a shape covers the
// pixels whose centres it holds.
type canvas struct{ img *image.RGBA }

func (c canvas) set(x, y int, t float64) {
	if (image.Point{x, y}).In(c.img.Rect) {
		c.img.SetRGBA(x, y, tone(t))
	}
}

// backdrop is a sky lightening towards the horizon and a grained ground
// below it.
func (c canvas) backdrop(horizon int) {
	for y := range VignetteH {
		for x := range VignetteW {
			if y < horizon {
				c.set(x, y, 0.6+0.32*float64(y)/float64(horizon))
				continue
			}
			d := float64(y-horizon) / float64(max(1, VignetteH-horizon))
			c.set(x, y, 0.44-0.22*d+0.07*grain(x/2, y/2))
		}
	}
}

// region fills the pixels of the box (x0, y0)-(x1, y1) for which in holds.
func (c canvas) region(x0, y0, x1, y1 float64, t float64, in func(x, y float64) bool) {
	for y := int(math.Floor(y0)); y <= int(math.Ceil(y1)); y++ {
		for x := int(math.Floor(x0)); x <= int(math.Ceil(x1)); x++ {
			if in(float64(x)+0.5, float64(y)+0.5) {
				c.set(x, y, t)
			}
		}
	}
}

func (c canvas) rect(x0, y0, x1, y1, t float64) {
	c.region(x0, y0, x1, y1, t, func(x, y float64) bool { return x >= x0 && x < x1 && y >= y0 && y < y1 })
}

func (c canvas) ellipse(cx, cy, rx, ry, t float64) {
	c.region(cx-rx, cy-ry, cx+rx, cy+ry, t, func(x, y float64) bool {
		dx, dy := (x-cx)/rx, (y-cy)/ry
		return dx*dx+dy*dy <= 1
	})
}

func (c canvas) circle(cx, cy, r, t float64) { c.ellipse(cx, cy, r, r, t) }

// poly fills a polygon by the even-odd rule.
func (c canvas) poly(t float64, pts ...[2]float64) {
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range pts {
		x0, y0 = math.Min(x0, p[0]), math.Min(y0, p[1])
		x1, y1 = math.Max(x1, p[0]), math.Max(y1, p[1])
	}
	c.region(x0, y0, x1, y1, t, func(x, y float64) bool {
		in := false
		for i, j := 0, len(pts)-1; i < len(pts); j, i = i, i+1 {
			a, b := pts[i], pts[j]
			if (a[1] > y) != (b[1] > y) && x < (b[0]-a[0])*(y-a[1])/(b[1]-a[1])+a[0] {
				in = !in
			}
		}
		return in
	})
}

// line is a stroke of width w from (x0, y0) to (x1, y1).
func (c canvas) line(x0, y0, x1, y1, w, t float64) {
	dx, dy := x1-x0, y1-y0
	l := math.Hypot(dx, dy)
	if l == 0 {
		return
	}
	nx, ny := -dy/l*w/2, dx/l*w/2
	c.poly(t, [2]float64{x0 + nx, y0 + ny}, [2]float64{x1 + nx, y1 + ny},
		[2]float64{x1 - nx, y1 - ny}, [2]float64{x0 - nx, y0 - ny})
}

// flash is a muzzle flash of radius r: a bright core with spikes.
func (c canvas) flash(cx, cy, r float64) {
	for i := range 8 {
		a := float64(i) * math.Pi / 4
		l := r * 1.8
		if i%2 == 1 {
			l = r * 1.1
		}
		c.line(cx, cy, cx+l*math.Cos(a), cy+l*math.Sin(a), math.Max(1, r/4), 0.95)
	}
	c.circle(cx, cy, r*0.7, 1)
}

// airScene: a biplane crossing the sky between drifting clouds, firing over
// its upper wing.
func airScene(c canvas, f int) {
	c.backdrop(160)
	c.poly(0.32, [2]float64{0, 180}, [2]float64{0, 156}, [2]float64{40, 148}, [2]float64{90, 158},
		[2]float64{130, 146}, [2]float64{152, 152}, [2]float64{152, 180})
	ff := float64(f)
	c.ellipse(34-2*ff, 38, 18, 6, 0.95)
	c.ellipse(116-3*ff, 62, 22, 7, 0.93)
	c.ellipse(70-2*ff, 140, 26, 6, 0.9)

	bob := [vignetteFrames]float64{0, 1, 2, 1, 0, -1}
	px, py := 58+4*ff, 90+bob[f]
	pt := func(x, y float64) [2]float64 { return [2]float64{px + x, py + y} }
	c.poly(0.2, pt(-34, -3), pt(-30, -14), pt(-24, -14), pt(-22, -4))
	c.rect(px-38, py-1, px-22, py+2, 0.2)
	c.poly(0.16, pt(-34, -3), pt(22, -6), pt(30, -2), pt(30, 4), pt(22, 6), pt(-34, 1))
	c.rect(px-10, py-20, px+22, py-16, 0.12)
	c.rect(px-6, py+4, px+20, py+7, 0.14)
	c.line(px-6, py-16, px-4, py+4, 1.2, 0.1)
	c.line(px+16, py-16, px+16, py+4, 1.2, 0.1)
	c.line(px+8, py+6, px+5, py+15, 1.2, 0.1)
	c.line(px+12, py+6, px+14, py+15, 1.2, 0.1)
	c.circle(px+9, py+16, 3.5, 0.08)
	c.circle(px+4, py-9, 3.2, 0.24)
	if f%2 == 0 {
		c.line(px+31, py-14, px+31, py+16, 1.6, 0.1)
	} else {
		c.ellipse(px+31, py+1, 1.5, 13, 0.55)
	}
	if f == 2 || f == 3 {
		c.flash(px+26, py-18, 4)
		c.line(px+32, py-18, px+60, py-19, 1, 0.98)
	}
}

// navalScene: a warship on a rolling sea, firing to the right with a splash
// rising on the horizon.
func navalScene(c canvas, f int) {
	const horizon = 92
	c.backdrop(horizon)
	ff := float64(f)
	for y := horizon; y < VignetteH; y++ {
		for x := range VignetteW {
			d := float64(y-horizon) / float64(VignetteH-horizon)
			w := math.Sin(float64(x)*0.22+float64(y)*0.7-ff*0.9) * math.Sin(float64(y)*0.35)
			c.set(x, y, 0.5-0.26*d+0.07*w)
		}
	}
	for i := range 3 {
		fi := float64(i)
		c.ellipse(24-3*ff-12*fi, 78-6*fi, 6+2*fi, 4+fi, 0.6+0.08*fi)
	}
	c.poly(0.12, [2]float64{6, 108}, [2]float64{118, 108}, [2]float64{112, 126}, [2]float64{14, 126})
	c.rect(40, 92, 72, 108, 0.2)
	c.rect(46, 84, 64, 92, 0.18)
	c.line(55, 84, 55, 58, 1.4, 0.1)
	c.line(46, 64, 64, 64, 1, 0.1)
	c.rect(22, 86, 29, 108, 0.15)
	c.rect(31, 88, 38, 108, 0.15)
	c.rect(84, 100, 98, 108, 0.18)
	c.line(96, 103, 118, 97, 2, 0.1)
	for x := 10.0; x < 116; x += 7 {
		c.circle(x, 116, 1, 0.35)
	}
	switch f {
	case 1, 2:
		c.flash(121, 96, 6-2*float64(f-1))
	case 3, 4, 5:
		h := [3]float64{6, 14, 8}[f-3]
		c.poly(0.9, [2]float64{136, horizon + 1}, [2]float64{139, horizon - h}, [2]float64{142, horizon - h + 2},
			[2]float64{145, horizon + 1})
	}
}

// artilleryScene: a field gun recoiling as it fires, its smoke drifting up.
func artilleryScene(c canvas, f int) {
	c.backdrop(118)
	rc := [vignetteFrames]float64{0, 4, 2, 1, 0, 0}[f]
	c.line(50, 128, 10, 150, 5, 0.16)
	c.rect(6, 148, 16, 154, 0.12)
	c.poly(0.18, [2]float64{44, 118}, [2]float64{76, 110}, [2]float64{80, 122}, [2]float64{48, 130})
	c.line(58-rc*0.9, 124+rc*0.4, 128-rc*0.9, 96+rc*0.4, 5, 0.12)
	c.poly(0.22, [2]float64{66, 98}, [2]float64{80, 94}, [2]float64{84, 126}, [2]float64{70, 130})
	c.circle(58, 134, 17, 0.14)
	c.circle(58, 134, 14, 0.5)
	for i := range 8 {
		a := float64(i) * math.Pi / 4
		c.line(58, 134, 58+14*math.Cos(a), 134+14*math.Sin(a), 1.6, 0.14)
	}
	c.circle(58, 134, 3, 0.1)
	if f == 1 || f == 2 {
		c.flash(130, 95, 12-4*float64(f-1))
	}
	if f >= 2 {
		s := float64(f - 2)
		c.ellipse(132+6*s, 90-5*s, 6+3*s, 5+2*s, 0.76)
		c.ellipse(124+4*s, 86-7*s, 4+2*s, 3+2*s, 0.82)
	}
}

// armourScene: a rhomboid tank rolling to the right, firing from its
// sponson.
func armourScene(c canvas, f int) {
	c.backdrop(104)
	for i := range 5 {
		y := 160 + 4*float64(i)
		c.line(0, y, 152, y-6, 1, 0.2)
	}
	tx := 3*float64(f) - 8
	pt := func(x, y float64) [2]float64 { return [2]float64{tx + x, y} }
	for i := range 3 {
		fi := float64(i)
		c.ellipse(tx+14-5*fi, 96-7*fi, 5+2*fi, 4+fi, 0.62+0.06*fi)
	}
	c.poly(0.14, pt(14, 140), pt(34, 104), pt(112, 100), pt(140, 136), pt(124, 152), pt(26, 152))
	c.poly(0.24, pt(24, 138), pt(38, 112), pt(108, 108), pt(130, 136), pt(120, 145), pt(32, 145))
	for i := range 12 {
		x := tx + 30 + 8*float64(i) + 4*float64(f%2)
		c.line(x, 146, x, 152, 1.2, 0.08)
	}
	c.rect(tx+52, 94, tx+76, 102, 0.18)
	c.rect(tx+60, 118, tx+84, 136, 0.2)
	c.line(tx+84, 124, tx+106, 122, 3, 0.1)
	if f == 3 || f == 4 {
		c.flash(tx+109, 122, 7-2*float64(f-3))
	}
}

// infantryScene: three riflemen behind a sandbag parapet, firing in turn.
func infantryScene(c canvas, f int) {
	c.backdrop(100)
	for x := 6.0; x < 152; x += 22 {
		c.line(x, 100, x+1, 90, 1, 0.3)
		c.line(x-6, 94, x+8, 96, 0.6, 0.3)
	}
	shooter := [vignetteFrames]int{0, -1, 2, -1, 1, -1}[f]
	for i, x := range []float64{34, 76, 118} {
		c.poly(0.18, [2]float64{x - 11, 132}, [2]float64{x - 7, 114}, [2]float64{x + 7, 114},
			[2]float64{x + 11, 132})
		c.circle(x, 108, 4.5, 0.3)
		c.circle(x, 104, 5, 0.12)
		c.ellipse(x, 105, 8, 2, 0.12)
		c.line(x-2, 121, x+24, 109, 2, 0.08)
		c.line(x-4, 124, x+4, 119, 3, 0.2)
		if i == shooter {
			c.flash(x+27, 108, 5)
		}
	}
	c.rect(0, 134, VignetteW, VignetteH, 0.1)
	for row, y := range []float64{128, 134} {
		for x := -8.0 + 11*float64(row); x < 160; x += 22 {
			c.ellipse(x, y, 11, 5, 0.34+0.04*float64(row))
		}
	}
}
