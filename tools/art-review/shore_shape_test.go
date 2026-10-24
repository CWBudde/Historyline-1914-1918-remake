package main

import (
	"fmt"
	"image"
	"io/fs"
	"slices"
	"sort"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// campaignShores collects the derived shore tiles the shipped and campaign
// levels select that has lacks, by ID.
func campaignShores(has func(string) bool) (map[string]mapcompose.Gap, error) {
	out := map[string]mapcompose.Gap{}
	for _, source := range []struct {
		fsys    fs.FS
		pattern string
	}{{levels.FS, "*.json"}, {levels.Campaign, "campaign/*.json"}} {
		names, err := fs.Glob(source.fsys, source.pattern)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			b, err := fs.ReadFile(source.fsys, name)
			if err != nil {
				return nil, err
			}
			l, err := level.Parse(b)
			if err != nil {
				return nil, err
			}
			s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
				h := l.Lines[c.Line][c.Col]
				return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
			}, has)
			for _, g := range s.Gaps {
				if g.Family == mapcompose.Shore {
					out[g.Descriptor] = g
				}
			}
		}
	}
	return out, nil
}

// derivedShoreShapes are the shapes the levels need beyond the roster, plus
// open water and the all-land shape (which stands in for a land neighbour in
// the edge band).
func derivedShoreShapes(t *testing.T) []mapcompose.ShoreShape {
	t.Helper()
	gaps, err := campaignShores(func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) < 40 {
		t.Fatalf("only %d derived shores found", len(gaps))
	}
	ids := make([]string, 0, len(gaps))
	for id := range gaps {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	shapes := []mapcompose.ShoreShape{{}, {Land: 0x3f}}
	for _, id := range ids {
		s, ok := mapcompose.ParseShore(id)
		if !ok {
			t.Fatalf("unparsable shore %s", id)
		}
		shapes = append(shapes, s)
	}
	for _, v := range shoreRoster() {
		s, _ := mapcompose.ParseShore(v.ID)
		shapes = append(shapes, s)
	}
	return shapes
}

func TestShoreArcs(t *testing.T) {
	for _, tt := range []struct {
		land uint8
		want [][]int
	}{
		{0x00, nil},
		{0x3f, nil},
		{0x03, [][]int{{0, 1}}},
		{0x21, [][]int{{5, 0}}},
		{0x12, [][]int{{1}, {4}}},
		{0x3b, [][]int{{3, 4, 5, 0, 1}}},
		{0x2d, [][]int{{2, 3}, {5, 0}}},
	} {
		if got := shoreArcs(tt.land); fmt.Sprint(got) != fmt.Sprint(tt.want) {
			t.Errorf("arcs of %06b: %v, want %v", tt.land, got, tt.want)
		}
	}
}

// One arc without a channel is a roster profile and keeps its exact field.
func TestShoreShapeKeepsRosterFields(t *testing.T) {
	for _, v := range shoreRoster() {
		s, _ := mapcompose.ParseShore(v.ID)
		f := newShoreShapeField(s)
		for y := -4.; y < 84; y += 3.7 {
			for x := -4.; x < 100; x += 3.3 {
				p := connectionPoint{x, y}
				if got, want := f.field(p), shoreField(v, p); got != want {
					t.Fatalf("%s at %v: %g, want %g", v.ID, p, got, want)
				}
			}
		}
	}
}

// Two tiles that agree on an edge's vertices and channel paint the same
// colours through its band, so derived shores join each other, the roster and
// open water without a seam.
func TestShoreShapesJoinEveryCompatibleNeighbour(t *testing.T) {
	path, m := shoreTestFixture(t)
	m["bank_profile"] = authoredBankProfile
	shoreTestWrite(t, path, m)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	shapes := derivedShoreShapes(t)
	fields := make([]shoreShapeField, len(shapes))
	for i, s := range shapes {
		fields[i] = newShoreShapeField(s)
	}
	vertices := sharedConnectionGeometry().Polygon
	neighbors := []connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	bit := func(m uint8, i int) bool { return m&(1<<i) != 0 }
	pairs := 0
	for i, a := range shapes {
		for j, b := range shapes {
			for direction := 1; direction <= 6; direction++ {
				opposite := (direction+2)%6 + 1
				e, o := direction%6, opposite%6
				if bit(a.Land, e) != bit(b.Land, (o+1)%6) || bit(a.Land, (e+1)%6) != bit(b.Land, o) || bit(a.Channels, e) != bit(b.Channels, o) {
					continue
				}
				pairs++
				p0, p1 := vertices[e], vertices[(e+1)%6]
				edge := p1.sub(p0)
				tangent := edge.mul(1 / edge.length())
				normal := connectionPoint{-tangent[1], tangent[0]}
				for _, depth := range []float64{-8, -4, 0, 4, 8} {
					for _, along := range []float64{.02, .25, .4, .5, .6, .75, .98} {
						p := p0.add(edge.mul(along)).add(normal.mul(depth))
						q := p.sub(neighbors[direction-1])
						ca := shoreMaterialColor(in, p, fields[i].field(p))
						cb := shoreMaterialColor(in, q, fields[j].field(q))
						if ca != cb {
							t.Fatalf("%s/%s dir%d depth%g t%g: %v/%v", a.ID(), b.ID(), direction, depth, along, ca, cb)
						}
					}
				}
			}
		}
	}
	if pairs < 1000 {
		t.Fatalf("only %d compatible pairs", pairs)
	}
}

// Every derived tile keeps one body of water that reaches each of its water
// openings (open and mixed edges, channels); a pond holds the centre.
func TestShoreShapesKeepOneConnectedWater(t *testing.T) {
	g := sharedConnectionGeometry()
	for _, s := range derivedShoreShapes(t) {
		if s.Land == 0 {
			continue
		}
		f := newShoreShapeField(s)
		const w, h = 96, 80
		label := make([]int, w*h)
		water := func(x, y int) bool {
			p := connectionPoint{float64(x) + .5, float64(y) + .5}
			return connectionInside(g, p) && f.field(p) < 0
		}
		regions := 0
		for y := range h {
			for x := range w {
				if label[y*w+x] != 0 || !water(x, y) {
					continue
				}
				regions++
				stack := [][2]int{{x, y}}
				label[y*w+x] = regions
				for len(stack) > 0 {
					c := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
						nx, ny := c[0]+d[0], c[1]+d[1]
						if nx >= 0 && ny >= 0 && nx < w && ny < h && label[ny*w+nx] == 0 && water(nx, ny) {
							label[ny*w+nx] = regions
							stack = append(stack, [2]int{nx, ny})
						}
					}
				}
			}
		}
		// Tiny specks at a corner of a single water vertex do not count.
		sizes := map[int]int{}
		for _, l := range label {
			if l != 0 {
				sizes[l]++
			}
		}
		big := 0
		for _, n := range sizes {
			if n >= 12 {
				big++
			}
		}
		if big != 1 {
			t.Errorf("%s: %d bodies of water (sizes %v)", s.ID(), big, sizes)
			continue
		}
		var openings []connectionPoint
		for e := range 6 {
			a, b := g.Polygon[e], g.Polygon[(e+1)%6]
			landA, landB := s.Land&(1<<e) != 0, s.Land&(1<<((e+1)%6)) != 0
			mid := a.add(b).mul(.5)
			inward := connectionPoint{48, 40}.sub(mid)
			inward = inward.mul(3 / inward.length())
			switch {
			case s.Channels&(1<<e) != 0, !landA && !landB:
				openings = append(openings, mid.add(inward))
			case landA != landB:
				// The coast crosses a mixed edge at its midpoint: water lies
				// toward the water vertex.
				toward := a
				if landA {
					toward = b
				}
				openings = append(openings, mid.add(toward.sub(mid).mul(.3)).add(inward))
			}
		}
		if s.Land == 0x3f && s.Channels == 0 {
			openings = append(openings, connectionPoint{48, 40})
		}
		var seen []int
		for _, p := range openings {
			l := label[int(p[1])*w+int(p[0])]
			if l == 0 || sizes[l] < 12 {
				t.Errorf("%s: opening %v is not in the water", s.ID(), p)
				continue
			}
			if !slices.Contains(seen, l) {
				seen = append(seen, l)
			}
		}
		if len(seen) > 1 {
			t.Errorf("%s: openings reach %d bodies of water", s.ID(), len(seen))
		}
	}
}

// A derived tile is a whole hex at both densities, opaque inside and clear
// outside, with water at a pond's centre and meadow at its corners.
func TestShoreShapeTilesFillTheHex(t *testing.T) {
	path, m := shoreTestFixture(t)
	m["bank_profile"] = authoredBankProfile
	shoreTestWrite(t, path, m)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := exportShoreShapeTiles(in, mapcompose.ShoreShape{Land: 0x3f})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{2, 4} {
		im := tiles[d]
		if im.Bounds() != image.Rect(0, 0, 96*d, 80*d) {
			t.Fatalf("%dx bounds %v", d, im.Bounds())
		}
		for y := range 80 * d {
			for x := range 96 * d {
				inside := connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)})
				if a := im.NRGBAAt(x, y).A; inside != (a == 255) || !inside && a != 0 {
					t.Fatalf("%dx alpha %d at %d,%d (inside %v)", d, a, x, y, inside)
				}
			}
		}
		if centre, corner := im.NRGBAAt(48*d, 40*d), im.NRGBAAt(4*d, 40*d); centre.B <= centre.G || corner.G <= corner.B {
			t.Errorf("%dx pond centre %v, corner %v", d, centre, corner)
		}
	}
}
