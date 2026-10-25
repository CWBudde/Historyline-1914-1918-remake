//go:build grassgfx

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// This private candidate leaves all active grass exports intact; its report goes to
// build/review/grass-refdesign-detail-v1/grass-variations.json.
// go test -tags grassgfx ./tools/art-review -run TestGenerateGrassDetailCandidate -count=1 -v
func TestGenerateGrassDetailCandidate(t *testing.T) {
	root := filepath.Join("..", "..")
	directory := filepath.Join(root, "build/review/grass-refdesign-detail-v1")
	if err := os.MkdirAll(filepath.Join(directory, "exports"), 0o755); err != nil {
		t.Fatal(err)
	}
	read := func(path, want string) (*image.NRGBA, string) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		hash := hex.EncodeToString(h[:])
		if want != hash {
			t.Fatalf("unreviewed source: %s", path)
		}
		decoded, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		im := image.NewNRGBA(image.Rect(0, 0, decoded.Bounds().Dx(), decoded.Bounds().Dy()))
		draw.Draw(im, im.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
		return im, hash
	}
	write := func(path string, im *image.NRGBA) packOutput {
		t.Helper()
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b.Bytes())
		return packOutput{path, hex.EncodeToString(h[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
	}
	basePath := "art/terrain/grass/grass-summer-v1.png"
	base, baseHash := read(basePath, "8e751025b89fe32233f01a9f8090dd280fb702ebbd9954f279583ba732fa8a3b")
	materials := map[string]*image.NRGBA{"base": base}
	materialPins := map[string]any{"base": map[string]any{"source": basePath, "source_sha256": baseHash, "source_dimensions": []int{768, 480}}}
	for _, spec := range []struct{ id, hash string }{{"flowers", "368607bd87ee0b7ecbe39fc87b970982b001d6d3eeaabc3d9ee4f36e811cac89"}, {"worn", "d0a4b7a2eec6551a246581d961f139f434ad9f3f9694d8a54fdbec5a48b77078"}} {
		masterPath := "art/terrain/grass/grass-" + spec.id + "-painting-summer-v1.png"
		master, _ := read(masterPath, spec.hash)
		material, err := resizeLanczos(master, 768, 480)
		if err != nil {
			t.Fatal(err)
		}
		path := "art/terrain/grass/grass-" + spec.id + "-material-summer-v1.png"
		pin := write(filepath.Join(root, path), material)
		pin.Path = path
		materials[spec.id] = material
		materialPins[spec.id] = map[string]any{"source": path, "source_sha256": pin.SHA256, "source_dimensions": []int{768, 480}, "master_source": masterPath, "master_sha256": spec.hash, "master_dimensions": []int{master.Bounds().Dx(), master.Bounds().Dy()}, "projection": "Whole source alpha-aware antialiased Lanczos3 to768x480, preserving the edited canonical grass composition and relative feature scale."}
	}
	selection := [16]string{"base", "flowers", "worn", "flowers", "worn", "base", "flowers", "worn", "flowers", "worn", "base", "flowers", "worn", "flowers", "worn", "base"}
	patchBounds := image.Rect(0, 0, (4*142+96)*2, (8*40+80)*2)
	before, after := image.NewNRGBA(patchBounds), image.NewNRGBA(patchBounds)
	var variants []map[string]any
	for index, id := range selection {
		old, err := renderGrassVariation(base, index)
		if err != nil {
			t.Fatal(err)
		}
		next, err := renderGrassDetailVariation(base, materials[id], index)
		if err != nil {
			t.Fatal(err)
		}
		var outputs []packOutput
		for _, d := range []int{2, 4} {
			name := fmt.Sprintf("grass-detail-%02d-%dx.png", index, d)
			pin := write(filepath.Join(directory, "exports", name), next[d])
			pin.Path = "assets/terrain/grass/" + name
			outputs = append(outputs, pin)
		}
		variants = append(variants, map[string]any{"runtime_id": fmt.Sprintf("grass-detail-%02d", index), "index": index, "detail_material": id, "detail_strength": grassDetailStrength, "source_phase_world": grassVariationPhase(index), "outputs": outputs})
		for line := range 9 {
			for col := range 5 - line%2 {
				if 4*(line%4)+(col+3*line)%4 != index {
					continue
				}
				x, y := (142*col+71*(line%2))*2, 40*line*2
				r := image.Rect(x, y, x+192, y+160)
				draw.Draw(before, r, old[2], image.Point{}, draw.Over)
				draw.Draw(after, r, next[2], image.Point{}, draw.Over)
			}
		}
	}
	write(filepath.Join(directory, "candidate-2x.png"), after)
	gameBefore, err := resizeLanczos(before, patchBounds.Dx()*3/4, patchBounds.Dy()*3/4)
	if err != nil {
		t.Fatal(err)
	}
	gameAfter, err := resizeLanczos(after, patchBounds.Dx()*3/4, patchBounds.Dy()*3/4)
	if err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(directory, "baseline-game1.5.png"), gameBefore)
	write(filepath.Join(directory, "candidate-game1.5.png"), gameAfter)
	comparison := image.NewNRGBA(image.Rect(0, 0, gameBefore.Bounds().Dx()*2, gameBefore.Bounds().Dy()))
	draw.Draw(comparison, gameBefore.Bounds(), gameBefore, image.Point{}, draw.Src)
	draw.Draw(comparison, gameAfter.Bounds().Add(image.Pt(gameBefore.Bounds().Dx(), 0)), gameAfter, image.Point{}, draw.Src)
	write(filepath.Join(directory, "comparison-game1.5.png"), comparison)
	recipe := map[string]any{"schema_version": 1, "id": "grass-variations"}
	recipe["detail_materials"] = materialPins
	recipe["variants"] = variants
	recipe["generation_command"] = "go test -tags grassgfx ./tools/art-review -run TestGenerateGrassDetailCandidate -count=1 -v"
	recipe["profile"] = "Sixteen authored material/phase variants: four unchanged phased canonical meadow, six wildflower/pebble/weed and six subtly worn meadow. Detail masters precisely edit canonical grass and project the complete output to768x480 with existing alpha-aware antialiased Lanczos3, retaining relative feature scale. Interior paint is65percent authored detail material plus35percent the existing phased canonical meadow. Canonical source and sampler unchanged through8worldpixels from everyedge; cubic smoothstep8..18 to interior paint. Unmasked416x352 opaque4x overscan, existing whole-image Lanczos3 to208x176, crop4worldpixels then complementary final half-open ownership. All final6worldpixel guards exactly match canonical exports at2x/4x. New botanical/pebble details are imagegen-authored, never drawn procedurally."
	encoded, err := json.MarshalIndent(recipe, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "grass-variations.json"), append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	for id, im := range materials {
		var rgb [3]float64
		for y := range im.Bounds().Dy() {
			for x := range im.Bounds().Dx() {
				c := im.NRGBAAt(x, y)
				rgb[0] += float64(c.R)
				rgb[1] += float64(c.G)
				rgb[2] += float64(c.B)
			}
		}
		n := float64(im.Bounds().Dx() * im.Bounds().Dy())
		t.Logf("%s meanRGB=(%.2f,%.2f,%.2f)", id, rgb[0]/n, rgb[1]/n, rgb[2]/n)
	}
}
