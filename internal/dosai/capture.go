package dosai

import (
	"cmp"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The goal values of the computer side's capturers. Every depot and factory
// counts the same and above the enemy HQ, so that the route decides: a
// capturer takes the nearest building first, and the HQ when none is left
// for it. A value is far above any route length.
const (
	buildingValue = 3850
	hqValue       = 450
)

// mission is a capturer's order to take a building.
type mission struct {
	goal  hexgrid.Coord
	value int // the goal's value less the unit's route length
}

// captureGoal is a building to take and what it is worth.
type captureGoal struct {
	at    hexgrid.Coord
	value int
}

// goalValue is what taking building b is worth to the side.
func goalValue(b dosplay.BuildingView) int {
	if b.HQ {
		return hqValue
	}
	return buildingValue
}

// assignMissions gives the side's capturers on the map their buildings, as the
// original does: goals by value, each to the capturer with the shortest route,
// at any distance, if the value less that route beats the unit's mission so
// far. A goal left without a capturer that way then takes the nearest one
// still free, where the original leaves it for the next phase. Units in
// buildings leave first and get a mission the next time.
func (pl *plan) assignMissions() error {
	pl.missions = map[int]mission{}
	goals := pl.captureGoals()
	capturers := pl.capturers()
	for _, g := range goals {
		if err := pl.assign(g, capturers, false); err != nil {
			return err
		}
	}
	for _, g := range goals {
		served := false
		for _, m := range pl.missions {
			served = served || m.goal == g.at
		}
		if !served {
			if err := pl.assign(g, capturers, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// captureGoals are the buildings to take: the enemy HQs, all other buildings
// that are not the side's, by value, then in hex order.
func (pl *plan) captureGoals() []captureGoal {
	var goals []captureGoal
	for _, c := range sorted(pl.view.Buildings) {
		b := pl.view.Buildings[c]
		if b.Owner != pl.s.Faction(pl.side) && (!b.HQ || b.Owner == pl.s.Faction(1-pl.side)) {
			goals = append(goals, captureGoal{c, goalValue(b)})
		}
	}
	slices.SortStableFunc(goals, func(a, b captureGoal) int { return cmp.Compare(b.value, a.value) })
	return goals
}

// capturers are the hexes of the side's units on the map that may capture and
// take an order, by unit.
func (pl *plan) capturers() []hexgrid.Coord {
	var out []hexgrid.Coord
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		if u.Side == pl.side && pl.profile.Tables.Move.Types[u.Type].Flags&captureFlag != 0 &&
			pl.s.CanOrder(pl.side, u.Index) == nil {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b hexgrid.Coord) int {
		return cmp.Compare(pl.view.Units[a].Index, pl.view.Units[b].Index)
	})
	return out
}

// assign gives goal g to the capturer with the shortest route whose mission it
// beats; with free only to one without a mission.
func (pl *plan) assign(g captureGoal, capturers []hexgrid.Coord, free bool) error {
	best, bestHops := -1, 0
	for _, c := range capturers {
		u := pl.view.Units[c]
		m, busy := pl.missions[u.Index]
		if free && busy {
			continue
		}
		hops, ok, err := pl.hops(u, g.at, c)
		if err != nil {
			return err
		}
		if ok && g.value-hops > m.value && (best < 0 || hops < bestHops) {
			best, bestHops = u.Index, hops
		}
	}
	if best >= 0 {
		pl.missions[best] = mission{g.at, g.value - bestHops}
	}
	return nil
}

// routeKey names a unit type's route data to a goal.
type routeKey struct {
	typ  int
	goal hexgrid.Coord
}

// hops is the number of hexes on unit u's shortest route from from to goal, on
// the terrain alone; false if it has none.
func (pl *plan) hops(u dosplay.UnitView, goal, from hexgrid.Coord) (int, bool, error) {
	k := routeKey{u.Type, goal}
	dist, ok := pl.hopFields[k]
	if !ok {
		costs, err := pl.s.Costs(pl.side, u.Index)
		if err != nil {
			return 0, false, err
		}
		dist = map[hexgrid.Coord]int{}
		if _, open := costs[goal]; open {
			dist[goal] = 0
			for q := []hexgrid.Coord{goal}; len(q) > 0; q = q[1:] {
				for _, n := range hexgrid.Around(q[0]) {
					if _, open := costs[n]; !open {
						continue
					}
					if _, seen := dist[n]; !seen {
						dist[n] = dist[q[0]] + 1
						q = append(q, n)
					}
				}
			}
		}
		pl.hopFields[k] = dist
	}
	d, ok := dist[from]
	return d, ok, nil
}

// missionField is the route cost to the goal of unit u's open mission; false if
// it has none or the side has taken the goal since.
func (pl *plan) missionField(u dosplay.UnitView) (map[hexgrid.Coord]int, bool, error) {
	m, ok := pl.missions[u.Index]
	if !ok || pl.view.Buildings[m.goal].Owner == pl.s.Faction(pl.side) {
		return nil, false, nil
	}
	k := routeKey{u.Type, m.goal}
	f, ok := pl.routes[k]
	if !ok {
		costs, err := pl.s.Costs(pl.side, u.Index)
		if err != nil {
			return nil, false, err
		}
		f = dijkstra(costs, map[hexgrid.Coord]int{m.goal: 0})
		pl.routes[k] = f
	}
	return f, true, nil
}
