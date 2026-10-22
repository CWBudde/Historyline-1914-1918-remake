package gesture

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

const pitch = 40 // drag threshold in device px for the tests

// Frame builders: a mouse at (x, y), held or not; a right click; keys held.
func mouse(x, y float64, down bool) Input {
	return Input{Pointer: Pointer{X: x, Y: y, Active: true, Down: down}}
}

func rightClick(x, y float64) Input {
	return Input{Pointer: Pointer{X: x, Y: y, Active: true, RightPressed: true}}
}

func keys(k Keys) Input { return Input{Keys: k} }

var (
	idle  = Keys{}
	fire  = Keys{Fire: true}
	up    = Keys{Up: true}
	left  = Keys{Left: true}
	right = Keys{Right: true}
	down  = Keys{Down: true}
)

func with(a, b Keys) Keys {
	return Keys{Up: a.Up || b.Up, Down: a.Down || b.Down, Left: a.Left || b.Left, Right: a.Right || b.Right, Fire: a.Fire || b.Fire}
}

// Event builders.
func press(x, y float64) Event { return Event{Kind: Press, X: x, Y: y, Mouse: true} }
func pick(x, y float64, d Dir, a Action, ok bool) Event {
	return Event{Kind: Pick, Dir: d, Action: a, Enabled: ok, X: x, Y: y, Mouse: true}
}
func run(x, y float64, a Action) Event { return Event{Kind: Run, Action: a, X: x, Y: y, Mouse: true} }
func click(x, y float64) Event         { return Event{Kind: Click, X: x, Y: y, Mouse: true} }
func cancel(x, y float64) Event {
	return Event{Kind: Cancelled, Action: Cancel, X: x, Y: y, Mouse: true}
}
func hover(x, y float64) Event    { return Event{Kind: Hover, X: x, Y: y, Mouse: true} }
func ringOpen(x, y float64) Event { return Event{Kind: RingOpen, X: x, Y: y, Mouse: true} }
func keyPick(d Dir, a Action, ok bool) Event {
	return Event{Kind: Pick, Dir: d, Action: a, Enabled: ok}
}
func step(d Dir) Event { return Event{Kind: Step, Dir: d} }

var (
	keyPress  = Event{Kind: Press}
	keyClick  = Event{Kind: Click}
	keyCancel = Event{Kind: Cancelled, Action: Cancel}
)

func keyRun(a Action) Event { return Event{Kind: Run, Action: a} }

func TestTracker(t *testing.T) {
	moveMode, attackMode := ForMap(false, false), ForMap(true, false)
	noMove := moveMode.Restrict(func(a Action) bool { return a != Move })
	tests := []struct {
		name   string
		place  Placement
		frames []Input
		want   []Event
	}{
		{"hover reports pointer moves only", moveMode,
			[]Input{mouse(10, 10, false), mouse(10, 10, false), mouse(12, 10, false), {}},
			[]Event{hover(12, 10)}},
		{"click: press and release without crossing the threshold", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(110, 95, true), mouse(110, 95, false)},
			[]Event{press(100, 100), click(100, 100)}},
		{"hold without drag shows cancel and clicks on release", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(100, 100, true), mouse(100, 100, true), mouse(100, 100, false)},
			[]Event{press(100, 100), click(100, 100)}},
		{"drag up beyond the threshold picks move, release runs it", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(105, 60, true), mouse(105, 50, true), mouse(105, 50, false)},
			[]Event{press(100, 100), pick(100, 100, Up, Move, true), run(100, 100, Move)}},
		{"drag up in the attack mode picks attack", attackMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(100, 40, true), mouse(100, 40, false)},
			[]Event{press(100, 100), pick(100, 100, Up, Attack, true), run(100, 100, Attack)}},
		{"left, right, down by the dominant axis", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(50, 80, true), mouse(150, 120, true), mouse(130, 160, true), mouse(130, 160, false)},
			[]Event{press(100, 100), pick(100, 100, Left, ModeChange, true), pick(100, 100, Right, Map, true),
				pick(100, 100, Down, Info, true), run(100, 100, Info)}},
		{"a diagonal tie counts as vertical", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(140, 140, true), mouse(140, 140, false)},
			[]Event{press(100, 100), pick(100, 100, Down, Info, true), run(100, 100, Info)}},
		{"dragging back inside the threshold returns to cancel", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(100, 50, true), mouse(100, 90, true), mouse(100, 90, false)},
			[]Event{press(100, 100), pick(100, 100, Up, Move, true), pick(100, 100, Centre, Cancel, true), cancel(100, 100)}},
		{"releasing on a disabled action cancels", noMove,
			[]Input{mouse(100, 100, false), mouse(100, 100, true), mouse(100, 50, true), mouse(100, 50, false)},
			[]Event{press(100, 100), pick(100, 100, Up, Move, false), cancel(100, 100)}},
		{"a button already down when the pointer arrives does not press", moveMode,
			[]Input{mouse(100, 100, true), mouse(100, 40, true), mouse(100, 40, false)},
			[]Event{hover(100, 40)}},
		{"right click opens the ring, a click outside it cancels", moveMode,
			[]Input{mouse(100, 100, false), rightClick(100, 100), mouse(400, 400, false), mouse(400, 400, true), mouse(400, 400, false)},
			[]Event{ringOpen(100, 100), cancel(400, 400)}},
		{"right click again closes the ring", moveMode,
			[]Input{mouse(100, 100, false), rightClick(100, 100), rightClick(100, 100)},
			[]Event{ringOpen(100, 100), cancel(100, 100)}},
		{"keys: arrows step the cursor", moveMode,
			[]Input{keys(idle), keys(up), keys(up), keys(idle), keys(left), keys(with(left, down)), keys(idle)},
			[]Event{step(Up), step(Left), step(Down)}},
		{"keys: fire tap clicks", moveMode,
			[]Input{keys(idle), keys(fire), keys(fire), keys(idle)},
			[]Event{keyPress, keyClick}},
		{"keys: fire and up picks move, release runs it", moveMode,
			[]Input{keys(idle), keys(fire), keys(with(fire, up)), keys(fire), keys(idle)},
			[]Event{keyPress, keyPick(Up, Move, true), keyRun(Move)}},
		{"keys: the opposite arrow returns to cancel, a perpendicular one jumps", attackMode,
			[]Input{keys(idle), keys(fire), keys(with(fire, up)), keys(fire), keys(with(fire, down)), keys(fire),
				keys(with(fire, left)), keys(fire), keys(with(fire, up)), keys(fire), keys(with(fire, down)), keys(idle)},
			[]Event{keyPress, keyPick(Up, Attack, true), keyPick(Centre, Cancel, true), keyPick(Left, ModeChange, true),
				keyPick(Up, Attack, true), keyPick(Centre, Cancel, true), keyCancel}},
		{"keys: fire and an arrow in the same frame press and pick", moveMode,
			[]Input{keys(idle), keys(with(fire, right)), keys(idle)},
			[]Event{keyPress, keyPick(Right, Map, true), keyRun(Map)}},
		{"keys: an arrow held before fire does not pick", moveMode,
			[]Input{keys(idle), keys(down), keys(with(down, fire)), keys(down)},
			[]Event{step(Down), keyPress, keyClick}},
		{"keys: fire closes the ring", moveMode,
			[]Input{mouse(100, 100, false), rightClick(100, 100), keys(fire), keys(idle)},
			[]Event{ringOpen(100, 100), keyCancel}},
		{"keys are ignored while the mouse holds", moveMode,
			[]Input{mouse(100, 100, false), mouse(100, 100, true),
				{Pointer: Pointer{X: 100, Y: 100, Active: true, Down: true}, Keys: with(fire, up)},
				mouse(100, 100, false)},
			[]Event{press(100, 100), click(100, 100)}},
		{"the mouse is ignored while the keys hold", moveMode,
			[]Input{mouse(100, 100, false), {Pointer: Pointer{X: 100, Y: 100, Active: true}, Keys: fire},
				{Pointer: Pointer{X: 100, Y: 100, Active: true, Down: true}, Keys: fire}, mouse(100, 40, true), mouse(100, 40, false)},
			[]Event{keyPress, keyClick, hover(100, 40)}},
	}
	for _, tt := range tests {
		tr := NewTracker(Config{Threshold: pitch})
		tr.Placement = tt.place
		var got []Event
		for _, in := range tt.frames {
			got = append(got, tr.Update(in)...)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s:\n got %v\nwant %v", tt.name, got, tt.want)
		}
	}
}

func TestTrackerPlate(t *testing.T) {
	tr := NewTracker(Config{Threshold: pitch})
	tr.Placement = ForMap(false, false).Restrict(func(a Action) bool { return a != Info })
	plate := func() [3]any { a, ok, held := tr.Plate(); return [3]any{a, ok, held} }
	tr.Update(mouse(100, 100, false))
	if got := plate(); got != [3]any{Cancel, true, false} {
		t.Errorf("idle plate = %v", got)
	}
	tr.Update(mouse(100, 100, true))
	if got := plate(); got != [3]any{Cancel, true, true} {
		t.Errorf("held plate = %v, want cancel first", got)
	}
	tr.Update(mouse(100, 160, true))
	if got := plate(); got != [3]any{Info, false, true} {
		t.Errorf("plate after drag down = %v, want disabled info", got)
	}
	tr.Update(mouse(100, 160, false))
	if got := plate(); got != [3]any{Cancel, true, false} {
		t.Errorf("plate after release = %v", got)
	}
}

func TestTrackerRingRuns(t *testing.T) {
	tr := NewTracker(Config{Threshold: pitch})
	tr.Placement = ForMap(false, false)
	tr.Update(mouse(200, 200, false))
	if ev := tr.Update(rightClick(200, 200)); len(ev) != 1 || ev[0].Kind != RingOpen {
		t.Fatalf("right click gave %v", ev)
	}
	if _, ok := tr.Ring(); ok {
		t.Error("ring set before SetRing")
	}
	r := NewRing(200, 200, 0.5, tr.Placement.RingActions())
	tr.SetRing(r)
	x, y := r.Pos(1) // Map, clockwise from the top
	tr.Update(mouse(x, y, false))
	if got := tr.RingHover(); got != 1 {
		t.Errorf("RingHover = %d, want 1", got)
	}
	ev := tr.Update(mouse(x, y, true))
	if want := []Event{run(x, y, Map)}; !reflect.DeepEqual(ev, want) {
		t.Errorf("click on the ring gave %v, want %v", ev, want)
	}
	if _, ok := tr.Ring(); ok {
		t.Error("ring still open after a pick")
	}
	if ev := tr.Update(mouse(x, y, false)); len(ev) != 0 {
		t.Errorf("release after a ring pick gave %v, want nothing", ev)
	}
	if tr.RingHover() != -1 {
		t.Error("RingHover with the ring closed")
	}
}

func TestTrackerKeyRepeat(t *testing.T) {
	tr := NewTracker(Config{Threshold: pitch, RepeatDelay: 3, RepeatEvery: 2})
	var got []Event
	for range 8 { // frames 0..7 held: step at 0, then 3, 5, 7
		got = append(got, tr.Update(keys(right))...)
	}
	got = append(got, tr.Update(keys(idle))...)
	want := []Event{step(Right), step(Right), step(Right), step(Right)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("repeat gave %v, want %v", got, want)
	}
}

func TestTrackerReset(t *testing.T) {
	tr := NewTracker(Config{Threshold: pitch})
	tr.Update(mouse(100, 100, false))
	tr.Update(mouse(100, 100, true))
	tr.Reset()
	if _, _, held := tr.Plate(); held {
		t.Error("held after Reset")
	}
	if ev := tr.Update(mouse(100, 100, false)); len(ev) != 0 {
		t.Errorf("release after Reset gave %v", ev)
	}
}

func TestKeySets(t *testing.T) {
	held := map[ebiten.Key]bool{ebiten.KeyD: true, ebiten.KeyControlLeft: true, ebiten.KeyArrowLeft: true}
	pressed := func(k ebiten.Key) bool { return held[k] }
	if got, want := KeySetA.Read(pressed), (Keys{Left: true}); got != want {
		t.Errorf("key set A = %+v, want %+v", got, want)
	}
	if got, want := KeySetB.Read(pressed), (Keys{Up: true, Fire: true}); got != want {
		t.Errorf("key set B = %+v, want %+v", got, want)
	}
	// Two trackers on one keyboard: B picks while A steps.
	a, b := NewTracker(Config{Threshold: pitch}), NewTracker(Config{Threshold: pitch})
	b.Placement = ForMap(false, false)
	var gotA, gotB []Event
	for _, frame := range []map[ebiten.Key]bool{
		{},
		{ebiten.KeyControlLeft: true},
		{ebiten.KeyControlLeft: true, ebiten.KeyX: true, ebiten.KeyArrowDown: true},
		{},
	} {
		p := func(k ebiten.Key) bool { return frame[k] }
		gotA = append(gotA, a.Update(Input{Keys: KeySetA.Read(p)})...)
		gotB = append(gotB, b.Update(Input{Keys: KeySetB.Read(p)})...)
	}
	if want := []Event{step(Down)}; !reflect.DeepEqual(gotA, want) {
		t.Errorf("key set A gave %v, want %v", gotA, want)
	}
	if want := []Event{keyPress, keyPick(Left, ModeChange, true), keyRun(ModeChange)}; !reflect.DeepEqual(gotB, want) {
		t.Errorf("key set B gave %v, want %v", gotB, want)
	}
}
