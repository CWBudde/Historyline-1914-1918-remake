package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"slices"
	"testing"
)

// A map-specific painting or duplicated patch cannot establish tile reuse.
// Check independent world corners and reassemble the saved PNGs literally.
func TestShoreReuseProofUsesDistinctCompatibleSavedTiles(t *testing.T) {
	path, _ := shoreTestFixture(t)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := buildShoreBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Scenes     []shoreScene `json:"scenes"`
		Reuse      []shoreScene `json:"reuse_scenes"`
		Catalogues []packOutput `json:"reuse_catalogues"`
	}
	for _, a := range bundle.artifacts {
		if a.name == "report.json" {
			if err := json.Unmarshal(a.data, &report); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(report.Reuse) < 6 || len(report.Catalogues) != 2 {
		t.Fatalf("missing reusable proof: %d arrangements, %d catalogues", len(report.Reuse), len(report.Catalogues))
	}
	if len(report.Scenes) != 20 {
		t.Fatalf("changed legacy scenes: %d", len(report.Scenes))
	}
	profiles := shoreTestProfiles()
	profiles["grass"] = []int{0, 1, 2, 3, 4, 5}
	profiles["water"] = []int{}
	vertices := []image.Point{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}
	seen := map[string]bool{}
	var arrangements [][]string
	for _, scene := range report.Reuse {
		if len(scene.Cells) != 7 {
			t.Fatalf("%s: %d cells", scene.ID, len(scene.Cells))
		}
		if scene.CenterID != "lakeside23" {
			t.Fatalf("%s does not demonstrate reuse of the same central tile", scene.ID)
		}
		corners := map[image.Point]bool{}
		key := ""
		var ids []string
		for _, cell := range scene.Cells {
			profile, ok := profiles[cell.RuntimeID]
			if !ok {
				t.Fatalf("invented tile %s", cell.RuntimeID)
			}
			key += cell.RuntimeID + "/"
			ids = append(ids, cell.RuntimeID)
			for v, p := range vertices {
				at := p.Add(image.Pt(int(cell.Origin[0]), int(cell.Origin[1])))
				land := slices.Contains(profile, v)
				if old, exists := corners[at]; exists && old != land {
					t.Fatalf("%s: incompatible corner %v", scene.ID, at)
				}
				corners[at] = land
			}
		}
		if seen[key] {
			t.Fatalf("duplicate arrangement %s", scene.ID)
		}
		seen[key] = true
		for _, earlier := range arrangements {
			changes := 0
			for i, id := range ids {
				if id != earlier[i] {
					changes++
				}
			}
			if changes < 2 {
				t.Fatalf("%s differs by only %d tiles; proof needs more varied arrangements", scene.ID, changes)
			}
		}
		arrangements = append(arrangements, ids)
		for _, density := range []int{2, 4} {
			name := fmt.Sprintf("%s-%dx.png", scene.ID, density)
			got := bundle.pixels[name]
			if got == nil {
				t.Fatalf("missing %s", name)
			}
			want := image.NewNRGBA(image.Rect(0, 0, 238*density, 240*density))
			for _, cell := range scene.Cells {
				src := bundle.pixels[fmt.Sprintf("tiles/%s-%dx.png", cell.RuntimeID, density)]
				at := image.Pt(int(cell.Origin[0]-scene.CanvasOrigin[0])*density, int(cell.Origin[1]-scene.CanvasOrigin[1])*density)
				draw.Draw(want, src.Bounds().Add(at), src, src.Bounds().Min, draw.Over)
			}
			if !bytes.Equal(got.Pix, want.Pix) {
				t.Fatalf("%s: changed saved tile pixels", name)
			}
		}
	}
	// Export metadata must identify real native-density sheets, not an HTML desk.
	for i, scale := range []int{2, 4} {
		c := report.Catalogues[i]
		if c.Path != fmt.Sprintf("shore-catalogue-%dx.png", scale) || bundle.pixels[c.Path] == nil {
			t.Fatalf("missing catalogue: %+v", c)
		}
		catalogue := bundle.pixels[c.Path]
		// The nine source tiles occupy a 3x3 catalogue in roster order. Check
		// every owned pixel to detect resampling or a missing/wrong tile.
		for tile, id := range []string{"lakeside12", "lakeside16", "lakeside23", "lakeside34", "lakeside56", "lakeside123", "lakeside126", "lakeside156", "lakeside3456"} {
			src := bundle.pixels[fmt.Sprintf("tiles/%s-%dx.png", id, scale)]
			at := image.Pt((83+tile%3*254)*scale, (32+tile/3*104)*scale)
			for y := range src.Bounds().Dy() {
				for x := range src.Bounds().Dx() {
					want := src.NRGBAAt(x, y)
					if want.A == 255 && catalogue.NRGBAAt(at.X+x, at.Y+y) != want {
						t.Fatalf("%s %dx catalogue altered native tile pixel %d,%d", id, scale, x, y)
					}
				}
			}
		}
	}
}
