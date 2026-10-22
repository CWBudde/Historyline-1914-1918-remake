package minimap

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

var rasterCases = []struct {
	name   string
	grid   hexgrid.Grid
	cw, ch int
}{
	{"pulse 4x4", hexgrid.Grid{Width: 8, Lines: 47}, 4, 4},
	{"campaign 2x2", hexgrid.Grid{Width: 16, Lines: 63}, 2, 2},
	{"campaign 7x8", hexgrid.Grid{Width: 16, Lines: 63}, 7, 8},
	{"narrow cells 3x6", hexgrid.Grid{Width: 5, Lines: 9}, 3, 6},
	{"even line count", hexgrid.Grid{Width: 4, Lines: 6}, 4, 4},
	{"one hex", hexgrid.Grid{Width: 1, Lines: 1}, 4, 4},
	{"one line", hexgrid.Grid{Width: 5, Lines: 1}, 4, 4},
	{"one column", hexgrid.Grid{Width: 1, Lines: 5}, 4, 4},
}

func hexes(g hexgrid.Grid) []hexgrid.Coord {
	var out []hexgrid.Coord
	for l := 0; l < g.Lines; l++ {
		for c := 0; c < g.Width; c++ {
			if at := (hexgrid.Coord{Line: l, Col: c}); g.Contains(at) {
				out = append(out, at)
			}
		}
	}
	return out
}

// Every pixel inside the map belongs to exactly one hex: blocks never overlap, they
// fill the raster except the half cells at the jagged top and bottom edge, and At
// names the hex whose block holds the pixel.
func TestCellsTileEveryPixelOnce(t *testing.T) {
	for _, tc := range rasterCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Cells(tc.grid, tc.cw, tc.ch)
			size := r.Size()
			owner := make([]int, size.X*size.Y) // index+1 of the covering hex, 0 = none
			all := hexes(tc.grid)
			for i, at := range all {
				b := r.Block(at)
				if b.Dx() != tc.cw || b.Dy() != tc.ch {
					t.Fatalf("block of %v is %v", at, b)
				}
				if !b.In(image.Rect(0, 0, size.X, size.Y)) {
					t.Fatalf("block of %v %v leaves the raster %v", at, b, size)
				}
				for y := b.Min.Y; y < b.Max.Y; y++ {
					for x := b.Min.X; x < b.Max.X; x++ {
						if o := owner[y*size.X+x]; o != 0 {
							t.Fatalf("pixel %d,%d in %v and %v", x, y, all[o-1], at)
						}
						owner[y*size.X+x] = i + 1
					}
				}
			}
			covered := 0
			for y := 0; y < size.Y; y++ {
				for x := 0; x < size.X; x++ {
					o := owner[y*size.X+x]
					at, ok := r.At(x, y)
					if o == 0 {
						if ok {
							t.Fatalf("At(%d,%d) = %v on an empty pixel", x, y, at)
						}
						continue
					}
					covered++
					if !ok || at != all[o-1] {
						t.Fatalf("At(%d,%d) = %v %v, want %v", x, y, at, ok, all[o-1])
					}
				}
			}
			if covered != len(all)*tc.cw*tc.ch {
				t.Fatalf("covered %d pixels, want %d", covered, len(all)*tc.cw*tc.ch)
			}
			// No gap: with both line parities present, every row but the half cells at
			// the top and bottom is covered from edge to edge.
			if tc.grid.Lines > 1 && tc.grid.Width > 1 {
				for y := tc.ch / 2; y < size.Y-tc.ch/2; y++ {
					for x := 0; x < size.X; x++ {
						if owner[y*size.X+x] == 0 {
							t.Fatalf("gap at %d,%d", x, y)
						}
					}
				}
			}
			for _, p := range []image.Point{{-1, 0}, {0, -1}, {size.X, 0}, {0, size.Y}} {
				if at, ok := r.At(p.X, p.Y); ok {
					t.Errorf("At(%v) = %v outside the raster", p, at)
				}
			}
		})
	}
}

// Same-line neighbours sit two block columns apart, odd lines one block column
// (half a hex) to the right and half a block down, and line l+2 directly below
// line l: the raster keeps the hex layout without a half-line shift.
func TestCellsFollowTheHexLayout(t *testing.T) {
	r := Cells(hexgrid.Grid{Width: 8, Lines: 47}, 4, 4)
	if got := r.Block(hexgrid.Coord{}); got != image.Rect(0, 0, 4, 4) {
		t.Fatalf("block of 0,0 = %v", got)
	}
	if got := r.Block(hexgrid.Coord{Line: 1}); got != image.Rect(4, 2, 8, 6) {
		t.Fatalf("block of 1,0 = %v", got)
	}
	if got := r.Block(hexgrid.Coord{Line: 6, Col: 3}); got != image.Rect(24, 12, 28, 16) {
		t.Fatalf("block of 6,3 = %v", got)
	}
	if got := r.Size(); got != image.Pt(15*4, 48*2) {
		t.Fatalf("size = %v", got)
	}
	for _, at := range hexes(r.Grid) {
		b := r.Block(at)
		// The block centre is the hex centre.
		wx, wy := r.ToWorld(float64(b.Min.X)+2, float64(b.Min.Y)+2)
		if cx, cy := hexgrid.Center(at); wx != cx || wy != cy {
			t.Fatalf("centre of block %v = %v,%v, hex centre %v,%v", at, wx, wy, cx, cy)
		}
		// The four central pixels lie inside the hex.
		for _, p := range []image.Point{{1, 1}, {2, 1}, {1, 2}, {2, 2}} {
			wx, wy := r.ToWorld(float64(b.Min.X+p.X)+0.5, float64(b.Min.Y+p.Y)+0.5)
			if got, _ := r.Grid.PixelToHex(wx, wy); got != at {
				t.Fatalf("pixel %v of %v lies in hex %v", p, at, got)
			}
		}
	}
}

func TestRasterWorldRoundTrip(t *testing.T) {
	r := Cells(hexgrid.Grid{Width: 16, Lines: 63}, 7, 8)
	for _, p := range [][2]float64{{12.5, 0}, {500, 700}, {2213.5, 2560}, {48, 40}} {
		px, py := r.FromWorld(p[0], p[1])
		wx, wy := r.ToWorld(px, py)
		if math.Abs(wx-p[0]) > 1e-9 || math.Abs(wy-p[1]) > 1e-9 {
			t.Errorf("round trip of %v = %v,%v", p, wx, wy)
		}
	}
	// The raster spans the world without the slanted tips at the left and right.
	w, h := r.Grid.Bounds()
	px, py := r.FromWorld(w-12.5, h)
	if size := r.Size(); px != float64(size.X) || py != float64(size.Y) {
		t.Errorf("world corner at %v,%v, raster %v", px, py, size)
	}
}

// Raster.Minimap places the raster on screen with the same mapping, so a hex centre
// lands on its block centre and the minimap is exactly as large as the raster.
func TestRasterMinimapAgrees(t *testing.T) {
	r := Cells(hexgrid.Grid{Width: 8, Lines: 47}, 4, 4)
	m := r.Minimap(camera.Point{X: 100, Y: 50})
	if s := m.Size(); s.W != 60 || s.H != 96 {
		t.Fatalf("size = %v", s)
	}
	for _, at := range hexes(r.Grid) {
		b := r.Block(at)
		sx, sy := m.FromWorld(hexgrid.Center(at))
		if math.Abs(sx-(100+float64(b.Min.X)+2)) > 1e-9 || math.Abs(sy-(50+float64(b.Min.Y)+2)) > 1e-9 {
			t.Fatalf("hex %v at %v,%v, block %v", at, sx, sy, b)
		}
		wx, wy := m.ToWorld(sx, sy)
		if got, _ := r.Grid.PixelToHex(wx, wy); got != at {
			t.Fatalf("click on block %v picks %v", at, got)
		}
	}
	if !m.Contains(100, 50) || m.Contains(160, 50) || m.Contains(100, 146) {
		t.Error("Contains boundaries wrong")
	}
}

// The view rectangle's edges meet block edges when the camera stands on a cell edge.
func TestViewRectLinesUpWithCells(t *testing.T) {
	g := hexgrid.Grid{Width: 16, Lines: 63}
	r := Cells(g, 4, 4)
	m := r.Minimap(camera.Point{})
	w, h := g.Bounds()
	cam := camera.New(camera.Size{W: w, H: h}, camera.Size{W: 710, H: 400})
	cam.Origin = camera.Point{X: 12.5 + 3*71, Y: 5 * 40}
	x, y, vw, vh := m.ViewRect(cam)
	for i, p := range [][2]float64{{x, 12}, {y, 10}, {vw, 40}, {vh, 20}} {
		if math.Abs(p[0]-p[1]) > 1e-9 {
			t.Fatalf("rect value %d = %v, want %v", i, p[0], p[1])
		}
	}
	// At the map's left edge the rectangle is clipped to the raster.
	cam.Origin = camera.Point{}
	if x, y, _, _ := m.ViewRect(cam); x != 0 || y != 0 {
		t.Fatalf("rect at the origin starts at %v,%v", x, y)
	}
}

func TestFitGrid(t *testing.T) {
	view := camera.Size{W: 1000, H: 700}
	for _, tc := range []struct {
		grid   hexgrid.Grid
		max    float64
		cw, ch int
	}{
		{hexgrid.Grid{Width: 8, Lines: 47}, 280, 9, 10},  // 135×240
		{hexgrid.Grid{Width: 16, Lines: 63}, 280, 7, 8},  // 217×256
		{hexgrid.Grid{Width: 16, Lines: 63}, 80, 2, 2},   // 62×64
		{hexgrid.Grid{Width: 50, Lines: 120}, 280, 2, 2}, // 198×121
		{hexgrid.Grid{Width: 3, Lines: 3}, 80, 16, 18},   // 80×36
		{hexgrid.Grid{Width: 50, Lines: 300}, 80, 2, 2},  // the smallest cells overflow
	} {
		m, r := FitGrid(tc.grid, camera.Size{W: tc.max, H: tc.max}, view, 12)
		if r.CellW != tc.cw || r.CellH != tc.ch {
			t.Errorf("%v in %v: cells %dx%d, want %dx%d", tc.grid, tc.max, r.CellW, r.CellH, tc.cw, tc.ch)
		}
		size := r.Size()
		if s := m.Size(); s.W != float64(size.X) || s.H != float64(size.Y) {
			t.Errorf("%v: minimap %v, raster %v", tc.grid, s, size)
		}
		if m.Pos.X+float64(size.X) != 988 || m.Pos.Y+float64(size.Y) != 688 {
			t.Errorf("%v: bottom-right at %v + %v", tc.grid, m.Pos, size)
		}
		// The cells keep the world's proportions: 71 px per block column, 80 per cell.
		if ratio := float64(r.CellW) / float64(r.CellH); math.Abs(ratio-71.0/80) > 0.12 {
			t.Errorf("%v: cell ratio %v", tc.grid, ratio)
		}
	}
}

func TestCellsNormalisesItsArguments(t *testing.T) {
	r := Cells(hexgrid.Grid{Width: 3, Lines: 3}, 0, 3)
	if r.CellW != 1 || r.CellH != 4 {
		t.Fatalf("cells %dx%d, want 1x4", r.CellW, r.CellH)
	}
	if s := Cells(hexgrid.Grid{}, 4, 4).Size(); s != (image.Point{}) {
		t.Fatalf("empty grid size %v", s)
	}
}

func TestRenderFillsEachBlock(t *testing.T) {
	r := Cells(hexgrid.Grid{Width: 4, Lines: 5}, 4, 4)
	fill := func(at hexgrid.Coord) color.RGBA {
		return color.RGBA{uint8(10 * at.Line), uint8(10 * at.Col), 7, 255}
	}
	img := r.Render(fill)
	if img.Bounds() != image.Rect(0, 0, r.Size().X, r.Size().Y) {
		t.Fatalf("bounds %v", img.Bounds())
	}
	for y := 0; y < r.Size().Y; y++ {
		for x := 0; x < r.Size().X; x++ {
			got := img.RGBAAt(x, y)
			if at, ok := r.At(x, y); ok {
				if got != fill(at) {
					t.Fatalf("pixel %d,%d = %v, want %v", x, y, got, fill(at))
				}
			} else if got != (color.RGBA{}) {
				t.Fatalf("pixel %d,%d off the map = %v", x, y, got)
			}
		}
	}
}

func TestUnitDotIsShadedSquare(t *testing.T) {
	r := Cells(hexgrid.Grid{Width: 4, Lines: 5}, 8, 8)
	img := image.NewRGBA(image.Rect(0, 0, r.Size().X, r.Size().Y))
	at := hexgrid.Coord{Line: 2, Col: 1}
	r.DrawUnit(img, image.Point{}, Unit{At: at, Shade: Blue})
	b := r.Block(at) // 16,8 – 24,16; the dot is inset 2 px: 18,10 – 22,14
	for _, tc := range []struct {
		x, y int
		want color.RGBA
	}{
		{b.Min.X + 2, b.Min.Y + 2, Blue.Lit},
		{b.Min.X + 3, b.Min.Y + 2, Blue.Lit},
		{b.Min.X + 3, b.Min.Y + 3, Blue.Mid},
		{b.Min.X + 5, b.Min.Y + 2, Blue.Mid}, // top-right corner
		{b.Min.X + 5, b.Min.Y + 5, Blue.Dark},
		{b.Min.X + 4, b.Min.Y + 5, Blue.Dark},
		{b.Min.X + 1, b.Min.Y + 1, color.RGBA{}},
		{b.Min.X + 6, b.Min.Y + 6, color.RGBA{}},
	} {
		if got := img.RGBAAt(tc.x, tc.y); got != tc.want {
			t.Errorf("pixel %d,%d = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
	// On the smallest cells the dot fills its block.
	small := Cells(hexgrid.Grid{Width: 4, Lines: 5}, 2, 2)
	img = image.NewRGBA(image.Rect(0, 0, small.Size().X, small.Size().Y))
	small.DrawUnit(img, image.Point{}, Unit{At: at, Shade: Yellow})
	b = small.Block(at)
	if img.RGBAAt(b.Min.X, b.Min.Y) != Yellow.Lit || img.RGBAAt(b.Max.X-1, b.Max.Y-1) != Yellow.Dark {
		t.Errorf("2x2 dot %v %v", img.RGBAAt(b.Min.X, b.Min.Y), img.RGBAAt(b.Max.X-1, b.Max.Y-1))
	}
}
