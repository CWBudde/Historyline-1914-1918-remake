package mapcompose

import (
	"strings"
	"testing"
)

// The compositor must use available grass detail tiles without hiding missing
// terrain contexts or accidentally varying winter grass and other terrains.
func TestComposeGrassDetails(t *testing.T) {
	g, at := grid("grass grass grass grass", "grass grass grass", "grass_w water grass grass")
	has := func(id string) bool { return strings.HasPrefix(id, "tile/grass-detail-") }
	s := Compose(g, at, has)
	seen := map[string]bool{}
	for line := range s.Cells {
		for col, c := range s.Cells[line] {
			if line == 2 && col < 2 {
				want := []string{"tile/grass_w", "tile/water"}[col]
				if c.Background != want {
					t.Errorf("other terrain changed: %s, want %s", c.Background, want)
				}
				continue
			}
			if !has(c.Background) {
				t.Errorf("grass at %d,%d did not select detail: %s", line, col, c.Background)
			}
			seen[c.Background] = true
		}
	}
	if len(seen) < 6 {
		t.Errorf("nearby grass repeats too few interiors: %d", len(seen))
	}
	if len(s.Gaps) != 0 {
		t.Fatalf("optional grass details caused context gaps: %+v", s.Gaps)
	}
	old := Compose(g, at, none)
	if old.Cells[0][0].Background != "tile/grass" || len(old.Gaps) != 0 {
		t.Fatalf("catalog without optional details changed: %+v", old)
	}
}
