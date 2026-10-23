package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terrainlight"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func openTerrainPack(t *testing.T, out string, d int, group string) *assets.Pack {
	t.Helper()
	r, err := os.OpenRoot(filepath.Join(out, terrainDir(d, group)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	p, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// sourcePNG decodes a runtime PNG of assets/.
func sourcePNG(t *testing.T, name string) image.Image {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(assetsRoot, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return im
}

// pixels is im's NRGBA pixels from its top-left corner.
func pixels(im image.Image) []byte {
	b := im.Bounds()
	n := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(n, n.Rect, im, b.Min, draw.Src)
	return n.Pix
}

func TestRetainedTerrainCatalogBuilds(t *testing.T) {
	out := filepath.Join(t.TempDir(), "terrain")
	packs, err := buildTerrain(assetsRoot, out)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"backgrounds": 104, "infrastructure": 75,
		"context-ground": 273, "context-ground-2": 273, "context-ground-3": 273, "context-ground-4": 198,
		"context-forest": 130,
		"context-hills":  273, "context-hills-2": 273, "context-hills-3": 273, "context-hills-4": 151}
	for _, e := range retainedIndex(t) {
		if e.Density == 2 {
			if strings.HasPrefix(e.Pack, "terrain-standing-stamps-") {
				want[standingStampGroup]++
			} else if strings.HasPrefix(e.Pack, "terrain-standing-") {
				want[standingGroup]++
			}
			if strings.HasPrefix(e.Pack, "terrain-stamps-") {
				want[stampGroup]++
			}
			if strings.HasPrefix(e.Pack, "terrain-materials-") {
				want["materials"]++
			}
		}
	}
	if len(packs) != len(want)*2 {
		t.Fatalf("built %d packs, want %d", len(packs), len(want)*2)
	}
	for _, p := range packs {
		if n := len(p.Manifest.Entries); n != want[p.Group] || p.Manifest.Status != "candidate" {
			t.Errorf("%s: %d entries (want %d), status %s", p.Manifest.ID, n, want[p.Group], p.Manifest.Status)
		}
	}
	// Every terrain and infrastructure ID of the shipped levels has a tile at
	// both densities, registered on the hex centre.
	for _, d := range terrainDensities {
		bg := openTerrainPack(t, out, d, "backgrounds")
		inf := openTerrainPack(t, out, d, "infrastructure")
		for _, feature := range []struct {
			group, id string
			w, h      int
			anchor    [2]float64
		}{
			{stampGroup, "stamp/crater/4_-1.1_1.1_0.2/1", 180, 120, [2]float64{90, 60}},
			{stampGroup, "stamp/crater/4_-1.1_1.1_0.2/2", 192, 160, [2]float64{122, 51}},
			{stampGroup, "stamp/crater_w/1/1", 120, 80, [2]float64{60, 40}},
			{stampGroup, "stamp/crater_w/2_1.1/1", 192, 128, [2]float64{60, 44}},
			{stampGroup, "stamp/crater_w/2_-1.1/1", 192, 128, [2]float64{132, 44}},
			{stampGroup, "stamp/crater_w/6_-1.1_-4.2_-2.2_-3.3_-1.3/1", 384, 256, [2]float64{330, 64}},
			{stampGroup, "stamp/boulders_w/1/1", 96, 80, [2]float64{48, 40}},
			{stampGroup, "stamp/boulders_w/3_-1.1_-2.2/1", 240, 160, [2]float64{191, 40}},
			{stampGroup, "stamp/boulders_w/2_0.2/1", 128, 192, [2]float64{64, 56}},
			{stampGroup, "stamp/knoll_w/1/1", 144, 120, [2]float64{72, 60}},
			{standingStampGroup, "standing/stamp/knoll_w/1/1", 144, 120, [2]float64{72, 60}},
			{stampGroup, "stamp/knoll_w/3_-1.1_-2.2/1", 240, 160, [2]float64{191, 40}},
			{standingStampGroup, "standing/stamp/knoll_w/3_-1.1_-2.2/1", 240, 160, [2]float64{191, 40}},
			{stampGroup, "stamp/knoll_w/3_2.0_1.1/1", 240, 160, [2]float64{48, 40}},
			{standingStampGroup, "standing/stamp/knoll_w/3_2.0_1.1/1", 240, 160, [2]float64{48, 40}},
			{stampGroup, "stamp/knoll_w/4_1.1_0.2_0.4/1", 192, 288, [2]float64{96, 48}},
			{standingStampGroup, "standing/stamp/knoll_w/4_1.1_0.2_0.4/1", 192, 288, [2]float64{96, 48}},
			{stampGroup, "stamp/knoll/1/1", 144, 120, [2]float64{72, 60}},
			{standingStampGroup, "standing/stamp/knoll/1/1", 144, 120, [2]float64{72, 60}},
		} {
			e, err := openTerrainPack(t, out, d, feature.group).Entry(feature.id, d)
			if err != nil {
				t.Fatal(err)
			}
			if e.Width != feature.w*d || e.Height != feature.h*d || e.Anchor != feature.anchor {
				t.Errorf("retained feature geometry: %+v", e)
			}
		}
		names, err := fs.Glob(levels.FS, "*.json")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			b, _ := fs.ReadFile(levels.FS, name)
			l, err := level.Parse(b)
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range l.Lines {
				for _, h := range line {
					e, err := bg.Entry(mapcompose.TileID(h.Terrain), d)
					if err != nil {
						t.Fatal(err)
					}
					if e.Anchor != [2]float64{48, 40} || e.Width != 96*d || e.Height != 80*d {
						t.Fatalf("%s: %v %dx%d", e.ID, e.Anchor, e.Width, e.Height)
					}
					for _, id := range h.Infra.IDs {
						if _, err := inf.Entry(mapcompose.TileID(id), d); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		}
	}
	// Terrain is graded as the map renderer grades it; painted cast shadows are
	// drawn from their standing item with the one light, hlkit's are copied.
	meadow := sourcePNG(t, "terrain/grass/grass-4x.png")
	got, err := openTerrainPack(t, out, 4, "backgrounds").Image("tile/hqN", 4)
	if err != nil {
		t.Fatal(err)
	}
	graded := terraingrade.Grade(sourcePNG(t, "terrain/hq/hqN-4x.png"), "tile/hqN", "terrain/hq", meadow)
	if !bytes.Equal(pixels(got), graded.Pix) {
		t.Fatal("tile/hqN is not graded as terraingrade grades it")
	}
	if g := pixels(got); bytes.Equal(g, pixels(sourcePNG(t, "terrain/hq/hqN-4x.png"))) {
		t.Fatal("tile/hqN is not graded")
	}
	standing := openTerrainPack(t, out, 4, standingGroup)
	crowns := sourcePNG(t, "terrain/forest/crowns-default-4x.png")
	cast, err := standing.Image("standing/shadow/forest", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pixels(cast), terrainlight.Cast(crowns, 4, terrainlight.Diameter(crowns, 4), terrainlight.Crown).Pix) {
		t.Fatal("standing/shadow/forest is not cast as terrainlight casts it")
	}
	src, err := os.ReadFile(filepath.Join(assetsRoot, "terrain", "forest", "scatter-shadow-crown-oak-16-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	var shadow bytes.Buffer
	if err := standing.WritePNG(&shadow, "standing/shadow/scatter/crown/oak-16", 4); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(shadow.Bytes(), src) {
		t.Fatal("the build changed a scattered crown's shadow PNG bytes")
	}
	if b, err := os.ReadFile(filepath.Join(out, "grade")); err != nil || string(b) != terraingrade.Version+"\n"+terrainlight.Version+"\n" {
		t.Fatalf("grade stamp %q, %v", b, err)
	}
	if _, err := buildTerrain(assetsRoot, out); err == nil {
		t.Fatal("existing output directory reused")
	}
}

func TestTerrainPlanIncludesGrassVariations(t *testing.T) {
	packs, err := planTerrain(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[int]int{}
	for _, p := range packs {
		for _, e := range p.Manifest.Entries {
			if strings.HasPrefix(e.ID, "tile/grass-detail-") {
				if p.Group != "backgrounds" {
					t.Errorf("grass detail outside the background pack: %+v", e)
				}
				counts[p.Density]++
			}
		}
	}
	if counts[2] != 16 || counts[4] != 16 {
		t.Fatalf("terrain catalog has grass details %v, want sixteen per density", counts)
	}
}

func TestTerrainBuildRefusesBadInputs(t *testing.T) {
	entry := func(es []indexEntry, pack, id string) *indexEntry {
		for i := range es {
			if es[i].Pack == pack && es[i].ID == id {
				return &es[i]
			}
		}
		t.Fatalf("no %s in %s", id, pack)
		return nil
	}
	cases := []struct {
		name   string
		damage func([]indexEntry) []indexEntry
		want   string
	}{
		{"wrong registration", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-backgrounds-2x", "tile/hqN").Anchor = &[2]float64{0, 0}
			return es
		}, "tile/hqN"},
		{"wrong size", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-backgrounds-4x", "tile/hqN").Height = 160
			return es
		}, "tile/hqN"},
		{"wrong group", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-infrastructure-2x", "tile/road124").Pack = "terrain-context-hills-2x"
			return es
		}, "tile/road124"},
		{"missing pack", func(es []indexEntry) []indexEntry {
			var kept []indexEntry
			for _, e := range es {
				if e.Pack != "terrain-context-forest-4x" {
					kept = append(kept, e)
				}
			}
			return kept
		}, "terrain-context-forest-4x"},
		{"unknown pack", func(es []indexEntry) []indexEntry {
			entry(es, "terrain-backgrounds-2x", "tile/hqN").Pack = "terrain-roofs-2x"
			return es
		}, "terrain-roofs-2x"},
	}
	for _, c := range cases {
		if _, err := planTerrain(c.damage(retainedIndex(t))); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want it to mention %q", c.name, err, c.want)
		}
	}
	// A tampered pin fails validation before the output directory exists.
	packs, err := planTerrain(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	m := packs[0].Manifest
	m.Entries[0].SHA256 = strings.Repeat("0", 64)
	out := filepath.Join(t.TempDir(), "terrain")
	if _, err := writePacks(assetsRoot, out, []assets.Manifest{m}, func(int) string { return "." }, nil); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatal("tampered pin accepted", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed build left its output directory")
	}
}

func TestTerrainGapsReport(t *testing.T) {
	out := filepath.Join(t.TempDir(), "terrain")
	var b bytes.Buffer
	if err := run([]string{"build-terrain", "-assets", assetsRoot, "-out", out}, &b); err != nil {
		t.Fatal(err)
	}
	planned, err := planTerrain(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	groundParts := 0
	for _, p := range planned {
		if p.Density == 2 && strings.HasPrefix(p.Group, "context-ground") {
			groundParts++
		}
	}
	partName := fmt.Sprintf("context-ground-%d", groundParts+1)
	variants := 0
	for _, e := range retainedIndex(t) {
		if strings.HasPrefix(e.Pack, "terrain-") {
			variants++
		}
	}
	if !strings.Contains(b.String(), fmt.Sprintf("built %d terrain packs (%d PNG variants", len(planned), variants)) {
		t.Errorf("build output %q", b.String())
	}
	// Move an existing context into an additional paired part: the report must still
	// select it, rather than silently reporting its hexes as missing.
	for _, d := range terrainDensities {
		base := filepath.Join(out, terrainDir(d, "context-ground"))
		part := filepath.Join(out, terrainDir(d, partName))
		raw, err := os.ReadFile(filepath.Join(base, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		m, err := assets.ParseManifest(raw)
		if err != nil {
			t.Fatal(err)
		}
		extra := m
		extra.ID = fmt.Sprintf("terrain-%s-%dx", partName, d)
		extra.Entries = append([]assets.Entry(nil), m.Entries[:1]...)
		m.Entries = m.Entries[1:]
		imagePath := filepath.Join(part, extra.Entries[0].Path)
		if err := os.MkdirAll(filepath.Dir(imagePath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(base, extra.Entries[0].Path), imagePath); err != nil {
			t.Fatal(err)
		}
		for dir, manifest := range map[string]assets.Manifest{base: m, part: extra} {
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	b.Reset()
	if err := run([]string{"terrain-gaps", "-terrain", out}, &b); err != nil {
		t.Fatal(err)
	}
	s := b.String()
	if strings.Contains(s, "  unstamped lake ") {
		t.Error("reusable lake tiles reported as a missing whole-lake painting")
	}
	if strings.Contains(s, "  unstamped ") {
		t.Errorf("shipped terrain cluster has no installed stamp:\n%s", s)
	}
	for _, want := range []string{
		"level1_cp: 156 context tiles selected, 0 missing, 0 undefined, 0 compound issues",
		"level2_cp: 59 context tiles selected, 0 missing, 0 undefined",
		"level3_cp: 78 context tiles selected, 0 missing, 0 undefined",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("report lacks %q", want)
		}
	}
	// A catalog with an omitted optional stamp must still report the fallback.
	// Remove the same known crater shape at both densities, leaving other art intact.
	for _, d := range terrainDensities {
		p := filepath.Join(out, terrainDir(d, stampGroup), "manifest.json")
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var m assets.Manifest
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		kept, removed := m.Entries[:0], false
		for _, e := range m.Entries {
			if e.ID == "stamp/crater_w/1/1" {
				removed = true
			} else {
				kept = append(kept, e)
			}
		}
		if !removed {
			t.Fatal("missing retained single-crater stamp")
		}
		m.Entries = kept
		raw, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, append(raw, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	b.Reset()
	if err := run([]string{"terrain-gaps", "-terrain", out}, &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "  unstamped crater_w 1 ") {
		t.Error("omitted winter crater stamp disappeared from the report")
	}
}

// withStanding adds a standing item at both densities: a 144×120 canvas that
// overhangs its hex by a quarter hex on every side.
func withStanding(es []indexEntry) []indexEntry {
	for _, d := range terrainDensities {
		es = append(es, indexEntry{ID: "standing/forest", Pack: fmt.Sprintf("terrain-standing-%dx", d), Density: d,
			Path: fmt.Sprintf("terrain/forest/standing-%dx.png", d), SHA256: strings.Repeat("0", 64),
			Width: 144 * d, Height: 120 * d, Anchor: &[2]float64{72, 60}, Layer: "standing", Art: "terrain/forest"})
	}
	return es
}

// Synthetic optional-pack fixtures start without retained standing or stamp
// entries, so their missing-density and uniqueness cases stay isolated.
func withoutOptionalTerrain(es []indexEntry) []indexEntry {
	var kept []indexEntry
	for _, e := range es {
		if !strings.HasPrefix(e.Pack, "terrain-standing-") && !strings.HasPrefix(e.Pack, "terrain-stamps-") {
			kept = append(kept, e)
		}
	}
	return kept
}

func TestTerrainPlanStandingLayer(t *testing.T) {
	packs, err := planTerrain(withStanding(withoutOptionalTerrain(retainedIndex(t))))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, p := range packs {
		if p.Group == standingGroup {
			n++
			if len(p.Manifest.Entries) != 1 || p.Manifest.Entries[0].ID != "standing/forest" {
				t.Errorf("%s: %+v", p.Manifest.ID, p.Manifest.Entries)
			}
		}
	}
	if n != len(terrainDensities) {
		t.Errorf("%d standing packs, want one per density", n)
	}
	last := func(es []indexEntry) *indexEntry { return &es[len(es)-1] }
	cases := []struct {
		name   string
		damage func([]indexEntry) []indexEntry
		want   string
	}{
		{"overhang beyond half a hex", func(es []indexEntry) []indexEntry {
			last(es).Anchor = &[2]float64{120, 60}
			return es
		}, "more than half a hex"},
		{"one density only", func(es []indexEntry) []indexEntry { return es[:len(es)-1] }, "at 1 of 2 densities"},
		{"ground layer in the standing pack", func(es []indexEntry) []indexEntry {
			last(es).Layer = ""
			return es
		}, "layer ground, want standing"},
		{"standing layer in a ground pack", func(es []indexEntry) []indexEntry {
			for i := range es {
				if es[i].Pack == "terrain-backgrounds-2x" {
					es[i].Layer = "standing"
					break
				}
			}
			return es
		}, "want ground"},
	}
	for _, c := range cases {
		if _, err := planTerrain(c.damage(withStanding(withoutOptionalTerrain(retainedIndex(t))))); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want it to mention %q", c.name, err, c.want)
		}
	}
}

func TestTerrainPlanStamps(t *testing.T) {
	es := withoutOptionalTerrain(retainedIndex(t))
	for _, d := range terrainDensities {
		for _, e := range []indexEntry{
			{ID: "stamp/crater/4_-1.1_1.1_0.2/1", Pack: fmt.Sprintf("terrain-stamps-%dx", d), Width: 300 * d, Height: 160 * d, Anchor: &[2]float64{140, 40}},
			// A stamp's standing part spans the cluster, beyond a single hex's overhang.
			{ID: "standing/stamp/crater/4_-1.1_1.1_0.2/1", Pack: fmt.Sprintf("terrain-standing-%dx", d), Layer: "standing", Width: 300 * d, Height: 160 * d, Anchor: &[2]float64{140, 40}},
		} {
			e.Density, e.SHA256, e.Art = d, strings.Repeat("0", 64), "terrain/mud"
			e.Path = fmt.Sprintf("terrain/mud/%s-%d-%dx.png", strings.ReplaceAll(e.ID, "/", "-"), len(es), d)
			es = append(es, e)
		}
	}
	packs, err := planTerrain(es)
	if err != nil {
		t.Fatal(err)
	}
	groups := map[string]int{}
	for _, p := range packs {
		groups[p.Group] += len(p.Manifest.Entries)
	}
	if groups[stampGroup] != 2 || groups[standingGroup] != 2 {
		t.Errorf("packs %v, want a stamp and its standing part per density", groups)
	}
	es[len(es)-2].ID = "tile/crater"
	if _, err := planTerrain(es); err == nil || !strings.Contains(err.Error(), "is not a stamp/") {
		t.Errorf("tile in the stamp pack: %v", err)
	}
}

func TestTerrainPlanSeparateStandingStamps(t *testing.T) {
	es := withoutOptionalTerrain(retainedIndex(t))
	for _, d := range terrainDensities {
		es = append(es, indexEntry{ID: "standing/stamp/knoll_w/1/1", Pack: fmt.Sprintf("terrain-standing-stamps-%dx", d),
			Density: d, Path: fmt.Sprintf("terrain/mountain/knoll-winter-%dx.png", d), SHA256: strings.Repeat("0", 64),
			Width: 300 * d, Height: 160 * d, Anchor: &[2]float64{140, 40}, Layer: "standing", Art: "terrain/mountain"})
	}
	packs, err := planTerrain(es)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, p := range packs {
		if p.Group == "standing-stamps" {
			n++
			if len(p.Manifest.Entries) != 1 || p.Manifest.Entries[0].ID != "standing/stamp/knoll_w/1/1" {
				t.Fatalf("cluster pack entries: %+v", p.Manifest)
			}
		}
	}
	if n != 2 {
		t.Fatalf("got %d cluster-standing packs, want both densities", n)
	}
	if _, err := planTerrain(es[:len(es)-1]); err == nil || !strings.Contains(err.Error(), "at 1 of 2 densities") {
		t.Errorf("missing density: %v", err)
	}
	es[len(es)-1].ID = "standing/forest"
	if _, err := planTerrain(es); err == nil || !strings.Contains(err.Error(), "standing/stamp/") {
		t.Errorf("non-cluster standing item: %v", err)
	}
}
