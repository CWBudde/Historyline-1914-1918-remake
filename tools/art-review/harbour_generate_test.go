//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateHarbourTiles -count=1 -v
// Renders the paving of the three harbour pieces (tile/harbourN, harbourW,
// harbourS), their buildings and cargo (standing/harbour*) and the shadows,
// summer and winter (_w), from the selected harbour paintings into
// assets/terrain/harbour/, replacing earlier exports, and lists them in
// assets/index.json; run `go run ./cmd/hl art index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGenerateHarbourTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && e.Art == "terrain/harbour"
	})
	dir := filepath.Join(root, "assets/terrain/harbour")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// r is the tiles' rectangle of logical pixels on the hex canvas, whose
	// centre is the anchor.
	write := func(id, file, pack, layer string, tiles map[int]*image.NRGBA, r image.Rectangle) {
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/harbour/%s-%dx.png", file, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			raw, err := json.Marshal(airfieldEntry{campaignEntry{ID: id, Pack: fmt.Sprintf(pack, d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: r.Dx() * d, Height: r.Dy() * d, Anchor: [2]float64{float64(48 - r.Min.X), float64(40 - r.Min.Y)}, Art: "terrain/harbour"}, layer})
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
		}
	}
	for _, season := range []string{"summer", "winter"} {
		suffix := map[string]string{"summer": "", "winter": "_w"}[season]
		paving := readNRGBA(t, filepath.Join(root, "art/terrain/harbour/harbour-paving-"+season+"-v1.png"))
		for _, id := range []string{"harbourN", "harbourW", "harbourS"} {
			write("tile/"+id+suffix, id+suffix, "terrain-infrastructure-%dx", "", harbourPavingTiles(paving, id), image.Rect(0, 0, 96, 80))
		}
		pieces := harbourKitPieces(readNRGBA(t, filepath.Join(root, "art/terrain/harbour/harbour-kit-"+season+"-v1.png")))
		for i, p := range harbourPlaces {
			t.Logf("%s%s: painted box %v", p.id, suffix, opaqueBox(pieces[i]))
			sprite, shadow, err := harbourSprite(pieces[i], p.place)
			if err != nil {
				t.Fatal(err)
			}
			// Cut to what they cover: the standing pack stays within its budget.
			r := harbourCrop(sprite, shadow)
			write("standing/"+p.id+suffix, p.id+suffix+"-standing", "terrain-standing-%dx", "standing", cropTiles(sprite, r), r)
			write("standing/shadow/"+p.id+suffix, p.id+suffix+"-shadow", "terrain-standing-%dx", "standing", cropTiles(shadow, r), r)
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
