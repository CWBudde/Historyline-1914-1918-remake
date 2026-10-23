package dosscreen

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The frame's steel is a painting of one riveted hex plate: every cell that is
// not an opening is drawn as that plate, clipped by the panel, so the teeth
// are the plates of the cells around the view, as in the original. Without a
// painting the frame is the drawn steel sheet.

// plateArt is the plate painting in use; gen counts changes so the Frame
// repaints its panels.
var plateArt struct {
	sync.RWMutex
	tex *plateTex
	gen int
}

// SetPlate makes img the painting of the frame's hex plate, nil for the drawn
// steel. Its opaque box (alpha over half) is taken as the cell's 96×80 box.
func SetPlate(img image.Image) {
	t := newPlateTex(img)
	plateArt.Lock()
	defer plateArt.Unlock()
	plateArt.tex = t
	plateArt.gen++
}

// currentPlate is the plate painting in use and its generation.
func currentPlate() (*plateTex, int) {
	plateArt.RLock()
	defer plateArt.RUnlock()
	return plateArt.tex, plateArt.gen
}

// plateTex is the plate painting cropped to its cell box, premultiplied, at
// the full size and halved down to a few texels.
type plateTex struct {
	mips []plateMip
}

type plateMip struct {
	w, h int
	d    []float32 // premultiplied RGBA, 0..1
}

func newPlateTex(img image.Image) *plateTex {
	if img == nil {
		return nil
	}
	b := img.Bounds()
	box := image.Rectangle{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a > 0x8000 {
				box = box.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if box.Empty() {
		return nil
	}
	m := plateMip{box.Dx(), box.Dy(), make([]float32, 4*box.Dx()*box.Dy())}
	for y := range m.h {
		for x := range m.w {
			r, g, bl, a := img.At(box.Min.X+x, box.Min.Y+y).RGBA()
			i := 4 * (y*m.w + x)
			m.d[i], m.d[i+1], m.d[i+2], m.d[i+3] = float32(r)/0xffff, float32(g)/0xffff, float32(bl)/0xffff, float32(a)/0xffff
		}
	}
	t := &plateTex{mips: []plateMip{m}}
	for m.w > 8 && m.h > 8 {
		h := plateMip{m.w / 2, m.h / 2, make([]float32, 4*(m.w/2)*(m.h/2))}
		for y := range h.h {
			for x := range h.w {
				for ch := range 4 {
					at := func(dx, dy int) float32 { return m.d[4*((2*y+dy)*m.w+2*x+dx)+ch] }
					h.d[4*(y*h.w+x)+ch] = (at(0, 0) + at(1, 0) + at(0, 1) + at(1, 1)) / 4
				}
			}
		}
		t.mips = append(t.mips, h)
		m = h
	}
	return t
}

// sample is the premultiplied colour at (u, v) in 0..1 across the cell box,
// filtered for texels per device pixel of the full size.
func (t *plateTex) sample(u, v, texels float64) [4]float32 {
	m := t.mips[0]
	for i := 1; i < len(t.mips) && texels > 1.5; i++ {
		m, texels = t.mips[i], texels/2
	}
	fx, fy := u*float64(m.w)-0.5, v*float64(m.h)-0.5
	x0, y0 := int(math.Floor(fx)), int(math.Floor(fy))
	ax, ay := float32(fx-float64(x0)), float32(fy-float64(y0))
	at := func(x, y int) []float32 {
		x, y = max(0, min(m.w-1, x)), max(0, min(m.h-1, y))
		i := 4 * (y*m.w + x)
		return m.d[i : i+4]
	}
	var out [4]float32
	p00, p10, p01, p11 := at(x0, y0), at(x0+1, y0), at(x0, y0+1), at(x0+1, y0+1)
	for ch := range 4 {
		top := p00[ch] + ax*(p10[ch]-p00[ch])
		bot := p01[ch] + ax*(p11[ch]-p01[ch])
		out[ch] = top + ay*(bot-top)
	}
	return out
}

// seam is what shows between two plates where their painted edges let it
// through: the darkest steel.
var seam = steelDark

// plateAt is the colour of the plate under world point (wx, wy) at scale
// device pixels per world pixel, over the seam.
func (t *plateTex) plateAt(wx, wy, scale float64) color.RGBA {
	cx, cy := cellCentre(wx, wy)
	u := (wx - cx + hexgrid.HexWidth/2) / hexgrid.HexWidth
	v := (wy - cy + hexgrid.HexHeight/2) / hexgrid.HexHeight
	p := t.sample(u, v, float64(t.mips[0].w)/hexgrid.HexWidth/scale)
	over := func(c float32, s uint8) uint8 {
		return uint8(math.Round(float64(c*255 + float32(s)*(1-p[3]))))
	}
	return color.RGBA{over(p[0], seam.R), over(p[1], seam.G), over(p[2], seam.B), 255}
}

// cellCentre is the centre of the hex polygon holding world point (wx, wy),
// on the map or not.
func cellCentre(wx, wy float64) (float64, float64) {
	const hw, hh = hexgrid.HexWidth / 2, hexgrid.HexHeight / 2
	const run = hexgrid.HexWidth - hexgrid.HalfPitch
	line0 := int(math.Floor((wy - hh) / hexgrid.LinePitch))
	bx, by, bd := 0.0, 0.0, math.Inf(1)
	for line := line0 - 1; line <= line0+2; line++ {
		shift := float64((line & 1) * hexgrid.HalfPitch)
		col0 := int(math.Floor((wx - hw - shift) / hexgrid.ColPitch))
		for col := col0; col <= col0+1; col++ {
			cx, cy := hexgrid.Center(hexgrid.Coord{Line: line, Col: col})
			dx, dy := math.Abs(wx-cx), math.Abs(wy-cy)
			if dy <= hh && dx <= hw-run*dy/hh {
				return cx, cy
			}
			if d := dx*dx + dy*dy; d < bd {
				bx, by, bd = cx, cy, d
			}
		}
	}
	return bx, by
}
