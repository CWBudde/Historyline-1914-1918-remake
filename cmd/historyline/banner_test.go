package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestEffectsNilSafe(t *testing.T) {
	var e *effects
	e.Show("WON", victoryFrames)
	e.Mark(hexgrid.Coord{}, false)
	e.Tick()
	e.Dismiss()
	if e.Banner() != nil || e.Marks() != nil {
		t.Fatal("nil effects must be empty")
	}
}

func TestMarks(t *testing.T) {
	var e effects
	at := hexgrid.Coord{Line: 1, Col: 2}
	e.Mark(at, true)
	e.Mark(at, false)
	m := e.Marks()
	if len(m) != 2 || m[0].At != at || m[0].Kind != markDeath || m[1].Kind != markHit {
		t.Fatalf("marks = %+v", m)
	}
}

func TestEffectsBounded(t *testing.T) {
	var e effects
	e.Mark(hexgrid.Coord{}, true)
	e.Show("WON", victoryFrames)
	for range victoryFrames {
		e.Tick()
	}
	if len(e.Marks()) != 0 || e.Banner() != nil {
		t.Fatalf("effects outlived their duration: %+v %+v", e.Marks(), e.Banner())
	}
}

func TestBannerReplacedAndDismissed(t *testing.T) {
	var e effects
	e.Show("TAKEN", captureFrames)
	if b := e.Banner(); b == nil || b.Frames != captureFrames {
		t.Fatalf("banner = %+v", b)
	}
	e.Show("WON", victoryFrames)
	if b := e.Banner(); b.Text != "WON" || b.Frames != victoryFrames {
		t.Fatal("a new banner must replace the shown one")
	}
	e.Dismiss()
	if e.Banner() != nil {
		t.Fatal("click must dismiss the banner")
	}
}

// TestExplosion plays the death animation of the interface art: its frames
// and times are the ones assets/interface/animations.json gives.
func TestExplosion(t *testing.T) {
	raw, err := os.ReadFile("../../assets/interface/animations.json")
	if err != nil {
		t.Fatal(err)
	}
	type anim struct {
		ID      string   `json:"id"`
		Frames  []string `json:"frames"`
		FrameMS []int    `json:"frame_ms"`
	}
	var anims []anim
	if err := json.Unmarshal(raw, &anims); err != nil {
		t.Fatal(err)
	}
	i := slices.IndexFunc(anims, func(a anim) bool { return a.ID == "death" })
	if i < 0 {
		t.Fatal("no death animation")
	}
	a := anims[i]
	if len(a.Frames) != explosionFrames || animTicks(a.FrameMS) != deathFrames {
		t.Fatalf("death animation of %d frames in %d ticks, want %d in %d", len(a.Frames), animTicks(a.FrameMS), explosionFrames, deathFrames)
	}
	for f, id := range a.Frames {
		if want := fmt.Sprintf("effect/explosion-%02d", f); id != want || a.FrameMS[f] != explosionMS {
			t.Errorf("frame %d: %s for %d ms, want %s for %d", f, id, a.FrameMS[f], want, explosionMS)
		}
	}

	// A death mark steps through the frames once, then ends.
	var e effects
	e.Mark(hexgrid.Coord{Col: 1}, true)
	var seen []int
	for len(e.Marks()) > 0 {
		if f := e.Marks()[0].frame(false); len(seen) == 0 || seen[len(seen)-1] != f {
			seen = append(seen, f)
		}
		e.Tick()
	}
	if !slices.Equal(seen, []int{0, 1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Errorf("explosion frames %v", seen)
	}
	if (mark{Kind: markDeath, Frames: 5, Total: deathFrames}).frame(true) != explosionFrames-1 {
		t.Error("reduced motion does not show the still smoke frame")
	}
}

// animTicks is the length in frames of one loop of an animation whose frames
// show frameMS milliseconds each.
func animTicks(frameMS []int) int {
	ms := 0
	for _, f := range frameMS {
		ms += f
	}
	return (ms*tps + 999) / 1000
}
