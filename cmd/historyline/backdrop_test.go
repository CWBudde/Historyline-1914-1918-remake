package main

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
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
