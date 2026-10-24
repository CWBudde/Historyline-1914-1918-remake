//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCompoundShedMasters -count=1 -v
// Writes the depot and factory masters whose buildings stay inside the hex
// edge (compoundShedEdits: cmp-summer-v3, cmp-winter-v4, factory-summer-v3,
// factory-winter-v5) from the masters before them; run it before
// TestGenerateCompoundTiles and `go run ./cmd/hlart index`.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCompoundShedMasters(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, e := range compoundShedEdits() {
		base, err := loadPNG(filepath.Join(root, e.Base))
		if err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		if err := png.Encode(&b, compoundShedApply(base, e)); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, e.Out), b.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		t.Logf("%s: %s", e.Out, hex.EncodeToString(sum[:]))
	}
}
