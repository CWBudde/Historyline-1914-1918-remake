package artreview

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func uniform(w, h int, c color.NRGBA) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(im.Pix); i += 4 {
		im.Pix[i], im.Pix[i+1], im.Pix[i+2], im.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	return im
}

// alphaAt is the 2x tile's alpha at world point p of its hex.
func alphaAt(tiles map[int]*image.NRGBA, p connectionPoint) uint8 {
	return tiles[2].NRGBAAt(int(p[0]*2), int(p[1]*2)).A
}

// The runway covers a band down the hex centre; the apron runs from the
// hangar's doors over its hex's east corner into the runway hexes at its NE
// and SE, up to the runway.
func TestAirfieldGeometry(t *testing.T) {
	strip := newRunwayStrip(uniform(1536, 1024, color.NRGBA{120, 100, 80, 255}), 512, 750/80.)
	runway, err := runwayTiles(strip)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		p    connectionPoint
		want uint8
	}{{connectionPoint{48, 1}, 255}, {connectionPoint{48, 79}, 255}, {connectionPoint{33, 40}, 255}, {connectionPoint{18, 40}, 0}, {connectionPoint{78, 40}, 0}} {
		if got := alphaAt(runway, c.p); got != c.want {
			t.Errorf("runway alpha at %v = %d, want %d", c.p, got, c.want)
		}
	}
	apron := map[string]map[int]*image.NRGBA{}
	for id, offset := range apronPieces {
		apron[id] = apronTiles(strip, offset)
	}
	for _, c := range []struct {
		id   string
		p    connectionPoint // world point relative to the hangar's centre
		want uint8
	}{
		{"hangar", connectionPoint{35, 0}, 255},
		{"hangar", connectionPoint{0, 0}, 0},   // under the hangar
		{"hangar", connectionPoint{35, 20}, 0}, // beside the apron
		{"apron-sw", connectionPoint{47, -5}, 255},
		{"apron-nw", connectionPoint{47, 5}, 255},
		{"apron-nw", connectionPoint{65, 5}, 0}, // the runway itself
	} {
		q := c.p.sub(apronPieces[c.id]).add(connectionPoint{48, 40})
		if got := alphaAt(apron[c.id], q); got != c.want {
			t.Errorf("%s alpha at hangar%+v = %d, want %d", c.id, c.p, got, c.want)
		}
	}
}

// The hangar is scaled to hangarLength world px, centred on the hex, and casts
// its shadow toward the lower right.
func TestHangarSprite(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 1536, 1024))
	body := uniform(500, 360, color.NRGBA{150, 140, 110, 255})
	for y := range 360 {
		copy(im.Pix[im.PixOffset(500, 300+y):], body.Pix[body.PixOffset(0, y):body.PixOffset(0, y+1)])
	}
	sprite, shadow, err := hangarSprite(im)
	if err != nil {
		t.Fatal(err)
	}
	if box := opaqueBox(sprite[2]); box.Dx() != 2*hangarLength || box.Min.X != 96-hangarLength || math.Abs(float64(box.Dy())-2*360*hangarLength/500) > 1.5 {
		t.Errorf("hangar box at 2x %v", box)
	}
	if a := shadow[2].NRGBAAt(96+hangarLength+2, 80).A; a == 0 {
		t.Error("no shadow beside the hangar's lower right")
	}
	if a := shadow[2].NRGBAAt(96-hangarLength-2, 80).A; a != 0 {
		t.Errorf("shadow on the hangar's upper left: %d", a)
	}
	if _, _, err := hangarSprite(uniform(1536, 1024, color.NRGBA{})); err == nil {
		t.Error("an empty painting makes a hangar")
	}
}
