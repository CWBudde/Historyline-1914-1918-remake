package dosscreen

import (
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
)

// isGunmetal is a near-black blue-grey: dark, no warm cast.
func isGunmetal(c color.RGBA) bool {
	return c.A == 255 && max(c.R, c.G, c.B) <= 0x40 && c.B >= c.R
}

// isBrass is a warm gold: red well above blue.
func isBrass(c color.RGBA) bool {
	return c.A == 255 && c.R >= 0x70 && int(c.R) > int(c.B)+0x30 && c.G > c.B
}

// The slim chrome is gunmetal with the panels' openings transparent, a brass
// hairline around each opening and strip, the status strips in gunmetal and a
// brass knob at the divider's foot, at every window size and UI scale.
func TestPaintSlimChrome(t *testing.T) {
	for _, tc := range []struct {
		name string
		w, h int
		ui   float64
	}{
		{"1080p", 1920, 1080, 1},
		{"1024x768", 1024, 768, 1},
		{"4K at 2x", 3840, 2160, 2},
		{"800p at 1.5x", 1280, 800, 1.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := Layout(tc.w, tc.h, tc.ui)
			img := PaintSlimChrome(s)
			l := hair(tc.ui)
			at := func(x, y int) color.RGBA { return img.RGBAAt(x, y) }
			for side := range 2 {
				o := s.SlimOpening(side)
				p := s.Panels[side]
				if o.W <= 0 || o.H <= 0 || o.X < p.X || o.Right() > p.Right() || o.Y < p.Y || o.Bottom() > p.Bottom() {
					t.Fatalf("side %d: opening %+v outside panel %+v", side, o, p)
				}
				mx, my := o.X+o.W/2, o.Y+o.H/2
				for _, pt := range [][2]int{{mx, my}, {o.X, o.Y}, {o.Right() - 1, o.Bottom() - 1}} {
					if c := at(pt[0], pt[1]); c.A != 0 {
						t.Errorf("side %d: opening pixel %v is %v, want transparent", side, pt, c)
					}
				}
				checks := []struct {
					what string
					x, y int
					ok   func(color.RGBA) bool
				}{
					{"rim above", mx, 0, isGunmetal},
					{"shadow above", mx, o.Y - 1, isDark},
					{"brass above", mx, o.Y - 2*l, isBrass},
					{"brass below", mx, o.Bottom() + l, isBrass},
					{"brass beside, divider side", [2]int{o.Right() + 2*l, o.X - 2*l - 1}[side], my, isBrass},
					{"brass beside, window side", [2]int{o.X - 2*l, o.Right() + 2*l - 1}[side], my, isBrass},
					{"outer rim", [2]int{0, s.W - 1}[side], my, isGunmetal},
				}
				wt := s.WellText(side)
				checks = append(checks, struct {
					what string
					x, y int
					ok   func(color.RGBA) bool
				}{"status strip", wt.X + wt.W/2, wt.Y + wt.H/2, isGunmetal})
				for _, c := range checks {
					if got := at(c.x, c.y); !c.ok(got) {
						t.Errorf("side %d %s at %d,%d: %v", side, c.what, c.x, c.y, got)
					}
				}
			}
			b := s.Bolt
			if c := at(b.X+b.W/2, b.Y+b.H/2); !isBrass(c) {
				t.Errorf("knob centre %v", c)
			}
			if c := at(s.Divider.X+s.Divider.W/2, s.Divider.H/2); !isGunmetal(c) {
				t.Errorf("divider post %v", c)
			}
			for y := range s.H {
				for x := range s.W {
					inOpen := false
					for side := range 2 {
						o := s.SlimOpening(side)
						inOpen = inOpen || x >= o.X && x < o.Right() && y >= o.Y && y < o.Bottom()
					}
					if a := img.RGBAAt(x, y).A; inOpen != (a == 0) || !inOpen && a != 255 {
						t.Fatalf("pixel %d,%d alpha %d, in an opening %v", x, y, a, inOpen)
					}
				}
			}
		})
	}
}

// Beyond the world a panel shows gunmetal, edged along the world's edge with
// the ring of lines, all inside the opening and outside the world.
func TestBeyondPieces(t *testing.T) {
	open := Rect{10, 10, 400, 300}
	for _, tc := range []struct {
		name  string
		world Rect
		brass [][2]int // pixels that must be brass
	}{
		{"world covers the panel", Rect{0, 0, 1000, 1000}, nil},
		{"world exactly the panel", open, nil},
		{"panel past the left end", Rect{60, -500, 1000, 2000}, [][2]int{{58, 100}, {57, 200}}},
		{"panel past the top and right", Rect{-100, 50, 400, 2000}, [][2]int{{100, 48}, {301, 200}}},
		{"world inside", Rect{100, 100, 100, 100}, [][2]int{{98, 150}, {150, 201}}},
		{"no world in view", Rect{2000, 2000, 10, 10}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ps := beyondPieces(open, tc.world, 1)
			in := tc.world.intersect(open)
			if in == open {
				if len(ps) != 0 {
					t.Fatalf("%d pieces where the world covers the panel", len(ps))
				}
				return
			}
			metal := 0
			kind := map[[2]int]pieceKind{}
			for _, p := range ps {
				if p.r.intersect(open) != p.r || p.r.empty() {
					t.Errorf("piece %+v leaves the opening", p)
				}
				if !p.r.intersect(tc.world).empty() {
					t.Errorf("piece %+v covers the world", p)
				}
				if p.kind == pieceMetal {
					metal += p.r.W * p.r.H
				}
				for y := p.r.Y; y < p.r.Bottom(); y++ {
					for x := p.r.X; x < p.r.Right(); x++ {
						if p.kind != pieceMetal {
							kind[[2]int{x, y}] = p.kind
						}
					}
				}
			}
			if want := open.W*open.H - in.W*in.H; metal != want {
				t.Errorf("gunmetal covers %d pixels, want %d", metal, want)
			}
			for _, b := range tc.brass {
				if k := kind[b]; k != pieceBrass && k != pieceBrassLit {
					t.Errorf("pixel %v is %v, want brass", b, k)
				}
			}
		})
	}
}

// The slim frame paints its chrome once per window size and no panel frames;
// switching the style repaints the chrome.
func TestFrameSlimCache(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	s := Layout(640, 400, 1)
	dst := ebiten.NewImage(640, 400)
	f := Frame{Slim: true}
	defer f.Release()
	views := [2]View{s.View(0, 5, 5), s.View(1, 9, 9)}
	f.Draw(dst, s, views, g)
	f.Draw(dst, s, [2]View{s.View(0, -1, 0), s.View(1, 8, 10)}, g)
	if f.painted != 1 || len(f.panels) != 0 {
		t.Fatalf("slim frame painted %d images, %d panels; want the chrome only", f.painted, len(f.panels))
	}
	f.Slim = false
	f.Draw(dst, s, views, g)
	if f.painted != 4 {
		t.Errorf("the plate frame after the slim one painted %d, want chrome and both panels", f.painted)
	}
	f.Slim = true
	f.Draw(dst, s, views, g)
	if f.painted != 5 {
		t.Errorf("back to the slim frame painted %d, want a new chrome", f.painted)
	}
	f.DrawBeyond(dst, s.SlimOpening(0), Rect{50, -10, 1000, 1000}, s.UI)
	if f.metal == nil {
		t.Error("DrawBeyond past the world made no gunmetal")
	}
}

// isDark is a near-black line.
func isDark(c color.RGBA) bool { return c.A == 255 && max(c.R, c.G, c.B) <= 0x20 }

// SlimFit moves a view looking past a map's end back onto the map by whole
// device pixels, leaves one inside it alone and centres a map smaller than
// the opening.
func TestSlimFit(t *testing.T) {
	s := Layout(1920, 1080, 1)
	world := WorldRect{X1: 1090, Y1: 1880}
	inside := func(v View, side int) bool {
		o := s.SlimOpening(side)
		x0, y0 := v.ScreenToWorld(float64(o.X), float64(o.Y))
		x1, y1 := v.ScreenToWorld(float64(o.Right()), float64(o.Bottom()))
		return x0 >= world.X0-eps && y0 >= world.Y0-eps && x1 <= world.X1+eps && y1 <= world.Y1+eps
	}
	whole := func(f float64) bool { return math.Abs(f-math.Round(f)) < 1e-6 }
	for side := range 2 {
		for _, at := range [][2]int{{0, 0}, {4, 0}, {0, 18}, {4, 18}} { // the map's corners, past each end
			v := s.SlimFit(s.View(side, at[0], at[1]), side, world)
			if !inside(v, side) {
				t.Errorf("side %d at %v: the opening looks past the map (shift %.2f, %.2f)", side, at, v.ShiftX, v.ShiftY)
			}
			if !whole(v.ShiftX*v.Scale) || !whole(v.ShiftY*v.Scale) {
				t.Errorf("side %d at %v: shift %.3f, %.3f is not whole device pixels", side, at, v.ShiftX, v.ShiftY)
			}
		}
	}
	// A view well inside a large map stays on its cells.
	big := WorldRect{X1: 10000, Y1: 10000}
	if v := s.SlimFit(s.View(0, 10, 10), 0, big); v.ShiftX != 0 || v.ShiftY != 0 {
		t.Errorf("inside the map: shift %.2f, %.2f", v.ShiftX, v.ShiftY)
	}
	// A map smaller than the opening sits in its middle.
	small := WorldRect{X1: 300, Y1: 200}
	v := s.SlimFit(s.View(1, 0, 0), 1, small)
	o := s.SlimOpening(1)
	x0, y0 := v.WorldToScreen(small.X0, small.Y0)
	x1, y1 := v.WorldToScreen(small.X1, small.Y1)
	if math.Abs(x0-float64(o.X)-(float64(o.Right())-x1)) > 1 || math.Abs(y0-float64(o.Y)-(float64(o.Bottom())-y1)) > 1 {
		t.Errorf("small map at %.0f,%.0f-%.0f,%.0f in opening %+v", x0, y0, x1, y1, o)
	}
}

// FillZoom is the lowest level at which the map still covers the openings:
// the level below it does not, and a map too small even at the fitted scale
// keeps DefaultZoom.
func TestFillZoom(t *testing.T) {
	for _, tc := range []struct {
		name   string
		w, h   int
		ww, wh float64
	}{
		{"PULSE 1080p", 1920, 1080, 1090, 1880},
		{"PULSE 1024x768", 1024, 768, 1090, 1880},
		{"wide map", 1920, 1080, 5000, 5000},
		{"tiny map", 1920, 1080, 200, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := Layout(tc.w, tc.h, 1)
			o := s.SlimOpening(0)
			covers := func(i int) bool {
				sc := s.zoomScale(i)
				return tc.ww*sc >= float64(o.W)-eps && tc.wh*sc >= float64(o.H)-eps
			}
			z := s.FillZoom(tc.ww, tc.wh)
			lo, _ := s.zoomRange()
			switch {
			case z > DefaultZoom || z < lo:
				t.Fatalf("level %d outside %d..%d", z, lo, DefaultZoom)
			case !covers(z) && z != DefaultZoom:
				t.Errorf("level %d does not cover the opening", z)
			case z > lo && covers(z-1):
				t.Errorf("level %d covers too; the floor is %d", z-1, z)
			}
		})
	}
}
