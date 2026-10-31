package mapcompose

import (
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Season is the look a map is drawn in. Winter comes from the terrain (its
// IDs end in _w) and the overlays follow it (mostlyWinter). Autumn and spring
// change the look only: every drawn ID but the cast shadows takes its
// seasonal form, the summer ID plus _au or _sp, where the catalog has it, and
// keeps its summer form otherwise.
type Season string

// The seasons.
const (
	Summer Season = "summer"
	Winter Season = "winter"
	Autumn Season = "autumn"
	Spring Season = "spring"
)

// suffix is the ID suffix of the season's forms, "" for summer.
func (s Season) suffix() string {
	switch s {
	case Winter:
		return "_w"
	case Autumn:
		return "_au"
	case Spring:
		return "_sp"
	}
	return ""
}

// SeasonalID is the form of summer asset ID id in season s: id with the
// season's suffix ("tile/road36_au"), id itself in summer.
func SeasonalID(id string, s Season) string { return id + s.suffix() }

// SummerID is asset ID id without an autumn or spring suffix; winter forms
// keep their _w.
func SummerID(id string) string {
	for _, s := range []Season{Autumn, Spring} {
		if base, ok := strings.CutSuffix(id, s.suffix()); ok {
			return base
		}
	}
	return id
}

// IDSeason is the season of asset ID id by its suffix: winter for _w, autumn
// for _au, spring for _sp, summer otherwise.
func IDSeason(id string) Season {
	for _, s := range []Season{Winter, Autumn, Spring} {
		if strings.HasSuffix(id, s.suffix()) {
			return s
		}
	}
	return Summer
}

// LevelSeason is the season l is drawn in: the one its scenario gives
// (level.Extra), else winter for a mostly winter map and summer otherwise.
func LevelSeason(l *level.Level) Season {
	if l.Extra != nil {
		switch s := Season(l.Extra.Season); s {
		case Summer, Winter, Autumn, Spring:
			return s
		}
	}
	if mostlyWinter(l.Grid(), LevelHex(l)) {
		return Winter
	}
	return Summer
}

// ComposeSeason composes g as Compose does and draws it in season: in autumn
// and spring every drawn ID but the cast shadows takes its seasonal form
// where has it, and Scene.Fallbacks counts the summer IDs still drawn. Summer
// and winter compose as Compose; winter comes from the terrain.
func ComposeSeason(g hexgrid.Grid, at func(hexgrid.Coord) Hex, has func(id string) bool, season Season) Scene {
	s := Compose(g, at, has)
	if season == Autumn || season == Spring {
		s.inSeason(season, has)
	}
	return s
}

// inSeason swaps every drawn ID of the scene but the cast shadows for its form
// in season where has it, counting the others in Fallbacks. Winter IDs stay.
func (s *Scene) inSeason(season Season, has func(string) bool) {
	s.Fallbacks = map[string]int{}
	swap := func(id *string) {
		if *id == "" || strings.HasSuffix(*id, "_w") {
			return
		}
		if form := SeasonalID(*id, season); has(form) {
			*id = form
			return
		}
		s.Fallbacks[*id]++
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			c := &s.Cells[l][col]
			swap(&c.Background)
			swap(&c.Under)
			for _, ids := range [][]string{c.Overlays, c.Standing} {
				for i := range ids {
					swap(&ids[i])
				}
			}
			for _, ps := range [][]Placed{c.Scatter, c.Ground} {
				for i := range ps {
					swap(&ps[i].ID)
				}
			}
		}
	}
	for i := range s.Clusters {
		swap(&s.Clusters[i].Stamp)
		swap(&s.Clusters[i].Standing)
	}
}

// IDFamily groups drawn asset ID id for reports, its season ignored: a tile or
// standing item by its name up to the first digit, dash or underscore and
// without a compound piece's direction ("tile/road36_au" is "tile/road",
// "tile/hqNE" "tile/hq", "standing/rBridge36" "standing/rBridge"), a context,
// stamp or nested standing item by its first path parts ("context/forest",
// "stamp/knoll", "standing/scatter/crown").
func IDFamily(id string) string {
	parts := strings.Split(SummerID(id), "/")
	switch {
	case len(parts) == 1:
		return parts[0]
	case len(parts) == 2:
		name := parts[1]
		if end := strings.IndexFunc(name, func(r rune) bool { return r >= '0' && r <= '9' || r == '-' || r == '_' }); end >= 0 {
			name = name[:end]
		}
		if base := strings.TrimRight(name, "NESWC"); base != "" {
			name = base
		}
		return parts[0] + "/" + name
	case parts[0] == "standing" && len(parts) > 3:
		return strings.Join(parts[:3], "/")
	}
	return parts[0] + "/" + parts[1]
}
