//go:build bridgegfx

package main

// Explicit offline generator, run from the repository root's tools/terrain-stamps:
//
//	go test -tags bridgegfx ./tools/terrain-stamps -run TestGenerateBridgeDecks -count=1 -v
//
// Regrades the rail heads of the painted rail decks (v1) onto tile/rail36's dark
// steel as the v2 masters (raildeck_test.go), then registers the bridge deck
// masters of art/pipeline/bridge-standing-v2.json (the rail deck and the winter
// decks, each painted as an edit of a registered v1 export) onto the 1152x960
// export frame of the export they were made from, writes the -export- masters
// and art/terrain/<family>/<family>-registration-v<registrationVersion>.json,
// pins their hashes in the manifest, exports the standing items with run() and
// puts them into assets/ and assets/index.json. Run `go run ./cmd/hlart index`
// afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// deckSource is one native deck master and the registered export whose frame
// it takes. rails, when set, is the native x of the middle between a vertical
// deck's two rails: the deck then moves sideways so that its rails continue
// tile/rail36's, whose middle lies at logical x 47.875 (export x 574.5).
type deckSource struct {
	id, native, frame, export string
	rails                     float64
}

// registrationVersion is the version of each family's registration record: the
// rBridge record names the v2 rail deck masters.
var registrationVersion = map[string]int{"rBridge": 3, "bridge": 2}

// rail36Middle is the middle between tile/rail36's rails in export pixels.
const rail36Middle = 47.875 * 12

var deckSources = []deckSource{
	{"standing/rBridge36-rail", "art/terrain/rBridge/rbridge-36-rail-summer-v2.png", "art/terrain/rBridge/rbridge-36-export-shared-v1.png", "art/terrain/rBridge/rbridge-36-rail-export-summer-v2.png", 686},
	{"standing/rBridge36-rail_w", "art/terrain/rBridge/rbridge-36-rail-winter-v2.png", "art/terrain/rBridge/rbridge-36-rail-export-summer-v2.png", "art/terrain/rBridge/rbridge-36-rail-export-winter-v2.png", 688},
	{"standing/rBridge36_w", "art/terrain/rBridge/rbridge-36-winter-v1.png", "art/terrain/rBridge/rbridge-36-export-shared-v1.png", "art/terrain/rBridge/rbridge-36-export-winter-v1.png", 0},
	{"standing/rBridge14_w", "art/terrain/rBridge/rbridge-14-winter-v1.png", "art/terrain/rBridge/rbridge-14-export-shared-v1.png", "art/terrain/rBridge/rbridge-14-export-winter-v1.png", 0},
	{"standing/rBridge25_w", "art/terrain/rBridge/rbridge-25-winter-v1.png", "art/terrain/rBridge/rbridge-25-export-shared-v1.png", "art/terrain/rBridge/rbridge-25-export-winter-v1.png", 0},
	{"standing/bridge36_w", "art/terrain/bridge/bridge-36-winter-v1.png", "art/terrain/bridge/bridge-36-export-shared-v1.png", "art/terrain/bridge/bridge-36-export-winter-v1.png", 0},
	{"standing/bridge25_w", "art/terrain/bridge/bridge-25-winter-v1.png", "art/terrain/bridge/bridge-25-export-shared-v1.png", "art/terrain/bridge/bridge-25-export-winter-v1.png", 0},
}

func TestGenerateBridgeDecks(t *testing.T) {
	root := filepath.Join("..", "..")
	type registration struct {
		ID           string     `json:"id"`
		Source       string     `json:"source"`
		SourceSHA256 string     `json:"source_sha256"`
		NativeDims   [2]int     `json:"native_dimensions"`
		Frame        string     `json:"frame"`
		SourceBox    [4]int     `json:"source_opaque_box"`
		FrameBox     [4]int     `json:"frame_opaque_box"`
		Affine       [6]float64 `json:"source_to_prepared_affine"`
		Prepared     string     `json:"prepared_source"`
		PreparedSHA  string     `json:"prepared_sha256"`
		Policy       string     `json:"policy"`
	}
	for _, r := range railDeckRegrades {
		from := readNRGBA(t, filepath.Join(root, r.from))
		im := regradeRailHeads(from, readNRGBA(t, filepath.Join(root, r.track)), r.rails)
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, r.to), b.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s: rail heads %v regraded onto %s", r.to, railHeadColumns(from, r.rails), r.track)
	}
	regs := map[string][]registration{}
	hashes := map[string]string{}
	for _, d := range deckSources {
		native := readNRGBA(t, filepath.Join(root, d.native))
		frame := readNRGBA(t, filepath.Join(root, d.frame))
		sb, fb := opaqueBox(native), opaqueBox(frame)
		sx := float64(fb.Dx()) / float64(sb.Dx())
		sy := float64(fb.Dy()) / float64(sb.Dy())
		aff := f64.Aff3{sx, 0, float64(fb.Min.X) - float64(sb.Min.X)*sx, 0, sy, float64(fb.Min.Y) - float64(sb.Min.Y)*sy}
		if d.rails != 0 {
			aff[2] = rail36Middle - d.rails*sx
		}
		dst := image.NewRGBA(frame.Bounds())
		draw.CatmullRom.Transform(dst, aff, native, native.Bounds(), draw.Src, nil)
		out := image.NewNRGBA(dst.Bounds())
		draw.Draw(out, out.Bounds(), dst, image.Point{}, draw.Src)
		var b bytes.Buffer
		if err := png.Encode(&b, out); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, d.export), b.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		hashes[d.export] = sum(b.Bytes())
		raw, err := os.ReadFile(filepath.Join(root, d.native))
		if err != nil {
			t.Fatal(err)
		}
		family := filepath.Dir(d.export)
		regs[family] = append(regs[family], registration{
			ID: d.id, Source: d.native, SourceSHA256: sum(raw), NativeDims: [2]int{native.Bounds().Dx(), native.Bounds().Dy()},
			Frame: d.frame, SourceBox: box(sb), FrameBox: box(fb), Affine: aff,
			Prepared: d.export, PreparedSHA: hashes[d.export],
			Policy: "Edit of the frame export; its opaque box (alpha > 128) scaled per axis and moved onto the frame's opaque box, CatmullRom at 12 pixels per logical pixel; final terrain-stamps Lanczos3 at 4x then 2x. Fixed canvas 96x80, anchor 48,40.",
		})
		t.Logf("%s: box %v -> %v, scale %.4f x %.4f", d.id, sb, fb, sx, sy)
	}
	for family, r := range regs {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		v, ok := registrationVersion[filepath.Base(family)]
		if !ok {
			t.Fatalf("%s: no registration version", family)
		}
		name := fmt.Sprintf("%s-registration-v%d.json", strings.ToLower(filepath.Base(family)), v)
		if err := os.WriteFile(filepath.Join(root, family, name), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Pin the exports in the manifest and export the standing items.
	manifestPath := filepath.Join(root, "art/pipeline/bridge-standing-v2.json")
	var m manifest
	readJSONFile(t, manifestPath, &m)
	for i := range m.Assets {
		h, ok := hashes[m.Assets[i].Source]
		if !ok {
			t.Fatalf("%s: no registered source %s", m.Assets[i].ID, m.Assets[i].Source)
		}
		m.Assets[i].SourceSHA256 = h
	}
	mb, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(mb, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "build/scratch/bridge-standing-v2")
	if err := os.RemoveAll(out); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	err = run([]string{"-manifest", "art/pipeline/bridge-standing-v2.json", "-out", "build/scratch/bridge-standing-v2"}, io.Discard)
	if cerr := os.Chdir(wd); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		t.Fatal(err)
	}
	var entries []indexEntry
	readJSONFile(t, filepath.Join(out, "index-entries.json"), &entries)
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(out, e.Path))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "assets", e.Path), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mergeIndex(t, filepath.Join(root, "assets/index.json"), entries)
}

// mergeIndex replaces the index entries of the exported IDs and inserts new
// ones after the v1 bridge standing items.
func mergeIndex(t *testing.T, path string, entries []indexEntry) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Schema int               `json:"schema_version"`
		Assets []json.RawMessage `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, e := range entries {
		ids[e.ID] = true
	}
	idOf := func(raw json.RawMessage) string {
		var e struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		return e.ID
	}
	assets := slices.DeleteFunc(index.Assets, func(raw json.RawMessage) bool { return ids[idOf(raw)] })
	at := len(assets)
	for i, raw := range assets {
		if id := idOf(raw); strings.HasPrefix(id, "standing/rBridge") || strings.HasPrefix(id, "standing/bridge") {
			at = i + 1
		}
	}
	var add []json.RawMessage
	for _, e := range entries {
		raw, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		add = append(add, raw)
	}
	assets = slices.Insert(assets, at, add...)
	var out bytes.Buffer
	out.WriteString("{\n  \"schema_version\": 1,\n  \"assets\": [\n")
	for i, e := range assets {
		out.WriteString("    ")
		out.Write(e)
		if i+1 < len(assets) {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString("  ]\n}\n")
	if err := os.WriteFile(path, out.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func opaqueBox(im *image.NRGBA) image.Rectangle {
	r := image.Rectangle{}
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A > 128 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}

func box(r image.Rectangle) [4]int { return [4]int{r.Min.X, r.Min.Y, r.Dx(), r.Dy()} }

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func readNRGBA(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(src.Bounds())
	draw.Draw(im, im.Bounds(), src, src.Bounds().Min, draw.Src)
	return im
}

func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(fmt.Errorf("%s: %w", path, err))
	}
}
