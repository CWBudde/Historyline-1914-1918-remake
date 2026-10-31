package llmplay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func rhein(t *testing.T) *Game {
	t.Helper()
	lvl, err := ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(s, 0, t.TempDir(), "1918-02:cp", "RHEIN", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = g.Close() })
	return g
}

// RHEIN loads and the fair prompt names both sides' strength without
// leaking TrueView-only data (enemy section is the stale view).
func TestObserveRhein(t *testing.T) {
	g := rhein(t)
	p, err := g.Prompt()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"YOU=CP", "PHASE movement", "hold72", "ENEMY LAST-KNOWN (positions stale", "LEGAL THIS PHASE:", "exit->", "[19,28] depot NONE"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	if !strings.Contains(StaticMap(g.S), "STATIC MAP 32x47") {
		t.Error("static map wrong size")
	}
}

// A batch with one illegal order applies nothing.
func TestOrdersAtomic(t *testing.T) {
	g := rhein(t)
	before, _ := StateHash(g.S)
	rej, _, err := g.Step("{\"op\":\"move\",\"unit\":49,\"to\":[26,23]}\n{\"op\":\"move\",\"unit\":9999,\"to\":[1,1]}\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(rej) == 0 {
		t.Fatal("expected rejection")
	}
	after, _ := StateHash(g.S)
	if before != after {
		t.Error("rejected batch changed the state")
	}
}

// A legal batch plus dosai's plan resolves; the transcript then
// verifies hash-for-hash (VerifyReplay).
func TestStepAndReplay(t *testing.T) {
	dir := t.TempDir()
	lvl, err := ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(s, 0, dir, "1918-02:cp", "RHEIN", false)
	if err != nil {
		t.Fatal(err)
	}
	rej, evs, err := g.Step("{\"op\":\"move\",\"unit\":49,\"to\":[26,23]}\n")
	if err != nil || len(rej) > 0 {
		t.Fatalf("step: rej=%v err=%v", rej, err)
	}
	if len(evs) == 0 {
		t.Fatal("no events from resolution")
	}
	if _, _, err := g.Step(""); err != nil { // fallback action phase
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	_, n, err := VerifyReplay(dir)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("replayed %d resolutions, want 2", n)
	}
	rep, err := Review(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if rep.Resolutions != 2 {
		t.Errorf("review resolutions %d", rep.Resolutions)
	}
}

// The fallback-only game runs several resolutions without a refusal
// erroring the loop: this is also a headless RHEIN run vs dosai.
func TestFallbackPlays(t *testing.T) {
	g := rhein(t)
	if _, err := g.Play(FallbackDriver{}, 6); err != nil {
		t.Fatal(err)
	}
	if g.FallbackPhases != 6 {
		t.Errorf("fallback phases %d, want 6", g.FallbackPhases)
	}
	if _, _, err := ReadTranscript(g.Dir); err != nil {
		t.Fatal(err)
	}
	if _, n, err := VerifyReplay(g.Dir); err != nil || n != 6 {
		t.Errorf("verify: n=%d err=%v", n, err)
	}
}

// The Coach WINS RHEIN against dosai: annihilation at resolution 146
// (deterministic; see docs/llmplay-strategy.md battle log v1-v32).
// Regression: coach v2/v3 aborted at resolution 17 with "No such unit"
// (destroyed cargo record in a hold; two-cell ship rears ordered).
func TestCoachPlaysFullGame(t *testing.T) {
	g := rhein(t)
	end, err := g.PlayCoach(&Coach{}, 600)
	if err != nil {
		t.Fatal(err)
	}
	if end == nil {
		t.Fatal("coach game did not end within 600 resolutions")
	}
	if end.Winner != 0 {
		t.Errorf("winner side %d won %v score %d, want CP (side 0) to win RHEIN", end.Winner, end.Won, end.Score)
	}
	if _, n, err := VerifyReplay(g.Dir); err != nil || n != 146 {
		t.Errorf("verify: n=%d err=%v, want the deterministic 146-resolution win", n, err)
	}
	t.Logf("coach game: winner side %d won %v score %d", end.Winner, end.Won, end.Score)
}

// RenderReplay replays a recorded game and writes a frame per
// resolution (opening + 2 resolutions here).
func TestRenderReplay(t *testing.T) {
	dir := t.TempDir()
	lvl, err := ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(s, 0, dir, "1918-02:cp", "RHEIN", false)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := g.StepPlayer(&Coach{}); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	n, err := RenderReplay(dir, r, RenderOptions{Every: 1, Density: 2, Div: 16})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("frames %d, want 3 (opening + 2 resolutions)", n)
	}
	for _, f := range []string{"0000.png", "0001.png", "0002.png"} {
		if _, err := os.Stat(filepath.Join(dir, "frames", f)); err != nil {
			t.Error(err)
		}
	}
}

// RenderReplay with the level's own painted backdrop writes the same
// hash-verified frames on the painting instead of terrain tiles.
func TestRenderReplayBackdrop(t *testing.T) {
	dir := t.TempDir()
	lvl, err := ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	bd, err := LevelBackdrop("1918-02", lvl)
	if err != nil {
		t.Fatal(err)
	}
	if got := bd.Bounds(); got.Dx() != 7737 || got.Dy() != 3302 {
		t.Fatalf("RHEIN backdrop bounds %v, want 7737x3302", got)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(s, 0, dir, "1918-02:cp", "RHEIN", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.StepPlayer(&Coach{}); err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	n, err := RenderReplay(dir, r, RenderOptions{Every: 1, Density: 2, Div: 16, Backdrop: bd})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("frames %d, want 2 (opening + 1 resolution)", n)
	}
	for _, f := range []string{"0000.png", "0001.png"} {
		if _, err := os.Stat(filepath.Join(dir, "frames", f)); err != nil {
			t.Error(err)
		}
	}
}

// Project keeps every on-map unit for the frame renderer.
func TestProjectUnits(t *testing.T) {
	g := rhein(t)
	proj := Project(g.S)
	n := 0
	for _, line := range proj.Lines {
		for _, h := range line {
			if h.Unit != "" {
				n++
			}
		}
	}
	if want := len(g.S.TrueView().Units); n == 0 || n > want {
		t.Errorf("projected units %d, true view %d", n, want)
	}
}
