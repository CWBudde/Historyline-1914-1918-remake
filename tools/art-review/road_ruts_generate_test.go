//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateRutRoads -count=1 -v
// Renders the summer and autumn roads (tile/road*, tile/road*_au,
// connected-terrain-v3) into assets/terrain/road/ and the level crossings of
// the same seasons (tile/cross*, tile/cross*_au, connected-mixed-v4) into
// assets/terrain/cross/, overwriting their PNGs and updating their hashes in
// assets/index.json; run `go run ./cmd/hl art index` afterwards. The winter
// forms belong to TestGenerateWinterRoads and TestGenerateWinterPathsAndRails.
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
	"testing"
)

func TestGenerateRutRoads(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{} // asset path -> new hash
	write := func(id, family string, tiles map[int]*image.NRGBA) {
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/%s/%s-%dx.png", family, id, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			hashes[path] = hex.EncodeToString(sum[:])
		}
	}
	defer func() { exportSeason = "" }()
	for _, season := range []string{"", "autumn"} {
		exportSeason = season
		roads, err := loadConnectionInputs(filepath.Join(root, "art/pipeline/connected-terrain-v3.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range roads.manifest.Families {
			if f.ID != "road" {
				continue
			}
			for _, v := range f.Variants {
				tiles, err := exportConnectionTiles(roads, &f, &v)
				if err != nil {
					t.Fatal(err)
				}
				write(seasonID(v.ID), "road", tiles)
			}
		}
		mixed, err := loadMixedInputs(filepath.Join(root, "art/pipeline/connected-mixed-v4.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range mixed.manifest.Variants {
			if v.Kind != "crossing" {
				continue
			}
			tiles, err := exportMixedTiles(mixed, v)
			if err != nil {
				t.Fatal(err)
			}
			write(seasonID(v.RuntimeID), "cross", tiles)
		}
	}
	changed := 0
	for i, raw := range entries {
		var e campaignEntry
		if json.Unmarshal(raw, &e) != nil {
			continue
		}
		if h, ok := hashes[e.Path]; ok {
			entries[i] = bytes.Replace(raw, []byte(`"`+e.SHA256+`"`), []byte(`"`+h+`"`), 1)
			changed++
		}
	}
	if changed != len(hashes) {
		t.Fatalf("updated %d index entries for %d files", changed, len(hashes))
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
	t.Logf("rendered %d road and crossing files", len(hashes))
}
