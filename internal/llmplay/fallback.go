package llmplay

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
)

// Fallback plays side's whole phase deterministically so a game never
// stalls on a bad model reply: dosai.Planner gives the phase's orders,
// repairs and builds, and confirms.
func Fallback(s *dosplay.Session, side int, p *dosai.Planner) ([]dosplay.Event, error) {
	return p.Plan(s, side)
}
