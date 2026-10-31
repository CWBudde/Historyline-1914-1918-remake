package artreview

import (
	"encoding/json"
	"image"
	"image/color"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testForestCanopyProfile = "canopy-26.5-30-v1"

func forestTestSelectCanopyProfile(t *testing.T, path string) {
	t.Helper()
	var manifest map[string]any
	if err := json.Unmarshal(groundTestRead(t, path), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest["summer_profile"] = testForestCanopyProfile
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, data)
}

func TestForestSummerCanopyProfileConfiguration(t *testing.T) {
	path := forestTestFixture(t)
	baseline, err := loadForestInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	old, err := buildForestBundle(baseline)
	if err != nil {
		t.Fatal(err)
	}
	forestTestSelectCanopyProfile(t, path)
	in, err := loadForestInputs(path)
	if err != nil {
		t.Fatalf("named summer canopy profile rejected: %v", err)
	}
	bundle, err := buildForestBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(bundle.report)
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report["summer_profile"] != testForestCanopyProfile || report["summer_crown_world_size"] != 26.5 || report["summer_tree_count"] != float64(30) {
		t.Fatal("profile geometry absent from report")
	}
	if !reflect.DeepEqual(bundle.report.SummerVariantOrder, []int{0, 1, 0, 1, 2, 0, 1, 3}) {
		t.Fatal("summer species roster missing from report")
	}
	if bundle.report.CrownWorldSize != 24 || len(bundle.report.Template) != 36 {
		t.Fatal("default/winter geometry changed")
	}
	homogeneous := 0
	for _, p := range bundle.report.Profiles {
		allForest := p.RuntimeID == "forest"
		for _, id := range p.NeighborIDs {
			allForest = allForest && id == "forest"
		}
		if allForest {
			homogeneous++
			if len(p.Instances) != 7*30 {
				t.Fatalf("summer placement report has %d instances; want210", len(p.Instances))
			}
			order := []int{0, 1, 0, 1, 2, 0, 1, 3}
			for _, instance := range p.Instances {
				if instance.Variant != order[instance.PointIndex%len(order)] {
					t.Fatalf("summer point%d species%d; want%d", instance.PointIndex, instance.Variant, order[instance.PointIndex%len(order)])
				}
			}
		}
	}
	if homogeneous != 1 {
		t.Fatal("homogeneous summer profile not exercised")
	}
	for _, output := range bundle.report.Exports {
		if strings.Contains(output.Path, "winter-") || strings.Contains(output.Path, "forest_w-") {
			a, b := bundle.pixels[output.Path], old.pixels[output.Path]
			if a == nil || b == nil || string(a.Pix) != string(b.Pix) {
				t.Fatalf("winter raster changed: %s", output.Path)
			}
		}
	}
}

func TestForestSummerCanopyRejectsInvalidProfileValues(t *testing.T) {
	path := forestTestFixture(t)
	original := string(groundTestRead(t, path))
	for _, value := range []string{`null`, `0`, `""`, `"unknown"`, `true`} {
		t.Run(value, func(t *testing.T) {
			data := strings.Replace(original, `"schema_version":1`, `"schema_version":1,"summer_profile":`+value, 1)
			var manifest forestManifest
			if err := decodeForestManifest([]byte(data), &manifest); err == nil {
				t.Fatal("accepted invalid profile value")
			}
		})
	}
}

func TestForestPaintedSupportIncludesEntirePixelSquare(t *testing.T) {
	crown := image.NewRGBA(image.Rect(0, 0, 106, 106))
	crown.SetRGBA(104, 96, color.RGBA{1, 1, 1, 1})
	want := math.Hypot(13, 11)
	if radius := forestCrownSupport(crown); radius != want || radius <= math.Hypot(12, 12) {
		t.Fatalf("painted square support=%g;want%g beyond original bound", radius, want)
	}
}

func TestForestSummerCanopyDrawUsesActualPanelHalfSize(t *testing.T) {
	path := forestTestFixture(t)
	in, err := loadForestInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	grounds, atlases := map[string]*image.NRGBA{}, map[string]*image.NRGBA{}
	for _, season := range []string{"summer", "winter"} {
		grounds[season], atlases[season] = in.Grounds[season].Image, in.Atlases[season].Image
	}
	r, err := newForestRenderer(grounds, atlases, testForestCanopyProfile)
	if err != nil {
		t.Fatal(err)
	}
	marker := color.RGBA{251, 1, 2, 255}
	for i := range 4 {
		crown := image.NewRGBA(image.Rect(0, 0, 106, 106))
		crown.SetRGBA(53, 53, marker)
		crowns := r.crowns["summer"]
		crowns[i] = crown
		r.crowns["summer"] = crowns
	}
	c, err := newForestContext("forest", []string{"forest", "forest", "forest", "forest", "forest", "forest"})
	if err != nil {
		t.Fatal(err)
	}
	// The first pinned centre is(20.5,71.5). Pixel53 is at the panel
	// half-size, so its destination is exactly that centre plus overscan.
	got := r.raw(c).NRGBAAt(98, 302)
	if got != color.NRGBA(marker) {
		t.Fatalf("marker at reported centre=%v;want%v", got, marker)
	}
}

func TestForestSummerCanopyRejectsPaintOutsideOriginalSupport(t *testing.T) {
	path := forestTestFixture(t)
	forestTestSelectCanopyProfile(t, path)
	atlas := forestTestAtlas()
	// Existing atlas guards permit translucent outer pixels. At the larger
	// profile size these painted corners exceed the original seven-site bound.
	for y := range 16 {
		for x := range 16 {
			atlas.SetNRGBA(x, y, color.NRGBA{30, 60, 20, 249})
		}
	}
	if err := validateForestAtlas(atlas); err != nil {
		t.Fatal(err)
	}
	ref := forestTestWritePNG(t, filepath.Join(filepath.Dir(path), "atlas-summer.png"), atlas)
	var manifest map[string]any
	if err := json.Unmarshal(groundTestRead(t, path), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest["atlases"].(map[string]any)["summer"] = ref
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, data)
	in, err := loadForestInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = buildForestBundle(in); err == nil || !strings.Contains(err.Error(), "support") {
		t.Fatalf("oversized painted alpha accepted: %v", err)
	}
}
