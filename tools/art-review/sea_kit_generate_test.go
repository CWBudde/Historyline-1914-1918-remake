//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateSeaKit -count=1 -v
// Cuts the shoals and skerries of the C2 kit, art/terrain/water/
// water-seakit-shared-v1.png (12 px per world px), into hex overlays:
// tile/shoal1..4, the two sandy beds and their mirror images at the kit's
// scale, drawn over shallow open water by a sea; and tile/skerry1..3, two of
// the three rock clusters each at half the kit's scale, the rocks of stones in
// water by a sea. The winter forms (suffix _w) are the same pixels under the
// winter water grade. Writes them to assets/terrain/water/, replacing earlier
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
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// seaKitPixels is the kit's scale, source pixels per world pixel.
const seaKitPixels = 12

// The kit's elements: rough boxes in the painting that hold one each.
var (
	seaKitShoals   = []image.Rectangle{image.Rect(820, 110, 1150, 680), image.Rect(1160, 110, 1520, 680)}
	seaKitSkerries = []image.Rectangle{image.Rect(100, 680, 520, 1000), image.Rect(600, 680, 1000, 1000), image.Rect(1020, 680, 1500, 1000)}
)

// seaKitPart is one placed element of an overlay: which element, where its
// centre lies from the hex centre in world px, at what scale of the kit and
// whether mirrored.
type seaKitPart struct {
	element int
	at      [2]float64
	scale   float64
	mirror  bool
}

func TestGenerateSeaKit(t *testing.T) {
	root := filepath.Join("..", "..")
	f, err := os.Open(filepath.Join(root, "art/terrain/water/water-seakit-shared-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	kit := image.NewNRGBA(src.Bounds())
	draw.Draw(kit, kit.Rect, src, src.Bounds().Min, draw.Src)
	shoals, skerries := seaKitElements(t, kit, seaKitShoals), seaKitElements(t, kit, seaKitSkerries)
	overlays := map[string][]seaKitPart{
		"shoal1":  {{0, [2]float64{-6, 2}, 1, false}},
		"shoal2":  {{1, [2]float64{5, -2}, 1, false}},
		"shoal3":  {{0, [2]float64{6, -3}, 1, true}},
		"shoal4":  {{1, [2]float64{-5, 3}, 1, true}},
		"skerry1": {{0, [2]float64{-13, -7}, 0.7, false}, {1, [2]float64{12, 9}, 0.7, false}},
		"skerry2": {{1, [2]float64{-11, 8}, 0.7, true}, {2, [2]float64{13, -6}, 0.7, false}},
		"skerry3": {{2, [2]float64{-12, -5}, 0.7, true}, {0, [2]float64{11, 10}, 0.7, true}},
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && (strings.HasPrefix(e.ID, "tile/shoal") || strings.HasPrefix(e.ID, "tile/skerry"))
	})
	ids := make([]string, 0, len(overlays))
	for id := range overlays {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		elements := shoals
		if strings.HasPrefix(id, "skerry") {
			elements = skerries
		}
		for _, d := range []int{2, 4} {
			tile := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
			for _, p := range overlays[id] {
				e := elements[p.element]
				w := int(math.Round(float64(e.Rect.Dx()) * p.scale * float64(d) / seaKitPixels))
				h := int(math.Round(float64(e.Rect.Dy()) * p.scale * float64(d) / seaKitPixels))
				part, err := resizeLanczos(e, w, h)
				if err != nil {
					t.Fatal(err)
				}
				if p.mirror {
					part = mirrored(part)
				}
				at := image.Pt(int(math.Round((48+p.at[0])*float64(d)))-w/2, int(math.Round((40+p.at[1])*float64(d)))-h/2)
				draw.Draw(tile, part.Rect.Add(at), part, image.Point{}, draw.Over)
			}
			for y := range 80 * d {
				for x := range 96 * d {
					if !connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}) {
						tile.SetNRGBA(x, y, color.NRGBA{})
					}
				}
			}
			var buf bytes.Buffer
			if err := png.Encode(&buf, tile); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(buf.Bytes())
			for _, name := range []string{id, id + "_w"} {
				path := fmt.Sprintf("terrain/water/%s-%dx.png", name, d)
				if err := os.WriteFile(filepath.Join(root, "assets", path), buf.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(campaignEntry{ID: "tile/" + name, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/water"})
				if err != nil {
					t.Fatal(err)
				}
				entries = append(entries, raw)
			}
		}
	}
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

// seaKitElements cuts each box of kit to the element it holds: the box of its
// pixels of alpha 24 and more, with the faint haze around them cleared.
func seaKitElements(t *testing.T, kit *image.NRGBA, boxes []image.Rectangle) []*image.NRGBA {
	t.Helper()
	var out []*image.NRGBA
	for _, box := range boxes {
		var r image.Rectangle
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				if kit.NRGBAAt(x, y).A >= 24 {
					r = r.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
		if r.Empty() {
			t.Fatalf("no element in %v", box)
		}
		if r.Min.X == box.Min.X || r.Min.Y == box.Min.Y || r.Max.X == box.Max.X || r.Max.Y == box.Max.Y {
			t.Fatalf("the element in %v reaches its box (%v)", box, r)
		}
		e := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
		draw.Draw(e, e.Rect, kit, r.Min, draw.Src)
		for i := 3; i < len(e.Pix); i += 4 {
			if e.Pix[i] < 12 {
				e.Pix[i-3], e.Pix[i-2], e.Pix[i-1], e.Pix[i] = 0, 0, 0, 0
			}
		}
		out = append(out, e)
	}
	return out
}

func mirrored(im *image.NRGBA) *image.NRGBA {
	out := image.NewNRGBA(im.Rect)
	w := im.Rect.Dx()
	for y := range im.Rect.Dy() {
		for x := range w {
			out.SetNRGBA(w-1-x, y, im.NRGBAAt(x, y))
		}
	}
	return out
}
