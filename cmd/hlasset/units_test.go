package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

var (
	unitRecipes = filepath.Join("..", "..", "art", "recipes")
	unitExports = filepath.Join("..", "..", "art", "exports")
)

func TestRetainedUnitCatalogBuildsEveryFamilyAndDensity(t *testing.T) {
	out := filepath.Join(t.TempDir(), "units")
	packs, err := buildUnits(unitRecipes, unitExports, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != len(unitsprite.Families())*2 {
		t.Fatalf("built %d packs, want %d", len(packs), len(unitsprite.Families())*2)
	}
	for _, p := range packs {
		m := p.Manifest
		if m.Status != "candidate" || m.ID != fmt.Sprintf("unit-%s-%dx", p.Family, p.Density) || len(m.Entries) != 36 {
			t.Fatalf("pack %s: status %s, %d entries", m.ID, m.Status, len(m.Entries))
		}
		dir := filepath.Join(out, fmt.Sprintf("%dx", p.Density), p.Family)
		r, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		pack, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
		_ = r.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, pal := range unitsprite.Palettes {
			for f := range unitsprite.Facings {
				for _, s := range unitsprite.States {
					e, err := pack.Entry(unitsprite.ID(p.Family, pal, f, s), p.Density)
					if err != nil {
						t.Fatal(err)
					}
					if e.Anchor != [2]float64{48, 49} || e.Width != 96*p.Density || e.Height != 96*p.Density {
						t.Fatalf("%s: registration %v %dx%d", e.ID, e.Anchor, e.Width, e.Height)
					}
				}
			}
		}
	}
	// Bytes are copied, not re-encoded: compare one sprite per pack with its export.
	src, err := os.ReadFile(filepath.Join(unitExports, "unit-light-artillery-v1", "sprites", "lightArtillery", "ec", "nw-spent-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := os.OpenRoot(filepath.Join(out, "4x", "light-artillery"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	pack, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := pack.WritePNG(&got, "unit/light-artillery/ec/nw-spent", 4); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Bytes(), src) {
		t.Fatal("runtime build changed the source PNG bytes")
	}
	var output bytes.Buffer
	if err := run([]string{"validate", "-root", filepath.Join(out, "2x", "infantry"), "-allow-candidates"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "validated unit-infantry-2x: 36 PNG variants") {
		t.Fatal(output.String())
	}
	before, err := os.ReadFile(filepath.Join(out, "2x", "infantry", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := buildUnits(unitRecipes, unitExports, out); err == nil {
		t.Fatal("existing unit catalog overwritten")
	}
	after, err := os.ReadFile(filepath.Join(out, "2x", "infantry", "manifest.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed rebuild changed the existing catalog", err)
	}
}

func TestBuildUnitsCommand(t *testing.T) {
	out := filepath.Join(t.TempDir(), "units")
	var output bytes.Buffer
	if err := run([]string{"build-units", "-recipes", unitRecipes, "-exports", unitExports, "-out", out}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "built 30 unit packs (15 families × 2 densities, 1080 PNG variants; candidate") {
		t.Fatal(output.String())
	}
	if err := run([]string{"build-units"}, &bytes.Buffer{}); err == nil {
		t.Fatal("missing -out accepted")
	}
}

// unitRecipe loads a real recipe so a test can damage one detail of it.
func unitRecipe(t *testing.T, name string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(unitRecipes, name))
	if err != nil {
		t.Fatal(err)
	}
	var r map[string]any
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestImportUnitsRejectsDamagedRecipes(t *testing.T) {
	variant := func(r map[string]any, i int) map[string]any {
		return r["bundle_report"].(map[string]any)["variants"].([]any)[i].(map[string]any)
	}
	cases := []struct {
		name   string
		damage func(map[string]any)
		want   string
	}{
		{"unknown family", func(r map[string]any) { r["id"] = "unit-zeppelin-v1" }, "no sprite family"},
		{"class drift", func(r map[string]any) {
			r["bundle_report"].(map[string]any)["class_ids"] = []any{"CP_cavalry"}
		}, "class IDs"},
		{"unknown palette", func(r map[string]any) { variant(r, 0)["palette"] = "red" }, "palette"},
		{"unknown state", func(r map[string]any) { variant(r, 0)["state"] = "dead" }, "state"},
		{"unknown facing", func(r map[string]any) { variant(r, 0)["facing_label"] = "e" }, "facing"},
		{"duplicate variant", func(r map[string]any) {
			v := r["bundle_report"].(map[string]any)["variants"].([]any)
			v[1] = v[0]
		}, "duplicate"},
		{"missing density", func(r map[string]any) {
			v := variant(r, 3)
			v["outputs"] = v["outputs"].([]any)[:1]
		}, "density"},
		{"density from size", func(r map[string]any) {
			o := variant(r, 0)["outputs"].([]any)[0].(map[string]any)
			o["dimensions"] = map[string]any{"width": 100, "height": 100}
		}, "96"},
	}
	for _, c := range cases {
		r := unitRecipe(t, "unit-cavalry-v1.json")
		c.damage(r)
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := importUnits(b); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want %q", c.name, err, c.want)
		}
	}
}

// A sprite whose bytes do not match the recipe pin fails before any output exists.
func TestBuildUnitsLeavesNoOutputOnHashMismatch(t *testing.T) {
	root := t.TempDir()
	recipes := filepath.Join(root, "recipes")
	if err := os.Mkdir(recipes, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, fam := range unitsprite.Families() {
		r := unitRecipe(t, "unit-"+fam+"-v1.json")
		if fam == "sappers" {
			o := r["bundle_report"].(map[string]any)["variants"].([]any)[5].(map[string]any)["outputs"].([]any)[1].(map[string]any)
			o["sha256"] = strings.Repeat("0", 64)
		}
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(recipes, "unit-"+fam+"-v1.json"), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(root, "units")
	if _, err := buildUnits(recipes, unitExports, out); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatal("tampered pin accepted", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed build left output", err)
	}
	// A catalog missing a family is incomplete and refused too.
	if err := os.Remove(filepath.Join(recipes, "unit-sappers-v1.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := buildUnits(recipes, unitExports, out); err == nil || !strings.Contains(err.Error(), "sappers") {
		t.Fatal("incomplete catalog accepted", err)
	}
}
