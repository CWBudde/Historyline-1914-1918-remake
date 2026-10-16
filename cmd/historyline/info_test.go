package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func testGame(t *testing.T) *game {
	t.Helper()
	c := newTestControl(t)
	g := newGame(testLevel(t), c.st, c.player, false)
	g.ctl = c
	return g
}

// hexWith finds a hex matching ok.
func hexWith(t *testing.T, g *game, ok func(hexgrid.Coord, *state.Hex) bool) hexgrid.Coord {
	t.Helper()
	for l, line := range g.state.Hexes {
		for col := range line {
			c := hexgrid.Coord{Line: l, Col: col}
			if ok(c, g.state.At(c)) {
				return c
			}
		}
	}
	t.Fatal("no such hex in level1_cp")
	return hexgrid.Coord{}
}

func TestStatusLineKeepsUnseenTerrainHidden(t *testing.T) {
	g := testGame(t)
	base := g.statusText() // nothing hovered
	g.hover = hexWith(t, g, func(c hexgrid.Coord, h *state.Hex) bool {
		return g.fogOf(c) == state.Unseen && !strings.Contains(base, h.Terrain)
	})
	g.hoverOnMap = true
	h := g.state.At(g.hover)
	if s := g.statusText(); strings.Contains(s, h.Terrain) {
		t.Errorf("status line names unseen terrain %q: %q", h.Terrain, s)
	}
	if lines := g.infoLines(); !slices.Contains(lines, "unexplored") || slices.Contains(lines, h.Terrain) {
		t.Errorf("info box on an unseen hex: %q", lines)
	}
}

func TestInfoVisibleHexAndUnits(t *testing.T) {
	g := testGame(t)
	at := ownMovers(g.ctl)[0]
	u := g.state.At(at).Unit
	g.hover, g.hoverOnMap = at, true
	lines := g.infoLines()
	for _, want := range []string{g.state.At(at).Terrain, u.Def.Name, "size 6/6"} {
		if !slices.ContainsFunc(lines, func(l string) bool { return strings.Contains(l, want) }) {
			t.Errorf("info lacks %q: %q", want, lines)
		}
	}
	// Selecting the unit shows it once, above the hex.
	g.ctl.Select(at)
	lines = g.infoLines()
	if lines[0] != "selected:" || strings.Count(strings.Join(lines, "\n"), u.Def.Name+" (") != 1 {
		t.Errorf("selected unit shown twice or not first: %q", lines)
	}
}

func TestInfoSeenHexShowsOnlyMemory(t *testing.T) {
	g := testGame(t)
	enemy := hexWith(t, g, func(c hexgrid.Coord, h *state.Hex) bool {
		return h.Unit != nil && h.Unit.Owner != g.player && g.fogOf(c) == state.Unseen
	})
	g.hover, g.hoverOnMap = enemy, true
	if lines := g.infoLines(); strings.Contains(strings.Join(lines, "\n"), g.state.At(enemy).Unit.Def.Name) {
		t.Errorf("hidden enemy described: %q", lines)
	}
	// A scout next to it sees it, then leaves: the hex is Seen.
	spot := hexWith(t, g, func(c hexgrid.Coord, h *state.Hex) bool {
		return hexgrid.Distance(c, enemy) == 1 && h.Unit == nil && h.Container == nil
	})
	scout := g.state.At(ownMovers(g.ctl)[0]).Unit
	g.state.At(spot).Unit = scout
	g.state.UpdateVision(g.player)
	g.state.At(spot).Unit = nil
	g.state.UpdateVision(g.player)
	if g.fogOf(enemy) != state.Seen {
		t.Skipf("enemy hex is %v; another own unit sees it", g.fogOf(enemy))
	}
	g.state.At(enemy).Unit.Size = 3 // changes after the player looked away
	lines := strings.Join(g.infoLines(), "\n")
	if !strings.Contains(lines, "last seen") || !strings.Contains(lines, "size 6") || strings.Contains(lines, "fuel") {
		t.Errorf("seen enemy: %q", lines)
	}
}

func TestUnitInfoWeapons(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	u, err := state.NewUnit(set, "CP_infantry")
	if err != nil {
		t.Fatal(err)
	}
	zero := 0
	u.Ammo[2] = &zero
	u.Weapons[3] = nil
	lines := unitInfo(u)
	if len(lines) != 2+3 {
		t.Fatalf("want 3 weapon rows, empty slot skipped: %q", lines)
	}
	if !strings.HasSuffix(lines[4], "ammo 0") {
		t.Errorf("zero ammo not shown: %q", lines[4])
	}
	for _, l := range lines[2:] {
		if strings.HasPrefix(l, "4 ") {
			t.Errorf("empty slot listed: %q", l)
		}
	}
	u.Moved = true
	if got := unitInfo(u); got[2] != "moved" {
		t.Errorf("moved state: %q", got)
	}
}

func TestTerrainInfoModifiers(t *testing.T) {
	two := 2
	h := &state.Hex{Terrain: "grass", Infra: []string{"road14"}, Mods: data.Hex{Movement: &two}}
	got := terrainInfo(h, state.Seen, state.None)
	want := []string{"grass", "road14", "move 2  battle -  sight -"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	h.Container = &state.Container{}
	if got := terrainInfo(h, state.Visible, state.Entente); got[len(got)-1] != "owner Entente Cordial" {
		t.Errorf("building owner: %q", got)
	}
}

func TestInfoRectFitsScreen(t *testing.T) {
	l := layoutHUD(320, 240)
	r := l.infoRect(1000, 4)
	if r.X < hudMargin || r.X+r.W > l.DevRight || r.Y != infoTop || r.H != 4*infoLineHeight+infoPad {
		t.Errorf("info rect %+v on a 320 px screen", r)
	}
}
