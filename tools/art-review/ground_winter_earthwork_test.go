package main

import (
	"encoding/json"
	"image"
	"image/color"
	"strings"
	"testing"
)

// A missing or permissive opt-in would reject authored winter craters or accept malformed modes.
func TestGroundWinterEarthworkManifest(t *testing.T) {
	path, _ := groundTestFixture(t)
	valid := string(groundTestRead(t, path))
	for _, value := range []string{`"circular-earthwork"`, `null`, `""`, `"noise"`, `true`} {
		data := strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"shelled_feature":`+value, 1)
		var m groundManifest
		err := decodeGroundManifest([]byte(data), &m)
		if (err == nil) != (value == `"circular-earthwork"`) {
			t.Fatalf("shelled_feature %s: %v", value, err)
		}
	}
}

// Mirroring the winter crater creates multiple impacts; composing onto summer grass breaks snow edges.
func TestGroundWinterEarthworkPreservesCompositionAndSnowGuards(t *testing.T) {
	path, _ := groundTestFixture(t)
	data := strings.Replace(string(groundTestRead(t, path)), `"schema_version":1`, `"schema_version":1,"shelled_feature":"circular-earthwork"`, 1)
	var m groundManifest
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Fatal(err)
	}
	materials := groundTestConstants()
	source := image.NewNRGBA(image.Rect(0, 0, 400, 400))
	for y := range 400 {
		for x := range 400 {
			p := color.NRGBA{220, 30, 20, 255}
			if x >= 200 {
				p = color.NRGBA{20, 30, 220, 255}
			}
			source.SetNRGBA(x, y, p)
		}
	}
	materials["shelled_w"] = source
	in := groundInputs{Manifest: m, Sources: map[string]compoundSource{}}
	for id, im := range materials {
		in.Sources[id] = compoundSource{Image: im}
	}
	b, err := buildGroundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	r := newGroundRenderer(materials)
	base, err := r.tiles(groundTestContext(t, "grass_w", [6]string{"grass_w", "grass_w", "grass_w", "grass_w", "grass_w", "grass_w"}))
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range b.report.Profiles {
		if profile.RuntimeID != "shelled_w" {
			continue
		}
		for _, out := range profile.Outputs {
			tile := b.pixels[out.Path]
			scale := tile.Bounds().Dx() / 96
			left, right := tile.NRGBAAt(36*scale, 40*scale), tile.NRGBAAt(60*scale, 40*scale)
			if left.R < 200 || left.B > 40 || right.B < 200 || right.R > 40 {
				t.Fatalf("winter crater composition mirrored: left=%v right=%v", left, right)
			}
			for y := range tile.Bounds().Dy() {
				for x := range tile.Bounds().Dx() {
					if !groundTestOwned(x, y, scale) {
						continue
					}
					for n, delta := range groundTestDeltas {
						if profile.NeighborIDs[n] == "grass_w" && groundTestDistance((float64(x)+.5)/float64(scale)-float64(delta[0]), (float64(y)+.5)/float64(scale)-float64(delta[1])) <= 4 && tile.NRGBAAt(x, y) != base[scale].NRGBAAt(x, y) {
							t.Fatal("winter crater polluted snow quiet band")
						}
					}
				}
			}
		}
	}
}
