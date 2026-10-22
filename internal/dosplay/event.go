package dosplay

import (
	"fmt"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// EventKind names an event.
type EventKind int

// The events of a session.
const (
	Moved        EventKind = iota // a move or an exit ran: Path, From, To
	Refused                       // the order code refused a move or an exit: Text
	Ordered                       // an attack order entered the list: From the attacker, To the target
	Confirmed                     // the side confirmed its orders
	Unconfirmed                   // the side withdrew its confirmation
	Battle                        // an attack order was fought: Battle, From, To
	Skipped                       // an attack order whose hex held no unit: To
	Revealed                      // a move of the movement side, shown to the other after the phase change: Path
	PhaseChanged                  // the phases swapped: Phase, Turn
	Credited                      // a side got its buildings' energy: Energy
	Repaired                      // a unit was repaired: Unit, To the building, Energy left
	Built                         // a unit was built: Unit, To the building, Energy left
	Ended                         // the battle ended: Side the winner (-1 a draw), End
	Captured                      // Side took a building with Unit: To the building, Hold its record, Text
)

var kindNames = [...]string{"moved", "refused", "ordered", "confirmed", "unconfirmed", "battle", "skipped",
	"revealed", "phase", "credited", "repaired", "built", "ended", "captured"}

func (k EventKind) String() string {
	if k >= 0 && int(k) < len(kindNames) {
		return kindNames[k]
	}
	return fmt.Sprintf("EventKind(%d)", int(k))
}

// Fighter is one side of a battle: the unit as it stood (its Size, Alt and
// Exp are the values before), its group size, alternate strength (the
// strength of the types that count it) and experience before and after,
// and whether the battle destroyed it.
type Fighter struct {
	Unit                UnitView
	Before, After       int
	ExpBefore, ExpAfter int
	Destroyed           bool
	AltBefore, AltAfter int
}

// BattleInfo is a fought attack order: the fighters, whether the defender
// fired back (it stood next to the attacker) and the support terms.
type BattleInfo struct {
	Attacker, Defender Fighter
	ReturnFire         bool
	Pincer, Blocking   int
}

// Event is one thing that happened, in order. Path runs from the start hex
// to the goal; Raw is the rules engine's event.
type Event struct {
	Kind     EventKind
	Side     int
	Unit     int
	Path     []hexgrid.Coord
	From, To hexgrid.Coord
	Battle   *BattleInfo
	Phase    [2]int
	Turn     int
	Energy   int
	End      *dosrules.End
	Hold     int // Captured: the building's container record
	Text     string
	Raw      dosrules.Event
}

// The status messages of the rules engine that events carry, beyond the
// dosrules Msg* constants.
const (
	msgEnemyThere   = 0x37
	msgArrived      = 0x38
	msgFriendThere  = 0x39
	msgNoCarry      = 0x3A
	msgFull         = 0x3B
	msgTerrain      = 0x40
	msgNoRoute      = 0x48
	msgMovementTurn = 0x35
	msgActionTurn   = 0x36
)

// MessageText is our wording of a status message code of the rules engine.
func MessageText(code int) string {
	switch code {
	case dosrules.MsgReadyUndone:
		return "Orders open again."
	case msgMovementTurn:
		return "Movement phase."
	case msgActionTurn:
		return "Action phase."
	case msgEnemyThere:
		return "An enemy unit holds that hex."
	case msgArrived:
		return "Move complete."
	case msgFriendThere:
		return "A friendly unit holds that hex."
	case msgNoCarry:
		return "That transport cannot take this unit now."
	case msgFull:
		return "There is no room left there."
	case msgTerrain:
		return "The unit cannot stop on that terrain."
	case dosrules.MsgNoTarget:
		return "No target there."
	case msgNoRoute:
		return "No route leads there."
	case dosrules.MsgAttack:
		return "Attack ordered."
	case dosrules.MsgLimit:
		return "No orders left this phase."
	case 0:
		return ""
	}
	return fmt.Sprintf("Message %#x.", code)
}

// text is the wording of the first message an event shows its side.
func text(r dosrules.Event) string {
	for _, m := range r.Messages {
		if m.Side == r.Side {
			return MessageText(m.Code)
		}
	}
	return ""
}

// battles is the state Resolve converts battles with: the units as the
// action side's map showed them before, and the running group size,
// alternate strength and experience.
type battles struct {
	side  int
	views map[int]UnitView
	state map[int][3]int
}

// fighter is a unit's part in an outcome; it updates the running state.
func (b *battles) fighter(idx, size, alt, exp int, gone []int) Fighter {
	st := b.state[idx]
	v := b.views[idx]
	v.Size, v.Alt, v.Exp = st[0], st[1], st[2]
	f := Fighter{Unit: v, Before: st[0], After: size, AltBefore: st[1], AltAfter: alt, ExpBefore: st[2], ExpAfter: exp,
		Destroyed: slices.Contains(gone, idx)}
	b.state[idx] = [3]int{size, alt, exp}
	return f
}

// event turns an engine event; b is set during a resolution.
func (s *Session) event(r dosrules.Event, b *battles) Event {
	ev := Event{Side: r.Side, Unit: r.Unit, Phase: r.Phase, Turn: r.Turn, Energy: r.Energy, End: r.End, Raw: r}
	switch r.Kind {
	case dosrules.EvAttack:
		ev.Kind, ev.To, ev.Text = Ordered, s.coord(r.Cell), text(r)
		ev.From = s.coord(s.Game.Units[r.Unit].Cell(r.Side))
	case dosrules.EvReady:
		ev.Kind, ev.Text = Confirmed, "Orders confirmed."
	case dosrules.EvUnready:
		ev.Kind, ev.Text = Unconfirmed, text(r)
	case dosrules.EvBattle:
		o := r.Outcome
		ev.Kind = Battle
		ev.Battle = &BattleInfo{ReturnFire: o.ReturnFire, Pincer: o.Pincer, Blocking: o.Blocking}
		if b != nil {
			ev.Battle.Attacker = b.fighter(r.Unit, o.AttackerSize, o.AttackerAlt, o.AttackerExp, o.Destroyed)
			ev.Battle.Defender = b.fighter(o.Defender, o.DefenderSize, o.DefenderAlt, o.DefenderExp, o.Destroyed)
			ev.From, ev.To = ev.Battle.Attacker.Unit.At, ev.Battle.Defender.Unit.At
		}
	case dosrules.EvSkipped:
		ev.Kind, ev.To = Skipped, s.coord(r.Cell)
		if b != nil {
			ev.From = b.views[r.Unit].At
		}
	case dosrules.EvPhase:
		ev.Kind, ev.Text = PhaseChanged, fmt.Sprintf("Turn %d.", r.Turn)
	case dosrules.EvCredit:
		ev.Kind = Credited
	case dosrules.EvRepaired:
		ev.Kind, ev.To = Repaired, s.coord(r.Cell)
	case dosrules.EvBuilt:
		ev.Kind, ev.To = Built, s.coord(r.Cell)
	case dosrules.EvEnd:
		ev.Kind = Ended
	case dosrules.EvCaptured:
		ev.Kind, ev.To, ev.Hold = Captured, s.coord(r.Cell), r.Hold
		ev.Text = captureText(kindOf(r.Hold), len(s.slots(r.Hold, r.Side))-1)
	case dosrules.EvMoved, dosrules.EvRefused:
		return s.moved(r, -1)
	}
	return ev
}

// captureText is the wording of a capture of a building of kind k in which
// found units changed sides.
func captureText(k BuildingKind, found int) string {
	t := [...]string{"HQ captured", "Depot captured", "Factory captured"}[k]
	switch {
	case found == 1:
		return t + ", 1 unit found."
	case found > 1:
		return fmt.Sprintf("%s, %d units found.", t, found)
	}
	return t + "."
}

// moved turns a move or an exit that started on cell start. The rules
// engine reports a unit's cell after each step; for an arrival in a
// building the last one is the building's position word, so the path ends
// on the ordered cell instead.
func (s *Session) moved(r dosrules.Event, start int) Event {
	ev := Event{Kind: Moved, Side: r.Side, Unit: r.Unit, Phase: r.Phase, Turn: r.Turn, Raw: r, To: s.coord(r.Cell)}
	if r.Kind == dosrules.EvRefused || len(r.Route) == 0 {
		ev.Kind, ev.Text = Refused, text(r)
		if ev.Text == "" {
			ev.Text = "The unit cannot go there."
		}
		if start >= 0 {
			ev.From = s.coord(start)
		}
		return ev
	}
	ev.Path = make([]hexgrid.Coord, 0, len(r.Route)+1)
	if start >= 0 {
		ev.Path = append(ev.Path, s.coord(start))
	}
	for _, c := range r.Route[:len(r.Route)-1] {
		ev.Path = append(ev.Path, s.coord(c))
	}
	ev.Path = append(ev.Path, ev.To)
	ev.From = ev.Path[0]
	return ev
}
