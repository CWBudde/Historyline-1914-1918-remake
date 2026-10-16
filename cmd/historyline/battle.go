package main

import "fmt"

// reportFrames is how long the battle panel stays up (about 3 s at 60 TPS);
// any click or key dismisses it earlier.
const reportFrames = 180

// reportFramesReduced is the panel time with reduced motion.
const reportFramesReduced = 60

// side is one fighter in the battle panel, with its size before and after.
type side struct {
	Name          string
	Weapon        string // weapon fired, empty when it did not fire
	Rank          int
	Before, After int
	Max           int
	Hidden        bool // the player cannot see this side: show nothing but the name
}

// report is what the battle panel shows; it is built from the engine's Battle
// result and the units' state before the shot, never from its own rolls.
type report struct {
	Attacker, Defender side
	frames             int
}

// Active reports whether the panel is still shown.
func (r *report) Active() bool { return r != nil && r.frames > 0 }

// Dismiss hides the panel at once (a map click or key press).
func (r *report) Dismiss() {
	if r != nil {
		r.frames = 0
	}
}

// Tick counts the display time down.
func (r *report) Tick() {
	if r.Active() {
		r.frames--
	}
}

func bar(s side, width int) string {
	if s.Max <= 0 {
		return ""
	}
	filled := s.After * width / s.Max
	out := make([]byte, width)
	for i := range out {
		out[i] = '-'
		if i < filled {
			out[i] = '#'
		}
	}
	return string(out)
}

// Lines renders the panel text, one row per line.
func (r *report) Lines() []string {
	row := func(role string, s side) string {
		if s.Hidden {
			return fmt.Sprintf("%s %s\n", role, s.Name)
		}
		w := s.Weapon
		if w == "" {
			w = "no reply"
		}
		return fmt.Sprintf("%s %s rank %d  %s\n  [%s] %d -> %d / %d", role, s.Name, s.Rank, w, bar(s, 16), s.Before, s.After, s.Max)
	}
	return []string{row("ATK", r.Attacker), row("DEF", r.Defender)}
}
