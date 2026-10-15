package presentation

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func animations() []Animation {
	return []Animation{{ID: "hit", Frames: []string{"hit"}, FrameMS: []int{1000}, ReducedFrames: []string{"ring"}, ReducedMS: []int{250}, Anchor: [2]int{36, 36}, Skippable: true}, {ID: "death", Frames: []string{"boom0", "boom1"}, FrameMS: []int{100, 200}, ReducedFrames: []string{"smoke"}, ReducedMS: []int{200}, Anchor: [2]int{48, 48}, Skippable: true}, {ID: "capture", Frames: []string{"capture"}, FrameMS: []int{900}, ReducedFrames: []string{"capture"}, ReducedMS: []int{250}, Skippable: true}, {ID: "round", Frames: []string{"round"}, FrameMS: []int{900}, ReducedFrames: []string{"round"}, ReducedMS: []int{250}, Skippable: true}}
}

func TestOrderingBoundariesAndImmutableInputs(t *testing.T) {
	defs := animations()
	steps := []Step{{ID: "move", Position: Point{0, 80}, To: Point{0, 0}, Duration: 600 * time.Millisecond}, {ID: "hit", Effect: "hit"}, {ID: "death", Effect: "death"}, {ID: "capture", Effect: "capture"}, {ID: "round", Effect: "round"}}
	p, err := New(defs, steps, false)
	if err != nil {
		t.Fatal(err)
	}
	defs[0].Frames[0] = "changed"
	steps[0].To = Point{99, 99}
	if err := p.Advance(300 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if s := p.State(); s.Position != (Point{0, 40}) || s.Facing != "n" || s.Frame != "" {
		t.Fatal(s)
	}
	for _, check := range []struct {
		advance   time.Duration
		id, frame string
	}{{300 * time.Millisecond, "hit", "hit"}, {time.Second, "death", "boom0"}, {100 * time.Millisecond, "death", "boom1"}, {200 * time.Millisecond, "capture", "capture"}, {900 * time.Millisecond, "round", "round"}, {900 * time.Millisecond, "", ""}} {
		if err := p.Advance(check.advance); err != nil {
			t.Fatal(err)
		}
		s := p.State()
		if s.ID != check.id || s.Frame != check.frame {
			t.Fatalf("%+v expected %s %s", s, check.id, check.frame)
		}
	}
	if !p.State().Done || p.Skip() {
		t.Fatal("must finish once")
	}
}

func TestTimePartitionAndSkip(t *testing.T) {
	steps := []Step{{ID: "move", Position: Point{}, To: Point{71, -40}, Duration: 600 * time.Millisecond}, {ID: "death", Effect: "death", Position: Point{71, -40}}, {ID: "round", Effect: "round", Position: Point{71, -40}}}
	a, err := New(animations(), steps, false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := New(animations(), steps, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Advance(815 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	for range 163 {
		if err := b.Advance(5 * time.Millisecond); err != nil {
			t.Fatal(err)
		}
	}
	if a.State() != b.State() {
		t.Fatalf("partition changes playback: %+v / %+v", a.State(), b.State())
	}
	if !a.Skip() || a.State().ID != "round" || a.State().Position != (Point{71, -40}) {
		t.Fatal(a.State())
	}
	if !a.Skip() || !a.State().Done {
		t.Fatal(a.State())
	}
	c, err := New(animations(), steps, false)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Skip() || c.State().Position != (Point{71, -40}) || c.State().Facing != "ne" {
		t.Fatal(c.State())
	}
}

func TestReducedSnapsWithoutFlashing(t *testing.T) {
	p, err := New(animations(), []Step{{ID: "move", Position: Point{}, To: Point{-71, 40}, Duration: time.Second}, {ID: "death", Effect: "death", Position: Point{-71, 40}}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if s := p.State(); s.ID != "death" || s.Position != (Point{-71, 40}) || s.Facing != "sw" || s.Frame != "smoke" {
		t.Fatal(s)
	}
	if err := p.Advance(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if !p.State().Done {
		t.Fatal(p.State())
	}
}

func TestSixFacingsAndInvalidInputs(t *testing.T) {
	for _, c := range []struct {
		to   Point
		want string
	}{{Point{0, -80}, "n"}, {Point{-71, -40}, "nw"}, {Point{-71, 40}, "sw"}, {Point{0, 80}, "s"}, {Point{71, 40}, "se"}, {Point{71, -40}, "ne"}} {
		got, err := Facing(Point{}, c.to)
		if err != nil || got != c.want {
			t.Fatalf("%+v: %s %v", c, got, err)
		}
	}
	for _, to := range []Point{{0, 0}, {math.NaN(), 1}, {math.Inf(1), 0}} {
		if _, err := Facing(Point{}, to); err == nil {
			t.Fatal("bad move accepted", to)
		}
	}
	invalid := []Animation{{ID: "loop", Frames: []string{"x"}, FrameMS: []int{1}, ReducedFrames: []string{"x"}, ReducedMS: []int{1}, Loop: true}, {ID: "empty"}, {ID: "zero", Frames: []string{"x"}, FrameMS: []int{0}, ReducedFrames: []string{"x"}, ReducedMS: []int{1}}}
	for _, a := range invalid {
		if _, err := New([]Animation{a}, nil, false); err == nil {
			t.Fatal("bad animation accepted", a)
		}
	}
	p, err := New(animations(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	before := p.State()
	for _, delta := range []time.Duration{-1, 25 * time.Hour} {
		if err := p.Advance(delta); err == nil || !reflect.DeepEqual(before, p.State()) {
			t.Fatal("invalid time accepted or state mutated", delta)
		}
	}
}

func TestNonSkippableEffectStillExpires(t *testing.T) {
	defs := animations()
	defs[0].Skippable = false
	p, err := New(defs, []Step{{ID: "hit", Effect: "hit"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if p.Skip() {
		t.Fatal("non-skippable effect skipped")
	}
	if err := p.Advance(time.Second); err != nil {
		t.Fatal(err)
	}
	if !p.State().Done {
		t.Fatal(p.State())
	}
}
