package unitsprite

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestEveryUnitTypeHasAFamily(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for id := range set.Units {
		_, ok := Family(id)
		if !ok {
			t.Errorf("unit type %s must draw authored art, not a pending dot", id)
		}
	}
	for id := range families {
		if _, ok := set.Units[id]; !ok {
			t.Errorf("family table names unknown unit type %s", id)
		}
	}
}

func TestCapturedAndSharedUnitsKeepTheirSubjects(t *testing.T) {
	for _, tc := range []struct{ id, family string }{
		{"CP_ft17", "renault-ft17"}, {"EC_ft17", "renault-ft17"},
		{"CP_markI", "mark-i"}, {"CP_markIV", "mark-iv"},
		{"EC_armCar", "armored-car"}, {"EC_transTrain", "transport-train"},
		{"CP_junkersJI", "junkers-j4"}, {"CP_gothaGV", "gotha-g-vb"},
	} {
		if family, ok := Family(tc.id); !ok || family != tc.family {
			t.Errorf("%s draws %q, want %q", tc.id, family, tc.family)
		}
	}
}

// The art folders are the other half of the contract: every unit pipeline
// manifest unit-<family>-v<N>.json names a family of this table, and its prompt.md lists
// the mapped class IDs.
func TestFamiliesMatchArt(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "art", "pipeline", "unit-*-v*.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := regexp.MustCompile(`^unit-(.+)-v[0-9]+\.json$`)
	var seen, long []string
	for _, p := range paths {
		m := manifest.FindStringSubmatch(filepath.Base(p))
		if m == nil {
			continue
		}
		family := m[1]
		if family == "panel" || slices.Contains(seen, family) || slices.Contains(long, family) {
			continue // the unit panel manifest, or another version of a family
		}
		if strings.HasSuffix(family, "-long") {
			long = append(long, family) // drawn from its ship family's art
			continue
		}
		seen = append(seen, family)
		got := ClassIDs(family)
		for _, kind := range []string{"units", "portraits"} {
			want := promptRuntimeIDs(t, kind, family)
			if !slices.Equal(got, want) {
				t.Errorf("%s/%s: class IDs %v, prompt.md has %v", kind, family, got, want)
			}
		}
	}
	slices.Sort(seen)
	if !slices.Equal(seen, Families()) {
		t.Errorf("pipeline families %v, table %v", seen, Families())
	}
	slices.Sort(long)
	if !slices.Equal(long, LongFamilies()) {
		t.Errorf("long pipeline families %v, table %v", long, LongFamilies())
	}
	for _, f := range Families() {
		if l, ok := Long(f); ok && (Size(l) != LongSize || Size(f) != 96) {
			t.Errorf("%s: long family %s", f, l)
		}
	}
}

// promptRuntimeIDs reads the sorted runtime_ids of an art family's prompt.md.
func promptRuntimeIDs(t *testing.T, kind, family string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "art", kind, family, "prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.Lines(string(b)) {
		v, ok := strings.CutPrefix(line, "runtime_ids:")
		if !ok {
			continue
		}
		v, _, _ = strings.Cut(v, "#")
		var ids []string
		for id := range strings.SplitSeq(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
		slices.Sort(ids)
		return ids
	}
	t.Fatalf("%s: prompt.md has no runtime_ids", family)
	return nil
}

func TestIDNamesFamilyOwnerFacingAndState(t *testing.T) {
	cases := []struct {
		family  string
		palette Palette
		face    Facing
		state   State
		want    string
	}{
		{"infantry", CP, N, Normal, "unit/infantry/cp/n-normal"},
		{"light-artillery", EC, NW, Spent, "unit/light-artillery/ec/nw-spent"},
		{"patrol-boat", Neutral, SE, Normal, "unit/patrol-boat/neutral/se-normal"},
	}
	for _, c := range cases {
		if got := ID(c.family, c.palette, c.face, c.state); got != c.want {
			t.Errorf("ID = %q, want %q", got, c.want)
		}
	}
}

func TestPaletteFollowsOwnerNotCountry(t *testing.T) {
	if PaletteOf(game.CentralPowers) != CP || PaletteOf(game.Entente) != EC || PaletteOf(game.None) != Neutral {
		t.Fatal("owner palettes")
	}
	// A captured Entente unit owned by the Central Powers is drawn in CP colours:
	// the palette ignores the "EC_" prefix of its type.
	fam, _ := Family("EC_cavalry")
	if got := ID(fam, PaletteOf(game.CentralPowers), S, Normal); !strings.HasPrefix(got, "unit/cavalry/cp/") {
		t.Fatal(got)
	}
}

func TestFacingLabelsRoundTrip(t *testing.T) {
	for f := range Facings {
		g, ok := ParseFacing(f.String())
		if !ok || g != f {
			t.Errorf("%v does not round-trip", f)
		}
	}
	if _, ok := ParseFacing("e"); ok {
		t.Error("east is not an authored facing")
	}
}

// Every neighbour step faces its own direction; in-between vectors snap to the
// nearest authored heading.
func TestFacingToward(t *testing.T) {
	g := hexgrid.Grid{Width: 10, Lines: 10}
	for _, from := range []hexgrid.Coord{{Line: 4, Col: 4}, {Line: 5, Col: 4}} {
		fx, fy := hexgrid.Center(from)
		got := map[Facing]bool{}
		for _, n := range g.Neighbors(from) {
			nx, ny := hexgrid.Center(n)
			f, ok := Toward(nx-fx, ny-fy)
			if !ok {
				t.Fatal("no facing for a neighbour step")
			}
			got[f] = true
		}
		if len(got) != 6 {
			t.Errorf("from %v: neighbours give facings %v, want all six", from, got)
		}
	}
	cases := []struct {
		dx, dy float64
		want   Facing
	}{
		{0, -1, N}, {0, 1, S}, {-71, -40, NW}, {-71, 40, SW}, {71, 40, SE}, {71, -40, NE},
		{-100, -1, NW}, {-100, 1, SW}, {100, -1, NE}, {100, 1, SE},
		{142, -400, N}, // two lines up and a column across: still mostly north
	}
	for _, c := range cases {
		if f, _ := Toward(c.dx, c.dy); f != c.want {
			t.Errorf("Toward(%v, %v) = %v, want %v", c.dx, c.dy, f, c.want)
		}
	}
	if _, ok := Toward(0, 0); ok {
		t.Error("a zero vector has no facing")
	}
}
