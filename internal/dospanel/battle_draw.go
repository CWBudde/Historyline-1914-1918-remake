package dospanel

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BattleSideRects are the parts of one side of the battle scene in device
// pixels: its half of the panel, the photo's recess and the photo, the name
// plate, the hex with the figures, and on the photo the flag, the badge and
// the count plate.
type BattleSideRects struct {
	Half, PhotoWell, Photo, Name, Hex image.Rectangle
	Flag, Badge, Count                image.Rectangle
}

// BattleRects are the battle scene's parts inside Panel: the attacker's
// side above the defender's, mirrored, so the two hexes face each other
// across the panel. Scale is the device pixels per reference pixel.
type BattleRects struct {
	Scale float64
	Panel image.Rectangle
	Sides [2]BattleSideRects
}

// The battle scene in reference pixels: a 152×180 block of two halves. The
// attacker's has the photo at the left and the hex at the right; the
// defender's is its mirror image. The hex is 60×50, a map hex's shape,
// padded above and below in its 60×82 slot, and holds the ring.
const (
	battleW, battleH = 152, 180
	battleHalfH      = battleH / 2
	ringHexW         = 60
	figureHalf       = 10.5
)

// ring is the places of the figures around the hex's centre in reference
// pixels: top, upper right, lower right, bottom, lower left, upper left.
// The ring is as flat as the hex, so the men stand on it.
var ring = [MaxFigures][2]float64{{0, -13}, {17, -6.5}, {17, 6.5}, {0, 13}, {-17, 6.5}, {-17, -6.5}}

// BattleLayout lays out the battle scene centred in panel at scale, or the
// largest scale that fits when that is smaller.
func BattleLayout(panel image.Rectangle, scale float64) BattleRects {
	g := fit(panel, scale, battleW, battleH)
	l := BattleRects{Scale: g.k, Panel: panel}
	for s := range l.Sides {
		r := func(x0, y0, x1, y1 float64) image.Rectangle {
			if s == 1 {
				x0, x1 = battleW-x1, battleW-x0
			}
			oy := float64(s * battleHalfH)
			return g.rect(x0, y0+oy, x1, y1+oy)
		}
		l.Sides[s] = BattleSideRects{
			Half:      r(2, 2, 150, 88),
			PhotoWell: r(5, 5, 85, 73),
			Photo:     r(7, 7, 83, 71),
			Name:      r(5, 75, 85, 86),
			Hex:       r(88, 20, 148, 70),
			Flag:      r(10, 10, 25, 25),
			Badge:     r(27, 9, 44, 26),
			Count:     r(54, 50, 80, 68),
		}
	}
	return l
}

// scale is the device pixels per reference pixel of the side's hex.
func (r BattleSideRects) scale() float64 { return float64(r.Hex.Dx()) / ringHexW }

// centre is the device centre of the side's hex.
func (r BattleSideRects) centre() (x, y float64) {
	return float64(r.Hex.Min.X+r.Hex.Max.X) / 2, float64(r.Hex.Min.Y+r.Hex.Max.Y) / 2
}

// spot is the device centre of figure i of a group; on a ship a place
// along its hull.
func (r BattleSideRects) spot(i int, ship bool) (x, y float64) {
	cx, cy := r.centre()
	k := r.scale()
	i = min(max(i, 0), MaxFigures-1)
	if ship {
		return cx + k*(float64(i%3)-1)*10, cy + k*(float64(i/3)*6-4)
	}
	return cx + k*ring[i][0], cy + k*ring[i][1]
}

// Figure is the device box of figure i on the ring.
func (r BattleSideRects) Figure(i int) image.Rectangle {
	x, y := r.spot(i, false)
	h := figureHalf * r.scale()
	return image.Rect(int(math.Round(x-h)), int(math.Round(y-h)), int(math.Round(x+h)), int(math.Round(y+h)))
}

// The dark sepia ground of the hex, from shadow to its lit centre.
var (
	fieldDark  = color.RGBA{0x14, 0x0E, 0x08, 0xFF}
	fieldLight = color.RGBA{0x7A, 0x5E, 0x3C, 0xFF}
)

// field is the hex of half size a×b around (cx, cy): a recessed steel rim
// around dark sepia ground, lit in the middle.
func (p painter) field(cx, cy, a, b float64) {
	rim := 2 * float64(p.bw())
	q := max(1, int(math.Round(p.k)))
	for y := int(cy - b - 1); y <= int(cy+b+1); y++ {
		for x := int(cx - a - 1); x <= int(cx+a+1); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			n := hexNorm(dx, dy, a, b)
			if n > 1 || !(image.Point{x, y}).In(p.img.Rect) {
				continue
			}
			if d, light := hexEdge(dx, dy, a, b); d < rim {
				p.img.SetRGBA(x, y, ramp(p.steelAt(x, y, toneBack-bevelTone*light)))
				continue
			}
			t := 0.12 + 0.5*(1-n)*(1-n) + 0.05*grain(x/q, y/q) + 0.03*grain(x/(4*q), y/(3*q))
			p.img.SetRGBA(x, y, lerp(fieldDark, fieldLight, t))
		}
	}
}

// paintBattle paints the plates of the battle scene.
func paintBattle(l BattleRects) *image.RGBA {
	img := image.NewRGBA(l.Panel)
	p := painter{img, l.Scale}
	p.steel(img.Rect, toneBack)
	for _, s := range l.Sides {
		p.plate(s.Half, tonePlate)
		p.well(s.PhotoWell)
		p.plate(s.Photo, toneBlank)
		p.namePlate(s.Name)
		cx, cy, a, b := hexOf(s.Hex)
		p.field(cx, cy, a, b)
	}
	// Rivets along the seam between the halves.
	g := fit(l.Panel, l.Scale, battleW, battleH)
	for x := 10.0; x < battleW; x += 22 {
		p.rivet(float64(g.x(x)), float64(g.y(battleHalfH)), math.Max(1.2, 1.3*l.Scale))
	}
	return img
}

// Colours of the scene's effects and the count.
var (
	countFloor = color.RGBA{0x1E, 0x16, 0x0E, 0xFF}
	hitRed     = color.RGBA{0xFF, 0x5A, 0x38, 0xFF}
	flashCore  = color.RGBA{0xFF, 0xF6, 0xDC, 0xFF}
	fireAmber  = color.RGBA{0xF4, 0xA8, 0x40, 0xFF}
	smokeBrown = color.RGBA{0x3A, 0x2C, 0x1E, 0xFF}
	dustTan    = color.RGBA{0xB8, 0x9C, 0x74, 0xFF}
)

// DrawBattle draws battle b's scene at tick over panel at scale (see
// BattleLayout), on steel that fills the panel: the plates, each
// side's photo, flag, badge, count and name, the figures on the hexes and
// the shots. Under reduced motion it shows the end. art may be empty; text
// may be nil.
func DrawBattle(dst *ebiten.Image, panel image.Rectangle, scale float64, b Battle, tick int, reduced bool, art BattleArt, text Text) {
	if panel.Empty() {
		return
	}
	size := panel.Size()
	l0 := BattleLayout(image.Rectangle{Max: size}, scale)
	drawPlate(dst, panel, cachedPlate(battlePlate, size, l0.Scale, func() *image.RGBA { return paintBattle(l0) }))

	l := BattleLayout(panel, scale)
	st := BattleAt(b, tick, reduced)
	c := clip(dst, panel)
	w := writer{c, text}
	fighters := b.sides()
	for s, f := range fighters {
		drawSide(c, w, l, s, f, st)
	}
	for _, fl := range st.Flash {
		drawFlash(c, l, fighters, fl, art)
	}
}

// drawSide draws side s of the scene: photo, flag, badge, count, name, the
// ring's places and the figures.
func drawSide(c *ebiten.Image, w writer, l BattleRects, s int, f Fighter, st BattleState) {
	r, ss, k := l.Sides[s], st.Sides[s], l.Scale
	drawPhoto(c, r.Photo, f.Photo)
	drawFlag(c, r.Flag, f, k)
	drawBadge(c, r, s, f, st, k)
	drawCount(c, r, ss, k)
	cx, cy := centre2(r.Name)
	w.centre(f.Name, int(cx), int(cy), cream)

	if f.Ship {
		drawShip(c, r, f.Sprite, ss)
		return
	}
	// The ring's places a full group would fill: the men missing from it
	// leave empty foxholes.
	places := min(max(f.Full, ss.Shown, 0), MaxFigures)
	for i := range places {
		x, y := r.spot(i, false)
		hk := r.scale()
		fillEllipse(c, x, y+7*hk, 7*hk, 2.6*hk, color.NRGBA{0, 0, 0, 0x50})
	}
	// Figures from the top of the ring down, so the lower ones stand in
	// front.
	for _, i := range [...]int{0, 5, 1, 4, 2, 3} {
		if i >= ss.Shown {
			continue
		}
		lost := i >= ss.Shown-ss.Lost
		rise := 1.0
		if i == ss.Shown-1 {
			rise = ss.Rise
		}
		drawFigure(c, r, i, f.Sprite, lost, rise)
	}
}

// centre2 is the device centre of r.
func centre2(r image.Rectangle) (float64, float64) {
	return float64(r.Min.X+r.Max.X) / 2, float64(r.Min.Y+r.Max.Y) / 2
}

// drawFlag draws the nation's flag, or a bar in its colour, on a dark edge.
func drawFlag(c *ebiten.Image, r image.Rectangle, f Fighter, k float64) {
	if f.Flag != nil {
		drawSprite(c, r, f.Flag, 1)
		return
	}
	if f.FlagColour.A == 0 {
		return
	}
	bw := float32(math.Max(1, math.Round(k)))
	h := float32(r.Dy()) * 0.7
	y := float32(r.Min.Y) + (float32(r.Dy())-h)/2
	vector.FillRect(c, float32(r.Min.X), y, float32(r.Dx()), h, plateEdge, false)
	vector.FillRect(c, float32(r.Min.X)+bw, y+bw, float32(r.Dx())-2*bw, h-2*bw, f.FlagColour, false)
}

// drawBadge draws the experience badge; a gain switches it to the new one
// with a glow and a "+1".
func drawBadge(c *ebiten.Image, r BattleSideRects, s int, f Fighter, st BattleState, k float64) {
	if f.Badge == nil {
		return
	}
	exp := f.ExpBefore
	if st.Gained[s] {
		exp = f.ExpAfter
	}
	img := f.Badge(exp)
	gl := st.Glow[s]
	cx, cy := centre2(r.Badge)
	if img == nil && gl <= 0 {
		return
	}
	// A dark medallion with a brass rim carries the badge on the photo.
	rad := float64(r.Badge.Dx()) / 2
	fillCircle(c, cx, cy, rad, color.NRGBA{plateEdge.R, plateEdge.G, plateEdge.B, 0xFF}, ebiten.BlendSourceOver)
	fillCircle(c, cx, cy, rad-math.Max(1, k*0.8), color.NRGBA{brass.R, brass.G, brass.B, 0xFF}, ebiten.BlendSourceOver)
	fillCircle(c, cx, cy, rad-math.Max(2, k*1.8), color.NRGBA{countFloor.R, countFloor.G, countFloor.B, 0xFF}, ebiten.BlendSourceOver)
	if gl > 0 {
		glow(c, cx, cy, rad*(1.6+0.8*gl), fireAmber, 0.9*gl)
	}
	if img != nil {
		// The badges are drawn with a wide margin: twice the medallion fits
		// the mark inside it.
		drawImageAt(c, img, cx, cy, 2*rad*1.9*(1+0.25*gl), 0, nil, 1)
	}
	if gl > 0 {
		q := math.Max(1, math.Round(k*1.1))
		gw := glyphsWidth("+1", q)
		x := float64(r.Badge.Max.X) + k
		if s == 1 {
			x = float64(r.Badge.Min.X) - k - gw
		}
		drawGlyphs(c, "+1", x, cy-3.5*q-(1-gl)*4*k, q, lerp(amber, flashCore, gl))
	}
}

// glyphs are the scene's large figures on a 5×7 grid.
var glyphs = map[rune][7]string{
	'0': {" ### ", "#   #", "#  ##", "# # #", "##  #", "#   #", " ### "},
	'1': {"  #  ", " ##  ", "  #  ", "  #  ", "  #  ", "  #  ", " ### "},
	'2': {" ### ", "#   #", "    #", "   # ", "  #  ", " #   ", "#####"},
	'3': {"#####", "   # ", "  #  ", "   # ", "    #", "#   #", " ### "},
	'4': {"   # ", "  ## ", " # # ", "#  # ", "#####", "   # ", "   # "},
	'5': {"#####", "#    ", "#### ", "    #", "    #", "#   #", " ### "},
	'6': {"  ## ", " #   ", "#    ", "#### ", "#   #", "#   #", " ### "},
	'+': {"     ", "  #  ", "  #  ", "#####", "  #  ", "  #  ", "     "},
}

// glyphsWidth is the width of s in glyphs of cell size q.
func glyphsWidth(s string, q float64) float64 {
	n := len([]rune(s))
	if n == 0 {
		return 0
	}
	return q * float64(6*n-1)
}

// drawGlyphs draws s with its top left at (x, y) in cells of size q, with a
// shadow below right; runes without a glyph leave a gap.
func drawGlyphs(c *ebiten.Image, s string, x, y, q float64, col color.Color) {
	sh := math.Max(1, math.Round(q/2))
	for pass, cc := range []color.Color{color.RGBA{0, 0, 0, 0xFF}, col} {
		off := sh * float64(1-pass)
		for n, r := range []rune(s) {
			g, ok := glyphs[r]
			if !ok {
				continue
			}
			gx := x + off + float64(6*n)*q
			for row, line := range g {
				for i, ch := range line {
					if ch == '#' {
						vector.FillRect(c, float32(gx+float64(i)*q), float32(y+off+float64(row)*q),
							float32(q), float32(q), cc, false)
					}
				}
			}
		}
	}
}

// drawCount draws the count plate: the group's size in large figures and,
// once men fell, the size before the battle small and struck out; the
// figures flash red as a man falls.
func drawCount(c *ebiten.Image, r BattleSideRects, ss SideState, k float64) {
	rr := r.Count
	bw := float32(math.Max(1, math.Round(k)))
	x, y, wd, ht := float32(rr.Min.X), float32(rr.Min.Y), float32(rr.Dx()), float32(rr.Dy())
	vector.FillRect(c, x, y, wd, ht, plateEdge, false)
	vector.FillRect(c, x+bw, y+bw, wd-2*bw, ht-2*bw, brassDark, false)
	vector.FillRect(c, x+2*bw, y+2*bw, wd-4*bw, ht-4*bw, countFloor, false)

	q := math.Max(1, math.Floor(float64(rr.Dy())*0.62/7))
	big := strconv.Itoa(ss.Count)
	dy := float64(rr.Min.Y) + (float64(rr.Dy())-7*q)/2
	col := lerp(amber, hitRed, ss.Hit)
	if ss.Lost == 0 {
		drawGlyphs(c, big, float64(rr.Min.X)+(float64(rr.Dx())-glyphsWidth(big, q))/2, dy, q, col)
		return
	}
	drawGlyphs(c, big, float64(rr.Min.X)+3.5*k, dy, q, col)
	old := strconv.Itoa(ss.Count + ss.Lost)
	sq := math.Max(1, math.Round(q*0.55))
	ox := float64(rr.Max.X) - 3.5*k - glyphsWidth(old, sq)
	oy := float64(rr.Max.Y) - 3.5*k - 7*sq
	drawGlyphs(c, old, ox, oy, sq, lerp(cream, plateBrown, 0.25))
	vector.StrokeLine(c, float32(ox-sq), float32(oy+8*sq), float32(ox+glyphsWidth(old, sq)+sq), float32(oy-sq),
		float32(math.Max(1, sq*0.9)), hitRed, true)
}

// fallback figures for a missing sprite: a rifleman and a warship, drawn
// once.
var fallback struct {
	once      sync.Once
	man, ship *ebiten.Image
}

func fallbackArt() (man, ship *ebiten.Image) {
	fallback.once.Do(func() {
		fallback.man = paintMan()
		fallback.ship = paintShip()
	})
	return fallback.man, fallback.ship
}

// paintMan paints a rifleman in grey, drawn on a 48 grid, at 4× with a
// margin like the map sprites'.
func paintMan() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 192, 192))
	const k = 4 * 0.62
	c := shapes{img, k, 96 - k*24, 96 - k*28}
	grey := func(v uint8) color.RGBA { return color.RGBA{v, v, v, 0xFF} }
	c.poly(grey(0x58), [2]float64{17, 40}, [2]float64{19, 20}, [2]float64{29, 20}, [2]float64{31, 40})
	c.ellipse(24, 16, 4, 4.4, grey(0xA8))
	c.ellipse(24, 12.5, 7, 2.2, grey(0x48))
	c.ellipse(24, 11.5, 4.5, 3, grey(0x48))
	c.line(14, 34, 36, 10, 2.2, grey(0x2C))
	c.line(19, 40, 18, 46, 3, grey(0x40))
	c.line(29, 40, 30, 46, 3, grey(0x40))
	return ebiten.NewImageFromImage(img)
}

// paintShip paints a warship in grey, drawn on a 96×48 grid, at 4× with a
// margin.
func paintShip() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 512, 320))
	const k = 4 * 1.1
	c := shapes{img, k, 256 - k*48, 160 - k*24}
	grey := func(v uint8) color.RGBA { return color.RGBA{v, v, v, 0xFF} }
	c.poly(grey(0x50), [2]float64{2, 30}, [2]float64{94, 30}, [2]float64{86, 40}, [2]float64{10, 40})
	c.poly(grey(0x68), [2]float64{30, 30}, [2]float64{32, 22}, [2]float64{62, 22}, [2]float64{64, 30})
	c.poly(grey(0x40), [2]float64{40, 22}, [2]float64{41, 12}, [2]float64{45, 12}, [2]float64{46, 22})
	c.poly(grey(0x40), [2]float64{50, 22}, [2]float64{51, 13}, [2]float64{55, 13}, [2]float64{56, 22})
	c.line(66, 26, 80, 22, 2, grey(0x38))
	c.line(28, 26, 14, 22, 2, grey(0x38))
	c.line(48, 12, 48, 2, 1, grey(0x30))
	return ebiten.NewImageFromImage(img)
}

// shapes fills shapes on a CPU image by pixel centres; a shape's
// coordinates are scaled by k and moved by (ox, oy).
type shapes struct {
	img       *image.RGBA
	k, ox, oy float64
}

func (s shapes) at(x, y float64) (float64, float64) { return s.ox + x*s.k, s.oy + y*s.k }

func (s shapes) region(x0, y0, x1, y1 float64, c color.RGBA, in func(x, y float64) bool) {
	for y := int(math.Floor(y0)); y <= int(math.Ceil(y1)); y++ {
		for x := int(math.Floor(x0)); x <= int(math.Ceil(x1)); x++ {
			if (image.Point{x, y}).In(s.img.Rect) && in(float64(x)+0.5, float64(y)+0.5) {
				s.img.SetRGBA(x, y, c)
			}
		}
	}
}

func (s shapes) ellipse(cx, cy, rx, ry float64, c color.RGBA) {
	cx, cy = s.at(cx, cy)
	rx, ry = rx*s.k, ry*s.k
	s.region(cx-rx, cy-ry, cx+rx, cy+ry, c, func(x, y float64) bool {
		dx, dy := (x-cx)/rx, (y-cy)/ry
		return dx*dx+dy*dy <= 1
	})
}

func (s shapes) poly(c color.RGBA, pts ...[2]float64) {
	dev := make([][2]float64, len(pts))
	for i, p := range pts {
		dev[i][0], dev[i][1] = s.at(p[0], p[1])
	}
	s.devPoly(c, dev)
}

// devPoly fills a polygon in device pixels by the even-odd rule.
func (s shapes) devPoly(c color.RGBA, pts [][2]float64) {
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range pts {
		x0, y0 = math.Min(x0, p[0]), math.Min(y0, p[1])
		x1, y1 = math.Max(x1, p[0]), math.Max(y1, p[1])
	}
	s.region(x0, y0, x1, y1, c, func(x, y float64) bool {
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

func (s shapes) line(x0, y0, x1, y1, w float64, c color.RGBA) {
	dx, dy := x1-x0, y1-y0
	l := math.Hypot(dx, dy)
	if l == 0 {
		return
	}
	nx, ny := -dy/l*w/2, dx/l*w/2
	s.poly(c, [2]float64{x0 + nx, y0 + ny}, [2]float64{x1 + nx, y1 + ny},
		[2]float64{x1 - nx, y1 - ny}, [2]float64{x0 - nx, y0 - ny})
}

// drawImageAt draws img centred on (cx, cy), scaled so its larger side is
// size, turned by rot radians, through the colour matrix m (nil: as it is)
// at opacity alpha.
func drawImageAt(c *ebiten.Image, img *ebiten.Image, cx, cy, size, rot float64, m *colorm.ColorM, alpha float64) {
	if img == nil || size <= 0 || alpha <= 0 {
		return
	}
	b := img.Bounds()
	if b.Empty() {
		return
	}
	s := size / float64(max(b.Dx(), b.Dy()))
	var geo ebiten.GeoM
	geo.Translate(-float64(b.Min.X)-float64(b.Dx())/2, -float64(b.Min.Y)-float64(b.Dy())/2)
	geo.Scale(s, s)
	geo.Rotate(rot)
	geo.Translate(cx, cy)
	if m == nil {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear, GeoM: geo}
		op.ColorScale.ScaleAlpha(float32(alpha))
		c.DrawImage(img, op)
		return
	}
	mm := *m
	mm.Scale(1, 1, 1, alpha)
	colorm.DrawImage(c, img, mm, &colorm.DrawImageOptions{Filter: ebiten.FilterLinear, GeoM: geo})
}

// toned is the sepia matrix brightened or darkened by v.
func toned(v float64) *colorm.ColorM {
	m := sepia()
	m.Scale(v, v, v, 1)
	return &m
}

func fillCircle(c *ebiten.Image, cx, cy, r float64, col color.NRGBA, blend ebiten.Blend) {
	if r <= 0 || col.A == 0 {
		return
	}
	var p vector.Path
	p.Arc(float32(cx), float32(cy), float32(r), 0, 2*math.Pi, vector.Clockwise)
	p.Close()
	fillPath(c, &p, col, blend)
}

func fillEllipse(c *ebiten.Image, cx, cy, rx, ry float64, col color.NRGBA) {
	if rx <= 0 || ry <= 0 || col.A == 0 {
		return
	}
	var p vector.Path
	const n = 20
	for i := range n {
		a := 2 * math.Pi * float64(i) / n
		x, y := float32(cx+rx*math.Cos(a)), float32(cy+ry*math.Sin(a))
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	fillPath(c, &p, col, ebiten.BlendSourceOver)
}

func fillPath(c *ebiten.Image, p *vector.Path, col color.NRGBA, blend ebiten.Blend) {
	op := &vector.DrawPathOptions{AntiAlias: true, Blend: blend}
	op.ColorScale.ScaleWithColor(color.RGBA{col.R, col.G, col.B, 0xFF})
	op.ColorScale.ScaleAlpha(float32(col.A) / 0xFF)
	vector.FillPath(c, p, &vector.FillOptions{}, op)
}

func ease(t float64) float64 { return 1 - (1-t)*(1-t) }

// drawFigure draws figure i of a side: standing in sepia, rising into its
// place as it arrives, or a dark wreck on a scorch mark.
func drawFigure(c *ebiten.Image, r BattleSideRects, i int, sprite *ebiten.Image, lost bool, rise float64) {
	if sprite == nil {
		sprite, _ = fallbackArt()
	}
	x, y := r.spot(i, false)
	hk := r.scale()
	size := 2 * figureHalf * hk * 1.25
	if lost {
		fillEllipse(c, x, y+4*hk, 10*hk, 4.2*hk, color.NRGBA{0x08, 0x05, 0x02, 0xB0})
		fillEllipse(c, x-2*hk, y+3*hk, 5*hk, 2*hk, color.NRGBA{0, 0, 0, 0xC0})
		tilt := 0.45
		if i%2 == 1 {
			tilt = -0.45
		}
		drawImageAt(c, sprite, x, y+3*hk, size*0.92, tilt, toned(0.32), 1)
		return
	}
	e := ease(rise)
	drawImageAt(c, sprite, x, y-(1-e)*6*hk, size*(1.2-0.2*e), 0, toned(0.95), e)
}

// drawShip draws a ship's one large picture across its hex, a dark wreck
// listing once the last man is lost.
func drawShip(c *ebiten.Image, r BattleSideRects, sprite *ebiten.Image, ss SideState) {
	if ss.Shown == 0 {
		return
	}
	var ship *ebiten.Image
	if sprite == nil {
		_, ship = fallbackArt()
	} else {
		ship = sprite
	}
	x, y := r.centre()
	hk := r.scale()
	size := 64 * hk
	if ss.Lost >= ss.Shown {
		fillEllipse(c, x, y+8*hk, 26*hk, 7*hk, color.NRGBA{0x08, 0x05, 0x02, 0xB0})
		drawImageAt(c, ship, x, y+5*hk, size, 0.16, toned(0.32), 1)
		return
	}
	e := ease(ss.Rise)
	drawImageAt(c, ship, x, y, size*(1.1-0.1*e), 0, toned(0.95), e)
}

// flashAt is the device centre an effect sits on: figure fig of side s.
func flashAt(l BattleRects, fighters [2]Fighter, s, fig int) (x, y float64) {
	if fig < 0 {
		return l.Sides[s].centre()
	}
	return l.Sides[s].spot(fig, fighters[s].Ship)
}

// drawFlash draws one effect: a muzzle flash pointing at the other hex, a
// tracer, a blast (the art's frames in sepia or a drawn fireball) or a
// puff of dust.
func drawFlash(c *ebiten.Image, l BattleRects, fighters [2]Fighter, f Flash, art BattleArt) {
	if f.Side < 0 || f.Side > 1 {
		return
	}
	k := l.Sides[f.Side].scale()
	p := f.Progress
	switch f.Kind {
	case Muzzle:
		x, y := flashAt(l, fighters, f.Side, f.Figure)
		tx, ty := l.Sides[1-f.Side].centre()
		dx, dy := tx-x, ty-y
		d := math.Max(1, math.Hypot(dx, dy))
		dx, dy = dx/d, dy/d
		x, y = x+dx*7*k, y+dy*7*k
		env := p / 0.2
		if p > 0.2 {
			env = (1 - p) / 0.8
		}
		drawMuzzle(c, x, y, dx, dy, 9*k*env, env)
	case Blast:
		x, y := flashAt(l, fighters, f.Side, f.Figure)
		size := 40 * k
		if fighters[f.Side].Ship {
			size = 50 * k
		}
		drawBlast(c, x, y-2*k, size, p, art)
	case Miss:
		x, y := flashAt(l, fighters, f.Side, f.Figure)
		side := 1.0
		if f.Figure%2 == 0 {
			side = -1
		}
		drawDust(c, x+side*11*k, y+5*k, 9*k, p)
	}
}

// drawMuzzle draws a star of fire at (x, y) of radius r, long towards
// (dx, dy), at strength a.
func drawMuzzle(c *ebiten.Image, x, y, dx, dy, r, a float64) {
	if r <= 0 || a <= 0 {
		return
	}
	a = math.Min(1, a)
	glow(c, x, y, r*2.6, fireAmber, 0.9*a)
	// The blast's outline in the shot's frame: angle from the shot's
	// direction and reach, a long tongue forward, short tongues aside.
	tongues := [...][2]float64{
		{0, 1.7}, {0.2, 0.45}, {0.42, 1.0}, {0.7, 0.4}, {1.15, 0.62}, {1.6, 0.3}, {2.4, 0.34}, {math.Pi, 0.26},
	}
	var path vector.Path
	pt := func(ang, reach float64, first bool) {
		u, v := math.Cos(ang)*reach*r, math.Sin(ang)*reach*r
		px, py := x+u*dx-v*dy, y+u*dy+v*dx
		if first {
			path.MoveTo(float32(px), float32(py))
		} else {
			path.LineTo(float32(px), float32(py))
		}
	}
	for i, t := range tongues {
		pt(t[0], t[1], i == 0)
	}
	for i := len(tongues) - 2; i > 0; i-- {
		pt(-tongues[i][0], tongues[i][1], false)
	}
	path.Close()
	fillPath(c, &path, color.NRGBA{fireAmber.R, fireAmber.G, fireAmber.B, uint8(240 * a)}, ebiten.BlendSourceOver)
	fillCircle(c, x+dx*r*0.15, y+dy*r*0.15, r*0.36, color.NRGBA{flashCore.R, flashCore.G, flashCore.B, uint8(255 * a)}, ebiten.BlendSourceOver)
}

// glowSize is the side of the glow's falloff image.
const glowSize = 64

var glowTex struct {
	once sync.Once
	img  *ebiten.Image
}

// glowImage is a white disc fading from its centre to its edge.
func glowImage() *ebiten.Image {
	glowTex.once.Do(func() {
		img := image.NewRGBA(image.Rect(0, 0, glowSize, glowSize))
		for y := range glowSize {
			for x := range glowSize {
				d := math.Hypot(float64(x)+0.5-glowSize/2, float64(y)+0.5-glowSize/2) / (glowSize / 2)
				v := uint8(math.Round(255 * math.Pow(math.Max(0, 1-d), 2)))
				img.SetRGBA(x, y, color.RGBA{v, v, v, v})
			}
		}
		glowTex.img = ebiten.NewImageFromImage(img)
	})
	return glowTex.img
}

// glow adds a soft round light of col and radius r at (x, y), at strength
// a.
func glow(c *ebiten.Image, x, y, r float64, col color.RGBA, a float64) {
	if r <= 0 || a <= 0 {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear, Blend: ebiten.BlendLighter}
	op.GeoM.Scale(2*r/glowSize, 2*r/glowSize)
	op.GeoM.Translate(x-r, y-r)
	op.ColorScale.ScaleWithColor(col)
	op.ColorScale.ScaleAlpha(float32(math.Min(1, a)))
	c.DrawImage(glowImage(), op)
}

// fireTone is the explosion art's tone: sepia pushed towards amber.
func fireTone() *colorm.ColorM {
	m := sepia()
	m.Scale(1.08, 0.9, 0.66, 1)
	return &m
}

// blobs are the puffs of the drawn fireball: offset and size as shares of
// its radius.
var blobs = [...][3]float64{
	{0, 0, 0.62}, {-0.42, 0.12, 0.42}, {0.4, 0.18, 0.44}, {-0.16, -0.38, 0.46},
	{0.24, -0.34, 0.4}, {0.02, 0.36, 0.36}, {-0.5, -0.24, 0.3}, {0.52, -0.1, 0.32},
}

// drawBlast draws an explosion of size at (x, y) at progress p: the art's
// frames in an amber sepia, or a drawn fireball of puffs dying into smoke.
func drawBlast(c *ebiten.Image, x, y, size, p float64, art BattleArt) {
	fade := 1 - p
	glow(c, x, y, size*0.9*(0.6+0.4*fade), fireAmber, 0.5*fade*fade)
	if n := len(art.Explosion); n > 0 {
		if img := art.Explosion[min(int(p*float64(n)), n-1)]; img != nil {
			drawImageAt(c, img, x, y, size*1.25*(0.8+0.2*ease(p)), 0, fireTone(), math.Min(1, fade/0.2))
			return
		}
	}
	grow := ease(math.Min(1, p/0.55))
	rad := size * 0.5 * (0.3 + 0.7*grow)
	rise := rad * 0.35 * p
	smoke := uint8(230 * math.Min(1, fade*1.5))
	for _, b := range blobs {
		fillCircle(c, x+b[0]*rad, y+b[1]*rad-rise, b[2]*rad*1.15, color.NRGBA{smokeBrown.R, smokeBrown.G, smokeBrown.B, smoke}, ebiten.BlendSourceOver)
	}
	heat := math.Max(0, 1-p/0.75)
	for _, b := range blobs {
		fillCircle(c, x+b[0]*rad*0.9, y+b[1]*rad*0.9-rise, b[2]*rad*heat, color.NRGBA{fireAmber.R, fireAmber.G, fireAmber.B, uint8(255 * heat)}, ebiten.BlendSourceOver)
	}
	fillCircle(c, x, y-rise, rad*0.42*heat*heat, color.NRGBA{flashCore.R, flashCore.G, flashCore.B, 0xFF}, ebiten.BlendSourceOver)
	for i := range 7 {
		a := float64(i)*2*math.Pi/7 + 0.3
		d := rad * (0.7 + 0.9*grow)
		fillCircle(c, x+d*math.Cos(a), y+d*math.Sin(a)-rise*2, size*0.025*fade+0.5, color.NRGBA{smokeBrown.R, smokeBrown.G, smokeBrown.B, uint8(255 * fade)}, ebiten.BlendSourceOver)
	}
}

// drawDust draws the puff of a shot that missed: tan dust rising and
// thinning.
func drawDust(c *ebiten.Image, x, y, size, p float64) {
	a := uint8(200 * (1 - p))
	g := ease(p)
	for i, o := range [][2]float64{{0, 0}, {-0.45, -0.25}, {0.4, -0.35}} {
		rad := size * (0.3 + 0.35*g) * (1 - 0.15*float64(i))
		fillCircle(c, x+o[0]*size*g, y+o[1]*size*g-size*0.4*g, rad, color.NRGBA{dustTan.R, dustTan.G, dustTan.B, a}, ebiten.BlendSourceOver)
	}
}
