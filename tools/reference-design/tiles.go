package referencedesign

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

var hexPolygon = [6][2]float64{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}

// registration maps game world pixels onto the reference image:
// source = source_anchor + (world - world_anchor) * num/den, per axis.
type registration struct {
	WorldAnchor  [2]float64 `json:"world_anchor"`
	SourceAnchor [2]float64 `json:"source_anchor"`
	ScaleX       [2]float64 `json:"scale_x"`
	ScaleY       [2]float64 `json:"scale_y"`
}

func (r registration) sourcePoint(wx, wy float64) (float64, float64) {
	return r.SourceAnchor[0] + (wx-r.WorldAnchor[0])*r.ScaleX[0]/r.ScaleX[1],
		r.SourceAnchor[1] + (wy-r.WorldAnchor[1])*r.ScaleY[0]/r.ScaleY[1]
}

// prepareSource removes only the exterior black canvas, never dark scenery.
func prepareSource(src image.Image) *image.NRGBA {
	im := image.NewNRGBA(src.Bounds())
	draw.Draw(im, im.Bounds(), src, src.Bounds().Min, draw.Src)
	var queue []image.Point
	add := func(x, y int) {
		if !image.Pt(x, y).In(im.Bounds()) {
			return
		}
		c := im.NRGBAAt(x, y)
		if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 255 {
			im.SetNRGBA(x, y, color.NRGBA{})
			queue = append(queue, image.Pt(x, y))
		}
	}
	b := im.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		add(x, b.Min.Y)
		add(x, b.Max.Y-1)
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		add(b.Min.X, y)
		add(b.Max.X-1, y)
	}
	for i := 0; i < len(queue); i++ {
		p := queue[i]
		add(p.X-1, p.Y)
		add(p.X+1, p.Y)
		add(p.X, p.Y-1)
		add(p.X, p.Y+1)
	}
	return im
}

// tileOrigin is the top-left world pixel of the hex at zero-based row and column.
func tileOrigin(row, col int) (int, int) {
	x, y := hexgrid.Center(hexgrid.Coord{Line: row, Col: col})
	return int(x) - hexgrid.HexWidth/2, int(y) - hexgrid.HexHeight/2
}

// insideTile uses the same complementary edge ownership as the terrain exports,
// so a shared edge belongs to exactly one of two adjoining hexes.
func insideTile(x, y, density int) bool {
	px, py := (float64(x)+.5)/float64(density), (float64(y)+.5)/float64(density)
	for i, a := range hexPolygon {
		b := hexPolygon[(i+1)%len(hexPolygon)]
		dx, dy := b[0]-a[0], b[1]-a[1]
		cross := dx*(py-a[1]) - dy*(px-a[0])
		if cross < 0 || (cross == 0 && !(dy < 0 || (dy == 0 && dx > 0))) {
			return false
		}
	}
	return true
}

// sampleSource reads the source at a world point with premultiplied bilinear filtering.
func sampleSource(src image.Image, reg registration, wx, wy float64) color.NRGBA {
	sx, sy := reg.sourcePoint(wx, wy)
	b := src.Bounds()
	if sx < float64(b.Min.X) || sy < float64(b.Min.Y) || sx >= float64(b.Max.X) || sy >= float64(b.Max.Y) {
		return color.NRGBA{}
	}
	sx, sy = sx-.5, sy-.5
	x, y := int(math.Floor(sx)), int(math.Floor(sy))
	tx, ty := sx-float64(x), sy-float64(y)
	var channels [4]float64
	for j := range 2 {
		for i := range 2 {
			weight := math.Abs(float64(1-i)-tx) * math.Abs(float64(1-j)-ty)
			r, g, bb, a := src.At(max(b.Min.X, min(b.Max.X-1, x+i)), max(b.Min.Y, min(b.Max.Y-1, y+j))).RGBA()
			channels[0] += float64(r) / 257 * weight
			channels[1] += float64(g) / 257 * weight
			channels[2] += float64(bb) / 257 * weight
			channels[3] += float64(a) / 257 * weight
		}
	}
	a := uint8(math.Floor(channels[3] + .5))
	if a == 0 {
		return color.NRGBA{}
	}
	channel := func(v float64) uint8 {
		return uint8(math.Min(255, math.Max(0, math.Floor(v*255/channels[3]+.5))))
	}
	return color.NRGBA{R: channel(channels[0]), G: channel(channels[1]), B: channel(channels[2]), A: a}
}

// renderTile cuts one hex at the given density; pixels outside the hex are zero RGBA.
func renderTile(src image.Image, reg registration, ox, oy, density int) (*image.NRGBA, int) {
	im := image.NewNRGBA(image.Rect(0, 0, hexgrid.HexWidth*density, hexgrid.HexHeight*density))
	missing := 0
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			if !insideTile(x, y, density) {
				continue
			}
			c := sampleSource(src, reg, float64(ox)+(float64(x)+.5)/float64(density), float64(oy)+(float64(y)+.5)/float64(density))
			if c.A < 255 {
				missing++
			}
			im.SetNRGBA(x, y, c)
		}
	}
	return im, missing
}
