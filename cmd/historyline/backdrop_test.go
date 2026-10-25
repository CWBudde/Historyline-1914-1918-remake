package main

import (
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

func TestBackdropScale(t *testing.T) {
	const ww, wh = 1090, 1920 // PULSE's world
	for _, tc := range []struct {
		name  string
		w, h  int
		scale float64
		err   string
	}{
		{"world size", 1090, 1920, 1, ""},
		{"4x", 4360, 7680, 4, ""},
		{"painted 1.409x", 1536, 2706, 1536.0 / 1090, ""},
		{"half", 545, 960, 0.5, ""},
		{"aspect 0.5 % off", 1090, 1930, 1, ""},
		{"aspect 2 % off", 1090, 1960, 0, "does not fit"},
		{"landscape", 1920, 1090, 0, "does not fit"},
		{"empty", 0, 1920, 0, "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := backdropScale(tc.w, tc.h, ww, wh)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("err %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil || math.Abs(s-tc.scale) > 1e-9 {
				t.Fatalf("scale %v, %v; want %v", s, err, tc.scale)
			}
		})
	}
}

func TestBackdropTiles(t *testing.T) {
	for _, tc := range []struct {
		name       string
		w, h       int
		cols, rows int
	}{
		{"1x in one", 1090, 1920, 1, 1},
		{"painted in two", 1536, 2706, 1, 2},
		{"4x", 4360, 7680, 3, 4},
		{"exact step", 2 * (backdropTileSize - 2*backdropPad), backdropTileSize - 2*backdropPad, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tiles := backdropTiles(tc.w, tc.h, backdropTileSize, backdropPad)
			if len(tiles) != tc.cols*tc.rows {
				t.Fatalf("%d tiles, want %dx%d", len(tiles), tc.cols, tc.rows)
			}
			all := image.Rect(0, 0, tc.w, tc.h)
			area := 0
			for i, a := range tiles {
				if a.src.Dx() > backdropTileSize || a.src.Dy() > backdropTileSize {
					t.Errorf("tile %d holds %v, over %d", i, a.src, backdropTileSize)
				}
				if !a.core.In(a.src) || !a.src.In(all) {
					t.Errorf("tile %d: core %v, src %v, picture %v", i, a.core, a.src, all)
				}
				if want := a.core.Inset(-backdropPad).Intersect(all); a.src != want {
					t.Errorf("tile %d holds %v, want %v", i, a.src, want)
				}
				for j, b := range tiles[i+1:] {
					if a.core.Overlaps(b.core) {
						t.Errorf("cores %d and %d overlap: %v %v", i, i+1+j, a.core, b.core)
					}
				}
				area += a.core.Dx() * a.core.Dy()
			}
			if area != tc.w*tc.h {
				t.Fatalf("cores cover %d pixels, want %d", area, tc.w*tc.h)
			}
		})
	}
}

// TestBackdropOnLevel places pictures on PULSE's 8x47 grid: a hex centre lands
// on the picture pixel scaled from its world point.
func TestBackdropOnLevel(t *testing.T) {
	grid := hexgrid.Grid{Width: 8, Lines: 47}
	for _, tc := range []struct {
		w, h int
		ok   bool
	}{{1090, 1920, true}, {4360, 7680, true}, {1536, 2706, true}, {1536, 1536, false}} {
		b, err := (&backdropImage{w: tc.w, h: tc.h}).on(grid)
		if (err == nil) != tc.ok {
			t.Fatalf("%dx%d: err %v, want ok %v", tc.w, tc.h, err, tc.ok)
		}
		if !tc.ok {
			continue
		}
		last := hexgrid.Coord{Line: 46, Col: 7}
		x, y := hexgrid.Center(last)
		px, py := x*b.scale, y*b.scale
		// The last hex's centre lies half a hex in from the picture's corner.
		if math.Abs(float64(tc.w)-px-48*b.scale) > 1 || math.Abs(float64(tc.h)-py-40*b.scale) > 1 {
			t.Fatalf("%dx%d: centre of %v at picture (%.1f, %.1f)", tc.w, tc.h, last, px, py)
		}
	}
}

// In the slim frame of a backdrop the human panel picks every playable hex
// whose centre lies in its opening, also those the plate frame's teeth cover,
// and nothing on the frame.
func TestSlimPick(t *testing.T) {
	uncovered := 0 // picked hexes the plate frame covers, over all windows
	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"1800x1000", 1800, 1000},
		{"1024x768", 1024, 768},
		{"1920x1080", 1920, 1080},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, d := newSplitTest(t, passPlanner{}, tc.w, tc.h)
			g.split.frame.Slim = true
			// A view inside the map, with hexes beyond the block on every side.
			s := g.split
			pair, row := clampView(g.playGrid(), s.screen.Cols, s.screen.Rows, 2, 3)
			s.cams[d.side].moveTo(pair, row, true)
			g.syncSplitCam()
			v := g.splitView(d.side)
			o := g.split.screen.SlimOpening(d.side)
			play := g.playGrid()
			for line := v.Row*2 - 2; line < (v.Row+v.Rows)*2+3; line++ {
				for col := v.Pair - 1; col < v.Pair+(v.Cols+1)/2+2; col++ {
					c := hexgrid.Coord{Line: line, Col: col}
					sx, sy := v.WorldToScreen(hexgrid.Center(c))
					want := play.Contains(c) && inRect(o, sx, sy)
					if got := g.panelShows(d.side, c); got != want {
						t.Fatalf("%v at %.0f,%.0f: shown %v, want %v", c, sx, sy, got, want)
					}
					if !want {
						continue
					}
					if got, ok := g.pickSplit(int(sx), int(sy)); !ok || got != c {
						t.Fatalf("pick at the centre of %v: %v %v", c, got, ok)
					}
					if !viewShows(v, play, c) {
						uncovered++
					}
				}
			}
			// The frame and the status strip pick nothing.
			st := g.split.screen.Status[d.side]
			for _, p := range [][2]int{{o.X - 1, o.Y + o.H/2}, {o.X + o.W/2, o.Y - 1}, {o.Right(), o.Y + o.H/2}, {st.X + st.W/2, st.Y + st.H/2}} {
				if c, ok := g.pickSplit(p[0], p[1]); ok {
					t.Errorf("frame pixel %v picks %v", p, c)
				}
			}
			// The plate frame keeps picking through its openings only.
			g.split.frame.Slim = false
			for line := v.Row * 2; line < (v.Row+v.Rows)*2; line++ {
				for col := v.Pair; col < v.Pair+(v.Cols+1)/2; col++ {
					c := hexgrid.Coord{Line: line, Col: col}
					if g.panelShows(d.side, c) != viewShows(v, play, c) {
						t.Fatalf("plate frame: %v shown %v", c, g.panelShows(d.side, c))
					}
				}
			}
		})
	}
	if uncovered == 0 {
		t.Error("no window picks a hex the plate frame covers")
	}
}

// In the slim frame of a backdrop each panel's opening ends on the map at every
// resting place and every zoom level it allows, and the wheel stops zooming out
// where the map would no longer fill the opening.
func TestSlimNoGap(t *testing.T) {
	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"1920x1080", 1920, 1080},
		{"1024x768", 1024, 768},
		{"1800x1000", 1800, 1000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := newSplitTest(t, passPlanner{}, tc.w, tc.h)
			ww, wh := maprender.Size(g.grid)
			g.backdrop = &backdrop{ww: float64(ww), wh: float64(wh)}
			g.layoutSplit()
			s := g.split
			if !s.frame.Slim {
				t.Fatal("a backdrop lays out the slim frame")
			}
			g.zoomSplit(-len(dosscreen.Zooms))
			base := dosscreen.Layout(tc.w, tc.h, 1)
			floor := base.FillZoom(g.slimWorld().X1, g.slimWorld().Y1)
			if s.screen.Zoom != floor {
				t.Fatalf("zoomed out to level %d, want the floor %d", s.screen.Zoom, floor)
			}
			world := g.slimWorld()
			for z := floor; z <= base.ClampZoom(len(dosscreen.Zooms)); z++ {
				s.zoom = z
				g.layoutSplit()
				for side := range 2 {
					plo, phi := pairRange(g.playGrid(), s.screen.Cols)
					rlo, rhi := rowRange(g.playGrid(), s.screen.Rows)
					for pair := plo; pair <= phi; pair++ {
						for row := rlo; row <= rhi; row++ {
							s.cams[side].moveTo(pair, row, true)
							cam := g.splitCamera(side)
							o := s.screen.SlimOpening(side)
							x0, y0 := cam.ScreenToWorld(float64(o.X), float64(o.Y))
							x1, y1 := cam.ScreenToWorld(float64(o.Right()), float64(o.Bottom()))
							const eps = 1e-6
							if x0 < world.X0-eps || y0 < world.Y0-eps || x1 > world.X1+eps || y1 > world.Y1+eps {
								t.Fatalf("zoom %d side %d at %d,%d: opening sees %.1f,%.1f-%.1f,%.1f beyond %+v",
									z, side, pair, row, x0, y0, x1, y1, world)
							}
						}
					}
				}
			}
		})
	}
}

// PULSE's scenario file names its painting: the level carries it as a path
// below the levels directory, the embedded copy loads and fits the map, and a
// battle on it draws the painting unless -backdrop off.
func TestLevelBackdrop(t *testing.T) {
	lvl, err := loadLevel(t.TempDir(), "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	if got := lvl.Extra.Backdrop; got != "monthly/1914-08.jpg" {
		t.Fatalf("backdrop %q", got)
	}
	bd, err := levelBackdrop(t.TempDir(), lvl.Extra.Backdrop)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bd.on(lvl.Grid()); err != nil {
		t.Fatal(err)
	}
	if _, err := levelBackdrop("", "nowhere/x.jpg"); err == nil {
		t.Error("a backdrop outside the scenario directories loaded")
	}
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, off := range []bool{false, true} {
		gm, err := newMatch(matchConfig{lvl: lvl, set: set, player: state.CentralPowers, rules: "dos", opponent: "pass", noBackdrop: off}, &meter{})
		if err != nil {
			t.Fatal(err)
		}
		if (gm.backdrop != nil) == off {
			t.Errorf("-backdrop off %v: backdrop %v", off, gm.backdrop != nil)
		}
	}
}

// A hex's minimap colour comes from the picture: a picture of one colour tints
// every hex of PULSE's grid that colour; one painted red left and green right of
// world x = 500 tints a hex well inside each half beyond that half's colour,
// away from the map's mean, and one across the line between them.
func TestBackdropTints(t *testing.T) {
	grid := hexgrid.Grid{Width: 8, Lines: 47}
	const scale = 2.0 // picture pixels per world pixel
	red, green := color.RGBA{200, 40, 20, 255}, color.RGBA{30, 160, 50, 255}
	paint := func(split int) *image.RGBA {
		pic := image.NewRGBA(image.Rect(0, 0, 2180, 3840))
		for y := range pic.Rect.Dy() {
			for x := range pic.Rect.Dx() {
				c := red
				if x >= split {
					c = green
				}
				pic.SetRGBA(x, y, c)
			}
		}
		return pic
	}
	tint := hexTints(reduce(paint(1<<20), backdropThumb), scale/backdropThumb, grid)
	if len(tint) != 8*24+7*23 {
		t.Fatalf("%d hexes tinted", len(tint))
	}
	for c, got := range tint {
		if got != red {
			t.Fatalf("one colour: %v tinted %v", c, got)
		}
	}
	tint = hexTints(reduce(paint(1000), backdropThumb), scale/backdropThumb, grid)
	left, across, right := tint[hexgrid.Coord{Line: 10, Col: 1}], tint[hexgrid.Coord{Line: 10, Col: 3}], tint[hexgrid.Coord{Line: 10, Col: 6}]
	if left.R < red.R || left.G > red.G || right.G < green.G || right.R > green.R {
		t.Errorf("left %v, right %v: not stretched beyond %v and %v", left, right, red, green)
	}
	// Hex (10, 3) spans x 426..522: partly red, partly green.
	if across.R <= right.R || across.R >= left.R || across.G <= left.G || across.G >= right.G {
		t.Errorf("hex across the line %v, not between %v and %v", across, left, right)
	}
}
