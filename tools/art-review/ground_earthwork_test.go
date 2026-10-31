package artreview

import (
	"encoding/json"
	"image"
	"image/color"
	"strings"
	"testing"
)

// Rejecting the opt-in would silently leave the selected feature unrenderable;
// accepting null/aliases/unknown values would weaken the strict input contract.
func TestGroundCircularEarthworkManifest(t *testing.T) {
	path, _ := groundTestFixture(t)
	valid := string(groundTestRead(t, path))
	for _, value := range []string{`"circular-earthwork"`, `null`, `""`, `"noise"`, `true`} {
		data := strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"mud_feature":`+value, 1)
		var m groundManifest
		err := decodeGroundManifest([]byte(data), &m)
		if (err == nil) != (value == `"circular-earthwork"`) {
			t.Fatalf("mud_feature %s: %v", value, err)
		}
	}
	var m groundManifest
	if err := json.Unmarshal([]byte(`{"mud_feature":"circular-earthwork"}`), &m); err != nil || m.MudFeature != "circular-earthwork" {
		t.Fatalf("feature selection not retained: %+v %v", m, err)
	}
}

// A mirrored source sampler would mix opposite quadrants and lose this single
// asymmetric composition; a full-tile fill would contaminate shared guards.
func TestGroundCircularEarthworkPreservesCompositionAndGrassGuards(t *testing.T) {
	materials := groundTestConstants()
	source := image.NewNRGBA(image.Rect(0, 0, 400, 400))
	for y := range 400 {
		for x := range 400 {
			c := color.NRGBA{220, 30, 20, 255}
			if x >= 200 {
				c = color.NRGBA{20, 30, 220, 255}
			}
			source.SetNRGBA(x, y, c)
		}
	}
	materials["mud"] = source
	r := newGroundRenderer(materials)
	r.mudFeature = "circular-earthwork"
	c := groundTestContext(t, "mud", [6]string{"mud", "mud", "mud", "mud", "mud", "mud"})
	tiles, err := r.tiles(c)
	if err != nil {
		t.Fatal(err)
	}
	tile := tiles[4]
	left, right := tile.NRGBAAt(36*4, 40*4), tile.NRGBAAt(60*4, 40*4)
	if left.R < 200 || left.B > 40 || right.B < 200 || right.R > 40 {
		t.Fatalf("single feature composition lost: left=%v right=%v", left, right)
	}
	grass, err := r.tiles(groundTestContext(t, "grass", [6]string{"grass", "grass", "grass", "grass", "grass", "grass"}))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []image.Point{{48 * 4, 2 * 4}, {48 * 4, 77 * 4}, {5 * 4, 40 * 4}, {90 * 4, 40 * 4}} {
		if got, want := tile.NRGBAAt(p.X, p.Y), grass[4].NRGBAAt(p.X, p.Y); got != want {
			t.Fatalf("earthwork polluted shared grass guard %v: %v != %v", p, got, want)
		}
	}
}
