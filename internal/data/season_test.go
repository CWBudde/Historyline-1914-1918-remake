package data

import "testing"

// Every winter terrain of the table is the winter form of one kind, and
// InSeason and Kind are each other's reverse.
func TestInSeason(t *testing.T) {
	for kind, want := range map[string]string{
		"grass": "grass_w", "hills": "rough_w", "mud": "shelled_w", "stone": "stoney_w", "hqN": "hqN_w", "water": "water",
	} {
		if got := InSeason(kind, "winter"); got != want || Kind(got) != kind || InSeason(kind, "autumn") != kind {
			t.Errorf("%s: winter %s (kind %s), want %s", kind, got, Kind(got), want)
		}
	}
	for id := range terrainTable() {
		if k := Kind(id); InSeason(k, "winter") != id && k != id {
			t.Errorf("%s: kind %s, whose winter form is %s", id, k, InSeason(k, "winter"))
		}
		if k := Kind(id); !HasTerrain(k) {
			t.Errorf("%s: kind %s has no summer terrain", id, k)
		}
	}
}
