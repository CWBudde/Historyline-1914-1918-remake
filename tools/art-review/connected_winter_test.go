package main

import (
	"image/color"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// winterLab converts an sRGB colour to CIELAB (D65).
func winterLab(c color.NRGBA) (l, a, b float64) {
	lin := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= .04045 {
			return s / 12.92
		}
		return math.Pow((s+.055)/1.055, 2.4)
	}
	f := func(t float64) float64 {
		if t > 216.0/24389 {
			return math.Cbrt(t)
		}
		return (24389.0/27*t + 16) / 116
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	x := (.4124564*r + .3575761*g + .1804375*bl) / .95047
	y := .2126729*r + .7151522*g + .0721750*bl
	z := (.0193339*r + .1191920*g + .9503041*bl) / 1.08883
	return 116*f(y) - 16, 500 * (f(x) - f(y)), 200 * (f(y) - f(z))
}

// summerGreen is the summer meadow's colour: a* < -6 and chroma > 12.
func summerGreen(c color.NRGBA) bool {
	_, a, b := winterLab(c)
	return a < -6 && math.Hypot(a, b) > 12
}

func loadWinterManifest(t *testing.T, name string) (connectionInputs, connectionFamily) {
	t.Helper()
	in, err := loadConnectionInputs(filepath.Join("..", "..", "art", "pipeline", name))
	if err != nil {
		t.Fatal(err)
	}
	return in, in.manifest.Families[0]
}

func TestConnectedSeasonValidation(t *testing.T) {
	for _, tt := range []struct {
		f  connectionFamily
		ok bool
	}{
		{connectionFamily{ID: "stream", BankProfile: authoredBankProfile, Season: "winter"}, true},
		{connectionFamily{ID: "trench", Season: "winter"}, true},
		{connectionFamily{ID: "road"}, true},
		{connectionFamily{ID: "stream", Season: "winter"}, false},
		{connectionFamily{ID: "road", Season: "winter"}, false},
		{connectionFamily{ID: "trench", Season: "summer"}, false},
	} {
		if err := validateConnectionSeason(tt.f); (err == nil) != tt.ok {
			t.Errorf("%+v: %v", tt.f, err)
		}
	}
	for _, name := range []string{"connected-stream-v2.json", "connected-trench-v2.json"} {
		_, f := loadWinterManifest(t, name)
		if f.Season != "" {
			t.Errorf("%s: season %q", name, f.Season)
		}
	}
}

func TestWinterLatticeNoiseRepeatsOnTheLattice(t *testing.T) {
	g := sharedConnectionGeometry()
	lo, hi := math.Inf(1), math.Inf(-1)
	for y := -3.; y < 83; y += .37 {
		for x := -3.; x < 99; x += .41 {
			p := connectionPoint{x, y}
			w := winterIceWidth(p)
			lo, hi = min(lo, w), max(hi, w)
			for _, n := range g.Neighbors {
				if d := math.Abs(winterIceWidth(p.add(n)) - w); d > 1e-6 {
					t.Fatalf("ice width at %v differs by %g after %v", p, d, n)
				}
			}
		}
	}
	if lo < winterIceMin || hi > winterIceMin+winterIceSpan || hi-lo < 1.5 {
		t.Fatalf("ice width %.2f..%.2f", lo, hi)
	}
}

// The winter forms keep the summer geometry, so they join the summer forms'
// neighbours (bridges, outlets) where those are drawn in winter as well.
func TestWinterConnectedKeepsSummerGeometry(t *testing.T) {
	for _, tt := range []struct{ summer, winter string }{
		{"connected-stream-v2.json", "connected-stream-winter-v1.json"},
		{"connected-trench-v2.json", "connected-trench-winter-v1.json"},
	} {
		sin, sf := loadWinterManifest(t, tt.summer)
		win, wf := loadWinterManifest(t, tt.winter)
		if len(sf.Variants) != len(wf.Variants) {
			t.Fatalf("%s: %d variants, summer %d", tt.winter, len(wf.Variants), len(sf.Variants))
		}
		for i, v := range sf.Variants {
			w := wf.Variants[i]
			if w.ID != v.ID+"_w" {
				t.Fatalf("%s: variant %s for %s", tt.winter, w.ID, v.ID)
			}
			w.ID = v.ID
			if !slices.Equal(w.Ports, v.Ports) || !slices.Equal(w.BlindSpurs, v.BlindSpurs) || w.ClosedRing != v.ClosedRing || w.DugoutDirection != v.DugoutDirection {
				t.Fatalf("%s: %s changes the summer shape", tt.winter, v.ID)
			}
		}
		if wf.Width != sf.Width || wf.Border != sf.Border || wf.BankProfile != sf.BankProfile || wf.Season != "winter" {
			t.Fatalf("%s: profile %+v", tt.winter, wf)
		}
		for _, i := range []int{0, len(sf.Variants) / 2} {
			a, err := exportConnectionTiles(sin, &sf, &sf.Variants[i])
			if err != nil {
				t.Fatal(err)
			}
			b, err := exportConnectionTiles(win, &wf, &wf.Variants[i])
			if err != nil {
				t.Fatal(err)
			}
			for j := 3; j < len(a[4].Pix); j += 4 {
				if a[4].Pix[j] != b[4].Pix[j] {
					t.Fatalf("%s: alpha differs at byte %d", wf.Variants[i].ID, j)
				}
			}
		}
	}
}

// Winter streams: frosted banks without summer green, a pale ice rim in the
// outer two world px of the channel, open water in the centre after the
// winter grade.
func TestWinterStreamIceRimAndFrostedBanks(t *testing.T) {
	in, f := loadWinterManifest(t, "connected-stream-winter-v1.json")
	g := in.manifest.Geometry
	var bank, green, greenGraded, rim, ice, centre, water int
	var centreL float64
	for _, v := range f.Variants {
		tiles, err := exportConnectionTiles(in, &f, &v)
		if err != nil {
			t.Fatal(err)
		}
		tile := tiles[4]
		graded := terraingrade.Grade(tile, "tile/"+v.ID, "terrain/stream", nil)
		route := connectionRoutes(g, f, v)
		for y := range 320 {
			for x := range 384 {
				if tile.NRGBAAt(x, y).A != 255 {
					continue
				}
				p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
				core, outer, _ := connectionGravelStreamProfile(g, f, route, p, p)
				outer -= connectionVergeEnvelope(g, p) * (2 + .9*authoredRockOffset(sampleAuthoredBank(in.materials["bank"], p)))
				c := graded.NRGBAAt(x, y)
				l, a, b := winterLab(c)
				switch {
				case core > 1 && outer < -.5:
					bank++
					if summerGreen(tile.NRGBAAt(x, y)) {
						green++
					}
					if summerGreen(c) {
						greenGraded++
					}
				case core >= -2 && core < 0:
					rim++
					if l >= 70 && math.Hypot(a, b) <= 10 {
						ice++
					}
				}
				if connectionRouteDistance(route, p) <= 3 {
					centre++
					centreL += l
					if l <= 45 {
						water++
					}
				}
			}
		}
	}
	t.Logf("bank %d px: summer green %.2f%% (graded %.2f%%); rim %d px: ice %.1f%%; centre %d px: L*<=45 %.1f%%, mean L* %.1f",
		bank, 100*float64(green)/float64(bank), 100*float64(greenGraded)/float64(bank), rim, 100*float64(ice)/float64(rim),
		centre, 100*float64(water)/float64(centre), centreL/float64(centre))
	if float64(green) >= .01*float64(bank) || float64(greenGraded) >= .01*float64(bank) {
		t.Errorf("bank carries summer green")
	}
	if float64(ice) < .6*float64(rim) {
		t.Errorf("ice rim covers too little of the outer channel")
	}
	if float64(water) < .9*float64(centre) || centreL/float64(centre) > 40 {
		t.Errorf("centre is not open water")
	}
}

// Two winter streams that both link an edge paint the same colours through
// its band; the ice rim's noise repeats on the lattice.
func TestWinterStreamsJoinAcrossEdges(t *testing.T) {
	in, f := loadWinterManifest(t, "connected-stream-winter-v1.json")
	g := in.manifest.Geometry
	routes := make([]connectionRoute, len(f.Variants))
	for i, v := range f.Variants {
		routes[i] = connectionRoutes(g, f, v)
	}
	pairs := 0
	for i, a := range f.Variants {
		for j, b := range f.Variants {
			for port := 1; port <= 6; port++ {
				opposite := (port+2)%6 + 1
				if !slices.Contains(a.Ports, port) || !slices.Contains(b.Ports, opposite) {
					continue
				}
				pairs++
				p0, p1 := g.Polygon[port%6], g.Polygon[(port+1)%6]
				edge := p1.sub(p0)
				tangent := edge.mul(1 / edge.length())
				normal := connectionPoint{-tangent[1], tangent[0]}
				for _, depth := range []float64{-6, -3, 0, 3, 6} {
					for along := .2; along <= .8; along += .025 {
						p := p0.add(edge.mul(along)).add(normal.mul(depth))
						q := p.sub(g.Neighbors[port-1])
						ca := connectionFeature(in, f, routes[i], p, p, .25)
						cb := connectionFeature(in, f, routes[j], q, q, .25)
						if ca != cb {
							t.Fatalf("%s/%s port%d depth%g t%g: %v/%v", a.ID, b.ID, port, depth, along, ca, cb)
						}
					}
				}
			}
		}
	}
	if pairs < 30 {
		t.Fatalf("only %d linked pairs", pairs)
	}
}

// Winter trenches: snow on the parapet and parados close to the winter
// ground, no summer green, a dark slush floor.
func TestWinterTrenchSnowAndSlush(t *testing.T) {
	in, f := loadWinterManifest(t, "connected-trench-winter-v1.json")
	g := in.manifest.Geometry
	ground, err := exportConnectionTiles(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var gl, ga, gb float64
	var gn int
	for y := range 320 {
		for x := range 384 {
			if c := ground[4].NRGBAAt(x, y); c.A == 255 {
				l, a, b := winterLab(c)
				gl, ga, gb, gn = gl+l, ga+a, gb+b, gn+1
			}
		}
	}
	gl, ga, gb = gl/float64(gn), ga/float64(gn), gb/float64(gn)
	variants := slices.Clone(f.Variants)
	for i, s := range derivedTrenchShapes(t) {
		if i%7 == 0 {
			v := derivedTrenchVariant(f.Variants, s)
			v.ID += "_w"
			variants = append(variants, v)
		}
	}
	for _, v := range variants {
		tiles, err := exportConnectionTiles(in, &f, &v)
		if err != nil {
			t.Fatal(err)
		}
		tile := tiles[4]
		graded := terraingrade.Grade(tile, "tile/"+v.ID, "terrain/trench", nil)
		r := connectionTrenchRoutes(g, f, v)
		var bl, ba, bb, floorL, floorGraded float64
		var band, green, floor int
		for y := range 320 {
			for x := range 384 {
				c := tile.NRGBAAt(x, y)
				if c.A != 255 {
					continue
				}
				p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
				core, _ := connectionTrenchCore(g, f, r, p)
				l, a, b := winterLab(c)
				switch {
				case core >= 1.5:
					bl, ba, bb, band = bl+l, ba+a, bb+b, band+1
					if summerGreen(c) {
						green++
					}
				case core <= -1.5:
					lg, _, _ := winterLab(graded.NRGBAAt(x, y))
					floorL, floorGraded, floor = floorL+l, floorGraded+lg, floor+1
				}
			}
		}
		bl, ba, bb = bl/float64(band), ba/float64(band), bb/float64(band)
		de := math.Sqrt((bl-gl)*(bl-gl) + (ba-ga)*(ba-ga) + (bb-gb)*(bb-gb))
		floorL, floorGraded = floorL/float64(floor), floorGraded/float64(floor)
		t.Logf("%-16s band %5d px L*a*b* %.1f %.1f %.1f, ΔE %.1f to ground %.1f %.1f %.1f, green %.2f%%; floor %4d px L* %.1f (graded %.1f)",
			v.ID, band, bl, ba, bb, de, gl, ga, gb, 100*float64(green)/float64(band), floor, floorL, floorGraded)
		if de > 10 || float64(green) >= .01*float64(band) || floorL > 40 || floorGraded > 40 {
			t.Errorf("%s: not a winter trench", v.ID)
		}
		if !strings.HasSuffix(v.ID, "_w") {
			t.Errorf("%s: winter IDs end in _w", v.ID)
		}
	}
}

// The winter trenches, derived shapes included, join and reach their ends as
// the summer ones do, with the real winter materials.
func TestWinterDerivedTrenchesJoinAndReachTheirEnds(t *testing.T) {
	in, f := loadWinterManifest(t, "connected-trench-winter-v1.json")
	checkDerivedTrenchesJoin(t, in, f)
	checkDerivedTrenchEnds(t, in, f)
}
