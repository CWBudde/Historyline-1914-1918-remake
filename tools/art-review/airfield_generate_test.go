//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateAirfieldTiles -count=1 -v
// Renders tile/runway36, the aprons (tile/hangar, apron-sw, apron-nw), the
// hangar (standing/hangar) and its shadow, summer and winter (_w), from the
// selected airfield paintings into assets/terrain/airfield/, replacing earlier
// exports, and lists them in assets/index.json; run `go run ./cmd/hl art index`
// afterwards. With -args -airfield-season autumn (or spring) it writes only
// that season's forms (_au, _sp) from the seasonal edits beside the paintings,
// keeping summer and winter; a painting without an edit (the hangar, which
// has no vegetation) keeps its summer form.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

var airfieldSeason = flag.String("airfield-season", "", "autumn or spring: write only that season's airfield forms")

type airfieldEntry struct {
	campaignEntry
	Layer string `json:"layer,omitempty"`
}

func readNRGBA(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(src.Bounds())
	draw.Draw(im, im.Rect, src, src.Bounds().Min, draw.Src)
	return im
}

func TestGenerateAirfieldTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	seasons := []string{"summer", "winter"}
	if *airfieldSeason != "" {
		if _, ok := airfieldSuffix[*airfieldSeason]; !ok || *airfieldSeason == "summer" || *airfieldSeason == "winter" {
			t.Fatalf("-airfield-season %q: want autumn or spring", *airfieldSeason)
		}
		seasons = []string{*airfieldSeason}
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && e.Art == "terrain/airfield" && slices.Contains(seasons, airfieldSeasonOf(e.ID))
	})
	art := filepath.Join(root, "art/terrain/airfield")
	// source is the painting of piece for season, checked against the
	// summer painting's size; "" when a seasonal edit is missing.
	source := func(piece, season string) *image.NRGBA {
		path, err := airfieldSource(art, piece, season)
		if err != nil {
			t.Fatal(err)
		}
		if path == "" {
			t.Logf("%s: no %s edit, the summer form serves", piece, season)
			return nil
		}
		im := readNRGBA(t, path)
		if season != "summer" && season != "winter" {
			if s := readNRGBA(t, filepath.Join(art, "airfield-"+piece+"-summer-v1.png")); s.Bounds() != im.Bounds() {
				t.Fatalf("%s is %v, the summer painting %v", path, im.Bounds(), s.Bounds())
			}
		}
		return im
	}
	dir := filepath.Join(root, "assets/terrain/airfield")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(id, file, pack, layer string, tiles map[int]*image.NRGBA) {
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/airfield/%s-%dx.png", file, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			raw, err := json.Marshal(airfieldEntry{campaignEntry{ID: id, Pack: fmt.Sprintf(pack, d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/airfield"}, layer})
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
		}
	}
	for _, season := range seasons {
		suffix := airfieldSuffix[season]
		// Both strips paint the runway centred on row 502, its boundary stones
		// 375 px apart: two spacings make the 80 world px period.
		if im := source("runway", season); im != nil {
			strip := newRunwayStrip(im, 502, 750/80.)
			runway, err := runwayTiles(strip)
			if err != nil {
				t.Fatal(err)
			}
			write("tile/runway36"+suffix, "runway36"+suffix, "terrain-infrastructure-%dx", "", runway)
			for _, id := range slices.Sorted(maps.Keys(apronPieces)) {
				write("tile/"+id+suffix, id+suffix, "terrain-infrastructure-%dx", "", apronTiles(strip, apronPieces[id]))
			}
		}
		if im := source("hangar", season); im != nil {
			sprite, shadow, err := hangarSprite(im)
			if err != nil {
				t.Fatal(err)
			}
			write("standing/hangar"+suffix, "hangar"+suffix+"-standing", "terrain-standing-%dx", "standing", sprite)
			// Cast shadows have no season.
			if season == "summer" || season == "winter" {
				write("standing/shadow/hangar"+suffix, "hangar"+suffix+"-shadow", "terrain-standing-%dx", "standing", shadow)
			}
		}
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
