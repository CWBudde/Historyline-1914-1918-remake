//go:build riverbankgfx

package main

// Explicit deterministic source authoring:
// go test -tags riverbankgfx ./tools/art-review -run '^TestGenerateRiverBankGravelCrop$' -count=1
// Normal tests never write artwork. The parent and every crop pixel are checked.
//
// The stream bank material is an unchanged-pixel crop (400,640)-(784,960) of the
// pinned lakeside bank master: no interpolation, recolouring or retouching.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	const source = "art/terrain/stream/bank.png"
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
	full := filepath.Join("..", "..", source)
	if old, e := os.ReadFile(full); e == nil && !bytes.Equal(old, encoded.Bytes()) {
		t.Fatalf("refusing to replace different retained crop bytes: %s", source)
	} else if e != nil && !os.IsNotExist(e) {
		t.Fatal(e)
	}
	if err = os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(full, encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("retained %s SHA256 %s; all 122880 pixels match pinned parent", source, hash(encoded.Bytes()))
}
