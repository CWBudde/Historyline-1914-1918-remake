package game

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"

// This package has a single rule set (the former Standard): movement spends
// a point budget of 2 × speed by terrain cost, rBridge* infrastructure reads
// as a road, a unit leaving a carrier uses its own stats, a full building
// blocks entry, a unit moves or attacks once per turn, weapons spend ammo
// per shot and enforce their air/ground/water flags, and a defender without
// a counter weapon deals no counter damage. The Classic (Python) rule set
// was removed; see PLAN.md.

// pointsPerSpeed is the movement point budget per point of speed.
const pointsPerSpeed = 2

// Route is a path to a hex with what it costs.
type Route struct {
	Path []hexgrid.Coord // from the start hex to the target
	Cost int             // movement points spent
	Left int             // movement points left at the target
}
