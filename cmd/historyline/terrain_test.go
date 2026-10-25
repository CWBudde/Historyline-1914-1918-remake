package main

import (
	"bytes"
	"cmp"
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
	l, err := loadLevel("levels-not-here", "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// terrainCatalog is a terrain catalog of transparent tiles holding every tile
// level l draws, minus skip, plus the context and standing IDs in contexts.
func terrainCatalog(t *testing.T, l *level.Level, skip string, contexts ...string) fstest.MapFS {
	t.Helper()
	s := mapcompose.Compose(l.Grid(), mapcompose.LevelHex(l), func(string) bool { return false })
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
		if strings.HasPrefix(id, "standing/stamp/") {
			ids[id] = "standing-stamps"
			continue
		}
		if strings.HasPrefix(id, "standing/") {
			ids[id] = standingGroup
			continue
		}
		if strings.HasPrefix(id, "stamp/") {
			ids[id] = stampGroup
			continue
		}
		ids[id] = "context-" + strings.Split(id, "/")[1]
	}
	delete(ids, skip)
	fsys := fstest.MapFS{}
	for _, d := range terrainDensities {
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
			if m == nil {
				m = &assets.Manifest{Schema: 1, ID: fmt.Sprintf("terrain-%s-%dx", ids[id], d), Status: "candidate"}
				packs[ids[id]] = m
			}
			p := fmt.Sprintf("images/%04d.png", len(m.Entries))
			fsys[path.Join(fmt.Sprintf("%dx", d), ids[id], p)] = &fstest.MapFile{Data: buf.Bytes()}
			m.Entries = append(m.Entries, assets.Entry{
				ID: id, Path: p, SHA256: hex.EncodeToString(sum[:]),
				Width: 96 * d, Height: 80 * d, Density: d, Anchor: [2]float64{48, 40},
			})
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
	plain := mapcompose.Compose(l.Grid(), mapcompose.LevelHex(l), func(string) bool { return false })
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
	rows := [][]string{{"house", "mud", "grass"}, {"mud", "grass"}, {"grass", "grass", "grass"}}
	stamp := mapcompose.StampID("crater", "2_-1.1", 1)
	have := map[string]bool{"standing/house": true, stamp: true}
	s := mapcompose.Compose(hexgrid.Grid{Width: 3, Lines: 3}, func(c hexgrid.Coord) mapcompose.Hex {
		return mapcompose.Hex{Terrain: rows[c.Line][c.Col]}
	}, func(id string) bool { return have[id] })
	house, crater, rim := hexgrid.Coord{Line: 0, Col: 0}, hexgrid.Coord{Line: 0, Col: 1}, hexgrid.Coord{Line: 1, Col: 0}
	fog := map[hexgrid.Coord]state.Fog{house: state.Seen, rim: state.Unseen}
	fogOf := func(c hexgrid.Coord) state.Fog {
		if f, ok := fog[c]; ok {
			return f
		}
		return state.Visible
	}
	if bg, _, _, st, dim := terrainLayers(&s, house, fogOf); bg != "tile/house" || !slices.Equal(st, []string{"standing/house"}) || !dim {
		t.Errorf("seen: %s %v %v", bg, st, dim)
	}
	if bg, stamp, ov, st, dim := terrainLayers(&s, rim, fogOf); bg != "tile/unseen" || stamp != "" || ov != nil || st != nil || dim {
		t.Errorf("unseen: %s %s %v %v %v", bg, stamp, ov, st, dim)
	}
	// Part of the crater is unseen: its known hex keeps the per-hex art.
	if bg, stamp, _, _, dim := terrainLayers(&s, crater, fogOf); bg != "tile/mud" || stamp != "" || dim {
		t.Errorf("partly known crater: %s %s %v", bg, stamp, dim)
	}
	fog[rim] = state.Visible
	if bg, got, _, _, _ := terrainLayers(&s, crater, fogOf); bg != "tile/grass" || got != stamp {
		t.Errorf("known crater: %s %s", bg, got)
	}
	// A stamp anchored outside the view is still drawn when the crater reaches in.
	if got := withClusterAnchors(&s, []hexgrid.Coord{house, rim}); !slices.Equal(got, []hexgrid.Coord{house, crater, rim}) {
		t.Errorf("cells drawn %v", got)
	}
}

func TestTerrainShadowsFollowFog(t *testing.T) {
	s := mapcompose.Scene{Cells: [][]mapcompose.Cell{{{Standing: []string{"standing/forest"}, Shadows: []string{"standing/shadow/forest"}}}}}
	for _, f := range []state.Fog{state.Unseen, state.Seen, state.Visible} {
		got, dim := terrainShadows(&s, hexgrid.Coord{}, func(hexgrid.Coord) state.Fog { return f })
		if f == state.Unseen {
			if len(got) != 0 || dim {
				t.Errorf("unseen shadows %v dim %v", got, dim)
			}
		} else if !slices.Equal(got, []string{"standing/shadow/forest"}) || dim != (f == state.Seen) {
			t.Errorf("fog %v shadows %v dim %v", f, got, dim)
		}
	}
}

func TestTerrainDevIndicators(t *testing.T) {
	g := &game{}
	g.terrain = &terrainArt{
		artPacks: &artPacks{candidate: true},
		scene:    mapcompose.Scene{Gaps: make([]mapcompose.Gap, 3), Issues: make([]mapcompose.Issue, 1)},
	}
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
			got := mapcompose.LevelHex(l)(hexgrid.Coord{Line: li, Col: c})
			if got.Terrain != h.Terrain || !slices.Equal(got.Infra, h.Infra.IDs) || got.Vacant != (h.Content == nil) {
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
	// No 1x unit art exists: with unit packs the 2x art is drawn smaller.
	g.art = &artPacks{}
	if d := g.drawnDensity(); d != 2 {
		t.Errorf("with unit packs at zoom 1: %dx, want 2x", d)
	}
	// Terrain has 1x packs, derived from the 2x ones.
	g.terrain = &terrainArt{artPacks: &artPacks{}}
	if d := g.drawnDensity(); d != 1 {
		t.Errorf("with terrain packs at zoom 1: %dx, want 1x", d)
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

func TestOpenTerrainArtAddsStandingItems(t *testing.T) {
	l := testLevel(t)
	a, err := openTerrainArt(terrainCatalog(t, l, "", mapcompose.StandingID("house")), assets.Options{AllowCandidates: true}, l)
	if err != nil {
		t.Fatal(err)
	}
	if a.bleed[standingGroup] {
		t.Error("standing items bleed into their border")
	}
	houses := 0
	for li, line := range l.Lines {
		for c, h := range line {
			st := a.scene.Cells[li][c].Standing
			if h.Terrain == "house" {
				houses++
				if !slices.Equal(st, []string{"standing/house"}) {
					t.Errorf("house at (%d,%d) has standing %v", li, c, st)
				}
			} else if st != nil {
				t.Errorf("(%d,%d) %s has standing %v", li, c, h.Terrain, st)
			}
		}
	}
	if houses == 0 {
		t.Fatal("level has no house")
	}
}

func TestOpenTerrainArtAddsSeparateClusterStandingPack(t *testing.T) {
	l := testLevel(t)
	stamp := mapcompose.StampID("knoll", "1", 1)
	standing := mapcompose.StandingID(stamp)
	fsys := terrainCatalog(t, l, "", stamp, standing)
	a, err := openTerrainArt(fsys, assets.Options{AllowCandidates: true}, l)
	if err != nil {
		t.Fatal(err)
	}
	if a.group[standing] != "standing-stamps" || a.bleed["standing-stamps"] {
		t.Fatalf("standing stamp group=%s, bleed=%v", a.group[standing], a.bleed["standing-stamps"])
	}
	selected := 0
	for _, cl := range a.scene.Clusters {
		if cl.Stamp == stamp {
			selected++
			if cl.Standing != standing {
				t.Errorf("selected cluster has standing %q", cl.Standing)
			}
		}
	}
	if selected == 0 {
		t.Fatal("fixture did not exercise a stamped cluster")
	}
	delete(fsys, "4x/standing-stamps/manifest.json")
	if _, err := openTerrainArt(fsys, assets.Options{AllowCandidates: true}, l); err == nil {
		t.Fatal("accepted cluster standing pack with a missing density")
	}
}

// Backgrounds have a bleeding texture, drawn beneath all backgrounds, and a
// bare one drawn over them; overlays have only the bare one. Terrain loads 1x.
func TestBackgroundsHaveBareTextures(t *testing.T) {
	l := testLevel(t)
	a, err := openTerrainArt(terrainCatalog(t, l, ""), assets.Options{AllowCandidates: true}, l)
	if err != nil {
		t.Fatal(err)
	}
	var bg, overlay string
	for _, line := range a.scene.Cells {
		for _, c := range line {
			bg = cmp.Or(bg, c.Background)
			if len(c.Overlays) > 0 {
				overlay = cmp.Or(overlay, c.Overlays[0])
			}
		}
	}
	if bg == "" || overlay == "" {
		t.Fatal("fixture has no background or overlay")
	}
	for _, d := range terrainDensities {
		for _, tt := range []struct {
			id    string
			apart bool
		}{{bg, true}, {overlay, false}} {
			bled, _, err := a.image(a.group[tt.id], tt.id, d)
			if err != nil {
				t.Fatal(err)
			}
			bare, _, err := a.bareImage(a.group[tt.id], tt.id, d)
			if err != nil {
				t.Fatal(err)
			}
			if (bled != bare) != tt.apart || bare.Bounds() != bled.Bounds() {
				t.Errorf("%s at %dx: textures apart %v, want %v", tt.id, d, bled != bare, tt.apart)
			}
		}
	}
	if terrainDensities[0] != 1 || terrainDensity(1) != 1 || terrainDensity(0.5) != 1 {
		t.Errorf("terrain densities %v draw %dx at zoom 1", terrainDensities, terrainDensity(1))
	}
}
