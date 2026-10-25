//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateSeaKit -count=1 -v
// Cuts the shoals and skerries of the C2 kit, art/terrain/water/
// water-seakit-shared-v1.png (12 px per world px), into hex overlays:
// tile/shoal1..4, the two sandy beds and their mirror images at the kit's
// scale, drawn over shallow open water by a sea; and tile/skerry1..3, two of
// the three rock clusters each at 0.7 of the kit's scale, the rocks of stones
// in water by a sea. The winter shoals (suffix _w) are the same pixels under
// the winter water grade; the winter skerries are cut the same way from the
// kit's winter edit, water-seakit-winter-v1.png (snow caps, frosted weed, ice
// collars), its paint under the summer kit's alpha (waterRocksOutline), so they
// keep the summer skerries' outline. Writes them to assets/terrain/water/,
// replacing earlier exports, and lists them in assets/index.json; run
// `go run ./cmd/hlart index` afterwards.
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
	kit := loadSeaKit(t, filepath.Join(root, "art/terrain/water/water-seakit-shared-v1.png"))
	winterKit, err := waterRocksOutline(loadSeaKit(t, filepath.Join(root, "art/terrain/water/water-seakit-winter-v1.png")), kit)
	if err != nil {
		t.Fatal(err)
	}
	shoals, skerries := seaKitElements(t, kit, seaKitShoals), seaKitElements(t, kit, seaKitSkerries)
	winterSkerries := seaKitElements(t, winterKit, seaKitSkerries)
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
	// The new entries take the place of the earlier ones (the end for a first
	// export), so a re-export leaves the index's order alone.
	ours := func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && (strings.HasPrefix(e.ID, "tile/shoal") || strings.HasPrefix(e.ID, "tile/skerry"))
	}
	place := slices.IndexFunc(entries, ours)
	entries = slices.DeleteFunc(entries, ours)
	if place < 0 {
		place = len(entries)
	}
	var fresh []json.RawMessage
	ids := make([]string, 0, len(overlays))
	for id := range overlays {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		elements, winterElements := shoals, shoals
		if strings.HasPrefix(id, "skerry") {
			elements, winterElements = skerries, winterSkerries
		}
		for _, d := range []int{2, 4} {
			summer := seaKitTile(t, elements, overlays[id], d)
			winter := summer
			if strings.HasPrefix(id, "skerry") {
				winter = seaKitTile(t, winterElements, overlays[id], d)
			}
			for _, o := range []struct {
				name string
				png  []byte
			}{{id, summer}, {id + "_w", winter}} {
				path := fmt.Sprintf("terrain/water/%s-%dx.png", o.name, d)
				if err := os.WriteFile(filepath.Join(root, "assets", path), o.png, 0644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(o.png)
				raw, err := json.Marshal(campaignEntry{ID: "tile/" + o.name, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/water"})
				if err != nil {
					t.Fatal(err)
				}
				fresh = append(fresh, raw)
			}
		}
	}
	entries = slices.Insert(entries, place, fresh...)
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

// loadSeaKit reads a sea kit painting as NRGBA.
func loadSeaKit(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	kit := image.NewNRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(kit, kit.Rect, src, src.Bounds().Min, draw.Src)
	return kit
}

// seaKitTile places parts of elements on a hex overlay at density d, clipped
// to the hex, and returns its PNG bytes.
func seaKitTile(t *testing.T, elements []*image.NRGBA, parts []seaKitPart, d int) []byte {
	t.Helper()
	tile := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
	for _, p := range parts {
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
	return buf.Bytes()
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
