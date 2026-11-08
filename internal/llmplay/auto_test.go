package llmplay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// StepAuto plays the movement orders and the action phase after them by
// suggest, records both resolutions so the game replays, and reports the
// events of both; a brief window drops the legends and own moves.
func TestStepAutoBrief(t *testing.T) {
	g := rhein(t)
	g.Window.Brief = true
	full := Observe(g.S, g.Side, "", Window{})
	brief, err := g.Prompt()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(brief, "Reply with JSONL") || strings.Contains(brief, "destination marks") || len(brief) >= len(full) {
		t.Errorf("brief prompt keeps the legends (%d vs %d bytes)", len(brief), len(full))
	}
	for range 3 {
		if g.S.Phase(g.Side) != dosrules.PhaseMovement {
			t.Fatalf("StepAuto left an action phase to the player")
		}
		rej, evs, err := g.StepAuto(`{"op":"done"}`)
		if err != nil || len(rej) > 0 || len(evs) == 0 {
			t.Fatalf("StepAuto: %v %v (%d events)", rej, err, len(evs))
		}
		if !strings.Contains(g.digest, "(action phase played by suggest)") {
			t.Errorf("digest does not mark the suggested phase:\n%s", g.digest)
		}
	}
	if g.resolution != 6 {
		t.Errorf("%d resolutions after 3 auto steps, want 6", g.resolution)
	}
	if rej, _, err := g.StepAuto(`{"op":"attack","unit":9999,"target":[1,1]}`); err != nil || len(rej) == 0 {
		t.Errorf("bad orders: rejected %v, err %v", rej, err)
	}
	if _, n, err := VerifyReplay(g.Dir); err != nil || n != 6 {
		t.Errorf("verify: %d resolutions, %v", n, err)
	}
	root := t.TempDir()
	game := filepath.Join(root, "sweep", "1918-02-cp-test")
	raw, err := os.ReadFile(TranscriptPath(g.Dir))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(game, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(TranscriptPath(game), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	ps, err := ForkPoints(root, "1918-02:cp")
	if err != nil || len(ps) != 1 || ps[0].Resolutions != 6 || ps[0].Result != "open" || ps[0].At < 1 {
		t.Errorf("fork points %+v, %v", ps, err)
	}
}
