package hlwebapp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackdropIsAFileOfItsOwn(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 600, 300))
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			im.SetRGBA(x, y, color.RGBA{90, 110, 70, 255})
		}
	}
	b := blobWriter{root: t.TempDir()}
	ref, err := b.backdrop(context.Background(), im)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(b.root, ref.Path))
	if err != nil {
		t.Fatal(err)
	}
	if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != ref.SHA256 || len(raw) != ref.Bytes || !strings.HasSuffix(ref.Path, ".json") {
		t.Fatalf("backdrop reference does not pin its file: %+v", ref)
	}
	var p Pyramid
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	if p.Width != 600 || p.Height != 300 || len(p.Mips) == 0 {
		t.Fatalf("pyramid: %+v", p)
	}
	for _, mip := range p.Mips {
		for _, tile := range mip.Tiles {
			if _, err := os.Stat(filepath.Join(b.root, tile.File.Path)); err != nil {
				t.Fatal(err)
			}
		}
	}

	// The index names the file and never inlines the pyramid.
	encoded, err := json.Marshal(LevelAssets{BackdropFile: &ref})
	if err != nil {
		t.Fatal(err)
	}
	var level map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &level); err != nil {
		t.Fatal(err)
	}
	if _, ok := level["backdropFile"]; !ok {
		t.Fatalf("level lost its backdrop file: %s", encoded)
	}
	if _, ok := level["backdrop"]; ok {
		t.Fatalf("level inlines its backdrop: %s", encoded)
	}
}
