package seasonblend

import (
	"image"
	"image/color"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestExportWeightsReproduceMix(t *testing.T) {
	b := New(winterEast(t), all)
	c := hexgrid.Coord{Line: 10, Col: 8}
	x, y := hexgrid.Center(c)
	bounds := image.Rect(int(x)-64, int(y)-56, int(x)+64, int(y)+56)
	one := image.NewRGBA(bounds)
	zero := image.NewRGBA(bounds)
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			one.SetRGBA(px, py, color.RGBA{255, 255, 255, 255})
		}
	}
	seasons := b.Seasons(c)
	layers := make([][]Part, len(seasons))
	for i := range layers {
		im := zero
		if i == 0 {
			im = one
		}
		layers[i] = []Part{{Image: im, At: bounds.Min}}
	}
	mixed := b.Mix(c, layers, 1)
	for py := bounds.Min.Y; py < bounds.Max.Y; py += 3 {
		for px := bounds.Min.X; px < bounds.Max.X; px += 3 {
			q := b.Weights(c, float64(px)+0.5, float64(py)+0.5)
			sum := 0
			for _, v := range q {
				sum += v
			}
			if sum != 256 {
				t.Fatalf("weights sum to %d", sum)
			}
			if got, want := mixed.RGBAAt(px, py).R, uint8((q[0]*255+128)>>8); got != want {
				t.Fatalf("%d,%d: %d != %d", px, py, got, want)
			}
		}
	}
	if q := (*Blend)(nil).Weights(c, 0, 0); q != nil {
		t.Fatal(q)
	}
}
