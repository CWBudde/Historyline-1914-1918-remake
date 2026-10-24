package main

import (
	"fmt"
	"image"
	"math"
)

// A harbour (art/terrain/harbour) is three hexes: the quayside warehouse
// (harbourN), the customs house at its SW (harbourW) and the quay yard at its
// S (harbourS). One opaque painting of granite setts paves all three, sampled
// in the group's frame so the paving runs on over the edges the pieces share;
// it fades out with a ragged edge along the outer edges, and at the yard's S
// edge it leaves the road coming in uncovered for a stretch. A kit painting
// holds the warehouse, the yard's cargo and crane and the customs house in
// three columns; each stands on its hex with a cast shadow made here.
const (
	harbourPavingK  = 6.  // source px per world px of the paving painting
	harbourInset    = 2.  // world px of an outer edge left unpaved at least
	harbourFray     = 4.  // world px by which the unpaved margin wanders beyond that
	harbourFeather  = 5.  // world px over which the paving fades out
	harbourRound    = 4.  // world px: rounding of the corners between outer edges
	yardRoadHalf    = 11. // world px from the road's axis to where the setts cover it fully
	yardRoadDepth   = 8.  // world px the road runs into the yard uncovered
	yardRoadBlend   = 20. // world px the road's fade into the setts adds to harbourFeather
	harbourMinPiece = 64  // px: smaller opaque shapes of the kit are haze
)

// harbourPavingPieces are the pieces' centres relative to the warehouse's and
// the edges each shares with another piece (hexgrid.Around directions).
var harbourPavingPieces = map[string]struct {
	at     connectionPoint
	shared []int
}{
	"harbourN": {connectionPoint{0, 0}, []int{2, 3}},
	"harbourW": {connectionPoint{-71, 40}, []int{0, 1}},
	"harbourS": {connectionPoint{0, 80}, []int{4, 5}},
}

// harbourCentre is the middle of the group's extent in its frame, sampled at
// the middle of the paving painting.
var harbourCentre = connectionPoint{-35.5, 40}

// hexEdgeDistance is the distance of q (relative to the hex centre) inside the
// hex edge facing hexgrid.Around direction dir.
func hexEdgeDistance(q connectionPoint, dir int) float64 {
	poly := sharedConnectionGeometry().Polygon
	// The polygon runs N, NE, SE, S, SW, NW from its top-left corner.
	i := (dir + 1) % 6
	a, b := poly[i].sub(connectionPoint{48, 40}), poly[(i+1)%6].sub(connectionPoint{48, 40})
	e := b.sub(a)
	return (e[0]*(q[1]-a[1]) - e[1]*(q[0]-a[0])) / e.length()
}

// valueNoise is smooth noise in [-1, 1] over cells of cell world px.
func valueNoise(p connectionPoint, cell float64) float64 {
	x, y := p[0]/cell, p[1]/cell
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := connectionSmoothStep(x-x0), connectionSmoothStep(y-y0)
	h := func(i, j float64) float64 {
		v := uint64(int64(i)*73856093) ^ uint64(int64(j)*19349663)
		v ^= v >> 13
		v *= 0x5bd1e995
		v ^= v >> 15
		return float64(v%2001)/1000 - 1
	}
	top := h(x0, y0)*(1-fx) + h(x0+1, y0)*fx
	bottom := h(x0, y0+1)*(1-fx) + h(x0+1, y0+1)*fx
	return top*(1-fy) + bottom*fy
}

// harbourPavingTiles renders tile/<piece> from the paving painting: the setts
// over the hex, faded out along its outer edges with rounded corners and a
// wandering margin.
func harbourPavingTiles(paving *image.NRGBA, piece string) map[int]*image.NRGBA {
	pp := harbourPavingPieces[piece]
	b := paving.Bounds()
	s := quayStrip{im: paving}
	mid := connectionPoint{float64(b.Min.X+b.Max.X) / 2, float64(b.Min.Y+b.Max.Y) / 2}
	return hexTile(harbourPavingK, func(q connectionPoint) [4]float64 {
		g := q.add(pp.at) // in the group's frame
		// The smooth minimum of the distances inside the outer edges rounds
		// the corners between them.
		var sum, road float64
		for dir := range 6 {
			if dir == pp.shared[0] || dir == pp.shared[1] {
				continue
			}
			d := hexEdgeDistance(q, dir)
			if piece == "harbourS" && dir == 2 {
				// Over the road the S edge lies deeper and fades out slower.
				road = 1 - connectionSmoothStep((math.Abs(q[0])-yardRoadHalf)/4)
				d = (d - yardRoadDepth*road) * harbourFeather / (harbourFeather + yardRoadBlend*road)
			}
			sum += math.Exp(-d / harbourRound)
		}
		inside := -harbourRound * math.Log(sum)
		// The margin wanders, never onto the hex edge, and keeps still over the road.
		margin := harbourInset + harbourFray*(valueNoise(g, 10)+1)/2*(1-road)
		a := connectionSmoothStep((inside - margin) / harbourFeather)
		if a <= 0 {
			return [4]float64{}
		}
		p := mid.add(g.sub(harbourCentre).mul(harbourPavingK))
		return scaled(s.sample(p[0], p[1]), a)
	})
}

// harbourKitPieces splits the kit painting into its three columns' pieces:
// each opaque shape (alpha 32 and more, 8-connected) goes to the column its
// centre lies in, so a part reaching over a column edge stays with its piece;
// shapes under harbourMinPiece px are left out.
func harbourKitPieces(sheet *image.NRGBA) [3]*image.NRGBA {
	b := sheet.Bounds()
	var out [3]*image.NRGBA
	for i := range out {
		out[i] = image.NewNRGBA(b)
	}
	seen := make([]bool, b.Dx()*b.Dy())
	at := func(x, y int) int { return (y-b.Min.Y)*b.Dx() + x - b.Min.X }
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if seen[at(x, y)] || sheet.NRGBAAt(x, y).A < 32 {
				continue
			}
			seen[at(x, y)] = true
			shape := []image.Point{{x, y}}
			var sx float64
			for i := 0; i < len(shape); i++ {
				p := shape[i]
				sx += float64(p.X)
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						n := image.Pt(p.X+dx, p.Y+dy)
						if n.In(b) && !seen[at(n.X, n.Y)] && sheet.NRGBAAt(n.X, n.Y).A >= 32 {
							seen[at(n.X, n.Y)] = true
							shape = append(shape, n)
						}
					}
				}
			}
			if len(shape) < harbourMinPiece {
				continue
			}
			col := min(2, max(0, int(sx/float64(len(shape))-float64(b.Min.X))*3/b.Dx()))
			for _, p := range shape {
				out[col].SetNRGBA(p.X, p.Y, sheet.NRGBAAt(p.X, p.Y))
			}
		}
	}
	return out
}

// harbourPlace is where a kit piece stands on its hex: scaled to width world
// px across its opaque box, the box centred at at from the hex centre.
type harbourPlace struct {
	width float64
	at    connectionPoint
}

// harbourPlaces are the kit pieces' places, in column order: the warehouse
// on harbourN, the cargo and crane on harbourS, east of the road, and the
// customs house on harbourW.
var harbourPlaces = [3]struct {
	id    string
	place harbourPlace
}{
	{"harbourN", harbourPlace{24, connectionPoint{0, -3}}},
	{"harbourS", harbourPlace{40, connectionPoint{17, -6}}},
	{"harbourW", harbourPlace{32, connectionPoint{0, 0}}},
}

// harbourSprite renders a standing piece and its shadow from im, which holds
// only the piece, at 2x and 4x on a hex-sized canvas.
func harbourSprite(im *image.NRGBA, at harbourPlace) (sprite, shadow map[int]*image.NRGBA, err error) {
	box := opaqueBox(im)
	if box.Dx() < 32 || box.Dy() < 32 {
		return nil, nil, fmt.Errorf("harbour piece: opaque box %v", box)
	}
	k := float64(box.Dx()) / at.width
	centre := connectionPoint{float64(box.Min.X+box.Max.X) / 2, float64(box.Min.Y+box.Max.Y) / 2}
	s := quayStrip{im: im, k: k}
	px := func(q connectionPoint) [4]float64 {
		p := centre.add(q.sub(at.at).mul(k))
		return s.sample(p[0], p[1])
	}
	sprite, shadow = map[int]*image.NRGBA{}, map[int]*image.NRGBA{}
	for _, d := range []int{2, 4} {
		sub := max(1, int(math.Ceil(k/float64(d))))
		img, sh := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d)), image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		for y := range 80 * d {
			for x := range 96 * d {
				var acc [4]float64
				var cast float64
				for sy := range sub {
					for sx := range sub {
						q := connectionPoint{(float64(x) + (float64(sx)+.5)/float64(sub)) / float64(d), (float64(y) + (float64(sy)+.5)/float64(sub)) / float64(d)}.sub(connectionPoint{48, 40})
						c := px(q)
						for i := range acc {
							acc[i] += c[i]
						}
						cast += px(q.sub(connectionPoint{shadowOffset, shadowOffset}))[3]
					}
				}
				n := float64(sub * sub)
				img.SetNRGBA(x, y, straight(acc, n))
				if a := cast / n * shadowAlpha; a > 0 {
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

// harbourCrop is the rectangle of logical pixels of the hex canvas that holds
// every pixel of a standing piece and of its shadow.
func harbourCrop(sprite, shadow map[int]*image.NRGBA) image.Rectangle {
	var box image.Rectangle
	for _, im := range []*image.NRGBA{sprite[4], shadow[4]} {
		b := im.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if im.NRGBAAt(x, y).A > 0 {
					box = box.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
	}
	return image.Rect(box.Min.X/4, box.Min.Y/4, (box.Max.X+3)/4, (box.Max.Y+3)/4)
}

// cropTiles cuts the 2x and 4x tiles to r, in logical pixels.
func cropTiles(tiles map[int]*image.NRGBA, r image.Rectangle) map[int]*image.NRGBA {
	out := map[int]*image.NRGBA{}
	for d, im := range tiles {
		c := image.NewNRGBA(image.Rect(0, 0, r.Dx()*d, r.Dy()*d))
		for y := range c.Rect.Dy() {
			copy(c.Pix[c.PixOffset(0, y):c.PixOffset(0, y+1)], im.Pix[im.PixOffset(r.Min.X*d, r.Min.Y*d+y):im.PixOffset(r.Max.X*d, r.Min.Y*d+y)])
		}
		out[d] = c
	}
	return out
}
