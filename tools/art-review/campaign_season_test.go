package artreview

// Seasonal forms of the campaign generators (tag campaigngfx). With
// -args -season autumn -season-out <dir>, TestGenerateCampaignShores and
// TestGenerateCampaignTrenches render every installed summer tile of their
// family from the seasonal edits of its sources (season.go) instead of filling
// gaps, and write "<summer file stem><suffix>-<d>x.png" into the new directory.
// The runtime index is not touched: the files are installed beside their
// summer forms by a seasonal install that copies each summer entry.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var campaignSeasonal = regexp.MustCompile(`_(w|au|sp)$`)

// campaignSeasonExport renders, with exportSeason set to season, the seasonal
// form of every summer terrain tile in root's runtime index whose ID starts
// with one of prefixes, and writes it to the new directory out under the
// summer file's stem plus the season suffix. load runs once the season is set
// and returns the renderer, called with the ID without "tile/". It returns the
// number of IDs written.
func campaignSeasonExport(root, season, out string, prefixes []string, load func() (func(id string) (map[int]*image.NRGBA, error), error)) (int, error) {
	suffix := seasonSuffixes[season]
	if suffix == "" {
		return 0, fmt.Errorf("-season %q: want autumn or spring", season)
	}
	if out == "" {
		return 0, fmt.Errorf("-season needs -season-out")
	}
	if _, err := os.Stat(out); err == nil {
		return 0, fmt.Errorf("%s exists", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		return 0, err
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Pack    string `json:"pack"`
			Density int    `json:"density"`
			Path    string `json:"path"`
			Width   int    `json:"width"`
			Height  int    `json:"height"`
		} `json:"assets"`
	}
	if err = json.Unmarshal(b, &index); err != nil {
		return 0, err
	}
	type file struct {
		path          string
		width, height int
	}
	files := map[string]map[int]file{}
	for _, e := range index.Assets {
		if !strings.HasPrefix(e.Pack, "terrain-") || campaignSeasonal.MatchString(e.ID) {
			continue
		}
		for _, p := range prefixes {
			if strings.HasPrefix(e.ID, p) {
				if files[e.ID] == nil {
					files[e.ID] = map[int]file{}
				}
				files[e.ID][e.Density] = file{e.Path, e.Width, e.Height}
				break
			}
		}
	}
	ids := make([]string, 0, len(files))
	for id := range files {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	previous := exportSeason
	exportSeason = season
	defer func() { exportSeason = previous }()
	render, err := load()
	if err != nil {
		return 0, err
	}
	if err = os.MkdirAll(out, 0o755); err != nil {
		return 0, err
	}
	for _, id := range ids {
		tiles, err := render(strings.TrimPrefix(id, "tile/"))
		if err != nil {
			return 0, fmt.Errorf("%s: %w", id, err)
		}
		for d, f := range files[id] {
			im := tiles[d]
			if im == nil {
				return 0, fmt.Errorf("%s: no %dx render", id, d)
			}
			if im.Bounds().Dx() != f.width || im.Bounds().Dy() != f.height {
				return 0, fmt.Errorf("%s@%d: render %dx%d, summer %dx%d", id, d, im.Bounds().Dx(), im.Bounds().Dy(), f.width, f.height)
			}
			tail := fmt.Sprintf("-%dx.png", d)
			name := filepath.Base(f.path)
			if !strings.HasSuffix(name, tail) {
				return 0, fmt.Errorf("%s: unexpected file name %s", id, name)
			}
			var buf bytes.Buffer
			if err = png.Encode(&buf, im); err != nil {
				return 0, err
			}
			if err = os.WriteFile(filepath.Join(out, strings.TrimSuffix(name, tail)+suffix+tail), buf.Bytes(), 0o644); err != nil {
				return 0, err
			}
		}
	}
	return len(ids), nil
}

func TestCampaignSeasonExport(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	index := `{"schema_version":1,"assets":[
{"id":"tile/trench14","pack":"terrain-infrastructure-2x","density":2,"path":"terrain/trench/trench14-2x.png","width":2,"height":1},
{"id":"tile/trench14","pack":"terrain-infrastructure-4x","density":4,"path":"terrain/trench/trench14-4x.png","width":4,"height":2},
{"id":"tile/trench14_w","pack":"terrain-infrastructure-2x","density":2,"path":"terrain/trench/trench14_w-2x.png","width":2,"height":1},
{"id":"tile/trench14_au","pack":"terrain-infrastructure-2x","density":2,"path":"terrain/trench/trench14_au-2x.png","width":2,"height":1},
{"id":"tile/road14","pack":"terrain-infrastructure-2x","density":2,"path":"terrain/road/road14-2x.png","width":2,"height":1}]}`
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), []byte(index), 0o600); err != nil {
		t.Fatal(err)
	}
	var seen []string
	load := func() (func(string) (map[int]*image.NRGBA, error), error) {
		if exportSeason != "autumn" {
			t.Fatalf("inputs load in season %q", exportSeason)
		}
		return func(id string) (map[int]*image.NRGBA, error) {
			seen = append(seen, id)
			out := map[int]*image.NRGBA{}
			for _, d := range []int{2, 4} {
				im := image.NewNRGBA(image.Rect(0, 0, d, d/2))
				im.SetNRGBA(0, 0, color.NRGBA{uint8(d), 1, 2, 255})
				out[d] = im
			}
			return out, nil
		}, nil
	}
	out := filepath.Join(root, "out")
	n, err := campaignSeasonExport(root, "autumn", out, []string{"tile/trench"}, load)
	if err != nil {
		t.Fatal(err)
	}
	if exportSeason != "" {
		t.Fatal("export season left set")
	}
	if n != 1 || len(seen) != 1 || seen[0] != "trench14" {
		t.Fatalf("rendered %d %v, want trench14 alone", n, seen)
	}
	for _, name := range []string{"trench14_au-2x.png", "trench14_au-4x.png"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = campaignSeasonExport(root, "autumn", out, []string{"tile/trench"}, load); err == nil {
		t.Fatal("an existing output directory is accepted")
	}
	if _, err = campaignSeasonExport(root, "winter", filepath.Join(root, "w"), []string{"tile/trench"}, load); err == nil {
		t.Fatal("winter is accepted as a seasonal export")
	}
}
