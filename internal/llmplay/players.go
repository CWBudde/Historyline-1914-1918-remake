package llmplay

import (
	"fmt"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Player is a registered programmatic player: `run -driver NAME`, the
// corpus sweep and regeneration play it by name. Players must be
// deterministic, so a recording made by one can be made again.
type Player struct {
	Name string
	Doc  string
	// Levels the player plays ("1918-02:cp"); nil for every level.
	Levels []string
	New    func(ref string) PhasePlayer
}

// Players are the registered players.
var Players = []Player{
	{Name: "coach", Doc: "the general self-player", New: func(ref string) PhasePlayer { return &Coach{Config: CoachConfigFor(ref)} }},
	{Name: "lookahead", Doc: "the coach choosing its posture by playing each ahead against dosai (an oracle)", New: func(ref string) PhasePlayer { return &Lookahead{Config: CoachConfigFor(ref)} }},
	{Name: "lookahead-deep", Doc: "lookahead playing 16 resolutions ahead", New: func(ref string) PhasePlayer { return &Lookahead{Config: CoachConfigFor(ref), Horizon: 16} }},
	{Name: "lookahead-wide", Doc: "lookahead also weighing the grab postures", New: func(ref string) PhasePlayer { return &Lookahead{Config: CoachConfigFor(ref), Wide: true} }},
	{Name: "fallback", Doc: "the computer side's own planner", New: func(string) PhasePlayer { return &FallbackPlayer{} }},
	{Name: "strike", Doc: "RHEIN's scripted HQ strike", Levels: []string{"1918-02:cp"}, New: func(string) PhasePlayer { return &Strike{} }},
}

// FindPlayer is the registered player name, if it plays ref.
func FindPlayer(name, ref string) (Player, error) {
	for _, p := range Players {
		if p.Name != name {
			continue
		}
		if p.Levels != nil && !slices.Contains(p.Levels, ref) {
			return p, fmt.Errorf("player %s plays only %s, not %s", name, strings.Join(p.Levels, ", "), ref)
		}
		return p, nil
	}
	return Player{}, fmt.Errorf("no player %q (have %s)", name, strings.Join(PlayerNames(), ", "))
}

// PlayerNames are the registered players' names.
func PlayerNames() []string {
	names := make([]string, len(Players))
	for i, p := range Players {
		names[i] = p.Name
	}
	return names
}

// FallbackPlayer plays a side with the computer side's own planner.
type FallbackPlayer struct{ p dosai.Planner }

// PlayPhase implements PhasePlayer.
func (f *FallbackPlayer) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	return Fallback(s, side, &f.p)
}

// SetupCode is the code word of a shipped setup ("RHEIN" for 1918-02, cp).
func SetupCode(month string, mode scenario.Mode) (string, error) {
	f, err := scenario.Read(levels.Levels, month)
	if err != nil {
		return "", err
	}
	if st := f.Setup(mode); st != nil {
		return st.Code, nil
	}
	return "", fmt.Errorf("llmplay: %s has no %s setup", month, mode)
}

// StartGame starts a recorded game on the shipped setup ref ("1918-02:cp")
// in dir, the side the mode gives (LLMSide) played by driver.
func StartGame(ref, dir, driver string, oracle bool) (*Game, error) {
	month, mode, err := ParseLevelRef(ref)
	if err != nil {
		return nil, err
	}
	lvl, err := ShipLevel(month, mode)
	if err != nil {
		return nil, err
	}
	code, err := SetupCode(month, mode)
	if err != nil {
		return nil, err
	}
	side := LLMSide(mode)
	s, err := NewSession(lvl, side)
	if err != nil {
		return nil, err
	}
	return NewGame(s, dir, Header{Level: ref, Code: code, LLMSide: side, Driver: driver, Oracle: oracle})
}
