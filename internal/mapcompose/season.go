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

// SeasonAt is the season each hex of l is drawn in: its entry in the
// scenario's Seasons, else LevelSeason.
func SeasonAt(l *level.Level) func(hexgrid.Coord) Season {
	all := LevelSeason(l)
	if l.Extra == nil || len(l.Extra.Seasons) == 0 {
		return Everywhere(all)
	}
	seasons := l.Extra.Seasons
	return func(c hexgrid.Coord) Season {
		if s, ok := seasons[c]; ok {
			return Season(s)
		}
		return all
	}
}

// Everywhere is season s on every hex.
func Everywhere(s Season) func(hexgrid.Coord) Season {
	return func(hexgrid.Coord) Season { return s }
}

// ComposeSeason composes g as Compose does and draws each hex in its
// season: on an autumn or spring hex every drawn ID but the cast shadows
// takes its seasonal form where has it, a cluster's in the season of its
// anchor (Cells[0]), and Scene.Fallbacks counts the summer IDs still drawn
// there. Summer and winter hexes compose as Compose; winter comes from the
// terrain.
func ComposeSeason(g hexgrid.Grid, at func(hexgrid.Coord) Hex, has func(id string) bool, season func(hexgrid.Coord) Season) Scene {
	s := Compose(g, at, has)
	s.inSeason(season, has)
	return s
}

// neutralFamilies are the IDFamily groups whose summer form serves every
// season: they hold no vegetation, so a summer ID drawn for them is no
// fallback.
var neutralFamilies = map[string]bool{
	"tile/trench": true, "tile/bridge": true, "standing/bridge": true, "tile/rBridge": true,
	"standing/rBridge": true, "tile/barbedWire": true, "tile/quay": true, "tile/rail": true,
	"tile/harbour": true, "standing/harbour": true, "standing/hangar": true,
}

// SeasonNeutral reports whether drawn asset ID id belongs to a family whose
// summer form serves autumn and spring.
func SeasonNeutral(id string) bool { return neutralFamilies[IDFamily(id)] }

// inSeason swaps every drawn ID of an autumn or spring hex but the cast
// shadows for its form in that season where has it, counting the others in
// Fallbacks but those of the season-neutral families; nil Fallbacks when no
// hex is in autumn or spring. Winter IDs stay.
func (s *Scene) inSeason(seasonAt func(hexgrid.Coord) Season, has func(string) bool) {
	var season Season
	swap := func(id *string) {
		if season != Autumn && season != Spring || *id == "" || strings.HasSuffix(*id, "_w") {
			return
		}
		if s.Fallbacks == nil {
			s.Fallbacks = map[string]int{}
		}
		if form := SeasonalID(*id, season); has(form) {
			*id = form
			return
		}
		if !SeasonNeutral(*id) {
			s.Fallbacks[*id]++
		}
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			season = seasonAt(hexgrid.Coord{Line: l, Col: col})
			c := &s.Cells[l][col]
			swap(&c.Background)
			swap(&c.Under)
			if c.Field != nil && (season == Autumn || season == Spring) {
				if form := SeasonalID(c.Field.Exemplar, season); has(form) {
					f := *c.Field
					f.Exemplar = form
					c.Field = &f
				}
			}
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
		season = seasonAt(s.Clusters[i].Cells[0])
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
