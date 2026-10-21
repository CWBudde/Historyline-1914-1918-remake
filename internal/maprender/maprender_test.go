package maprender

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Catch odd-line shifts and anchors drifting between authored densities.
func TestPlace(t *testing.T) {
	for _, density := range []int{1, 2, 4} {
		src := image.NewNRGBA(image.Rect(0, 0, 96*density, 80*density))
		src.SetNRGBA(48*density, 40*density, color.NRGBA{R: 255, A: 255})
		dst := image.NewNRGBA(image.Rect(0, 0, 300*density, 200*density))
		Place(dst, src, hexgrid.Coord{Line: 1, Col: 0}, density, [2]float64{48, 40})
		if got := dst.NRGBAAt(119*density, 80*density); got.R != 255 {
			t.Errorf("%dx odd-line anchor: %+v", density, got)
		}
		clear(dst.Pix)
		Place(dst, src, hexgrid.Coord{Line: 1, Col: 0}, density, [2]float64{48, 49})
		if got := dst.NRGBAAt(119*density, 71*density); got.R != 255 {
			t.Errorf("%dx unit anchor: %+v", density, got)
		}
	}
}

func TestRenderPulse(t *testing.T) {
	r, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	details := 0
	for id := range r.has {
		if strings.HasPrefix(id, "tile/grass-detail-") {
			details++
		}
	}
	if details != 16 {
		t.Errorf("%d grass details at both densities, want sixteen", details)
	}
	lvl, err := level.Load("../../levels/level1_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := r.Render(lvl, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	w, h := Size(lvl.Grid())
	if b := m.Image.Bounds(); b.Dx() != 2*w || b.Dy() != 2*h {
		t.Errorf("size %v, want %dx%d", b, 2*w, 2*h)
	}
	if m.Units == 0 {
		t.Error("no units drawn")
	}
}

// overlay serves files from over before base.
type overlay struct {
	over fstest.MapFS
	base fs.FS
}

func (o overlay) Open(name string) (fs.File, error) {
	if _, ok := o.over[name]; ok {
		return o.over.Open(name)
	}
	return o.base.Open(name)
}

var red = color.NRGBA{R: 255, A: 255}

// extra is a solid red asset of w×h logical pixels added to the index.
type extra struct {
	id, pack string // pack without its -<d>x suffix
	w, h     int
	anchor   [2]float64
}

// withAssets keeps the repository's required art, replaces optional standing and
// stamp art with the extras, and supplies both densities for every extra.
func withAssets(t *testing.T, extras ...extra) fs.FS {
	t.Helper()
	base := os.DirFS("../../assets")
	raw, err := fs.ReadFile(base, "index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index map[string]any
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	var retained []any
	for _, item := range index["assets"].([]any) {
		entry := item.(map[string]any)
		pack, _ := entry["pack"].(string)
		if !strings.HasPrefix(pack, "terrain-standing-") && !strings.HasPrefix(pack, "terrain-stamps-") {
			retained = append(retained, item)
		}
	}
	index["assets"] = retained
	over := fstest.MapFS{}
	for i, x := range extras {
		for _, d := range []int{2, 4} {
			im := image.NewNRGBA(image.Rect(0, 0, x.w*d, x.h*d))
			draw.Draw(im, im.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)
			var b bytes.Buffer
			if err := png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			p := fmt.Sprintf("test/%d-%dx.png", i, d)
			over[p] = &fstest.MapFile{Data: b.Bytes()}
			index["assets"] = append(index["assets"].([]any), map[string]any{
				"id": x.id, "pack": fmt.Sprintf("%s-%dx", x.pack, d), "density": d, "path": p,
				"sha256": fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), "logical_anchor": x.anchor,
			})
		}
	}
	if raw, err = json.Marshal(index); err != nil {
		t.Fatal(err)
	}
	over["index.json"] = &fstest.MapFile{Data: raw}
	return overlay{over, base}
}

func renderPulse(t *testing.T, assets fs.FS) (*level.Level, Map) {
	t.Helper()
	r, err := New(assets)
	if err != nil {
		t.Fatal(err)
	}
	lvl, err := level.Load("../../levels/level1_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := r.Render(lvl, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	return lvl, m
}

// A standing item overhangs its hex over the ground of its neighbours, and is
// clipped at the map edge.
func TestStandingOverhang(t *testing.T) {
	lvl, m := renderPulse(t, withAssets(t, extra{"standing/house", "terrain-standing", 144, 120, [2]float64{72, 60}}))
	houses := 0
	for line, hs := range lvl.Lines {
		for col, h := range hs {
			if h.Terrain != "house" {
				continue
			}
			houses++
			x, y := hexgrid.Center(hexgrid.Coord{Line: line, Col: col})
			// 60 logical pixels west of the centre lies in a neighbour's ground.
			if got := m.Image.NRGBAAt(int(x-60)*2, int(y)*2); got != red {
				t.Errorf("house (%d,%d): overhang %+v, want it over the neighbour", line, col, got)
			}
		}
	}
	if houses == 0 {
		t.Fatal("level has no house")
	}
	// At the map edge the overhang is clipped.
	src := image.NewNRGBA(image.Rect(0, 0, 144, 120))
	draw.Draw(src, src.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)
	dst := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	Place(dst, src, hexgrid.Coord{}, 1, [2]float64{72, 60})
	if dst.NRGBAAt(0, 0) != red || dst.NRGBAAt(48+72-1, 40+60-1) != red || dst.NRGBAAt(48+72, 40) == red {
		t.Error("overhang at the map edge drawn wrongly")
	}
}

// A cluster with a stamp is drawn as the stamp, anchored on its first hex.
func TestClusterStamp(t *testing.T) {
	const shape = "4_-1.1_1.1_0.2" // PULSE's two craters
	lvl, m := renderPulse(t, withAssets(t, extra{"stamp/crater/" + shape + "/1", "terrain-stamps", 8, 8, [2]float64{4, 4}}))
	stamped := 0
	for line, hs := range lvl.Lines {
		for col, h := range hs {
			if h.Terrain != "mud" {
				continue
			}
			x, y := hexgrid.Center(hexgrid.Coord{Line: line, Col: col})
			if m.Image.NRGBAAt(int(x)*2, int(y)*2) == red {
				stamped++
			}
		}
	}
	// One stamp per crater, on its first (northernmost) hex.
	if stamped != 2 {
		t.Errorf("%d crater hexes carry the stamp, want 2", stamped)
	}
}
