package main

import (
	"fmt"
	"testing"
)

// Zero borders are accepted input. Interior shoulder wear must never remove
// pixels that the unchanged pavement or water core fully covers.
func TestConnectedZeroBorderPreservesOpaqueCore(t *testing.T) {
	for _, family := range []string{"road", "stream"} {
		t.Run(family, func(t *testing.T) {
			dir, manifest := connectedFixture(t)
			material, border, width := "road", "earth", 22
			if family == "stream" {
				material, border, width = "water", "bank", 18
				manifest["materials"].(map[string]any)["bank"] = manifest["materials"].(map[string]any)["earth"]
			}
			f := map[string]any{"id": family, "width": width, "border": 0, "material": material, "border_material": border}
			if family == "road" {
				f["markings"] = "dashed"
			}
			var variants []any
			for port := 1; port <= 3; port++ {
				variants = append(variants, map[string]any{"id": fmt.Sprintf("%s%d%d", family, port, port+3), "ports": []int{port, port + 3}})
			}
			f["variants"] = variants
			manifest["families"] = []any{f}
			in, err := loadConnectionInputs(writeConnectedManifest(t, dir, manifest))
			if err != nil {
				t.Fatal(err)
			}
			profile := in.manifest.Families[0]
			for _, variant := range profile.Variants {
				route := connectionRoutes(in.manifest.Geometry, profile, variant)
				checked := 0
				for y := 10.; y < 70; y += .5 {
					for x := 10.; x < 86; x += .5 {
						p := connectionPoint{x, y}
						distance := connectionRouteDistance(route, p) - profile.Width/2
						if family == "stream" {
							distance, _, _ = connectionStreamProfile(in.manifest.Geometry, profile, route, p, p)
						}
						if connectionRamp(-distance, .25) != 1 {
							continue
						}
						checked++
						if got := connectionFeature(in, profile, route, p, p, .25); got.A != 255 {
							t.Fatalf("%s cut an opaque core at %v: alpha=%d", variant.ID, p, got.A)
						}
					}
				}
				if checked == 0 {
					t.Fatalf("%s did not exercise the opaque core", variant.ID)
				}
			}
		})
	}
}
