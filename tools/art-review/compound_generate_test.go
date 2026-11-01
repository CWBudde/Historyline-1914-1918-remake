//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCompoundTiles -count=1 -v
// Renders the selected HQ, depot, factory and massif scenes and replaces their
// installed tile PNGs (tile/hq*, cmp*, factory*, mount*, mountain*, both seasons) in place,
// updating each file's hash in assets/index.json; run `go run ./cmd/hl art index`
// afterwards. A manifest whose summer artwork has an autumn or spring edit
// also writes that season's tiles (summer ID + _au/_sp beside the summer
// file), adding the entries it lacks with the summer entry's pack and art key.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCompoundTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	at := map[string]int{}
	for i, raw := range entries {
		var e campaignEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		at[campaignExportKey(e.ID, e.Density)] = i
	}
	replaced := 0
	defer func() { exportSeason = "" }()
	for _, season := range []string{"", "autumn", "spring"} {
		exportSeason = season
		for _, manifest := range []string{"compound-hq-v4.json", "compound-depot-v6.json", "compound-factory-v6.json", "compound-factory-summer-v4.json", "compound-mountain-v3.json"} {
			replaced += generateCompoundTiles(t, root, manifest, &entries, exported, at)
		}
	}
	t.Logf("replaced or added %d compound tiles", replaced)
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

// generateCompoundTiles writes the tiles of one manifest in the export season;
// a seasonal export without an artwork edit writes nothing.
func generateCompoundTiles(t *testing.T, root, manifest string, entries *[]json.RawMessage, exported map[string]campaignEntry, at map[string]int) int {
	t.Helper()
	in, err := loadCompoundInputs(filepath.Join(root, "art/pipeline", manifest))
	if err != nil {
		t.Fatal(err)
	}
	if seasonActive() && !compoundSeasonEdited(in) {
		return 0
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	written := 0
	for _, a := range b.artifacts {
		name, ok := strings.CutPrefix(a.name, "tiles/")
		if !ok {
			continue
		}
		id, density, ok := strings.Cut(strings.TrimSuffix(name, ".png"), "-")
		if !ok || (density != "2x" && density != "4x") {
			t.Fatalf("%s: unexpected tile %s", manifest, a.name)
		}
		d := int(density[0] - '0')
		key := campaignExportKey("tile/"+id, d)
		sum := sha256.Sum256(a.data)
		e, ok := exported[key]
		if !ok {
			summerID, seasonal := strings.CutSuffix(id, seasonSuffixes[exportSeason])
			s, found := exported[campaignExportKey("tile/"+summerID, d)]
			if !seasonActive() || !seasonal || !found {
				t.Fatalf("%s: %s is not installed", manifest, key)
			}
			dir, _ := path.Split(s.Path)
			e = campaignEntry{ID: "tile/" + id, Pack: s.Pack, Density: d, Path: dir + name, SHA256: hex.EncodeToString(sum[:]), Width: s.Width, Height: s.Height, Anchor: s.Anchor, Art: s.Art}
			raw, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			at[key] = len(*entries)
			*entries = append(*entries, raw)
			exported[key] = e
		}
		if err := os.WriteFile(filepath.Join(root, "assets", e.Path), a.data, 0o644); err != nil {
			t.Fatal(err)
		}
		i := at[key]
		(*entries)[i] = bytes.Replace((*entries)[i], []byte(e.SHA256), []byte(hex.EncodeToString(sum[:])), 1)
		written++
	}
	return written
}
