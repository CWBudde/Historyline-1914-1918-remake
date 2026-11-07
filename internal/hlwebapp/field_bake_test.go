package hlwebapp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// The bake reworks as the tile map does and marks the result for it alone.
func TestFieldBakeMatchesNativeRework(t *testing.T) {
	lv := &level.Level{Lines: [][]level.Hex{{{Terrain: "mud"}, {Terrain: "mud"}}, {{Terrain: "mud"}}, {{Terrain: "mud"}, {Terrain: "grass"}}}}
	descriptor, ok := mapcompose.Descriptor(mapcompose.Ground, "mud", [6]string{"mud", "mud", "mud", "mud", "mud", "mud"}, [2]int{})
	if !ok {
		t.Fatal("no descriptor")
	}
	exemplar := mapcompose.ContextID(mapcompose.Ground, mapcompose.Key(descriptor))
	dir := t.TempDir()
	pictures := map[int]map[string]image.Image{}
	for _, d := range []int{1, 2, 4} {
		pack := filepath.Join(dir, strconv.Itoa(d)+"x", "ground")
		if err := os.MkdirAll(pack, 0o755); err != nil {
			t.Fatal(err)
		}
		m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("test/%d", d)}
		pictures[d] = map[string]image.Image{}
		for i, id := range []string{exemplar, "tile/mud", "tile/grass"} {
			im := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
			draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{90, 70, 40, 255}), image.Point{}, draw.Src)
			for y := range im.Rect.Dy() {
				for x := range im.Rect.Dx() {
					if (x/(5*d)+y/(7*d))%2 == 0 {
						im.SetNRGBA(x, y, color.NRGBA{50, 40, uint8(20 + i), 255})
					}
				}
			}
			pictures[d][id] = im
			var raw bytes.Buffer
			if err := png.Encode(&raw, im); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw.Bytes())
			name := fmt.Sprintf("%d.png", i)
			m.Entries = append(m.Entries, assets.Entry{ID: id, Path: name, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Density: d, Anchor: [2]float64{48, 40}})
			if err := os.WriteFile(filepath.Join(pack, name), raw.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		raw, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pack, "manifest.json"), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	scene := mapcompose.Compose(lv.Grid(), mapcompose.LevelHex(lv), func(id string) bool { return id == exemplar })
	b := blobWriter{root: t.TempDir()}
	baker, err := newWaterBaker(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	layers := BakedLayers{Scene: &scene}
	if err := baker.bake(context.Background(), lv, &layers); err != nil {
		t.Fatal(err)
	}
	c := hexgrid.Coord{}
	id := scene.Cells[0][0].Background
	f := scene.FieldAt(c, id)
	if f == nil {
		t.Fatal("the fixture has no field")
	}
	for _, d := range []int{1, 2, 4} {
		e := layers.Backgrounds["0,0"][strconv.Itoa(d)]
		if e.ID != id || !e.TileOnly || e.Full != e.Lite || e.Lite.Path == "" {
			t.Fatalf("%dx: %+v", d, e)
		}
		if _, err := os.Stat(filepath.Join(b.root, e.Lite.Path)); err != nil {
			t.Fatal(err)
		}
		got, err := baker.reworked(&scene, c, id, pictures[d][id], waterOrigin(c, [2]float64{48, 40}, d), d)
		if err != nil {
			t.Fatal(err)
		}
		want := mapcompose.NewFieldMaterial(pictures[d][exemplar], [2]float64{48, 40}, d).Rework(pictures[d][id], waterOrigin(c, [2]float64{48, 40}, d), c, f.Edges)
		if got.Bounds().Size() != want.Bounds().Size() {
			t.Fatalf("%dx: size %v", d, got.Bounds())
		}
		for y := range want.Rect.Dy() {
			for x := range want.Rect.Dx() {
				if color.NRGBAModel.Convert(got.At(x, y)) != color.Color(want.NRGBAAt(x, y)) {
					t.Fatalf("%dx: pixel %d,%d differs", d, x, y)
				}
			}
		}
	}
	if len(layers.Backgrounds["2,1"]) != 0 {
		t.Error("meadow baked")
	}
}
