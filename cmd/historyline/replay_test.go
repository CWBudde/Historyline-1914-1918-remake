package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// replayResolutions is how many resolutions the recorded test game has.
const replayResolutions = 6

// recordRhein records a RHEIN game of n resolutions played by the
// fallback driver and returns its directory.
func recordRhein(t *testing.T, n int) string {
	t.Helper()
	lvl, err := llmplay.ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := llmplay.NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	g, err := llmplay.NewGame(s, 0, dir, "1918-02:cp", "RHEIN", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Play(llmplay.FallbackDriver{}, n); err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	return dir
}

// newReplayTest is the split screen replaying the game in dir.
func newReplayTest(t *testing.T, dir string) (*game, *llmplay.Replay) {
	t.Helper()
	rp, err := llmplay.OpenReplay(dir)
	if err != nil {
		t.Fatal(err)
	}
	g := replayGame(rp, false)
	g.applyLayout(1800, 1000, 1)
	g.layoutSplit()
	return g, rp
}

// replayRun is what playing a replay out showed.
type replayRun struct {
	ticks     int
	walks     [2]int  // moves that walked on each side's panel
	camMoved  [2]bool // each panel's camera left its first place
	ownerless int     // frames a side's walk showed while the other acted
}

// playReplay ticks g's replay until it is done, as Update does.
func playReplay(t *testing.T, g *game) replayRun {
	t.Helper()
	d := g.dos
	var r replayRun
	var last [2]*walk
	type place struct{ pair, row int }
	var start [2]place
	for side := range 2 {
		start[side] = place{g.split.cams[side].pair, g.split.cams[side].row}
	}
	for !d.replayDone() {
		if r.ticks == 400000 {
			t.Fatalf("replay not done after %d ticks (line %d of %d)", r.ticks, d.replay.at, len(d.replay.lines))
		}
		d.Tick()
		g.updateSplit()
		r.ticks++
		if a := d.act; a != nil {
			if w := d.walkOf(a.side); !w.Done() && w != last[a.side] {
				r.walks[a.side]++
				last[a.side] = w
			}
			if !d.walkOf(1 - a.side).Done() {
				r.ownerless++
			}
		}
		for side := range 2 {
			c := g.split.cams[side]
			if (place{c.pair, c.row}) != start[side] {
				r.camMoved[side] = true
			}
		}
	}
	return r
}

// TestReplayPlaysBothSides: a recorded game replays in the split screen,
// both sides' moves walking through the computer's act path, every
// resolution matching the transcript, and two runs take the same frames.
func TestReplayPlaysBothSides(t *testing.T) {
	dir := recordRhein(t, replayResolutions)
	g, rp := newReplayTest(t, dir)
	d := g.dos
	if d.side != rp.Header.LLMSide || d.self == nil || d.planner != d.self {
		t.Fatalf("replay control: side %d (recorded %d), self %v", d.side, rp.Header.LLMSide, d.self)
	}
	r := playReplay(t, g)
	if err := d.replayErr(); err != nil {
		t.Fatalf("replay error: %v", err)
	}
	if d.replay.at != len(rp.Lines) || len(rp.Lines) != replayResolutions {
		t.Fatalf("replayed %d of %d lines, recorded %d", d.replay.at, len(rp.Lines), replayResolutions)
	}
	got, err := llmplay.StateHash(d.s)
	if err != nil {
		t.Fatal(err)
	}
	if want := rp.Lines[len(rp.Lines)-1].StateAfter; got != want {
		t.Errorf("final state %s, transcript %s", got, want)
	}
	t.Logf("%d ticks, walks %v, cameras moved %v", r.ticks, r.walks, r.camMoved)
	for side := range 2 {
		if r.walks[side] == 0 {
			t.Errorf("no move of side %d walked", side)
		}
		if !r.camMoved[side] {
			t.Errorf("panel %d's camera never moved", side)
		}
	}
	if r.ownerless > 0 {
		t.Errorf("%d frames a walk showed on the side not acting", r.ownerless)
	}
	again, _ := newReplayTest(t, dir)
	if r2 := playReplay(t, again); r2.ticks != r.ticks || r2.walks != r.walks {
		t.Errorf("second run: %d ticks, walks %v; first %d ticks, walks %v", r2.ticks, r2.walks, r.ticks, r.walks)
	}
}

// TestReplayStopsOnMismatch: a resolution whose state differs from the
// transcript stops the replay with the error, giving no further orders.
func TestReplayStopsOnMismatch(t *testing.T) {
	dir := recordRhein(t, 3)
	_, lines, err := llmplay.ReadTranscript(dir)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "transcript.jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bad := strings.Repeat("0", len(lines[1].StateAfter))
	raw = bytes.Replace(raw, []byte(`"stateAfter":"`+lines[1].StateAfter), []byte(`"stateAfter":"`+bad), 1)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	g, _ := newReplayTest(t, dir)
	d := g.dos
	playReplay(t, g)
	err = d.replayErr()
	if err == nil || !strings.Contains(err.Error(), "transcript has "+bad) {
		t.Fatalf("replay error %v, want the mismatch of resolution 2", err)
	}
	if d.replay.at != 1 || d.s.Ready(0) || d.s.Ready(1) {
		t.Errorf("after the mismatch: line %d, ready %v %v", d.replay.at, d.s.Ready(0), d.s.Ready(1))
	}
	if !strings.Contains(d.Message, "transcript has "+bad) {
		t.Errorf("message %q, want the mismatch", d.Message)
	}
	before := d.s.Version()
	for range 100 {
		d.Tick()
	}
	if d.s.Version() != before {
		t.Error("the stopped replay still gave orders")
	}
}

// TestReplayEndsWithTheBattle: a transcript through to the battle's end
// plays its end.
func TestReplayEndsWithTheBattle(t *testing.T) {
	if testing.Short() {
		t.Skip("plays a whole battle")
	}
	dir := recordRhein(t, 300)
	g, rp := newReplayTest(t, dir)
	r := playReplay(t, g)
	d := g.dos
	if err := d.replayErr(); err != nil {
		t.Fatal(err)
	}
	if d.s.Ended() == nil || !d.over || d.replay.at != len(rp.Lines) {
		t.Errorf("after %d ticks: ended %v, over %v, line %d of %d", r.ticks, d.s.Ended(), d.over, d.replay.at, len(rp.Lines))
	}
}

// TestReplayLevelBackdrop: a replayed level's backdrop is found below the
// levels directory, as the menu's level is.
func TestReplayLevelBackdrop(t *testing.T) {
	rp, err := llmplay.OpenReplay(recordRhein(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	lvl, err := replayLevel(rp)
	if err != nil {
		t.Fatal(err)
	}
	if lvl.Extra == nil || lvl.Extra.Backdrop != "1918-02/backdrop.jpg" {
		t.Fatalf("backdrop %+v, want 1918-02/backdrop.jpg", lvl.Extra)
	}
	if _, err := levelBackdrop("levels", lvl.Extra.Backdrop); err != nil {
		t.Fatal(err)
	}
}
