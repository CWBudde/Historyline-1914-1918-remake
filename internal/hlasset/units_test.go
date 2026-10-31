package hlasset

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

var assetsRoot = filepath.Join("..", "..", "assets")

func TestRetainedUnitCatalogBuildsEveryFamilyAndDensity(t *testing.T) {
	out := filepath.Join(t.TempDir(), "units")
	packs, err := buildUnits(assetsRoot, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != len(unitsprite.AllFamilies())*2 {
		t.Fatalf("built %d packs, want %d", len(packs), len(unitsprite.AllFamilies())*2)
	}
	for _, p := range packs {
		m := p.Manifest
		if m.Status != "candidate" || m.ID != fmt.Sprintf("unit-%s-%dx", p.Family, p.Density) || len(m.Entries) != 36 {
			t.Fatalf("pack %s: status %s, %d entries", m.ID, m.Status, len(m.Entries))
		}
		dir := filepath.Join(out, fmt.Sprintf("%dx", p.Density), p.Family)
		r, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		pack, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
		_ = r.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, pal := range unitsprite.Palettes {
			for f := range unitsprite.Facings {
				for _, s := range unitsprite.States {
					e, err := pack.Entry(unitsprite.ID(p.Family, pal, f, s), p.Density)
					if err != nil {
						t.Fatal(err)
					}
					size, anchor := 96, [2]float64{48, 49}
					if unitsprite.Size(p.Family) == unitsprite.LongSize {
						// A long ship's anchor is the midpoint between its two hexes.
						size, anchor = unitsprite.LongSize, [2]float64{unitsprite.LongSize / 2, unitsprite.LongSize / 2}
					}
					if e.Anchor != anchor || e.Width != size*p.Density || e.Height != size*p.Density {
						t.Fatalf("%s: registration %v %dx%d", e.ID, e.Anchor, e.Width, e.Height)
					}
				}
			}
		}
	}
	// Bytes are copied, not re-encoded.
	src, err := os.ReadFile(filepath.Join(assetsRoot, "units", "light-artillery", "ec", "nw-spent-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := os.OpenRoot(filepath.Join(out, "4x", "light-artillery"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	pack, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := pack.WritePNG(&got, "unit/light-artillery/ec/nw-spent", 4); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Bytes(), src) {
		t.Fatal("runtime build changed the source PNG bytes")
	}
	var output bytes.Buffer
	if err := Run([]string{"validate", "-root", filepath.Join(out, "2x", "infantry"), "-allow-candidates"}, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "validated unit-infantry-2x: 36 PNG variants") {
		t.Fatal(output.String())
	}
	before, err := os.ReadFile(filepath.Join(out, "2x", "infantry", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := buildUnits(assetsRoot, out); err == nil {
		t.Fatal("existing unit catalog overwritten")
	}
	after, err := os.ReadFile(filepath.Join(out, "2x", "infantry", "manifest.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed rebuild changed the existing catalog", err)
	}
}

func TestBuildUnitsCommand(t *testing.T) {
	out := filepath.Join(t.TempDir(), "units")
	var output bytes.Buffer
	if err := Run([]string{"build-units", "-assets", assetsRoot, "-out", out}, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	families := len(unitsprite.AllFamilies())
	if !strings.Contains(output.String(), fmt.Sprintf("built %d unit packs (%d families × 2 densities, %d PNG variants; candidate", families*2, families, families*72)) {
		t.Fatal(output.String())
	}
	if err := Run([]string{"build-units"}, &bytes.Buffer{}, io.Discard); err == nil {
		t.Fatal("missing -out accepted")
	}
}

// retainedIndex reads the real index so a test can damage one detail of it.
func retainedIndex(t *testing.T) []indexEntry {
	t.Helper()
	entries, err := readIndex(assetsRoot)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func TestPlanUnitsRejectsDamagedIndex(t *testing.T) {
	first := func(entries []indexEntry, pack string) *indexEntry {
		for i := range entries {
			if entries[i].Pack == pack {
				return &entries[i]
			}
		}
		t.Fatalf("no %s entry", pack)
		return nil
	}
	cases := []struct {
		name   string
		damage func([]indexEntry) []indexEntry
		want   string
	}{
		{"unknown family", func(es []indexEntry) []indexEntry {
			first(es, "unit-cavalry-2x").Pack = "unit-zeppelin-2x"
			return es
		}, "no sprite family"},
		{"unknown variant", func(es []indexEntry) []indexEntry {
			first(es, "unit-cavalry-2x").ID = "unit/cavalry/red/n-normal"
			return es
		}, "unknown or duplicate"},
		{"wrong size", func(es []indexEntry) []indexEntry {
			first(es, "unit-cavalry-2x").Width = 100
			return es
		}, "96"},
		{"missing variant", func(es []indexEntry) []indexEntry {
			e := first(es, "unit-cavalry-4x")
			e.Pack = ""
			return es
		}, "35 sprite variants"},
		{"missing family", func(es []indexEntry) []indexEntry {
			var kept []indexEntry
			for _, e := range es {
				if !strings.HasPrefix(e.Pack, "unit-sappers-") {
					kept = append(kept, e)
				}
			}
			return kept
		}, "sappers"},
		{"missing density", func(es []indexEntry) []indexEntry {
			var kept []indexEntry
			for _, e := range es {
				if e.Pack != "unit-bunker-4x" {
					kept = append(kept, e)
				}
			}
			return kept
		}, "densities"},
	}
	for _, c := range cases {
		if _, err := planUnits(c.damage(retainedIndex(t))); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want %q", c.name, err, c.want)
		}
	}
}

// A sprite whose bytes do not match its index pin fails before any output exists.
func TestBuildUnitsLeavesNoOutputOnHashMismatch(t *testing.T) {
	packs, err := planUnits(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	var manifests []assets.Manifest
	for _, p := range packs {
		if p.Family == "sappers" {
			p.Manifest.Entries[5].SHA256 = strings.Repeat("0", 64)
		}
		manifests = append(manifests, p.Manifest)
	}
	out := filepath.Join(t.TempDir(), "units")
	if _, err := writePacks(assetsRoot, out, manifests, func(i int) string { return fmt.Sprint(i) }, nil); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatal("tampered pin accepted", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed build left output", err)
	}
}
