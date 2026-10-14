package game

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"

// Fog is what a player knows about a hex.
type Fog uint8

const (
	Unseen  Fog = iota // never seen
	Seen               // seen before, not visible now
	Visible            // in sight now
)

func (f Fog) String() string {
	switch f {
	case Unseen:
		return "Unseen"
	case Seen:
		return "Seen"
	case Visible:
		return "Visible"
	}
	return "Fog(?)"
}

// vision is the fog of war of one player, indexed like State.Hexes.
type vision struct {
	seen, visible [][]bool
}

// Fog returns what p knows about the hex at c. Off-map hexes and None are Unseen.
func (s *State) Fog(p Player, c hexgrid.Coord) Fog {
	if p != CentralPowers && p != Entente || !s.Grid.Contains(c) {
		return Unseen
	}
	v := &s.vision[p]
	switch {
	case v.visible == nil:
		return Unseen
	case v.visible[c.Line][c.Col]:
		return Visible
	case v.seen[c.Line][c.Col]:
		return Seen
	}
	return Unseen
}

// UpdateVision recomputes the fog of war of p, as gui.py calculateFOW does: each
// map unit of p sees the hexes within its sight in steps (units inside
// containers see nothing), and each building p owns shows itself and its
// neighbours. Unlike Python, the building hexes are also marked as seen, so
// they stay Seen when the building is lost.
func (s *State) UpdateVision(p Player) {
	if p != CentralPowers && p != Entente {
		return
	}
	v := &s.vision[p]
	if v.visible == nil {
		v.seen = make([][]bool, len(s.Hexes))
		v.visible = make([][]bool, len(s.Hexes))
		for i := range s.Hexes {
			v.seen[i] = make([]bool, len(s.Hexes[i]))
			v.visible[i] = make([]bool, len(s.Hexes[i]))
		}
	}
	for i := range v.visible {
		clear(v.visible[i])
	}
	show := func(c hexgrid.Coord) {
		v.visible[c.Line][c.Col] = true
		v.seen[c.Line][c.Col] = true
	}
	for i := range s.Hexes {
		for j := range s.Hexes[i] {
			h, c := &s.Hexes[i][j], hexgrid.Coord{Line: i, Col: j}
			if h.Unit != nil && h.Unit.Owner == p {
				s.within(c, h.Unit.Def.Sight, show)
			}
			if h.Container != nil && h.Owner == p {
				show(c)
				for _, n := range s.Grid.Neighbors(c) {
					show(n)
				}
			}
		}
	}
}

// within calls f for every hex at most steps steps from c, c included. On this
// grid the step count equals hexgrid.Distance (tested in internal/hexgrid).
func (s *State) within(c hexgrid.Coord, steps int, f func(hexgrid.Coord)) {
	for l := max(0, c.Line-2*steps); l <= min(s.Grid.Lines-1, c.Line+2*steps); l++ {
		for col := range s.Hexes[l] {
			if n := (hexgrid.Coord{Line: l, Col: col}); hexgrid.Distance(c, n) <= steps {
				f(n)
			}
		}
	}
}

func (s *State) updateVision() {
	s.UpdateVision(CentralPowers)
	s.UpdateVision(Entente)
}
