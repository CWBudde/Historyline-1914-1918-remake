package mapcompose

import (
	"encoding/json"
	"image"
	"image/color"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// fieldExemplarImage is a tile of a material with a checker of hollows.
func fieldExemplarImage(d int) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			v := uint8(100 + 60*((x/(7*d)+y/(5*d))%2))
			im.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	return im
}

func allMud(lines, width int, terrain string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	g := hexgrid.Grid{Width: width, Lines: lines}
	return g, func(hexgrid.Coord) Hex { return Hex{Terrain: terrain} }
}

func TestCraterHexesCarryTheirField(t *testing.T) {
	g := hexgrid.Grid{Width: 4, Lines: 7}
	centre := hexgrid.Coord{Line: 3, Col: 1}
	field := map[hexgrid.Coord]bool{centre: true}
	around := hexgrid.Around(centre)
	for _, n := range around[:3] {
		field[n] = true
	}
	at := func(c hexgrid.Coord) Hex {
		if field[c] {
			return Hex{Terrain: "mud"}
		}
		return Hex{Terrain: "grass"}
	}
	exemplar := fieldExemplar("mud")
	s := Compose(g, at, func(id string) bool { return id == exemplar })
	bg := s.Cells[centre.Line][centre.Col].Background
	f := s.FieldAt(centre, bg)
	if f == nil || f.Exemplar != exemplar {
		t.Fatalf("field %+v", f)
	}
	if want := [6]bool{false, false, false, true, true, true}; f.Edges != want {
		t.Errorf("edges %v, want %v", f.Edges, want)
	}
	if s.FieldAt(centre, "tile/other") != nil {
		t.Error("a different background is reworked")
	}
	if s.FieldAt(hexgrid.Coord{}, s.Cells[0][0].Background) != nil {
		t.Error("meadow is reworked")
	}
	if none := Compose(g, at, func(string) bool { return false }); none.FieldAt(centre, none.Cells[centre.Line][centre.Col].Background) != nil {
		t.Error("a catalog without the exemplar is reworked")
	}
	raw, err := json.Marshal(s.Cells[centre.Line][centre.Col])
	if err != nil {
		t.Fatal(err)
	}
	var cell map[string]any
	if err := json.Unmarshal(raw, &cell); err != nil {
		t.Fatal(err)
	}
	if _, ok := cell["Field"]; ok {
		t.Error("the scene exports its draw-time rework")
	}
}

func TestStampedCraterIsNoField(t *testing.T) {
	g, at := allMud(1, 1, "shelled_w")
	exemplar := fieldExemplar("shelled_w")
	stampID := StampID("crater_w", "1", 1)
	s := Compose(g, at, func(id string) bool { return id == exemplar || id == stampID })
	c := hexgrid.Coord{}
	under, stamp, _, _ := s.Layers(c, func(*Cluster) bool { return true })
	if stamp != stampID || s.FieldAt(c, under) != nil {
		t.Errorf("stamp %q, under %q reworked as a field", stamp, under)
	}
	bg, _, _, _ := s.Layers(c, func(*Cluster) bool { return false })
	if s.FieldAt(c, bg) == nil {
		t.Error("the stamp's per-hex art under fog keeps no field")
	}
}

func TestFieldReworkRepeatsOnlyAtItsPeriod(t *testing.T) {
	const d = 2
	tile := fieldExemplarImage(d)
	m := NewFieldMaterial(tile, [2]float64{48, 40}, d)
	render := func(c hexgrid.Coord) *image.NRGBA {
		x, y := hexgrid.Center(c)
		return m.Rework(tile, image.Pt(int(x-48)*d, int(y-40)*d), c, [6]bool{})
	}
	diff := func(a, b *image.NRGBA) int {
		n := 0
		for i := range a.Pix {
			if a.Pix[i] != b.Pix[i] {
				n++
			}
		}
		return n
	}
	base := hexgrid.Coord{Line: 4, Col: 3}
	first := render(base)
	for _, c := range []hexgrid.Coord{{Line: 4, Col: 4}, {Line: 5, Col: 3}, {Line: 6, Col: 3}} {
		if diff(first, render(c)) < len(first.Pix)/4 {
			t.Errorf("%v repeats %v", c, base)
		}
	}
	for _, c := range []hexgrid.Coord{{Line: 4, Col: 3 + FieldPeriodX/hexgrid.ColPitch}, {Line: 4 + FieldPeriodY/hexgrid.LinePitch, Col: 3}} {
		if n := diff(first, render(c)); n != 0 {
			t.Errorf("%v differs from %v one period away in %d bytes", c, base, n)
		}
		x, y := hexgrid.Center(c)
		bx, by := hexgrid.Center(base)
		if a, b := FieldPhase(image.Pt(int(x-48)*d, int(y-40)*d), d), FieldPhase(image.Pt(int(bx-48)*d, int(by-40)*d), d); a != b {
			t.Errorf("phase %v, %v", a, b)
		}
	}
}

func TestFieldReworkKeepsTheTileAtOtherGround(t *testing.T) {
	const d = 2
	tile := fieldExemplarImage(d)
	// A hole in the mask stays a hole.
	tile.SetNRGBA(0, 0, color.NRGBA{})
	m := NewFieldMaterial(fieldExemplarImage(d), [2]float64{48, 40}, d)
	c := hexgrid.Coord{Line: 4, Col: 3}
	x, y := hexgrid.Center(c)
	out := m.Rework(tile, image.Pt(int(x-48)*d, int(y-40)*d), c, [6]bool{5: true}) // the north side
	if out.NRGBAAt(0, 0).A != 0 {
		t.Error("the tile's alpha is lost")
	}
	if got, want := out.NRGBAAt(48*d, 2*d), tile.NRGBAAt(48*d, 2*d); got != want {
		t.Errorf("at the north side %v, want the tile's %v", got, want)
	}
	changed := false
	for dy := range 10 {
		changed = changed || out.NRGBAAt(48*d, (60+dy)*d) != tile.NRGBAAt(48*d, (60+dy)*d)
	}
	if !changed {
		t.Error("the field does not reach the far side")
	}
}

func TestStampBorderWithRoughGroundIsReworked(t *testing.T) {
	g := hexgrid.Grid{Width: 4, Lines: 7}
	stamped := hexgrid.Coord{Line: 3, Col: 1}
	around := hexgrid.Around(stamped)
	rough := around[0] // north-east
	at := func(c hexgrid.Coord) Hex {
		switch c {
		case stamped:
			return Hex{Terrain: "shelled_w"}
		case rough:
			return Hex{Terrain: "rough_w"}
		}
		return Hex{Terrain: "grass_w"}
	}
	ground := fieldExemplar("rough_w")
	stampID := StampID("crater_w", "1", 1)
	s := Compose(g, at, func(id string) bool { return id == ground || id == stampID })
	yes := func(*Cluster) bool { return true }
	under, _, _, _ := s.Layers(stamped, yes)
	f := s.FringeAt(stamped, under)
	if f == nil || f.Exemplar != ground || !f.Border[0] || f.Border[1] || f.Other[0] || !f.Other[1] {
		t.Fatalf("stamped hex: %+v", f)
	}
	bg, _, _, _ := s.Layers(rough, yes)
	r := s.FringeAt(rough, bg)
	if r == nil || !r.Border[3] || r.Other[3] { // the south-west side faces the stamp
		t.Fatalf("rough hex: %+v", r)
	}
	if other, _, _, _ := s.Layers(around[2], yes); s.FringeAt(around[2], other) != nil {
		t.Error("snow beside the stamp is reworked")
	}
	perHex, _, _, _ := s.Layers(stamped, func(*Cluster) bool { return false })
	if s.FringeAt(stamped, perHex) != nil {
		t.Error("the stamp's per-hex art under fog is reworked as its under-tile")
	}
}

func TestStampFringeWeight(t *testing.T) {
	f := StampFringe{}
	f.Border[0] = true // north-east
	f.Other[1] = true  // south-east
	cx, cy := 500., 500.
	mid := func(a, b [2]float64) (float64, float64) { return (a[0] + b[0]) / 2, (a[1] + b[1]) / 2 }
	x, y := mid(hexVertices[1], hexVertices[2]) // on the north-east side
	if w := f.weight(x, y, cx+x, cy+y); w < .9 {
		t.Errorf("on the border %.2f, want about 1", w)
	}
	if w := f.weight(-20, 20, cx-20, cy+20); w != 0 {
		t.Errorf("far from the border %.2f, want 0", w)
	}
	// The corner the border shares with other ground takes none: the tile beyond
	// is snow.
	if w := f.weight(hexVertices[2][0], hexVertices[2][1], cx+hexVertices[2][0], cy+hexVertices[2][1]); w > .01 {
		t.Errorf("at the shared corner %.2f, want 0", w)
	}
}
