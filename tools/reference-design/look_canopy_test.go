package referencedesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func TestCanopyPlaneAveragesAlphaWithoutDiscardingEdges(t *testing.T) {
	im := image.NewAlpha(image.Rect(7, 9, 11, 13))
	im.SetAlpha(7, 9, color.Alpha{255})
	for y := 9; y < 11; y++ {
		for x := 9; x < 11; x++ {
			im.SetAlpha(x, y, color.Alpha{128})
		}
	}
	p, err := averageCanopy(im, 2)
	if err != nil {
		t.Fatal(err)
	}
	if p.w != 2 || p.h != 2 || p.alpha[0] != .25 || p.alpha[1] != 128.0/255 || p.alpha[2] != 0 || p.alpha[3] != 0 {
		t.Fatalf("averaged plane %+v", p)
	}
	for _, d := range []int{0, 1, 3} {
		if _, err := averageCanopy(im, d); err == nil {
			t.Errorf("invalid density%d accepted", d)
		}
	}
	if _, err := averageCanopy(image.NewAlpha(image.Rect(0, 0, 3, 4)), 2); err == nil {
		t.Fatal("partial block accepted")
	}
}

func TestCanopyCoverageUsesEveryCleanCellAndNamesUnsupported(t *testing.T) {
	p := loadFixture(t)
	plane := &canopyPlane{w: 380, h: 200, alpha: make([]float64, 380*200)}
	for i := range plane.alpha {
		plane.alpha[i] = .5
	}
	got := p.measureCanopy(plane, nil)
	if !got.Available {
		t.Fatal(got.Reason)
	}
	for _, group := range got.Groups {
		if group.Coverage.Value == nil || math.Abs(*group.Coverage.Value-.5) > 1e-12 {
			t.Errorf("group %+v", group)
		}
	}
	if len(got.Cells) != 4 {
		t.Fatalf("got%d clean forest/grass cells, want4", len(got.Cells))
	}
	unsupported := []hexgrid.Coord{{Line: 0, Col: 1}}
	got = p.measureCanopy(plane, unsupported)
	if len(got.Unsupported) != 1 || got.Unsupported[0] != "r1c2" {
		t.Fatalf("unsupported cells %v", got.Unsupported)
	}
	for _, c := range got.Cells {
		if c.Cell == "r1c2" && (c.Coverage.Value != nil || c.Coverage.Reason != "canopy_layer_unavailable") {
			t.Fatalf("unsupported winter treated as empty canopy: %+v", c)
		}
	}
}

func TestCanopyInteriorRequiresSixForestNeighbours(t *testing.T) {
	p := &project{excluded: map[int]bool{}, level: &level.Level{Lines: make([][]level.Hex, 5)}}
	for l := range p.level.Lines {
		p.level.Lines[l] = make([]level.Hex, 4-l%2)
		for c := range p.level.Lines[l] {
			p.level.Lines[l][c].Terrain = "forest"
		}
	}
	center := cell{3, 2}
	if got := p.canopyGroup(center); got != "forest_interior" {
		t.Fatal(got)
	}
	n := hexgrid.Around(hexgrid.Coord{Line: 2, Col: 1})[0]
	p.level.Lines[n.Line][n.Col].Terrain = "grass"
	if got := p.canopyGroup(center); got != "forest_frontier" {
		t.Fatal(got)
	}
	if got := p.canopyGroup(cell{1, 1}); got != "forest_frontier" {
		t.Fatal("map edge", got)
	}
}

func TestLookReportIncludesCanopyEvidenceAndMask(t *testing.T) {
	p := loadFixture(t)
	out := t.TempDir()
	if err := p.writeLookReport(out); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "look.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report lookReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	for _, source := range report.Sources {
		if source.Canopy == nil {
			t.Fatalf("%s lacks canopy availability", source.ID)
		}
		if source.ID != "render" {
			if source.Canopy.Available || source.Canopy.Reason == "" {
				t.Fatalf("unannotated canopy fabricated for%s", source.ID)
			}
			continue
		}
		c := source.Canopy
		if !c.Available || len(c.Cells) != 4 || c.MaskFile == "" {
			t.Fatalf("render canopy evidence missing: %+v", c)
		}
		b, err := os.ReadFile(filepath.Join(out, c.MaskFile))
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != c.MaskSHA256 {
			t.Fatal("canopy mask hash mismatch")
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if im.Bounds() != image.Rect(0, 0, 380, 200) {
			t.Fatal(im.Bounds())
		}
	}
}
