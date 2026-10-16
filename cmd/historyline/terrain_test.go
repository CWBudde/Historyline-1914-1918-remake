package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

func testLevel(t *testing.T) *level.Level {
	t.Helper()
	l, err := loadLevel("levels-not-here", "level1_cp")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// terrainCatalog is a terrain catalog of transparent tiles holding every tile
// level l draws, minus skip, plus the context IDs in contexts.
func terrainCatalog(t *testing.T, l *level.Level, skip string, contexts ...string) fstest.MapFS {
	t.Helper()
	s := mapcompose.Compose(l.Grid(), levelHex(l), func(string) bool { return false })
	ids := map[string]string{mapcompose.TileID("unseen"): "backgrounds"}
	for _, line := range s.Cells {
		for _, c := range line {
			ids[c.Background] = "backgrounds"
			for _, o := range c.Overlays {
				ids[o] = "infrastructure"
			}
		}
	}
	// Every context pack holds at least one tile, as the built catalog does.
	for _, f := range mapcompose.Families {
		contexts = append(contexts, mapcompose.ContextID(f, "unused"))
	}
	for _, id := range contexts {
		ids[id] = "context-" + strings.Split(id, "/")[1]
	}
	delete(ids, skip)
	fsys := fstest.MapFS{}
	for _, d := range spriteDensities {
		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(buf.Bytes())
		packs := map[string]*assets.Manifest{}
		for _, g := range terrainGroups {
			packs[g] = &assets.Manifest{Schema: 1, ID: fmt.Sprintf("terrain-%s-%dx", g, d), Status: "candidate"}
		}
		var sorted []string
		for id := range ids {
			sorted = append(sorted, id)
		}
		slices.Sort(sorted)
		for _, id := range sorted {
			m := packs[ids[id]]
			p := fmt.Sprintf("images/%04d.png", len(m.Entries))
			fsys[path.Join(fmt.Sprintf("%dx", d), ids[id], p)] = &fstest.MapFile{Data: buf.Bytes()}
			m.Entries = append(m.Entries, assets.Entry{ID: id, Path: p, SHA256: hex.EncodeToString(sum[:]),
				Width: 96 * d, Height: 80 * d, Density: d, Anchor: [2]float64{48, 40}})
		}
		for g, m := range packs {
			b, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			fsys[path.Join(fmt.Sprintf("%dx", d), g, "manifest.json")] = &fstest.MapFile{Data: b}
		}
	}
	return fsys
}

func TestOpenTerrainArtSelectsAvailableContexts(t *testing.T) {
	l := testLevel(t)
	plain := mapcompose.Compose(l.Grid(), levelHex(l), func(string) bool { return false })
	var gap mapcompose.Gap
	for _, x := range plain.Gaps {
		if x.Reason == mapcompose.Missing {
			gap = x
			break
		}
	}
	ctx := mapcompose.ContextID(gap.Family, mapcompose.Key(gap.Descriptor))
	a, err := openTerrainArt(terrainCatalog(t, l, "", ctx), assets.Options{AllowCandidates: true}, l)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.scene.Cells[gap.At.Line][gap.At.Col].Background; got != ctx {
		t.Errorf("gap hex draws %s, want the exported context %s", got, ctx)
	}
	// Every hex with the same descriptor is closed by the one export.
	want := 0
	for _, x := range plain.Gaps {
		if x.Descriptor != gap.Descriptor {
			want++
		}
	}
	if len(a.scene.Gaps) != want || want == len(plain.Gaps) || !a.candidate {
		t.Errorf("%d gaps (want %d of %d), candidate %v", len(a.scene.Gaps), want, len(plain.Gaps), a.candidate)
	}
}

func TestOpenTerrainArtRefusesCandidatesAndGaps(t *testing.T) {
	l := testLevel(t)
	if _, err := openTerrainArt(terrainCatalog(t, l, ""), assets.Options{}, l); err == nil || !strings.Contains(err.Error(), "opt-in") {
		t.Errorf("candidates loaded without opt-in: %v", err)
	}
	for _, skip := range []string{"tile/unseen", "tile/hqN", "tile/road36"} {
		if _, err := openTerrainArt(terrainCatalog(t, l, skip), assets.Options{AllowCandidates: true}, l); err == nil || !strings.Contains(err.Error(), skip) {
			t.Errorf("catalog without %s loaded: %v", skip, err)
		}
	}
}

func TestTerrainLayersFollowFog(t *testing.T) {
	c := mapcompose.Cell{Background: "tile/grass", Overlays: []string{"tile/road14"}}
	if bg, ov, dim := terrainLayers(c, state.Unseen); bg != "tile/unseen" || ov != nil || dim {
		t.Errorf("unseen: %s %v %v", bg, ov, dim)
	}
	if bg, ov, dim := terrainLayers(c, state.Seen); bg != "tile/grass" || len(ov) != 1 || !dim {
		t.Errorf("seen: %s %v %v", bg, ov, dim)
	}
	if bg, ov, dim := terrainLayers(c, state.Visible); bg != "tile/grass" || len(ov) != 1 || dim {
		t.Errorf("visible: %s %v %v", bg, ov, dim)
	}
}

func TestTerrainDevIndicators(t *testing.T) {
	g := &game{}
	g.terrain = &terrainArt{artPacks: &artPacks{candidate: true},
		scene: mapcompose.Scene{Gaps: make([]mapcompose.Gap, 3), Issues: make([]mapcompose.Issue, 1)}}
	got := g.devIndicators()
	for _, want := range []string{"CANDIDATE-ART", "CONTEXT-GAP 3", "COMPOUND-ISSUE 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("dev label %q lacks %q", got, want)
		}
	}
}

func TestLevelHex(t *testing.T) {
	l := testLevel(t)
	for li, line := range l.Lines {
		for c, h := range line {
			got := levelHex(l)(hexgrid.Coord{Line: li, Col: c})
			if got.Terrain != h.Terrain || !slices.Equal(got.Infra, h.Infra.IDs) {
				t.Fatalf("(%d,%d) = %+v", li, c, got)
			}
		}
	}
}

func TestDrawnDensity(t *testing.T) {
	g := newGame(testLevel(t), newTestControl(t).st, state.CentralPowers, false)
	g.cam.Zoom = 1
	if d := g.drawnDensity(); d != assetDensity(g.cam.Effective()) {
		t.Errorf("without packs: %dx", d)
	}
	// No 1x art exists: with packs the 2x art is drawn smaller.
	g.terrain = &terrainArt{artPacks: &artPacks{}}
	if d := g.drawnDensity(); d != 2 {
		t.Errorf("with terrain packs at zoom 1: %dx, want 2x", d)
	}
}

func TestPaddedTextureKeepsAnchor(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	src.Pix[3] = 255 // top-left texel opaque black
	pix := padded(src, 2, false)
	if len(pix) != 8*6*4 {
		t.Fatalf("%d bytes", len(pix))
	}
	if pix[(2*8+2)*4+3] != 255 || pix[3] != 0 || pix[(2*8+1)*4+3] != 0 {
		t.Error("texel not moved inside a transparent border")
	}
	// The anchor stays on the screen point whatever the border.
	e := assets.Entry{Anchor: [2]float64{48, 40}}
	for _, pad := range []int{0, 2} {
		m := anchorGeoM(e, 2, pad, 1.5, 100, 50)
		if x, y := m.Apply(float64(48*2+pad), float64(40*2+pad)); x != 100 || y != 50 {
			t.Errorf("pad %d: anchor at %v,%v", pad, x, y)
		}
	}
}

func TestPaddedBleedGrowsOpaqueEdge(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	copy(src.Pix[0:4], []byte{200, 10, 20, 255}) // opaque left texel, rest transparent
	pix := padded(src, 2, true)
	w := 3 + 4
	at := func(x, y int) []byte { i := (y*w + x) * 4; return pix[i : i+4] }
	if !bytes.Equal(at(2, 2), []byte{200, 10, 20, 255}) {
		t.Fatalf("source texel changed: %v", at(2, 2))
	}
	// Two steps reach two texels out, in every direction, and no further.
	for _, p := range [][2]int{{0, 0}, {0, 2}, {4, 2}, {2, 4}, {4, 4}} {
		if !bytes.Equal(at(p[0], p[1]), []byte{200, 10, 20, 255}) {
			t.Errorf("(%d,%d) = %v, want the edge colour", p[0], p[1], at(p[0], p[1]))
		}
	}
	if at(5, 2)[3] != 0 || at(6, 2)[3] != 0 {
		t.Error("bleed went beyond the border width")
	}
}
