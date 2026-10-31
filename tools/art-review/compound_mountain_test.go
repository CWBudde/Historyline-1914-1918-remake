package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type mountainSceneWant struct {
	id, layout, season string
	width, height      int
	tiles              []compoundTile
}

// Literal campaign/runtime contracts, independent of compoundLayoutFor.
func mountainScenes() []mountainSceneWant {
	return []mountainSceneWant{
		{"mountain-summer", "mountain", "summer", 238, 160, []compoundTile{
			{Cell: "N", RuntimeID: "mountN", Origin: connectionPoint{71, 0}},
			{Cell: "W", RuntimeID: "mountW", Origin: connectionPoint{0, 40}},
			{Cell: "E", RuntimeID: "mountE", Origin: connectionPoint{142, 40}},
			{Cell: "S", RuntimeID: "mountS", Origin: connectionPoint{71, 80}},
		}},
		{"mountain-winter", "mountain", "winter", 238, 160, []compoundTile{
			{Cell: "N", RuntimeID: "mountN_w", Origin: connectionPoint{71, 0}},
			{Cell: "W", RuntimeID: "mountW_w", Origin: connectionPoint{0, 40}},
			{Cell: "E", RuntimeID: "mountE_w", Origin: connectionPoint{142, 40}},
			{Cell: "S", RuntimeID: "mountS_w", Origin: connectionPoint{71, 80}},
		}},
		{"mountain-single-summer", "mountain-single", "summer", 96, 80, []compoundTile{
			{Cell: "C", RuntimeID: "mountain", Origin: connectionPoint{0, 0}},
		}},
		{"mountain-single-winter", "mountain-single", "winter", 96, 80, []compoundTile{
			{Cell: "C", RuntimeID: "mountain_w", Origin: connectionPoint{0, 0}},
		}},
	}
}

func TestCompoundMountainParityRuntimeCoordinates(t *testing.T) {
	for _, want := range mountainScenes() {
		for parity := range 2 {
			t.Run(fmt.Sprintf("%s/parity%d", want.id, parity), func(t *testing.T) {
				layout, err := compoundLayoutFor(want.id)
				if err != nil {
					t.Fatal(err)
				}
				mountainCheckRuntimePatch(t, compoundPatchFor(layout, parity), want, parity)
			})
		}
	}
}

func mountainCheckRuntimePatch(t *testing.T, patch compoundPatch, want mountainSceneWant, parity int) {
	t.Helper()
	// Runtime placement is x=142*col+71*(line%2), y=40*line (gui.py).
	// Anchor the first terrain cell at col=0, line=parity, for either layout.
	anchor := connectionPoint{float64(71 * parity), float64(40 * parity)}
	origin := connectionPoint{anchor[0] - want.tiles[0].Origin[0], anchor[1] - want.tiles[0].Origin[1]}
	if patch.Parity != parity || patch.SceneOrigin != origin {
		t.Errorf("runtime parity%d scene origin: %v, want %v", parity, patch.SceneOrigin, origin)
	}
	for _, cell := range patch.Cells {
		line := int(cell.Origin[1] / 40)
		lineParity := (line%2 + 2) % 2
		column := (cell.Origin[0] - float64(71*lineParity)) / 142
		if cell.Origin[1] != float64(line*40) || column != math.Trunc(column) {
			t.Errorf("runtime lattice violation for %s: origin %v gives line=%d col=%g", cell.RuntimeID, cell.Origin, line, column)
		}
		if cell.Compound && cell.RuntimeID == want.tiles[0].RuntimeID && cell.Origin != anchor {
			t.Errorf("runtime parity%d terrain anchor: %v, want %v", parity, cell.Origin, anchor)
		}
	}
}

func TestCompoundMountainCLIActualPNGs(t *testing.T) {
	path, manifest := mountainFixture(t)
	out := filepath.Join(filepath.Dir(path), "bundle")
	if err := run([]string{"-compound", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report compoundReport
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Scenes) != 4 || len(report.Grounds) != 2 {
		t.Fatalf("report scene/ground counts: %d/%d", len(report.Scenes), len(report.Grounds))
	}
	manifestBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mountainCheckSnapshot(t, out, report.ManifestSnapshot, manifestBytes, report.ManifestSHA256)
	for season, source := range report.Grounds {
		want := manifest.Grounds[season]
		if source.connectionMaterial != want || source.Dimensions != (dimensions{384, 320}) {
			t.Fatalf("ground provenance: %+v", source)
		}
		original, err := os.ReadFile(filepath.Join(filepath.Dir(path), want.Source))
		if err != nil {
			t.Fatal(err)
		}
		mountainCheckSnapshot(t, out, source.RetainedSource, original, source.SourceSHA256)
	}
	for i, want := range mountainScenes() {
		t.Run(want.id, func(t *testing.T) {
			scene := report.Scenes[i]
			if scene.ID != want.id {
				t.Fatalf("scene ID: %q", scene.ID)
			}
			mountainCheckLayout(t, scene.compoundLayout, want)
			ref := manifest.Scenes[i].Artwork
			if scene.InputArtwork.connectionMaterial != ref || scene.InputArtwork.Dimensions != (dimensions{1024, 640}) || scene.InputArtwork.RetainedSource != "sources/"+want.id+".png" {
				t.Fatalf("art provenance: %+v", scene.InputArtwork)
			}
			original, err := os.ReadFile(filepath.Join(filepath.Dir(path), ref.Source))
			if err != nil {
				t.Fatal(err)
			}
			mountainCheckSnapshot(t, out, scene.InputArtwork.RetainedSource, original, ref.SourceSHA256)
			wantScale, wantOffset := .2091796875, (connectionPoint{11.9, 13.0625})
			if want.layout == "mountain-single" {
				wantScale, wantOffset = .084375, (connectionPoint{4.8, 13})
			}
			if scene.Transform.Source != (dimensions{1024, 640}) || math.Abs(scene.Transform.Scale-wantScale) > 1e-12 || math.Abs(scene.Transform.Offset[0]-wantOffset[0]) > 1e-12 || math.Abs(scene.Transform.Offset[1]-wantOffset[1]) > 1e-12 || scene.Guard.GuardedOpaque != 0 {
				t.Fatalf("fit/guard: %+v %+v", scene.Transform, scene.Guard)
			}
			if len(scene.Patches) != 2 {
				t.Fatalf("runtime parity count: %d", len(scene.Patches))
			}
			for parity := range 2 {
				mountainCheckRuntimePatch(t, scene.Patches[parity], want, parity)
			}
			for _, density := range []int{2, 4} {
				full := mountainLoadPNG(t, out, fmt.Sprintf("%s-scene-%dx.png", want.id, density))
				if full.Bounds() != image.Rect(0, 0, want.width*density, want.height*density) {
					t.Fatalf("scene bounds: %v", full.Bounds())
				}
				groundID := "grass"
				if want.season == "winter" {
					groundID = "grass_w"
				}
				ground := mountainLoadPNG(t, out, fmt.Sprintf("ground/%s-%dx.png", groundID, density))
				assembled := image.NewNRGBA(full.Bounds())
				base := image.NewNRGBA(full.Bounds())
				tiles := map[string]*image.NRGBA{}
				for _, tile := range want.tiles {
					im := mountainLoadPNG(t, out, fmt.Sprintf("tiles/%s-%dx.png", tile.RuntimeID, density))
					if im.Bounds() != image.Rect(0, 0, 96*density, 80*density) {
						t.Fatalf("tile %s bounds: %v", tile.RuntimeID, im.Bounds())
					}
					for y := range im.Bounds().Dy() {
						for x := range im.Bounds().Dx() {
							c := im.NRGBAAt(x, y)
							inside := mountainOwned(x, y, density)
							if (inside && c.A != 255) || (!inside && c != (color.NRGBA{})) {
								t.Fatalf("%s ownership at %d,%d: %v inside=%v", tile.RuntimeID, x, y, c, inside)
							}
						}
					}
					tiles[tile.RuntimeID] = im
					at := image.Pt(int(tile.Origin[0])*density, int(tile.Origin[1])*density)
					draw.Draw(assembled, im.Bounds().Add(at), im, image.Point{}, draw.Over)
					draw.Draw(base, ground.Bounds().Add(at), ground, image.Point{}, draw.Over)
				}
				mountainEqualPNG(t, assembled, full, "ordinary Over assembly")
				mountainEqualPNG(t, mountainLoadPNG(t, out, fmt.Sprintf("%s-assembled-%dx.png", want.id, density)), full, "saved assembly")
				if bytes.Equal(full.Pix, base.Pix) {
					t.Fatal("artwork missing from saved scene")
				}
				for parity := range 2 {
					patch := mountainLoadPNG(t, out, fmt.Sprintf("%s-patch-parity%d-%dx.png", want.id, parity, density))
					mountainEqualPNG(t, patch, mountainSurround(want, ground, tiles, density, parity), fmt.Sprintf("parity%d surround", parity))
					if want.layout == "mountain-single" {
						material := mountainLoadPNG(t, filepath.Dir(path), manifest.Grounds[want.season].Source)
						for _, p := range []image.Point{{48, 1}, {1, 40}, {94, 40}, {48, 78}} {
							x, y := p.X*density, p.Y*density
							c := full.NRGBAAt(x, y)
							if c != ground.NRGBAAt(x, y) {
								t.Fatalf("single quiet ground changed at %v, %dx", p, density)
							}
							if density == 4 {
								// 4x ground is sampled directly; verify its saved phase at
								// the independent runtime world coordinate for each parity.
								world := connectionPoint{float64(71*parity) + (float64(x)+.5)/4, float64(40*parity) + (float64(y)+.5)/4}
								sample := sampleConnectionMaterial
								if want.season == "summer" {
									sample = sampleGrassMaterial
								}
								if c != sample(material, world) {
									t.Fatalf("single runtime ground phase mismatch at %v parity%d", p, parity)
								}
							}
						}
					}
				}
			}
		})
	}
	for _, output := range report.Exports {
		b, err := os.ReadFile(filepath.Join(out, output.Path))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		im := mountainLoadPNG(t, out, output.Path)
		if hex.EncodeToString(sum[:]) != output.SHA256 || output.Dimensions != (dimensions{im.Bounds().Dx(), im.Bounds().Dy()}) {
			t.Fatalf("export pin/dimensions: %+v", output)
		}
	}
}

func TestCompoundMountainLayoutsAndGuardBoundary(t *testing.T) {
	for _, want := range mountainScenes() {
		t.Run(want.id, func(t *testing.T) {
			l, err := compoundLayoutFor(want.id)
			if err != nil {
				t.Fatal(err)
			}
			mountainCheckLayout(t, l, want)
			x, join := 48, connectionPoint{48, 40}
			boundary := []connectionSegment{
				{connectionPoint{25, 0}, connectionPoint{71, 0}},
				{connectionPoint{71, 0}, connectionPoint{96, 40}},
				{connectionPoint{96, 40}, connectionPoint{71, 80}},
				{connectionPoint{71, 80}, connectionPoint{25, 80}},
				{connectionPoint{25, 80}, connectionPoint{0, 40}},
				{connectionPoint{0, 40}, connectionPoint{25, 0}},
			}
			if want.layout == "mountain" {
				x, join = 119, connectionPoint{83.5, 60}
				boundary = []connectionSegment{
					{connectionPoint{96, 0}, connectionPoint{142, 0}},
					{connectionPoint{142, 0}, connectionPoint{167, 40}},
					{connectionPoint{71, 40}, connectionPoint{96, 0}},
					{connectionPoint{25, 40}, connectionPoint{71, 40}},
					{connectionPoint{71, 120}, connectionPoint{25, 120}},
					{connectionPoint{25, 120}, connectionPoint{0, 80}},
					{connectionPoint{0, 80}, connectionPoint{25, 40}},
					{connectionPoint{167, 40}, connectionPoint{213, 40}},
					{connectionPoint{213, 40}, connectionPoint{238, 80}},
					{connectionPoint{238, 80}, connectionPoint{213, 120}},
					{connectionPoint{213, 120}, connectionPoint{167, 120}},
					{connectionPoint{167, 120}, connectionPoint{142, 160}},
					{connectionPoint{142, 160}, connectionPoint{96, 160}},
					{connectionPoint{96, 160}, connectionPoint{71, 120}},
				}
			}
			if !reflect.DeepEqual(compoundBoundary(l), boundary) || compoundDistance(l, join) < 15 {
				t.Fatalf("union boundary: %v, join distance %g", compoundBoundary(l), compoundDistance(l, join))
			}
			if d := compoundDistance(l, connectionPoint{float64(x), 1}); d != 1 {
				t.Fatalf("top boundary distance: %g", d)
			}
			raw := image.NewNRGBA(image.Rect(0, 0, want.width*4, want.height*4))
			raw.SetNRGBA(x*4, 4, color.NRGBA{120, 80, 50, 255})
			if _, _, err := compoundGuard(raw, l); err == nil {
				t.Fatal("opaque terrain at external boundary accepted")
			}
			raw.SetNRGBA(x*4, 4, color.NRGBA{})
			raw.SetNRGBA(int(join[0]*4), int(join[1]*4), color.NRGBA{120, 80, 50, 255})
			guarded, _, err := compoundGuard(raw, l)
			if err != nil || !bytes.Equal(raw.Pix, guarded.Pix) {
				t.Fatalf("interior terrain changed: %v", err)
			}
		})
	}
}

func TestCompoundMountainBoundedManifestAndLoadedPins(t *testing.T) {
	path, m := mountainFixture(t)
	for _, id := range []string{"depot-summer", "depot-winter", "hq-summer", "hq-winter", "factory-winter"} {
		m.Scenes = append(m.Scenes, compoundSceneInput{ID: id, Artwork: m.Scenes[0].Artwork, Inset: .9})
	}
	compoundTestWriteManifest(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Manifest.Scenes) != 9 || len(in.Artwork) != 9 {
		t.Fatal("all nine fixed scenes were not loaded")
	}
	for i, want := range mountainScenes() {
		src := in.Artwork[want.id]
		original, err := os.ReadFile(filepath.Join(filepath.Dir(path), m.Scenes[i].Artwork.Source))
		if err != nil {
			t.Fatal(err)
		}
		if src.Ref != m.Scenes[i].Artwork || !bytes.Equal(src.Bytes, original) || src.Image.Bounds() != image.Rect(0, 0, 1024, 640) {
			t.Fatalf("loaded source changed for %s", want.id)
		}
	}
	if !bytes.Equal(in.Artwork["mountain-summer"].Bytes, in.Artwork["mountain-single-summer"].Bytes) || !bytes.Equal(in.Artwork["mountain-winter"].Bytes, in.Artwork["mountain-single-winter"].Bytes) {
		t.Fatal("shared seasonal master was changed")
	}
	for _, tc := range []struct{ name, id, errorPart string }{
		{"thirteenth", "mountain-summer", "one to twelve"},
		{"duplicate", "mountain-summer", "duplicate compound scene"},
		{"unsupported-single", "mountain-single-other", "unsupported compound scene"},
		{"unsupported-season", "mountain-autumn", "unsupported compound scene"},
		{"unsupported-alias", "mountN-summer", "unsupported compound scene"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := m
			bad.Scenes = append([]compoundSceneInput(nil), m.Scenes...)
			if tc.name == "thirteenth" {
				for range 4 {
					bad.Scenes = append(bad.Scenes, compoundSceneInput{ID: tc.id, Artwork: m.Scenes[0].Artwork, Inset: .9})
				}
			} else {
				bad.Scenes[1].ID = tc.id
			}
			compoundTestWriteManifest(t, path, bad)
			out := filepath.Join(filepath.Dir(path), tc.name)
			err := run([]string{"-compound", path, "-out", out}, io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.errorPart) {
				t.Fatalf("expected %q: %v", tc.errorPart, err)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("invalid input created output")
			}
		})
	}
}

func mountainCheckLayout(t *testing.T, got compoundLayout, want mountainSceneWant) {
	t.Helper()
	if got.Layout != want.layout || got.Season != want.season || got.Width != want.width || got.Height != want.height || got.EntryCell != "" || len(got.Tiles) != len(want.tiles) {
		t.Fatalf("layout: %+v; want %+v with empty entry", got, want)
	}
	for i, tile := range got.Tiles {
		tile.Outputs = nil
		if !reflect.DeepEqual(tile, want.tiles[i]) {
			t.Fatalf("tile %d: %+v, want %+v", i, tile, want.tiles[i])
		}
	}
}

func mountainFixture(t *testing.T) (string, compoundManifest) {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, im *image.NRGBA) connectionMaterial {
		var data bytes.Buffer
		if err := png.Encode(&data, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data.Bytes())
		return connectionMaterial{name, hex.EncodeToString(sum[:])}
	}
	m := compoundManifest{SchemaVersion: 1, Grounds: map[string]connectionMaterial{}}
	artwork := map[string]connectionMaterial{}
	for i, season := range []string{"summer", "winter"} {
		ground := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := range 320 {
			for x := range 384 {
				ground.SetNRGBA(x, y, color.NRGBA{uint8(x*3 + y + i*50), uint8(x + y*7), uint8(x ^ y), 255})
			}
		}
		m.Grounds[season] = write("ground-"+season+".png", ground)
		art := image.NewNRGBA(image.Rect(0, 0, 1024, 640))
		for y := 285; y < 355; y++ {
			for x := 450; x < 574; x++ {
				art.SetNRGBA(x, y, color.NRGBA{uint8(150 + i*60), uint8(x), uint8(y), 255})
			}
		}
		artwork[season] = write("master-"+season+".png", art)
	}
	for _, scene := range mountainScenes() {
		m.Scenes = append(m.Scenes, compoundSceneInput{ID: scene.id, Artwork: artwork[scene.season], Inset: .9})
	}
	path := filepath.Join(dir, "manifest.json")
	compoundTestWriteManifest(t, path, m)
	return path, m
}

func mountainCheckSnapshot(t *testing.T, out, path string, want []byte, hash string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(out, path))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(got)
	if !bytes.Equal(got, want) || hex.EncodeToString(sum[:]) != hash {
		t.Fatalf("retained source/pin mismatch: %s", path)
	}
}

func mountainLoadPNG(t *testing.T, out, name string) *image.NRGBA {
	t.Helper()
	im, err := loadPNG(filepath.Join(out, name))
	if err != nil {
		t.Fatal(err)
	}
	return im
}

func mountainEqualPNG(t *testing.T, got, want *image.NRGBA, context string) {
	t.Helper()
	if got.Bounds() != want.Bounds() || !bytes.Equal(got.Pix, want.Pix) {
		t.Fatalf("%s differs: %v versus %v", context, got.Bounds(), want.Bounds())
	}
}

// Integer half-plane ownership for the literal hex (25,0),(71,0),(96,40),
// (71,80),(25,80),(0,40), sampled at pixel centers. Left/top ties belong here.
func mountainOwned(x, y, density int) bool {
	xx, yy, d := 2*x+1, 2*y+1, 2*density
	return yy >= 0 && yy < 80*d &&
		40*xx+25*yy >= 1000*d && 40*xx-25*yy < 2840*d &&
		40*xx+25*yy < 4840*d && 40*xx-25*yy >= -1000*d
}

func mountainSurround(want mountainSceneWant, ground *image.NRGBA, tiles map[string]*image.NRGBA, density, parity int) *image.NRGBA {
	// First cell at runtime col=0, line=parity, derived from the world formula.
	offset := image.Pt(71*parity-int(want.tiles[0].Origin[0]), 40*parity-int(want.tiles[0].Origin[1]))
	cells := map[image.Point]string{}
	for _, tile := range want.tiles {
		cells[image.Pt(int(tile.Origin[0]), int(tile.Origin[1])).Add(offset)] = tile.RuntimeID
	}
	for _, tile := range want.tiles {
		p := image.Pt(int(tile.Origin[0]), int(tile.Origin[1])).Add(offset)
		for _, n := range []image.Point{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}} {
			if _, exists := cells[p.Add(n)]; !exists {
				cells[p.Add(n)] = ""
			}
		}
	}
	minX, minY, maxX, maxY := 1000, 1000, -1000, -1000
	for p := range cells {
		minX, minY = min(minX, p.X), min(minY, p.Y)
		maxX, maxY = max(maxX, p.X+96), max(maxY, p.Y+80)
	}
	im := image.NewNRGBA(image.Rect(0, 0, (maxX-minX)*density, (maxY-minY)*density))
	for p := range cells {
		at := image.Pt((p.X-minX)*density, (p.Y-minY)*density)
		draw.Draw(im, ground.Bounds().Add(at), ground, image.Point{}, draw.Over)
	}
	for p, id := range cells {
		if id != "" {
			at := image.Pt((p.X-minX)*density, (p.Y-minY)*density)
			draw.Draw(im, tiles[id].Bounds().Add(at), tiles[id], image.Point{}, draw.Over)
		}
	}
	return im
}
