package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"maps"
	"math"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// trackTileHash is the SHA-256 of the PNG encoding of one exported tile.
func trackTileHash(t *testing.T, im *image.NRGBA) string {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b.Bytes())
	return hex.EncodeToString(sum[:])
}

// The earlier rail, path and mixed manifests keep their narrow profiles: a
// sample of their tiles still renders byte for byte as before the profiles
// were derived from the family width.
func TestEarlierTrackManifestsRenderUnchanged(t *testing.T) {
	root := filepath.Join("..", "..")
	connected := []struct{ manifest, id, x2, x4 string }{
		{"connected-rail-v1.json", "rail14", "75ef860f167248b7c459f6d27c7e00f3ed2de8a8189e3f8e2cf1833a2eed52e8", "3714d1be86654746ee1b64c9a78ffb39906e45643b8519142802898ee7f4a465"},
		{"connected-rail-v2.json", "rail236", "0eb07ed837e319ca79fad6c7fd98e5a141dcbf49c4217b949aec267ad62a196c", "5edb2590934d8a7e8396eb3122e41f3cca19a2e31351d163f966a847f95c55b0"},
		{"connected-paths-v1.json", "path14", "1af5ea7f8173750a9779153497dd1128ce8a30a9f52153296826cae09c53430b", "589c59f36872ec6901ac0d14ba0f6dd549ecd534ddd55a40644c2b2f35d6017e"},
		{"connected-paths-v2.json", "path135", "ff2268cbe6d6884828e99fdd6603e6d6c08b4fe7c70b0b54bee397e345c3818a", "8d7f344fb8c7c446068868e914b368c053aa9ed70d85638b5d2175db799ce10e"},
		{"connected-rail-winter-v1.json", "rail346_w", "3d5a8dbcabdf16cc0c779e40d8aeb7a630581e6907a94e6f3bbe69a84ab09b0f", "e3ae6674fc79bbc1baca469816973d5f02197797ff9b9fc379178c7916bd246d"},
		{"connected-paths-winter-v1.json", "path26_w", "3d5cd8ce24e81d130dd9eb44130e60075cb701380ba107cb7fea27655d9bc12d", "0cd76be360e53240c112d50fc3865ca26cef315335d09d582061fadb8984a772"},
	}
	for _, c := range connected {
		t.Run(c.id, func(t *testing.T) {
			t.Parallel()
			in, err := loadConnectionInputs(filepath.Join(root, "art/pipeline", c.manifest))
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range in.manifest.Families {
				for _, v := range f.Variants {
					if v.ID != c.id {
						continue
					}
					tiles, err := exportConnectionTiles(in, &f, &v)
					if err != nil {
						t.Fatal(err)
					}
					if x2, x4 := trackTileHash(t, tiles[2]), trackTileHash(t, tiles[4]); x2 != c.x2 || x4 != c.x4 {
						t.Errorf("%s %s changed: 2x %s 4x %s", c.manifest, c.id, x2, x4)
					}
					return
				}
			}
			t.Fatalf("%s has no %s", c.manifest, c.id)
		})
	}
	mixed := []struct{ manifest, id, x2, x4 string }{
		{"connected-mixed-v1.json", "bridge25", "53bbcbad3283dc793cf961c19d69bad8ec7f79795e4219a88de967ec23f3f365", "1e9c91663dbde217a435995e6739ebfd6586dd9a53f35211c8b18154359e9a8f"},
		{"connected-mixed-v2.json", "cross1", "d070a8113c71c5326bf9904a5898226edb7f813db13f3bfca27d868a057161e1", "d8983868fa58b1b646ca69ff3a807fac44a1375357abd94d8334f2a6777896a2"},
		{"connected-mixed-v2.json", "bridge36", "39f2f0d05b117b3e015785e21ea1573fbb8aeb7e15410ed6edec59267cd862b1", "fec2d430a157686460b5d60426fa26e8bc159a7ed5097da006e4e649a3bf353f"},
		{"connected-mixed-winter-v1.json", "cross2", "0581da8a1e63dea764023584932ff310495179abcd57554f9c927f51db054443", "7bc983638e519c575880f1644a5f7d9769f857d844a00caa6aae51098e6bad8f"},
	}
	for _, c := range mixed {
		t.Run(c.manifest+"/"+c.id, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(root, "art/pipeline", c.manifest)
			var in mixedInputs
			var variants []mixedVariant
			if c.manifest == "connected-mixed-winter-v1.json" {
				w, err := loadWinterMixedInputs(path)
				if err != nil {
					t.Fatal(err)
				}
				in, variants = w.inputs(), w.variants
			} else {
				var err error
				if in, err = loadMixedInputs(path); err != nil {
					t.Fatal(err)
				}
				variants = in.manifest.Variants
			}
			for _, v := range variants {
				if v.ID != c.id {
					continue
				}
				tiles, err := exportMixedTiles(in, v)
				if err != nil {
					t.Fatal(err)
				}
				if x2, x4 := trackTileHash(t, tiles[2]), trackTileHash(t, tiles[4]); x2 != c.x2 || x4 != c.x4 {
					t.Errorf("%s %s changed: 2x %s 4x %s", c.manifest, c.id, x2, x4)
				}
				return
			}
			t.Fatalf("%s has no %s", c.manifest, c.id)
		})
	}
}

// The selected rail and path manifests: tracks as wide as the road.
var wideTrackManifests = []string{"connected-rail-v3.json", "connected-paths-v3.json", "connected-rail-winter-v2.json", "connected-paths-winter-v2.json"}

// trackPortProfile is a track's cross-section at one port, in world px: the
// core (surface) width, the outer width (core and border) and, on a rail, the
// centre-to-centre gauge of the two steel rails.
type trackPortProfile struct{ core, outer, gauge float64 }

// measureTrackPort measures a 4x tile at port across the route, averaged over
// the straight guard 2..7.5 world px inside the port. The outer edge is where
// alpha falls to one half; the core edge is the colour step between the core
// and the border.
func measureTrackPort(im *image.NRGBA, port int) trackPortProfile {
	g := sharedConnectionGeometry()
	inward := connectionRailUnit(g.Neighbors[port-1]).mul(-1)
	normal := connectionRailNormal(inward)
	const step, reach = .25, 16.
	bins := int(2*reach/step) + 1
	alpha := make([]float64, bins)
	colour := make([][3]float64, bins)
	for i := range bins {
		a := -reach + float64(i)*step
		opaque, rows := 0., 0.
		for d := 2.; d <= 7.5; d += step {
			q := g.Ports[port-1].add(inward.mul(d)).add(normal.mul(a))
			c := im.NRGBAAt(int(math.Floor(q[0]*4)), int(math.Floor(q[1]*4)))
			alpha[i] += float64(c.A) / 255
			rows++
			if c.A == 255 {
				colour[i][0] += float64(c.R)
				colour[i][1] += float64(c.G)
				colour[i][2] += float64(c.B)
				opaque++
			}
		}
		alpha[i] /= rows
		if opaque > 0 {
			for k := range 3 {
				colour[i][k] /= opaque
			}
		}
	}
	// The core edge is fitted on RGB and on chroma alone (R-G, G-B), and the
	// clearer step is taken: the materials' luminance texture would swamp a
	// faint step such as a winter rail's, whose snowed ballast shoulder and
	// snowed earth border differ mostly in tint, while a winter path's grey
	// track and snowed verge differ mostly in lightness.
	chroma := make([][3]float64, bins)
	for i, c := range colour {
		chroma[i] = [3]float64{c[0] - c[1], c[1] - c[2], 0}
	}
	at := func(a float64) int { return int(math.Round((a + reach) / step)) }
	mean := func(from, to float64) [3]float64 {
		var m [3]float64
		n := 0.
		for i := at(min(from, to)); i <= at(max(from, to)); i++ {
			for k := range 3 {
				m[k] += colour[i][k]
			}
			n++
		}
		for k := range 3 {
			m[k] /= n
		}
		return m
	}
	// side measures one half of the profile (sign -1 left, +1 right) and
	// returns the outer and core edge distances from the centre line.
	side := func(sign float64) (float64, float64) {
		outer := reach
		for a := step; a <= reach; a += step {
			if cur := alpha[at(sign*a)]; cur < .5 {
				prev := alpha[at(sign*(a-step))]
				outer = a - step + step*(prev-.5)/(prev-cur)
				break
			}
		}
		// Fit one colour step to the bins 0.25..2.6 world px inside the outer
		// edge (the core side and the border), then place the edge where the
		// colour crosses half-way between the two fitted plateaus in the two
		// bins either side of the step.
		lo, hi := at(math.Ceil((outer-2.6)/step)*step), at(math.Floor((outer-.25)/step)*step)
		if sign < 0 {
			lo, hi = at(-math.Floor((outer-.25)/step)*step), at(-math.Ceil((outer-2.6)/step)*step)
		}
		rgb, rgbF := trackStepEdge(colour, lo, hi, sign, step, reach)
		tint, tintF := trackStepEdge(chroma, lo, hi, sign, step, reach)
		if tintF > rgbF {
			return outer, tint
		}
		return outer, rgb
	}
	lo, lc := side(-1)
	ro, rc := side(1)
	p := trackPortProfile{core: lc + rc, outer: lo + ro}
	// The rails: the centroids of the colour departure from the gauge's
	// centre, above half its peak, either side within 1 world px and 0.3 of
	// the outer width.
	centre := mean(-.5, .5)
	rail := func(sign float64) float64 {
		dev := func(a float64) float64 {
			c := colour[at(sign*a)]
			return math.Sqrt((c[0]-centre[0])*(c[0]-centre[0]) + (c[1]-centre[1])*(c[1]-centre[1]) + (c[2]-centre[2])*(c[2]-centre[2]))
		}
		peak, best := 0., 1.
		for a := 1.; a <= .3*p.outer; a += step {
			if d := dev(a); d > peak {
				peak, best = d, a
			}
		}
		sum, weight := best*peak, peak
		for _, dir := range []float64{-step, step} {
			for a := best + dir; a >= 1 && a <= .3*p.outer && dev(a) >= peak/2; a += dir {
				sum += a * dev(a)
				weight += dev(a)
			}
		}
		return sum / weight
	}
	p.gauge = rail(-1) + rail(1)
	return p
}

// trackStepEdge fits one step to the profile bins lo..hi (the core side
// toward the centre line, the border outward; sign -1 is the left half) and
// returns the edge's distance from the centre line, where the profile crosses
// half-way between the two fitted plateaus next to the step, and the step's
// Fisher ratio (between-plateau over within-plateau variance).
func trackStepEdge(profile [][3]float64, lo, hi int, sign, step, reach float64) (float64, float64) {
	pos := func(i int) float64 { return math.Abs(-reach + float64(i)*step) }
	dist := func(c, m [3]float64) float64 {
		return (c[0]-m[0])*(c[0]-m[0]) + (c[1]-m[1])*(c[1]-m[1]) + (c[2]-m[2])*(c[2]-m[2])
	}
	plateau := func(from, to int) [3]float64 {
		var m [3]float64
		for i := from; i <= to; i++ {
			for k := range 3 {
				m[k] += profile[i][k] / float64(to-from+1)
			}
		}
		return m
	}
	var inner, rim [3]float64
	best, cut := math.Inf(1), 0
	for split := lo + 1; split <= hi; split++ {
		a, b := plateau(lo, split-1), plateau(split, hi)
		cost := 0.
		for i := lo; i <= hi; i++ {
			if i < split {
				cost += dist(profile[i], a)
			} else {
				cost += dist(profile[i], b)
			}
		}
		if cost < best {
			best, inner, rim, cut = cost, a, b, split
		}
	}
	n, left := float64(hi-lo+1), float64(cut-lo)
	fisher := left * (n - left) / n * dist(inner, rim) / (best/(n-2) + 1e-9)
	// Bins cut-1 and cut straddle the step; on the right side the lower
	// index is the inner one.
	in, out := cut-1, cut
	if sign < 0 {
		inner, rim = rim, inner
		in, out = cut, cut-1
	}
	span := [3]float64{inner[0] - rim[0], inner[1] - rim[1], inner[2] - rim[2]}
	norm := span[0]*span[0] + span[1]*span[1] + span[2]*span[2]
	share := func(i int) float64 {
		c := profile[i]
		return ((c[0]-rim[0])*span[0] + (c[1]-rim[1])*span[1] + (c[2]-rim[2])*span[2]) / norm
	}
	edge := (pos(in) + pos(out)) / 2
	if s1, s2 := share(in), share(out); s1 > .5 && s2 < .5 {
		edge = pos(in) + (pos(out)-pos(in))*(s1-.5)/(s1-s2)
	}
	return edge, fisher
}

// trackTiles lists the installed rail and path tiles (4x) by ID.
func trackTiles(t *testing.T, root string) map[string]string {
	t.Helper()
	tiles := map[string]string{}
	for _, family := range []string{"rail", "path"} {
		matches, err := filepath.Glob(filepath.Join(root, "assets/terrain", family, family+"*-4x.png"))
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range matches {
			tiles[strings.TrimSuffix(filepath.Base(m), "-4x.png")] = m
		}
	}
	if len(tiles) != 52 {
		t.Fatalf("%d installed rail and path tiles, want 52", len(tiles))
	}
	return tiles
}

// trackPorts reads the ports from a rail or path tile ID such as path135_w.
func trackPorts(id string) []int {
	var ports []int
	for _, r := range strings.TrimSuffix(strings.TrimLeft(id, "pathril"), "_w") {
		ports = append(ports, int(r-'0'))
	}
	return ports
}

// Criterion: rails and paths as wide as the road. At every port of every
// installed rail and path tile, summer and winter, the core is 22 ± 0.5 and
// the outer width 26 ± 0.75 world px, as road36's; the rails stand 9 ± 0.25
// world px apart centre to centre.
func TestTracksAsWideAsTheRoad(t *testing.T) {
	root := filepath.Join("..", "..")
	road, err := loadPNG(filepath.Join(root, "assets/terrain/road/road36-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	want := measureTrackPort(road, 6)
	t.Logf("road36 port 6: core %.2f outer %.2f", want.core, want.outer)
	if math.Abs(want.core-22) > .5 || math.Abs(want.outer-26) > .75 {
		t.Fatalf("road36 measures core %.2f outer %.2f, want 22 and 26", want.core, want.outer)
	}
	tiles := trackTiles(t, root)
	for _, id := range slices.Sorted(maps.Keys(tiles)) {
		im, err := loadPNG(tiles[id])
		if err != nil {
			t.Fatal(err)
		}
		line := id
		for _, port := range trackPorts(id) {
			p := measureTrackPort(im, port)
			line += fmt.Sprintf("  %d: %.2f/%.2f", port, p.core, p.outer)
			if strings.HasPrefix(id, "rail") {
				line += fmt.Sprintf(" g%.2f", p.gauge)
			}
			if math.Abs(p.core-22) > .5 || math.Abs(p.outer-26) > .75 || math.Abs(p.core-want.core) > .5 || math.Abs(p.outer-want.outer) > .75 {
				t.Errorf("%s port %d: core %.2f outer %.2f, want 22 ± 0.5 and 26 ± 0.75 as road36 (%.2f/%.2f)", id, port, p.core, p.outer, want.core, want.outer)
			}
			if strings.HasPrefix(id, "rail") && math.Abs(p.gauge-9) > .25 {
				t.Errorf("%s port %d: gauge %.2f, want 9 ± 0.25", id, port, p.gauge)
			}
		}
		t.Log(line)
	}
}

// trackSteelBreaks counts the samples along the rail centre lines of r, every
// quarter world pixel inside the hex, whose installed 4x pixel is not the
// steel the family paints there.
func trackSteelBreaks(im *image.NRGBA, in connectionInputs, f connectionFamily, r connectionRailGeometry) (breaks, samples int) {
	g := sharedConnectionGeometry()
	for _, seg := range r.rails.segments {
		d := seg.b.sub(seg.a)
		steps := max(1, int(math.Ceil(d.length()*4)))
		for i := range steps + 1 {
			p := seg.a.add(d.mul(float64(i) / float64(steps)))
			if !connectionInside(g, p) || mixedEdgeDistance(g, p) < .5 {
				continue
			}
			x, y := int(math.Floor(p[0]*4)), int(math.Floor(p[1]*4))
			centre := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
			samples++
			if im.NRGBAAt(x, y) != connectionRailSteel(in, f, centre) {
				breaks++
			}
		}
	}
	return breaks, samples
}

// The installed wide tracks join and run: at every port each rail and path
// tile is pixel for pixel the straight tile of that axis (so any two tiles
// sharing a port join as the straight tiles do), and the steel of every rail
// tile and level crossing, summer and winter, runs unbroken along both rails.
func TestInstalledWideTracksJoinAndRailsRun(t *testing.T) {
	root := filepath.Join("..", "..")
	g := sharedConnectionGeometry()
	inputs := map[string]connectionInputs{}
	for _, name := range wideTrackManifests {
		in, err := loadConnectionInputs(filepath.Join(root, "art/pipeline", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range in.manifest.Families {
			if f.Width != 22 || f.Border != 2 {
				t.Fatalf("%s: %s profile %v/%v, want the road's 22/2", name, f.ID, f.Width, f.Border)
			}
			for _, v := range f.Variants {
				inputs[v.ID] = in
			}
		}
	}
	load := func(family, id string, d int) *image.NRGBA {
		im, err := loadPNG(filepath.Join(root, "assets/terrain", family, fmt.Sprintf("%s-%dx.png", id, d)))
		if err != nil {
			t.Fatal(err)
		}
		return im
	}
	axis := map[int]string{1: "14", 4: "14", 2: "25", 5: "25", 3: "36", 6: "36"}
	for _, id := range slices.Sorted(maps.Keys(trackTiles(t, root))) {
		in, ok := inputs[id]
		if !ok {
			t.Fatalf("%s is in no selected manifest", id)
		}
		f := in.manifest.Families[0]
		winter := strings.HasSuffix(id, "_w")
		for _, d := range []int{2, 4} {
			tile := load(f.ID, id, d)
			for _, port := range trackPorts(id) {
				straight := f.ID + axis[port]
				if winter {
					straight += "_w"
				}
				ref := load(f.ID, straight, d)
				inward := connectionRailUnit(g.Neighbors[port-1]).mul(-1)
				normal := connectionRailNormal(inward)
				for y := range 80 * d {
					for x := range 96 * d {
						q := connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}.sub(g.Ports[port-1])
						along, across := q[0]*inward[0]+q[1]*inward[1], math.Abs(q[0]*normal[0]+q[1]*normal[1])
						if along > 4 || across > 15 {
							continue
						}
						if a, b := tile.NRGBAAt(x, y), ref.NRGBAAt(x, y); a != b && (d == 4 || !railTestSamePixel(a, b)) {
							t.Fatalf("%s %dx port %d differs from %s at %v: %v != %v", id, d, port, straight, q, a, b)
						}
					}
				}
			}
		}
		if f.ID == "rail" {
			v := connectionVariant{ID: id, Ports: trackPorts(id)}
			breaks, samples := trackSteelBreaks(load("rail", id, 4), in, f, connectionRailRoutes(g, f, v))
			if breaks > 0 || samples < 500 {
				t.Errorf("%s: %d of %d rail centre samples are not steel", id, breaks, samples)
			}
		}
	}
	summer, err := loadMixedInputs(filepath.Join(root, "art/pipeline/connected-mixed-v3.json"))
	if err != nil {
		t.Fatal(err)
	}
	w, err := loadWinterMixedInputs(filepath.Join(root, "art/pipeline/connected-mixed-winter-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	for suffix, mixed := range map[string]mixedInputs{"": summer, "_w": w.inputs()} {
		for _, v := range mixedRoster() {
			if v.Kind != "crossing" {
				continue
			}
			track := mixed.components[v.Layers[1]]
			if track.family.Width != 22 || mixedRailHalf(track.family) != 9 {
				t.Fatalf("%s%s: track %v wide, plank half-width %v", v.RuntimeID, suffix, track.family.Width, mixedRailHalf(track.family))
			}
			breaks, samples := trackSteelBreaks(load("cross", v.RuntimeID+suffix, 4), track.input, track.family, connectionRailRoutes(g, track.family, track.variant))
			if breaks > 0 || samples < 500 {
				t.Errorf("%s%s: %d of %d rail centre samples are not steel over the crossing", v.RuntimeID, suffix, breaks, samples)
			}
		}
	}
}

// Crossings and footbridges over road-wide paths and rails: the plank infill
// spans the rail's 9-pixel half-width, the steel stays continuous over it and
// the procedural footbridge deck is as wide as the path.
func TestMixedWideTracks(t *testing.T) {
	for _, f := range []connectionFamily{{ID: "path", Width: 22, Border: 2}, {ID: "rail", Width: 22, Border: 2}, {ID: "path", Width: 7, Border: .75}, {ID: "rail", Width: 11, Border: 1}} {
		if !mixedProfileCompatible(f) {
			t.Fatalf("%s %v/%v refused", f.ID, f.Width, f.Border)
		}
	}
	for _, f := range []connectionFamily{{ID: "path", Width: 22, Border: 1}, {ID: "rail", Width: 21, Border: 2}, {ID: "rail", Width: 22, Border: .75}, {ID: "stream", Width: 22, Border: 2}} {
		if mixedProfileCompatible(f) {
			t.Fatalf("%s %v/%v accepted", f.ID, f.Width, f.Border)
		}
	}
	if p := mixedProfileFor(22); p.FootWidth != 22 || !strings.Contains(p.Method, "Footbridge 48x22 timber deck, side curbs at transverse 11..12.5, path-width 22 approaches") {
		t.Fatalf("wide footbridge profile %v: %s", p.FootWidth, p.Method)
	}
	if p := mixedProfileFor(7); p != fixedMixedProfile() {
		t.Fatal("narrow footbridge profile changed")
	}
	path, _ := mixedFixtureTracks(t, [2]float64{22, 2}, [2]float64{22, 2})
	in, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	steel := in.merged.materials["steel"].NRGBAAt(0, 0)
	ballast := in.merged.materials["rail"].NRGBAAt(0, 0)
	raw := renderMixedRaw(in, mixedRoster()[6]) // road14 over rail36
	at := func(im *image.NRGBA, x, y float64) color.NRGBA { return im.NRGBAAt(int((x+4)*4), int((y+4)*4)) }
	for y := 1.; y < 79; y += .25 {
		for _, x := range []float64{48 - 4.5, 48 + 4.5} {
			if got := at(raw, x, y); got != steel {
				t.Fatalf("crossing steel broken at %v,%v: %v", x, y, got)
			}
		}
	}
	// Planks reach past the 15-pixel sleepers to 9 world px from the track's
	// centre; ballast lies beyond them.
	for _, x := range []float64{48 - 8, 48 + 7.75} {
		if c := at(raw, x, 41); c.R < 125 || c.B > 90 || c.A != 255 {
			t.Fatalf("no plank infill at %v: %v", x, c)
		}
	}
	if c := at(raw, 48+10, 41); c != ballast {
		t.Fatalf("infill wider than the rail's 9-pixel half-width: %v", c)
	}
	// The procedural footbridge deck over the 22-pixel path.
	tiles, err := exportMixedTiles(in, mixedRoster()[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		im := tiles[scale]
		pick := func(x, y float64) color.NRGBA { return im.NRGBAAt(int(x*float64(scale)), int(y*float64(scale))) }
		if deck := pick(48+10.25, 30); deck.A != 255 || deck.R < 100 {
			t.Fatalf("%dx footbridge deck narrower than 22: %v", scale, deck)
		}
		if post, beam := pick(48+11.75, 40), pick(48+11.75, 44); int(post.R)-int(beam.R) < 15 {
			t.Fatalf("%dx no guard post on the wide deck's side beam: post=%v beam=%v", scale, post, beam)
		}
		if c := pick(48+14.5, 40); c.A != 0 {
			t.Fatalf("%dx footbridge wider than its side beams: %v", scale, c)
		}
	}
}

// A summer path with markings ruts carries two faint cart ruts 5 world px out,
// worn toward its earth border, and keeps that plain earth verge.
func TestConnectedSummerPathRuts(t *testing.T) {
	dir, m := connectedFixture(t)
	f := connectedPathFamily()
	f["width"], f["border"], f["markings"] = 22, 2, "ruts"
	f["variants"] = []any{map[string]any{"id": "path36", "ports": []int{3, 6}}}
	m["families"] = []any{f}
	solid := func(name string, c color.NRGBA) connectionMaterial {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for i := 0; i < len(im.Pix); i += 4 {
			im.Pix[i], im.Pix[i+1], im.Pix[i+2], im.Pix[i+3] = c.R, c.G, c.B, c.A
		}
		return unitTestPNG(t, filepath.Join(dir, name), im)
	}
	tan, earth := color.NRGBA{210, 175, 115, 255}, color.NRGBA{90, 60, 35, 255}
	mats := m["materials"].(map[string]any)
	mats["path"], mats["earth"] = solid("tan.png", tan), solid("earth.png", earth)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := exportConnectionTiles(in, &in.manifest.Families[0], &in.manifest.Families[0].Variants[0])
	if err != nil {
		t.Fatal(err)
	}
	at := func(x float64) color.NRGBA { return tiles[4].NRGBAAt(int(x*4), 2*4) }
	if c := at(48); c != tan {
		t.Fatalf("centre between the ruts %v, want the plain surface %v", c, tan)
	}
	offset, _, _ := connectionPathRutProfile(22)
	rut := connectionBlend(tan, earth, connectionPathSummerRutStrength)
	for _, x := range []float64{48 - offset, 48 + offset} {
		if c := at(x); math.Abs(float64(c.R)-float64(rut.R)) > 2 || c.A != 255 {
			t.Fatalf("rut at %v is %v, want about %v", x, c, rut)
		}
	}
	if c := at(48 + 12); c != earth {
		t.Fatalf("summer verge %v, want the plain earth %v", c, earth)
	}
}
