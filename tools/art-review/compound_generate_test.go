//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCompoundTiles -count=1 -v
// Renders the selected HQ, depot and factory scenes and replaces their
// installed tile PNGs (tile/hq*, cmp*, factory*, both seasons) in place,
// updating each file's hash in assets/index.json; run `go run ./cmd/hlart index`
// afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
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
	for _, manifest := range []string{"compound-hq-v4.json", "compound-depot-v5.json", "compound-factory-v5.json", "compound-factory-summer-v3.json"} {
		in, err := loadCompoundInputs(filepath.Join(root, "art/pipeline", manifest))
		if err != nil {
			t.Fatal(err)
		}
		b, err := buildCompoundBundle(in)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range b.artifacts {
			name, ok := strings.CutPrefix(a.name, "tiles/")
			if !ok {
				continue
			}
			id, density, ok := strings.Cut(strings.TrimSuffix(name, ".png"), "-")
			if !ok || (density != "2x" && density != "4x") {
				t.Fatalf("%s: unexpected tile %s", manifest, a.name)
			}
			key := campaignExportKey("tile/"+id, int(density[0]-'0'))
			e, ok := exported[key]
			if !ok {
				t.Fatalf("%s: %s is not installed", manifest, key)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", e.Path), a.data, 0644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(a.data)
			i := at[key]
			entries[i] = bytes.Replace(entries[i], []byte(e.SHA256), []byte(hex.EncodeToString(sum[:])), 1)
			replaced++
		}
	}
	t.Logf("replaced %d compound tiles", replaced)
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
