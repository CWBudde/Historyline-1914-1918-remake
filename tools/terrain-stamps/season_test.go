package terrainstamps

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A seasonal export takes the newest edit beside a summer source, keeps the
// summer geometry, suffixes IDs and files, carries the transparent companion
// of an edited standing stamp and leaves out stamps without an edit.
func TestSeasonExport(t *testing.T) {
	root, m := fixture(t)
	summer := filepath.Join(root, "knoll-summer-v1.png")
	sum := writeFlat(t, summer, 40, 24, color.NRGBA{80, 100, 20, 255})
	writeFlat(t, filepath.Join(root, "knoll-autumn-v1.png"), 40, 24, color.NRGBA{1, 2, 3, 255})
	writeFlat(t, filepath.Join(root, "knoll-autumn-v2.png"), 40, 24, color.NRGBA{150, 110, 40, 255})
	writeFlat(t, filepath.Join(root, "knoll-winter-v9.png"), 40, 24, color.NRGBA{250, 250, 250, 255})
	m.Assets = append(m.Assets,
		asset{ID: "standing/stamp/knoll/1/1", Source: summer, SourceSHA256: sum, Crop: [4]int{4, 0, 32, 24}, Canvas: [2]int{8, 6}, Anchor: [2]float64{4, 3}, Layer: "standing", Art: "terrain/mountain", Path: "terrain/mountain/knoll-standing"},
		asset{ID: "stamp/knoll/1/1", Transparent: true, Canvas: [2]int{8, 6}, Anchor: [2]float64{4, 3}, Layer: "ground", Art: "terrain/mountain", Path: "terrain/mountain/knoll-ground"},
		asset{ID: "stamp/crater/1/1", Transparent: true, Canvas: [2]int{8, 6}, Anchor: [2]float64{4, 3}, Layer: "ground", Art: "terrain/ground", Path: "terrain/ground/crater"},
	)
	out := filepath.Join(root, "out")
	if err := Run([]string{"-manifest", writeManifest(t, root, m), "-out", out, "-season", "autumn"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	var entries []indexEntry
	readJSON(t, filepath.Join(out, "index-entries.json"), &entries)
	ids := map[string]int{}
	for _, e := range entries {
		ids[e.ID]++
		if !strings.HasSuffix(e.Path, "_au-2x.png") && !strings.HasSuffix(e.Path, "_au-4x.png") {
			t.Fatalf("path not suffixed: %s", e.Path)
		}
		if e.Width != 8*e.Density || e.Height != 6*e.Density || e.Anchor != [2]float64{4, 3} {
			t.Fatalf("geometry changed: %+v", e)
		}
	}
	if len(ids) != 2 || ids["standing/stamp/knoll/1/1_au"] != 2 || ids["stamp/knoll/1/1_au"] != 2 {
		t.Fatalf("exported IDs %v, want the edited knoll and its companion", ids)
	}
	f, err := os.Open(filepath.Join(out, "terrain/mountain/knoll-standing_au-2x.png"))
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(f)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	if p := color.NRGBAModel.Convert(im.At(8, 6)).(color.NRGBA); p != (color.NRGBA{150, 110, 40, 255}) {
		t.Fatalf("pixel %v, want the newest autumn edit", p)
	}
	var report struct {
		Season        string            `json:"season"`
		IDSuffix      string            `json:"id_suffix"`
		SeasonSources map[string]string `json:"season_sources"`
	}
	readJSON(t, filepath.Join(out, "report.json"), &report)
	if report.Season != "autumn" || report.IDSuffix != "_au" || !strings.HasSuffix(report.SeasonSources[summer], "knoll-autumn-v2.png") {
		t.Fatalf("report %+v", report)
	}
}

func TestSeasonExportRejects(t *testing.T) {
	for _, tc := range []struct {
		name, want, edit string
		w, h             int
		pin              string
		season           string
	}{
		{"no edit", "no autumn edit", "", 0, 0, "", "autumn"},
		{"geometry", "geometry", "knoll-autumn-v1.png", 40, 25, "", "autumn"},
		{"pin", "SHA256", "knoll-autumn-v1.png", 40, 24, strings.Repeat("0", 64), "autumn"},
		{"season", "autumn or spring", "knoll-autumn-v1.png", 40, 24, "", "winter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, m := fixture(t)
			summer := filepath.Join(root, "knoll-summer-v1.png")
			sum := writeFlat(t, summer, 40, 24, color.NRGBA{80, 100, 20, 255})
			if tc.edit != "" {
				writeFlat(t, filepath.Join(root, tc.edit), tc.w, tc.h, color.NRGBA{150, 110, 40, 255})
			}
			if tc.pin != "" {
				sum = tc.pin
			}
			m.Assets = []asset{{ID: "standing/stamp/knoll/1/1", Source: summer, SourceSHA256: sum, Crop: [4]int{4, 0, 32, 24}, Canvas: [2]int{8, 6}, Anchor: [2]float64{4, 3}, Layer: "standing", Art: "terrain/mountain", Path: "terrain/mountain/knoll-standing"}}
			out := filepath.Join(root, "out")
			err := Run([]string{"-manifest", writeManifest(t, root, m), "-out", out, "-season", tc.season}, io.Discard, io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("rejected export published output")
			}
		})
	}
}

func writeFlat(t *testing.T, p string, w, h int, c color.NRGBA) string {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.SetNRGBA(x, y, c)
		}
	}
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
