// Package game is the rules engine: map, units, containers and players. It has
// no rendering code; the UI and the AI drive it through commands.
package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Player is a side, numbered as in hlrData.py Info.nameDict and the level
// files' owner field.
type Player int

const (
	None          Player = 0
	CentralPowers Player = 1
	Entente       Player = 2
)

var playerNames = map[string]Player{"None": None, "Central Powers": CentralPowers, "Entente Cordial": Entente}

// PlayerByName maps the level file's player string to a Player.
func PlayerByName(name string) (Player, error) {
	p, ok := playerNames[name]
	if !ok {
		return None, fmt.Errorf("unknown player %q", name)
	}
	return p, nil
}

func (p Player) String() string {
	for name, q := range playerNames {
		if p == q {
			return name
		}
	}
	return fmt.Sprintf("Player(%d)", int(p))
}

// FactionOf returns the side of a unit from its country, as in hlrData.py Unit.
func FactionOf(country string) Player {
	switch country {
	case "Germany", "Austria", "Bulgaria", "Ottoman":
		return CentralPowers
	}
	return Entente
}

// Unit is one unit on the map or inside a container.
type Unit struct {
	Type       string // unit id, a key of data.Set.Units
	Def        data.Unit
	Owner      Player
	Size       int // current size, 0..MaxSize
	MaxSize    int
	Experience int // rank 0..10
	Fuel       int
	Weapons    [4]*data.Weapon // per weapon slot; nil for an empty slot
	Ammo       [4]*int         // per weapon slot; nil for an empty slot or unlimited ammo
	Moved      bool
	Attacked   bool
	Content    *Container // nil unless the unit type can carry units
}

// NewUnit creates a fresh unit of type id at the full strength of its original
// group, with no experience and full fuel and ammo.
func NewUnit(set *data.Set, id string) (*Unit, error) {
	def, ok := set.Units[id]
	if !ok {
		return nil, fmt.Errorf("unknown unit %q", id)
	}
	u := &Unit{Type: id, Def: def, Owner: FactionOf(def.Country), Size: def.Group, MaxSize: def.Group, Fuel: def.Fuel}
	for i, sl := range def.Weapons {
		if sl == nil {
			continue
		}
		wd := set.SlotWeapon(sl)
		u.Weapons[i] = &wd
		if a := wd.Ammo; a != nil {
			v := *a
			u.Ammo[i] = &v
		}
	}
	if def.StorageMax > 0 {
		u.Content = &Container{Capacity: def.StorageMax}
	}
	return u, nil
}

// Container holds units in a 2 × 9 grid with a weight capacity: an HQ, depot,
// factory, or a unit with storage. Slots[row][col].
type Container struct {
	Capacity int
	Slots    [2][9]*Unit
}

// Used returns the total weight of the contained units.
func (c *Container) Used() int {
	w := 0
	for _, u := range c.Units() {
		w += u.Def.Weight
	}
	return w
}

// Fits reports whether u can be added without exceeding the capacity.
func (c *Container) Fits(u *Unit) bool {
	return u.Def.Weight+c.Used() <= c.Capacity
}

// Add puts u into the first free slot in the order of Content.addUnit: column
// by column, row 0 before row 1. It does not check the weight;
// callers check Fits.
func (c *Container) Add(u *Unit) error {
	for col := range 9 {
		for row := range 2 {
			if c.Slots[row][col] == nil {
				c.Slots[row][col] = u
				return nil
			}
		}
	}
	return errors.New("container is full")
}

// Remove frees the slot of u and reports whether u was contained.
func (c *Container) Remove(u *Unit) bool {
	for row := range 2 {
		for col := range 9 {
			if c.Slots[row][col] == u {
				c.Slots[row][col] = nil
				return true
			}
		}
	}
	return false
}

// Units returns the contained units, row 0 before row 1, as getAllUnits lists them.
func (c *Container) Units() []*Unit {
	var out []*Unit
	for row := range 2 {
		for col := range 9 {
			if u := c.Slots[row][col]; u != nil {
				out = append(out, u)
			}
		}
	}
	return out
}

// FillOrder returns the contained units in the order Add fills the slots.
func (c *Container) FillOrder() []*Unit {
	var out []*Unit
	for col := range 9 {
		for row := range 2 {
			if u := c.Slots[row][col]; u != nil {
				out = append(out, u)
			}
		}
	}
	return out
}

// Hex is one map hex.
type Hex struct {
	Terrain   string
	Infra     []string
	Mods      data.Hex // terrain modifiers after infrastructure
	Unit      *Unit
	Owner     Player     // owner of a building; None elsewhere
	Container *Container // non-nil for an HQ, depot or factory
}

// State is a game in progress.
type State struct {
	Name   string
	No     int
	Grid   hexgrid.Grid
	Hexes  [][]Hex // Hexes[line][col]
	Human  Player  // the side named by the level, which moves first
	Turn   Player  // side to move
	Round  int
	Winner Player // side that captured an HQ; None while the game runs

	vision [3]vision // fog of war, indexed by Player
}

// At returns the hex at c, or nil when c is off the map.
func (s *State) At(c hexgrid.Coord) *Hex {
	if !s.Grid.Contains(c) {
		return nil
	}
	return &s.Hexes[c.Line][c.Col]
}

// Units lists the units of p (all units for None): map units in line order,
// each followed by its contents, then the contents of buildings.
func (s *State) Units(p Player) []*Unit {
	var out []*Unit
	var add func(u *Unit)
	add = func(u *Unit) {
		if p == None || u.Owner == p {
			out = append(out, u)
		}
		if u.Content != nil {
			for _, v := range u.Content.Units() {
				add(v)
			}
		}
	}
	for i := range s.Hexes {
		for j := range s.Hexes[i] {
			h := &s.Hexes[i][j]
			if h.Unit != nil {
				add(h.Unit)
			}
			if h.Container != nil {
				for _, u := range h.Container.Units() {
					add(u)
				}
			}
		}
	}
	return out
}

// buildingCapacity returns the capacity of a building hex, or 0 for other
// terrain. HQ, depot and factoryE are the hexes HexSquare treats as buildings;
// other factory pieces count only when the level gives them content.
func buildingCapacity(terrain string, hasContent bool) int {
	switch strings.TrimSuffix(terrain, "_w") {
	case "hqN":
		return 50
	case "cmpN", "factoryE":
		return 40
	}
	if hasContent && strings.HasPrefix(terrain, "factory") {
		return 40
	}
	return 0
}

// New builds the starting state of a level.
func New(lv *level.Level, set *data.Set) (*State, error) {
	human, err := PlayerByName(lv.Player)
	if err != nil {
		return nil, err
	}
	s := &State{
		Name: lv.MapName, No: lv.MapNo, Grid: lv.Grid(), Human: human, Turn: human, Round: 1,
		Hexes: make([][]Hex, len(lv.Lines)),
	}
	for i, line := range lv.Lines {
		s.Hexes[i] = make([]Hex, len(line))
		for col, lh := range line {
			h, err := newHex(set, lh)
			if err != nil {
				return nil, fmt.Errorf("line%d[%d]: %w", i+1, col, err)
			}
			s.Hexes[i][col] = h
		}
	}
	s.updateVision()
	return s, nil
}

func newHex(set *data.Set, lh level.Hex) (Hex, error) {
	t, ok := set.Terrain[lh.Terrain]
	if !ok {
		return Hex{}, fmt.Errorf("unknown terrain %q", lh.Terrain)
	}
	h := Hex{Terrain: lh.Terrain, Infra: lh.Infra.IDs, Mods: data.FixedModifiers(t, lh.Infra.IDs)}
	if lh.Unit != "" {
		u, err := NewUnit(set, lh.Unit)
		if err != nil {
			return Hex{}, err
		}
		h.Unit = u
	}
	capacity := buildingCapacity(lh.Terrain, lh.Content != nil)
	if capacity == 0 {
		if lh.Content != nil {
			return Hex{}, fmt.Errorf("%s cannot hold content", lh.Terrain)
		}
		return h, nil
	}
	h.Container = &Container{Capacity: capacity}
	if lh.Content == nil {
		return h, nil
	}
	h.Owner = Player(lh.Content.Owner)
	if h.Owner < None || h.Owner > Entente {
		return Hex{}, fmt.Errorf("unknown owner %d", lh.Content.Owner)
	}
	for _, id := range lh.Content.Units {
		u, err := NewUnit(set, id)
		if err != nil {
			return Hex{}, err
		}
		if err := h.Container.Add(u); err != nil {
			return Hex{}, err
		}
	}
	return h, nil
}
