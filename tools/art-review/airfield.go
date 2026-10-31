package artreview

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// An airfield (art/terrain/airfield) is a runway strip and a hangar. The
// runway is painted once as an opaque horizontal strip and turned to the
// north-south axis (ports 3–6), periodic over the 80 world px between two hex
// centres. The apron in front of a
// hangar's doors is cut from the same strip, its tracks running east to the
// runway; it lies on the hangar hex and on the corners of the runway hexes at
// the hangar's NE and SE. The hangar stands on the centre of its hex, doors
// east, with a cast shadow made here.
const (
	runwayWidth   = 48.  // world px across the runway, its boundary stones included
	runwayFeather = 3.   // world px over which a runway or apron edge fades out
	apronWidth    = 26.  // world px across the apron
	hangarLength  = 56.  // world px from the hangar's west gable to its doors
	shadowOffset  = 2.5  // world px toward the lower right
	shadowAlpha   = 0.45 // opacity of the hangar's shadow
)

var hangarShadowColor = color.NRGBA{18, 26, 16, 255}

// apronPieces are the hexes the apron lies on, as the offset of their centre
// from the hangar's: the hangar hex and the runway hexes at its NE and SE.
var apronPieces = map[string]connectionPoint{"hangar": {0, 0}, "apron-sw": {71, -40}, "apron-nw": {71, 40}}

// newRunwayStrip is the opaque strip painting as a quayStrip: the runway's
// centreline on row centre, k source px per world px. A k that puts a whole
// number of the painted boundary stones' spacings into 80 world px lets the
// periodic crossfade meet stone on stone.
func newRunwayStrip(im *image.NRGBA, centre, k float64) quayStrip {
	b := im.Bounds()
	return quayStrip{im: im, centre: centre, k: k, start: float64(b.Min.X) + 4, end: float64(b.Max.X) - 4}
}

// edgeAlpha is the coverage at distance across from a centreline of a band w
// world px wide.
func edgeAlpha(across, w float64) float64 {
	return 1 - connectionSmoothStep((math.Abs(across)-w/2+runwayFeather)/(2*runwayFeather))
}

// hexTile renders a 2x and a 4x hex tile, supersampled to the source's
// resolution k (source px per world px), from px: the premultiplied colour at
// world point q relative to the hex centre. Pixels outside the hex stay
// transparent.
func hexTile(k float64, px func(q connectionPoint) [4]float64) map[int]*image.NRGBA {
	out := map[int]*image.NRGBA{}
	for _, d := range []int{2, 4} {
		sub := max(1, int(math.Ceil(k/float64(d))))
		im := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		for y := range 80 * d {
			for x := range 96 * d {
				if !connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}) {
					continue
				}
				var acc [4]float64
				for sy := range sub {
					for sx := range sub {
						c := px(connectionPoint{(float64(x) + (float64(sx)+.5)/float64(sub)) / float64(d), (float64(y) + (float64(sy)+.5)/float64(sub)) / float64(d)}.sub(connectionPoint{48, 40}))
						for i := range acc {
							acc[i] += c[i]
						}
					}
				}
				im.SetNRGBA(x, y, straight(acc, float64(sub*sub)))
			}
		}
		out[d] = im
	}
	return out
}

// straight turns the sum of n premultiplied samples into a straight colour.
func straight(acc [4]float64, n float64) color.NRGBA {
	a := acc[3] / n
	if a <= 0 {
		return color.NRGBA{}
	}
	c := func(v float64) uint8 { return uint8(math.Round(math.Max(0, math.Min(255, v/n/a)))) }
	return color.NRGBA{c(acc[0]), c(acc[1]), c(acc[2]), uint8(math.Round(a * 255))}
}

// scaled is a premultiplied sample s with its alpha multiplied by a.
func scaled(s [4]float64, a float64) [4]float64 {
	for i := range s {
		s[i] *= a
	}
	return s
}

// runwayTiles renders tile/runway36: the runway band down the hex, periodic
// over the 80 world px between hex centres.
func runwayTiles(s quayStrip) (map[int]*image.NRGBA, error) {
	if s.end-s.start < 1.3*80*s.k {
		return nil, fmt.Errorf("runway strip too short")
	}
	return hexTile(s.k, func(q connectionPoint) [4]float64 {
		// Along the runway is the strip's x, west of it the strip's top.
		return scaled(s.periodic(q[1]*s.k, 80*s.k, s.centre+q[0]*s.k), edgeAlpha(q[0], runwayWidth))
	}), nil
}

// apronTiles renders the apron piece of the hex at offset from the hangar hex:
// a band apronWidth wide on the hangar's centreline, from its doors to the
// runway's west edge, sampled across the middle of the runway strip.
func apronTiles(s quayStrip, offset connectionPoint) map[int]*image.NRGBA {
	from, to := hangarLength/2-2, 71-runwayWidth/2+runwayFeather
	return hexTile(s.k, func(q connectionPoint) [4]float64 {
		h := q.add(offset) // relative to the hangar's centre
		a := edgeAlpha(h[1], apronWidth) * (1 - connectionSmoothStep((from-h[0])/runwayFeather)) * (1 - connectionSmoothStep((h[0]-to)/runwayFeather))
		if a <= 0 {
			return [4]float64{}
		}
		return scaled(s.sample(s.start+(h[0]+71)*s.k, s.centre+h[1]*s.k), a)
	})
}

// hangarSprite renders standing/hangar and its shadow from the cut-out
// painting im (doors facing east): scaled so its opaque box is hangarLength
// world px long, centred on the hex centre, at 2x and 4x on a hex-sized canvas.
func hangarSprite(im *image.NRGBA) (sprite, shadow map[int]*image.NRGBA, err error) {
	box := opaqueBox(im)
	if box.Dx() < 64 || box.Dy() < 32 {
		return nil, nil, fmt.Errorf("hangar: opaque box %v", box)
	}
	k := float64(box.Dx()) / hangarLength
	if float64(box.Dy())/k > 76 {
		return nil, nil, fmt.Errorf("hangar: %v is too deep for a hex", box)
	}
	centre := connectionPoint{float64(box.Min.X+box.Max.X) / 2, float64(box.Min.Y+box.Max.Y) / 2}
	s := quayStrip{im: im, k: k}
	at := func(q connectionPoint) [4]float64 {
		p := centre.add(q.mul(k))
		return s.sample(p[0], p[1])
	}
	sprite, shadow = map[int]*image.NRGBA{}, map[int]*image.NRGBA{}
	for _, d := range []int{2, 4} {
		sub := max(1, int(math.Ceil(k/float64(d))))
		img, sh := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d)), image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		for y := range 80 * d {
			for x := range 96 * d {
				var acc, cast [4]float64
				for sy := range sub {
					for sx := range sub {
						q := connectionPoint{(float64(x) + (float64(sx)+.5)/float64(sub)) / float64(d), (float64(y) + (float64(sy)+.5)/float64(sub)) / float64(d)}.sub(connectionPoint{48, 40})
						c := at(q)
						for i := range acc {
							acc[i] += c[i]
						}
						cast[3] += at(q.sub(connectionPoint{shadowOffset, shadowOffset}))[3]
					}
				}
				n := float64(sub * sub)
				img.SetNRGBA(x, y, straight(acc, n))
				if a := cast[3] / n * shadowAlpha; a > 0 {
					c := hangarShadowColor
					c.A = uint8(math.Round(a * 255))
					sh.SetNRGBA(x, y, c)
				}
			}
		}
		sprite[d], shadow[d] = img, sh
	}
	return sprite, shadow, nil
}

// opaqueBox is the bounding box of the pixels of im with alpha above 128.
func opaqueBox(im *image.NRGBA) image.Rectangle {
	var box image.Rectangle
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A > 128 {
				box = box.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return box
}
