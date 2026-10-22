package dosplay

import (
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// The data of the overlays: the unit info (TypeInfo), the game info (Stats)
// and the contents with repair and production (CanRepair, CanBuild,
// Buildable).

// Weapon is a type's attack on one target class: its power and the range
// the unit info shows (UNIT.DAT stores one more).
type Weapon struct{ Power, Range int }

// TypeView is a unit type as the unit info shows it for a side: the name of
// the side's unit of the type, the weapon against air, ground and water
// targets (the original has one surface range, so ground and water reach
// alike), armour, speed, weight, the full group and the production price
// (-1 when never sold) and its battle domain.
type TypeView struct {
	Type                         int
	ID                           string
	Name                         string
	Domain                       Domain
	Air, Ground, Water           Weapon
	Armour, Speed, Weight, Group int
	Cost                         int
}

// TypeInfo is UNIT.DAT type typ as side's unit info shows it; false for no
// such type.
func (s *Session) TypeInfo(side, typ int) (TypeView, bool) {
	types := s.Game.Profile.Tables.Types
	if typ < 0 || typ >= len(types) {
		return TypeView{}, false
	}
	t := types[typ]
	v := TypeView{Type: typ, ID: s.typeID(typ, side), Air: Weapon{t.AirPower, t.AirRange - 1},
		Ground: Weapon{t.GroundPower, t.SurfaceRange - 1}, Water: Weapon{t.WaterPower, t.SurfaceRange - 1},
		Armour: t.Armour, Group: t.Group, Cost: t.Cost, Domain: TypeDomain(s.Game.Profile.Tables, typ)}
	v.Name = s.typeName(v.ID, typ)
	if m := s.Game.Profile.Tables.Move; m != nil && typ < len(m.Types) {
		v.Speed, v.Weight = m.Types[typ].Speed, m.Types[typ].Weight
	}
	if t.Cost == neverSold {
		v.Cost = -1
	}
	return v, true
}

// neverSold is the price of a type no factory sells.
const neverSold = 0xFF

// BuildingKind is what a building record is.
type BuildingKind int

// The building kinds, by container record: HQs, depots and factories.
const (
	HQ BuildingKind = iota
	Depot
	Factory
)

var buildingKindNames = [...]string{"hq", "depot", "factory"}

func (k BuildingKind) String() string {
	if k >= 0 && int(k) < len(buildingKindNames) {
		return buildingKindNames[k]
	}
	return fmt.Sprintf("BuildingKind(%d)", int(k))
}

// kindOf is the kind of building record hold (HoldFlag40 and up).
func kindOf(hold int) BuildingKind {
	switch {
	case hold < dosrules.HoldFlag100:
		return HQ
	case hold < dosrules.HoldFlag400:
		return Depot
	}
	return Factory
}

// Holdings are what one owner has: its units not destroyed (held ones too),
// depots and factories.
type Holdings struct{ Units, Depots, Factories int }

// add counts h into the holdings.
func (a *Holdings) add(h Holdings) {
	a.Units += h.Units
	a.Depots += h.Depots
	a.Factories += h.Factories
}

// Stats are the numbers of the game info: the round (the turn counter, the
// command cycles completed), the level's name and number, each side's mode
// (its phase), the current score, the holdings of side 0, side 1 and the
// neutral owner and of the whole map, each side's energy account and the
// income its buildings pay at every phase change, the orders each side has
// given this phase (repairs and builds do not count) and the move limit (0:
// none).
//
// The original's game info shows both sides' holdings, energy and income to
// either side, so Stats is not per side and counts the true state, not a
// side's map. The original's third column is the whole map (Map), which
// its manual calls what neither side owns (Sides[2]).
type Stats struct {
	Round   int
	Level   string
	LevelNo int
	Mode    [2]int
	Score   int
	Sides   [3]Holdings
	Map     Holdings
	Energy  [2]int
	Income  [2]int
	Orders  [2]int
	Limit   int
}

// Stats are the game info's numbers.
func (s *Session) Stats() Stats {
	g := s.Game
	st := Stats{Round: g.Turn, Level: s.Level.MapName, LevelNo: s.Level.MapNo, Score: g.Score(), Energy: g.Energy,
		Limit: g.Limit}
	for k := range 2 {
		st.Mode[k] = g.Players[k].Phase
		st.Orders[k] = len(g.Players[k].Orders)
		st.Income[k] = g.Income(k)
	}
	for _, u := range g.Units {
		if !u.Destroyed() {
			st.Sides[sideOf(u.Flags4)].Units++
		}
	}
	for _, b := range s.buildings(-1) {
		switch h := &st.Sides[s.SideOf(b.Owner)]; b.Kind {
		case Depot:
			h.Depots++
		case Factory:
			h.Factories++
		}
	}
	for _, h := range st.Sides {
		st.Map.add(h)
	}
	return st
}

// CanRepair reports whether Repair would take the unit in slot of side's
// part of building record hold now, without repairing it.
func (s *Session) CanRepair(side, hold, slot int) error {
	return wrap(s.Game.Check(dosrules.Command{Kind: dosrules.CmdRepair, Side: side, Hold: hold, Slot: slot}))
}

// CanBuild reports whether Build would build a unit of type typ in building
// record hold now, without building it.
func (s *Session) CanBuild(side, hold, typ int) error {
	return wrap(s.Game.Check(dosrules.Command{Kind: dosrules.CmdBuild, Side: side, Hold: hold, Type: typ}))
}

// BuildOption is a type a factory builds, named for the side, and its
// price.
type BuildOption struct {
	Type     int
	ID, Name string
	Cost     int
}

// Buildable are the types building record hold builds, in the scenario's
// order, as side sees them; none for a building that only repairs. Whether
// the side may build one now is CanBuild's.
func (s *Session) Buildable(side, hold int) []BuildOption {
	types := s.Game.Profile.Tables.Types
	b, ok := s.Game.Buildings[hold]
	if side != 0 && side != 1 || !ok || hold < dosrules.HoldFlag40 {
		return nil
	}
	var out []BuildOption
	for _, t := range b.Builds {
		if t < 0 || t >= len(types) || types[t].Cost == neverSold {
			continue
		}
		id := s.typeID(t, side)
		out = append(out, BuildOption{Type: t, ID: id, Name: s.typeName(id, t), Cost: types[t].Cost})
	}
	return out
}
