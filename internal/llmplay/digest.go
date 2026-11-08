package llmplay

import (
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// EventRec is the transcript's serialisable view of a dosplay.Event.
type EventRec struct {
	Kind   string `json:"kind"`
	Side   int    `json:"side"`
	Unit   int    `json:"unit,omitempty"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Path   string `json:"path,omitempty"`
	Text   string `json:"text,omitempty"`
	Battle string `json:"battle,omitempty"`
	Energy int    `json:"energy,omitempty"`
	End    string `json:"end,omitempty"`
}

func pathStr(p []hexgrid.Coord) string {
	parts := make([]string, len(p))
	for i, c := range p {
		parts[i] = show(c)
	}
	return strings.Join(parts, "->")
}

func battleStr(b *dosplay.BattleInfo) string {
	if b == nil {
		return ""
	}
	a, d := b.Attacker, b.Defender
	return fmt.Sprintf("atk u%d %s %d->%d%s vs def u%d %s %d->%d%s returnFire=%v pincer=%d blocking=%d",
		a.Unit.Index, a.Unit.ID, a.Before, a.After, dead(a.Destroyed),
		d.Unit.Index, d.Unit.ID, d.Before, d.After, dead(d.Destroyed),
		b.ReturnFire, b.Pincer, b.Blocking)
}

func dead(d bool) string {
	if d {
		return " DESTROYED"
	}
	return ""
}

// Record converts an event for the transcript and the prompt digest.
func Record(ev dosplay.Event) EventRec {
	r := EventRec{
		Kind: ev.Kind.String(), Side: ev.Side, Unit: ev.Unit,
		From: show(ev.From), To: show(ev.To), Text: ev.Text, Energy: ev.Energy,
	}
	if len(ev.Path) > 0 {
		r.Path = pathStr(ev.Path)
	}
	if ev.Battle != nil {
		r.Battle = battleStr(ev.Battle)
	}
	if ev.End != nil {
		r.End = fmt.Sprintf("over=%v winner=%d won=%v score=%d", ev.End.Over, ev.End.Winner, ev.End.Won, ev.End.Score)
	}
	return r
}

// Digest renders events for the next prompt: what the side may learn —
// its battles with before/after strength, captures, revealed enemy
// moves (the only fair enemy-move channel), income and the end.
func Digest(evs []dosplay.Event) string {
	var b strings.Builder
	for _, ev := range evs {
		switch ev.Kind {
		case dosplay.Battle:
			fmt.Fprintf(&b, " battle %s at %s\n", battleStr(ev.Battle), show(ev.To))
		case dosplay.Skipped:
			fmt.Fprintf(&b, " attack on %s found no unit there\n", show(ev.To))
		case dosplay.Revealed:
			fmt.Fprintf(&b, " move revealed (movement side, shown after phase change): u%d %s\n", ev.Unit, pathStr(ev.Path))
		case dosplay.Captured:
			fmt.Fprintf(&b, " %s captured building at %s (hold %d): %s\n", Sides[ev.Side], show(ev.To), ev.Hold, ev.Text)
		case dosplay.Credited:
			fmt.Fprintf(&b, " %s credited, energy now %d\n", Sides[ev.Side], ev.Energy)
		case dosplay.Repaired:
			fmt.Fprintf(&b, " repaired u%d at %s, energy left %d\n", ev.Unit, show(ev.To), ev.Energy)
		case dosplay.Built:
			fmt.Fprintf(&b, " built u%d at %s, energy left %d\n", ev.Unit, show(ev.To), ev.Energy)
		case dosplay.Dug:
			fmt.Fprintf(&b, " sapper work at %s: %s\n", show(ev.To), ev.Text)
		case dosplay.Constructed:
			fmt.Fprintf(&b, " %s depot work at %s: %s\n", Sides[ev.Side], show(ev.To), ev.Text)
		case dosplay.Ended:
			if ev.End != nil {
				fmt.Fprintf(&b, " GAME OVER winner=side%d (%s) won=%v score=%d\n", ev.End.Winner, Sides[max(ev.End.Winner, 0)], ev.End.Won, ev.End.Score)
			}
		case dosplay.PhaseChanged:
			fmt.Fprintf(&b, " phase changed, turn %d\n", ev.Turn)
		}
	}
	return b.String()
}

// BriefDigest is Digest for side in fewer words: units as uN (side's) and
// eN (the enemy's) without the side prefix, battles without the zero
// support terms, enemy moves as start->goal, and side's own moves left
// out (YOUR UNITS shows where they ended).
func BriefDigest(evs []dosplay.Event, side int) string {
	name := func(u dosplay.UnitView) string {
		p := "u"
		if u.Side != side {
			p = "e"
		}
		id := u.ID
		if _, rest, ok := strings.Cut(id, "_"); ok {
			id = rest
		}
		return fmt.Sprintf("%s%d %s", p, u.Index, id)
	}
	fighter := func(f dosplay.Fighter) string {
		return fmt.Sprintf("%s %d->%d%s", name(f.Unit), f.Before, f.After, dead(f.Destroyed))
	}
	var b strings.Builder
	for _, ev := range evs {
		switch ev.Kind {
		case dosplay.Battle:
			if bi := ev.Battle; bi != nil {
				fmt.Fprintf(&b, " %s: %s hits %s", show(ev.To), fighter(bi.Attacker), fighter(bi.Defender))
				if bi.Pincer != 0 {
					fmt.Fprintf(&b, " pincer=%d", bi.Pincer)
				}
				b.WriteByte('\n')
			}
		case dosplay.Revealed:
			if ev.Side != side && len(ev.Path) > 0 {
				fmt.Fprintf(&b, " e%d moved %s->%s\n", ev.Unit, show(ev.Path[0]), show(ev.Path[len(ev.Path)-1]))
			}
		case dosplay.Credited:
			if ev.Side == side {
				fmt.Fprintf(&b, " energy %d\n", ev.Energy)
			}
		case dosplay.Repaired, dosplay.Built:
			verb := map[dosplay.EventKind]string{dosplay.Repaired: "repaired", dosplay.Built: "built"}[ev.Kind]
			if ev.Side == side {
				fmt.Fprintf(&b, " %s u%d at %s, energy left %d\n", verb, ev.Unit, show(ev.To), ev.Energy)
			} else {
				fmt.Fprintf(&b, " enemy %s e%d at %s\n", verb, ev.Unit, show(ev.To))
			}
		case dosplay.Skipped, dosplay.Captured, dosplay.Dug, dosplay.Constructed, dosplay.Ended, dosplay.PhaseChanged:
			b.WriteString(Digest([]dosplay.Event{ev}))
		}
	}
	return b.String()
}
