//go:build campaigngfx

package artreview

// Explicit offline generators: go test -tags campaigngfx ./tools/art-review -run 'TestGenerate(Ruin|Site)Tiles' -count=1 -v
// Cut the destroyed depot (art/pipeline/compound-ruin-v2.json) into tile/ruinN,
// ruinW, ruinE and ruinS, and the depot construction site
// (art/pipeline/compound-site-v1.json) into tile/siteN, siteW, siteE and siteS,
// in both seasons (and in autumn and spring where the summer scene has an edit,
// suffix _au/_sp) through the compound pipeline, write them to
// assets/terrain/<family>/, replacing earlier exports, and list them in
// assets/index.json; run `go run ./cmd/hl art index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGenerateRuinTiles(t *testing.T) { generateDiamondTiles(t, "compound-ruin-v2.json", "ruin") }

func TestGenerateSiteTiles(t *testing.T) { generateDiamondTiles(t, "compound-site-v1.json", "site") }

// generateDiamondTiles cuts the four-hex scene of manifest into the tiles of
// art family terrain/<family>, named <family>N, W, E and S.
func generateDiamondTiles(t *testing.T, manifest, family string) {
	t.Helper()
	root := filepath.Join("..", "..")
	in, err := loadCompoundInputs(filepath.Join(root, "art/pipeline", manifest))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, a := range bundle.artifacts {
		files[a.name] = a.data
	}
	suffixes := []string{"", "_w"}
	defer func() { exportSeason = "" }()
	for _, season := range []string{"autumn", "spring"} {
		exportSeason = season
		in, err := loadCompoundInputs(filepath.Join(root, "art/pipeline", manifest))
		if err != nil {
			t.Fatal(err)
		}
		if !compoundSeasonEdited(in) {
			continue
		}
		b, err := buildCompoundBundle(in)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range b.artifacts {
			files[a.name] = a.data
		}
		suffixes = append(suffixes, seasonSuffixes[season])
	}
	exportSeason = ""
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && e.Art == "terrain/"+family
	})
	if err := os.MkdirAll(filepath.Join(root, "assets/terrain", family), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{2, 4} {
		for _, sfx := range suffixes {
			for _, p := range []string{"E", "N", "S", "W"} {
				id := family + p + sfx
				data, ok := files[fmt.Sprintf("tiles/%s-%dx.png", id, d)]
				if !ok {
					t.Fatalf("compound bundle has no %s at %dx", id, d)
				}
				path := fmt.Sprintf("terrain/%s/%s-%dx.png", family, id, d)
				if err := os.WriteFile(filepath.Join(root, "assets", path), data, 0o644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(data)
				raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-backgrounds-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/" + family})
				if err != nil {
					t.Fatal(err)
				}
				entries = append(entries, raw)
			}
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
