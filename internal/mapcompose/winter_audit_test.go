package mapcompose

import (
	"maps"
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
// water, shore, road or other infrastructure whose winter form exists. Autumn
// and spring maps (by their month or map.json) draw no summer form of a
// family that has forms of their season in the catalog (IDFamily), so a season
// is filled family by family. The test logs every family still drawn in summer
// form.
func TestWinterMapsDrawNoSummerInfrastructure(t *testing.T) {
	have := indexIDs(t)
	for _, c := range []struct {
		season Season
		least  int // the maps of the season at least
	}{{Winter, 16}, {Autumn, 4}, {Spring, 6}} {
		t.Run(string(c.season), func(t *testing.T) {
			done := func(f string) bool { return slices.Contains(winterAuditDone, f) }
			if c.season != Winter {
				seasonal := map[string]bool{}
				for id := range have {
					if IDSeason(id) == c.season {
						seasonal[IDFamily(id)] = true
					}
				}
				done = func(f string) bool { return seasonal[f] }
			}
			summer, audited := seasonAudit(t, have, c.season)
			if audited < c.least {
				t.Fatalf("audited %d %s maps, want every %s scenario map (at least %d)", audited, c.season, c.season, c.least)
			}
			for _, f := range slices.Sorted(maps.Keys(summer)) {
				total := 0
				for _, n := range summer[f] {
					total += n
				}
				t.Logf("%-24s %5d summer IDs on %2d of %d %s maps", f, total, len(summer[f]), audited, c.season)
				if done(f) {
					t.Errorf("%s: %d summer IDs on %s maps %v", f, total, c.season, summer[f])
				}
			}
		})
	}
}

// seasonAudit composes the shipped maps of season against the catalog have
// and counts the summer IDs they draw per family (auditFamily in winter,
// IDFamily otherwise) and map. Winter maps are the mostly winter ones, the
// others go by LevelSeason.
func seasonAudit(t *testing.T, have map[string]bool, season Season) (summer map[string]map[string]int, audited int) {
	t.Helper()
	summer = map[string]map[string]int{} // family -> map -> count
	count := func(f, name string, n int) {
		if summer[f] == nil {
			summer[f] = map[string]int{}
		}
		summer[f][name] += n
	}
	for _, nl := range allLevels(t) {
		name, l := nl.Name, nl.Level
		g := l.Grid()
		at := func(c hexgrid.Coord) Hex {
			h := l.Lines[c.Line][c.Col]
			return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}
		if season == Winter && !mostlyWinter(g, at) || season != Winter && LevelSeason(l) != season {
			continue
		}
		audited++
		s := ComposeSeason(g, at, func(id string) bool { return have[id] }, season)
		if season != Winter {
			for id, n := range s.Fallbacks {
				count(IDFamily(id), name, n)
			}
			continue
		}
		for _, line := range s.Cells {
			for _, c := range line {
				ids := slices.Concat([]string{c.Background, c.Under}, c.Overlays, c.Standing)
				for _, id := range ids {
					if id == "" || strings.HasSuffix(id, "_w") {
						continue
					}
					if f := auditFamily(id); f != "" {
						count(f, name, 1)
					}
				}
			}
		}
	}
	return summer, audited
}
