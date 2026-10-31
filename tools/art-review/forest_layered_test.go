package artreview

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func layeredTestRenderer(t *testing.T) *layeredForestRenderer {
	t.Helper()
	ground := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	floor := image.NewNRGBA(ground.Bounds())
	draw.Draw(ground, ground.Bounds(), image.NewUniform(color.NRGBA{100, 110, 40, 255}), image.Point{}, draw.Src)
	draw.Draw(floor, floor.Bounds(), image.NewUniform(color.NRGBA{35, 45, 20, 255}), image.Point{}, draw.Src)
	var atlases []*image.NRGBA
	for range 3 {
		a := image.NewNRGBA(image.Rect(0, 0, 512, 512))
		for i := range 4 {
			r := image.Rect(i%2*256+45, i/2*256+45, i%2*256+210, i/2*256+210)
			draw.Draw(a, r, image.NewUniform(color.NRGBA{30, 65, 20, 255}), image.Point{}, draw.Src)
		}
		atlases = append(atlases, a)
	}
	r, err := newLayeredForestRenderer(ground, floor, atlases)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestLayeredForestCLIRejectsUnsafeOrIncompleteInputs(t *testing.T) {
	for _, tt := range []struct{ name, data string }{
		{"missing atlases", `{"schema_version":1}`},
		{"unknown field", `{"schema_version":1,"alien":true}`},
		{"trailing document", `{"schema_version":1} {}`},
		{"duplicate field", `{"schema_version":1,"schema_version":1,"atlases":[{},{},{}]}`},
		{"bad source", `{"schema_version":1,"atlases":[{},{},{}]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "manifest.json")
			out := filepath.Join(dir, "out")
			if err := os.WriteFile(path, []byte(tt.data), 0o600); err != nil {
				t.Fatal(err)
			}
			err := runLayeredForest(path, out, new(bytes.Buffer))
			if err == nil {
				t.Fatal("invalid manifest accepted")
			}
			if tt.name == "duplicate field" && !strings.Contains(err.Error(), "duplicate JSON key") {
				t.Fatalf("duplicate key not detected before source loading: %v", err)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("invalid input left output: %v", err)
			}
		})
	}
	dir := t.TempDir()
	if err := runLayeredForest("missing", dir, new(bytes.Buffer)); err == nil || !strings.Contains(err.Error(), "must be new") {
		t.Fatal("existing output not protected")
	}
	if err := run([]string{"-forest-layered", "missing", "-out", filepath.Join(dir, "new"), "-forest", "other"}, new(bytes.Buffer)); err == nil {
		t.Fatal("conflicting mode accepted")
	}
}

func TestLayeredForestOwnsWholeCrownsAndSeparateShadows(t *testing.T) {
	r := layeredTestRenderer(t)
	interior, _ := newForestContext("forest", []string{"forest", "forest", "forest", "forest", "forest", "forest"})
	edge, _ := newForestContext("forest", []string{"grass", "grass", "grass", "grass", "grass", "grass"})
	if got := len(r.crowns); got != 12 {
		t.Fatalf("crowns %d, want 12", got)
	}
	if len(r.instances(edge)) >= len(r.instances(interior)) {
		t.Fatal("forest edge did not thin crowns")
	}
	a := r.layers(interior)
	b := r.layers(interior)
	for i := range a {
		if !bytes.Equal(a[i].Pix, b[i].Pix) {
			t.Fatal("nondeterministic forest pixels")
		}
	}
	// Interior canopies must actually cross the owning hex; the old exporter clipped them.
	overhang, shadows := false, false
	for y := range a[1].Bounds().Dy() {
		for x := range a[1].Bounds().Dx() {
			p := connectionPoint{float64(x)/4 - 16, float64(y)/4 - 16}
			if a[1].NRGBAAt(x, y).A > 0 && !connectionInside(sharedConnectionGeometry(), p) {
				overhang = true
			}
			if c := a[2].NRGBAAt(x, y); c.A > 0 {
				shadows = true
				if c.A > 95 {
					t.Fatal("cast shadow is opaque")
				}
			}
		}
	}
	if !overhang || !shadows {
		t.Fatalf("overhang=%v shadows=%v", overhang, shadows)
	}
	// Ground is independent of canopy: an interior floor remains dark under every tree.
	if c := a[0].NRGBAAt(48*4, 40*4); c.R >= 70 {
		t.Fatalf("forest floor not present: %v", c)
	}
}

func TestLayeredForestFloorMasksAndPropFootClearance(t *testing.T) {
	r := layeredTestRenderer(t)
	edge, _ := newForestContext("forest", []string{"grass", "forest", "forest", "forest", "forest", "forest"})
	if r.floorMask(edge, connectionPoint{90, 20}) > .15 {
		t.Fatal("floor does not blend toward NE meadow")
	}
	if r.floorMask(edge, connectionPoint{30, 45}) < .95 {
		t.Fatal("interior floor mask lost coverage")
	}
	for v := 1; v <= 4; v++ {
		layers := r.props(v)
		if layers[0].Bounds() != image.Rect(0, 0, 512, 448) {
			t.Fatal("prop canvas changed")
		}
		// Leave the unit foot point and its immediate eight-pixel neighbourhood clear.
		for y := (56 - 8) * 4; y < (56+8)*4; y++ {
			for x := (64 - 8) * 4; x < (64+8)*4; x++ {
				if layers[0].NRGBAAt(x, y).A != 0 {
					t.Fatalf("prop %d covers unit foot", v)
				}
			}
		}
	}
}

func TestLayeredForestVertexMasksAgreeAcrossForestNeighbours(t *testing.T) {
	r := layeredTestRenderer(t)
	a, _ := newForestContext("forest", []string{"forest", "forest", "forest", "forest", "forest", "grass"})
	b, _ := newForestContext("forest", []string{"forest", "forest", "forest", "forest", "grass", "forest"})
	delta := sharedConnectionGeometry().Neighbors[0]
	for i := 1; i < 10; i++ {
		time := float64(i) / 10
		p := connectionPoint{71 + 25*time, 40 * time}
		if x, y := r.floorMask(a, p), r.floorMask(b, p.sub(delta)); math.Abs(x-y) > 1e-8 {
			t.Fatalf("forest vertex seam at %v: %.9f != %.9f", p, x, y)
		}
	}
}

func TestLayeredForestGroundFiltersBeforeFinalOwnership(t *testing.T) {
	r := layeredTestRenderer(t)
	c, _ := newForestContext("forest", []string{"forest", "forest", "forest", "forest", "forest", "forest"})
	raw := r.layers(c)[0]
	for _, density := range []int{1, 2, 4} {
		tile, err := layeredGroundTile(raw, density)
		if err != nil {
			t.Fatal(err)
		}
		for y := range tile.Bounds().Dy() {
			for x := range tile.Bounds().Dx() {
				inside := connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)})
				want := uint8(0)
				if inside {
					want = 255
				}
				if a := tile.NRGBAAt(x, y).A; a != want {
					t.Fatalf("%dx pixel %d,%d alpha%d want%d: polygon filtered before ownership", density, x, y, a, want)
				}
			}
		}
	}
}
