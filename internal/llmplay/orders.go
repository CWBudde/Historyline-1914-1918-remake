package llmplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Order is one JSONL line from the model. Coordinates are 1-based
// [line,col]. Type is the UNIT.DAT type for build (as Buildable prints).
type Order struct {
	Op     string `json:"op"`
	Unit   int    `json:"unit,omitempty"`
	Hold   int    `json:"hold,omitempty"`
	Slot   int    `json:"slot,omitempty"`
	To     [2]int `json:"to,omitempty"`
	Target [2]int `json:"target,omitempty"`
	Type   int    `json:"type,omitempty"`
}

// ParseOrders reads a JSONL batch. Blank lines and lines without '{'
// (commentary the model wrapped around its orders) are skipped; malformed
// JSON lines are errors naming the line.
func ParseOrders(text string) ([]Order, error) {
	var out []Order
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var o Order
		dec := json.NewDecoder(bytes.NewReader([]byte(line)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&o); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no orders in reply")
	}
	return out, nil
}

func applyOne(s *dosplay.Session, side int, o Order) ([]dosplay.Event, error) {
	switch o.Op {
	case "move":
		return s.Move(side, o.Unit, parseCoord(o.To))
	case "exit":
		return s.Exit(side, o.Hold, o.Slot, parseCoord(o.To))
	case "attack":
		return s.Attack(side, o.Unit, parseCoord(o.Target))
	case "repair":
		return s.Repair(side, o.Hold, o.Slot)
	case "build":
		return s.Build(side, o.Hold, o.Type)
	case "confirm", "done", "end":
		return nil, nil // the loop confirms after a clean batch
	case "surrender":
		return s.Surrender(side), nil
	default:
		return nil, fmt.Errorf("unknown op %q", o.Op)
	}
}

// Apply validates the whole batch on a Clone (each order through the
// session's own Move/Exit/Attack/Repair/Build, so refusals carry the
// rules engine's reason) and commits it to s via Replay only if every
// order was taken. On failure nothing is applied and errs names every
// failing order. It does not Confirm; Game.Step does.
func Apply(s *dosplay.Session, side int, orders []Order) (cmds []dosrules.Command, errs []string, err error) {
	clone, err := s.Clone()
	if err != nil {
		return nil, nil, err
	}
	mark := clone.Mark()
	for i, o := range orders {
		evs, e := applyOne(clone, side, o)
		if e != nil {
			errs = append(errs, fmt.Sprintf("line %d (%s unit %d hold %d): %v", i+1, o.Op, o.Unit, o.Hold, e))
			continue
		}
		for _, ev := range evs {
			if ev.Kind == dosplay.Refused {
				errs = append(errs, fmt.Sprintf("line %d (%s unit %d): refused: %s", i+1, o.Op, o.Unit, ev.Text))
			}
		}
	}
	if len(errs) > 0 {
		return nil, errs, nil
	}
	cmds = clone.CommandsSince(mark)
	if _, err := s.Replay(cmds); err != nil {
		return nil, nil, fmt.Errorf("llmplay: committing validated batch: %w", err)
	}
	return cmds, nil, nil
}
