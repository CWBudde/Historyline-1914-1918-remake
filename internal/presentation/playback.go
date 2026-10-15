// Package presentation plays immutable visual descriptions after game rules have
// resolved. It neither imports nor changes gameplay state, and uses caller time.
package presentation

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// Animation is the exported interface candidate's finite animation contract.
type Animation struct {
	ID            string   `json:"id"`
	Frames        []string `json:"frames"`
	FrameMS       []int    `json:"frame_ms"`
	ReducedFrames []string `json:"reduced_frames"`
	ReducedMS     []int    `json:"reduced_ms"`
	Anchor        [2]int   `json:"anchor"`
	Loop          bool     `json:"loop"`
	Skippable     bool     `json:"skippable"`
	Notes         string   `json:"notes"`
}

type Point struct{ X, Y float64 }

// Facing chooses one of the six independently authored sprites. Screen Y grows
// downward. No sprite is rotated or reflected to supply a missing view.
func Facing(from, to Point) (string, error) {
	dx, dy := to.X-from.X, to.Y-from.Y
	if !finite(dx) || !finite(dy) || (dx == 0 && dy == 0) {
		return "", errors.New("movement needs two finite, distinct points")
	}
	angle := math.Atan2(dx, -dy)
	index := int(math.Round(angle / (math.Pi / 3)))
	return []string{"n", "ne", "se", "s", "sw", "nw"}[(index+6)%6], nil
}

func finite(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) }

// Step is either a move (nonzero Duration) or an effect (Effect names Animation).
// Position and To are logical unit anchors, never top-left image coordinates.
type Step struct {
	ID       string
	Position Point
	To       Point
	Duration time.Duration
	Effect   string
}

// State is a value snapshot. Frame is empty during movement or after completion.
type State struct {
	ID       string
	Index    int
	Position Point
	Facing   string
	Frame    string
	Anchor   [2]int
	Elapsed  time.Duration
	Done     bool
}

type segment struct {
	step      Step
	animation Animation
	duration  time.Duration
	facing    string
}

// Player owns only a copied visual schedule. Reduced moves snap immediately to
// their destination and reduced effects use the export's static alternatives.
type Player struct {
	segments []segment
	index    int
	elapsed  time.Duration
	position Point
	facing   string
	reduced  bool
}

func New(animations []Animation, steps []Step, reduced bool) (*Player, error) {
	defs := make(map[string]Animation, len(animations))
	for _, a := range animations {
		if a.ID == "" || a.Loop || len(a.Frames) == 0 || len(a.Frames) != len(a.FrameMS) || len(a.ReducedFrames) == 0 || len(a.ReducedFrames) != len(a.ReducedMS) {
			return nil, fmt.Errorf("invalid finite animation %q", a.ID)
		}
		if _, exists := defs[a.ID]; exists {
			return nil, fmt.Errorf("duplicate animation %q", a.ID)
		}
		for i, frames := range [][]string{a.Frames, a.ReducedFrames} {
			durations := a.FrameMS
			if i == 1 {
				durations = a.ReducedMS
			}
			for j, frame := range frames {
				if frame == "" || durations[j] <= 0 || durations[j] > 60000 {
					return nil, fmt.Errorf("invalid frame in %q", a.ID)
				}
			}
		}
		a.Frames = append([]string(nil), a.Frames...)
		a.FrameMS = append([]int(nil), a.FrameMS...)
		a.ReducedFrames = append([]string(nil), a.ReducedFrames...)
		a.ReducedMS = append([]int(nil), a.ReducedMS...)
		defs[a.ID] = a
	}
	p := &Player{reduced: reduced, facing: "n"}
	for _, step := range steps {
		if step.ID == "" || !finite(step.Position.X) || !finite(step.Position.Y) {
			return nil, errors.New("step needs an ID and finite position")
		}
		s := segment{step: step}
		if step.Effect != "" {
			if step.Duration != 0 {
				return nil, errors.New("effect cannot also be a move")
			}
			var ok bool
			s.animation, ok = defs[step.Effect]
			if !ok {
				return nil, fmt.Errorf("unknown animation %q", step.Effect)
			}
			ms := s.animation.FrameMS
			if reduced {
				ms = s.animation.ReducedMS
			}
			for _, n := range ms {
				s.duration += time.Duration(n) * time.Millisecond
			}
		} else {
			if step.Duration <= 0 || step.Duration > time.Minute {
				return nil, errors.New("move duration must be positive and at most one minute")
			}
			var err error
			s.facing, err = Facing(step.Position, step.To)
			if err != nil {
				return nil, err
			}
			if !reduced {
				s.duration = step.Duration
			}
		}
		p.segments = append(p.segments, s)
	}
	if len(steps) > 0 {
		p.position = steps[0].Position
	}
	p.normalize()
	return p, nil
}

func (p *Player) normalize() {
	for p.index < len(p.segments) {
		s := p.segments[p.index]
		if s.step.Effect == "" {
			p.facing = s.facing
		}
		if s.duration > p.elapsed {
			return
		}
		p.elapsed -= s.duration
		p.position = s.step.Position
		if s.step.Effect == "" {
			p.position = s.step.To
		}
		p.index++
	}
	p.elapsed = 0
}

// Advance consumes a nonnegative interval, including boundaries crossed within
// that interval. Its outcome does not depend on the caller's update frequency.
func (p *Player) Advance(delta time.Duration) error {
	if delta < 0 {
		return errors.New("cannot reverse presentation time")
	}
	// A bounded interval avoids duration overflow on hostile caller input.
	if delta > 24*time.Hour {
		return errors.New("advance exceeds 24 hours")
	}
	p.elapsed += delta
	p.normalize()
	return nil
}

// Skip finishes the active move/effect, retaining the move's destination anchor.
// A non-skippable effect returns false and continues until its finite expiry.
func (p *Player) Skip() bool {
	if p.index == len(p.segments) {
		return false
	}
	s := p.segments[p.index]
	if s.step.Effect != "" && !s.animation.Skippable {
		return false
	}
	p.elapsed = s.duration
	p.normalize()
	return true
}

func (p *Player) State() State {
	s := State{Index: p.index, Position: p.position, Facing: p.facing, Elapsed: p.elapsed, Done: p.index == len(p.segments)}
	if s.Done {
		return s
	}
	seg := p.segments[p.index]
	s.ID = seg.step.ID
	s.Position = seg.step.Position
	if seg.step.Effect == "" {
		fraction := float64(p.elapsed) / float64(seg.duration)
		s.Position.X += (seg.step.To.X - seg.step.Position.X) * fraction
		s.Position.Y += (seg.step.To.Y - seg.step.Position.Y) * fraction
		return s
	}
	s.Anchor = seg.animation.Anchor
	frames, durations := seg.animation.Frames, seg.animation.FrameMS
	if p.reduced {
		frames, durations = seg.animation.ReducedFrames, seg.animation.ReducedMS
	}
	elapsed := p.elapsed
	for i, frame := range frames {
		d := time.Duration(durations[i]) * time.Millisecond
		if elapsed < d {
			s.Frame = frame
			return s
		}
		elapsed -= d
	}
	return s
}
