package main

import (
	"bytes"
	"encoding/json"
	"image/color"
	"math"
	"os"
	"testing"
)

func standingMixedFixture(t *testing.T) mixedInputs {
	t.Helper()
	path, _ := mixedFixture(t)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m["bridge_artwork"] = "standing"
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
	in, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	return in
}

func TestMixedStandingBridgeTilesRetainOnlyApproaches(t *testing.T) {
	in := standingMixedFixture(t)
	for _, v := range mixedRoster() {
		if v.Kind == "crossing" {
			continue
		}
		t.Run(v.ID, func(t *testing.T) {
			tiles, err := exportMixedTiles(in, v)
			if err != nil {
				t.Fatal(err)
			}
			axis := connectionRailUnit(sharedConnectionGeometry().Neighbors[v.Ports[0]-1])
			for density, tile := range tiles {
				if v.Kind == "road-bridge" {
					for _, c := range tile.Pix {
						if c != 0 {
							t.Fatal("road bridge tile still paints finite artwork")
						}
					}
					continue
				}
				id := "path" + mixedAxis(v.Ports)
				component := in.components[id]
				paths, err := exportConnectionTiles(component.input, &component.family, &component.variant)
				if err != nil {
					t.Fatal(err)
				}
				for _, along := range []float64{-31, 31} {
					p := connectionPoint{48 + along*axis[0], 40 + along*axis[1]}
					x, y := int(p[0]*float64(density)), int(p[1]*float64(density))
					if got, want := tile.NRGBAAt(x, y), paths[density].NRGBAAt(x, y); got != want || got.A == 0 {
						t.Fatalf("%dx approach at%v: got%v want%v", density, p, got, want)
					}
				}
				for y := 0; y < tile.Bounds().Dy(); y++ {
					for x := 0; x < tile.Bounds().Dx(); x++ {
						q := connectionPoint{(float64(x)+.5)/float64(density) - 48, (float64(y)+.5)/float64(density) - 40}
						if math.Abs(q[0]*axis[0]+q[1]*axis[1]) < 20 && tile.NRGBAAt(x, y) != (color.NRGBA{}) {
							t.Fatalf("%dx finite span underpaint remains at%d,%d", density, x, y)
						}
					}
				}
			}
		})
	}
}

func TestMixedStandingModeLeavesCrossingsAndDefaultUnchanged(t *testing.T) {
	in := standingMixedFixture(t)
	path, _ := mixedFixture(t)
	legacy, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range mixedRoster() {
		before := renderMixedRaw(legacy, v)
		after := renderMixedRaw(in, v)
		if v.Kind == "crossing" && !bytes.Equal(before.Pix, after.Pix) {
			t.Fatalf("standing mode changed crossing %s", v.ID)
		}
		if v.Kind != "crossing" && before.NRGBAAt(208, 176).A == 0 {
			t.Fatalf("default lost finite deck for%s", v.ID)
		}
	}
}

func TestMixedBridgeArtworkRejectsInvalidValues(t *testing.T) {
	path, _ := mixedFixture(t)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err = json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{"", "tile", nil, true} {
		raw["bridge_artwork"] = value
		b, err = json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = decodeMixedManifest(b); err == nil {
			t.Fatalf("accepted bridge_artwork=%v", value)
		}
	}
}
