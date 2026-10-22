package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The tests below play the original order cycle through dosControl as the
// player does: clicks, the contents menu, F1 and frames.

// newDosTest is a session on a shipped level played by side 0 (the level's
// side) against planner.
func newDosTest(t *testing.T, name string, planner dosPlanner) *dosControl {
	t.Helper()
	lvl, err := loadLevel("", name)
	if err != nil {
		t.Fatal(err)
	}
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	s, err := dosplay.New(lvl, set, dosplay.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return newDosControl(s, 0, planner, false)
}

// ownUnits are the hexes of the side's units on its map, in line order.
func ownUnits(d *dosControl) []hexgrid.Coord {
	var out []hexgrid.Coord
	for _, p := range dosplay.Coords(toSet(d.own().Units)) {
		if d.own().Units[p].Side == d.side {
			out = append(out, p)
		}
	}
	return out
}

// nearestEnemy is the choice closest to a unit of the other side.
func nearestEnemy(d *dosControl, choices []hexgrid.Coord) hexgrid.Coord {
	best, bd := choices[0], 1<<30
	for p, u := range d.own().Units {
		if u.Side != 1-d.side {
			continue
		}
		for _, c := range choices {
			if dist := hexgrid.Distance(c, p); dist < bd || dist == bd && (c.Line < best.Line || c.Line == best.Line && c.Col < best.Col) {
				best, bd = c, dist
			}
		}
	}
	return best
}

// advance moves every own unit that can towards the enemy by clicks.
func advance(t *testing.T, d *dosControl) {
	t.Helper()
	for _, at := range ownUnits(d) {
		d.Click(at)
		if !d.selected || len(d.moves) == 0 {
			d.Clear()
			continue
		}
		to := nearestEnemy(d, dosplay.Coords(d.moves))
		d.Click(to)
		d.SkipWalk()
	}
}

// attackAll gives every own unit with a target an attack order by clicks.
func attackAll(t *testing.T, d *dosControl) int {
	t.Helper()
	n := 0
	for _, at := range ownUnits(d) {
		d.Click(at)
		if !d.selected || len(d.targets) == 0 {
			d.Clear()
			continue
		}
		before := len(d.s.Orders(d.side))
		d.Click(dosplay.Coords(d.targets)[0])
		if len(d.s.Orders(d.side)) != before+1 {
			t.Fatalf("a click on a red hex gave no order: %q", d.Message)
		}
		n++
	}
	return n
}

// played is one event of a resolution and what the playback showed.
type played struct {
	ev     dosplay.Event
	report *report
	frozen map[hexgrid.Coord]dosplay.UnitView
}

// finish confirms, resolves and plays the resolution back to the next phase,
// closing panels and walks as a waiting player would, and returns what was
// played.
func finish(t *testing.T, d *dosControl) []played {
	t.Helper()
	if !d.s.Ready(d.side) {
		d.EndTurn()
	}
	var out []played
	for range 10000 {
		var next *dosplay.Event
		if d.busy() && len(d.queue) > 0 && d.walk.Done() && !d.report.Active() {
			next = &d.queue[0]
		}
		var ev dosplay.Event
		if next != nil {
			ev = *next
		}
		d.Tick()
		if next != nil {
			p := played{ev: ev, report: d.report}
			if d.frozen != nil {
				p.frozen = maps.Clone(d.frozen.Units)
			}
			out = append(out, p)
		}
		d.SkipWalk()
		d.DismissReport()
		if d.over || !d.busy() && !d.s.Ready(d.side) && d.s.Ready(1-d.side) {
			return out
		}
	}
	t.Fatal("the resolution did not end")
	return nil
}

func TestDosSelectMoveConfirm(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	d.Tick()
	if !d.s.Ready(1) || d.s.Ready(0) {
		t.Fatalf("after the first frame: ready %v %v", d.s.Ready(0), d.s.Ready(1))
	}
	var at hexgrid.Coord
	for _, p := range ownUnits(d) {
		if dests, err := d.s.Destinations(0, d.own().Units[p].Index); err == nil && len(dests) > 0 {
			at = p
			break
		}
	}
	u := d.own().Units[at]
	d.Click(at)
	want, _ := d.s.Destinations(0, u.Index)
	if !d.selected || d.unit != u.Index || !maps.Equal(d.moves, want) {
		t.Fatalf("selection of %v: selected %v unit %d, %d blue hexes, want %d", at, d.selected, d.unit, len(d.moves), len(want))
	}
	var open []hexgrid.Coord
	for _, p := range dosplay.Coords(d.moves) {
		if _, ok := d.own().Buildings[p]; !ok {
			open = append(open, p)
		}
	}
	to := nearestEnemy(d, open)
	d.Click(to)
	if d.walk.Done() || d.walk.path[len(d.walk.path)-1] != to || d.walk.unit != u.Index {
		t.Fatalf("no walk to %v: %+v", to, d.walk)
	}
	if v, ok := d.own().Units[to]; !ok || v.Index != u.Index || !v.Moved || !strings.Contains(d.Message, "moved") {
		t.Fatalf("after the move: %+v, %q", v, d.Message)
	}
	// Esc finishes the walk and keeps the order.
	d.Cancel()
	if !d.walk.Done() || d.own().Units[to].Index != u.Index {
		t.Error("Esc did not finish the walk")
	}
	d.EndTurn()
	if !d.s.Ready(0) || !strings.Contains(d.Message, "confirmed") {
		t.Fatalf("F1: ready %v, %q", d.s.Ready(0), d.Message)
	}
	d.Click(ownUnits(d)[0])
	if d.selected || !strings.Contains(d.Message, "F1") {
		t.Errorf("a click while confirmed: selected %v, %q", d.selected, d.Message)
	}
	turn := d.s.Turn()
	d.Tick()
	if d.frozen == nil || d.s.Phase(0) != dosrules.PhaseAction || d.s.Phase(1) != dosrules.PhaseMovement {
		t.Fatalf("no resolution: frozen %v, phases %d %d", d.frozen != nil, d.s.Phase(0), d.s.Phase(1))
	}
	for range 100 {
		if !d.busy() {
			break
		}
		d.Tick()
	}
	if d.busy() || d.frozen != nil {
		t.Fatal("the playback did not end")
	}
	if b := d.fx.Banner(); b == nil || !strings.Contains(b.Text, "action phase") {
		t.Errorf("phase banner %+v", b)
	}
	d.Tick() // the computer confirms its movement phase
	if d.s.Ready(0) || !d.s.Ready(1) || d.s.Turn() != turn {
		t.Errorf("after the playback: ready %v %v, turn %d (was %d)", d.s.Ready(0), d.s.Ready(1), d.s.Turn(), turn)
	}
}

func TestDosConfirmAndReopen(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	d.EndTurn()
	d.EndTurn()
	if d.s.Ready(0) || !strings.Contains(d.Message, "open again") {
		t.Errorf("second F1: ready %v, %q", d.s.Ready(0), d.Message)
	}
	d.Tick()
	if d.busy() {
		t.Error("resolved while the player's orders were open")
	}
}

// TestDosBattles plays the level against a side that never moves: attack
// orders by clicks get markers, the battles show the panel and take the
// destroyed units off the frozen view, and the end blocks input.
func TestDosBattles(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	d.Tick()
	ordered, battles, destroyed := 0, 0, 0
	for range 60 {
		if d.over {
			break
		}
		if d.s.Phase(0) == dosrules.PhaseMovement {
			advance(t, d)
		} else {
			n := attackAll(t, d)
			if n > 0 && ordered == 0 {
				os := d.s.Orders(0)
				if len(os) != n || os[0].Kind != dosplay.Ordered || os[0].From == os[0].To {
					t.Fatalf("orders %+v", os)
				}
				if u := d.own().Units[os[0].From]; u.Target == nil || *u.Target != os[0].To {
					t.Errorf("the attacker shows no target: %+v", u)
				}
			}
			ordered += n
		}
		for _, p := range finish(t, d) {
			if p.ev.Kind != dosplay.Battle {
				continue
			}
			battles++
			b := p.ev.Battle
			if p.report == nil || p.report.Attacker.Name != b.Attacker.Unit.Name || p.report.Defender.Name != b.Defender.Unit.Name ||
				p.report.Defender.Before != b.Defender.Before || p.report.Defender.After != b.Defender.After || p.report.Defender.Max != b.Defender.Unit.Full {
				t.Fatalf("battle panel %+v for %+v", p.report, b)
			}
			for _, f := range []dosplay.Fighter{b.Attacker, b.Defender} {
				if !f.Destroyed {
					continue
				}
				destroyed++
				if slices.ContainsFunc(slices.Collect(maps.Values(p.frozen)), func(u dosplay.UnitView) bool { return u.Index == f.Unit.Index }) {
					t.Errorf("destroyed unit %d still in the frozen view", f.Unit.Index)
				}
			}
		}
	}
	t.Logf("turn %d: %d orders, %d battles, %d destroyed, end %+v", d.s.Turn(), ordered, battles, destroyed, d.s.Ended())
	if ordered == 0 || battles == 0 || destroyed == 0 {
		t.Fatalf("%d orders, %d battles, %d destroyed", ordered, battles, destroyed)
	}
	if !d.over {
		t.Fatalf("the battle did not end (turn %d)", d.s.Turn())
	}
	if d.s.Ended() == nil || !strings.Contains(d.Message, "VICTORY") {
		t.Errorf("end message %q", d.Message)
	}
	ready := d.s.Ready(0)
	d.Click(ownUnits(d)[0])
	d.EndTurn()
	d.Tick()
	if d.selected || d.s.Ready(0) != ready || d.busy() {
		t.Error("input after the end")
	}
}

func TestDosExitThroughContents(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	d.Tick()
	// Enter an own building, then leave it in the next movement phase.
	var building hexgrid.Coord
	var unit int
	found := false
	for _, at := range ownUnits(d) {
		d.Click(at)
		for _, p := range dosplay.Coords(d.moves) {
			if b, ok := d.own().Buildings[p]; ok && b.Owner == d.s.Faction(0) {
				building, unit, found = p, d.unit, true
				d.Click(p)
				break
			}
		}
		if found {
			break
		}
		d.Clear()
	}
	if !found {
		t.Fatal("no unit reaches an own building")
	}
	if u := d.s.Unit(unit); !u.Held || u.At != building {
		t.Fatalf("the unit did not enter: %+v, %q", u, d.Message)
	}
	finish(t, d)
	d.Click(building)
	if d.menuKind != menuContents || len(d.menu) == 0 || d.menu[0].Usable {
		t.Fatalf("contents in the action phase: %v %q", d.menuKind, rows(d.menu))
	}
	finish(t, d)
	d.Click(building)
	row := slices.IndexFunc(d.stored, func(u dosplay.UnitView) bool { return u.Index == unit })
	if d.menuKind != menuContents || row < 0 || !d.menu[row].Usable {
		t.Fatalf("contents: %v %q", d.menuKind, rows(d.menu))
	}
	d.ChooseMenu(row)
	if !d.selected || d.hold < 0 || len(d.moves) == 0 || d.MenuOpen() {
		t.Fatalf("exit choice: selected %v hold %d, %d blue hexes, %q", d.selected, d.hold, len(d.moves), d.Message)
	}
	to := dosplay.Coords(d.moves)[0]
	d.Click(to)
	if v := d.own().Units[to]; v.Index != unit || v.Held {
		t.Fatalf("after the exit %v shows %+v (%q)", to, v, d.Message)
	}
	if d.walk.Done() || d.walk.path[0] != building {
		t.Errorf("exit walk %+v", d.walk)
	}
}

func TestDosInfoAndStatus(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	lvl := d.s.Level
	g := newMapGame(lvl, false)
	g.dos = d
	at := ownUnits(d)[0]
	u := d.own().Units[at]
	g.hover, g.hoverOnMap = at, true
	lines := strings.Join(g.infoLines(), "\n")
	if !strings.Contains(lines, u.Name) || !strings.Contains(lines, fmt.Sprintf("size %d/%d", u.Size, u.Full)) {
		t.Errorf("info box for %s: %q", u.Name, lines)
	}
	if !strings.Contains(lines, lvl.Lines[at.Line][at.Col].Terrain) {
		t.Errorf("no terrain in %q", lines)
	}
	for p, b := range d.own().Buildings {
		g.hover = p
		if lines := strings.Join(g.infoLines(), "\n"); !strings.Contains(lines, b.Owner.String()) {
			t.Errorf("building at %v: %q", p, lines)
		}
		break
	}
	mine, msg := g.wellText(0)
	other, none := g.wellText(1)
	if !strings.Contains(mine, "TURN 1 ") || !strings.Contains(mine, "MOVEMENT PHASE, 0 ORDERS") || !strings.Contains(other, "ACTION PHASE") || none != "" {
		t.Errorf("status wells %q (%q), %q (%q)", mine, msg, other, none)
	}
	if g.artErr = "broken"; !strings.Contains(g.statusText(), "unit sprites: broken") {
		t.Errorf("art failure not shown: %q", g.statusText())
	}
	if shown, remembered := g.unitShown(at); !shown || remembered || g.fogOf(at) != state.Visible {
		t.Errorf("unit on %v: shown %v remembered %v", at, shown, remembered)
	}
	if g.drv() != driver(d) {
		t.Error("the game does not drive the session")
	}
}

// approachPlanner plays a computer side that moves every unit it can towards
// the enemy and attacks the first target of every unit that has one.
type approachPlanner struct{}

func (approachPlanner) Plan(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	d := &dosControl{s: s, side: side}
	var evs []dosplay.Event
	for _, at := range ownUnits(d) {
		u := d.own().Units[at]
		if s.Phase(side) == dosrules.PhaseAction {
			if ts := dosplay.Coords(s.Targets(side, u.Index)); len(ts) > 0 {
				ev, err := s.Attack(side, u.Index, ts[0])
				if err != nil {
					return evs, err
				}
				evs = append(evs, ev...)
			}
			continue
		}
		dests, err := s.Destinations(side, u.Index)
		if err != nil || len(dests) == 0 {
			continue
		}
		ev, err := s.Move(side, u.Index, nearestEnemy(d, dosplay.Coords(dests)))
		if err != nil {
			return evs, err
		}
		evs = append(evs, ev...)
	}
	return append(evs, s.Confirm(side)...), nil
}

func TestDosRevealedWalks(t *testing.T) {
	d := newDosTest(t, "level1_cp", approachPlanner{})
	d.Tick()
	finish(t, d) // the computer has planned its movement phase now
	if d.s.Phase(1) != dosrules.PhaseMovement || !d.s.Ready(1) {
		t.Fatalf("the computer did not plan: phase %d ready %v", d.s.Phase(1), d.s.Ready(1))
	}
	before := map[int]hexgrid.Coord{}
	for p, u := range d.own().Units {
		if u.Side == 1 {
			before[u.Index] = p
		}
	}
	d.fx.Dismiss() // the banner of the last phase change
	d.EndTurn()
	d.Tick()
	walks, banner := 0, false
	for range 10000 {
		if !d.busy() {
			break
		}
		if !d.walk.Done() {
			w := d.walk
			if d.frozen != nil || banner || w.path[0] != before[w.unit.(int)] {
				var ks []string
				for _, ev := range d.queue {
					ks = append(ks, ev.Kind.String()+fmt.Sprint(ev.Side, ev.Path))
				}
				t.Fatalf("walk %v of unit %v (%d before): frozen %v, banner %v %+v, stood on %v; queue %v", w.path, w.unit, walks, d.frozen != nil, banner, d.fx.Banner(), before[w.unit.(int)], ks)
			}
			// Units whose move is still to come stand where they stood.
			for _, ev := range d.queue {
				if ev.Kind == dosplay.Revealed && ev.Side == 1 {
					if u, ok := d.units(false)[ev.Path[0]]; !ok || u.Index != ev.Unit {
						t.Errorf("unit %d waiting for its walk is not on %v", ev.Unit, ev.Path[0])
					}
				}
			}
			walks++
			for !d.walk.Done() {
				d.Tick()
			}
		}
		d.DismissReport()
		d.Tick()
		if d.fx.Banner() != nil {
			banner = true
		}
	}
	if walks == 0 || !banner {
		t.Errorf("%d revealed walks, banner %v", walks, banner)
	}
	for idx, p := range before {
		if u := d.s.Unit(idx); !u.Destroyed && !u.Held && d.own().Units[u.At].Index != idx {
			t.Errorf("unit %d (was %v) not shown at %v after the reveal", idx, p, u.At)
		}
	}
}
