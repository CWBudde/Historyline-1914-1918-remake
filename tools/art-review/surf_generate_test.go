//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateSurf -count=1 -v
// Packs the four surf strips of art/terrain/water/water-surf-shared-v1.png
// (12 px per world px, 1536 px wide and tileable left to right) into
// tile/surf: a 96×80 world px texture of four rows of 20 world px, row k the
// first 20 world px of strip k from its coast side down, each strip's 128
// world px pressed into the 96 of the tile. The renderers read it by the
// distance from the coast (internal/waterdepth). The winter form (suffix _w)
// is the same pixels. Writes them to assets/terrain/water/, replacing earlier
// exports, and lists them in assets/index.json; run `go run ./cmd/hl art index`
// afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// surfTops are the rows of the painting where the foam of each strip begins,
// its coast side.
var surfTops = [4]int{100, 336, 578, 814}

func TestGenerateSurf(t *testing.T) {
	root := filepath.Join("..", "..")
	f, err := os.Open(filepath.Join(root, "art/terrain/water/water-surf-shared-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	paint := image.NewNRGBA(src.Bounds())
	draw.Draw(paint, paint.Rect, src, src.Bounds().Min, draw.Src)
	if paint.Rect.Dx() != 1536 || paint.Rect.Dy() != 1024 {
		t.Fatalf("surf painting is %v, want 1536x1024", paint.Rect.Size())
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && strings.TrimSuffix(e.ID, "_w") == "tile/surf"
	})
	for _, d := range []int{2, 4} {
		tile := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		// Source pixels per tile pixel: 16 along the coast, 12 across it.
		sx, sy := 1536.0/float64(96*d), 12.0/float64(d)
		for y := range 80 * d {
			k := y / (20 * d)
			top := float64(surfTops[k]) + float64(y-20*d*k)*sy
			for x := range 96 * d {
				left := float64(x) * sx
				var sum [4]float64
				n := 0.
				for py := int(top); py < int(math.Ceil(top+sy)); py++ {
					for px := int(left); px < int(math.Ceil(left+sx)); px++ {
						c := paint.NRGBAAt(px%1536, py)
						a := float64(c.A)
						sum[0] += float64(c.R) * a
						sum[1] += float64(c.G) * a
						sum[2] += float64(c.B) * a
						sum[3] += a
						n++
					}
				}
				if sum[3] == 0 {
					continue
				}
				tile.SetNRGBA(x, y, color.NRGBA{uint8(math.Round(sum[0] / sum[3])), uint8(math.Round(sum[1] / sum[3])), uint8(math.Round(sum[2] / sum[3])), uint8(math.Round(sum[3] / n))})
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, tile); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(buf.Bytes())
		for _, id := range []string{"surf", "surf_w"} {
			path := fmt.Sprintf("terrain/water/%s-%dx.png", id, d)
			if err := os.WriteFile(filepath.Join(root, "assets", path), buf.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/water"})
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
		}
	}
	var out bytes.Buffer
	out.WriteString("{\n  \"schema_version\": 1,\n  \"assets\": [\n")
	for i, e := range entries {
		out.WriteString("    ")
		out.Write(e)
		if i+1 < len(entries) {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString("  ]\n}\n")
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
