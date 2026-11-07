package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/hajimehoshi/ebiten/v2"
)

// A crater field is reworked on the tile map, and left as the tile on a painted
// map's sapper hexes.
func TestCraterFieldIsReworkedOnTheTileMapOnly(t *testing.T) {
	l := testLevel(t)
	descriptor, ok := mapcompose.Descriptor(mapcompose.Ground, "mud", [6]string{"mud", "mud", "mud", "mud", "mud", "mud"}, [2]int{})
	if !ok {
		t.Fatal("no descriptor")
	}
	exemplar := mapcompose.ContextID(mapcompose.Ground, mapcompose.Key(descriptor))
	a, err := openTerrainArt(terrainCatalog(t, l, "", exemplar), l)
	if err != nil {
		t.Fatal(err)
	}
	defer a.release()
	var field hexgrid.Coord
	found := false
	for line, cells := range a.scene.Cells {
		for col, cell := range cells {
			if c := (hexgrid.Coord{Line: line, Col: col}); !found && a.scene.FieldAt(c, cell.Background) != nil {
				field, found = c, true
			}
		}
	}
	if !found {
		t.Fatal("no crater field on the level")
	}
	id := a.scene.Cells[field.Line][field.Col].Background
	for _, d := range terrainDensities {
		img, e, err := a.fieldImage(field, id, d, true)
		if err != nil {
			t.Fatal(err)
		}
		if got := img.Bounds().Dx(); got != e.Width+2*tilePad {
			t.Errorf("%dx: width %d", d, got)
		}
		again, _, err := a.fieldImage(field, id, d, true)
		if err != nil {
			t.Fatal(err)
		}
		if again != img {
			t.Errorf("%dx: the tile is made twice", d)
		}
	}
	a.fields.release(a)

	dst := ebiten.NewImage(8, 8)
	g := &game{terrain: a, backdrop: &backdrop{}, cam: camera.New(camera.Size{W: 800, H: 600}, camera.Size{W: 800, H: 600})}
	g.drawTerrainTile(dst, id, 0, 0, false, &field, true)
	if len(a.fields.tiles) != 0 {
		t.Error("a painted map's hex is reworked")
	}
	g.backdrop = nil
	g.drawTerrainTile(dst, id, 0, 0, false, &field, true)
	if len(a.fields.tiles) != 1 {
		t.Errorf("%d reworked tiles on the tile map, want 1", len(a.fields.tiles))
	}
}
