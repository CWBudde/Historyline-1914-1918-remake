//go:build riverbankgfx

package main

// Explicit deterministic source authoring:
// go test -tags riverbankgfx ./tools/art-review -run '^TestGenerateRiverBankGravelCrop$' -count=1
// Normal tests never write artwork. The parent and every crop pixel are checked.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateRiverBankGravelCrop(t *testing.T) {
	const parent = "art/terrain/lakeside/bank.png"
	const parentPin = "d43b40a2ed64e95cbb6afdacca4a384a39ec59c8a54d05bdcbe77a60e6aa6613"
	const parentRecipe = "art/recipes/terrain/shore-reference-v2/bank.json"
	const source = "art/terrain/stream/bank.png"
	const recipe = "art/recipes/terrain/refdesign-v1/river-bank-gravel.json"
	read := func(path string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", "..", path))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	hash := func(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
	parentBytes := read(parent)
	if hash(parentBytes) != parentPin {
		t.Fatal("river bank crop parent pin changed")
	}
	master, err := png.Decode(bytes.NewReader(parentBytes))
	if err != nil {
		t.Fatal(err)
	}
	box := image.Rect(400, 640, 784, 960)
	if !box.In(master.Bounds()) {
		t.Fatal("crop exceeds parent bounds")
	}
	crop := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(crop, crop.Bounds(), master, box.Min, draw.Src)
	var encoded bytes.Buffer
	if err = png.Encode(&encoded, crop); err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for y := range 320 {
		for x := range 384 {
			got := color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA)
			want := color.NRGBAModel.Convert(master.At(x+400, y+640)).(color.NRGBA)
			if got != want {
				t.Fatalf("crop modified parent pixel at %d,%d", x, y)
			}
		}
	}
	record := map[string]any{
		"schema_version": 1, "id": "river-bank-gravel", "date": "2026-10-20", "art_direction": "refdesign-v1", "status": "selected-source-awaiting-connected-bundle-review",
		"tool":   "Go image/png and image/draw; deterministic unchanged-pixel crop",
		"source": source, "source_dimensions": []int{384, 320}, "source_sha256": hash(encoded.Bytes()), "transparent_background": false,
		"generation_inputs":     []any{map[string]any{"path": parent, "sha256": parentPin, "recipe": parentRecipe, "recipe_sha256": hash(read(parentRecipe)), "role": "new-authored-parent-material; exact decoded RGBA crop, no original raster input"}},
		"original_image_inputs": []any{}, "crop_rectangle_xyxy": []int{400, 640, 784, 960},
		"method":             "Decode the pinned 1254x1254 parent PNG. image.Draw Src copies rectangle (400,640)-(784,960) to a 384x320 NRGBA canvas; encode PNG. All 122880 decoded output RGBA pixels are checked against the corresponding decoded parent pixel. No interpolation, recoloring, retouching or raster painting. Horizontal material features become 1254/384=3.265625 times coarser and vertical features 1254/320=3.91875 times coarser under the unchanged lattice material sampler.",
		"generation_command": "go test -tags riverbankgfx ./tools/art-review -run '^TestGenerateRiverBankGravelCrop$' -count=1",
		"consumer":           "connected stream bank only, width18 border3.5 with canonical eight-world-pixel port guards and 1.1-world-pixel interior bank wear; nine lakeside variants keep their existing full parent master",
		"review":             "384x320 crop selected after comparison with full master and 512x384 crop at 2x and 1.5x assembled scale; larger irregular loose stone detail without a regular paving pattern. Integrated map visual acceptance remains open.",
	}
	recipeBytes, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string][]byte{source: encoded.Bytes(), recipe: append(recipeBytes, '\n')} {
		full := filepath.Join("..", "..", path)
		if old, e := os.ReadFile(full); e == nil && !bytes.Equal(old, data) {
			t.Fatalf("refusing to replace different retained crop bytes: %s", path)
		} else if e != nil && !os.IsNotExist(e) {
			t.Fatal(e)
		}
		if err = os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(full, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("retained %s SHA256 %s; all 122880 pixels match pinned parent", source, hash(encoded.Bytes()))
}
