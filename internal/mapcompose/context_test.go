package mapcompose

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type recipeProfile struct {
	RuntimeID  string    `json:"runtime_id"`
	Season     string    `json:"season"`
	Neighbors  [6]string `json:"neighbor_ids"`
	Phase      *[2]int   `json:"phase"`
	Descriptor string    `json:"descriptor"`
	Key        string    `json:"context_key"`
}

func readProfiles(t *testing.T, f Family) []recipeProfile {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "art", "recipes", string(f)+"-terrain-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		Profiles []recipeProfile `json:"profiles"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Profiles) == 0 {
		t.Fatalf("%s: no profiles", f)
	}
	return r.Profiles
}

// The port of the offline context contract must reproduce every exported
// profile's descriptor and key from its centre, neighbours and phase.
func TestDescriptorMatchesRecipeProfiles(t *testing.T) {
	n := 0
	for _, f := range Families {
		for _, p := range readProfiles(t, f) {
			var phase [2]int
			if p.Phase != nil {
				phase = *p.Phase
			}
			d, ok := Descriptor(f, p.RuntimeID, p.Neighbors, phase)
			if !ok {
				t.Errorf("%s %s %v: not a valid context", f, p.RuntimeID, p.Neighbors)
				continue
			}
			if d != p.Descriptor || Key(d) != p.Key {
				t.Errorf("%s profile %s: got %q key %s, want %q key %s", f, p.RuntimeID, d, Key(d), p.Descriptor, p.Key)
			}
			n++
		}
	}
	if n != 51+43+122 {
		t.Errorf("checked %d profiles, want 216", n)
	}
}

func TestFamilyOf(t *testing.T) {
	for id, want := range map[string]Family{
		"mud": Ground, "stone": Ground, "rough_w": Ground, "stoney_w": Ground, "shelled_w": Ground,
		"forest": Forest, "forest_w": Forest, "hills": Hills,
		// Bases are drawn from their default tile; every family's profile for a base
		// centre is the canonical default.
		"grass": "", "grass_w": "", "water": "", "mountN": "",
	} {
		if got, _ := FamilyOf(id); got != want {
			t.Errorf("FamilyOf(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestDescriptorRejectsMixedNeighbours(t *testing.T) {
	g := [6]string{"grass", "grass", "grass", "grass", "grass", "grass"}
	mixed := g
	mixed[2] = "hills"
	if _, ok := Descriptor(Forest, "forest", mixed, [2]int{}); ok {
		t.Error("forest next to hills has a descriptor")
	}
	winter := g
	winter[0] = "grass_w"
	if _, ok := Descriptor(Ground, "mud", winter, [2]int{}); ok {
		t.Error("summer mud next to winter grass has a descriptor")
	}
	if _, ok := Descriptor(Hills, "hills", g, [2]int{}); !ok {
		t.Error("hills among grass has no descriptor")
	}
}

func TestHillsPhase(t *testing.T) {
	for _, c := range []struct{ line, col, x, y int }{
		{0, 0, 0, 0}, {0, 1, 142, 0}, {1, 0, 71, 40}, {8, 4, 0, 0}, {9, 5, 213, 40}, {17, 3, 497, 40},
	} {
		x, y := Phase(c.line, c.col)
		if x != c.x || y != c.y {
			t.Errorf("Phase(%d,%d) = %d,%d, want %d,%d", c.line, c.col, x, y, c.x, c.y)
		}
	}
}
