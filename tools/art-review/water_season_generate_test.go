//go:build campaigngfx

package artreview

// Explicit offline generator:
//
//	go test -tags campaigngfx ./tools/art-review -run TestGenerateSeasonalWater -count=1 -v -args -season autumn -season-out <new dir>
//
// Writes the seasonal forms (suffix _au, _sp) of open water, the open sea, the
// surf, the shoals and the skerries into the new directory, named as their
// installed summer files plus the suffix, for a seasonal install beside them.
// Open water, sea, surf and shoals are the installed summer pixels: the shores,
// lakesides and streams of a season are painted with the summer water
// material, so open water keeps meeting them as one material, and a seasonal
// water colour belongs in the pack grade (terraingrade), as winter's does. The
// skerries are cut as in summer (TestGenerateSeaKit) from the newest seasonal
// edit of the sea kit, water-seakit-<season>-v<N>.png, under the summer kit's
// alpha, so they keep the summer outline; without an edit they are left out.
import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestGenerateSeasonalWater(t *testing.T) {
	if *campaignSeason == "" {
		t.Skip("needs -args -season autumn|spring -season-out <dir>")
	}
	suffix := seasonSuffixes[*campaignSeason]
	if suffix == "" {
		t.Fatalf("-season %q: want autumn or spring", *campaignSeason)
	}
	out := *campaignSeasonOut
	if out == "" {
		t.Fatal("-season needs -season-out")
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatalf("%s exists", out)
	}
	root := filepath.Join("..", "..")
	_, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	copies := []string{"tile/water", "tile/surf"}
	for _, k := range []string{"00", "02", "20", "22", "11", "13", "31", "33"} {
		copies = append(copies, "tile/sea-"+k)
	}
	var skerries []string
	for id := range seaKitOverlays {
		if strings.HasPrefix(id, "shoal") {
			copies = append(copies, "tile/"+id)
		} else {
			skerries = append(skerries, id)
		}
	}
	slices.Sort(copies)
	slices.Sort(skerries)
	if err = os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(id string, d int, data []byte) {
		summer, ok := exported[campaignExportKey(id, d)]
		if !ok {
			t.Fatalf("%s at %dx is not installed", id, d)
		}
		tail := fmt.Sprintf("-%dx.png", d)
		name := filepath.Base(summer.Path)
		if !strings.HasSuffix(name, tail) {
			t.Fatalf("%s: unexpected file name %s", id, name)
		}
		if err := os.WriteFile(filepath.Join(out, strings.TrimSuffix(name, tail)+suffix+tail), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range copies {
		for _, d := range []int{2, 4} {
			e, ok := exported[campaignExportKey(id, d)]
			if !ok {
				t.Fatalf("%s at %dx is not installed", id, d)
			}
			data, err := os.ReadFile(filepath.Join(root, "assets", e.Path))
			if err != nil {
				t.Fatal(err)
			}
			write(id, d, data)
		}
	}
	t.Logf("%s water, sea, surf and shoals: %d IDs as the summer pixels", *campaignSeason, len(copies))
	edit := newestSeasonEdit(t, filepath.Join(root, "art/terrain/water"), "water-seakit", *campaignSeason)
	if edit == "" {
		t.Logf("no %s edit of the sea kit: skerries left out", *campaignSeason)
		return
	}
	summerKit := loadSeaKit(t, filepath.Join(root, "art/terrain/water/water-seakit-shared-v1.png"))
	kit := loadSeaKit(t, edit)
	if kit.Rect != summerKit.Rect {
		t.Fatalf("%s is %v, the summer kit %v", edit, kit.Rect.Size(), summerKit.Rect.Size())
	}
	for i := 3; i < len(kit.Pix); i += 4 {
		kit.Pix[i] = summerKit.Pix[i]
	}
	elements := seaKitElements(t, kit, seaKitSkerries)
	summerElements := seaKitElements(t, summerKit, seaKitSkerries)
	for _, id := range skerries {
		for _, d := range []int{2, 4} {
			data := seaKitTile(t, elements, seaKitOverlays[id], d)
			if !sameAlpha(t, data, seaKitTile(t, summerElements, seaKitOverlays[id], d)) {
				t.Fatalf("%s at %dx: the %s skerry's alpha differs from summer", id, d, *campaignSeason)
			}
			write("tile/"+id, d, data)
		}
	}
	t.Logf("%s skerries from %s: %d IDs", *campaignSeason, filepath.Base(edit), len(skerries))
}

// newestSeasonEdit is the path of the newest <stem>-<season>-v<N>.png in dir,
// or "".
func newestSeasonEdit(t *testing.T, dir, stem, season string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	best, version := "", 0
	for _, e := range entries {
		m := seasonSourceName.FindStringSubmatch(e.Name())
		if m == nil || m[1] != stem || m[2] != season {
			continue
		}
		if v, _ := strconv.Atoi(m[3]); v > version {
			best, version = filepath.Join(dir, e.Name()), v
		}
	}
	return best
}

// sameAlpha reports whether two PNG images have the same size and alpha.
func sameAlpha(t *testing.T, a, b []byte) bool {
	t.Helper()
	decode := func(data []byte) *image.NRGBA {
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		n, ok := im.(*image.NRGBA)
		if !ok {
			t.Fatalf("not NRGBA: %T", im)
		}
		return n
	}
	x, y := decode(a), decode(b)
	if x.Rect != y.Rect {
		return false
	}
	for i := 3; i < len(x.Pix); i += 4 {
		if x.Pix[i] != y.Pix[i] {
			return false
		}
	}
	return true
}
