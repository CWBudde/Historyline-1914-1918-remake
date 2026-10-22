package main

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

func TestTerrainReleaseClearsMaterialAndLegacyTextures(t *testing.T) {
	key := materialTileKey{mask: "mask", density: 2}
	a := &terrainArt{
		artPacks: &artPacks{tex: map[texKey]*ebiten.Image{{id: "legacy", density: 2}: ebiten.NewImage(2, 2)}},
		materials: &terrainMaterials{
			tiles: map[materialTileKey]*ebiten.Image{key: ebiten.NewImage(2, 2)},
			order: []materialTileKey{key},
			bytes: 16,
		},
	}
	a.release()
	if len(a.tex) != 0 || len(a.materials.tiles) != 0 || len(a.materials.order) != 0 || a.materials.bytes != 0 {
		t.Fatal("release retained terrain textures or cache accounting")
	}
	a.release()
	var absent *terrainArt
	absent.release()
	(&terrainArt{}).release()
}

func TestMaterialManifestStrict(t *testing.T) {
	entries := []assets.Entry{}
	for _, id := range []string{worldmaterial.MeadowID, worldmaterial.ScrubID} {
		entries = append(entries, assets.Entry{ID: id, Density: 2, Width: 1136, Height: 800})
	}
	good := assets.Manifest{Entries: entries}
	if err := validateMaterialManifest(good, 2); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*assets.Entry){func(e *assets.Entry) { e.ID = "unexpected" }, func(e *assets.Entry) { e.Width++ }, func(e *assets.Entry) { e.Height++ }, func(e *assets.Entry) { e.Density = 4 }, func(e *assets.Entry) { e.Anchor[0] = 1 }} {
		bad := assets.Manifest{Entries: append([]assets.Entry(nil), entries...)}
		change(&bad.Entries[0])
		if validateMaterialManifest(bad, 2) == nil {
			t.Fatal("accepted malformed material")
		}
	}
	if validateMaterialManifest(assets.Manifest{Entries: entries[:1]}, 2) == nil {
		t.Fatal("accepted partial material")
	}
}

func TestMaterialBackgroundEligibility(t *testing.T) {
	for _, terrain := range []string{"grass", "hills", "grass_w", "hills_w", "water", "forest"} {
		m := &terrainMaterials{level: &level.Level{Lines: [][]level.Hex{{{Terrain: terrain}}}}}
		c := hexgrid.Coord{}
		want := terrain == "grass" || terrain == "hills"
		if m.eligible(c, "background") != want {
			t.Fatalf("terrain %s", terrain)
		}
		if m.eligible(c, mapcompose.TileID("unseen")) {
			t.Fatal("unseen material")
		}
	}
}

func TestMaterialOriginMatchesPlaceRounding(t *testing.T) {
	for _, c := range []hexgrid.Coord{{Line: 2, Col: 3}, {Line: 3, Col: 3}} {
		for _, d := range []int{2, 4} {
			e := assets.Entry{Anchor: [2]float64{47.375, 39.625}}
			p := materialOrigin(c, e, d)
			x, y := hexgrid.Center(c)
			if p != image.Pt(int(math.Round((x-e.Anchor[0])*float64(d))), int(math.Round((y-e.Anchor[1])*float64(d)))) {
				t.Fatal(p)
			}
		}
	}
}

func TestGameMaterialSamplerUsesMaskAndGlobalOrigin(t *testing.T) {
	grass := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	scrub := image.NewNRGBA(grass.Bounds())
	for y := range 8 {
		for x := range 8 {
			grass.SetNRGBA(x, y, color.NRGBA{R: uint8((x % 7) * 25), A: 255})
			scrub.SetNRGBA(x, y, color.NRGBA{G: 200, A: 255})
		}
	}
	s, err := worldmaterial.New(grass, scrub, image.Pt(8, 8), 1)
	if err != nil {
		t.Fatal(err)
	}
	mask := image.NewNRGBA(image.Rect(7, 9, 10, 11))
	mask.SetNRGBA(7, 9, color.NRGBA{A: 127})
	mask.SetNRGBA(8, 9, color.NRGBA{A: 255})
	field := mapmaterial.New(nil)
	origin := image.Pt(-3, 18)
	got, err := sampleMaterial(s, field, mask, origin)
	if err != nil {
		t.Fatal(err)
	}
	want, err := s.Tile(mask, origin, field.Weight)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds() != image.Rect(0, 0, 3, 2) {
		t.Fatal(got.Bounds())
	}
	for y := range 2 {
		for x := range 3 {
			if got.NRGBAAt(x, y) != want.NRGBAAt(x, y) {
				t.Fatal("shared sampler mismatch")
			}
		}
	}
	if got.NRGBAAt(0, 0).A != 127 || got.NRGBAAt(2, 1) != (color.NRGBA{}) {
		t.Fatal("mask alpha not preserved")
	}
	if got.NRGBAAt(0, 0).R != 125 || got.NRGBAAt(1, 0).R != 150 {
		t.Fatal("material did not use negative global texel origin")
	}
}
