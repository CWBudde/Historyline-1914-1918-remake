package artreview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"testing"
)

func TestAuthoredBankKeepsPaintedRockContrast(t *testing.T) {
	path, manifest := shoreTestFixture(t)
	legacy, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest["bank_profile"] = "authored-rocks-v1"
	shoreTestWrite(t, path, manifest)
	authored, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []shoreInputs{legacy, authored} {
		bank := in.Materials["bank"].Image
		for y := range bank.Bounds().Dy() {
			for x := range bank.Bounds().Dx() {
				c := color.NRGBA{40, 45, 40, 255}
				if x < bank.Bounds().Dx()/2 {
					c = color.NRGBA{210, 205, 190, 255}
				}
				bank.SetNRGBA(x, y, c)
			}
		}
	}
	p := connectionPoint{48, 40}
	old := shoreMaterialColor(legacy, p, 3)
	got := shoreMaterialColor(authored, p, 3)
	if old.R >= 160 || got.R < 195 {
		t.Fatalf("painted light rock was averaged away: legacy=%v authored=%v", old, got)
	}
}

func TestAuthoredWaterStonesGroundMatchesShore(t *testing.T) {
	path, _ := waterStonesFixture(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m["water_profile"] = "authored-rocks-v1"
	shoreTestWrite(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	shore := shoreInputs{Manifest: shoreManifest{BankProfile: authoredBankProfile}, Materials: map[string]compoundSource{"water": in.Grounds["shared"]}}
	want, err := exportShoreTiles(shore, shoreVariant{ID: "water"})
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{2, 4} {
		name := fmt.Sprintf("ground/water-%dx.png", density)
		got := bundle.pixels[name]
		if !bytes.Equal(got.Pix, want[density].Pix) {
			t.Fatalf("%dx compound water differs from shore base", density)
		}
	}
}

func TestAuthoredBankProfileValidation(t *testing.T) {
	for _, profile := range []string{"", "authored-rocks-v1", "unknown"} {
		t.Run(profile, func(t *testing.T) {
			path, m := shoreTestFixture(t)
			if profile != "" {
				m["bank_profile"] = profile
			}
			shoreTestWrite(t, path, m)
			_, err := loadShoreInputs(path)
			if (err != nil) != (profile == "unknown") {
				t.Fatalf("profile %q: %v", profile, err)
			}
		})
	}
	dir, m := connectedFixture(t)
	f := map[string]any{"id": "stream", "width": 18, "border": 3.5, "material": "water", "border_material": "bank", "bank_profile": "authored-rocks-v1", "variants": []any{map[string]any{"id": "stream36", "ports": []int{3, 6}}}}
	m["families"] = []any{f}
	m["materials"].(map[string]any)["bank"] = m["materials"].(map[string]any)["earth"]
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err != nil {
		t.Fatal(err)
	}
	f["id"], f["material"], f["border_material"] = "road", "road", "earth"
	delete(m["materials"].(map[string]any), "bank")
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
		t.Fatal("authored bank profile accepted for road")
	}
}

func TestAuthoredMaterialLatticePeriodicity(t *testing.T) {
	_, _, source := packFixture(t, 1254, 1254, false)
	for _, sample := range []func(connectionPoint) color.NRGBA{func(p connectionPoint) color.NRGBA { return sampleAuthoredBank(source, p) }, func(p connectionPoint) color.NRGBA { return sampleAuthoredWater(source, p) }} {
		for _, p := range []connectionPoint{{0, 0}, {.125, .875}, {-22.5, 81.5}, {35.499999, 19.999999}, {71, 40}} {
			want := sample(p)
			for _, delta := range sharedConnectionGeometry().Neighbors {
				if got := sample(p.add(delta)); got != want {
					t.Fatalf("authored phase %v+%v: %v != %v", p, delta, got, want)
				}
			}
		}
	}
}
