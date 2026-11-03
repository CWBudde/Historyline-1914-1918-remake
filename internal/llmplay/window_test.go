package llmplay

import (
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func TestParseLines(t *testing.T) {
	got, err := ParseLines("1-20, 75-95,8")
	if err != nil {
		t.Fatal(err)
	}
	if want := [][2]int{{1, 20}, {75, 95}, {8, 8}}; !slices.Equal(got, want) {
		t.Errorf("ParseLines = %v, want %v", got, want)
	}
	for _, bad := range []string{"x", "5-3", "0-4", "3-y"} {
		if _, err := ParseLines(bad); err == nil {
			t.Errorf("ParseLines(%q) accepted", bad)
		}
	}
}

// candl is CANDL (1918-04:cp) at its first phase: CP moves while EC, the
// action side, has its long-range guns in reach of CP units.
func candl(t *testing.T) *dosplay.Session {
	t.Helper()
	lvl, err := ShipLevel("1918-04", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Every hex an enemy long-range gun can shoot at now carries the "!"
// mark for the unit standing there: the mark's reach is the engine's.
func TestDestNotesMarkGunReach(t *testing.T) {
	s := candl(t)
	v := s.View(0)
	notes := newDestNotes(s, 0, v)
	types := s.Game.Profile.Tables.Types
	checked := 0
	for _, e := range v.Units {
		if e.Side != 1 || types[e.Type].TargetFlags&longRangeFlag == 0 {
			continue
		}
		for _, c := range dosplay.Coords(s.Targets(1, e.Index)) {
			u, ok := v.Units[c]
			if !ok || u.Side != 0 {
				continue
			}
			checked++
			if !strings.Contains(notes(u, c), "!") {
				t.Errorf("e%d %s reaches u%d at %s, but the hex is not marked", e.Index, e.ID, u.Index, show(c))
			}
		}
	}
	if checked == 0 {
		t.Fatal("no long-range gun has a target at the start of CANDL")
	}
}

// A window keeps the header and drops units, threats and orders outside
// its lines; -full lists every destination.
func TestObserveWindow(t *testing.T) {
	s := candl(t)
	p := Observe(s, 0, "", Window{Lines: [][2]int{{76, 95}}, Full: true})
	for _, want := range []string{"GAME CANDL", "WINDOW lines 76-95", "u131 CP_cavalry", "e129 EC_hArtillery", "[93,2] factory CP"} {
		if !strings.Contains(p, want) {
			t.Errorf("windowed prompt lacks %q", want)
		}
	}
	for _, drop := range []string{"u21 CP_a7v", "e37 EC_eliteInf", "[8,11] hq", "u21 move->", "more legal"} {
		if strings.Contains(p, drop) {
			t.Errorf("windowed prompt keeps %q", drop)
		}
	}
	if whole := Observe(s, 0, "", Window{}); !strings.Contains(whole, "u21 move->") || strings.Contains(whole, "WINDOW") {
		t.Error("zero window does not give the whole prompt")
	}
}

// SuggestAct refuses movement phases; in the action phase its JSONL is a
// batch Apply takes whole, with the opening bunkers' attacks in it.
func TestSuggestAct(t *testing.T) {
	s := candl(t)
	if _, err := SuggestAct(s, 0); err == nil {
		t.Fatal("SuggestAct accepted a movement phase")
	}
	s.Confirm(0)
	s.Confirm(1)
	if _, err := s.Resolve(); err != nil {
		t.Fatal(err)
	}
	text, err := SuggestAct(s, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, `"op":"attack"`) {
		t.Fatalf("no attack suggested:\n%s", text)
	}
	orders, err := ParseOrders(text)
	if err != nil {
		t.Fatal(err)
	}
	if _, errs, err := Apply(s, 0, orders); err != nil || len(errs) > 0 {
		t.Fatalf("suggested batch rejected: %v %v", err, errs)
	}
}
