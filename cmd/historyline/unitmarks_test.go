package main

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestUnitMarks(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1920, 1080)
	var all []hexgrid.Coord
	for l, line := range g.lvl.Lines {
		for c := range line {
			all = append(all, hexgrid.Coord{Line: l, Col: c})
		}
	}
	// A veteran of the human side: its badge and size show on the map.
	v := d.s.View(d.side)
	at := ownUnits(d)[0]
	u := v.Units[at]
	u.Exp, u.Size = 3, 4
	v.Units[at] = u
	d.frozen = &v

	marks := g.unitMarks(all)
	sides := map[int]int{}
	var vet *unitMark
	for i, m := range marks {
		sides[m.unit.Side]++
		if m.at == at {
			vet = &marks[i]
		}
	}
	if len(marks) != len(v.Units) || sides[d.side] == 0 || sides[1-d.side] == 0 {
		t.Fatalf("%d marks for %d units, by side %v; want every unit of both sides", len(marks), len(v.Units), sides)
	}
	if vet == nil || vet.unit.Exp != 3 || vet.unit.Size != 4 {
		t.Fatalf("veteran's mark %+v", vet)
	}
	sx, sy := g.cam.WorldToScreen(hexgrid.Center(at))
	if vet.badgeX != sx || vet.badgeY >= sy || vet.sizeX <= sx || vet.sizeY <= sy {
		t.Errorf("badge at %.0f,%.0f and size at %.0f,%.0f around the hex centre %.0f,%.0f; want the badge above, the size lower right",
			vet.badgeX, vet.badgeY, vet.sizeX, vet.sizeY, sx, sy)
	}
	// Both stay inside the hex.
	e := g.cam.Effective()
	if dx, dy := vet.sizeX-sx, sy-vet.badgeY; dx > hexgrid.HexWidth/2*e || dy > hexgrid.HexHeight/2*e {
		t.Errorf("markers leave the hex: %.0f right, %.0f up at scale %.2f", dx, dy, e)
	}

	// A walking unit carries none; nor do hexes too small to read.
	d.walk = newWalkMotion([]hexgrid.Coord{{Line: at.Line, Col: at.Col + 1}, at}, u.Index, false)
	for _, m := range g.unitMarks(all) {
		if m.at == at {
			t.Error("the walking unit keeps its markers")
		}
	}
	d.walk = nil
	if marksShown(marksMinScale-0.01) || !marksShown(marksMinScale) {
		t.Error("marksShown threshold")
	}
	// Every zoom level of a small window still shows them.
	small, _ := newSplitTest(t, passPlanner{}, 1024, 768)
	small.zoomSplit(-len(dosscreen.Zooms))
	if !marksShown(small.cam.Effective()) {
		t.Errorf("the farthest zoom at 1024×768 (scale %.2f) hides the markers", small.cam.Effective())
	}

	// Drawing them works without art.
	g.drawUnitMarks(ebiten.NewImage(1920, 1080), all)
	for exp := range 8 {
		drawBrass(ebiten.NewImage(40, 40), exp, 20, 20, 30)
	}
}

func TestWellBadge(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1920, 1080)
	sc := g.split.screen
	r := sc.WellText(d.side)
	for _, exp := range []int{0, 1, 6} {
		var calls []image.Point
		w := well{unit: true, size: 6, exp: exp, name: "CAVALRY", progress: -1}
		drawWell(ebiten.NewImage(1920, 1080), sc, d.side, w, func(_ *ebiten.Image, e int, cx, cy, size float64) {
			if e != exp {
				t.Errorf("badge of exp %d, want %d", e, exp)
			}
			if cx-size/2 < float64(r.X) || cy-size/2 < float64(r.Y)-1 || cy+size/2 > float64(r.Y+r.H)+1 {
				t.Errorf("badge %.0f×%.0f at %.0f,%.0f leaves the well text %v", size, size, cx, cy, r)
			}
			calls = append(calls, image.Pt(int(cx), int(cy)))
		})
		if want := min(exp, 1); len(calls) != want {
			t.Errorf("exp %d: %d badges, want %d", exp, len(calls), want)
		}
	}
}
