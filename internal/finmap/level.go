package finmap

import (
	"errors"
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Format and Version identify the level JSON.
const (
	Format  = "historyline-level"
	Version = 1
)

// Rules are the fields of a build's tables that loading a map reads.
type Rules struct {
	// TerrainFlags is word 0 of each terrain record: 0x40 an HQ entrance,
	// 0x100 a depot entrance, 0x400 a factory entrance; 0x02 marks neutral,
	// else 0x01 side 1, neither side 0.
	TerrainFlags []int
	// TypeClass is word +0x0d of each unit type: 0x1000 a transporter,
	// 0x40 a ship two cells long (0x80 its head), 0x02 a neutral unit.
	TypeClass []int
}

// RulesFrom takes the rules from a build's tables (nil fields without the
// movement records).
func RulesFrom(t *dosrules.Tables) Rules {
	var r Rules
	if t == nil || t.Move == nil {
		return r
	}
	for _, tr := range t.Move.Terrains {
		r.TerrainFlags = append(r.TerrainFlags, tr.Flags)
	}
	for _, ty := range t.Move.Types {
		r.TypeClass = append(r.TypeClass, ty.Class)
	}
	return r
}

// Type class and terrain flag bits.
const (
	ClassTransporter = 0x1000
	ClassTwoCell     = 0x40
	ClassHead        = 0x80
	ClassNeutral     = 0x02

	FlagHQ      = 0x40
	FlagDepot   = 0x100
	FlagFactory = 0x400
)

// MaxUnits is the number of live unit records.
const MaxUnits = 0xf1

// Input are the decoded files of one map.
type Input struct {
	Index    int       // the map number (CODES.DAT record); -1 for none
	Map      *Map      // required
	Shop     *Shop     // nil without an SHP file
	Computer *Computer // nil without a COM file
	Code     *Code     // nil without a CODES.DAT record
}

// Level is version 1 of the imported level JSON.
type Level struct {
	Format  string   `json:"format"`
	Version int      `json:"version"`
	Source  Source   `json:"source"`
	Mission *Mission `json:"mission,omitempty"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	// Terrain is the terrain record of every cell, Terrain[row][col].
	Terrain [][]int `json:"terrain"`
	// Units are the units on the map in the order the game creates them.
	Units []Unit `json:"units"`
	// Buildings are the HQs, depots and factories by entrance cell.
	Buildings []Building `json:"buildings"`
	// Containers are the SHP records in file order; null without an SHP.
	Containers []Container `json:"containers"`
	// Producible are the unit types factories may produce on the map
	// (also subject to cost, side and free holds); null without an SHP.
	Producible []int `json:"producible"`
	// Computer reports a COM file: the parameters of the computer player,
	// which the level does not carry (not decoded).
	Computer bool `json:"computer"`
}

// Source names where a level came from; the importer fills it.
type Source struct {
	Build string   `json:"build"` // "full" or "demo"
	Map   string   `json:"map"`   // the file stem, "00" or "STATMAP"
	Files []string `json:"files"`
}

// Mission is the map's CODES.DAT record and the sides it implies.
type Mission struct {
	Code    string `json:"code"` // the map's password, its code in the menu
	Players int    `json:"players"`
	// HumanSides: with one player, maps below 24 give side 1 to the
	// computer and the others side 0; with two players both are human.
	HumanSides  []int `json:"humanSides"`
	Winter      bool  `json:"winter"`
	Next        *int  `json:"next"` // the map a win leads to; null ends the series
	CampaignEnd bool  `json:"campaignEnd"`
}

// Unit is a unit on the map.
type Unit struct {
	Cell [2]int `json:"cell"`
	Type int    `json:"type"`
	Side int    `json:"side"`
	// Rear is a two-cell ship's second cell (type Type+1): one row below the
	// head for side 1, one row above for side 0.
	Rear *[2]int `json:"rear,omitempty"`
	// Container is a transporter's hold: the transporters take holds 0, 1,
	// ... in creation order.
	Container *int `json:"container,omitempty"`
}

// Building is an HQ, depot or factory, found by its entrance cell (the cell
// whose terrain carries the building flag); buildings of a kind take their
// table entries in row-major order of their entrances.
type Building struct {
	Kind      string `json:"kind"`
	Cell      [2]int `json:"cell"`
	Owner     int    `json:"owner"`
	Container int    `json:"container"`
}

// Container is an SHP record applied to its container, in file order. A
// building without a record earns no energy, an HQ without one 50. Energy
// is the building's credit per command cycle (a transporter's record
// carries 0 there).
type Container struct {
	Container     int     `json:"container"`
	Kind          string  `json:"kind"`
	Index         int     `json:"index"`
	Side          int     `json:"side"`
	Energy        int     `json:"energy"`
	ComputerValue int     `json:"computerValue"`
	Units         []Cargo `json:"units"`
	// Overridden marks a record that a later record for the same
	// container replaces: its energy and computer value no longer apply.
	Overridden bool `json:"overridden,omitempty"`
	// Holder is what holds the container: "building", "unit" (a
	// transporter on the map), "cargo" (a transporter created by a record)
	// or "none".
	Holder string `json:"holder"`
}

// Cargo is a unit a record creates; a transporter gets the next free hold.
type Cargo struct {
	Type      int  `json:"type"`
	Container *int `json:"container,omitempty"`
}

// Build assembles the level JSON from a map's decoded files.
func Build(in Input, r Rules) (*Level, error) {
	m := in.Map
	switch {
	case m == nil:
		return nil, errors.New("finmap: no map")
	case m.Width < 1 || m.Height < 1 || len(m.Terrain) != m.Width*m.Height || len(m.Occupant) != m.Width*m.Height:
		return nil, fmt.Errorf("finmap: %dx%d map with %d terrain and %d occupant cells", m.Width, m.Height, len(m.Terrain), len(m.Occupant))
	case in.Code != nil && in.Index < 0:
		return nil, errors.New("finmap: a CODES.DAT record without a map number")
	}
	l := &Level{
		Format: Format, Version: Version, Width: m.Width, Height: m.Height,
		Terrain: make([][]int, m.Height), Units: []Unit{}, Buildings: []Building{},
	}
	for row := range m.Height {
		l.Terrain[row] = make([]int, m.Width)
		for col := range m.Width {
			t := m.TerrainAt(row, col)
			if t >= len(r.TerrainFlags) {
				return nil, fmt.Errorf("finmap: cell %d,%d terrain %d", row, col, t)
			}
			l.Terrain[row][col] = t
		}
	}
	if err := l.addBuildings(m, r); err != nil {
		return nil, err
	}
	created, holds, err := l.addUnits(m, r)
	if err != nil {
		return nil, err
	}
	if in.Shop != nil {
		if err := l.addShop(in.Shop, r, holds); err != nil {
			return nil, err
		}
	}
	for _, c := range l.Containers {
		created += len(c.Units)
	}
	if created > MaxUnits {
		return nil, fmt.Errorf("finmap: %d units, at most %d", created, MaxUnits)
	}
	l.Computer = in.Computer != nil
	if c := in.Code; c != nil {
		l.Mission = &Mission{Code: c.Password, Players: c.Players, Winter: c.Winter, CampaignEnd: c.CampaignEnd}
		switch {
		case c.Players == 2:
			l.Mission.HumanSides = []int{Central, Allied}
		case in.Index < 24:
			l.Mission.HumanSides = []int{Central}
		default:
			l.Mission.HumanSides = []int{Allied}
		}
		if c.Step != 0 {
			next := in.Index + c.Step
			l.Mission.Next = &next
		}
	}
	return l, nil
}

func ownerOf(flags int) int {
	switch {
	case flags&2 != 0:
		return Neutral
	case flags&1 != 0:
		return Allied
	}
	return Central
}

func (l *Level) addBuildings(m *Map, r Rules) error {
	var depots, factories int
	hqs := map[int]bool{}
	for row := range m.Height {
		for col := range m.Width {
			f := r.TerrainFlags[m.TerrainAt(row, col)]
			b := Building{Cell: [2]int{row, col}, Owner: ownerOf(f)}
			switch {
			case f&FlagFactory != 0:
				b.Kind, b.Container = Factory.String(), FactoryBase+factories
				factories++
			case f&FlagDepot != 0:
				b.Kind, b.Container = Depot.String(), DepotBase+depots
				depots++
			case f&FlagHQ != 0:
				if b.Owner == Neutral || hqs[b.Owner] {
					return fmt.Errorf("finmap: cell %d,%d: a second or neutral HQ (owner %d)", row, col, b.Owner)
				}
				hqs[b.Owner] = true
				b.Kind, b.Container = HQ.String(), HQBase+b.Owner
			default:
				continue
			}
			l.Buildings = append(l.Buildings, b)
		}
	}
	if depots > TableBuilding || factories > TableBuilding {
		return fmt.Errorf("finmap: %d depots and %d factories, the tables hold %d each", depots, factories, TableBuilding)
	}
	return nil
}

// addUnits creates the map's units in scan order and returns the number of
// units it created (ship rears included) and of holds the transporters took.
func (l *Level) addUnits(m *Map, r Rules) (int, int, error) {
	rears := map[[2]int]bool{} // rear cells a head claimed
	holds := 0
	for row := range m.Height {
		for col := range m.Width {
			occ := m.OccupantAt(row, col)
			if occ >= EmptyFrom {
				continue
			}
			t, side := occ>>1, occ&1
			if t >= TypeCount || t >= len(r.TypeClass) {
				return 0, 0, fmt.Errorf("finmap: cell %d,%d occupant %#02x", row, col, occ)
			}
			class := r.TypeClass[t]
			if class&ClassTwoCell != 0 && class&ClassHead == 0 {
				continue // a rear; its head creates it
			}
			u := Unit{Cell: [2]int{row, col}, Type: t, Side: side}
			if class&ClassTwoCell != 0 {
				if t+1 >= len(r.TypeClass) || r.TypeClass[t+1]&(ClassTwoCell|ClassHead) != ClassTwoCell {
					return 0, 0, fmt.Errorf("finmap: ship at %d,%d: type %d is no rear", row, col, t+1)
				}
				rr := row - 1
				if side == Allied {
					rr = row + 1
				}
				if rr < 0 || rr >= m.Height || m.OccupantAt(rr, col) != 2*(t+1)+side {
					return 0, 0, fmt.Errorf("finmap: ship at %d,%d has no rear at %d,%d", row, col, rr, col)
				}
				u.Rear = &[2]int{rr, col}
				rears[*u.Rear] = true
			} else if class&ClassNeutral != 0 {
				u.Side = Neutral
			}
			if class&ClassTransporter != 0 {
				if holds == Transporters {
					return 0, 0, fmt.Errorf("finmap: more than %d transporters", Transporters)
				}
				h := holds
				u.Container = &h
				holds++
			}
			l.Units = append(l.Units, u)
		}
	}
	for row := range m.Height {
		for col := range m.Width {
			occ := m.OccupantAt(row, col)
			if occ >= EmptyFrom {
				continue
			}
			c := r.TypeClass[occ>>1]
			if c&ClassTwoCell != 0 && c&ClassHead == 0 && !rears[[2]int{row, col}] {
				return 0, 0, fmt.Errorf("finmap: ship rear at %d,%d without a head", row, col)
			}
		}
	}
	return len(l.Units) + len(rears), holds, nil
}

func (l *Level) addShop(s *Shop, r Rules, holds int) error {
	l.Containers = []Container{}
	l.Producible = []int{}
	for t, locked := range s.Locked {
		if !locked {
			l.Producible = append(l.Producible, t)
		}
	}
	holder := map[int]string{}
	for _, b := range l.Buildings {
		holder[b.Container] = "building"
	}
	for _, u := range l.Units {
		if u.Container != nil {
			holder[*u.Container] = "unit"
		}
	}
	for _, rec := range s.Records {
		c := Container{
			Container: rec.Container(), Kind: rec.Kind.String(), Index: rec.Index, Side: rec.Side,
			Energy: rec.Energy, ComputerValue: rec.ComputerValue, Units: []Cargo{},
		}
		for _, t := range rec.Units() {
			cargo := Cargo{Type: t}
			if t < len(r.TypeClass) && r.TypeClass[t]&ClassTransporter != 0 {
				if holds == Transporters {
					return fmt.Errorf("finmap: more than %d transporters", Transporters)
				}
				h := holds
				cargo.Container = &h
				holder[h] = "cargo"
				holds++
			}
			c.Units = append(c.Units, cargo)
		}
		l.Containers = append(l.Containers, c)
	}
	last := map[int]int{}
	for i, c := range l.Containers {
		last[c.Container] = i
	}
	for i := range l.Containers {
		l.Containers[i].Overridden = last[l.Containers[i].Container] != i
		h, ok := holder[l.Containers[i].Container]
		if !ok {
			h = "none"
		}
		l.Containers[i].Holder = h
	}
	return nil
}
