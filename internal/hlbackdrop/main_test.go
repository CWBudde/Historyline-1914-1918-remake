package hlbackdrop

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

func TestQuality(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for _, q := range []int{1, 49, 50, 75, 90, 95, 100} {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, im, &jpeg.Options{Quality: q}); err != nil {
			t.Fatal(err)
		}
		got, err := quality(buf.Bytes())
		if err != nil || got != q {
			t.Errorf("quality of a q%d JPEG = %d, %v", q, got, err)
		}
	}
	if _, err := quality([]byte{0xFF, 0xD8}); err == nil {
		t.Error("a JPEG without a table gave a quality")
	}
}

// TestCheck walks a levels directory through a missing source, a written
// JPEG, a stale JPEG, a new backdrop and a source outside a level folder.
func TestCheck(t *testing.T) {
	levels := t.TempDir()
	write := func(name string, data []byte) {
		t.Helper()
		p := filepath.Join(levels, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	source := func(name string, c uint8) {
		t.Helper()
		im := image.NewRGBA(image.Rect(0, 0, 24, 16))
		for i := range im.Pix {
			im.Pix[i] = c + uint8(i%7)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, im); err != nil {
			t.Fatal(err)
		}
		write(name, buf.Bytes())
	}
	expect := func(want ...string) {
		t.Helper()
		errs := check(levels)
		if len(errs) != len(want) {
			t.Fatalf("check = %v, want %d errors %q", errs, len(want), want)
		}
		for i, err := range errs {
			if !strings.Contains(err.Error(), want[i]) {
				t.Errorf("check error %q, want it to say %q", err, want[i])
			}
		}
	}

	write("1914-08/map.json", []byte(`{}`))
	write("1914-09/map.json", []byte(`{}`))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 24, 16)), &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	write("1914-08/backdrop.jpg", buf.Bytes())
	expect("no source")

	source("1914-08/backdrop.png", 40)
	expect("is not")
	b, err := find(levels, "1914-08")
	if err != nil {
		t.Fatal(err)
	}
	if _, q, err := writeJPEG(b, 0); err != nil || q != 90 {
		t.Fatalf("writeJPEG = quality %d, %v; want the JPEG's own 90", q, err)
	}
	expect()

	source("1914-08/backdrop.png", 80)
	expect("is not")
	if _, _, err := writeJPEG(b, 0); err != nil {
		t.Fatal(err)
	}
	source("1914-09/backdrop.png", 80)
	expect("no JPEG")
	b, err = find(levels, "1914-09")
	if err != nil {
		t.Fatal(err)
	}
	if _, q, err := writeJPEG(b, 0); err != nil || q != defaultQuality {
		t.Fatalf("writeJPEG = quality %d, %v; want %d for a new backdrop", q, err, defaultQuality)
	}
	expect()

	source("1914-10/backdrop.png", 80)
	expect("not in a level folder")
	if _, err := find(levels, "1914-10"); err == nil {
		t.Error("find gave a backdrop outside a level folder")
	}
}

// TestShipped checks the repository's backdrops against their sources.
func TestShipped(t *testing.T) {
	if testing.Short() {
		t.Skip("decodes every backdrop source")
	}
	for _, err := range check(filepath.Join("..", "..", "levels")) {
		t.Error(err)
	}
}

// TestAudit records WATER (1918-11): its buildings, every setup's units
// and the two-cell ships with both halves on neighbouring water hexes.
func TestAudit(t *testing.T) {
	out := filepath.Join(t.TempDir(), "audit.json")
	var log bytes.Buffer
	if err := Run([]string{"audit", "-levels", filepath.Join("..", "..", "levels"), "-o", out, "1918-11"}, &log, &log); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var audit struct {
		Season    string
		Buildings []auditBuilding
		Setups    map[string]auditSetup
	}
	if err := json.Unmarshal(b, &audit); err != nil {
		t.Fatal(err)
	}
	if audit.Season != "winter" || len(audit.Buildings) != 7 || len(audit.Setups) != 3 {
		t.Fatalf("season %q, %d buildings, %d setups", audit.Season, len(audit.Buildings), len(audit.Setups))
	}
	for mode, s := range audit.Setups {
		if len(s.Units) != 101 || len(s.Ships) != 31 {
			t.Errorf("%s: %d units, %d ships", mode, len(s.Units), len(s.Ships))
		}
		for _, sh := range s.Ships {
			d := hexgrid.Distance(hexgrid.Coord{Line: sh.Head[0] - 1, Col: sh.Head[1] - 1}, hexgrid.Coord{Line: sh.Rear[0] - 1, Col: sh.Rear[1] - 1})
			if d != 1 || !strings.HasPrefix(sh.Terrain[0], "water") || !strings.HasPrefix(sh.Terrain[1], "water") {
				t.Errorf("%s: %s head %v rear %v on %v", mode, sh.ID, sh.Head, sh.Rear, sh.Terrain)
			}
		}
	}
}

// TestOpenSeaWeights weighs the open-sea shift at output scale: full on open
// water, none inside an island, less on a pier hex, and none for white crests.
func TestOpenSeaWeights(t *testing.T) {
	grid := hexgrid.Grid{Width: 9, Lines: 19}
	lines := make([][]level.Hex, grid.Lines)
	for l := range lines {
		lines[l] = make([]level.Hex, grid.Width-l%2)
		for c := range lines[l] {
			lines[l][c].Terrain = "waterSea"
		}
	}
	island := hexgrid.Coord{Line: 9, Col: 4}
	for _, c := range append(grid.Neighbors(island), island) {
		lines[c.Line][c.Col].Terrain = "grass"
	}
	lines[3][2] = level.Hex{Terrain: "water", Infra: level.Infra{IDs: []string{"rBridge36"}}}
	const scale = 1.41
	w, h := maprender.Size(grid)
	w, h = int(float64(w)*scale), int(float64(h)*scale)
	weights := waterWeights(image.Rect(0, 0, w, h), lines, grid, 1/scale, 1/scale)
	at := func(c hexgrid.Coord) float32 {
		x, y := hexgrid.Center(c)
		x, y = x*scale, y*scale
		return weights[int(y)/waterStep*(w/waterStep+1)+int(x)/waterStep]
	}
	if got := at(hexgrid.Coord{Line: 15, Col: 6}); got < 0.99 {
		t.Errorf("open sea weight %.2f, want 1", got)
	}
	if got := at(island); got > 0.01 {
		t.Errorf("island weight %.2f, want 0", got)
	}
	if got := at(hexgrid.Coord{Line: 3, Col: 2}); got > 0.5 {
		t.Errorf("pier weight %.2f, want under 0.5", got)
	}
	if seaShift(80, 80) != 1 || seaShift(255, 80) != 0 {
		t.Errorf("sea shift %.2f at the sea mean, %.2f at white", seaShift(80, 80), seaShift(255, 80))
	}
}
