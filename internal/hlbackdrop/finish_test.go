package hlbackdrop

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A green matching ring can have a different blue component without requiring
// a neutral roof inside it to turn purple. Brightness matching must preserve
// the roof's channel ratios while still adjusting its exposure.
func TestFinishPaintingBrightnessMatchKeepsMaterialColours(t *testing.T) {
	for _, tc := range []struct {
		name, match string
		want        color.NRGBA
	}{
		{"legacy default", "", color.NRGBA{107, 96, 160, 255}},
		{"explicit rgb", "rgb", color.NRGBA{107, 96, 160, 255}},
		{"brightness", "brightness", color.NRGBA{99, 99, 99, 255}},
		{"invalid mode", "green", color.NRGBA{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			base := image.NewNRGBA(image.Rect(0, 0, 87, 100))
			patch := image.NewNRGBA(image.Rect(0, 0, 40, 40))
			ground := color.NRGBA{80, 120, 70, 255}
			for y := range base.Bounds().Dy() {
				for x := range base.Bounds().Dx() {
					base.SetNRGBA(x, y, ground)
				}
			}
			for y := range patch.Bounds().Dy() {
				for x := range patch.Bounds().Dx() {
					patch.SetNRGBA(x, y, color.NRGBA{60, 100, 35, 255})
				}
			}
			for y := 18; y < 22; y++ {
				for x := 18; x < 22; x++ {
					patch.SetNRGBA(x, y, color.NRGBA{80, 80, 80, 255})
				}
			}
			write := func(name string, im image.Image) string {
				t.Helper()
				var data bytes.Buffer
				if err := png.Encode(&data, im); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, name)
				if err := os.WriteFile(path, data.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			}
			source := write("source.png", base)
			painting := write("patch.png", patch)
			finished := filepath.Join(dir, "finished.png")
			data, err := json.Marshal(map[string]any{
				"world": [2]int{2226, 2560}, "source": source, "finished": finished,
				"patches": []map[string]any{{
					"id": "roof", "painting": painting, "match": tc.match,
					"rectangle": [4]int{20, 30, 40, 40},
					"center":    [2]int{20, 20}, "radii": [2]int{8, 16},
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			plan := filepath.Join(dir, "plan.json")
			if err := os.WriteFile(plan, data, 0o644); err != nil {
				t.Fatal(err)
			}
			var log bytes.Buffer
			err = Run([]string{"jpeg", "-levels", filepath.Join("..", "..", "levels"), "-finish", plan, "1916-05"}, &log, &log)
			if tc.match == "green" {
				if err == nil || !strings.Contains(err.Error(), "match") {
					t.Fatalf("invalid matching mode: %v; want matching-mode error", err)
				}
				if _, err := os.Stat(finished); !os.IsNotExist(err) {
					t.Fatalf("invalid matching mode wrote output: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("finish: %v\n%s", err, &log)
			}
			data, err = os.ReadFile(finished)
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if got := color.NRGBAModel.Convert(im.At(40, 50)).(color.NRGBA); got != tc.want {
				t.Errorf("roof colour %v, want %v", got, tc.want)
			}
			if got := color.NRGBAModel.Convert(im.At(0, 0)).(color.NRGBA); got != ground {
				t.Errorf("ground outside mask changed to %v, want %v", got, ground)
			}
		})
	}
}
