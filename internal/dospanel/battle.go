package dospanel

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// MaxFigures is the most men a group shows: one figure per place on the
// hex's ring.
const MaxFigures = 6

// Fighter is one side of a battle scene: the unit's name, its photo and map
// sprite (nil: a blank plate and a drawn figure), its nation's flag (nil: a
// plain bar in FlagColour, none when that is transparent), its experience
// badge (nil, or a nil image: none), the group size before and after the
// battle with the full group, the experience before and after, whether the
// group was destroyed and whether it is a ship, shown as one large picture
// instead of a figure per man.
type Fighter struct {
	Name          string
	Photo, Sprite *ebiten.Image
	Flag          *ebiten.Image
	FlagColour    color.RGBA
	Badge         func(exp int) *ebiten.Image

	Before, After, Full int
	ExpBefore, ExpAfter int
	Destroyed, Ship     bool
}

// Battle is a battle scene: the attacker in the panel's top half, the
// defender in the bottom half. ReturnFire marks a defender that fired back:
// it shows a volley even when the attacker lost nobody.
type Battle struct {
	Attacker, Defender Fighter
	ReturnFire         bool
}

// BattleArt is optional art for the scene: the frames of an explosion,
// drawn toned to sepia. Without frames the explosion is drawn.
type BattleArt struct{ Explosion []*ebiten.Image }

// FlashKind is the kind of an effect in a battle scene.
type FlashKind int

// The effects: a muzzle flash at a shooter, the blast that takes a man and
// the dust of a shot that missed.
const (
	Muzzle FlashKind = iota
	Blast
	Miss
)

// Flash is an effect running in a battle scene on figure Figure of side
// Side: the shooter for Muzzle, the target for Blast and Miss. Progress
// runs from 0 to 1 over the effect.
type Flash struct {
	Kind         FlashKind
	Side, Figure int
	Progress     float64
}

// SideState is one side of a battle scene at a tick: the figures standing
// (0 to Before, in ring order) and how many of them are lost, the group size
// displayed, the newest figure's arrival (0 just placed, 1 settled) and the
// flash of the count after a loss (1 just hit, fading to 0).
type SideState struct {
	Shown, Lost int
	Count       int
	Rise, Hit   float64
}

// BattleState is a battle scene at a tick: the sides (0 attacker, 1
// defender), the effects running, whether each side's badge shows the new
// experience and the "+1" glow of a gain (1 at the gain, fading to 0).
type BattleState struct {
	Sides  [2]SideState
	Flash  []Flash
	Gained [2]bool
	Glow   [2]float64
}

// The scene's timing in ticks (60 a second).
const (
	buildStep   = 6  // between two figures of the build-up
	buildPause  = 12 // after the build-up
	shotStep    = 28 // from one shot to the next
	muzzleTicks = 10 // a muzzle flash
	blastFrom   = 8  // the blast, from the shot's start
	blastTicks  = 20 // a blast
	landAt      = 14 // the man falls, from the shot's start
	volleyGap   = 16 // between the attacker's fire and the return fire
	holdTicks   = 70 // the end held
	glowTicks   = 40 // the "+1" of a gain
	hitTicks    = 16 // the count's flash after a loss
	reducedHold = 20 // the whole scene under reduced motion
)

// Shooters by preference: the attacker fires down from the ring's lower
// places, the defender up from its upper ones.
var shooterOrder = [2][MaxFigures]int{{3, 2, 4, 1, 5, 0}, {0, 1, 5, 2, 4, 3}}

type build struct{ at, side, shown int }

type shot struct {
	at, side, fig, shooter int
	miss                   bool
}

// schedule is a battle laid out in time.
type schedule struct {
	before, after [2]int
	exp           [2][2]int
	ship          [2]bool
	builds        []build
	shots         []shot
	holdAt, total int
}

// sides are a battle's fighters in side order.
func (b Battle) sides() [2]Fighter { return [2]Fighter{b.Attacker, b.Defender} }

func plan(b Battle) schedule {
	var s schedule
	for i, f := range b.sides() {
		s.before[i] = min(max(f.Before, 0), MaxFigures)
		s.after[i] = min(max(f.After, 0), s.before[i])
		if f.Destroyed {
			s.after[i] = 0
		}
		s.exp[i] = [2]int{f.ExpBefore, f.ExpAfter}
		s.ship[i] = f.Ship
	}

	t := 0
	for i := range MaxFigures {
		for side := range 2 {
			if i >= s.before[side] || s.ship[side] && i > 0 {
				continue
			}
			n := i + 1
			if s.ship[side] {
				n = s.before[side]
			}
			s.builds = append(s.builds, build{t, side, n})
			t += buildStep
		}
	}
	t += buildPause

	// shooter is the figure of side that fires the k-th shot while its
	// first standing figures live; -1 when none does.
	shooter := func(side, standing, k int) int {
		var living []int
		for _, f := range shooterOrder[side] {
			if f < standing {
				living = append(living, f)
			}
		}
		if len(living) == 0 {
			return -1
		}
		return living[k%len(living)]
	}
	// fire takes target's losses, shot by the standing figures of the
	// other side.
	fire := func(target, standing int) {
		for k := range s.before[target] - s.after[target] {
			s.shots = append(s.shots, shot{t, target, s.before[target] - 1 - k, shooter(1-target, standing, k), false})
			t += shotStep
		}
	}
	fire(1, s.before[0]) // the attacker has lost nobody yet
	if len(s.shots) > 0 {
		t += volleyGap
	}
	if s.after[0] < s.before[0] {
		fire(0, s.after[1])
	} else if b.ReturnFire && s.after[1] > 0 && s.before[0] > 0 {
		s.shots = append(s.shots, shot{t, 0, s.before[0] - 1, shooter(1, s.after[1], 0), true})
		t += shotStep
	}
	s.holdAt = t
	s.total = t + holdTicks
	return s
}

// BattleTicks is the length of battle b's scene in ticks: the build-up, the
// losses and the hold; a short hold of the end alone under reduced motion.
func BattleTicks(b Battle, reduced bool) int {
	if reduced {
		return reducedHold
	}
	return plan(b).total
}

// FirstLoss is the tick the first man falls in battle b's scene, for the
// map to follow; with no loss the start of the hold. Under reduced motion
// the losses show at once.
func FirstLoss(b Battle, reduced bool) int {
	s := plan(b)
	for _, sh := range s.shots {
		if !sh.miss {
			if reduced {
				return 0
			}
			return sh.at + landAt
		}
	}
	if reduced {
		return reducedHold
	}
	return s.holdAt
}

// ramp01 is v from a to a+n as 0 to 1, clamped.
func ramp01(v, a, n int) float64 {
	return math.Max(0, math.Min(1, float64(v-a)/float64(n)))
}

// BattleAt is battle b's scene at tick: under reduced motion its end at
// every tick.
func BattleAt(b Battle, tick int, reduced bool) BattleState {
	s := plan(b)
	var st BattleState
	if reduced {
		for i := range 2 {
			st.Sides[i] = SideState{Shown: s.before[i], Lost: s.before[i] - s.after[i], Count: s.after[i], Rise: 1}
			st.Gained[i] = s.exp[i][1] > s.exp[i][0]
		}
		return st
	}
	tick = min(max(tick, 0), s.total)

	for i := range 2 {
		st.Sides[i].Rise = 1
	}
	for _, bd := range s.builds {
		if bd.at <= tick {
			st.Sides[bd.side].Shown = bd.shown
			st.Sides[bd.side].Rise = ramp01(tick, bd.at, buildStep)
		}
	}
	for _, sh := range s.shots {
		if !sh.miss && sh.at+landAt <= tick {
			ss := &st.Sides[sh.side]
			ss.Lost++
			ss.Hit = 1 - ramp01(tick, sh.at+landAt, hitTicks)
		}
		if sh.shooter >= 0 && tick >= sh.at && tick < sh.at+muzzleTicks {
			st.Flash = append(st.Flash, Flash{
				Kind: Muzzle, Side: 1 - sh.side, Figure: sh.shooter,
				Progress: ramp01(tick, sh.at, muzzleTicks),
			})
		}
		if tick >= sh.at+blastFrom && tick < sh.at+blastFrom+blastTicks {
			k := Blast
			if sh.miss {
				k = Miss
			}
			st.Flash = append(st.Flash, Flash{
				Kind: k, Side: sh.side, Figure: sh.fig,
				Progress: ramp01(tick, sh.at+blastFrom, blastTicks),
			})
		}
	}
	for i := range 2 {
		st.Sides[i].Count = s.before[i] - st.Sides[i].Lost
		if s.exp[i][1] > s.exp[i][0] && tick >= s.holdAt {
			st.Gained[i] = true
			st.Glow[i] = 1 - ramp01(tick, s.holdAt, glowTicks)
		}
	}
	return st
}
