package artreview

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// The road bridge's map tile is empty in every season: the standing deck is
// the whole bridge. Its winter form exists so that a winter map draws no
// summer ID (TestWinterMapsDrawNoSummerInfrastructure).
func TestWinterRoadBridgeTilesAreEmpty(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, id := range []string{"rBridge14_w", "rBridge25_w", "rBridge36_w"} {
		for d, name := range map[int]string{2: "-2x.png", 4: "-4x.png"} {
			f, err := os.Open(filepath.Join(root, "assets/terrain/rBridge", id+name))
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(f)
			if closeErr := f.Close(); closeErr != nil {
				t.Error(closeErr)
			}
			if err != nil {
				t.Fatal(err)
			}
			b := im.Bounds()
			if b.Dx() != 96*d || b.Dy() != 80*d {
				t.Errorf("%s%s: %v, want %dx%d", id, name, b.Size(), 96*d, 80*d)
			}
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					if _, _, _, a := im.At(x, y).RGBA(); a != 0 {
						t.Fatalf("%s%s: alpha %d at %d,%d", id, name, a, x, y)
					}
				}
			}
		}
	}
}
