package dosrules

import (
	"encoding/json"
	"fmt"
)

// SaveSchema is the version of the saved command-cycle state.
const SaveSchema = 2

// A saved game is the whole state of the cycle as JSON: the profile by its
// tables' build, both maps, the live and container records, each side's
// phase, confirmation and order list, the turn, the game options, the energy
// accounts and the buildings. Orders given but not confirmed and confirmed
// orders waiting for the other side are both part of that state (the order
// lists, the units' moved flags and targets), so a loaded game resolves
// exactly as the saved one would have.
type savedGame struct {
	Schema    int               `json:"schema"`
	Build     string            `json:"build"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Maps      [2][]int          `json:"maps"`
	Units     map[int]savedUnit `json:"units"`
	Holds     map[int]Hold      `json:"holds"`
	Players   [2]savedPlayer    `json:"players"`
	Turn      int               `json:"turn"`
	Limit     int               `json:"limit"`
	Concealed bool              `json:"concealed,omitempty"`
	Ended     *End              `json:"ended,omitempty"`
	Energy    [2]int            `json:"energy"`
	Buildings map[int]Building  `json:"buildings,omitempty"`
}

type savedPlayer struct {
	Phase    int   `json:"phase"`
	Ready    bool  `json:"ready,omitempty"`
	Computer bool  `json:"computer,omitempty"`
	Orders   []int `json:"orders,omitempty"`
}

type savedUnit struct {
	Budget int    `json:"budget"`
	Exp    int    `json:"exp"`
	Size   int    `json:"size"`
	Alt    int    `json:"alt"`
	Flags4 int    `json:"flags4"`
	Flags6 int    `json:"flags6"`
	Type   int    `json:"type"`
	Serial int    `json:"serial"`
	Hold   int    `json:"hold"`
	Pos    [2]int `json:"pos"`
	Facing [2]int `json:"facing"`
	Target int    `json:"target"`
	Action [7]int `json:"action"`
}

// Save returns the game's state as JSON.
func (g *Game) Save() ([]byte, error) {
	if g.Profile.Tables == nil {
		return nil, fmt.Errorf("dosrules: save: no tables")
	}
	s := savedGame{Schema: SaveSchema, Build: g.Profile.Tables.Build, Width: g.Width, Height: g.Height,
		Maps: g.Maps, Units: map[int]savedUnit{}, Holds: g.Holds, Turn: g.Turn, Limit: g.Limit, Concealed: g.Concealed, Ended: g.Ended,
		Energy: g.Energy, Buildings: g.Buildings}
	for i, u := range g.Units {
		s.Units[i] = savedUnit(u)
	}
	for k, p := range g.Players {
		s.Players[k] = savedPlayer(p)
	}
	return json.Marshal(s)
}

// LoadGame reads a game saved by Save. It refuses a save of another schema,
// of tables it does not know, or whose maps, units, order lists and phases
// do not fit together.
func LoadGame(data []byte) (*Game, error) {
	var s savedGame
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("dosrules: load: %w", err)
	}
	if s.Schema != SaveSchema {
		return nil, fmt.Errorf("dosrules: load: schema %d, want %d", s.Schema, SaveSchema)
	}
	var p Profile
	for _, q := range []Profile{Full, Demo} {
		if q.Tables.Build == s.Build {
			p = q
		}
	}
	if p.Tables == nil {
		return nil, fmt.Errorf("dosrules: load: unknown tables %q", s.Build)
	}
	g := &Game{Profile: p, Width: s.Width, Height: s.Height, Maps: s.Maps, Units: map[int]Live{},
		Holds: map[int]Hold{}, Turn: s.Turn, Limit: s.Limit, Concealed: s.Concealed, Ended: s.Ended,
		Energy: s.Energy, Buildings: s.Buildings}
	if s.Width < 1 || s.Height < 1 {
		return nil, fmt.Errorf("dosrules: load: %d×%d map", s.Width, s.Height)
	}
	for i, u := range s.Units {
		if i < 0 || i > SlotFree || u.Type < 0 || u.Type >= len(p.Tables.Types) {
			return nil, fmt.Errorf("dosrules: load: unit %d of type %d", i, u.Type)
		}
		g.Units[i] = Live(u)
	}
	for i, h := range s.Holds {
		g.Holds[i] = h
	}
	for k, m := range s.Maps {
		if len(m) != 2*s.Width*s.Height {
			return nil, fmt.Errorf("dosrules: load: map %d has %d entries for %d cells", k, len(m), s.Width*s.Height)
		}
		for c := range s.Width * s.Height {
			if occ := m[2*c+1]; occ != 0xFF {
				if _, ok := g.Units[occ]; !ok {
					return nil, fmt.Errorf("dosrules: load: map %d shows unit %d at %d", k, occ, c)
				}
			}
		}
	}
	for k, sp := range s.Players {
		for _, u := range sp.Orders {
			if _, ok := g.Units[u]; !ok {
				return nil, fmt.Errorf("dosrules: load: side %d orders unit %d", k, u)
			}
		}
		g.Players[k] = Player(sp)
	}
	if _, err := g.Action(); err != nil {
		return nil, fmt.Errorf("dosrules: load: %w", err)
	}
	return g, nil
}
