package maprender

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

func materialFS(t *testing.T) (fstest.MapFS, []Entry) {
	t.Helper()
	files := fstest.MapFS{}
	var entries []Entry
	for _, d := range []int{2, 4} {
		for _, id := range []string{worldmaterial.MeadowID, worldmaterial.ScrubID} {
			im := image.NewNRGBA(image.Rect(0, 0, 568*d, 400*d))
			for y := 0; y < im.Bounds().Dy(); y++ {
				for x := 0; x < im.Bounds().Dx(); x++ {
					im.SetNRGBA(x, y, color.NRGBA{R: uint8((x / d) % 97), G: 80, A: 255})
				}
			}
			// Match opposite endpoints while retaining a coordinate-dependent interior.
			for y := 0; y < im.Bounds().Dy(); y++ {
				im.SetNRGBA(im.Bounds().Max.X-1, y, im.NRGBAAt(0, y))
			}
			var b bytes.Buffer
			if err := png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			path := fmt.Sprintf("%d-%d.png", d, len(entries))
			files[path] = &fstest.MapFile{Data: append([]byte(nil), b.Bytes()...)}
			entries = append(entries, Entry{ID: id, Pack: fmt.Sprintf("terrain-materials-%dx", d), Density: d, Path: path, SHA: fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), Width: 568 * d, Height: 400 * d, Layer: "ground"})
		}
	}
	return files, entries
}

func indexedMaterialFS(t *testing.T, files fstest.MapFS, entries []Entry) fstest.MapFS {
	t.Helper()
	out := fstest.MapFS{}
	for k, v := range files {
		out[k] = v
	}
	b, err := json.Marshal(struct {
		Schema int     `json:"schema_version"`
		Assets []Entry `json:"assets"`
	}{1, entries})
	if err != nil {
		t.Fatal(err)
	}
	out["index.json"] = &fstest.MapFile{Data: b}
	return out
}

func TestMaterialIndexValidation(t *testing.T) {
	files, entries := materialFS(t)
	for _, tt := range []struct {
		name   string
		change func([]Entry) []Entry
		bad    bool
	}{
		{"valid", func(e []Entry) []Entry { return e }, false},
		{"partial", func(e []Entry) []Entry { return e[:3] }, true},
		{"duplicate", func(e []Entry) []Entry { return append(e, e[0]) }, true},
		{"unknown", func(e []Entry) []Entry { e[0].ID = "material/unknown"; return e }, true},
		{"pack", func(e []Entry) []Entry { e[0].Pack = "terrain-grass-2x"; return e }, true},
		{"size", func(e []Entry) []Entry { e[0].Width++; return e }, true},
		{"layer", func(e []Entry) []Entry { e[0].Layer = "standing"; return e }, true},
		{"anchor", func(e []Entry) []Entry { e[0].Anchor[0] = 1; return e }, true},
		{"hash", func(e []Entry) []Entry { e[0].SHA = "bad"; return e }, true},
		{"cached path with different hash", func(e []Entry) []Entry { e[1].Path = e[0].Path; e[1].SHA = "bad"; return e }, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(indexedMaterialFS(t, files, tt.change(append([]Entry(nil), entries...))))
			if (err != nil) != tt.bad {
				t.Fatalf("error=%v, want bad %v", err, tt.bad)
			}
		})
	}
}

func TestMaterialDrawUsesGlobalOriginAndMask(t *testing.T) {
	files, entries := materialFS(t)
	r, err := New(indexedMaterialFS(t, files, entries))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{2, 4} {
		for _, at := range []hexgrid.Coord{{Line: 0, Col: 1}, {Line: 1, Col: 0}} {
			mask := image.NewNRGBA(image.Rect(0, 0, 5, 3))
			mask.SetNRGBA(1, 1, color.NRGBA{R: 255, A: 128})
			id := "test-mask"
			path := "mask"
			r.files[fmt.Sprintf("%d/%s", d, id)] = Entry{Path: path, Anchor: [2]float64{48, 40}}
			r.cache[path] = mask
			dst := image.NewNRGBA(image.Rect(0, 0, 400*d, 200*d))
			if err := r.drawMaterial(dst, id, at, d, mapmaterial.New(nil)); err != nil {
				t.Fatal(err)
			}
			x, y := hexgrid.Center(at)
			origin := image.Pt(int((x-48)*float64(d)), int((y-40)*float64(d)))
			expected, err := r.materials[d].Tile(mask, origin, nil)
			if err != nil {
				t.Fatal(err)
			}
			for yy := 0; yy < 3; yy++ {
				for xx := 0; xx < 5; xx++ {
					if got, want := dst.NRGBAAt(origin.X+xx, origin.Y+yy), expected.NRGBAAt(xx, yy); got != want {
						t.Fatalf("%dx at %v pixel %d,%d: %v != %v", d, at, xx, yy, got, want)
					}
				}
			}
		}
	}
}

func TestMaterialImageValidation(t *testing.T) {
	files, entries := materialFS(t)
	for _, name := range []string{"alpha", "period", "decoded-size"} {
		t.Run(name, func(t *testing.T) {
			im := image.NewNRGBA(image.Rect(0, 0, 1136, 800))
			for y := 0; y < 800; y++ {
				for x := 0; x < 1136; x++ {
					im.SetNRGBA(x, y, color.NRGBA{R: 70, A: 255})
				}
			}
			switch name {
			case "alpha":
				im.SetNRGBA(5, 5, color.NRGBA{A: 128})
			case "period": // a ramp that jumps back at the wrap
				for y := 0; y < 800; y++ {
					for x := 0; x < 1136; x++ {
						im.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 200 / 1135), A: 255})
					}
				}
			case "decoded-size":
				im = image.NewNRGBA(image.Rect(0, 0, 1, 1))
			}
			var b bytes.Buffer
			if err := png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			es := append([]Entry(nil), entries...)
			es[0].SHA = fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
			fsys := indexedMaterialFS(t, files, es)
			fsys[es[0].Path] = &fstest.MapFile{Data: b.Bytes()}
			if _, err := New(fsys); err == nil {
				t.Fatal("accepted invalid material image")
			}
		})
	}
}

func TestMaterialRenderRealScene(t *testing.T) {
	files, entries := materialFS(t)
	base := os.DirFS("../../assets")
	raw, err := fs.ReadFile(base, "index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []Entry `json:"assets"`
	}
	if err = json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	var baseEntries []Entry
	for _, e := range index.Assets {
		if !strings.HasPrefix(e.ID, "material/") && !strings.HasPrefix(e.Pack, "terrain-materials") {
			baseEntries = append(baseEntries, e)
		}
	}
	files = indexedMaterialFS(t, files, append(baseEntries, entries...))
	r, err := New(overlay{over: files, base: base})
	if err != nil {
		t.Fatal(err)
	}
	lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "grass"}, {Terrain: "grass"}}, {{Terrain: "grass"}}, {{Terrain: "grass"}, {Terrain: "grass"}}}}
	for _, d := range []int{2, 4} {
		m, err := r.Render(lvl, d, false)
		if err != nil {
			t.Fatal(err)
		}
		for _, at := range []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 1, Col: 0}} {
			x, y := hexgrid.Center(at)
			px, py := int(x)*d, int(y)*d
			want := color.NRGBA{R: uint8((px / d) % 97), G: 80, A: 255}
			if got := m.Image.NRGBAAt(px, py); got != want {
				t.Fatalf("%dx at %v: got %v want %v", d, at, got, want)
			}
		}
	}
}
