package llmplay

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// walker is an own foot unit on the map with somewhere to go, and the
// destination farthest from where it stands.
func walker(t *testing.T, g *Game, not int) (dosplay.UnitView, hexgrid.Coord) {
	t.Helper()
	own := ownUnits(g.S, g.Side)
	idx := make([]int, 0, len(own))
	for i := range own {
		idx = append(idx, i)
	}
	slices.Sort(idx)
	for _, i := range idx {
		u := own[i]
		if i == not || u.Held || u.Domain != dosplay.Infantry {
			continue
		}
		dests, err := g.S.Destinations(g.Side, i)
		if err != nil || len(dests) == 0 {
			continue
		}
		best := u.At
		for _, c := range dosplay.Coords(dests) {
			if _, isB := g.S.View(g.Side).Buildings[c]; !isB && hexgrid.Distance(c, u.At) > hexgrid.Distance(best, u.At) {
				best = c
			}
		}
		if best != u.At {
			return u, best
		}
	}
	t.Fatal("no foot unit with a destination")
	return dosplay.UnitView{}, hexgrid.Coord{}
}

// A plan sends its unit to the named hex and keeps another where it
// stands, hands back on arrival, and the game replays.
func TestAdvanceFollowsPlan(t *testing.T) {
	g := rhein(t)
	u, to := walker(t, g, -1)
	stay, _ := walker(t, g, u.Index)
	d := &Directives{
		Groups: []Group{
			{Units: []int{u.Index}, Goal: dirGoto, To: &[2]int{to.Line + 1, to.Col + 1}},
			{Units: []int{stay.Index}, Goal: dirStay},
		},
		Stop: Stop{Res: 4, Arrived: true},
	}
	rep, err := g.Advance(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	own := ownUnits(g.S, g.Side)
	if got := own[u.Index].At; got != to || !strings.Contains(rep.Stop, "arrived") || rep.Resolutions != 1 {
		t.Errorf("u%d at %s, want %s; stop %q after %d", u.Index, show(got), show(to), rep.Stop, rep.Resolutions)
	}
	if got, ok := own[stay.Index]; ok && got.At != stay.At {
		t.Errorf("u%d was to stay at %s, stands at %s", stay.Index, show(stay.At), show(got.At))
	}
	if len(rep.Lines) != rep.Resolutions {
		t.Errorf("%d lines for %d resolutions", len(rep.Lines), rep.Resolutions)
	}
	if _, n, err := VerifyReplay(g.Dir); err != nil || n != rep.Resolutions {
		t.Errorf("verify: %d resolutions, %v", n, err)
	}
}

func TestDirectivesCheck(t *testing.T) {
	g := rhein(t)
	u, _ := walker(t, g, -1)
	one := 9999
	d, err := ParseDirectives(`{"posture":"sideways","groups":[
		{"units":[9999],"goal":"goto","to":[2,2]},
		{"units":[],"goal":"dance"},
		{"units":[],"goal":"goto","to":[900,900]},
		{"units":[],"goal":"capture","to":[1,1]}],
		"build":[{"hold":9999,"types":["infantry"]}]}`)
	if err != nil {
		t.Fatal(err)
	}
	d.Groups = append(d.Groups, Group{Units: []int{u.Index}, Goal: dirEscort, Of: &one}, Group{Units: []int{u.Index}, Goal: dirStay})
	errs := strings.Join(d.Check(g.S, g.Side), "\n")
	for _, want := range []string{"posture", "u9999 is not yours", "goal \"dance\"", "off the map", "no building entrance", "escorted u9999", "earlier group", "hold 9999", "no units"} {
		if !strings.Contains(errs, want) {
			t.Errorf("check does not name %q:\n%s", want, errs)
		}
	}
	if _, err := g.Advance(d, nil); err == nil {
		t.Error("a refused plan was played")
	}
	if _, err := ParseDirectives(`{"groupz":[]}`); err == nil {
		t.Error("unknown field accepted")
	}
	if errs := (&Directives{Posture: "turtle", Build: []BuildRule{{Types: []string{"infantry"}}}}).Check(g.S, g.Side); len(errs) > 0 {
		t.Errorf("good plan refused: %v", errs)
	}
}

// Try leaves the game as it is, refuses bad candidates on their own and
// gives the same answer twice.
func TestTry(t *testing.T) {
	g := rhein(t)
	before, err := StateHash(g.S)
	if err != nil {
		t.Fatal(err)
	}
	u, to := walker(t, g, -1)
	cands := []Candidate{
		{Name: "coach"},
		{Name: "orders", Orders: `{"op":"done"}`},
		{Name: "bad", Orders: `{"op":"move","unit":9999,"to":[1,1]}`},
		{Name: "plan", Plan: &Directives{Posture: "turtle", Groups: []Group{{Units: []int{u.Index}, Goal: dirGoto, To: &[2]int{to.Line + 1, to.Col + 1}}}}},
	}
	var runs [2][]TryResult
	for i := range runs {
		if runs[i], err = Try(g.S, g.Side, cands, 3, nil); err != nil {
			t.Fatal(err)
		}
	}
	for i, r := range runs[0] {
		if r.Value != runs[1][i].Value {
			t.Errorf("%s: values %d and %d", r.Name, r.Value, runs[1][i].Value)
		}
		if rejected := r.Rejected != nil; rejected != (r.Name == "bad") {
			t.Errorf("%s: rejected %v", r.Name, r.Rejected)
		} else if !rejected && r.Resolutions != 3 {
			t.Errorf("%s: %d resolutions", r.Name, r.Resolutions)
		}
	}
	after, err := StateHash(g.S)
	if err != nil {
		t.Fatal(err)
	}
	if before != after || g.S.Ready(g.Side) {
		t.Error("Try changed the game")
	}
	out := FormatTry(g.S, g.Side, runs[0], true)
	for _, want := range []string{"NOW value", "coach: after 3 res", "bad: REJECTED", "plan: after 3 res"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
}

func TestUndo(t *testing.T) {
	g := rhein(t)
	for range 3 {
		if _, _, err := g.Step(""); err != nil {
			t.Fatal(err)
		}
	}
	left, err := Undo(g.Dir, 2)
	if err != nil || left != 1 {
		t.Fatalf("undo: left %d, %v", left, err)
	}
	h, lines, err := ReadTranscript(g.Dir)
	if err != nil || len(lines) != 1 {
		t.Fatalf("%d lines, %v", len(lines), err)
	}
	if h.PlayedBy() != g.Driver && g.Driver != "" {
		t.Errorf("driver %s became %s", g.Driver, h.PlayedBy())
	}
	saved, err := os.ReadFile(filepath.Join(g.Dir, "save.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadSession(g.S.Level, saved)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := StateHash(s); err != nil || got != lines[0].StateAfter {
		t.Errorf("save is not the state after resolution 1: %v", err)
	}
	if _, err := Undo(g.Dir, 5); err == nil {
		t.Error("undo beyond the start")
	}
}
