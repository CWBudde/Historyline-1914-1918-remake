package dospanel

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SlotCount is the number of slots a side has in a container.
const SlotCount = 7

// Slot is one slot of a container: empty, or a unit with its name, group
// size of the full group, map sprite and portrait (either may be nil).
type Slot struct {
	Occupied         bool
	Name             string
	Size, Full       int
	Sprite, Portrait *ebiten.Image
}

// Contents is what the contents view shows: the container's name, the
// energy it stores and the total, the viewing side's slots, the slot under
// the cursor (-1 none) and the lines of the details plate.
type Contents struct {
	Name          string
	Stored, Total int
	Slots         [SlotCount]Slot
	Selected      int
	Details       []string
}

// ContentsRects are the parts of the contents view in device pixels inside
// Panel: the steel plate holding everything, the photo's recess and the
// photo, the container's plate, the unit's name plate, the details plate
// and the bounding boxes of the hex slots. Scale is the device pixels per
// reference pixel the layout uses.
type ContentsRects struct {
	Scale                                        float64
	Panel                                        image.Rectangle
	Plate, PhotoWell, Photo, Info, Name, Details image.Rectangle
	Slots                                        [SlotCount]image.Rectangle
}

// The contents in reference pixels: a 134×154 block. The slots are hexes of
// the map's cells (24×24, columns 16 apart) in two staggered rows near the
// bottom: even slots in the lower row, odd slots in the upper.
const (
	contentsW, contentsH = 134, 154
	slotX0, slotPitch    = 19, 16   // centre of slot 0 and the column pitch
	slotLowY, slotHighY  = 140, 128 // centres of the two rows
	slotHalf             = 12
)

// ContentsLayout lays out the contents view in panel at scale.
func ContentsLayout(panel image.Rectangle, scale float64) ContentsRects {
	g := fit(panel, scale, contentsW, contentsH)
	l := ContentsRects{
		Scale:     g.k,
		Panel:     panel,
		Plate:     g.rect(0, 0, contentsW, contentsH),
		PhotoWell: g.rect(2, 2, 58, 50),
		Photo:     g.rect(5, 5, 55, 47),
		Info:      g.rect(61, 4, 132, 40),
		Name:      g.rect(2, 52, 132, 63),
		Details:   g.rect(2, 65, 132, 113),
	}
	for i := range l.Slots {
		x, y := slotCentre(i)
		l.Slots[i] = g.rect(x-slotHalf, y-slotHalf, x+slotHalf, y+slotHalf)
	}
	return l
}

// slotCentre is the centre of slot i in reference pixels.
func slotCentre(i int) (x, y float64) {
	y = slotLowY
	if i%2 == 1 {
		y = slotHighY
	}
	return slotX0 + slotPitch*float64(i), y
}

// SlotAt is the slot whose hex holds device pixel (x, y).
func SlotAt(l ContentsRects, x, y int) (int, bool) {
	best, bestN := -1, math.Inf(1)
	for i, r := range l.Slots {
		cx, cy, a, b := hexOf(r)
		if n := hexNorm(float64(x)+0.5-cx, float64(y)+0.5-cy, a, b); n <= 1 && n < bestN {
			best, bestN = i, n
		}
	}
	return best, best >= 0
}

// SlotNeighbor is the slot next to slot i in direction d: left and right
// along the zigzag, up from a lower slot to the upper slot on its right and
// down back again.
func SlotNeighbor(i int, d Dir) (int, bool) {
	j := -1
	switch d {
	case Left:
		j = i - 1
	case Right:
		j = i + 1
	case Up:
		if i%2 == 0 {
			j = i + 1
		}
	case Down:
		if i%2 == 1 {
			j = i - 1
		}
	}
	if i < 0 || i >= SlotCount || j < 0 || j >= SlotCount {
		return 0, false
	}
	return j, true
}

// paintContents paints the plates and the empty slots of the contents view.
func paintContents(l ContentsRects) *image.RGBA {
	img := image.NewRGBA(l.Panel)
	p := painter{img, l.Scale}
	p.steel(img.Rect, toneBack)
	p.plate(l.Plate, tonePlate)
	p.well(l.PhotoWell)
	p.plate(l.Photo, toneBlank)
	p.namePlate(l.Info)
	p.namePlate(l.Name)
	p.namePlate(l.Details)
	for _, r := range l.Slots {
		cx, cy, a, b := hexOf(r)
		p.hex(cx, cy, a, b, toneWell+0.12, false)
	}
	return img
}

// Labels of the contents view.
const (
	labelStored = "RESOURCES"
	labelTotal  = "TOTAL"
)

// DrawContents draws the contents view of c over panel at scale: the photo
// of the unit under the cursor, the container's plate, the unit's name
// plate, the details and the slots with the cursor ring on the selected
// one.
func DrawContents(dst *ebiten.Image, panel image.Rectangle, scale float64, c Contents, t Text) {
	size := panel.Size()
	l0 := ContentsLayout(image.Rectangle{Max: size}, scale)
	drawPlate(dst, panel, cachedPlate(contentsPlate, size, scale, func() *image.RGBA { return paintContents(l0) }))

	l := ContentsLayout(panel, scale)
	for i, s := range c.Slots {
		if s.Occupied {
			drawSprite(dst, l.Slots[i].Intersect(panel), s.Sprite, 0.8)
		}
	}
	g := fit(panel, scale, contentsW, contentsH)
	w := writer{clip(dst, panel), t}
	w.left(c.Name, g.x(67), g.y(11), cream)
	w.left(labelStored, g.x(67), g.y(25), cream)
	w.right(strconv.Itoa(c.Stored), g.x(127), g.y(25), amber)
	w.left(labelTotal, g.x(67), g.y(31), cream)
	w.right(strconv.Itoa(c.Total), g.x(127), g.y(31), amber)
	for i, s := range c.Details {
		y := 71 + 7*float64(i)
		if y > 108 {
			break
		}
		w.left(s, g.x(8), g.y(y), cream)
	}

	if c.Selected < 0 || c.Selected >= SlotCount {
		return
	}
	drawRing(clip(dst, panel), l.Slots[c.Selected], l.Scale)
	if s := c.Slots[c.Selected]; s.Occupied {
		drawPhoto(dst, l.Photo.Intersect(panel), s.Portrait)
		if s.Size > 0 {
			w.left(strconv.Itoa(s.Size), g.x(8), g.y(57.5), amber)
		}
		w.centre(s.Name, g.x(67), g.y(57.5), cream)
	}
}

// drawRing draws the cursor ring around the hex in r: light steel with a
// dark outer line.
func drawRing(dst *ebiten.Image, r image.Rectangle, k float64) {
	cx, cy, a, b := hexOf(r)
	path := func(grow float64) *vector.Path {
		a, b := a+grow, b+grow
		var p vector.Path
		p.MoveTo(float32(cx-a), float32(cy))
		p.LineTo(float32(cx-a/3), float32(cy-b))
		p.LineTo(float32(cx+a/3), float32(cy-b))
		p.LineTo(float32(cx+a), float32(cy))
		p.LineTo(float32(cx+a/3), float32(cy+b))
		p.LineTo(float32(cx-a/3), float32(cy+b))
		p.Close()
		return &p
	}
	stroke := func(p *vector.Path, w float64, c color.RGBA) {
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(c)
		vector.StrokePath(dst, p, &vector.StrokeOptions{Width: float32(w), LineJoin: vector.LineJoinMiter}, op)
	}
	stroke(path(-0.5*k), 3*k, plateEdge)
	stroke(path(-0.5*k), 1.6*k, ringLight)
}
