package dospanel

import (
	"image"
	"image/color"
	"sync"

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
		p.steel(local, toneBack)
	case menuPlate:
		p.plate(local, tonePlate)
	case menuWell:
		p.well(local)
	case menuName:
		p.namePlate(local)
	case menuButton:
		switch state {
		case MenuPressed:
			p.steel(local, toneWell)
			p.bevel(local, toneBack, false)
		case MenuDisabled:
			p.plate(local, toneWell+0.08)
		case MenuFocus:
			p.plate(local, tonePlate)
			edges(local.Inset(p.bw()), p.bw(), func(x, y int, lit bool) {
				if lit {
					img.SetRGBA(x, y, brassLight)
				} else {
					img.SetRGBA(x, y, brass)
				}
			})
		default:
			p.plate(local, tonePlate)
		}
	default:
		paint(p, local)
	}
	_ = paint
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

// SteelBackground paints dst's rect r (device pixels) as flat steel backdrop.
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
