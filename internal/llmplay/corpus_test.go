package llmplay

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

const shippedLevels = "../../levels"

// Every live recording in levels/replays.json is still a win against the
// current computer side and rules, at the resolution and score the manifest
// gives; every recording file has an entry. A failure here means the engine
// or dosai changed: make the recordings again with
// `hl llmplay corpus regen` (stale ones are listed, not checked).
func TestShippedReplays(t *testing.T) {
	t.Parallel()
	m, err := ReadManifest(shippedLevels)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := Targets()
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 104 {
		t.Errorf("%d targets, want 52 months x cp, ec", len(targets))
	}
	refs := make([]string, len(targets))
	for i, tg := range targets {
		refs[i] = tg.Ref()
	}
	orphans, err := Orphans(shippedLevels, m)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range orphans {
		t.Errorf("%s has no entry in levels/%s", o, ManifestFile)
	}
	for _, e := range m.Replays {
		if !slices.Contains(refs, e.Level) {
			t.Errorf("%s is not a target", e.Level)
		}
		if e.Stale != "" {
			t.Logf("%s stale: %s", e.Level, e.Stale)
			continue
		}
		t.Run(e.Level, func(t *testing.T) {
			t.Parallel()
			if c := CheckEntry(shippedLevels, e); c.Err != nil {
				t.Errorf("%v; run hl llmplay corpus regen -level %s", c.Err, e.Level)
			}
		})
	}
}

// The recorded side's orders, played back against today's dosai, give the
// recorded game state for state.
func TestOrdersPlayerReplaysRecording(t *testing.T) {
	src := filepath.Join(shippedLevels, "1918-02", "replays", "cp.jsonl")
	h, lines, err := ReadTranscript(src)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := playOrders(dir, h, lines); err != nil {
		t.Fatal(err)
	}
	if err := sameStates(src, dir); err != nil {
		t.Fatal(err)
	}
	// One order fewer than recorded: the recording runs out, not a win.
	if err := playOrders(t.TempDir(), h, lines[:len(lines)-1]); err == nil {
		t.Error("orders one resolution short still won")
	}
}

// A fork keeps the first resolutions and goes on like a game; an unended
// game is not adopted.
func TestForkGoesOn(t *testing.T) {
	src := filepath.Join(shippedLevels, "1918-02", "replays", "cp.jsonl")
	dst := filepath.Join(t.TempDir(), "fork")
	if err := Fork(src, 10, dst); err != nil {
		t.Fatal(err)
	}
	h, lines, err := ReadTranscript(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 10 || h.PlayedBy() != DriverLLM {
		t.Fatalf("fork has %d lines, driver %s", len(lines), h.PlayedBy())
	}
	lvl, err := ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(filepath.Join(dst, "save.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadSession(lvl, saved)
	if err != nil {
		t.Fatal(err)
	}
	g, err := ResumeGame(s, dst)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := g.Step(""); err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	if _, n, err := VerifyReplay(dst); err != nil || n != 11 {
		t.Fatalf("verify fork: n=%d err=%v", n, err)
	}
	if _, err := Adopt(dst, t.TempDir()); err == nil {
		t.Error("an unended fork was adopted")
	}
	if err := Fork(src, 99, filepath.Join(t.TempDir(), "x")); err == nil {
		t.Error("fork past the end")
	}
}

// Adopt writes the compact recording and the manifest entry; Expand gives
// the events back.
func TestAdoptCompact(t *testing.T) {
	root := t.TempDir()
	res := Sweep([]Target{{Month: "1918-02", Mode: scenario.CP}}, []string{"strike", "fallback"}, t.TempDir(), 40, 2, nil)
	if len(res) != 2 || !res[0].Won() || res[1].Won() {
		t.Fatalf("sweep: %+v", res)
	}
	e, err := Adopt(res[0].Dir, root)
	if err != nil {
		t.Fatal(err)
	}
	if e.Driver != "strike" || e.Resolutions != 25 || e.Side != 0 || e.Code != "RHEIN" {
		t.Errorf("entry %+v", e)
	}
	path, err := RecordingFile(root, "1918-02:cp")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"events"`) {
		t.Error("the recording keeps its events")
	}
	if c := CheckEntry(root, e); c.Err != nil {
		t.Fatal(c.Err)
	}
	rp, err := OpenReplay(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Expand(rp); err != nil {
		t.Fatal(err)
	}
	_, full, err := ReadTranscript(res[0].Dir)
	if err != nil {
		t.Fatal(err)
	}
	for i, l := range rp.Lines {
		if len(l.Events) != len(full[i].Events) {
			t.Fatalf("resolution %d: %d events expanded, %d recorded", l.Resolution, len(l.Events), len(full[i].Events))
		}
	}
	if _, err := Adopt(res[1].Dir, root); err == nil {
		t.Error("a lost game was adopted")
	}
}

// Players refuse levels they are not made for.
func TestFindPlayer(t *testing.T) {
	if _, err := FindPlayer("strike", "1918-04:cp"); err == nil {
		t.Error("strike plays CANDL")
	}
	if _, err := FindPlayer("coach", "1918-04:ec"); err != nil {
		t.Error(err)
	}
	if _, err := FindPlayer("nobody", "1918-02:cp"); err == nil {
		t.Error("unknown player found")
	}
}
