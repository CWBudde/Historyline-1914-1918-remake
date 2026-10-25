package mapcompose_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// catalog returns the IDs of assets/index.json at 4x and the 4x entries.
func catalog(t *testing.T) (func(string) bool, map[string]maprender.Entry) {
	t.Helper()
	raw, err := os.ReadFile("../../assets/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []maprender.Entry `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	entries := map[string]maprender.Entry{}
	for _, e := range index.Assets {
		if e.Density == 4 {
			entries[e.ID] = e
		}
	}
	return func(id string) bool { _, ok := entries[id]; return ok }, entries
}

// bridgeAt composes a 3x3 map of ground with one hex carrying infra and
// returns that hex's standing items.
func bridgeAt(ground string, infra []string, has func(string) bool) []string {
	g := hexgrid.Grid{Width: 3, Lines: 3}
	c := hexgrid.Coord{Line: 1, Col: 1}
	s := mapcompose.Compose(g, func(x hexgrid.Coord) mapcompose.Hex {
		if x == c {
			return mapcompose.Hex{Terrain: ground, Infra: infra}
		}
		return mapcompose.Hex{Terrain: ground}
	}, has)
	return s.Cells[c.Line][c.Col].Standing
}

// window cuts the hexes around c out of l: an odd number of lines starting on
// an even line, so the hex layout is unchanged, and c's position in it.
func window(l *level.Level, c hexgrid.Coord, lines, cols int) (*level.Level, hexgrid.Coord) {
	l0 := max(0, c.Line-lines/2) &^ 1
	c0 := max(0, c.Col-cols/2)
	w := &level.Level{MapName: l.MapName, MapNo: l.MapNo, Player: l.Player}
	for li := l0; li < min(len(l.Lines), l0+lines); li++ {
		row := l.Lines[li]
		n := cols - (li-l0)&1
		w.Lines = append(w.Lines, slices.Clone(row[c0:min(len(row), c0+n)]))
	}
	return w, hexgrid.Coord{Line: c.Line - l0, Col: c.Col - c0}
}

// campaignLevel is a setup of a shipped scenario, "monthly/1915-02-venom:cp".
func campaignLevel(t *testing.T, name string) *level.Level {
	t.Helper()
	file, mode, _ := strings.Cut(name, ":")
	l, err := scenario.ShippedLevel(file, scenario.Mode(mode))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// lab converts an sRGB colour to CIE L*a*b* (D65).
func lab(c color.NRGBA) (l, a, b float64) {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	x := (0.4124*r + 0.3576*g + 0.1805*bl) / 0.95047
	y := 0.2126*r + 0.7152*g + 0.0722*bl
	z := (0.0193*r + 0.1192*g + 0.9505*bl) / 1.08883
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	return 116*f(y) - 16, 500 * (f(x) - f(y)), 200 * (f(y) - f(z))
}

// Railway bridges draw a rail deck: on the composed map the two rails of the
// approaches run on across the deck (dark steel within a logical pixel of each
// rail's centreline on at least 90 % of the 48 px deck), where the road deck
// broke them with asphalt and dashes. The deck carries the wide track: its
// rails (each the median over the rows of the darkest pixel near it) lie 9 ±
// 0.5 logical px apart, their middle within 0.25 of the approach rails', and
// at each end of the deck (its outer 8 px) each within 0.5 of the approach
// rail at that end (8 to 16 px beyond it). The rails on the deck are as dark
// and as thin as the approach track's (on the track 32 to 40 px from the
// centre), after the pack grade: per row the darkest half logical pixel
// within 0.75 of a rail's centreline has a mean L* within 5 of the track's,
// and the dark run across each rail (the pixels from that darkest one on with
// L* at most the track's + 12, within 2 of the centreline) is, as the median
// over the rows, within a quarter logical pixel as wide as the track's. Path
// bridges draw the footbridge.
func TestStandingRailDeck(t *testing.T) {
	has, _ := catalog(t)
	if got := bridgeAt("grass", []string{"stream14", "rBridge36", "rail36"}, has); !slices.Equal(got, []string{"standing/rBridge36-rail"}) {
		t.Errorf("rail bridge standing %v, want the rail deck", got)
	}
	if got := bridgeAt("grass", []string{"stream14", "rBridge36", "road36", "rail36"}, has); !slices.Equal(got, []string{"standing/rBridge36"}) {
		t.Errorf("road and rail bridge standing %v, want the road deck", got)
	}
	if got := bridgeAt("grass", []string{"stream14", "rBridge36", "path36"}, has); !slices.Equal(got, []string{"standing/bridge36"}) {
		t.Errorf("path bridge standing %v, want the footbridge", got)
	}
	if got := bridgeAt("grass", []string{"stream25", "rBridge36", "road36"}, has); !slices.Equal(got, []string{"standing/rBridge36"}) {
		t.Errorf("road bridge standing %v", got)
	}

	r, err := maprender.New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	// The wide track's rails lie 4.5 logical px either side of the hex centre;
	// each rail is found within railReach of there, on the approaches and on
	// the deck.
	nominal := [2]float64{-4.5, 4.5}
	const density, railReach = 4, 1.5
	for _, site := range []struct {
		level string
		at    hexgrid.Coord
	}{
		{"monthly/1917-10-atlas:cp", hexgrid.Coord{Line: 46, Col: 9}},
		{"monthly/1917-10-atlas:cp", hexgrid.Coord{Line: 48, Col: 9}},
		{"monthly/1914-10-civil:cp", hexgrid.Coord{Line: 10, Col: 4}},
		{"skirmish/13-grand:2p", hexgrid.Coord{Line: 31, Col: 24}}, // summer
	} {
		w, c := window(campaignLevel(t, site.level), site.at, 7, 5)
		m, err := r.Render(w, density, false)
		if err != nil {
			t.Fatal(err)
		}
		cx, cy := hexgrid.Center(c)
		name := fmt.Sprintf("%s %v", site.level, site.at)
		lstar := func(x, y int) float64 { l, _, _ := lab(m.Image.NRGBAAt(x, y)); return l }
		rows := func(spans ...[2]float64) []int {
			var ys []int
			for _, span := range spans {
				for y := int((cy + span[0]) * density); y < int((cy+span[1])*density); y++ {
					ys = append(ys, y)
				}
			}
			return ys
		}
		deck := rows([2]float64{-24, 24})
		deckEnds := [2][]int{rows([2]float64{-24, -16}), rows([2]float64{16, 24})}
		approachEnds := [2][]int{rows([2]float64{-40, -32}), rows([2]float64{32, 40})}
		approach := append(slices.Clone(approachEnds[0]), approachEnds[1]...)
		// darkest is, in row y, the darker x of the darkest two neighbouring
		// pixels within reach logical of off, and their mean L*.
		darkest := func(off, reach float64, y int) (int, float64) {
			x0 := int(math.Round((cx + off - reach) * density))
			x1 := int(math.Round((cx + off + reach) * density))
			bx, best := x0, math.Inf(1)
			for x := x0; x < x1; x++ {
				if v := (lstar(x, y) + lstar(x+1, y)) / 2; v < best {
					bx, best = x, v
				}
			}
			if lstar(bx+1, y) < lstar(bx, y) {
				bx++
			}
			return bx, best
		}
		median := func(v []float64) float64 {
			if len(v) == 0 {
				return math.NaN()
			}
			slices.Sort(v)
			return v[len(v)/2]
		}
		// railAt is the median over the rows ys of the middle of the dark run
		// (L* within 10 of its darkest pixel) around the darkest pixel within
		// reach of off, as a logical offset from the hex centre.
		railAt := func(off, reach float64, ys []int) float64 {
			var xs []float64
			for _, y := range ys {
				x, _ := darkest(off, reach, y)
				l := lstar(x, y) + 10
				a, b := x, x
				for a > x-8 && lstar(a-1, y) <= l {
					a--
				}
				for b < x+8 && lstar(b+1, y) <= l {
					b++
				}
				xs = append(xs, float64(a+b+1)/2/density-cx)
			}
			return median(xs)
		}
		// The approach rails, from both ends together and from each end.
		var rails [2]float64
		var railEnds [2][2]float64
		for i, off := range nominal {
			rails[i] = railAt(off, railReach, approach)
			for e := range 2 {
				railEnds[e][i] = railAt(off, railReach, approachEnds[e])
			}
		}
		// The deck rails, over the deck and at each end of it.
		var deckRails [2]float64
		var deckRailEnds [2][2]float64
		for i := range rails {
			deckRails[i] = railAt(rails[i], railReach, deck)
			for e := range 2 {
				deckRailEnds[e][i] = railAt(railEnds[e][i], railReach, deckEnds[e])
			}
		}
		gauge := deckRails[1] - deckRails[0]
		middle := (deckRails[0]+deckRails[1])/2 - (rails[0]+rails[1])/2
		msg := fmt.Sprintf("%s: approach rails at %+.2f %+.2f, deck rails at %+.2f %+.2f: gauge %.2f (want 9 ± 0.5), middle %+.2f off the approach's (want within 0.25)", name, rails[0], rails[1], deckRails[0], deckRails[1], gauge, middle)
		if math.Abs(gauge-9) > 0.5 || math.Abs(middle) > 0.25 {
			t.Error(msg)
		} else {
			t.Log(msg)
		}
		for e, end := range []string{"upper", "lower"} {
			for i := range rails {
				d := deckRailEnds[e][i] - railEnds[e][i]
				msg := fmt.Sprintf("%s: %s end, rail %d on the deck at %+.2f, on the approach at %+.2f (%+.2f, want within 0.5)", name, end, i, deckRailEnds[e][i], railEnds[e][i], d)
				if math.Abs(d) > 0.5 {
					t.Error(msg)
				} else {
					t.Log(msg)
				}
			}
		}
		mean := func(ys []int) float64 {
			var s float64
			for _, off := range rails {
				for _, y := range ys {
					_, v := darkest(off, 0.75, y)
					s += v
				}
			}
			return s / float64(len(rails)*len(ys))
		}
		heads, track := mean(deck), mean(approach)
		if math.Abs(heads-track) > 5 {
			t.Errorf("%s: rails on the deck L* %.1f, the approach track's rails %.1f, want within 5", name, heads, track)
		} else {
			t.Logf("%s: rails on the deck L* %.1f, the approach track's rails %.1f", name, heads, track)
		}
		// width is the median over the rows ys of the dark run across the rail
		// at off, in logical pixels, and the share of the rows it is measured
		// on: the rows where the run ends within 2 logical of the rail on both
		// sides, so the rail lies on a lighter deck (a dark sleeper as wide as
		// the window measures no rail).
		width := func(off float64, ys []int) (float64, float64) {
			lo := int(math.Round((cx + off - 2) * density))
			hi := int(math.Round((cx + off + 2) * density))
			var ws []float64
			for _, y := range ys {
				x, _ := darkest(off, 0.75, y)
				if lstar(x, y) > track+12 {
					ws = append(ws, 0)
					continue
				}
				a, b := x, x
				for a > lo && lstar(a-1, y) <= track+12 {
					a--
				}
				for b < hi && lstar(b+1, y) <= track+12 {
					b++
				}
				if a > lo && b < hi {
					ws = append(ws, float64(b-a+1)/density)
				}
			}
			if len(ws) == 0 {
				return math.NaN(), 0
			}
			return median(ws), float64(len(ws)) / float64(len(ys))
		}
		for _, off := range rails {
			wd, sd := width(off, deck)
			wa, sa := width(off, approach)
			msg := fmt.Sprintf("%s rail at %+.2f: dark run %.2f px wide on the deck (%.0f %% of the rows), %.2f on the approach track (%.0f %%)", name, off, wd, 100*sd, wa, 100*sa)
			if !(math.Abs(wd-wa) <= 0.25) || sd < 0.25 || sa < 0.25 {
				t.Errorf("%s, want within 0.25 on at least 25 %% of the rows", msg)
			} else {
				t.Log(msg)
			}
		}
		for _, off := range rails {
			x0 := int(math.Round((cx + off - 1) * density))
			x1 := int(math.Round((cx + off + 1) * density))
			steel, rows := 0, 0
			for y := int((cy - 24) * density); y < int((cy+24)*density); y++ {
				rows++
				for x := x0; x <= x1; x++ {
					l, a, b := lab(m.Image.NRGBAAt(x, y))
					if l < 36 && math.Hypot(a, b) < 15 {
						steel++
						break
					}
				}
			}
			if frac := float64(steel) / float64(rows); frac < 0.9 {
				t.Errorf("%s rail at %+.2f: dark steel on %.0f %% of the deck, want 90 %%", name, off, 100*frac)
			} else {
				t.Logf("%s rail at %+.2f: dark steel on %.0f %% of the deck", name, off, 100*frac)
			}
		}
	}
}

// Winter maps draw the winter bridge decks; the road deck is packed snow with
// ruts like the winter road (no white paint in its middle 8 px, surface within
// ΔE 8 of road36_w).
func TestStandingWinterBridges(t *testing.T) {
	has, entries := catalog(t)
	for _, tt := range []struct {
		infra []string
		want  string
	}{
		{[]string{"stream25", "rBridge36", "road36"}, "standing/rBridge36_w"},
		{[]string{"stream36", "rBridge14", "road14"}, "standing/rBridge14_w"},
		{[]string{"stream36", "rBridge25", "road25"}, "standing/rBridge25_w"},
		{[]string{"stream36", "bridge25"}, "standing/bridge25_w"},
		{[]string{"stream25", "bridge36"}, "standing/bridge36_w"},
		{[]string{"stream14", "rBridge36", "rail36"}, "standing/rBridge36-rail_w"},
		{[]string{"stream14", "rBridge36", "path36"}, "standing/bridge36_w"},
	} {
		if got := bridgeAt("grass_w", tt.infra, has); !slices.Equal(got, []string{tt.want}) {
			t.Errorf("winter %v: standing %v, want %s", tt.infra, got, tt.want)
		}
		summer := tt.want[:len(tt.want)-2]
		if got := bridgeAt("grass", tt.infra, has); !slices.Equal(got, []string{summer}) {
			t.Errorf("summer %v: standing %v, want %s", tt.infra, got, summer)
		}
	}

	// The road decks, sampled along their axis: the middle 40 px of the 48 px
	// deck (clear of the end walls), the 18 px between the verges for the
	// surface and the middle 8 px for paint.
	const target, density = 53.4, 4
	for _, deck := range []struct {
		id    string
		angle float64 // axis, degrees from horizontal (y down)
	}{
		{"standing/rBridge36_w", 90},
		{"standing/rBridge14_w", -29.396},
		{"standing/rBridge25_w", 29.396},
	} {
		e, ok := entries[deck.id]
		if !ok {
			t.Errorf("%s: not in the catalog", deck.id)
			continue
		}
		raw, err := os.ReadFile("../../assets/" + e.Path)
		if err != nil {
			t.Fatal(err)
		}
		src, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		im := image.NewNRGBA(src.Bounds())
		for y := range im.Bounds().Dy() {
			for x := range im.Bounds().Dx() {
				im.Set(x, y, src.At(x, y))
			}
		}
		ux, uy := math.Cos(deck.angle*math.Pi/180), math.Sin(deck.angle*math.Pi/180)
		var sl, sa, sb float64
		n, paint := 0, 0
		for t10 := -200; t10 <= 200; t10++ {
			for s10 := -90; s10 <= 90; s10++ {
				along, across := float64(t10)/10, float64(s10)/10
				x := (e.Anchor[0] + along*ux - across*uy) * density
				y := (e.Anchor[1] + along*uy + across*ux) * density
				c := im.NRGBAAt(int(x), int(y))
				if c.A < 250 {
					continue
				}
				l, a, b := lab(c)
				sl, sa, sb, n = sl+l, sa+a, sb+b, n+1
				if math.Abs(across) <= 4 && l > 85 && math.Hypot(a, b) < 8 {
					paint++
				}
			}
		}
		if n == 0 {
			t.Errorf("%s: no opaque deck", deck.id)
			continue
		}
		l, a, b := sl/float64(n), sa/float64(n), sb/float64(n)
		de := math.Sqrt((l-target)*(l-target) + (a-0.9)*(a-0.9) + (b+0.3)*(b+0.3))
		t.Logf("%s: deck L* %.1f a* %.1f b* %.1f, ΔE %.1f to road36_w; %d paint-white pixels", deck.id, l, a, b, de, paint)
		if de > 8 || paint > 0 {
			t.Errorf("%s: deck ΔE %.1f (want ≤ 8), %d paint-white pixels in the middle (want 0)", deck.id, de, paint)
		}
	}
}

// decode4x reads the 4x image of the catalog entry id.
func decode4x(t *testing.T, entries map[string]maprender.Entry, id string) (*image.NRGBA, maprender.Entry) {
	t.Helper()
	e, ok := entries[id]
	if !ok {
		t.Fatalf("%s: not in the catalog", id)
	}
	raw, err := os.ReadFile("../../assets/" + e.Path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(src.Bounds())
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			im.Set(x, y, src.At(x, y))
		}
	}
	return im, e
}

// The footbridges carry the wide path: their timber deck is 22 ± 1 logical px
// across (between the dark seams where the handrails meet it, from the luma
// across the deck averaged over its middle 40 px) and the whole bridge with
// its handrails at most 28; where the path meets the deck ends (1 px inside
// them and 3 px further, over the abutments) the bridge covers the path's core
// (its opaque middle on tile/path) across at least 90 % of its width; the
// planks lie across: the luma along the deck, averaged across its middle 16
// px, varies at least twice as much (σ) as the luma across it averaged along
// it; and each winter form keeps its summer outline (alpha IoU ≥ 0.97).
func TestStandingFootbridgeWidth(t *testing.T) {
	_, entries := catalog(t)
	const density = 4
	for _, b := range []struct {
		id, path string
		angle    float64 // axis, degrees from horizontal (y down)
	}{
		{"standing/bridge36", "tile/path36", 90},
		{"standing/bridge25", "tile/path25", 29.396},
	} {
		ux, uy := math.Cos(b.angle*math.Pi/180), math.Sin(b.angle*math.Pi/180)
		// at returns the pixel of im at along, across logical px from the anchor.
		at := func(im *image.NRGBA, e maprender.Entry, along, across float64) color.NRGBA {
			x := (e.Anchor[0] + along*ux - across*uy) * density
			y := (e.Anchor[1] + along*uy + across*ux) * density
			return im.NRGBAAt(int(math.Floor(x)), int(math.Floor(y)))
		}
		luma := func(c color.NRGBA) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }
		path, pe := decode4x(t, entries, b.path)
		// The path's core: the opaque run across it at the hex centre.
		var core [2]float64
		for _, dir := range []float64{-1, 1} {
			s := 0.0
			for at(path, pe, 0, dir*(s+0.25)).A >= 250 {
				s += 0.25
			}
			core[(int(dir)+1)/2] = dir * s
		}
		summer := map[bool]*image.NRGBA{}
		for _, winter := range []bool{false, true} {
			id := b.id
			if winter {
				id += "_w"
			}
			im, e := decode4x(t, entries, id)
			summer[winter] = im
			// The whole bridge across: the opaque run through the axis, as the
			// median over its middle 40 px.
			var whole []float64
			// profile is the mean luma across the deck over its middle 40
			// px, over the opaque pixels; along the mean luma along it over
			// its middle 16 px.
			profile, opaque := map[float64]float64{}, map[float64]int{}
			along := map[float64]float64{}
			for a := -20.0; a <= 20; a += 0.25 {
				lo, hi := 0.0, 0.0
				for at(im, e, a, lo-0.25).A >= 128 && lo > -40 {
					lo -= 0.25
				}
				for at(im, e, a, hi+0.25).A >= 128 && hi < 40 {
					hi += 0.25
				}
				whole = append(whole, hi-lo+0.25)
				for s := -16.0; s <= 16; s += 0.25 {
					if c := at(im, e, a, s); c.A >= 128 {
						profile[s] += luma(c)
						opaque[s]++
					}
					if math.Abs(s) <= 8 {
						along[a] += luma(at(im, e, a, s)) / 65
					}
				}
			}
			slices.Sort(whole)
			// The deck's edges: the darkest seam in each side's band, where
			// the bridge is opaque.
			edge := func(from, to float64) float64 {
				best, bv := math.NaN(), math.Inf(1)
				for s := from; s <= to; s += 0.25 {
					if opaque[s] < 150 {
						continue
					}
					if v := profile[s] / float64(opaque[s]); v < bv {
						best, bv = s, v
					}
				}
				return best
			}
			deck := edge(9, 14) - edge(-14, -9)
			// Planks across: their seams survive the mean across the deck but
			// not the mean along it.
			sd := func(m map[float64]float64, from, to float64, n map[float64]int) float64 {
				var s, s2, k float64
				for x := from; x <= to; x += 0.25 {
					v := m[x]
					if n != nil {
						v /= float64(n[x])
					}
					s, s2, k = s+v, s2+v*v, k+1
				}
				return math.Sqrt(s2/k - (s/k)*(s/k))
			}
			planks := sd(along, -20, 20, nil) / sd(profile, -8, 8, opaque)
			// Cover of the path's core at the deck ends.
			cover := 1.0
			for _, along := range []float64{-25, -23, 23, 25} {
				n, in := 0, 0
				for s := core[0]; s <= core[1]; s += 0.25 {
					n++
					if at(im, e, along, s).A >= 128 {
						in++
					}
				}
				cover = math.Min(cover, float64(in)/float64(n))
			}
			// The whole bridge along its axis.
			lo, hi := 0.0, 0.0
			for at(im, e, lo-0.25, 0).A >= 128 && lo > -48 {
				lo -= 0.25
			}
			for at(im, e, hi+0.25, 0).A >= 128 && hi < 48 {
				hi += 0.25
			}
			msg := fmt.Sprintf("%s: deck %.2f across (want 22 ± 1), whole bridge %.2f across (want ≤ 28) and %.2f along, path core %.2f..%+.2f covered %.0f %% at the deck ends (want ≥ 90 %%), seam contrast along/across %.2f (want ≥ 2)",
				id, deck, whole[len(whole)/2], hi-lo+0.25, core[0], core[1], 100*cover, planks)
			// The winter deck's snow and trodden path hide its seams and
			// planks; its outline is the summer one's.
			if whole[len(whole)/2] > 28 || cover < 0.9 || !winter && (!(math.Abs(deck-22) <= 1) || !(planks >= 2)) {
				t.Error(msg)
			} else {
				t.Log(msg)
			}
		}
		// The winter outline.
		var and, or int
		s, w := summer[false], summer[true]
		for i := 3; i < len(s.Pix); i += 4 {
			a, b := s.Pix[i] >= 128, w.Pix[i] >= 128
			if a && b {
				and++
			}
			if a || b {
				or++
			}
		}
		if iou := float64(and) / float64(or); iou < 0.97 {
			t.Errorf("%s_w: alpha IoU %.3f against summer, want ≥ 0.97", b.id, iou)
		} else {
			t.Logf("%s_w: alpha IoU %.3f against summer", b.id, iou)
		}
	}
}
