package main

import (
	"image"
	"math"
	"path/filepath"
	"slices"
	"testing"
)

// compoundEntryOffset is how far (world px, across the entry port) the
// painted street of scene s runs from the road tile it continues (extended
// straight through the cell) between depths near and far inside the port: the
// street's mirror axis in the artwork minus the road's, searched within
// ±shift world px over half world px either side.
func compoundEntryOffset(in compoundInputs, s compoundSceneInput, near, far float64, half, shift int) (float64, error) {
	l, err := compoundLayoutFor(s.ID)
	if err != nil {
		return 0, err
	}
	art, _, err := compoundAffine(in.Artwork[s.ID].Image, l, s.Inset, s.fit())
	if err != nil {
		return 0, err
	}
	road := in.EntryRoads[s.ID]
	origin, port, err := compoundEntryPort(l)
	if err != nil {
		return 0, err
	}
	f, v, err := compoundEntryRoad(road, l)
	if err != nil {
		return 0, err
	}
	raw := renderConnectionRaw(road, &f, &v) // 4x, 4 world px overscan
	g := road.manifest.Geometry
	at := g.Ports[port-1]
	inward := g.Neighbors[port-1].mul(-1 / g.Neighbors[port-1].length())
	across := connectionPoint{-inward[1], inward[0]}
	const reach, per = 24, 4 // world px either side, samples per world px
	// profile is the median colour, over depths near..far, at each
	// sample across the port of im, whose pixel (0,0) lies at local world
	// point o. The median ignores what crosses the street there (the depot's
	// barrier, the factory's gantry).
	profile := func(im *image.NRGBA, o connectionPoint) [][3]float64 {
		var p [][3]float64
		for i := -reach * per; i <= reach*per; i++ {
			var ch [3][]float64
			for d := near; d <= far; d += .5 {
				q := at.add(inward.mul(d)).add(across.mul(float64(i) / per))
				x, y := int(math.Floor((q[0]-o[0])*4)), int(math.Floor((q[1]-o[1])*4))
				if !(image.Point{x, y}).In(im.Bounds()) {
					continue
				}
				c := im.NRGBAAt(x, y)
				ch[0], ch[1], ch[2] = append(ch[0], float64(c.R)), append(ch[1], float64(c.G)), append(ch[2], float64(c.B))
			}
			var m [3]float64
			for k := range 3 {
				slices.Sort(ch[k])
				m[k] = ch[k][len(ch[k])/2]
			}
			p = append(p, m)
		}
		return p
	}
	want := profile(raw, connectionPoint{-4, -4})
	got := profile(art, connectionPoint{-origin[0], -origin[1]})
	// axis is the line across the port about which p is most nearly mirror
	// symmetric over half world px either side, within ±shift world px of the
	// port's centre line.
	axis := func(p [][3]float64) float64 {
		best, at := math.Inf(1), 0.
		for c := -shift * per; c <= shift*per; c++ {
			sum := 0.
			for t := 1; t <= half*per; t++ {
				a, b := p[reach*per+c+t], p[reach*per+c-t]
				sum += math.Sqrt((a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1]) + (a[2]-b[2])*(a[2]-b[2]))
			}
			if sum < best {
				best, at = sum, float64(c)/per
			}
		}
		return at
	}
	return axis(got) - axis(want), nil
}

// compoundEntryDeepBands is, per layout, the deepest 8 world px of the entry
// street before its first junction, bend or end (depths far-8..far inside the
// port), and the half width of its mirror window: the street with its verges,
// or only the street and its kerbs where the verges differ (snow beside the
// factory's tank row, a pipe on its north side).
var compoundEntryDeepBands = map[string]struct {
	far  float64
	half int
}{
	"hq":      {58, 14}, // the gatehouse roof
	"depot":   {76, 14}, // the central building
	"ruin":    {45, 14}, // the crater that breaks the road
	"factory": {37, 8},  // the first yard lane on the north side
}

// TestCompoundEntryStreetMeetsTheRoad: in every installed compound scene with
// an entry road, the painted street runs within 1 world px of the road tile it
// continues, so the gate blend joins two aligned streets: just inside the blend
// (which pins the factory's manifest shift: without it the street is 2 world
// px off) and deep inside, where a street a few degrees off the road's
// direction has drifted apart (the factory's v3 street: 1.5 world px, straightened in v4).
func TestCompoundEntryStreetMeetsTheRoad(t *testing.T) {
	for _, name := range []string{"compound-hq-v4.json", "compound-depot-v6.json", "compound-factory-v6.json", "compound-factory-summer-v4.json", "compound-ruin-v2.json"} {
		in, err := loadCompoundInputs(filepath.Join("..", "..", "art", "pipeline", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range in.Manifest.Scenes {
			if s.EntryRoad == nil {
				continue
			}
			t.Run(s.ID, func(t *testing.T) {
				l, err := compoundLayoutFor(s.ID)
				if err != nil {
					t.Fatal(err)
				}
				deep, ok := compoundEntryDeepBands[l.Layout]
				if !ok {
					t.Fatalf("%s: no deep band for layout %s", name, l.Layout)
				}
				off, err := compoundEntryOffset(in, s, compoundEntryBlend, compoundEntryBlend+8, 14, 6)
				if err != nil {
					t.Fatal(err)
				}
				if math.Abs(off) > 1 {
					t.Errorf("%s: street %+.2f world px off the road at the gate", name, off)
				}
				// Within ±3 world px: a wider search can mirror one rut onto the other.
				off, err = compoundEntryOffset(in, s, deep.far-8, deep.far, deep.half, 3)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("%s: depths %g-%g: %+.2f world px", name, deep.far-8, deep.far, off)
				if math.Abs(off) > 1 {
					t.Errorf("%s: street %+.2f world px off the road's extended axis at depths %g-%g", name, off, deep.far-8, deep.far)
				}
			})
		}
	}
}
