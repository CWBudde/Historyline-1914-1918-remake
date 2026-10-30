// Package llmplay lets a language model play a shipped level under the
// original (DOS) rules against the computer side (internal/dosai):
//
//   - Observe renders the fair, per-side state as canonical text (never
//     TrueView), with legal orders and exact Engage predictions annotated.
//   - ParseOrders/Apply take the model's JSONL orders, dry-run them on a
//     Clone and commit only a fully legal batch.
//   - Game drives the resolution loop, records a replayable transcript
//     plus a save per resolution, and never stalls: after repeated bad
//     replies a deterministic fallback (dosai.Planner + repair/build
//     service) plays the phase.
//   - Frame recording (frames/GIF) and post-game Review (blunder
//     detection feeding a LESSONS file) live in frame.go / review.go.
//
// Coordinates in every prompt and order are 1-based [line,col], matching
// the scenario files; hexgrid.Coord is 0-based.
package llmplay

import (
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Side names as the prompt prints them.
var Sides = [3]string{"CP", "EC", "NONE"}

// ShipLevel loads a shipped scenario setup, e.g. ("1918-02", scenario.CP)
// is RHEIN (the code word in levels/1918-02/cp.json).
func ShipLevel(month string, mode scenario.Mode) (*level.Level, error) {
	return scenario.ShippedLevel(month, mode)
}

// NewSession starts a DOS-rules session on lvl with side llmSide human and
// the other side computer-controlled.
func NewSession(lvl *level.Level, llmSide int) (*dosplay.Session, error) {
	set, err := data.Load()
	if err != nil {
		return nil, err
	}
	var comp [2]bool
	comp[1-llmSide] = true
	return dosplay.New(lvl, set, dosplay.Options{Computer: comp})
}

// LoadSession resumes a session from Session.Save bytes on its level.
func LoadSession(lvl *level.Level, saved []byte) (*dosplay.Session, error) {
	set, err := data.Load()
	if err != nil {
		return nil, err
	}
	return dosplay.Load(lvl, set, saved)
}

// ParseLevelRef parses "1918-02:cp" (mode cp|ec|2p, default cp).
func ParseLevelRef(ref string) (month string, mode scenario.Mode, err error) {
	month, mode = ref, scenario.CP
	for i := 0; i < len(ref); i++ {
		if ref[i] == ':' {
			month, mode = ref[:i], scenario.Mode(ref[i+1:])
			break
		}
	}
	switch mode {
	case scenario.CP, scenario.EC, scenario.TwoPlayer:
	default:
		return "", "", fmt.Errorf("llmplay: mode %q, want cp, ec or 2p", mode)
	}
	return month, mode, nil
}

// LLMSide is the side the model plays in a mode: CP (0) in cp, EC (1) in ec.
func LLMSide(mode scenario.Mode) int {
	if mode == scenario.EC {
		return 1
	}
	return 0
}

func show(c hexgrid.Coord) string { return fmt.Sprintf("[%d,%d]", c.Line+1, c.Col+1) }

func parseCoord(at [2]int) hexgrid.Coord { return hexgrid.Coord{Line: at[0] - 1, Col: at[1] - 1} }

func phaseName(s *dosplay.Session, side int) string {
	if s.Phase(side) == dosrules.PhaseMovement {
		return "movement"
	}
	return "action"
}
