//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateLongShips -count=1 -v
// Writes the long sprites of the two-cell ships (art/pipeline/unit-<family>-long-v1.json)
// to assets/units/<family>-long/ and replaces their entries in assets/index.json;
// run `go run ./cmd/hl art index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGenerateLongShips(t *testing.T) {
	root := filepath.Join("..", "..")
	name := filepath.Join(root, "assets/index.json")
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Schema int               `json:"schema_version"`
		Assets []json.RawMessage `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil || index.Schema != 1 {
		t.Fatalf("index: %v", err)
	}
	families := []string{"battleship", "destroyer", "submarine"}
	long := func(pack string) bool {
		return slices.ContainsFunc(families, func(f string) bool { return strings.HasPrefix(pack, "unit-"+f+"-long-") })
	}
	var entries []json.RawMessage
	for _, raw := range index.Assets {
		var e struct {
			Pack string `json:"pack"`
		}
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		if !long(e.Pack) {
			entries = append(entries, raw)
		}
	}
	for _, family := range families {
		in, err := loadShipInputs(filepath.Join(root, "art/pipeline/unit-"+family+"-long-v1.json"))
		if err != nil {
			t.Fatal(err)
		}
		sprites, frames, err := shipSprites(in)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range frames {
			t.Logf("%s %s: scale %.3f", family, f.Facing, f.Scale4x)
		}
		for _, path := range slices.Sorted(maps.Keys(sprites)) {
			im := sprites[path]
			var data bytes.Buffer
			if err := png.Encode(&data, im); err != nil {
				t.Fatal(err)
			}
			full := filepath.Join(root, "assets/units", path)
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, data.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data.Bytes())
			id := strings.TrimSuffix(path, ".png")
			d := 2
			if strings.HasSuffix(id, "-4x") {
				d = 4
			}
			id = strings.TrimSuffix(strings.TrimSuffix(id, "-2x"), "-4x")
			raw, err := json.Marshal(map[string]any{
				"art": "units/" + family, "density": d, "height": im.Bounds().Dy(), "id": "unit/" + id,
				"logical_anchor": [2]float64{shipCanvas / 2, shipCanvas / 2}, "pack": "unit-" + family + "-long-" + map[int]string{2: "2x", 4: "4x"}[d],
				"path": "units/" + path, "sha256": hex.EncodeToString(sum[:]), "width": im.Bounds().Dx(),
			})
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
	if err := os.WriteFile(name, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
