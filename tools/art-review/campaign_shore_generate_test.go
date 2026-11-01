//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignShores -count=1 -v
// Every derived shore tile the levels select (mapcompose.ShoreShape) that the
// runtime index lacks is rendered with the shore-terrain-v2 materials (winter
// forms, suffix _w, with shore-terrain-winter-v1), written to
// assets/terrain/lakeside/ and added to assets/index.json; run
// `go run ./cmd/hl art index` afterwards. With -args -season autumn -season-out
// <dir> it writes the seasonal form of every installed summer tile instead
// (campaign_season_test.go).
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

var campaignShoreMouths = flag.String("shore-mouths", "", "export only these new shoreline mouth tile names, comma-separated")

func TestGenerateCampaignShores(t *testing.T) {
	root := filepath.Join("..", "..")
	if *campaignShoreMouths != "" {
		generateShoreMouths(t, root, strings.Split(*campaignShoreMouths, ","))
		return
	}
	manifest := filepath.Join(root, "art/pipeline/shore-terrain-v2.json")
	if *campaignSeason != "" {
		n, err := campaignSeasonExport(root, *campaignSeason, *campaignSeasonOut, []string{"tile/lakeside", "tile/shore-"}, func() (func(string) (map[int]*image.NRGBA, error), error) {
			if err := seasonPreflight(manifest); err != nil {
				return nil, err
			}
			in, err := loadShoreInputs(manifest)
			if err != nil {
				return nil, err
			}
			return func(id string) (map[int]*image.NRGBA, error) {
				return campaignShoreTiles(in, shoreInputs{}, id)
			}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s shore tiles: %d", *campaignSeason, n)
		return
	}
	in, err := loadShoreInputs(manifest)
	if err != nil {
		t.Fatal(err)
	}
	winter, err := loadShoreInputs(filepath.Join(root, "art/pipeline/shore-terrain-winter-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	indexBefore, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	// The installed roster tiles must come from these inputs, or a derived tile
	// would not join its lakeside neighbours.
	for _, v := range shoreRoster() {
		tiles, err := exportShoreTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		e := exported[campaignExportKey("tile/"+v.ID, 4)]
		b, err := os.ReadFile(filepath.Join(root, "assets", e.Path))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(im.(*image.NRGBA).Pix, tiles[4].Pix) {
			t.Fatalf("installed %s differs from the shore-terrain-v2 render", v.ID)
		}
	}
	has := func(id string) bool {
		_, a := exported[campaignExportKey(id, 2)]
		_, b := exported[campaignExportKey(id, 4)]
		return a && b
	}
	// A winter mouth shows up as a gap only once its winter shore exists, so
	// export until the levels select no absent shore.
	written := 0
	for {
		gaps, err := campaignShores(has)
		if err != nil {
			t.Fatal(err)
		}
		if len(gaps) == 0 {
			break
		}
		ids := make([]string, 0, len(gaps))
		for id := range gaps {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		t.Logf("absent shore tiles: %d", len(ids))
		for _, id := range ids {
			tiles, err := campaignShoreTiles(in, winter, id)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range []int{2, 4} {
				path := fmt.Sprintf("assets/terrain/lakeside/%s-%dx.png", id, d)
				var b bytes.Buffer
				if err := png.Encode(&b, tiles[d]); err != nil {
					t.Fatal(err)
				}
				full := filepath.Join(root, path)
				if _, err := os.Stat(full); err == nil {
					t.Fatalf("unindexed export exists: %s", path)
				}
				if err := os.WriteFile(full, b.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(b.Bytes())
				e := campaignEntry{ID: mapcompose.TileID(id), Pack: fmt.Sprintf("terrain-backgrounds-%dx", d), Density: d, Path: strings.TrimPrefix(path, "assets/"), SHA256: hex.EncodeToString(sum[:]), Width: tiles[d].Bounds().Dx(), Height: tiles[d].Bounds().Dy(), Anchor: [2]float64{48, 40}, Art: "terrain/lakeside"}
				raw, err := json.Marshal(e)
				if err != nil {
					t.Fatal(err)
				}
				entries = append(entries, raw)
				exported[campaignExportKey(e.ID, d)] = e
				written++
			}
		}
	}
	if written == 0 {
		return
	}
	name := filepath.Join(root, "assets/index.json")
	current, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(current, indexBefore) {
		t.Fatal("runtime index changed during export")
	}
	var out bytes.Buffer
	out.WriteString("{\n  \"schema_version\": 1,\n  \"assets\": [\n")
	for i, e := range entries {
		out.WriteString("    ")
		out.Write(e)
		if i+1 < len(entries) {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString("  ]\n}\n")
	if err := os.WriteFile(name, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// campaignShoreTiles renders the shore tile id: a derived shape, or the winter
// form (suffix _w) of a roster shore, a stream mouth or a derived shape.
func campaignShoreTiles(summer, winter shoreInputs, id string) (map[int]*image.NRGBA, error) {
	in := summer
	base, cold := strings.CutSuffix(id, "_w")
	if cold {
		in = winter
	}
	shore, port := base, 0
	if i := strings.Index(base, "-outlet"); i >= 0 {
		shore = base[:i]
		if _, err := fmt.Sscanf(base[i:], "-outlet%d", &port); err != nil {
			return nil, fmt.Errorf("shore %s: %w", id, err)
		}
	}
	for _, v := range shoreRoster() {
		if v.ID != shore {
			continue
		}
		canonical, err := exportShoreTiles(in, v)
		if err != nil || port == 0 {
			return canonical, err
		}
		return exportShoreMouthOnBase(in, v, port, canonical)
	}
	shape, ok := mapcompose.ParseShore(shore)
	if !ok {
		return nil, fmt.Errorf("unknown shore %s", id)
	}
	if port != 0 {
		return exportShoreMouth(in, shoreVariant{ID: shore}, port)
	}
	return exportShoreShapeTiles(in, shape)
}

// generateShoreMouths adds only explicitly requested new variants. Preflight
// the entire set, retain every existing export, and detect concurrent index edits.
func generateShoreMouths(t *testing.T, root string, ids []string) {
	t.Helper()
	index := filepath.Join(root, "assets/index.json")
	before, err := os.ReadFile(index)
	if err != nil {
		t.Fatal(err)
	}
	entries, installed, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, id := range ids {
		base := strings.TrimSuffix(strings.TrimSuffix(id, "_au"), "_sp")
		base = strings.TrimSuffix(base, "_w")
		shore, tail, ok := strings.Cut(base, "-outlet")
		shape, valid := mapcompose.ParseShore(shore)
		var port int
		if !ok || !valid || seen[id] || (id != base && id != base+"_au" && id != base+"_sp" && id != base+"_w") {
			t.Fatalf("invalid or duplicate shore mouth %q", id)
		}
		if _, err := fmt.Sscanf(tail, "%d", &port); err != nil || port < 1 || port > 6 || base != fmt.Sprintf("%s-outlet%d", shore, port) ||
			shape.Land&(1<<(port%6)) == 0 || shape.Land&(1<<((port+1)%6)) == 0 {
			t.Fatalf("%s is not a mouth on a land edge", id)
		}
		seen[id] = true
		for _, d := range []int{2, 4} {
			if _, ok := installed[campaignExportKey("tile/"+id, d)]; ok {
				t.Fatalf("refusing to replace installed %s at %dx", id, d)
			}
			path := filepath.Join(root, "assets/terrain/lakeside", fmt.Sprintf("%s-%dx.png", id, d))
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				t.Fatalf("refusing existing or inaccessible output %s: %v", path, err)
			}
		}
	}
	previous := exportSeason
	defer func() { exportSeason = previous }()
	for _, id := range ids {
		exportSeason = ""
		base, manifest := id, "art/pipeline/shore-terrain-v2.json"
		switch {
		case strings.HasSuffix(id, "_au"):
			exportSeason, base = "autumn", strings.TrimSuffix(id, "_au")
		case strings.HasSuffix(id, "_sp"):
			exportSeason, base = "spring", strings.TrimSuffix(id, "_sp")
		case strings.HasSuffix(id, "_w"):
			manifest = "art/pipeline/shore-terrain-winter-v1.json"
		}
		in, err := loadShoreInputs(filepath.Join(root, manifest))
		if err != nil {
			t.Fatal(err)
		}
		tiles, err := campaignShoreTiles(in, in, base)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/lakeside/%s-%dx.png", id, d)
			var data bytes.Buffer
			if err := png.Encode(&data, tiles[d]); err != nil {
				t.Fatal(err)
			}
			f, err := os.OpenFile(filepath.Join(root, "assets", path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				t.Fatal(err)
			}
			_, writeErr := f.Write(data.Bytes())
			closeErr := f.Close()
			if writeErr != nil || closeErr != nil {
				t.Fatalf("write %s: %v; close: %v", path, writeErr, closeErr)
			}
			sum := sha256.Sum256(data.Bytes())
			raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-backgrounds-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/lakeside"})
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
		}
	}
	current, err := os.ReadFile(index)
	if err != nil || !bytes.Equal(current, before) {
		t.Fatalf("runtime index changed during export: %v", err)
	}
	writeWinterIndex(t, root, entries)
}
