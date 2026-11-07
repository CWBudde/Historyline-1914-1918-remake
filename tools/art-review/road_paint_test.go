package artreview

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	// roadPaintLevel is the lowest channel of a paint pixel: the paint is
	// RGB(242,240,237), and its thin lines keep about this much after the 2x filter.
	roadPaintLevel = 185
	// roadPaintLimit bounds the stray pale pebbles of one road tile at 4x;
	// crossPaintLimit adds the rail heads' highlights of a crossing.
	roadPaintLimit  = 40
	crossPaintLimit = 100
)

var roadPaintFamily = regexp.MustCompile(`^tile/(road|cross)[0-9]`)

// roadPaintPixels counts the opaque, paint-white pixels of a tile.
func roadPaintPixels(im image.Image) int {
	n := 0
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := im.At(x, y).RGBA()
			if a>>8 >= 128 && min(r, g, bl)>>8 >= roadPaintLevel {
				n++
			}
		}
	}
	return n
}

// Roads of the period carry no lane paint: no installed summer, autumn or
// spring road or road crossing tile shows paint-white pixels at any density.
func TestInstalledRoadsCarryNoPaint(t *testing.T) {
	root := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Density int    `json:"density"`
			Path    string `json:"path"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	worst := map[int]int{}
	for _, e := range index.Assets {
		if !roadPaintFamily.MatchString(e.ID) || campaignSeasonSuffix(e.ID) == "_w" {
			continue
		}
		f, err := os.Open(filepath.Join(root, "assets", e.Path))
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
		n := roadPaintPixels(im)
		worst[e.Density] = max(worst[e.Density], n)
		seen[campaignSeasonSuffix(e.ID)]++
		limit := roadPaintLimit
		if strings.HasPrefix(e.ID, "tile/cross") {
			limit = crossPaintLimit
		}
		if n > limit {
			t.Errorf("%s: %d paint-white pixels at %dx, want at most %d", e.ID, n, e.Density, limit)
		}
	}
	t.Logf("paint-white pixels at most per tile by density: %v; tiles by season suffix: %v", worst, seen)
	// 16 roads and 4 crossings at 2x and 4x, summer and autumn at least.
	if seen[""] != 40 || seen["_au"] != 40 {
		t.Errorf("checked %v tiles by season suffix, want 40 summer and 40 autumn", seen)
	}
}

// campaignSeasonSuffix is the runtime season suffix of id, "" for summer.
func campaignSeasonSuffix(id string) string {
	return campaignSeasonal.FindString(id)
}
