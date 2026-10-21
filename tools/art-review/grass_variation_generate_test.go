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
	canonical := filepath.Join(root, "art/source/terrain/refdesign-v1/grass-base-01.png")
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
	var recipe map[string]any
	if err := json.Unmarshal(data, &recipe); err != nil {
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
			if !filepath.IsLocal(pin.Path) || !strings.HasPrefix(pin.Path, "art/exports/grass-refdesign-v1/") || seen[pin.Path] {
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
	recipe["candidate_generation_command"] = recipe["generation_command"]
	recipe["generation_command"] = "go test -tags grassgfx ./tools/art-review -run TestGenerateGrassVariations -count=1 -v"
	recipe["material_mix"] = map[string]any{
		"base_indices": []int{0, 5, 10, 15}, "flowers_indices": []int{1, 3, 6, 8, 11, 13}, "worn_indices": []int{2, 4, 7, 9, 12, 14},
		"authored_detail_strength": .65, "phased_canonical_strength": .35,
		"description": "Four variants retain the unchanged phased canonical meadow. For the other twelve, interior paint is65percent the selected authored material plus35percent the same phased canonical meadow. The8worldpixel canonical guard and smoothstep8..18 boundary envelope remain unchanged.",
	}
	recipe["promotion_scope"] = "Only these32 grass-detail PNGs and this recipe. Canonical grass source and all other terrain-family exports stay byte-identical."
	data, err = json.MarshalIndent(recipe, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "art/recipes/terrain/refdesign-v1/grass-variations.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(canonical)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("canonical grass source changed: %v", err)
	}
}
