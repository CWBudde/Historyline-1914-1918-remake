package mapcompose_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"math"
	"os"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
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

func campaignLevel(t *testing.T, name string) *level.Level {
	t.Helper()
	b, err := fs.ReadFile(levels.Campaign, "campaign/"+name+".json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := level.Parse(b)
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
// broke them with asphalt and dashes. Path bridges draw the footbridge.
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
	// The rails of tile/rail36 lie 2.375 px left and 2.125 px right of the
	// hex centre (logical).
	rails := []float64{-2.375, 2.125}
	const density = 4
	for _, site := range []struct {
		level string
		at    hexgrid.Coord
	}{
		{"level20_cp", hexgrid.Coord{Line: 46, Col: 9}},
		{"level20_cp", hexgrid.Coord{Line: 48, Col: 9}},
		{"level2_cp", hexgrid.Coord{Line: 10, Col: 4}},
		{"level13_2p", hexgrid.Coord{Line: 31, Col: 24}}, // summer
	} {
		w, c := window(campaignLevel(t, site.level), site.at, 7, 5)
		m, err := r.Render(w, density, false)
		if err != nil {
			t.Fatal(err)
		}
		cx, cy := hexgrid.Center(c)
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
				t.Errorf("%s %v rail at %+.3f: dark steel on %.0f %% of the deck, want 90 %%", site.level, site.at, off, 100*frac)
			} else {
				t.Logf("%s %v rail at %+.3f: dark steel on %.0f %% of the deck", site.level, site.at, off, 100*frac)
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
