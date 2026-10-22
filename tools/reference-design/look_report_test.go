package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func TestLookCellsUseEveryCleanTerrainCell(t *testing.T) {
	p := &project{excluded: map[int]bool{2: true}, level: &level.Level{Lines: [][]level.Hex{
		{{Terrain: "grass"}, {Terrain: "grass", Infra: level.Infra{IDs: []string{"road14"}}}, {Terrain: "grass"}},
		{{Terrain: "grass"}, {Terrain: "grass"}},
		{{Terrain: "forest"}, {Terrain: "grass", Content: &level.Content{}}, {Terrain: "grass"}},
	}}}
	f := family{Kind: "terrain", RuntimeIDs: []string{"grass"}, Exemplars: []string{"r1c1"}}
	want := []cell{{1, 1}, {1, 3}, {3, 3}}
	if got := p.lookCells(f); !reflect.DeepEqual(got, want) {
		t.Fatalf("clean cells %v, want%v", got, want)
	}
	f.Kind = "infra"
	if got := p.lookCells(f); len(got) != 0 {
		t.Fatalf("infrastructure measured as clean terrain: %v", got)
	}
}

func TestLookReportRecordsMeasurementProvenanceWithoutAcceptanceClaim(t *testing.T) {
	o := comparisonOptions(t)
	// comparisonOptions deliberately uses translucent pixels for alpha tests.
	// This test needs an opaque registered reference with measurable samples.
	refPath := filepath.Join(o.root, o.ref)
	refBytes, err := os.ReadFile(refPath)
	if err != nil {
		t.Fatal(err)
	}
	var ref reference
	if err := json.Unmarshal(refBytes, &ref); err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(image.Rect(0, 0, ref.Image.Width, ref.Image.Height))
	draw.Draw(im, im.Bounds(), image.White, image.Point{}, draw.Src)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, im); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(o.root, ref.Image.Path), encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	ref.Image.SHA256 = fmt.Sprintf("%x", sha256.Sum256(encoded.Bytes()))
	refBytes, err = json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(refPath, refBytes, 0644); err != nil {
		t.Fatal(err)
	}
	o.compare = ""
	o.lookOut = filepath.Join(t.TempDir(), "look")
	if err := run(o); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(o.lookOut, "look.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report lookReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Schema != 1 || report.Accepted || len(report.Remaining) == 0 {
		t.Fatalf("false completion claim: %+v", report)
	}
	if len(report.LevelSHA256) != 64 || len(report.AssetsSHA256) != 64 {
		t.Fatal("missing level/catalogue provenance")
	}
	if len(report.Sources) != 3 {
		t.Fatalf("%d sources, want render/reference/original", len(report.Sources))
	}
	for _, source := range report.Sources {
		if source.ID == "original" {
			if source.Available || source.Reason == "" {
				t.Fatalf("original measurements fabricated: %+v", source)
			}
			continue
		}
		if !source.Available || source.Whole == nil || source.Whole.Samples == 0 || len(source.PixelsSHA256) != 64 {
			t.Fatalf("missing source measurement: %+v", source)
		}
		if len(source.Families) != 3 {
			t.Fatalf("%s: got%dterrainfamilies", source.ID, len(source.Families))
		}
	}
}

func TestLookReportMissingReferenceIsUnavailable(t *testing.T) {
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
		if source.ID == "reference" && (source.Available || source.Reason == "") {
			t.Fatalf("missing reference reported as measured: %+v", source)
		}
	}
}

func TestLookUnavailableColourIsNullWithReason(t *testing.T) {
	for _, test := range []string{"no clean cells", "transparent reference"} {
		t.Run(test, func(t *testing.T) {
			p := loadFixture(t)
			im := image.NewNRGBA(image.Rect(0, 0, 380, 200))
			if test == "no clean cells" {
				draw.Draw(im, im.Bounds(), image.White, image.Point{}, draw.Src)
				for l := range p.level.Lines {
					for c := range p.level.Lines[l] {
						p.level.Lines[l][c].Infra.IDs = []string{"road14"}
					}
				}
			}
			source := p.measureLookSource("reference", im)
			raw, err := json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Families []struct {
					Colour     json.RawMessage `json:"colour"`
					Reason     string          `json:"colour_unavailable_reason"`
					Statistics lookMetrics     `json:"statistics"`
				}
			}
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			for _, f := range decoded.Families {
				if string(f.Colour) != "null" || f.Reason != "insufficient_samples" || f.Statistics.Samples != 0 {
					t.Fatalf("missing colour fabricated as%s (reason%q, samples%d)", f.Colour, f.Reason, f.Statistics.Samples)
				}
			}
		})
	}
}
