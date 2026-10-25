//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWaterRocks -count=1 -v
// Renders tile/waterRocks from the water-stones scene of
// art/pipeline/compound-water-stones-v2.json and tile/waterRocks_w from the
// snow-capped rocks of compound-water-stones-winter-v1.json, kept to the summer
// rocks' alpha (waterRocksOutline): the rocks alone,
// without their water, so a stones hex draws them over the shore its
// neighbours give it. Over the open water tile they are tile/waterStones and
// tile/waterStones_w. Writes them to assets/terrain/water/, replacing earlier
// exports, and lists them in assets/index.json; run `go run ./cmd/hlart index`
// afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The water-stones scenes of each season.
const (
	waterStonesSummerManifest = "art/pipeline/compound-water-stones-v2.json"
	waterStonesWinterManifest = "art/pipeline/compound-water-stones-winter-v1.json"
)

// waterStonesExport is one density of a water-stones scene: the rocks alone
// (clipped to the hex) and the stones tile, the rocks over the scene's water,
// with its PNG bytes.
type waterStonesExport struct {
	rocks, stones *image.NRGBA
	stonesPNG     []byte
}

// waterStonesExports renders the water-stones scene of manifest at 2x and 4x.
// With outline set (another manifest's water-stones scene) the scene's artwork
// keeps that scene's alpha: its paint under the other artwork's outline
// (waterRocksOutline).
func waterStonesExports(t *testing.T, root, manifest, outline string) map[int]waterStonesExport {
	t.Helper()
	in, err := loadCompoundInputs(filepath.Join(root, manifest))
	if err != nil {
		t.Fatal(err)
	}
	if outline != "" {
		o, err := loadCompoundInputs(filepath.Join(root, outline))
		if err != nil {
			t.Fatal(err)
		}
		src := in.Artwork["water-stones"]
		if src.Image, err = waterRocksOutline(src.Image, o.Artwork["water-stones"].Image); err != nil {
			t.Fatal(err)
		}
		in.Artwork["water-stones"] = src
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	out := map[int]waterStonesExport{}
	for _, d := range []int{2, 4} {
		art := b.pixels[fmt.Sprintf("water-stones-artwork-%dx.png", d)]
		if art == nil {
			t.Fatalf("%s: no water-stones artwork at %dx", manifest, d)
		}
		rocks := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		for y := range 80 * d {
			for x := range 96 * d {
				if connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}) {
					rocks.SetNRGBA(x, y, art.NRGBAAt(x, y))
				}
			}
		}
		name := fmt.Sprintf("tiles/waterStones-%dx.png", d)
		e := waterStonesExport{rocks: rocks, stones: b.pixels[name]}
		for _, a := range b.artifacts {
			if a.name == name {
				e.stonesPNG = a.data
			}
		}
		if e.stones == nil || e.stonesPNG == nil {
			t.Fatalf("%s: no %s", manifest, name)
		}
		out[d] = e
	}
	return out
}

// waterRocksOutline returns paint under the alpha of outline: the winter edit
// of the lake rocks keeps the summer rocks' exact outline and their translucent
// ripple collar, which the edit does not reproduce to the pixel. Where the
// edit is nearly transparent its colour is unreliable; there the collar takes
// the summer colour as pale grey ice (its luminance, slightly cool).
func waterRocksOutline(paint, outline *image.NRGBA) (*image.NRGBA, error) {
	if paint.Rect.Size() != outline.Rect.Size() {
		return nil, fmt.Errorf("winter rocks %v and summer rocks %v differ in size", paint.Rect.Size(), outline.Rect.Size())
	}
	out := image.NewNRGBA(image.Rect(0, 0, paint.Rect.Dx(), paint.Rect.Dy()))
	for y := range out.Rect.Dy() {
		for x := range out.Rect.Dx() {
			p := paint.NRGBAAt(paint.Rect.Min.X+x, paint.Rect.Min.Y+y)
			o := outline.NRGBAAt(outline.Rect.Min.X+x, outline.Rect.Min.Y+y)
			if o.A == 0 {
				continue
			}
			w := min(1, float64(p.A)/64)
			grey := .299*float64(o.R) + .587*float64(o.G) + .114*float64(o.B)
			mix := func(v uint8, ice float64) uint8 {
				return uint8(math.Round(min(255, w*float64(v)+(1-w)*ice)))
			}
			out.SetNRGBA(x, y, color.NRGBA{mix(p.R, grey), mix(p.G, grey+3), mix(p.B, grey+6), o.A})
		}
	}
	return out, nil
}

func TestGenerateWaterRocks(t *testing.T) {
	root := filepath.Join("..", "..")
	summer := waterStonesExports(t, root, waterStonesSummerManifest, "")
	winter := waterStonesExports(t, root, waterStonesWinterManifest, waterStonesSummerManifest)
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{2, 4} {
		// The rocks over the installed water are the installed stones tile, in
		// summer, and the winter scene's stones tile in winter.
		water, stones := installedTile(t, root, exported, "tile/water", d), installedTile(t, root, exported, "tile/waterStones", d)
		if got := compoundCompose(water, summer[d].rocks); !bytes.Equal(got.Pix, stones.Pix) {
			t.Fatalf("%dx: the rocks over tile/water differ from tile/waterStones", d)
		}
		if got := compoundCompose(installedTile(t, root, exported, "tile/water_w", d), winter[d].rocks); !bytes.Equal(got.Pix, winter[d].stones.Pix) {
			t.Fatalf("%dx: the winter rocks over tile/water_w differ from the winter stones tile", d)
		}
		for _, id := range []string{"waterRocks", "waterRocks_w"} {
			rocks := summer[d].rocks
			if id == "waterRocks_w" {
				rocks = winter[d].rocks
			}
			var buf bytes.Buffer
			if err := png.Encode(&buf, rocks); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(buf.Bytes())
			path := fmt.Sprintf("terrain/water/%s-%dx.png", id, d)
			if err := os.WriteFile(filepath.Join(root, "assets", path), buf.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/water"})
			if err != nil {
				t.Fatal(err)
			}
			entries = setCampaignEntry(entries, raw)
		}
	}
	writeCampaignIndex(t, root, entries)
}

// setCampaignEntry replaces the entry of raw's id and density in place, or
// appends it.
func setCampaignEntry(entries []json.RawMessage, raw json.RawMessage) []json.RawMessage {
	var want campaignEntry
	if json.Unmarshal(raw, &want) != nil {
		return append(entries, raw)
	}
	i := slices.IndexFunc(entries, func(r json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(r, &e) == nil && e.ID == want.ID && e.Density == want.Density
	})
	if i < 0 {
		return append(entries, raw)
	}
	entries[i] = raw
	return entries
}

// writeCampaignIndex writes the runtime index, one entry per line.
func writeCampaignIndex(t *testing.T, root string, entries []json.RawMessage) {
	t.Helper()
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func installedTile(t *testing.T, root string, exported map[string]campaignEntry, id string, d int) *image.NRGBA {
	t.Helper()
	e, ok := exported[campaignExportKey(id, d)]
	if !ok {
		t.Fatalf("%s at %dx is not installed", id, d)
	}
	f, err := os.Open(filepath.Join(root, "assets", e.Path))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	out := image.NewNRGBA(im.Bounds())
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			out.Set(x, y, im.At(x, y))
		}
	}
	return out
}
