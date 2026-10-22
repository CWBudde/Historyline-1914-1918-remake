package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

func TestLiteMaterialsPreservePeriodicPixels(t *testing.T) {
	for _, d := range []int{2, 4} {
		t.Run(fmt.Sprintf("%dx", d), func(t *testing.T) {
			period := image.Pt(568, 400)
			src := fstest.MapFS{}
			m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("terrain-materials-%dx", d), Status: "candidate"}
			originals := map[string]*image.NRGBA{}
			for i, id := range []string{worldmaterial.MeadowID, worldmaterial.ScrubID} {
				im := image.NewNRGBA(image.Rect(0, 0, period.X*d, period.Y*d))
				state := uint32(7 + i)
				for y := 0; y < im.Rect.Dy()-1; y++ {
					for x := 0; x < im.Rect.Dx()-1; x++ {
						state ^= state << 13
						state ^= state >> 17
						state ^= state << 5
						im.SetNRGBA(x, y, color.NRGBA{uint8(state), uint8(state >> 8), uint8(state >> 16), 255})
					}
				}
				for y := 0; y < im.Rect.Dy()-1; y++ {
					im.SetNRGBA(im.Rect.Dx()-1, y, im.NRGBAAt(0, y))
				}
				for x := 0; x < im.Rect.Dx(); x++ {
					im.SetNRGBA(x, im.Rect.Dy()-1, im.NRGBAAt(x, 0))
				}
				originals[id] = im
				var b bytes.Buffer
				if err := png.Encode(&b, im); err != nil {
					t.Fatal(err)
				}
				name := fmt.Sprintf("images/%d.png", i)
				src[name] = &fstest.MapFile{Data: b.Bytes()}
				sum := sha256.Sum256(b.Bytes())
				m.Entries = append(m.Entries, assets.Entry{ID: id, Path: name, SHA256: hex.EncodeToString(sum[:]), Width: im.Rect.Dx(), Height: im.Rect.Dy(), Density: d})
			}
			p, err := assets.Load(src, m, assets.Options{AllowCandidates: true})
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(t.TempDir(), "lite")
			o, err := newOutput(out)
			if err != nil {
				t.Fatal(err)
			}
			defer o.close()
			var st liteStats
			if err = o.writeLitePack("materials", src, p, &st); err != nil {
				t.Fatal(err)
			}
			o.done()
			got, err := assets.Open(os.DirFS(filepath.Join(out, "materials")), "manifest.json", assets.Options{AllowCandidates: true})
			if err != nil {
				t.Fatal(err)
			}
			decoded := map[string]image.Image{}
			for _, e := range got.Manifest().Entries {
				if path.Ext(e.Path) != ".png" {
					t.Errorf("%s encoded as %s; periodic material requires PNG", e.ID, e.Path)
				}
				im, err := got.Image(e.ID, d)
				if err != nil {
					t.Fatal(err)
				}
				decoded[e.ID] = im
				original := originals[e.ID]
				for _, pt := range []image.Point{{0, 13}, {im.Bounds().Dx() - 1, 13}, {17, 0}, {17, im.Bounds().Dy() - 1}, {37, 41}} {
					if c := color.NRGBAModel.Convert(im.At(pt.X, pt.Y)).(color.NRGBA); c != original.NRGBAAt(pt.X, pt.Y) {
						t.Errorf("%s changed pixel %v: %v", e.ID, pt, c)
					}
				}
			}
			if _, err = worldmaterial.New(decoded[worldmaterial.MeadowID], decoded[worldmaterial.ScrubID], period, d); err != nil {
				t.Fatalf("lite materials rejected: %v", err)
			}
		})
	}
}
