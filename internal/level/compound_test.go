package level

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestCompoundPiece(t *testing.T) {
	for _, c := range []struct {
		terrain, prefix, piece, suffix string
	}{
		{"hqN", "hq", "N", ""},
		{"hqSE_w", "hq", "SE", "_w"},
		{"cmpW", "cmp", "W", ""},
		{"factoryE_w", "factory", "E", "_w"},
		{"mountS", "mount", "S", ""},
		{"mountain", "", "", ""},
		{"hqX", "", "", ""},
		{"grass", "", "", ""},
	} {
		cp, piece, sfx := CompoundPiece(c.terrain)
		prefix := ""
		if cp != nil {
			prefix = cp.Prefix
		}
		if prefix != c.prefix || piece != c.piece || sfx != c.suffix {
			t.Errorf("CompoundPiece(%q) = %q %q %q, want %q %q %q", c.terrain, prefix, piece, sfx, c.prefix, c.piece, c.suffix)
		}
	}
}

// The pieces stand where the level's buildings stamp them: the HQ's seven
// hexes, the camp's and factory's diamonds, the factory entered at its E
// piece, SE of its N piece.
func TestCompoundAt(t *testing.T) {
	want := map[string]struct {
		entrance string
		pieces   map[string]hexgrid.Coord
	}{
		"hq": {"N", map[string]hexgrid.Coord{
			"N": {Line: 4, Col: 3}, "NW": {Line: 5, Col: 2}, "NE": {Line: 5, Col: 3}, "C": {Line: 6, Col: 3},
			"SW": {Line: 7, Col: 2}, "SE": {Line: 7, Col: 3}, "S": {Line: 8, Col: 3},
		}},
		"cmp":     {"N", map[string]hexgrid.Coord{"N": {Line: 4, Col: 3}, "W": {Line: 5, Col: 2}, "E": {Line: 5, Col: 3}, "S": {Line: 6, Col: 3}}},
		"factory": {"E", map[string]hexgrid.Coord{"N": {Line: 4, Col: 3}, "W": {Line: 5, Col: 2}, "E": {Line: 5, Col: 3}, "S": {Line: 6, Col: 3}}},
		"mount":   {"", map[string]hexgrid.Coord{"N": {Line: 4, Col: 3}, "W": {Line: 5, Col: 2}, "E": {Line: 5, Col: 3}, "S": {Line: 6, Col: 3}}},
	}
	if len(Compounds) != len(want) {
		t.Errorf("%d compounds, want %d", len(Compounds), len(want))
	}
	n := hexgrid.Coord{Line: 4, Col: 3}
	for i := range Compounds {
		cp := &Compounds[i]
		w, ok := want[cp.Prefix]
		if !ok {
			t.Errorf("compound %q", cp.Prefix)
			continue
		}
		if cp.Entrance != w.entrance {
			t.Errorf("%s: entrance %q, want %q", cp.Prefix, cp.Entrance, w.entrance)
		}
		if len(cp.Pieces) != len(w.pieces) {
			t.Errorf("%s: %d pieces, want %d", cp.Prefix, len(cp.Pieces), len(w.pieces))
		}
		for p, at := range w.pieces {
			if got := cp.At(n, p); got != at {
				t.Errorf("%s%s: at %v, want %v", cp.Prefix, p, got, at)
			}
		}
	}
}
