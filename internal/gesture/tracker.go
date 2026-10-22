package gesture

import (
	"fmt"
	"math"
)

// Pointer is the mouse (or a touch) in one frame, in device pixels. Active is
// false when the panel takes no pointer input this frame; keep it true during
// a drag even if the pointer leaves the panel.
type Pointer struct {
	X, Y         float64
	Active       bool
	Down         bool // left button held
	RightPressed bool // right button went down this frame
}

// Keys is one key set in one frame: which of its keys are held.
type Keys struct {
	Up, Down, Left, Right, Fire bool
}

// Input is one frame of a side's input.
type Input struct {
	Pointer Pointer
	Keys    Keys
}

// Kind tells what an Event is.
type Kind int

// Event kinds.
const (
	Hover     Kind = iota // the idle pointer moved to X, Y
	Step                  // an arrow moved the cursor one cell towards Dir (see Neighbor)
	Press                 // fire went down; the cursor shows the cancel plate
	Pick                  // the picked direction changed while held: Dir, Action, Enabled
	Run                   // released on Action, which is possible
	Click                 // pressed and released without picking: select or confirm
	Cancelled             // released on cancel or a disabled action, or the ring closed
	RingOpen              // right-click at X, Y: the caller sets the ring with SetRing
)

var kindNames = [...]string{"hover", "step", "press", "pick", "run", "click", "cancelled", "ring open"}

func (k Kind) String() string {
	if k < 0 || int(k) >= len(kindNames) {
		return fmt.Sprintf("Kind(%d)", int(k))
	}
	return kindNames[k]
}

// Event is what a frame of input did. X, Y are the pointer position for
// Hover and RingOpen, the click position on the ring, and the press position
// for every other event of a mouse gesture; Mouse tells the pointer from the
// keys.
type Event struct {
	Kind    Kind
	Dir     Dir
	Action  Action
	Enabled bool
	X, Y    float64
	Mouse   bool
}

func (e Event) String() string {
	src := "keys"
	if e.Mouse {
		src = fmt.Sprintf("mouse %v,%v", e.X, e.Y)
	}
	return fmt.Sprintf("{%v %v %v %v %s}", e.Kind, e.Dir, e.Action, e.Enabled, src)
}

// Config sets the gesture threshold and the arrow repeat.
type Config struct {
	Threshold   float64 // drag length in device px that picks: one cell pitch on screen
	RepeatDelay int     // frames an arrow is held before it repeats; 0 never repeats
	RepeatEvery int     // frames between repeats
}

type mode int

const (
	modeIdle mode = iota
	modeMouse
	modeKeys
	modeRing
)

// Tracker recognizes gestures for one side. Set Placement to the cursor's
// context whenever the cursor moves (on Hover, Step and Press); Pick and Run
// read the action from it.
type Tracker struct {
	Config
	Placement Placement

	mode     mode
	dir      Dir     // the picked direction while held
	picked   bool    // a direction was picked during this hold
	ox, oy   float64 // where the mouse pressed
	ring     Ring
	ringSet  bool
	pointer  Pointer // last active pointer
	seen     bool    // pointer was active last frame
	keys     Keys    // keys last frame
	heldFor  [4]int  // frames each arrow has been held
	arrowWas [4]bool
}

// NewTracker returns an idle tracker.
func NewTracker(cfg Config) *Tracker { return &Tracker{Config: cfg} }

// Reset drops a hold or an open ring, as when the panel changes mode. Buttons
// still held stay ignored until they are released.
func (t *Tracker) Reset() {
	t.mode, t.dir, t.picked, t.ringSet = modeIdle, Centre, false, false
}

// Plate returns the action the cursor plate shows, whether it is possible,
// and whether fire is held (the plate replaces the ring only then).
func (t *Tracker) Plate() (a Action, enabled, held bool) {
	held = t.mode == modeMouse || t.mode == modeKeys
	if !held {
		return Cancel, true, false
	}
	a, enabled = t.Placement.At(t.dir)
	return a, enabled, true
}

// SetRing sets the ring the caller opened on RingOpen. Until it is set, a
// click closes the ring.
func (t *Tracker) SetRing(r Ring) {
	if t.mode == modeRing {
		t.ring, t.ringSet = r, true
	}
}

// Ring returns the open ring.
func (t *Tracker) Ring() (Ring, bool) { return t.ring, t.mode == modeRing && t.ringSet }

// RingHover returns the ring item under the pointer, or -1.
func (t *Tracker) RingHover() int {
	if r, ok := t.Ring(); ok && t.seen {
		return r.Hit(t.pointer.X, t.pointer.Y)
	}
	return -1
}

// Update feeds one frame and returns what it did. The keys go first, so a
// key release and a pointer move in one frame give the release and the
// hover.
func (t *Tracker) Update(in Input) []Event {
	var out []Event
	out = t.updateKeys(in.Keys, out)
	out = t.updatePointer(in.Pointer, out)
	return out
}

var arrowDirs = [4]Dir{Up, Left, Right, Down}

func arrows(k Keys) [4]bool { return [4]bool{k.Up, k.Left, k.Right, k.Down} }

func (t *Tracker) updateKeys(k Keys, out []Event) []Event {
	firePressed := k.Fire && !t.keys.Fire
	fireReleased := !k.Fire && t.keys.Fire
	now := arrows(k)
	var newly [4]bool
	for i := range now {
		newly[i] = now[i] && !t.arrowWas[i]
		if !now[i] {
			t.heldFor[i] = 0
		} else if !newly[i] {
			t.heldFor[i]++
		}
	}
	t.keys, t.arrowWas = k, now

	switch t.mode {
	case modeIdle:
		if firePressed {
			t.mode, t.dir, t.picked = modeKeys, Centre, false
			out = append(out, Event{Kind: Press})
			return t.keyPicks(newly, out)
		}
		for i, d := range arrowDirs {
			if newly[i] || t.repeats(t.heldFor[i]) && now[i] {
				out = append(out, Event{Kind: Step, Dir: d})
			}
		}
	case modeKeys:
		out = t.keyPicks(newly, out)
		if fireReleased {
			out = append(out, t.release(Event{}))
		}
	case modeRing:
		if firePressed {
			t.Reset()
			out = append(out, Event{Kind: Cancelled, Action: Cancel})
		}
	}
	return out
}

func (t *Tracker) repeats(n int) bool {
	if t.RepeatDelay <= 0 || n < t.RepeatDelay {
		return false
	}
	return t.RepeatEvery <= 0 || (n-t.RepeatDelay)%t.RepeatEvery == 0
}

// keyPicks moves the pick over the plus of directions: an arrow from the
// centre picks its direction, the opposite arrow returns to the centre and a
// perpendicular one jumps to its direction.
func (t *Tracker) keyPicks(newly [4]bool, out []Event) []Event {
	for i, d := range arrowDirs {
		if !newly[i] {
			continue
		}
		next := d
		if opposite(t.dir) == d {
			next = Centre
		}
		out = t.setDir(next, Event{}, out)
	}
	return out
}

// opposite returns the direction facing d; Centre faces nothing.
func opposite(d Dir) Dir {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	case Right:
		return Left
	}
	return -1
}

func (t *Tracker) setDir(d Dir, base Event, out []Event) []Event {
	if d == t.dir {
		return out
	}
	t.dir = d
	t.picked = t.picked || d != Centre
	e := base
	e.Kind, e.Dir = Pick, d
	e.Action, e.Enabled = t.Placement.At(d)
	return append(out, e)
}

// release ends a hold: run the picked action, cancel, or click when nothing
// was ever picked.
func (t *Tracker) release(base Event) Event {
	a, ok := t.Placement.At(t.dir)
	e := base
	switch {
	case t.dir == Centre && !t.picked:
		e.Kind = Click
	case t.dir == Centre || !ok || a == Cancel:
		e.Kind, e.Action = Cancelled, Cancel
	default:
		e.Kind, e.Action = Run, a
	}
	t.Reset()
	return e
}

func (t *Tracker) updatePointer(p Pointer, out []Event) []Event {
	if !p.Active {
		t.seen = false
		if t.mode == modeMouse { // the drag lost its pointer
			t.Reset()
			out = append(out, Event{Kind: Cancelled, Action: Cancel, X: t.ox, Y: t.oy, Mouse: true})
		}
		return out
	}
	if !t.seen { // first frame over the panel: only take the state
		t.pointer, t.seen = p, true
		return out
	}
	last := t.pointer
	t.pointer = p
	pressed := p.Down && !last.Down
	moved := p.X != last.X || p.Y != last.Y
	at := Event{X: p.X, Y: p.Y, Mouse: true}

	switch t.mode {
	case modeIdle:
		switch {
		case p.RightPressed:
			t.mode, t.ringSet = modeRing, false
			at.Kind = RingOpen
			out = append(out, at)
		case pressed:
			t.mode, t.dir, t.picked, t.ox, t.oy = modeMouse, Centre, false, p.X, p.Y
			at.Kind = Press
			out = append(out, at)
		case moved:
			at.Kind = Hover
			out = append(out, at)
		}
	case modeMouse:
		origin := Event{X: t.ox, Y: t.oy, Mouse: true}
		out = t.setDir(dragDir(p.X-t.ox, p.Y-t.oy, t.Threshold), origin, out)
		if !p.Down {
			out = append(out, t.release(origin))
		}
	case modeRing:
		if !p.RightPressed && !pressed {
			break
		}
		at.Kind, at.Action = Cancelled, Cancel
		if r, ok := t.Ring(); ok && pressed && !p.RightPressed {
			if i := r.Hit(p.X, p.Y); i >= 0 {
				at.Kind, at.Action = Run, r.Actions[i]
			}
		}
		t.Reset()
		out = append(out, at)
	}
	return out
}

// dragDir is the direction of a drag by (dx, dy): its dominant axis once it
// is at least threshold long, the vertical one on a tie; Centre while shorter.
func dragDir(dx, dy, threshold float64) Dir {
	if math.Hypot(dx, dy) < threshold || dx == 0 && dy == 0 {
		return Centre
	}
	if math.Abs(dy) >= math.Abs(dx) {
		if dy < 0 {
			return Up
		}
		return Down
	}
	if dx < 0 {
		return Left
	}
	return Right
}
