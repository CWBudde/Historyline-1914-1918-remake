package minimap

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

var campaign = hexgrid.Grid{Width: 16, Lines: 63} // 31×32 cells

// The map takes the largest whole multiple of the 4×4 cell that fits inside the
// frame and sits centred in the panel.
func TestStrategicLayout(t *testing.T) {
	for _, tc := range []struct {
		panel image.Point
		scale int
		at    image.Rectangle
	}{
		{image.Pt(304, 360), 2, image.Rect(28, 52, 276, 308)},
		{image.Pt(152, 180), 1, image.Rect(14, 26, 138, 154)},
		{image.Pt(608, 360), 2, image.Rect(180, 52, 428, 308)},
		{image.Pt(50, 50), 1, image.Rect(-37, -39, 87, 89)}, // too small: centred and clipped
	} {
		s := NewStrategic(campaign, tc.panel)
		if s.Scale != tc.scale || s.Map != tc.at {
			t.Errorf("panel %v: scale %d at %v, want %d at %v", tc.panel, s.Scale, s.Map, tc.scale, tc.at)
		}
		if s.Raster.CellW != 4*tc.scale || s.Raster.CellH != 4*tc.scale {
			t.Errorf("panel %v: cells %dx%d", tc.panel, s.Raster.CellW, s.Raster.CellH)
		}
	}
}

var (
	grass   = color.RGBA{90, 120, 60, 255}
	plateC  = color.RGBA{48, 48, 36, 255}
	outerC  = color.RGBA{81, 81, 65, 255}
	middleC = color.RGBA{36, 36, 28, 255}
	innerC  = color.RGBA{28, 28, 24, 255}
	viewC   = color.RGBA{154, 154, 138, 255}
)

func strategicScene(cam *camera.Camera) Scene {
	return Scene{
		Terrain: func(hexgrid.Coord) color.RGBA { return grass },
		Units:   []Unit{{At: hexgrid.Coord{Line: 40, Col: 10}, Shade: Yellow}},
		Buildings: []Building{
			{At: hexgrid.Coord{Line: 50, Col: 4}, Kind: HQ, Shade: Blue},
			{At: hexgrid.Coord{Line: 54, Col: 12}, Kind: Depot},
			{At: hexgrid.Coord{Line: 30, Col: 6}, Kind: Factory, Shade: Yellow},
		},
		View: cam,
	}
}

func TestStrategicRender(t *testing.T) {
	s := NewStrategic(campaign, image.Pt(304, 360))
	w, h := campaign.Bounds()
	cam := camera.New(camera.Size{W: w, H: h}, camera.Size{W: 710, H: 400})
	cam.Origin = camera.Point{X: 12.5 + 3*71, Y: 5 * 40}
	img := s.Render(strategicScene(cam))
	if img.Bounds() != image.Rect(0, 0, 304, 360) {
		t.Fatalf("bounds %v", img.Bounds())
	}
	m := s.Map.Min // 28,52
	at := func(c hexgrid.Coord) image.Point { return s.Raster.Block(c).Min.Add(m) }
	hq, depot, unit := at(hexgrid.Coord{Line: 50, Col: 4}), at(hexgrid.Coord{Line: 54, Col: 12}), at(hexgrid.Coord{Line: 40, Col: 10})
	// plate is the top-left pixel of a 9×9 plate centred on c's block.
	plate := func(c hexgrid.Coord) image.Point {
		b := s.Raster.Block(c).Add(m)
		return b.Min.Add(image.Pt(b.Dx()/2, b.Dy()/2)).Sub(image.Pt(9*s.Scale/2, 9*s.Scale/2))
	}
	disc, square := plate(hexgrid.Coord{Line: 54, Col: 12}), plate(hexgrid.Coord{Line: 30, Col: 6})
	px := func(p image.Point, x, y int) image.Point { return p.Add(image.Pt(x*s.Scale, y*s.Scale)) }
	for _, tc := range []struct {
		name string
		p    image.Point
		want color.RGBA
	}{
		{"plate corner", image.Pt(0, 0), plateC},
		{"plate beside the frame", image.Pt(m.X-5, 200), plateC},
		{"frame outer line", image.Pt(m.X-4, 200), outerC},
		{"frame middle", image.Pt(m.X-3, 200), middleC},
		{"frame middle 2", image.Pt(m.X-2, 200), middleC},
		{"frame inner line", image.Pt(m.X-1, 200), innerC},
		{"frame bottom", image.Pt(150, s.Map.Max.Y+3), outerC},
		{"jagged top of an odd column", image.Pt(m.X+8+3, m.Y+1), innerC},
		{"terrain", image.Pt(m.X+3, m.Y+3), grass},
		{"unit lit corner", unit.Add(image.Pt(2, 2)), Yellow.Lit},
		{"unit dark corner", unit.Add(image.Pt(5, 5)), Yellow.Dark},
		{"unit inset", unit.Add(image.Pt(1, 1)), grass},
		{"HQ plate body", hq.Add(image.Pt(4, 4)), Blue.Mid},
		{"HQ plate outline", hq.Add(image.Pt(4-11, 4)), Blue.Dark},
		{"depot body", depot.Add(image.Pt(4, 4)), depotBody},
		{"depot D", depot.Add(image.Pt(2, 4)), depotMark},
		{"depot round corner", px(disc, 0, 0), grass},
		{"factory square corner", px(square, 0, 0), Yellow.Dark},
		{"factory body", px(square, 5, 6), depotBody},
		{"factory F", px(square, 3, 4), depotMark},
		{"view outline top left", image.Pt(m.X+24, m.Y+20), viewC},
		{"view outline bottom right", image.Pt(m.X+103, m.Y+59), viewC},
		{"view inner shadow", image.Pt(m.X+25, m.Y+30), viewShade},
		{"view drop shadow right", image.Pt(m.X+104, m.Y+30), outerC},
		{"view drop shadow bottom", image.Pt(m.X+30, m.Y+60), outerC},
		{"view inside", image.Pt(m.X+50, m.Y+30), grass},
	} {
		if got := img.RGBAAt(tc.p.X, tc.p.Y); got != tc.want {
			t.Errorf("%s at %v = %v, want %v", tc.name, tc.p, got, tc.want)
		}
	}
	if got := s.ViewRect(cam); got != image.Rect(m.X+24, m.Y+20, m.X+104, m.Y+60) {
		t.Errorf("view rect %v", got)
	}
}

// Draw paints the same picture into a larger image at an offset.
func TestStrategicDrawAtOffset(t *testing.T) {
	s := NewStrategic(campaign, image.Pt(152, 180))
	alone := s.Render(strategicScene(nil))
	dst := image.NewRGBA(image.Rect(0, 0, 400, 300))
	off := image.Pt(160, 20)
	s.Draw(dst, off, strategicScene(nil))
	for y := 0; y < 180; y++ {
		for x := 0; x < 152; x++ {
			if a, b := alone.RGBAAt(x, y), dst.RGBAAt(x+off.X, y+off.Y); a != b {
				t.Fatalf("pixel %d,%d: %v vs %v", x, y, a, b)
			}
		}
	}
	if dst.RGBAAt(off.X-1, off.Y) != (color.RGBA{}) || dst.RGBAAt(off.X+152, off.Y) != (color.RGBA{}) {
		t.Error("Draw paints outside the panel")
	}
}

// A click on a cell names the world point of that hex; the frame and plate are no hit.
func TestStrategicToWorld(t *testing.T) {
	s := NewStrategic(campaign, image.Pt(304, 360))
	for _, c := range hexes(campaign) {
		b := s.Raster.Block(c).Add(s.Map.Min)
		wx, wy, ok := s.ToWorld(float64(b.Min.X)+4, float64(b.Min.Y)+4)
		cx, cy := hexgrid.Center(c)
		if !ok || math.Abs(wx-cx) > 1e-9 || math.Abs(wy-cy) > 1e-9 {
			t.Fatalf("centre of %v: %v,%v %v, want %v,%v", c, wx, wy, ok, cx, cy)
		}
		wx, wy, _ = s.ToWorld(float64(b.Min.X)+2.5, float64(b.Min.Y)+5.5)
		if got, _ := campaign.PixelToHex(wx, wy); got != c {
			t.Fatalf("click inside %v picks %v", c, got)
		}
	}
	for _, p := range []image.Point{{0, 0}, {s.Map.Min.X - 1, 100}, {100, s.Map.Max.Y}} {
		if _, _, ok := s.ToWorld(float64(p.X), float64(p.Y)); ok {
			t.Errorf("%v hits the map", p)
		}
	}
}
