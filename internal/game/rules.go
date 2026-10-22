package game

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"

// Rules selects which Phase 3 fixes apply. The zero value is Classic, the
// behaviour of the Python version, which the oracle tests pin.
type Rules struct {
	// MovementPoints replaces "speed counts hex steps" by a point budget of
	// 2 × speed that each entered hex spends by its terrain cost (a road costs
	// 1, grass 2, forest 3, ...). Paths come from Dijkstra and never enter
	// terrain the unit's movement type cannot cross. A unit with its full
	// budget may always make one step.
	MovementPoints bool
	// RoadBridges reads rBridge* infrastructure as a road (data.FixedModifiers).
	RoadBridges bool
	// OwnStatsOnExit lets a unit leaving a carrier move with its own speed,
	// movement type and weight instead of the carrier's.
	OwnStatsOnExit bool
	// FullDepotBlocks makes a building that cannot hold the unit an illegal
	// target, instead of leaving the unit standing on the building hex.
	FullDepotBlocks bool
	// OneActionPerTurn lets a unit either move or attack in a turn, as in the
	// original: a unit that has moved cannot attack and one that has attacked
	// cannot move or attack again.
	OneActionPerTurn bool
	// AmmoPerShot spends one round of the weapon per shot (the counter fire
	// too) and disables a weapon without ammo for attacks and targeting.
	AmmoPerShot bool
	// TargetFlags enforces the weapons' air, ground and water flags: a weapon
	// hits an aircraft when it has Air, a boat when it has Water, any other
	// unit when it has Ground.
	TargetFlags bool
	// NoCounterNoDamage spares the attacker when the defender has no weapon to
	// fire back with; Python then still runs the formula with the attacker's
	// armour as the defender's power.
	NoCounterNoDamage bool
}

// Classic is the rule set of the Python version.
var Classic = Rules{}

// Standard is the rule set of new games: every Phase 3 fix that exists.
var Standard = Rules{MovementPoints: true, RoadBridges: true, OwnStatsOnExit: true, FullDepotBlocks: true,
	OneActionPerTurn: true, AmmoPerShot: true, TargetFlags: true, NoCounterNoDamage: true}

// Movement point constants of Rules.MovementPoints.
const pointsPerSpeed = 2

// Route is a path to a hex with what it costs.
type Route struct {
	Path []hexgrid.Coord // from the start hex to the target
	Cost int             // movement points spent (steps in Classic)
	Left int             // movement points left at the target
}
