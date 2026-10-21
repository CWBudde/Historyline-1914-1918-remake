//go:build grassgfx

package main

// Explicit selected-variant exporter:
// go test -tags grassgfx ./tools/art-review -run TestGenerateGrassVariations -count=1 -v
// Normal tests never write artwork. The canonical source is read-only here.
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

func TestGenerateGrassVariations(t *testing.T) {
	root := filepath.Join("..", "..")
	canonical := filepath.Join(root, "art/terrain/grass/grass-summer-v1.png")
	before, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	TestGenerateGrassDetailCandidate(t)
	directory := filepath.Join(root, "build/review/grass-refdesign-detail-v1")
	data, err := os.ReadFile(filepath.Join(directory, "grass-variations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Variants []struct {
			Outputs []packOutput `json:"outputs"`
		} `json:"variants"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if len(pins.Variants) != 16 {
		t.Fatal("expected sixteen selected grass variants")
	}
	seen := map[string]bool{}
	for _, variant := range pins.Variants {
		if len(variant.Outputs) != 2 {
			t.Fatal("every grass variant requires two densities")
		}
		for _, pin := range variant.Outputs {
			if !filepath.IsLocal(pin.Path) || !strings.HasPrefix(pin.Path, "assets/terrain/grass/") || seen[pin.Path] {
				t.Fatalf("invalid selected grass export %s", pin.Path)
			}
			seen[pin.Path] = true
			png, err := os.ReadFile(filepath.Join(directory, "exports", filepath.Base(pin.Path)))
			if err != nil {
				t.Fatal(err)
			}
			h := sha256.Sum256(png)
			if hex.EncodeToString(h[:]) != pin.SHA256 {
				t.Fatalf("candidate export hash mismatch: %s", pin.Path)
			}
			if err := os.WriteFile(filepath.Join(root, pin.Path), png, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	after, err := os.ReadFile(canonical)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("canonical grass source changed: %v", err)
	}
}
