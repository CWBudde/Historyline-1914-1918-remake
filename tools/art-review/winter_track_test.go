package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Winter paths: two narrow cart ruts on the snow-packed surface and snow
// instead of the earth verge, at the summer path width (7/0.75 and the road's
// 22/2, whose ruts lie 5 world px out).
func TestConnectedWinterPathActualPNG(t *testing.T) {
	for _, profile := range []struct{ width, border, verge, outside, rut float64 }{{7, .75, 3.95, 5, 1.6}, {22, 2, 12, 14.5, 5.03}} {
		t.Run(fmt.Sprint(profile.width), func(t *testing.T) {
			dir, m := connectedFixture(t)
			f := connectedPathFamily()
			f["width"], f["border"] = profile.width, profile.border
			f["markings"], f["verge"] = "ruts", "snow"
			f["variants"] = []any{map[string]any{"id": "path36_w", "ports": []int{3, 6}}}
			m["families"] = []any{f}
			solid := func(name string, c color.NRGBA) connectionMaterial {
				im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
				draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
				return unitTestPNG(t, filepath.Join(dir, name), im)
			}
			mats := m["materials"].(map[string]any)
			mats["path"] = solid("snow-track.png", color.NRGBA{140, 138, 140, 255})
			mats["grass"] = solid("snow-ground.png", color.NRGBA{218, 221, 225, 255})
			mats["earth"] = solid("earth.png", color.NRGBA{90, 60, 35, 255})
			in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
			if err != nil {
				t.Fatal(err)
			}
			if offset, _, _ := connectionPathRutProfile(profile.width); math.Abs(offset-profile.rut) > .01 {
				t.Fatalf("ruts %.2f world px out, want %.2f", offset, profile.rut)
			}
			tiles, err := exportConnectionTiles(in, &in.manifest.Families[0], &in.manifest.Families[0].Variants[0])
			if err != nil {
				t.Fatal(err)
			}
			im := tiles[4]
			// Near the port the path keeps its fixed profile.
			y := 2 * 4
			at := func(x float64) color.NRGBA { return im.NRGBAAt(int(x*4), y) }
			centre := at(48)
			for _, x := range []float64{48 - profile.rut, 48 + profile.rut} {
				if rut := at(x); int(rut.R)+15 > int(centre.R) || rut.A != 255 {
					t.Fatalf("rut at %v not darker than the centre: %v vs %v", x, rut, centre)
				}
			}
			for _, x := range []float64{48 - profile.verge, 48 + profile.verge} {
				if verge := at(x); verge.A < 128 || verge.R < 190 {
					t.Fatalf("verge at %v is not snow: %v", x, verge)
				}
			}
			for _, x := range []float64{48 - profile.outside, 48 + profile.outside} {
				if outside := at(x); outside.A > 0 {
					t.Fatalf("winter path wider than %v+%v: %v at %v", profile.width, profile.border, outside, x)
				}
			}
		})
	}
}

// Winter rails: snow on the ballast shoulders and the earth border, a lighter
// dusting between the sleepers, the sleepers still darker than the ballast
// between them and the steel unchanged; on the 11-pixel track and on the
// road-wide 22-pixel one, whose snow lines scale with the track.
func TestConnectedWinterRailActualPNG(t *testing.T) {
	for _, profile := range [][2]float64{{11, 1}, {22, 2}} {
		t.Run(fmt.Sprint(profile[0]), func(t *testing.T) {
			dir, m := railTestFixture(t)
			family := m["families"].([]any)[0].(map[string]any)
			family["verge"], family["width"], family["border"] = "snow", profile[0], profile[1]
			snow := color.NRGBA{220, 222, 226, 255}
			ground := image.NewNRGBA(image.Rect(0, 0, 384, 320))
			draw.Draw(ground, ground.Bounds(), image.NewUniform(snow), image.Point{}, draw.Src)
			m["materials"].(map[string]any)["grass"] = unitTestPNG(t, filepath.Join(dir, "snow.png"), ground)
			in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
			if err != nil {
				t.Fatal(err)
			}
			f := in.manifest.Families[0]
			var v connectionVariant
			for _, c := range f.Variants {
				if c.ID == "rail36" {
					v = c
				}
			}
			tiles, err := exportConnectionTiles(in, &f, &v)
			if err != nil {
				t.Fatal(err)
			}
			d := connectionRailDimsFor(f.Width)
			s := d.scale
			im := tiles[4]
			at := func(x, y float64) color.NRGBA { return im.NRGBAAt(int(x*4), int(y*4)) }
			// The straight interior sleepers lie at 16 + n pitches.
			gap, tie := 16+d.pitch/2, 16+d.pitch
			ballast := railTestColors["rail"]
			for _, x := range []float64{48 - 4.6*s, 48 + 4.6*s} {
				if c := at(x, gap); c.R < 200 {
					t.Fatalf("shoulder at %v not snowy: %v", x, c)
				}
			}
			for _, x := range []float64{48 - 6*s, 48 + 6*s} {
				if c := at(x, gap); c.A < 128 || c.R < 195 {
					t.Fatalf("border at %v not snowy: %v", x, c)
				}
			}
			between, sleeper := at(48, gap), at(48, tie)
			if between.R <= ballast.R+20 || between.R > snow.R-20 {
				t.Fatalf("ballast between sleepers not partly snowed: %v", between)
			}
			if int(sleeper.R)+30 > int(between.R) {
				t.Fatalf("sleeper %v not darker than the snowed ballast %v", sleeper, between)
			}
			if sleeper.R <= railTestColors["timber"].R {
				t.Fatalf("sleeper not dusted: %v", sleeper)
			}
			cold := connectionBlend(railTestColors["steel"], connectionRailWinterSteel, connectionRailWinterSteelMix)
			for _, x := range []float64{48 - d.gauge/2, 48 + d.gauge/2} {
				if c := at(x, gap); c != cold || c.R >= railTestColors["steel"].R {
					t.Fatalf("steel at %v is %v, want the darker cold steel %v", x, c, cold)
				}
			}
			if c := at(48+7.5*s, gap); c.A != 0 {
				t.Fatalf("winter rail wider than %v+%v: %v", profile[0], profile[1], c)
			}
		})
	}
}

// The winter mixed manifests (v1 narrow, v2 as wide as the road) derive the
// winter footbridges and crossings from the summer mixed tiles with every
// component swapped for its _w form of the same profile.
func TestWinterMixedManifestSwapsEveryComponent(t *testing.T) {
	root := filepath.Join("..", "..")
	for manifest, width := range map[string]float64{"connected-mixed-winter-v1.json": 0, "connected-mixed-winter-v2.json": 22} {
		w, err := loadWinterMixedInputs(filepath.Join(root, "art/pipeline", manifest))
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, v := range w.variants {
			ids = append(ids, v.RuntimeID)
		}
		if want := []string{"bridge25", "bridge36", "cross1", "cross2", "cross3", "cross4"}; !slices.Equal(ids, want) {
			t.Fatalf("%s: winter mixed variants %v, want %v", manifest, ids, want)
		}
		in := w.inputs()
		for _, v := range w.variants {
			for _, id := range winterMixedComponentIDs(v) {
				c, summer := in.components[id], w.base.components[id]
				if c.variant.ID != id+"_w" || !mixedProfileCompatible(c.family) || c.family.Width != summer.family.Width || c.family.Border != summer.family.Border {
					t.Fatalf("%s %s: component %s is %s (%s %v/%v)", manifest, v.ID, id, c.variant.ID, c.family.ID, c.family.Width, c.family.Border)
				}
				if c.family.Verge != "snow" {
					t.Fatalf("%s %s: component %s has no snow verge", manifest, v.ID, id)
				}
				if width > 0 && c.family.Width != width {
					t.Fatalf("%s %s: component %s is %v wide, want %v", manifest, v.ID, id, c.family.Width, width)
				}
			}
		}
	}
}

// A winter component of another profile than its summer form is refused.
func TestWinterMixedManifestRejectsProfileMix(t *testing.T) {
	root := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(root, "art/pipeline/connected-mixed-winter-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	// The narrow v2 summer base under the wide winter components.
	m["base"] = map[string]any{"source": "art/pipeline/connected-mixed-v2.json", "source_sha256": "71a4e6db42b6e1d1d9305ad3fb4c0edfab74639f8fe2b4a52136ba39eda44d1b"}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	old := sourceRoot
	sourceRoot = abs
	defer func() { sourceRoot = old }()
	p := filepath.Join(t.TempDir(), "winter-mixed.json")
	if err := os.WriteFile(p, out, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadWinterMixedInputs(p); err == nil || !strings.Contains(err.Error(), "profile") {
		t.Fatalf("narrow summer base with wide winter components accepted: %v", err)
	}
}

func TestWinterMixedManifestRejectsDrift(t *testing.T) {
	root := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(root, "art/pipeline/connected-mixed-winter-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]func(map[string]any){
		"unchanged":    nil,
		"base pin":     func(m map[string]any) { m["base"].(map[string]any)["source_sha256"] = strings.Repeat("0", 64) },
		"road bridge":  func(m map[string]any) { m["variants"] = append(m["variants"].([]any), "rbridge25") },
		"duplicate":    func(m map[string]any) { m["variants"] = append(m["variants"].([]any), "cross1") },
		"stray field":  func(m map[string]any) { m["season"] = "winter" },
		"no manifests": func(m map[string]any) { m["manifests"] = []any{} },
	} {
		t.Run(name, func(t *testing.T) {
			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}
			if edit != nil {
				edit(m)
			}
			out, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			// Relative sources resolve against the repository root.
			abs, err := filepath.Abs(root)
			if err != nil {
				t.Fatal(err)
			}
			old := sourceRoot
			sourceRoot = abs
			defer func() { sourceRoot = old }()
			p := filepath.Join(t.TempDir(), "winter-mixed.json")
			if err := os.WriteFile(p, out, 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := loadWinterMixedInputs(p); (err == nil) != (edit == nil) {
				t.Fatalf("drift %s: %v", name, err)
			}
		})
	}
}

// --- The winter path and rail criterion on the installed 4x tiles ---

func trackLab(c color.NRGBA) [3]float64 {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	f := func(t float64) float64 {
		if t > 216.0/24389 {
			return math.Cbrt(t)
		}
		return (24389.0/27*t + 16) / 116
	}
	r, g, b := lin(c.R), lin(c.G), lin(c.B)
	x := (.4124564*r + .3575761*g + .1804375*b) / .95047
	y := .2126729*r + .7151522*g + .0721750*b
	z := (.0193339*r + .1191920*g + .9503041*b) / 1.08883
	return [3]float64{116*f(y) - 16, 500 * (f(x) - f(y)), 200 * (f(y) - f(z))}
}

func winterDeltaE(a, b [3]float64) float64 {
	return math.Hypot(math.Hypot(a[0]-b[0], a[1]-b[1]), a[2]-b[2])
}

// summerGreen is the criterion's summer-green pixel: a* < -6 and C* > 12.
func trackSummerGreen(l [3]float64) bool { return l[1] < -6 && math.Hypot(l[1], l[2]) > 12 }

type winterLabMean struct {
	sum   [3]float64
	n     int
	green int
}

func (m *winterLabMean) add(l [3]float64) {
	for i := range l {
		m.sum[i] += l[i]
	}
	m.n++
	if trackSummerGreen(l) {
		m.green++
	}
}
func (m winterLabMean) mean() [3]float64 {
	return [3]float64{m.sum[0] / float64(m.n), m.sum[1] / float64(m.n), m.sum[2] / float64(m.n)}
}
func (m winterLabMean) greenShare() float64 { return float64(m.green) / float64(m.n) }

// winterTrackScene composites installed 4x tiles over the winter ground
// (grass-winter-v1 sampled like the ground tiles) and classifies pixels by
// the family profiles of the winter manifests.
type winterTrackScene struct {
	root     string
	ground   *image.NRGBA
	families map[string]connectionFamily
	variants map[string]connectionVariant
}

func newWinterTrackScene(t *testing.T) winterTrackScene {
	t.Helper()
	root := filepath.Join("..", "..")
	grass, err := loadPNG(filepath.Join(root, "art/terrain/grass/grass-winter-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	s := winterTrackScene{root: root, ground: image.NewNRGBA(image.Rect(0, 0, 384, 320)), families: map[string]connectionFamily{}, variants: map[string]connectionVariant{}}
	for y := range 320 {
		for x := range 384 {
			s.ground.SetNRGBA(x, y, sampleGrassMaterial(grass, connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}))
		}
	}
	for _, name := range []string{"connected-paths-winter-v2.json", "connected-rail-winter-v2.json"} {
		b, err := os.ReadFile(filepath.Join(root, "art/pipeline", name))
		if err != nil {
			t.Fatal(err)
		}
		var m connectionManifest
		if err := decodeConnectionJSON(b, &m); err != nil {
			t.Fatal(err)
		}
		for _, f := range m.Families {
			for _, v := range f.Variants {
				s.families[v.ID] = f
				s.variants[v.ID] = v
			}
		}
	}
	return s
}

func (s winterTrackScene) composite(t *testing.T, family, id string) *image.NRGBA {
	t.Helper()
	tile, err := loadPNG(filepath.Join(s.root, "assets/terrain", family, id+"-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	out := image.NewNRGBA(s.ground.Bounds())
	draw.Draw(out, out.Bounds(), s.ground, image.Point{}, draw.Src)
	draw.Draw(out, out.Bounds(), tile, image.Point{}, draw.Over)
	return out
}

// winterTrackBands measures one path or rail tile: the verge band (the earth
// border of the profile), the winter ground just beyond it and the surface
// well inside the core.
func (s winterTrackScene) bands(im *image.NRGBA, f connectionFamily, v connectionVariant) (verge, beside, surface winterLabMean) {
	g := sharedConnectionGeometry()
	var route connectionRoute
	if f.ID == "rail" {
		route = connectionRailRoutes(g, f, v).ballast
	} else {
		route = connectionRoutes(g, f, v)
	}
	for y := range 320 {
		for x := range 384 {
			p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
			if !connectionInside(g, p) || mixedEdgeDistance(g, p) < 1 {
				continue
			}
			var core, outer float64
			if f.ID == "path" {
				core, outer, _ = connectionPathProfile(g, f, route, p, p)
			} else {
				core = connectionRouteDistance(route, p) - f.Width/2
				outer = core - f.Border
			}
			l := trackLab(im.NRGBAAt(x, y))
			switch {
			case core > .25 && outer < -.1:
				verge.add(l)
			case outer > 1.5 && outer < 4:
				beside.add(l)
			case core < -1:
				surface.add(l)
			}
		}
	}
	return verge, beside, surface
}

// The road36_w surface the winter path surface must match.
func winterRoadSurface(t *testing.T, root string) [3]float64 {
	t.Helper()
	im, err := loadPNG(filepath.Join(root, "assets/terrain/road/road36_w-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	var m winterLabMean
	for y := 8 * 4; y < 72*4; y++ {
		for x := (48 - 9) * 4; x < (48+9)*4; x++ {
			m.add(trackLab(im.NRGBAAt(x, y)))
		}
	}
	return m.mean()
}

// railCentrelineDark is the share of both steel rail centre lines, sampled
// every quarter world pixel inside the hex, with L* <= 40.
func railCentrelineDark(im *image.NRGBA, f connectionFamily, v connectionVariant) float64 {
	g := sharedConnectionGeometry()
	dark, n := 0, 0
	for _, seg := range connectionRailRoutes(g, f, v).rails.segments {
		d := seg.b.sub(seg.a)
		steps := int(math.Ceil(d.length() * 4))
		for i := range steps {
			p := seg.a.add(d.mul(float64(i) / float64(max(steps, 1))))
			if !connectionInside(g, p) || mixedEdgeDistance(g, p) < .5 {
				continue
			}
			n++
			if trackLab(im.NRGBAAt(int(p[0]*4), int(p[1]*4)))[0] <= 40 {
				dark++
			}
		}
	}
	return float64(dark) / float64(n)
}

// railSleeperPeriod autocorrelates L* along a straight track, one sample per
// 4x pixel, each the mean across the gauge between the rails, and returns the
// best lag within 2 world px and 1.5 sleeper pitches and the correlation at
// the pitch.
func railSleeperPeriod(im *image.NRGBA, f connectionFamily, v connectionVariant) (int, float64) {
	pitch := connectionRailDimsFor(f.Width).pitch
	g := sharedConnectionGeometry()
	a, b := g.Ports[v.Ports[0]-1], g.Ports[v.Ports[1]-1]
	d := b.sub(a)
	n := connectionRailNormal(connectionRailUnit(d))
	steps := int(d.length() * 4)
	var xs []float64
	for i := 8; i < steps-8; i++ {
		p := a.add(d.mul(float64(i) / float64(steps)))
		l := 0.
		for _, across := range []float64{-1.5, -1, -.5, 0, .5, 1, 1.5} {
			q := p.add(n.mul(across))
			l += trackLab(im.NRGBAAt(int(q[0]*4), int(q[1]*4)))[0] / 7
		}
		xs = append(xs, l)
	}
	mean := 0.
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	corr := func(k int) float64 {
		var num, den float64
		for i := range xs {
			den += (xs[i] - mean) * (xs[i] - mean)
			if i+k < len(xs) {
				num += (xs[i] - mean) * (xs[i+k] - mean)
			}
		}
		return num / den
	}
	best, lag := math.Inf(-1), 0
	for k := 8; k <= int(pitch*6); k++ {
		if c := corr(k); c > best {
			best, lag = c, k
		}
	}
	return lag, corr(int(pitch * 4))
}

// winterCrossingInfill collects the plank infill of a crossing: where the
// way (a path or road ID such as "path14_w") and the track overlap, away from
// the hex edge, the rounded corners and the steel.
func winterCrossingInfill(im *image.NRGBA, rail connectionFamily, track connectionVariant, way string) winterLabMean {
	g := sharedConnectionGeometry()
	f := connectionFamily{ID: "path", Width: 22, Border: 2}
	if strings.HasPrefix(way, "road") {
		f = connectionFamily{ID: "road", Width: 22, Border: 2, Markings: "ruts"}
	}
	var ports []int
	for _, r := range strings.TrimSuffix(strings.TrimLeft(way, "pathrod"), "_w") {
		ports = append(ports, int(r-'0'))
	}
	route := connectionRoutes(g, f, connectionVariant{ID: way, Ports: ports})
	r := connectionRailRoutes(g, rail, track)
	var m winterLabMean
	for y := range 320 {
		for x := range 384 {
			p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
			if mixedEdgeDistance(g, p) > 12.5 && connectionRouteDistance(route, p) < f.Width/2+.5 && connectionRouteDistance(r.ballast, p) < mixedRailHalf(rail)-.5 && connectionRouteDistance(r.rails, p) > r.dims.steel/2+.5 {
				m.add(trackLab(im.NRGBAAt(x, y)))
			}
		}
	}
	return m
}

// Criterion for the winter paths, rails, crossings and footbridges (PLAN.md
// L7.4) on the installed 4x tiles over the winter ground.
func TestWinterPathsAndRailsMeetCriterion(t *testing.T) {
	s := newWinterTrackScene(t)
	road := winterRoadSurface(t, s.root)
	t.Logf("road36_w surface L*a*b* %.1f %.1f %.1f", road[0], road[1], road[2])
	var paths winterLabMean // the surface of every winter path
	for _, id := range slices.Sorted(func(yield func(string) bool) {
		for id := range s.variants {
			if !yield(id) {
				return
			}
		}
	}) {
		f, v := s.families[id], s.variants[id]
		if f.Width != 22 || f.Border != 2 {
			t.Fatalf("%s: profile %v/%v, want the road's 22/2", id, f.Width, f.Border)
		}
		im := s.composite(t, f.ID, id)
		verge, beside, surface := s.bands(im, f, v)
		dv := winterDeltaE(verge.mean(), beside.mean())
		line := fmt.Sprintf("%-9s verge ΔE %4.1f green %.2f%%", id, dv, 100*verge.greenShare())
		if dv > 8 || verge.greenShare() >= .01 {
			t.Errorf("%s: verge ΔE %.1f (want <= 8), summer green %.2f%% (want < 1%%)", id, dv, 100*verge.greenShare())
		}
		switch f.ID {
		case "path":
			for i := range paths.sum {
				paths.sum[i] += surface.sum[i]
			}
			paths.n += surface.n
			ds := winterDeltaE(surface.mean(), road)
			line += fmt.Sprintf("  surface L*a*b* %.1f %.1f %.1f ΔE %.1f", surface.mean()[0], surface.mean()[1], surface.mean()[2], ds)
			if ds > 8 {
				t.Errorf("%s: surface ΔE %.1f from road36_w (want <= 8)", id, ds)
			}
		case "rail":
			dark := railCentrelineDark(im, f, v)
			line += fmt.Sprintf("  rails dark %.1f%%", 100*dark)
			if dark < .9 {
				t.Errorf("%s: rail centre lines L* <= 40 along %.1f%% (want >= 90%%)", id, 100*dark)
			}
			if len(v.Ports) == 2 && (v.Ports[1]-v.Ports[0] == 3) {
				lag, c := railSleeperPeriod(im, f, v)
				line += fmt.Sprintf("  sleeper lag %d r %.2f", lag, c)
				// The interior pitch is fitted to the arc length, so on a diagonal
				// it may run a quarter world pixel long.
				if pitch := int(connectionRailDimsFor(f.Width).pitch * 4); lag < pitch || lag > pitch+1 || c < .3 {
					t.Errorf("%s: sleeper autocorrelation peak at lag %d (r %.2f at the summer period), want lag %d (+1) with r >= .3", id, lag, c, pitch)
				}
			}
		}
		t.Log(line)
	}
	path := paths.mean()
	t.Logf("winter path surface L*a*b* %.1f %.1f %.1f", path[0], path[1], path[2])
	// Crossings and footbridges: their rails stay dark and no summer green
	// shows anywhere on the tile. A crossing's plank infill reads as snowed
	// over: within ΔE 10 of the winter surface crossing the track, chroma <= 8.
	for _, v := range mixedRoster() {
		if v.Kind == "road-bridge" {
			continue
		}
		family := "cross"
		if v.Kind == "footbridge" {
			family = "bridge"
		}
		tile, err := loadPNG(filepath.Join(s.root, "assets/terrain", family, v.RuntimeID+"_w-4x.png"))
		if err != nil {
			t.Fatal(err)
		}
		var all winterLabMean
		for y := range 320 {
			for x := range 384 {
				if c := tile.NRGBAAt(x, y); c.A > 0 {
					all.add(trackLab(c))
				}
			}
		}
		line := fmt.Sprintf("%-9s green %.2f%%", v.RuntimeID+"_w", 100*all.greenShare())
		if all.greenShare() >= .01 {
			t.Errorf("%s_w: summer green %.2f%%", v.RuntimeID, 100*all.greenShare())
		}
		if v.Kind == "crossing" {
			rail := v.Layers[1] + "_w"
			im := s.composite(t, family, v.RuntimeID+"_w")
			dark := railCentrelineDark(im, s.families[rail], s.variants[rail])
			line += fmt.Sprintf("  rails dark %.1f%%", 100*dark)
			if dark < .9 {
				t.Errorf("%s_w: rail centre lines L* <= 40 along %.1f%%", v.RuntimeID, 100*dark)
			}
			infill := winterCrossingInfill(im, s.families[rail], s.variants[rail], v.Layers[0]+"_w")
			target, surface := path, "winter path"
			if strings.HasPrefix(v.Layers[0], "road") {
				target, surface = road, "road36_w"
			}
			m := infill.mean()
			de, chroma := winterDeltaE(m, target), math.Hypot(m[1], m[2])
			line += fmt.Sprintf("  infill L*a*b* %.1f %.1f %.1f ΔE %.1f from the %s surface, C* %.1f", m[0], m[1], m[2], de, surface, chroma)
			if infill.n < 100 || de > 10 || chroma > 8 {
				t.Errorf("%s_w: plank infill (%d px) ΔE %.1f from the %s surface (want <= 10), C* %.1f (want <= 8)", v.RuntimeID, infill.n, de, surface, chroma)
			}
		}
		t.Log(line)
	}
}
