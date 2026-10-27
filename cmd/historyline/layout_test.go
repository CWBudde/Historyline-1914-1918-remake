package main

import (
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestUIScaleAndLogicalSize(t *testing.T) {
	tests := []struct {
		name          string
		w, h          int
		dpi           float64
		scale, lw, lh int
	}{
		{"narrow 400", 400, 700, 1, 1, 400, 700},
		{"default window", 1800, 1000, 1, 1, 1800, 1000},
		{"4K at 1x", 3840, 2160, 1, 3, 1280, 720},
		{"4K at 2x dip", 1920, 1080, 2, 3, 1280, 720},
		{"retina small", 800, 500, 2, 2, 800, 500},
		{"hidpi 1.5", 1200, 700, 1.5, 2, 900, 525},
		{"zero dpi", 1000, 600, 0, 1, 1000, 600},
		{"tiny", 0, 0, 1, 1, 1, 1},
	}
	for _, tc := range tests {
		if got := uiScale(tc.w, tc.h, tc.dpi); got != tc.scale {
			t.Errorf("%s: scale %d, want %d", tc.name, got, tc.scale)
		}
		if lw, lh := logicalSize(tc.w, tc.h, tc.dpi); lw != tc.lw || lh != tc.lh {
			t.Errorf("%s: logical %dx%d, want %dx%d", tc.name, lw, lh, tc.lw, tc.lh)
		}
	}
}

func TestNewView(t *testing.T) {
	v := newView(1920, 1080, 2)
	if v != (view{PhysW: 3840, PhysH: 2160, UI: 3, LogW: 1280, LogH: 720}) {
		t.Errorf("4K at 2x: %+v", v)
	}
	if v := newView(1200, 700, 1.5); v.PhysW != 1800 || v.PhysH != 1050 || v.UI != 2 || v.LogW != 900 || v.LogH != 525 {
		t.Errorf("1.5x: %+v", v)
	}
	if v := newView(0, 0, 0); v.PhysW < 1 || v.PhysH < 1 || v.UI != 1 {
		t.Errorf("degenerate: %+v", v)
	}
}

// The PLAN 2.4 check: the world point under the view centre and the picked hex
// survive zoom, window resize and display-scale changes, and the map keeps its
// apparent size when only the display scale changes.
func TestWorldPointSurvivesZoomResizeAndDPI(t *testing.T) {
	lvl, err := loadLevel("", "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	g := newMapGame(lvl, false)
	g.applyLayout(1800, 1000, 1)
	g.cam.CenterOn(g.cam.World.W/2, g.cam.World.H/2)
	pick := func() (wx, wy float64) {
		return g.cam.ScreenToWorld(float64(g.view.PhysW)/2, float64(g.view.PhysH)/2)
	}
	wantX, wantY := pick()
	want, ok := g.grid.PixelToHex(wantX, wantY)
	if !ok {
		t.Fatal("view centre is off the map")
	}
	steps := []struct {
		name string
		do   func()
	}{
		{"zoom in", func() { g.cam.ZoomAt(1.3, float64(g.view.PhysW)/2, float64(g.view.PhysH)/2) }},
		{"resize", func() { g.applyLayout(1400, 900, 1) }},
		{"dpi 2", func() { g.applyLayout(1400, 900, 2) }},
		{"dpi 1.5", func() { g.applyLayout(1100, 700, 1.5) }},
		{"zoom out", func() { g.cam.ZoomAt(0.8, float64(g.view.PhysW)/2, float64(g.view.PhysH)/2) }},
		{"back to 1", func() { g.applyLayout(1800, 1000, 1) }},
	}
	for _, s := range steps {
		s.do()
		if g.cam.View.W != float64(g.view.PhysW) || g.cam.Scale != float64(g.view.UI) {
			t.Fatalf("%s: camera view %v scale %v, layout %+v", s.name, g.cam.View, g.cam.Scale, g.view)
		}
		wx, wy := pick()
		if math.Abs(wx-wantX) > 1e-6 || math.Abs(wy-wantY) > 1e-6 {
			t.Fatalf("%s: centre world point %v,%v, want %v,%v", s.name, wx, wy, wantX, wantY)
		}
		cx, cy := hexgrid.Center(want)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		if got, _ := g.grid.PixelToHex(g.cam.ScreenToWorld(sx, sy)); got != want {
			t.Fatalf("%s: centre of %v picks %v", s.name, want, got)
		}
	}
	// Same window in device-independent pixels, double density: same world area.
	g.applyLayout(1000, 600, 1)
	_, _, w1, h1 := g.cam.Visible()
	g.applyLayout(1000, 600, 2)
	_, _, w2, h2 := g.cam.Visible()
	if math.Abs(w1-w2) > 1e-6 || math.Abs(h1-h2) > 1e-6 {
		t.Errorf("display scale changed the visible world: %vx%v -> %vx%v", w1, h1, w2, h2)
	}
}

func TestLayoutHUDFitsScreen(t *testing.T) {
	for _, sz := range [][2]int{{400, 700}, {320, 200}, {1280, 720}, {1800, 1000}, {50, 50}} {
		w, h := sz[0], sz[1]
		l := layoutHUD(w, h)
		if l.Banner.W != w || l.Banner.Y != (h-bannerHeight)/2 {
			t.Errorf("%dx%d: banner %+v", w, h, l.Banner)
		}
		if l.DevRight != w-hudMargin {
			t.Errorf("%dx%d: dev indicators end at %d", w, h, l.DevRight)
		}
	}
}

func TestInfoRectFitsScreen(t *testing.T) {
	l := layoutHUD(320, 240)
	r := l.infoRect(1000, 4)
	if r.X < hudMargin || r.X+r.W > l.DevRight || r.Y != infoTop || r.H != 4*infoLineHeight+infoPad {
		t.Errorf("info rect %+v on a 320 px screen", r)
	}
}

func TestAssetDensity(t *testing.T) {
	for _, tc := range []struct {
		eff  float64
		want int
	}{{0.25, 1}, {1, 1}, {1.01, 2}, {2, 2}, {2.5, 4}, {4, 4}, {16, 4}, {0, 1}} {
		if got := assetDensity(tc.eff); got != tc.want {
			t.Errorf("assetDensity(%v) = %d, want %d", tc.eff, got, tc.want)
		}
	}
}
