//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCompoundShedMasters -count=1 -v
// Writes the depot and factory masters whose buildings stay inside the hex
// edge (compoundShedEdits: cmp-summer-v3/v4, cmp-winter-v4/v5,
// factory-summer-v3/v4, factory-winter-v5/v6) from the masters before them, in
// order; a master whose pixels already equal its edit is left as it is. Run it
// before TestGenerateCompoundTiles and `go run ./cmd/hlart index`.
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
		edit := compoundShedApply(base, e)
		if old, err := loadPNG(filepath.Join(root, e.Out)); err == nil && bytes.Equal(old.Pix, edit.Pix) {
			t.Logf("%s: unchanged", e.Out)
			continue
		}
		var b bytes.Buffer
		if err := png.Encode(&b, edit); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, e.Out), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		t.Logf("%s: %s", e.Out, hex.EncodeToString(sum[:]))
	}
}
