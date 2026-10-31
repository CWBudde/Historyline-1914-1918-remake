package artreview

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// The paving covers each harbour hex and runs on over the edges the pieces
// share, fades out along the outer edges, and leaves the road at the yard's S
// edge uncovered for a stretch.
func TestHarbourPavingGeometry(t *testing.T) {
	paving := uniform(1536, 1024, color.NRGBA{130, 120, 105, 255})
	for _, c := range []struct {
		piece string
		p     connectionPoint // world point relative to the piece's centre
		min   uint8
		max   uint8
	}{
		{"harbourN", connectionPoint{0, 0}, 255, 255},
		{"harbourN", connectionPoint{0, 39}, 255, 255},   // the yard's edge
		{"harbourN", connectionPoint{-35, 19}, 255, 255}, // the customs house's edge
		{"harbourN", connectionPoint{0, -39.5}, 0, 0},    // the outer N edge
		{"harbourN", connectionPoint{46, 0}, 0, 0},       // the outer E corner
		{"harbourW", connectionPoint{35, -19}, 255, 255},
		{"harbourW", connectionPoint{35, 19}, 255, 255},
		{"harbourW", connectionPoint{-46, 0}, 0, 0},
		{"harbourS", connectionPoint{0, -39}, 255, 255},
		{"harbourS", connectionPoint{-20, 10}, 255, 255},
		{"harbourS", connectionPoint{0, 25}, 0, 40}, // the road comes in
		{"harbourS", connectionPoint{0, 39.5}, 0, 0},
	} {
		tiles := harbourPavingTiles(paving, c.piece)
		if got := alphaAt(tiles, c.p.add(connectionPoint{48, 40})); got < c.min || got > c.max {
			t.Errorf("%s alpha at %v = %d, want %d–%d", c.piece, c.p, got, c.min, c.max)
		}
	}
}

// The pieces of the kit are the opaque shapes whose centre lies in its
// column, so a crane jib reaching into the next column stays with its crane.
func TestHarbourKitPieces(t *testing.T) {
	sheet := image.NewNRGBA(image.Rect(0, 0, 1536, 1024))
	fill := func(r image.Rectangle, c color.NRGBA) {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				sheet.SetNRGBA(x, y, c)
			}
		}
	}
	fill(image.Rect(100, 100, 400, 900), color.NRGBA{100, 100, 110, 255}) // the warehouse
	fill(image.Rect(600, 300, 900, 700), color.NRGBA{150, 120, 90, 255})  // the cargo
	fill(image.Rect(880, 300, 1100, 320), color.NRGBA{40, 40, 40, 255})   // the crane's jib
	fill(image.Rect(1150, 400, 1450, 800), color.NRGBA{150, 80, 60, 255}) // the customs house
	fill(image.Rect(1300, 100, 1302, 102), color.NRGBA{150, 80, 60, 20})  // haze
	pieces := harbourKitPieces(sheet)
	for i, want := range []image.Rectangle{image.Rect(100, 100, 400, 900), image.Rect(600, 300, 1100, 700), image.Rect(1150, 400, 1450, 800)} {
		if got := opaqueBox(pieces[i]); got != want {
			t.Errorf("piece %d: box %v, want %v", i, got, want)
		}
	}
}

// A building is scaled to its width in world px, placed on its spot of the
// hex and casts its shadow toward the lower right.
func TestHarbourSprite(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 512, 1024))
	body := uniform(200, 400, color.NRGBA{100, 100, 110, 255})
	for y := range 400 {
		copy(im.Pix[im.PixOffset(150, 300+y):], body.Pix[body.PixOffset(0, y):body.PixOffset(0, y+1)])
	}
	sprite, shadow, err := harbourSprite(im, harbourPlace{width: 24, at: connectionPoint{10, -2}})
	if err != nil {
		t.Fatal(err)
	}
	box := opaqueBox(sprite[2])
	if box.Dx() != 48 || box.Min.X != 2*(48+10-12) || math.Abs(float64(box.Dy())-96) > 1.5 || math.Abs(float64(box.Min.Y)-2*(40-2-24)) > 1.5 {
		t.Errorf("box at 2x %v", box)
	}
	if a := shadow[2].NRGBAAt(box.Max.X+2, box.Max.Y-10).A; a == 0 {
		t.Error("no shadow beside the lower right")
	}
	if a := shadow[2].NRGBAAt(box.Min.X-2, box.Min.Y+10).A; a != 0 {
		t.Errorf("shadow on the upper left: %d", a)
	}
	if _, _, err := harbourSprite(uniform(512, 1024, color.NRGBA{}), harbourPlace{width: 24}); err == nil {
		t.Error("an empty piece makes a sprite")
	}
}

// A standing piece is cut to the logical pixels its sprite and shadow cover,
// the same at both densities, so the standing pack stays small.
func TestHarbourCrop(t *testing.T) {
	sprite, shadow := map[int]*image.NRGBA{}, map[int]*image.NRGBA{}
	for _, d := range []int{2, 4} {
		sprite[d], shadow[d] = image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d)), image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
	}
	sprite[4].SetNRGBA(4*10+1, 4*20, color.NRGBA{1, 2, 3, 4})
	shadow[4].SetNRGBA(4*30+3, 4*50+3, color.NRGBA{1, 2, 3, 4})
	r := harbourCrop(sprite, shadow)
	if r != image.Rect(10, 20, 31, 51) {
		t.Fatalf("crop %v", r)
	}
	for _, d := range []int{2, 4} {
		c := cropTiles(sprite, r)[d]
		if c.Bounds() != image.Rect(0, 0, 21*d, 31*d) {
			t.Errorf("%dx: bounds %v", d, c.Bounds())
		}
	}
	if a := cropTiles(sprite, r)[4].NRGBAAt(1, 0).A; a != 4 {
		t.Errorf("the sprite's pixel moved: alpha %d", a)
	}
}

// Each kit column has its marker, and the piece placed there stays within the
// hex's width.
func TestHarbourPlaces(t *testing.T) {
	for i, want := range []string{"harbourN", "harbourS", "harbourW"} {
		p := harbourPlaces[i]
		if p.id != want {
			t.Errorf("column %d places %s, want %s", i, p.id, want)
		}
		if half := p.place.width / 2; p.place.at[0]-half < -48 || p.place.at[0]+half > 48 {
			t.Errorf("%s spans x %.0f..%.0f, beyond the hex", p.id, p.place.at[0]-half, p.place.at[0]+half)
		}
	}
}
