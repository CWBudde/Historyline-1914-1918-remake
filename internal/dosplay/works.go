package dosplay

import (
	"errors"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The extension's work units beyond the gas pioneers (dosrules wire.go,
// tunnel.go, ambulance.go; docs/new-units.md): the wiring party lays barbed
// wire, the tunnellers dig a mine under a neighbouring hex, the field
// ambulance heals the foot units next to it. A flamethrower attacks as any
// unit does.

// Wiring reports whether unit lays barbed wire instead of attacking: its
// action order goes to a neighbouring hex of open ground, where it wires
// up to three hexes after the battles, and fights no battle.
func (s *Session) Wiring(unit int) bool { return s.Game.Units[unit].Wiring() }

// Tunneller reports whether unit digs mines instead of attacking: its
// action order goes to a neighbouring hex, under which it digs a mine that
// blows after dosrules.TunnelTurns turns while it stays.
func (s *Session) Tunneller(unit int) bool { return s.Game.Units[unit].Tunneller() }

// Ambulance reports whether unit is a field ambulance, which heals the foot
// units next to it once per turn.
func (s *Session) Ambulance(unit int) bool { return s.Game.Units[unit].Ambulance() }

// Flamethrower reports whether unit's attacks ignore the defender's cover.
func (s *Session) Flamethrower(unit int) bool { return s.Game.Units[unit].Flamethrower() }

// Worker reports whether unit's action order is work on a neighbouring
// hex, not an attack: a sapper's, a construction unit's, gas pioneers',
// a wiring party's or tunnellers'.
func (s *Session) Worker(unit int) bool {
	return s.Sapper(unit) || s.Builder(unit) || s.GasPioneer(unit) || s.Wiring(unit) || s.Tunneller(unit)
}

// canWork is why a work unit of side has no work order now (no charge
// left, a tunnel dug already), nil when it has one somewhere: its own cell
// is never a target, so a refusal for no target there passes.
func (s *Session) canWork(side, unit int) error {
	u := s.Game.Units[unit]
	err := s.Game.Check(dosrules.Command{Kind: dosrules.CmdAttack, Side: side, Unit: unit, Cell: u.Cell(side)})
	var r *dosrules.Refusal
	if errors.As(err, &r) && r.Code == dosrules.MsgNoTarget {
		return nil
	}
	return wrap(err)
}

// CanWire is why wiring party unit of side can lay no wire now (none
// left), nil when it can.
func (s *Session) CanWire(side, unit int) error {
	if u, ok := s.Game.Units[unit]; !ok || !u.Wiring() {
		return refusal("not a wiring party")
	}
	return s.canWork(side, unit)
}

// CanTunnel is why tunnellers unit of side can start no tunnel now (one
// dug already), nil when they can.
func (s *Session) CanTunnel(side, unit int) error {
	if u, ok := s.Game.Units[unit]; !ok || !u.Tunneller() {
		return refusal("not tunnellers")
	}
	return s.canWork(side, unit)
}

// WireHexes are the hexes wiring party unit of side would wire from its
// hex on hex to, without ordering it; nil where it may not.
func (s *Session) WireHexes(side, unit int, to hexgrid.Coord) []hexgrid.Coord {
	cell, ok := s.Board.Cell(to)
	if !ok {
		return nil
	}
	return s.levelHexes(s.Game.WireCells(side, unit, cell))
}

// Tunnel is a mine side's tunnellers dig: the tunnellers (Unit), the hex
// they dig from and the hex the mine lies under, the turns dug and the
// turns it takes.
type Tunnel struct {
	Unit       int
	From, Mine hexgrid.Coord
	Dug, Turns int
}

// Tunnels are the mines side digs; the other side's stay hidden.
func (s *Session) Tunnels(side int) []Tunnel {
	var out []Tunnel
	for _, t := range s.Game.Tunnels {
		if t.Side == side {
			out = append(out, Tunnel{
				Unit: t.Unit, From: s.coord(t.From), Mine: s.coord(t.Cell), Dug: t.Dug, Turns: dosrules.TunnelTurns,
			})
		}
	}
	return out
}

// TunnelOf is the mine unit digs and whether it digs one.
func (s *Session) TunnelOf(unit int) (Tunnel, bool) {
	t, ok := s.Game.TunnelOf(unit)
	if !ok {
		return Tunnel{}, false
	}
	return Tunnel{Unit: unit, From: s.coord(t.From), Mine: s.coord(t.Cell), Dug: t.Dug, Turns: dosrules.TunnelTurns}, true
}
