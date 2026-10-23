package worldmaterial

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// disc paints an opaque disc of colour c and radius r at (cx, cy).
func disc(im *image.NRGBA, cx, cy, r float64, c color.NRGBA) {
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if dx, dy := float64(x)+.5-cx, float64(y)+.5-cy; dx*dx+dy*dy <= r*r {
				im.SetNRGBA(x, y, c)
			}
		}
	}
}

func TestCutClearsHazeOnly(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	im.SetNRGBA(0, 0, color.NRGBA{10, 20, 30, 31})
	im.SetNRGBA(1, 0, color.NRGBA{10, 20, 30, 32})
	im.SetNRGBA(2, 0, color.NRGBA{10, 20, 30, 255})
	if h := Haze(im, 32); h != 0.25 {
		t.Fatalf("haze %v, want 0.25", h)
	}
	out := Cut(im, 32)
	if a := out.NRGBAAt(0, 0).A; a != 0 {
		t.Fatalf("alpha 31 kept as %d", a)
	}
	if out.NRGBAAt(1, 0) != im.NRGBAAt(1, 0) || out.NRGBAAt(2, 0) != im.NRGBAAt(2, 0) {
		t.Fatal("cut changed pixels at or above the cut")
	}
	if Haze(out, 32) != 0 || im.NRGBAAt(0, 0).A != 31 {
		t.Fatal("cut left haze or changed its input")
	}
}

func TestItemsGroupsComponentsByCentroidCell(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	green := color.NRGBA{50, 62, 26, 255}
	// A disc centred in the left cell that reaches 10 px into the right cell.
	disc(im, 80, 50, 30, green)
	// Two touching-free stones in the right cell form one item.
	disc(im, 150, 40, 8, green)
	disc(im, 175, 60, 8, green)
	// A speck under minArea is dropped.
	im.SetNRGBA(130, 90, green)
	items, err := Items(im, 2, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("%d items, want 2", len(items))
	}
	left, right := items[0], items[1]
	if left.Cell != (image.Point{0, 0}) || right.Cell != (image.Point{1, 0}) {
		t.Fatalf("cells %v %v", left.Cell, right.Cell)
	}
	if left.Bounds.Max.X != 110 {
		t.Fatalf("left item bounds %v, want it to cross into the right cell to x 110", left.Bounds)
	}
	if right.Bounds != image.Rect(142, 32, 183, 68) || right.Components != 2 {
		t.Fatalf("right item %v with %d components", right.Bounds, right.Components)
	}
	if math.Abs(left.Centroid[0]-80) > .01 || math.Abs(left.Centroid[1]-50) > .01 {
		t.Fatalf("left centroid %v", left.Centroid)
	}
	// The item image holds only its own pixels.
	own := right.Image(im)
	if own.NRGBAAt(100, 50).A != 0 || own.NRGBAAt(150, 40).A != 255 || own.NRGBAAt(130, 90).A != 0 {
		t.Fatal("item image carries foreign pixels")
	}
}

func TestResampleAveragesPremultiplied(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		im.SetNRGBA(0, y, color.NRGBA{200, 100, 0, 255})
		im.SetNRGBA(1, y, color.NRGBA{255, 255, 255, 0}) // transparent white must not lighten the edge
	}
	out, err := Resample(im, image.Rect(0, 0, 4, 2), 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if c := out.NRGBAAt(0, 0); c != (color.NRGBA{200, 100, 0, 128}) {
		t.Fatalf("edge pixel %v, want the opaque colour at half alpha", c)
	}
	if c := out.NRGBAAt(1, 0); c.A != 0 {
		t.Fatalf("empty half %v", c)
	}
	if _, err := Resample(im, image.Rect(0, 0, 4, 2), 0, 1); err == nil {
		t.Fatal("zero width accepted")
	}
}

func TestGradeToLandsOnTheTargetMean(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			im.SetNRGBA(x, y, color.NRGBA{uint8(100 + x), uint8(80 + y), 50, 255})
		}
	}
	im.SetNRGBA(0, 0, color.NRGBA{0, 0, 0, 0})
	target := [3]float64{140, 132, 118}
	out, got, err := GradeTo(im, target, 0)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		if math.Abs(got[i]-target[i]) > 1.5 {
			t.Fatalf("mean %v, want %v", got, target)
		}
	}
	if out.NRGBAAt(0, 0).A != 0 || out.NRGBAAt(5, 5).A != 255 {
		t.Fatal("grade changed alpha")
	}
	// Variation survives: the grade shifts, it does not flatten.
	if out.NRGBAAt(39, 10) == out.NRGBAAt(1, 10) {
		t.Fatal("grade flattened the item")
	}
}

func TestGradeToKeepsFarColours(t *testing.T) {
	// A beige stone with a dark-teal water contact: the grade moves the stone
	// to the target and leaves the teal close to where it was.
	im := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			im.SetNRGBA(x, y, color.NRGBA{120, 105, 80, 255})
		}
	}
	teal := color.NRGBA{30, 70, 72, 255}
	for x := 0; x < 20; x++ {
		im.SetNRGBA(x, 9, teal)
	}
	out, got, err := GradeTo(im, [3]float64{140, 132, 118}, 20)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got[0]-140) > 2 || math.Abs(got[2]-118) > 2 {
		t.Fatalf("mean %v", got)
	}
	tl, _, tb := toLab(teal)
	gl, _, gb := toLab(out.NRGBAAt(3, 9))
	if math.Abs(gl-tl) > 3 || math.Abs(gb-tb) > 3 {
		t.Fatalf("teal moved from L %.1f b %.1f to L %.1f b %.1f", tl, tb, gl, gb)
	}
}
