package dospanel

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

// Steel tokens shared with the menus (cmd/historyline): one palette, so
// menu -> battle is one game. Values are the battle ramp in paint.go.
var (
	SteelDark   = steelDark
	SteelLight  = steelLight
	SteelBrass  = brass
	SteelBrassL = brassLight
	SteelBrassD = brassDark
	PlateBrown  = plateBrown
	RingLight   = ringLight
	MenuCream   = cream
	MenuAmber   = amber
	MenuPale    = pale
	// MenuMuted is mid-ramp steel for disabled/secondary menu text.
	MenuMuted = ramp(0.62)
)

// SteelColour is the steel ramp colour at t (0 dark .. 1 light).
func SteelColour(t float64) color.RGBA { return ramp(t) }

// MenuButtonState is a steel menu button's state.
type MenuButtonState int

const (
	MenuNormal MenuButtonState = iota
	MenuFocus
	MenuPressed
	MenuDisabled
)

type menuKey struct {
	kind  int
	size  image.Point
	scale float64
	state MenuButtonState
}

const (
	menuBack = iota
	menuPlate
	menuWell
	menuName
	menuButton
	menuPanel
)

var menuCache = struct {
	sync.Mutex
	m map[menuKey]*ebiten.Image
}{m: map[menuKey]*ebiten.Image{}}

func menuImage(kind int, r image.Rectangle, scale float64, state MenuButtonState, paint func(p painter, rr image.Rectangle)) *ebiten.Image {
	if r.Empty() {
		return nil
	}
	if scale <= 0 {
		scale = 1
	}
	key := menuKey{kind, r.Size(), scale, state}
	menuCache.Lock()
	defer menuCache.Unlock()
	if img, ok := menuCache.m[key]; ok {
		return img
	}
	img := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	p := painter{img, scale}
	local := img.Rect
	switch kind {
	case menuBack:
		p.menuBackground(local)
	case menuPlate:
		p.menuPlate(local, tonePlate)
	case menuWell:
		p.menuWell(local)
	case menuName:
		p.menuNamePlate(local)
	case menuButton:
		p.menuButton(local, state)
	case menuPanel:
		p.menuPlate(local, 0.30)
	default:
		paint(p, local)
	}
	out := ebiten.NewImageFromImage(img)
	if len(menuCache.m) >= 32 {
		for k, v := range menuCache.m {
			v.Deallocate()
			delete(menuCache.m, k)
		}
	}
	menuCache.m[key] = out
	return out
}

func drawMenu(dst *ebiten.Image, r image.Rectangle, img *ebiten.Image) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	dst.DrawImage(img, op)
}

// menuSteel fills r with menu steel of tone t: a light falloff from the
// top left, horizontal brush streaks and the fine grain of steelColour.
func (p painter) menuSteel(r image.Rectangle, t float64) {
	r = r.Intersect(p.img.Rect)
	if r.Empty() {
		return
	}
	q := max(1, int(math.Round(p.k)))
	w, h := float64(r.Dx()), float64(r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		fy := float64(y-r.Min.Y) / h
		for x := r.Min.X; x < r.Max.X; x++ {
			fx := float64(x-r.Min.X) / w
			tone := t + 0.055*(1-fy) - 0.045*fx*fy
			tone += 0.016 * grain(x/(36*q), y/q) // brushed along x
			tone += 0.012 * grain(x/(3*q), y/(9*q))
			p.img.SetRGBA(x, y, p.steelColourTone(x, y, tone))
		}
	}
}

// steelColourTone is steelColour without its own coarse grain (menuSteel
// supplies the grain), keeping a painted surface's modulation when set.
func (p painter) steelColourTone(x, y int, t float64) color.RGBA {
	if s := surfaces.steel; s != nil {
		dr, dg, db := s.At(x, y, p.k*steelSpan)
		return texture.Modulate(ramp(t), steelGain, dr, dg, db)
	}
	return ramp(t)
}

// menuEdge draws a one-line border of colour c inset by n lines.
func (p painter) menuEdge(r image.Rectangle, n int, c color.RGBA) {
	edges(r.Inset(n*p.bw()), p.bw(), func(x, y int, _ bool) { p.img.SetRGBA(x, y, c) })
}

// menuRivets sets brass rivets in r's corners when it is big enough.
func (p painter) menuRivets(r image.Rectangle) {
	if r.Dx() < int(28*p.k) || r.Dy() < int(20*p.k) {
		return
	}
	in := 5 * p.k
	rad := math.Max(1.5, 1.7*p.k)
	for _, y := range []float64{float64(r.Min.Y) + in, float64(r.Max.Y) - in} {
		p.rivet(float64(r.Min.X)+in, y, rad)
		p.rivet(float64(r.Max.X)-in, y, rad)
	}
}

// menuPlate is a raised menu plate: brushed steel, a dark outer edge, a
// two-step bevel, a soft inner shade at the foot and corner rivets.
func (p painter) menuPlate(r image.Rectangle, t float64) {
	p.menuSteel(r, t)
	p.menuEdge(r, 0, ramp(t-0.38))
	p.bevel(r.Inset(p.bw()), t, true)
	// Foot shade: the lowest bevel line reads as cast shadow.
	edges(r.Inset(2*p.bw()), p.bw(), func(x, y int, lit bool) {
		if !lit && y > r.Min.Y+r.Dy()/2 {
			p.img.SetRGBA(x, y, ramp(t-0.20))
		}
	})
	p.menuRivets(r)
}

// menuWell is a recessed menu well: dark steel, sunk bevel and a shadow
// band under its top lip.
func (p painter) menuWell(r image.Rectangle) {
	p.menuSteel(r, toneWell+0.10)
	p.menuEdge(r, 0, ramp(0.02))
	p.bevel(r.Inset(p.bw()), toneBack, false)
	r = r.Intersect(p.img.Rect)
	for y := r.Min.Y + 2*p.bw(); y < min(r.Max.Y, r.Min.Y+int(5*p.k)); y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := p.img.RGBAAt(x, y)
			p.img.SetRGBA(x, y, lerp(c, ramp(0.0), 0.35))
		}
	}
}

// menuBackground is the menu backdrop: darker hull plating with seams and
// rivets at the seams' crossings, dimming towards the corners.
func (p painter) menuBackground(r image.Rectangle) {
	p.menuSteel(r, toneBack-0.10)
	r = r.Intersect(p.img.Rect)
	// Vignette: corners fall off.
	cx, cy := float64(r.Min.X+r.Max.X)/2, float64(r.Min.Y+r.Max.Y)/2
	hw, hh := float64(r.Dx())/2, float64(r.Dy())/2
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			d := math.Hypot((float64(x)-cx)/hw, (float64(y)-cy)/hh) / math.Sqrt2
			if d > 0.55 {
				c := p.img.RGBAAt(x, y)
				p.img.SetRGBA(x, y, lerp(c, steelDark, 0.35*(d-0.55)))
			}
		}
	}
	// Seams every 160 reference pixels, with a lit line under each.
	step := int(160 * p.k)
	seam := func(at int, vertical bool) {
		for i := 0; i <= int(2*p.k)+1; i++ {
			lit := i > int(p.k)
			c := ramp(0.10)
			if lit {
				c = ramp(0.42)
			}
			if vertical {
				for y := r.Min.Y; y < r.Max.Y; y++ {
					if x := at + i; x < r.Max.X {
						p.img.SetRGBA(x, y, c)
					}
				}
			} else {
				for x := r.Min.X; x < r.Max.X; x++ {
					if y := at + i; y < r.Max.Y {
						p.img.SetRGBA(x, y, c)
					}
				}
			}
		}
	}
	var xs, ys []int
	for x := r.Min.X + step; x < r.Max.X-step/3; x += step {
		seam(x, true)
		xs = append(xs, x+int(p.k))
	}
	for y := r.Min.Y + step; y < r.Max.Y-step/3; y += step {
		seam(y, false)
		ys = append(ys, y+int(p.k))
	}
	rad := math.Max(1.5, 1.7*p.k)
	for _, x := range xs {
		for _, y := range ys {
			p.rivet(float64(x), float64(y), rad)
		}
	}
}

// menuNamePlate is the menus' title plate: brown steel with a vertical
// falloff, a dark edge, a brass line and corner rivets.
func (p painter) menuNamePlate(r image.Rectangle) {
	r = r.Intersect(p.img.Rect)
	w := p.bw()
	h := float64(r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		fy := float64(y-r.Min.Y) / h
		for x := r.Min.X; x < r.Max.X; x++ {
			c := lerp(plateBrown, plateEdge, 0.05+0.30*fy+0.15*grain(x/w, y/w))
			if s := surfaces.steel; s != nil {
				dr, dg, db := s.At(x, y, p.k*steelSpan)
				c = texture.Modulate(lerp(plateBrown, plateEdge, 0.05+0.30*fy), 1.5*steelGain, dr, dg, db)
			}
			p.img.SetRGBA(x, y, c)
		}
	}
	edges(r, w, func(x, y int, _ bool) { p.img.SetRGBA(x, y, plateEdge) })
	edges(r.Inset(w), w, func(x, y int, lit bool) {
		if lit {
			p.img.SetRGBA(x, y, brassLight)
		} else {
			p.img.SetRGBA(x, y, brassDark)
		}
	})
	p.menuRivets(r)
}

// menuButton paints a menu button in its state.
func (p painter) menuButton(r image.Rectangle, state MenuButtonState) {
	switch state {
	case MenuPressed:
		p.menuSteel(r, toneWell+0.16)
		p.menuEdge(r, 0, ramp(0.05))
		p.bevel(r.Inset(p.bw()), toneBack, false)
	case MenuDisabled:
		p.menuSteel(r, 0.16)
		p.menuEdge(r, 0, ramp(0.02))
		edges(r.Inset(p.bw()), p.bw(), func(x, y int, lit bool) {
			if lit {
				p.img.SetRGBA(x, y, ramp(0.32))
			} else {
				p.img.SetRGBA(x, y, ramp(0.14))
			}
		})
	case MenuFocus:
		p.menuPlate(r, 0.28)
		p.menuEdge(r, 1, brassDark)
		edges(r.Inset(2*p.bw()), p.bw(), func(x, y int, lit bool) {
			if lit {
				p.img.SetRGBA(x, y, brassLight)
			} else {
				p.img.SetRGBA(x, y, brass)
			}
		})
	default:
		p.menuPlate(r, 0.22)
		edges(r.Inset(2*p.bw()), p.bw(), func(x, y int, lit bool) {
			if lit {
				p.img.SetRGBA(x, y, ramp(0.34))
			}
		})
	}
}

// SteelPanel paints a dark raised steel column (the menu's side panel).
func SteelPanel(dst *ebiten.Image, r image.Rectangle, scale float64) {
	drawMenu(dst, r, menuImage(menuPanel, r, scale, MenuNormal, nil))
}

// SteelBackground paints dst's rect r (device pixels) as hull plating.
func SteelBackground(dst *ebiten.Image, r image.Rectangle, scale float64) {
	drawMenu(dst, r, menuImage(menuBack, r, scale, MenuNormal, nil))
}

// SteelPlate paints a raised steel plate.
func SteelPlate(dst *ebiten.Image, r image.Rectangle, scale float64) {
	drawMenu(dst, r, menuImage(menuPlate, r, scale, MenuNormal, nil))
}

// SteelWell paints a recessed steel well.
func SteelWell(dst *ebiten.Image, r image.Rectangle, scale float64) {
	drawMenu(dst, r, menuImage(menuWell, r, scale, MenuNormal, nil))
}

// SteelNamePlate paints the brown/brass name plate (titles).
func SteelNamePlate(dst *ebiten.Image, r image.Rectangle, scale float64) {
	drawMenu(dst, r, menuImage(menuName, r, scale, MenuNormal, nil))
}

// SteelButton paints a steel button in state.
func SteelButton(dst *ebiten.Image, r image.Rectangle, scale float64, state MenuButtonState) {
	drawMenu(dst, r, menuImage(menuButton, r, scale, state, nil))
}
