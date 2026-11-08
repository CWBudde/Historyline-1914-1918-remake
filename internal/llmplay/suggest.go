package llmplay

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// SuggestAct is the coach's action phase for side as JSONL orders, played
// on a Clone: repairs and builds by its production doctrine, sapper work
// and the focus-fire attacks the exact battle outcome favours. It reads
// no opponent plan (the coach's movement oracle is not used), so a player
// may take, edit or ignore it.
func SuggestAct(s *dosplay.Session, side int) (string, error) {
	if s.Phase(side) == dosrules.PhaseMovement {
		return "", fmt.Errorf("suggest covers the action phase; this is a movement phase")
	}
	clone, err := s.Clone()
	if err != nil {
		return "", err
	}
	mark := clone.Mark()
	if _, err := newCoachPlan(clone, side, nil, nil).act(); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, c := range clone.CommandsSince(mark) {
		// Order's fixed arrays do not omit when empty; print only what the op names.
		o := struct {
			Op     string  `json:"op"`
			Unit   int     `json:"unit,omitempty"`
			Hold   int     `json:"hold,omitempty"`
			Slot   int     `json:"slot,omitempty"`
			To     *[2]int `json:"to,omitempty"`
			Target *[2]int `json:"target,omitempty"`
			Type   int     `json:"type,omitempty"`
		}{Op: string(c.Kind), Unit: c.Unit, Hold: c.Hold, Slot: c.Slot, Type: c.Type}
		at, ok := clone.Board.Coord(c.Cell)
		pos := [2]int{at.Line + 1, at.Col + 1}
		switch c.Kind {
		case dosrules.CmdMove, dosrules.CmdExit:
			if ok {
				o.To = &pos
			}
		case dosrules.CmdAttack:
			if ok {
				o.Target = &pos
			}
		case dosrules.CmdRepair, dosrules.CmdBuild:
		default:
			continue
		}
		line, err := json.Marshal(o)
		if err != nil {
			return "", err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return b.String(), nil
}
