package dospanel

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Capture is the scene of a building taken, after the original's: a dark
// map room lights up, a hand reaches in over a model of the building on the
// staff map and plants a pin with the capturer's flag on its roof, then
// withdraws while the flag stirs. Flag nil draws a plain bar in FlagColour.
type Capture struct {
	Factory    bool
	Flag       *ebiten.Image
	FlagColour color.RGBA
}

// CaptureArt is the scene's paintings: the map table with the depot's and
// the factory's model, and the hand (art/screens/capture). Without a table
// the scene is not drawn; without the hand the pin comes down alone.
type CaptureArt struct{ Depot, Factory, Hand *ebiten.Image }

// The paintings' points, as fractions of their size: where the pin goes into
// each model's roof ridge and where the hand's thumb and index finger pinch.
var (
	depotRidge   = [2]float64{0.624, 0.303}
	factoryRidge = [2]float64{0.526, 0.300}
	handPinch    = [2]float64{0.184, 0.218}
)

// The scene's timing in ticks (60 a second).
const (
	captureLight    = 45  // the room lights up
	captureReach    = 20  // the hand starts to reach in
	captureHover    = 85  // the pin hovers above the ridge
	capturePlant    = 105 // the pin is in
	captureLetGo    = 120 // the hand starts to withdraw
	captureGone     = 165 // the hand is out of the panel
	captureTotal    = 230 // the end of the hold
	captureReduced  = 60  // the whole scene under reduced motion
	captureWaveTick = 0.11
)

// CaptureState is a capture scene at a tick: the room's light (0 dark to
// 1), the pin's lift above the ridge and its offset towards the panel's
// lower right while it is carried (fractions of the panel's height and
// width), the hand's offset from the pin (0 holding it), whether the hand
// shows and the flag's wave phase.
type CaptureState struct {
	Light       float64
	Lift, Carry float64
	HandAway    float64
	Hand        bool
	Planted     bool
	Wave        float64
}

// CaptureTicks is the length of the capture scene in ticks.
func CaptureTicks(reduced bool) int {
	if reduced {
		return captureReduced
	}
	return captureTotal
}

// CapturePlanted is the tick the pin goes in, for the map to follow.
func CapturePlanted(reduced bool) int {
	if reduced {
		return 0
	}
	return capturePlant
}

// CaptureAt is the capture scene at tick: under reduced motion its end.
func CaptureAt(tick int, reduced bool) CaptureState {
	if reduced {
		return CaptureState{Light: 1, Planted: true, HandAway: 1}
	}
	tick = min(max(tick, 0), captureTotal)
	st := CaptureState{Light: ease(ramp01(tick, 0, captureLight)), Wave: float64(tick) * captureWaveTick}
	switch {
	case tick < captureHover:
		t := ease(ramp01(tick, captureReach, captureHover-captureReach))
		st.Carry, st.Lift = 1-t, 0.08+0.5*(1-t)
	case tick < capturePlant:
		t := ramp01(tick, captureHover, capturePlant-captureHover)
		st.Lift = 0.08 * (1 - t*t)
	default:
		st.Planted = true
	}
	st.HandAway = ramp01(tick, captureLetGo, captureGone-captureLetGo)
	st.HandAway *= st.HandAway
	st.Hand = tick >= captureReach && st.HandAway < 1
	return st
}

// DrawCapture draws capture c's scene at tick over panel: the map table
// covering the panel, the pin with the flag and the hand. Nothing without
// the table's painting.
func DrawCapture(dst *ebiten.Image, panel image.Rectangle, c Capture, tick int, reduced bool, art CaptureArt) {
	table, ridge := art.Depot, depotRidge
	if c.Factory {
		table, ridge = art.Factory, factoryRidge
	}
	if panel.Empty() || table == nil || table.Bounds().Empty() {
		return
	}
	st := CaptureAt(tick, reduced)
	dc := clip(dst, panel)
	vector.FillRect(dc, float32(panel.Min.X), float32(panel.Min.Y), float32(panel.Dx()), float32(panel.Dy()), color.Black, false)

	s, ox, oy := cover(panel, table.Bounds())
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(ox, oy)
	op.ColorScale.Scale(float32(st.Light), float32(st.Light), float32(st.Light), 1)
	dc.DrawImage(table, op)

	w, h := float64(panel.Dx()), float64(panel.Dy())
	tb := table.Bounds()
	// The pin's foot: on the ridge, or carried from the panel's lower right.
	fx := ox + ridge[0]*float64(tb.Dx())*s + st.Carry*0.75*w
	fy := oy + ridge[1]*float64(tb.Dy())*s - st.Lift*h + st.Carry*0.35*h
	pinLen := 0.15 * h
	light := 0.35 + 0.65*st.Light
	drawPin(dc, fx, fy, pinLen, h, c, st, light)

	if hand := art.Hand; hand != nil && st.Hand && !hand.Bounds().Empty() {
		hb := hand.Bounds()
		hs := 0.8 * w / float64(hb.Dx())
		// The fingers pinch the pin near its point, well below the flag, and
		// leave towards the lower right.
		gx := fx + st.HandAway*0.8*w
		gy := fy - 0.08*pinLen + st.HandAway*0.7*h
		hop := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		hop.GeoM.Translate(-handPinch[0]*float64(hb.Dx()), -handPinch[1]*float64(hb.Dy()))
		hop.GeoM.Scale(hs, hs)
		hop.GeoM.Translate(gx, gy)
		m := warm(light)
		colorm.DrawImage(dc, hand, m, &colorm.DrawImageOptions{Filter: ebiten.FilterLinear, GeoM: hop.GeoM})
	}
}

// cover is the scale and the top left that make an image of bounds b cover
// r, centred.
func cover(r, b image.Rectangle) (s, ox, oy float64) {
	s = math.Max(float64(r.Dx())/float64(b.Dx()), float64(r.Dy())/float64(b.Dy()))
	ox = float64(r.Min.X) + (float64(r.Dx())-float64(b.Dx())*s)/2 - float64(b.Min.X)*s
	oy = float64(r.Min.Y) + (float64(r.Dy())-float64(b.Dy())*s)/2 - float64(b.Min.Y)*s
	return s, ox, oy
}

// drawPin draws the pin standing on (x, y), n long in a panel h high, with
// the flag at its head, lit by light.
func drawPin(c *ebiten.Image, x, y, n, h float64, cp Capture, st CaptureState, light float64) {
	lw := math.Max(1, 0.005*h)
	shade := func(v uint8) uint8 { return uint8(float64(v) * light) }
	vector.StrokeLine(c, float32(x), float32(y), float32(x), float32(y-n), float32(lw),
		color.RGBA{shade(0x9a), shade(0x86), shade(0x64), 0xff}, true)
	vector.FillCircle(c, float32(x), float32(y-n), float32(1.5*lw), color.RGBA{shade(0xe0), shade(0xbc), shade(0x6c), 0xff}, true)
	fw := 0.1 * h
	waveFlag(c, cp.Flag, cp.FlagColour, x+lw/2, y-n+lw, fw, fw*2/3, st.Wave, lampTint(light))
}

// lampTint is the oil lamp's warm light at strength light.
func lampTint(light float64) [3]float64 { return [3]float64{light, 0.88 * light, 0.7 * light} }

// warm is the map room's tone for the hand: its lightness in the painting's
// warm browns, at strength light.
func warm(light float64) colorm.ColorM {
	var m colorm.ColorM
	for i, k := range lampTint(light) {
		for j, v := range [3]float64{0.299, 0.587, 0.114} {
			m.SetElement(i, j, 0.8*k*v*[3]float64{1, 0.8, 0.56}[i])
		}
	}
	return m
}

// whiteTex is a white pixel, the source of the plain flag's strips, made on
// first use.
var whiteTex struct {
	once sync.Once
	img  *ebiten.Image
}

func whitePixel() *ebiten.Image {
	whiteTex.once.Do(func() {
		whiteTex.img = ebiten.NewImage(1, 1)
		whiteTex.img.Fill(color.White)
	})
	return whiteTex.img
}

// waveFlag draws flag (nil: a bar of colour, none when clear) hanging from
// its staff edge at (x, y), w wide and h high, stirred at phase and lit by
// tint: a band of strips whose edges rise and fall more towards the fly,
// shaded by the cloth's slope.
func waveFlag(c *ebiten.Image, flag *ebiten.Image, colour color.RGBA, x, y, w, h, phase float64, tint [3]float64) {
	src, k := flag, [3]float64{1, 1, 1}
	if src == nil || src.Bounds().Empty() {
		if colour.A == 0 {
			return
		}
		src, k = whitePixel(), [3]float64{float64(colour.R) / 255, float64(colour.G) / 255, float64(colour.B) / 255}
	}
	const strips = 24
	amp := 0.07 * h
	sb := src.Bounds()
	vs := make([]ebiten.Vertex, 0, 2*(strips+1))
	is := make([]uint16, 0, 6*strips)
	for i := range strips + 1 {
		t := float64(i) / strips
		dy := amp * t * math.Sin(phase-5*t)
		shade := 0.88 + 0.12*t*math.Cos(phase-5*t)
		sx := float32(sb.Min.X) + float32(t)*float32(sb.Dx())
		for j, yy := range [2]float64{y + dy, y + dy + h} {
			v := ebiten.Vertex{DstX: float32(x + t*w), DstY: float32(yy), SrcX: sx, SrcY: float32(sb.Min.Y + j*sb.Dy()), ColorA: 1}
			v.ColorR, v.ColorG, v.ColorB = float32(k[0]*tint[0]*shade), float32(k[1]*tint[1]*shade), float32(k[2]*tint[2]*shade)
			vs = append(vs, v)
		}
		if i > 0 {
			n := uint16(2 * i)
			is = append(is, n-2, n, n+1, n-2, n+1, n-1)
		}
	}
	c.DrawTriangles(vs, is, src, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear})
}
