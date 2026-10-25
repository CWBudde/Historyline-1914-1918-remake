package mapcompose

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// winterAuditDone lists the families whose summer form must not show on a
// winter map: they have winter forms in the catalog. Not listed: the road
// bridge tiles (tile/rBridge*), which are transparent under their standing
// deck.
var winterAuditDone = []string{
	"road", "cross", "path", "rail", "stream", "trench", "bridge",
	"water", "waterStones", "waterRocks", "lakeside", "quay", "standing/rBridge", "standing/bridge",
}

// auditFamily is the family of a drawn tile or standing ID: its name up to the
// first digit, dash or underscore ("tile/road36_w" is road, "standing/rBridge36"
// is standing/rBridge). Other IDs (ground contexts) give "".
func auditFamily(id string) string {
	kind, name, _ := strings.Cut(id, "/")
	if kind != "tile" && kind != "standing" {
		return ""
	}
	end := strings.IndexFunc(name, func(r rune) bool { return r >= '0' && r <= '9' || r == '-' || r == '_' })
	if end >= 0 {
		name = name[:end]
	}
	if kind == "standing" {
		return kind + "/" + name
	}
	return name
}

// Winter maps (Civil, Mouse and the winter campaign maps) draw no summer
// water, shore, road or other infrastructure whose winter form exists
// (PLAN.md L7.6). The test logs every family still drawn in summer form.
func TestWinterMapsDrawNoSummerInfrastructure(t *testing.T) {
	raw, err := os.ReadFile("../../assets/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Density int    `json:"density"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, entry := range index.Assets {
		if entry.Density == 4 {
			have[entry.ID] = true
		}
	}
	audited := 0
	summer := map[string]map[string]int{} // family -> map -> count
	for _, nl := range allLevels(t) {
		name, l := nl.Name, nl.Level
		g := l.Grid()
		at := func(c hexgrid.Coord) Hex {
			h := l.Lines[c.Line][c.Col]
			return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}
		if !mostlyWinter(g, at) {
			continue
		}
		audited++
		s := Compose(g, at, func(id string) bool { return have[id] })
		for _, line := range s.Cells {
			for _, c := range line {
				ids := slices.Concat([]string{c.Background, c.Under}, c.Overlays, c.Standing)
				for _, id := range ids {
					if id == "" || strings.HasSuffix(id, "_w") {
						continue
					}
					f := auditFamily(id)
					if f == "" {
						continue
					}
					if summer[f] == nil {
						summer[f] = map[string]int{}
					}
					summer[f][name]++
				}
			}
		}
	}
	if audited < 16 {
		t.Fatalf("audited %d winter maps, want every winter scenario map", audited)
	}
	families := slices.Sorted(func(yield func(string) bool) {
		for f := range summer {
			if !yield(f) {
				return
			}
		}
	})
	for _, f := range families {
		total := 0
		for _, n := range summer[f] {
			total += n
		}
		t.Logf("%-20s %5d summer IDs on %2d of %d winter maps", f, total, len(summer[f]), audited)
		if slices.Contains(winterAuditDone, f) {
			t.Errorf("%s: %d summer IDs on winter maps %v", f, total, summer[f])
		}
	}
}
