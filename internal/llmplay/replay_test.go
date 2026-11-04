package llmplay

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// recordFallback records a RHEIN game of n fallback resolutions.
func recordFallback(t *testing.T, n int) string {
	t.Helper()
	g := rhein(t)
	if _, err := g.Play(FallbackDriver{}, n); err != nil {
		t.Fatal(err)
	}
	return g.Dir
}

// Replaying side by side, as a front end does (the LLM side's orders,
// its confirmation, the other side's orders, its confirmation, Resolve),
// reproduces every recorded state, through to the end of the battle.
func TestReplaySideBySide(t *testing.T) {
	dir := recordFallback(t, 40)
	rp, err := OpenReplay(dir)
	if err != nil {
		t.Fatal(err)
	}
	end, n, err := VerifyReplay(dir)
	if err != nil {
		t.Fatal(err)
	}
	if end == nil || n != len(rp.Lines) {
		t.Fatalf("recorded game: end %v after %d of %d lines, want an ended game", end, n, len(rp.Lines))
	}
	s, llm := rp.Session, rp.Header.LLMSide
	kinds := map[dosrules.CommandKind]int{}
	for _, l := range rp.Lines {
		for _, side := range []int{llm, 1 - llm} {
			cmds := SideCommands(l, side)
			for _, c := range cmds {
				kinds[c.Kind]++
			}
			if _, err := s.Replay(cmds); err != nil {
				t.Fatalf("resolution %d side %d: %v", l.Resolution, side, err)
			}
			s.Confirm(side)
		}
		if _, err := s.Resolve(); err != nil {
			t.Fatalf("resolution %d: %v", l.Resolution, err)
		}
		if err := CheckResolution(s, l); err != nil {
			t.Fatal(err)
		}
	}
	if got := s.Ended(); got == nil || *got != *end {
		t.Errorf("end %v, recorded %v", got, end)
	}
	t.Logf("replayed orders %v", kinds)
	if kinds[dosrules.CmdMove] == 0 || kinds[dosrules.CmdAttack] == 0 {
		t.Errorf("replayed orders %v, want moves and attacks", kinds)
	}
}

// SideCommands keeps one side's orders, surrender included, in order and
// drops the cycle commands.
func TestSideCommands(t *testing.T) {
	l := Line{Commands: []dosrules.Command{
		{Kind: dosrules.CmdMove, Side: 0, Unit: 1, Cell: 5},
		{Kind: dosrules.CmdAttack, Side: 1, Unit: 2, Cell: 6},
		{Kind: dosplay.CmdUnconfirm, Side: 0},
		{Kind: dosrules.CmdBuild, Side: 0, Hold: 3, Type: 7},
		{Kind: dosplay.CmdConfirm, Side: 0},
		{Kind: dosplay.CmdConfirm, Side: 1},
		{Kind: dosplay.CmdSurrender, Side: 1},
		{Kind: dosplay.CmdResolve},
	}}
	got0 := SideCommands(l, 0)
	want0 := []dosrules.Command{l.Commands[0], l.Commands[3]}
	if !slices.Equal(got0, want0) {
		t.Errorf("side 0: %v, want %v", got0, want0)
	}
	got1 := SideCommands(l, 1)
	want1 := []dosrules.Command{l.Commands[1], l.Commands[6]}
	if !slices.Equal(got1, want1) {
		t.Errorf("side 1: %v, want %v", got1, want1)
	}
}

// OpenReplay refuses a transcript whose opening state is not the level's.
func TestOpenReplayOpeningMismatch(t *testing.T) {
	dir := recordFallback(t, 1)
	_, lines, err := ReadTranscript(dir)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "transcript.jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bad := strings.Repeat("0", len(lines[0].StateBefore))
	raw = bytes.Replace(raw, []byte(lines[0].StateBefore), []byte(bad), 1)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = OpenReplay(dir)
	if err == nil || !strings.Contains(err.Error(), "transcript has "+bad) {
		t.Fatalf("err %v, want the opening hash mismatch", err)
	}
}

// A transcript with only its header opens on the fresh session.
func TestOpenReplayHeaderOnly(t *testing.T) {
	g := rhein(t)
	rp, err := OpenReplay(g.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rp.Lines) != 0 || rp.Header.Level != "1918-02:cp" || rp.Session == nil {
		t.Errorf("replay %+v", rp)
	}
}
