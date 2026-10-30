package llmplay

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Fallback plays side's whole phase deterministically so a game never
// stalls on a bad model reply: in the action phase it first runs the
// service dosai never uses (repair the most damaged held unit, then
// build the cheapest capturer the side can afford), then lets
// dosai.Planner give the phase's orders and confirm. In the movement
// phase it is Planner alone.
func Fallback(s *dosplay.Session, side int, p *dosai.Planner) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	if s.Phase(side) == dosrules.PhaseAction {
		evs = append(evs, Service(s, side)...)
	}
	pevs, err := p.Plan(s, side)
	return append(evs, pevs...), err
}

// Service spends the action-phase economy dosai ignores: repair every
// repairable held unit (flat cost, full strength back), then build in
// each building with a free slot: the cheapest option (capturers are
// the cheapest buildable types). It stops at the first refusal.
func Service(s *dosplay.Session, side int) []dosplay.Event {
	var evs []dosplay.Event
	v := s.View(side)
	for _, at := range sortedCoords(v.Buildings) {
		bv := v.Buildings[at]
		if s.SideOf(bv.Owner) != side {
			continue
		}
		hold, slots, ok := s.Contents(side, at)
		if !ok {
			continue
		}
		for _, u := range slots {
			if u.Size < u.Full && s.CanRepair(side, hold, u.Slot) == nil {
				ev, err := s.Repair(side, hold, u.Slot)
				if err != nil {
					return evs
				}
				evs = append(evs, ev...)
			}
		}
		_, slots, _ = s.Contents(side, at)
		for _, o := range s.Buildable(side, hold) {
			if len(slots) >= dosplay.Slots {
				break
			}
			if s.CanBuild(side, hold, o.Type) != nil {
				continue
			}
			ev, err := s.Build(side, hold, o.Type)
			if err != nil {
				return evs
			}
			evs = append(evs, ev...)
			_, slots, _ = s.Contents(side, at)
		}
	}
	return evs
}
