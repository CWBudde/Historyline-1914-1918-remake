package dospanel

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Finale is the closing scene of a battle won by taking the HQ, after the
// original's: a low view up at the headquarters fades in, its doors swing
// open on the defenders surrendering inside, and the winner's flag goes up
// the bare pole on the gable. Flag nil draws a plain bar in FlagColour.
type Finale struct {
	Flag       *ebiten.Image
	FlagColour color.RGBA
}

// FinaleArt is the finale's paintings (art/screens/capture): the open
// headquarters and the leaves of the closed doors cut out of the closed one.
type FinaleArt struct {
	open   *ebiten.Image
	leaves [2]*ebiten.Image   // left and right, clear outside the arch
	at     [2]image.Rectangle // each leaf's place in the paintings' pixels
	arch   archRect           // in the paintings' pixels
}

// archRect is the doorway: its sides, the seam between the leaves, the
// crown and springing of the arch over it and its sill.
type archRect struct{ x0, xm, x1, crown, spring, sill float64 }

// The headquarters' points, as fractions of the paintings' size: the
// doorway's arch and the flagpole from its foot to its head.
var (
	hqArch = archRect{x0: 0.413, x1: 0.586, crown: 0.474, spring: 0.544, sill: 0.832}
	hqPole = [3]float64{0.500, 0.199, 0.050} // x, foot, head
)

// The finale's timing in ticks (60 a second).
const (
	finaleFade    = 40  // the view fades in
	finaleDoors   = 50  // the doors start to open
	finaleOpen    = 140 // the doors are open
	finaleHoist   = 150 // the flag starts up the pole
	finaleTop     = 270 // the flag is at the head
	finaleTotal   = 420 // the end of the hold
	finaleReduced = 90  // the whole scene under reduced motion
	finaleSwing   = 82 * math.Pi / 180
)

// NewFinaleArt cuts the leaves of the closed doors out of closed for the
// finale over open; nil when either is missing or their sizes differ.
func NewFinaleArt(closed, open image.Image) *FinaleArt {
	if closed == nil || open == nil || closed.Bounds().Size() != open.Bounds().Size() || closed.Bounds().Empty() {
		return nil
	}
	leaves, a := cutLeaves(closed)
	art := &FinaleArt{open: ebiten.NewImageFromImage(open), arch: a}
	for i, l := range leaves {
		art.leaves[i] = ebiten.NewImageFromImage(l)
		art.at[i] = l.Rect
	}
	return art
}

// cutLeaves cuts the closed doors' leaves out of closed, each in its own
// rectangle of the painting's pixels and clear outside the arch, and gives
// the arch in those pixels.
func cutLeaves(closed image.Image) (leaves [2]*image.RGBA, a archRect) {
	b := closed.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())
	a = archRect{hqArch.x0 * w, (hqArch.x0 + hqArch.x1) / 2 * w, hqArch.x1 * w, hqArch.crown * h, hqArch.spring * h, hqArch.sill * h}
	for i, xs := range [2][2]float64{{a.x0, a.xm}, {a.xm, a.x1}} {
		r := image.Rect(int(math.Floor(xs[0])), int(math.Floor(a.crown)), int(math.Ceil(xs[1])), int(math.Ceil(a.sill)))
		leaf := image.NewRGBA(r)
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				cov := a.cover(float64(x)+0.5, float64(y)+0.5)
				// Each leaf ends at the seam, half a pixel soft.
				cov *= math.Max(0, math.Min(1, math.Min(float64(x)+1-xs[0], xs[1]-float64(x))))
				if cov <= 0 {
					continue
				}
				cr, cg, cb, _ := closed.At(b.Min.X+x, b.Min.Y+y).RGBA()
				leaf.SetRGBA(x, y, color.RGBA{uint8(float64(cr>>8) * cov), uint8(float64(cg>>8) * cov), uint8(float64(cb>>8) * cov), uint8(255 * cov)})
			}
		}
		leaves[i] = leaf
	}
	return leaves, a
}

// cover is how much of the pixel at (x, y) lies in the doorway, soft over a
// pixel at its edge.
func (a archRect) cover(x, y float64) float64 {
	d := math.Min(math.Min(x-a.x0, a.x1-x), a.sill-y) // inside the sides and the sill
	if y < a.spring {
		rx, ry := (a.x1-a.x0)/2, a.spring-a.crown
		dx, dy := (x-a.xm)/rx, (y-a.spring)/ry
		// The distance to the ellipse, near enough at its edge.
		e := (1 - math.Hypot(dx, dy)) * math.Min(rx, ry)
		d = math.Min(d, e)
	}
	return math.Max(0, math.Min(1, d+0.5))
}

// FinaleState is the finale at a tick: the light (0 black to 1), the doors'
// swing (0 shut to 1 open) and the flag's hoist (0 at the pole's foot to 1
// at its head) and wave phase; Flag reports whether it shows.
type FinaleState struct {
	Light, Swing, Hoist, Wave float64
	Flag                      bool
}

// FinaleTicks is the length of the finale in ticks.
func FinaleTicks(reduced bool) int {
	if reduced {
		return finaleReduced
	}
	return finaleTotal
}

// FinaleAt is the finale at tick: under reduced motion its end.
func FinaleAt(tick int, reduced bool) FinaleState {
	if reduced {
		return FinaleState{Light: 1, Swing: 1, Hoist: 1, Flag: true}
	}
	tick = min(max(tick, 0), finaleTotal)
	smooth := func(t float64) float64 { return t * t * (3 - 2*t) }
	return FinaleState{
		Light: ease(ramp01(tick, 0, finaleFade)),
		Swing: smooth(ramp01(tick, finaleDoors, finaleOpen-finaleDoors)),
		Hoist: smooth(ramp01(tick, finaleHoist, finaleTop-finaleHoist)),
		Wave:  float64(tick) * captureWaveTick,
		Flag:  tick >= finaleHoist,
	}
}

// DrawFinale draws finale f at tick over r: the headquarters covering r,
// kept to its top where r is wider, the doors and the flag. Nothing without
// art.
func DrawFinale(dst *ebiten.Image, r image.Rectangle, f Finale, tick int, reduced bool, art *FinaleArt) {
	if art == nil || r.Empty() {
		return
	}
	c := clip(dst, r)
	c.Fill(color.Black)
	st := FinaleAt(tick, reduced)
	ob := art.open.Bounds()
	s, ox, _ := cover(r, ob)
	oy := float64(r.Min.Y) // a wider window crops the stairs, never the pole
	l := float32(st.Light)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(ox, oy)
	op.ColorScale.Scale(l, l, l, 1)
	c.DrawImage(art.open, op)

	at := func(x, y float64) (float32, float32) { return float32(ox + x*s), float32(oy + y*s) }
	// Each leaf turns on its hinge into the hall: narrower, its free edge
	// shorter and darker as it turns.
	a := art.arch
	turn := st.Swing * finaleSwing
	for i, leaf := range art.leaves {
		lb := art.at[i]
		hinge, free := float64(lb.Min.X), float64(lb.Max.X)
		if i == 1 {
			hinge, free = free, hinge
		}
		fx := hinge + (free-hinge)*math.Cos(turn)
		shrink := 0.12 * math.Sin(turn) * float64(lb.Dy())
		mid := (a.crown + a.sill) / 2
		top, bottom := float64(lb.Min.Y), float64(lb.Max.Y)
		ftop, fbottom := top+shrink*(mid-top)/(bottom-top), bottom-shrink*(bottom-mid)/(bottom-top)
		k := float32(st.Light * (1 - 0.6*math.Sin(turn)))
		quad := [4][4]float64{ // dst x, dst y, src x, src y
			{hinge, top, hinge, top},
			{fx, ftop, free, top},
			{fx, fbottom, free, bottom},
			{hinge, bottom, hinge, bottom},
		}
		var vs [4]ebiten.Vertex
		for j, q := range quad {
			x, y := at(q[0], q[1])
			vs[j] = ebiten.Vertex{DstX: x, DstY: y, SrcX: float32(q[2] - float64(lb.Min.X)), SrcY: float32(q[3] - float64(lb.Min.Y)), ColorR: k, ColorG: k, ColorB: k, ColorA: 1}
		}
		c.DrawTriangles(vs[:], []uint16{0, 1, 2, 0, 2, 3}, leaf, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear})
	}

	if !st.Flag {
		return
	}
	w := float64(ob.Dx()) * s
	px, foot, head := ox+hqPole[0]*float64(ob.Dx())*s, oy+hqPole[1]*float64(ob.Dy())*s, oy+hqPole[2]*float64(ob.Dy())*s
	fw := 0.07 * w
	fh := fw * 2 / 3
	y := foot - fh + (head-(foot-fh))*st.Hoist
	waveFlag(c, f.Flag, f.FlagColour, px+0.002*w, y, fw, fh, st.Wave, [3]float64{0.85 * st.Light, 0.85 * st.Light, 0.85 * st.Light})
}
