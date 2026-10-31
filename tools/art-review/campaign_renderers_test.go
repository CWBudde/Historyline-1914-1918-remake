//go:build campaigngfx

package artreview

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// The renderers used for new campaign exports must match the selected terrain.
// Summer forest is already fully exported as separate layers, checked by
// TestCampaignPreservesLayeredSummerForest; its baked renderer is not used.
func TestCampaignRenderersMatchSelectedTerrain(t *testing.T) {
	root := filepath.Join("..", "..")
	ground, forest, hills, err := loadCampaignRenderers(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ family, id string }{{"ground", "mud"}, {"ground", "shelled_w"}, {"forest", "forest_w"}, {"hills", "hills"}} {
		t.Run(tc.id, func(t *testing.T) {
			base := "grass"
			if tc.id == "shelled_w" || tc.id == "forest_w" {
				base = "grass_w"
			}
			neighbors := []string{base, base, base, base, base, base}
			var tiles map[int]*image.NRGBA
			switch tc.family {
			case "ground":
				c, e := newGroundContext(tc.id, neighbors)
				if e != nil {
					t.Fatal(e)
				}
				tiles, err = ground.tiles(c)
			case "forest":
				c, e := newForestContext(tc.id, neighbors)
				if e != nil {
					t.Fatal(e)
				}
				tiles, err = forest.tiles(c)
			case "hills":
				c, e := newHillsContext(tc.id, neighbors, connectionPoint{})
				if e != nil {
					t.Fatal(e)
				}
				tiles, err = hills.tiles(c)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, scale := range []int{2, 4} {
				path := filepath.Join(root, "assets/terrain", tc.family, fmt.Sprintf("%s-%dx.png", tc.id, scale))
				data, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				expected, e := png.Decode(bytes.NewReader(data))
				if e != nil {
					t.Fatal(e)
				}
				for y := range expected.Bounds().Dy() {
					for x := range expected.Bounds().Dx() {
						if tiles[scale].At(x, y) != expected.At(x, y) {
							t.Fatalf("campaign renderer differs from selected %s at (%d,%d)", path, x, y)
						}
					}
				}
			}
		})
	}
}
