package dosai

import (
	"cmp"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
)

// The computer's economy, as the original's computer side plans it: repairs
// first, then production, after all the units' orders of the action phase.
const (
	repairFloor = 14 // the side repairs when its energy is above this
	reserve     = 61 // energy a factory keeps: it builds with what is above
	maxBuilds   = 10 // units built per phase, over all factories
)

// computerBuilds are the types the computer may build, per side, on a map that
// gives no list of its own: never a construction unit, a transporter or a train.
var computerBuilds = [2][]int{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 15, 17, 18, 19, 21},
	{10, 11, 12, 14, 15, 17, 18, 21, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50},
}

// mayBuild reports whether the computer may build type typ for side, given the
// building's map list (nil: its side's default). A map list holds both sides'
// types: side 0 builds none above 35, side 1 none below 10.
func mayBuild(list []int, side, typ int) bool {
	if len(list) == 0 {
		return slices.Contains(computerBuilds[side], typ)
	}
	if side == 0 && typ > 35 || side == 1 && typ < 10 {
		return false
	}
	return slices.Contains(list, typ)
}

// allowedBuilds are the options of building record hold the computer may
// build for side, whatever the side can pay.
func allowedBuilds(s *dosplay.Session, side, hold int) []dosplay.BuildOption {
	list := s.Game.Buildings[hold].ComputerBuilds
	return slices.DeleteFunc(s.Buildable(side, hold), func(o dosplay.BuildOption) bool {
		return !mayBuild(list, side, o.Type)
	})
}

// pick is the candidate among n a factory's slot builds in a turn. The
// original draws at random; this draw is a hash of the arguments, so that
// equal sessions plan alike and the web edition's port gives the same.
func pick(turn, side, hold, slot, n int) int {
	x := uint32(turn)*73856093 ^ uint32(side+1)*19349663 ^ uint32(hold)*83492791 ^ uint32(slot+1)*2654435761
	x ^= x >> 15
	x *= 2654435761
	x ^= x >> 13
	return int(x % uint32(n))
}

// ownBuildings are the side's buildings, in record order.
func (pl *plan) ownBuildings() []dosplay.BuildingView {
	var out []dosplay.BuildingView
	for _, c := range sorted(pl.view.Buildings) {
		if b := pl.view.Buildings[c]; b.Owner == pl.s.Faction(pl.side) {
			out = append(out, b)
		}
	}
	slices.SortFunc(out, func(a, b dosplay.BuildingView) int { return cmp.Compare(a.Hold, b.Hold) })
	return out
}

// repairs repairs every damaged unit in the side's buildings, lowest unit
// record first, when the side's energy at the start of the economy, energy,
// is above repairFloor. What the side can no longer pay the rules refuse.
func (pl *plan) repairs(energy int) ([]dosplay.Event, error) {
	if energy <= repairFloor {
		return nil, nil
	}
	var units []dosplay.UnitView
	for _, b := range pl.ownBuildings() {
		for _, u := range b.Units {
			if u.Side == pl.side && u.Size < u.Full {
				units = append(units, u)
			}
		}
	}
	slices.SortFunc(units, func(a, b dosplay.UnitView) int { return cmp.Compare(a.Index, b.Index) })
	var evs []dosplay.Event
	for _, v := range units {
		u := pl.s.Unit(v.Index)
		if pl.s.CanRepair(pl.side, u.In, u.Slot) != nil {
			continue
		}
		ev, err := pl.s.Repair(pl.side, u.In, u.Slot)
		if err != nil {
			return evs, err
		}
		evs = append(evs, ev...)
	}
	return evs, nil
}

// produce fills the side's factories in record order. Each factory plans
// from the energy the economy started with, keeping the reserve, so several
// factories can together spend into it; a build the side can no longer pay
// the rules refuse.
func (pl *plan) produce(energy int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	built := 0
	for _, b := range pl.ownBuildings() {
		if b.Kind != dosplay.Factory {
			continue
		}
		e := energy
		for slot := len(b.Units); slot < dosplay.Slots && built < maxBuilds; slot++ {
			budget := e - reserve
			if budget <= 0 {
				break
			}
			opts := slices.DeleteFunc(allowedBuilds(pl.s, pl.side, b.Hold), func(o dosplay.BuildOption) bool {
				return o.Cost > budget
			})
			if len(opts) == 0 {
				break
			}
			o := opts[pick(pl.s.Turn(), pl.side, b.Hold, slot, len(opts))]
			e -= o.Cost
			built++
			if pl.s.CanBuild(pl.side, b.Hold, o.Type) != nil {
				continue
			}
			ev, err := pl.s.Build(pl.side, b.Hold, o.Type)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
		}
	}
	return evs, nil
}
