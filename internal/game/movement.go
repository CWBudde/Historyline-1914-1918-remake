package game

import (
	"container/heap"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// flies reports whether u is an aircraft: it crosses every terrain, but never
// boards a building or a carrier.
func flies(u *Unit) bool { return u.Def.MovementType == data.MoveAir }

// stepCost is what entering h costs a unit of the given movement type, and
// whether it may enter at all. Buildings are handled by the caller. A road,
// rail or water hex costs the minimum of 1, and so does every hex in the air.
func stepCost(h *Hex, moveType int) (cost int, ok bool) {
	if moveType == data.MoveAir {
		return 1, true
	}
	if moveType != data.MoveNormal {
		if !slices.Contains(h.Mods.MovementAllowed, moveType) {
			return 0, false
		}
	} else if h.Mods.Movement == nil {
		return 0, false
	}
	if h.Mods.Movement == nil {
		return 1, true
	}
	return max(*h.Mods.Movement, 1), true
}

// stopOK reports whether the mover m may end its move on h: an own
// carrier with room, or a building it may enter and that holds it, or a free hex.
// An aircraft only stops on a free hex without a building.
func (s *State) stopOK(h *Hex, m *Unit) bool {
	switch {
	case flies(m):
		return h.Unit == nil && h.Container == nil
	case h.Unit != nil:
		o := h.Unit
		return o.Owner == s.Turn && o.Content != nil && o.Content.Fits(m)
	case h.Container != nil:
		if h.Owner != s.Turn && !slices.Contains(m.Def.Skills, skillCapture) {
			return false
		}
		return h.Container.Fits(m)
	}
	return true
}

type node struct {
	c         hexgrid.Coord
	cost, seq int
}

type queue []node

func (q queue) Len() int { return len(q) }
func (q queue) Less(i, j int) bool {
	if q[i].cost != q[j].cost {
		return q[i].cost < q[j].cost
	}
	return q[i].seq < q[j].seq
}
func (q queue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *queue) Push(x any)   { *q = append(*q, x.(node)) }
func (q *queue) Pop() any {
	old := *q
	n := old[len(old)-1]
	*q = old[:len(old)-1]
	return n
}

// movablePoints is Movable: Dijkstra over the terrain costs. m is the unit
// whose speed, movement type and weight count.
// Unseen hexes and hexes with enemy units cannot be entered; own units can be
// passed. Buildings and carriers are only entered as the final stop, and
// never by an aircraft.
func (s *State) movablePoints(from hexgrid.Coord, m *Unit) map[hexgrid.Coord]Route {
	budget := m.Def.Speed * pointsPerSpeed
	best := map[hexgrid.Coord]int{from: 0}
	prev := map[hexgrid.Coord]hexgrid.Coord{}
	q := &queue{{c: from}}
	seq := 0
	for q.Len() > 0 {
		n := heap.Pop(q).(node)
		if n.cost > best[n.c] {
			continue
		}
		if n.c != from {
			if h := s.At(n.c); h.Container != nil || h.Unit != nil && h.Unit.Content != nil {
				continue // a stop, not a way through
			}
		}
		for _, c := range s.Grid.Neighbors(n.c) {
			h := s.At(c)
			if s.Fog(s.Turn, c) == Unseen || h.Unit != nil && h.Unit.Owner != s.Turn {
				continue
			}
			cost := 1
			switch {
			case h.Container != nil && flies(m):
				continue
			case h.Container == nil:
				var ok bool
				if cost, ok = stepCost(h, m.Def.MovementType); !ok {
					continue
				}
			}
			total := n.cost + cost
			if total > budget && n.c != from {
				continue
			}
			if old, seen := best[c]; seen && old <= total {
				continue
			}
			best[c], prev[c] = total, n.c
			seq++
			heap.Push(q, node{c, total, seq})
		}
	}
	out := map[hexgrid.Coord]Route{}
	for c, cost := range best {
		if c == from || !s.stopOK(s.At(c), m) {
			continue
		}
		path := []hexgrid.Coord{c}
		for p := c; p != from; {
			p = prev[p]
			path = append(path, p)
		}
		slices.Reverse(path)
		out[c] = Route{Path: path, Cost: cost, Left: max(budget-cost, 0)}
	}
	return out
}
