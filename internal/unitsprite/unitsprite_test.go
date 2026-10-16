package unitsprite

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		if _, ok := Family(id); !ok {
			t.Errorf("unit type %s has no sprite family", id)
		}
	}
	for id := range families {
		if _, ok := set.Units[id]; !ok {
			t.Errorf("family table names unknown unit type %s", id)
		}
	}
}

// The art recipes are the other half of the contract: each family slug is the
// recipe ID without "unit-" and "-v1", and lists exactly the class IDs mapped here.
func TestFamiliesMatchArtRecipes(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "art", "recipes", "unit-*-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var seen []string
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var r struct {
			ID     string `json:"id"`
			Report struct {
				ClassIDs []string `json:"class_ids"`
			} `json:"bundle_report"`
		}
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatal(p, err)
		}
		slug, ok := SlugOfRecipe(r.ID)
		if !ok {
			t.Fatalf("%s: recipe ID %q is not unit-<family>-v1", p, r.ID)
		}
		seen = append(seen, slug)
		got := ClassIDs(slug)
		want := slices.Clone(r.Report.ClassIDs)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("%s: class IDs %v, recipe has %v", slug, got, want)
		}
	}
	slices.Sort(seen)
	if !slices.Equal(seen, Families()) {
		t.Errorf("recipe families %v, table %v", seen, Families())
	}
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
