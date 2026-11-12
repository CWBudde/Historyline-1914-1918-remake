package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scores"
	"github.com/hajimehoshi/ebiten/v2"
)

// The tests below open the unit info and the game info in the human side's
// panel through the info gesture and close them again.

// emptyCell is a cell of the human panel without a unit or a building.
func emptyCell(t *testing.T, g *game, d *dosControl) hexgrid.Coord {
	t.Helper()
	for _, c := range g.splitView(d.side).Cells() {
		_, unit := d.units(false)[c]
		_, building := d.buildings()[c]
		if g.grid.Contains(c) && !unit && !building {
			return c
		}
	}
	t.Fatal("no empty cell in view")
	return hexgrid.Coord{}
}

// enemyUnit is a unit of the other side on the human side's map.
func enemyUnit(t *testing.T, d *dosControl) hexgrid.Coord {
	t.Helper()
	for _, c := range dosplay.Coords(toSet(d.units(false))) {
		if d.units(false)[c].Side != d.side {
			return c
		}
	}
	t.Fatal("no enemy unit on the map")
	return hexgrid.Coord{}
}

// infoDown runs the info gesture (fire held, down) on c with the keys.
func infoDown(g *game, c hexgrid.Coord) {
	g.setCursor(c)
	feed(g, []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{Fire: true, Down: true}), held(gesture.Keys{})})
}

func TestDosOverlaysInfoGesture(t *testing.T) {
	tests := []struct {
		name string
		at   func(t *testing.T, g *game, d *dosControl, unit hexgrid.Coord) hexgrid.Coord
		mode panelMode
	}{
		{"own unit", func(_ *testing.T, _ *game, _ *dosControl, unit hexgrid.Coord) hexgrid.Coord { return unit }, panelUnitInfo},
		{"own HQ", func(t *testing.T, _ *game, d *dosControl, _ hexgrid.Coord) hexgrid.Coord { return ownHQ(t, d) }, panelGameInfo},
		{"empty cell", func(t *testing.T, g *game, d *dosControl, _ hexgrid.Coord) hexgrid.Coord { return emptyCell(t, g, d) }, panelGameInfo},
		{"enemy unit", func(t *testing.T, _ *game, d *dosControl, _ hexgrid.Coord) hexgrid.Coord { return enemyUnit(t, d) }, panelMap},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, d, unit := newCursorTest(t)
			at := tt.at(t, g, d, unit)
			infoDown(g, at)
			if m := g.split.mode[d.side]; m != tt.mode || g.split.mode[1-d.side] != panelMap {
				t.Fatalf("modes %v, want %v in the human panel (%q)", g.split.mode, tt.mode, d.Message)
			}
			if tt.mode == panelMap && !strings.Contains(d.Message, "enemy") {
				t.Errorf("no refusal for an enemy unit: %q", d.Message)
			}
			if tt.mode == panelUnitInfo && g.split.infoUnit.Index != d.units(false)[at].Index {
				t.Errorf("unit info of %+v", g.split.infoUnit)
			}
		})
	}
}

func TestDosOverlaysClose(t *testing.T) {
	for name, close := range map[string]func(g *game){
		"fire click": func(g *game) { feed(g, []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{})}) },
		"cancel direction": func(g *game) {
			feed(g, []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{Fire: true, Right: true}), held(gesture.Keys{})})
		},
		"mouse click": func(g *game) {
			feed(g, []gesture.Input{mouseAt(5, 5, false), mouseAt(5, 5, true), mouseAt(5, 5, false)})
		},
		"Esc": func(g *game) { g.applyAction(actCancel) },
	} {
		t.Run(name, func(t *testing.T) {
			g, d, unit := newCursorTest(t)
			for _, at := range []hexgrid.Coord{unit, ownHQ(t, d)} {
				infoDown(g, at)
				if g.split.mode[d.side] == panelMap {
					t.Fatalf("no info on %v", at)
				}
				close(g)
				if g.split.mode[d.side] != panelMap || d.selected {
					t.Errorf("after %s on %v: mode %v, selected %v", name, at, g.split.mode[d.side], d.selected)
				}
			}
		})
	}
}

func TestDosOverlaysUnitInfo(t *testing.T) {
	g, d, at := newCursorTest(t)
	u := d.units(false)[at]
	tv, ok := d.s.TypeInfo(d.side, u.Type)
	if !ok {
		t.Fatalf("no type %d", u.Type)
	}
	got := g.unitInfo(u)
	want := dospanel.UnitInfo{
		Name: tv.Name, Air: dospanel.Weapon(tv.Air), Ground: dospanel.Weapon(tv.Ground), Water: dospanel.Weapon(tv.Water),
		Armour: tv.Armour, Speed: tv.Speed, Weight: tv.Weight, Group: tv.Group,
	}
	if got != want {
		t.Errorf("unit info %+v, want %+v", got, want)
	}
	if got.Name == "" || got.Group == 0 || got.Speed == 0 {
		t.Errorf("empty unit info %+v", got)
	}
}

func TestDosOverlaysGameInfo(t *testing.T) {
	g, d, _ := newCursorTest(t)
	g.scores = scores.OpenFile(filepath.Join(t.TempDir(), "scores.json"))
	if _, err := g.scores.Record("PULSE", 777); err != nil {
		t.Fatal(err)
	}
	st := d.s.Stats()
	gi := g.gameInfo()
	if gi.Level != "PULSE" || gi.Best != 777 || gi.Round != st.Round || gi.Score != st.Score || gi.Moves != st.Orders[d.side] || gi.Limit != st.Limit {
		t.Errorf("game info %+v from %+v", gi, st)
	}
	if gi.Mode != modeWord(st.Mode[d.side]) || gi.Energy != st.Energy || gi.Income != st.Income {
		t.Errorf("mode, energy or income %+v from %+v", gi, st)
	}
	for k := range 2 {
		if gi.Sides[k] != dospanel.Holdings(st.Sides[k]) {
			t.Errorf("side %d holdings %+v, want %+v", k, gi.Sides[k], st.Sides[k])
		}
	}
	if gi.Map != dospanel.Holdings(st.Map) || gi.Sides[0].Units == 0 {
		t.Errorf("map holdings %+v from %+v", gi.Map, st.Map)
	}
	// An order counts as a move; no store reads as no best.
	g.scores = nil
	if gi := g.gameInfo(); gi.Best != 0 {
		t.Errorf("best without a store: %d", gi.Best)
	}
	if levelKey(d.s.Level) != "PULSE" {
		t.Errorf("level key %q", levelKey(d.s.Level))
	}
}

func TestDosOverlaysRecordWin(t *testing.T) {
	for _, tt := range []struct {
		name   string
		winner int
		best   int
	}{{"win", 0, 321}, {"defeat", 1, 0}} {
		t.Run(tt.name, func(t *testing.T) {
			g, d := newSplitTest(t, passPlanner{}, 1280, 720)
			store := scores.OpenFile(filepath.Join(t.TempDir(), "scores.json"))
			a := &app{shell: testShell(t), base: matchConfig{scores: store}, battle: g, press: -1}
			d.s.Game.Ended = &dosrules.End{Over: true, Winner: tt.winner, Score: 321}
			d.over = true
			a.checkEnd()
			a.checkEnd() // announced once, recorded once
			if got := store.Best(levelKey(d.s.Level)); got != tt.best {
				t.Errorf("best %d, want %d", got, tt.best)
			}
		})
	}
}

func TestDosOverlaysHUDInfo(t *testing.T) {
	g, d, at := newCursorTest(t)
	g.setCursor(at)
	if lines := g.infoLines(); len(lines) != 0 {
		t.Errorf("the battle screen shows the info box: %q", lines)
	}
	g.hexNumbers = true
	if lines := strings.Join(g.infoLines(), "\n"); !strings.Contains(lines, "hex ") || !strings.Contains(lines, d.own().Units[at].Name) {
		t.Errorf("dev info box %q", lines)
	}
}

func TestDosOverlaysDraw(t *testing.T) {
	g, d, at := newCursorTest(t)
	img := ebiten.NewImage(1800, 1000)
	for _, c := range []hexgrid.Coord{at, ownHQ(t, d)} {
		infoDown(g, c)
		g.drawSplit(img)
		g.drawWells(img)
		g.applyAction(actCancel)
	}
	// The other panel keeps its map while the human one shows the info.
	infoDown(g, at)
	if g.split.mode[1-d.side] != panelMap {
		t.Errorf("other panel %v", g.split.mode[1-d.side])
	}
	// The phase change closes the plates.
	d.EndTurn()
	g.updateSplit()
	if g.split.mode[d.side] != panelMap {
		t.Errorf("the prompt left the plate open: %v", g.split.mode[d.side])
	}
}

func TestDosOverlaysText(t *testing.T) {
	tx := panelText{k: 2}
	w, h := tx.Measure("AB", dospanel.Small)
	if w2, h2 := tx.Measure("AB", dospanel.Large); w <= 0 || h <= 0 || h2 <= h || w2 <= w {
		t.Errorf("small %dx%d, large %dx%d", w, h, w2, h2)
	}
	for _, tt := range []struct {
		scale float64
		k     int
	}{{0.5, 1}, {1, 1}, {2, 2}, {2.9, 2}, {4, 4}} {
		if k := textScale(tt.scale); k != tt.k {
			t.Errorf("text scale at %v: %d, want %d", tt.scale, k, tt.k)
		}
	}
}

func TestDosOverlaysCaptureViews(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1920, 1080)
	planned := map[string]bool{}
	for _, f := range splitFramePlan() {
		planned[f.view] = true
	}
	for view, mode := range map[string]panelMode{"split-unitinfo": panelUnitInfo, "split-gameinfo": panelGameInfo, "split-contents": panelContents} {
		if !planned[view] {
			t.Errorf("%s is not captured", view)
		}
		g.stageCapture(view)
		if g.split.mode[d.side] != mode || g.split.mode[1-d.side] != panelMap {
			t.Errorf("%s: modes %v", view, g.split.mode)
		}
	}
	g.stageCapture("split")
	if g.split.mode[d.side] != panelMap || d.box != nil {
		t.Errorf("the opening keeps a plate: %v %+v", g.split.mode, d.box)
	}
}
