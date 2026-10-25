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
	memory        [][]Memory // each hex as it was when last visible
}

// Memory is what a player remembers of a hex: its map unit and building owner
// as they were when the hex was last visible. Units inside containers are not
// remembered.
type Memory struct {
	Unit  *UnitMemory // nil when the hex had no map unit
	Owner Player      // building owner; None elsewhere
}

// UnitMemory is a copy of the parts of a unit a player saw; it does not follow
// the live unit.
type UnitMemory struct {
	Type  string // unit id, a key of data.Set.Units
	Owner Player
	Size  int
}

// remember returns the current Memory of h.
func remember(h *Hex) Memory {
	var m Memory
	if u := h.Unit; u != nil {
		m.Unit = &UnitMemory{Type: u.Type, Owner: u.Owner, Size: u.Size}
	}
	if h.Container != nil {
		m.Owner = h.Owner
	}
	return m
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

// Remembered returns what p knows of the hex at c: the live hex when it is
// Visible, the last-seen Memory when it is Seen. ok is false for Unseen hexes,
// off-map hexes and None.
func (s *State) Remembered(p Player, c hexgrid.Coord) (m Memory, ok bool) {
	switch s.Fog(p, c) {
	case Visible:
		return remember(s.At(c)), true
	case Seen:
		if v := &s.vision[p]; v.memory != nil {
			m = v.memory[c.Line][c.Col]
		}
		return m, true
	}
	return Memory{}, false
}

// UpdateVision recomputes the fog of war of p: each
// map unit of p sees the hexes within its sight in steps (units inside
// containers see nothing), and each building p owns shows itself and its
// neighbours. The building hexes are also marked as seen, so
// they stay Seen when the building is lost. Every visible hex refreshes its
// Memory.
func (s *State) UpdateVision(p Player) {
	if p != CentralPowers && p != Entente {
		return
	}
	v := &s.vision[p]
	if v.visible == nil {
		v.seen = make([][]bool, len(s.Hexes))
		v.visible = make([][]bool, len(s.Hexes))
		v.memory = make([][]Memory, len(s.Hexes))
		for i := range s.Hexes {
			v.seen[i] = make([]bool, len(s.Hexes[i]))
			v.visible[i] = make([]bool, len(s.Hexes[i]))
			v.memory[i] = make([]Memory, len(s.Hexes[i]))
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
	for i := range v.visible {
		for j, ok := range v.visible[i] {
			if ok {
				v.memory[i][j] = remember(&s.Hexes[i][j])
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
