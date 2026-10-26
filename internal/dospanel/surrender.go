package dospanel

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// The surrender scene, the closing scene of a battle the other side gave
// up: three paintings of two rival commanders at dawn (art/screens/capture:
// surrender, surrender-offer, surrender-accepted). The stillness fades in
// and holds; the defeated rival then offers his sword flat on open palms,
// and the winner takes it while the white cloth opens behind him. The
// frames are full paintings, so the scene crossfades between them and
// never overlays them.

// SurrenderArt is the surrender scene's paintings, in scene order:
// stillness, the offer, the acceptance.
type SurrenderArt struct {
	frames [3]*ebiten.Image
}

// NewSurrenderArt takes the surrender scene's three paintings; nil when
// one is missing or their sizes differ.
func NewSurrenderArt(frames [3]image.Image) *SurrenderArt {
	art := &SurrenderArt{}
	for i, f := range frames {
		if f == nil || f.Bounds().Empty() || i > 0 && f.Bounds().Size() != frames[0].Bounds().Size() {
			return nil
		}
		art.frames[i] = ebiten.NewImageFromImage(f)
	}
	return art
}

// The surrender scene's timing in ticks (60 a second).
const (
	surrenderFade     = 40  // the stillness fades in
	surrenderOfferAt  = 100 // the offer's crossfade starts
	surrenderBlend    = 60  // a crossfade between two frames
	surrenderAcceptAt = 230 // the acceptance's crossfade starts
	surrenderTotal    = 420 // the end of the hold on the acceptance
	surrenderStill    = 90  // the whole scene under reduced motion
)

// SurrenderState is the surrender scene at a tick: the light (0 black to
// 1), the frame showing and, during a crossfade, the next frame and its
// blend in (0 to 1); Next is -1 while one frame holds.
type SurrenderState struct {
	Light       float64
	Frame, Next int
	Blend       float64
}

// SurrenderTicks is the length of the surrender scene in ticks.
func SurrenderTicks(reduced bool) int {
	if reduced {
		return surrenderStill
	}
	return surrenderTotal
}

// SurrenderAt is the surrender scene at tick: under reduced motion its
// end, the acceptance in full light.
func SurrenderAt(tick int, reduced bool) SurrenderState {
	if reduced {
		return SurrenderState{Light: 1, Frame: 2, Next: -1}
	}
	tick = min(max(tick, 0), surrenderTotal)
	smooth := func(t float64) float64 { return t * t * (3 - 2*t) }
	st := SurrenderState{Light: ease(ramp01(tick, 0, surrenderFade)), Next: -1}
	switch {
	case tick < surrenderOfferAt:
		st.Frame = 0
	case tick < surrenderOfferAt+surrenderBlend:
		st.Frame, st.Next = 0, 1
		st.Blend = smooth(ramp01(tick, surrenderOfferAt, surrenderBlend))
	case tick < surrenderAcceptAt:
		st.Frame = 1
	case tick < surrenderAcceptAt+surrenderBlend:
		st.Frame, st.Next = 1, 2
		st.Blend = smooth(ramp01(tick, surrenderAcceptAt, surrenderBlend))
	default:
		st.Frame = 2
	}
	return st
}

// DrawSurrender draws the surrender scene at tick over r: the frame
// covering r, crossfading to the next where the scene asks. Nothing
// without art.
func DrawSurrender(dst *ebiten.Image, r image.Rectangle, tick int, reduced bool, art *SurrenderArt) {
	if art == nil || r.Empty() {
		return
	}
	c := clip(dst, r)
	c.Fill(color.Black)
	st := SurrenderAt(tick, reduced)
	s, ox, oy := cover(r, art.frames[0].Bounds())
	draw := func(img *ebiten.Image, alpha float64) {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(ox, oy)
		l := float32(st.Light)
		op.ColorScale.Scale(l, l, l, float32(alpha))
		c.DrawImage(img, op)
	}
	draw(art.frames[st.Frame], 1)
	if st.Next >= 0 && st.Blend > 0 {
		draw(art.frames[st.Next], st.Blend)
	}
}
