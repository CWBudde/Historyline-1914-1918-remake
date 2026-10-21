package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// Literal integer half-planes independently pin the six game-grid vertices.
// Doubled pixel centers avoid floating-point ambiguity on shared edge ties.
func testTileOwned(x, y, density int) bool {
	xx, yy, d := 2*x+1, 2*y+1, 2*density
	return yy >= 0 && yy < 80*d &&
		40*xx+25*yy >= 1000*d &&
		40*xx-25*yy < 2840*d &&
		40*xx+25*yy < 4840*d &&
		40*xx-25*yy >= -1000*d
}

func TestInsideTileMatchesGamePolygon(t *testing.T) {
	for _, tc := range []struct {
		density, pixels int
	}{{1, 5680}, {2, 22720}, {4, 90880}} {
		t.Run(fmt.Sprintf("%dx", tc.density), func(t *testing.T) {
			owned := 0
			// Check beyond every side as well as the complete PNG bounds.
			for y := -2 * tc.density; y < 82*tc.density; y++ {
				for x := -2 * tc.density; x < 98*tc.density; x++ {
					got, want := insideTile(x, y, tc.density), testTileOwned(x, y, tc.density)
					if got != want {
						t.Fatalf("mask at (%d,%d): got %t, want %t", x, y, got, want)
					}
					if got {
						owned++
					}
				}
			}
			if owned != tc.pixels {
				t.Fatalf("owned pixels = %d, want %d", owned, tc.pixels)
			}
		})
	}
}

func TestTileOriginUsesStaggeredGameGrid(t *testing.T) {
	for _, tc := range []struct {
		row, col, x, y int
	}{
		{0, 0, 0, 0},
		{0, 1, 142, 0},
		{1, 0, 71, 40},
		{1, 2, 355, 40},
		{2, 0, 0, 80},
		{46, 7, 994, 1840},
	} {
		t.Run(fmt.Sprintf("row%d-col%d", tc.row, tc.col), func(t *testing.T) {
			x, y := tileOrigin(tc.row, tc.col)
			if x != tc.x || y != tc.y {
				t.Fatalf("origin = (%d,%d), want (%d,%d)", x, y, tc.x, tc.y)
			}
		})
	}
}

func TestTilePackingHasOneOwnerAcrossBothParities(t *testing.T) {
	for _, density := range []int{1, 2, 4} {
		for _, firstRow := range []int{0, 1} {
			t.Run(fmt.Sprintf("%dx-first-parity%d", density, firstRow), func(t *testing.T) {
				bounds := image.Rect(-density, (40*firstRow-1)*density, 523*density, (40*(firstRow+5)+81)*density)
				want := make([]uint8, bounds.Dx()*bounds.Dy())
				got := make([]uint8, len(want))
				index := func(x, y int) int { return (y-bounds.Min.Y)*bounds.Dx() + x - bounds.Min.X }
				cells := 0
				for row := firstRow; row < firstRow+6; row++ {
					for col := 0; col < 4-row%2; col++ {
						cells++
						ox, oy := tileOrigin(row, col)
						wx, wy := 142*col+71*(row%2), 40*row
						for y := 0; y < 80*density; y++ {
							for x := 0; x < 96*density; x++ {
								if testTileOwned(x, y, density) {
									want[index(wx*density+x, wy*density+y)]++
								}
								if insideTile(x, y, density) {
									p := image.Pt(ox*density+x, oy*density+y)
									if !p.In(bounds) {
										t.Fatalf("tile (%d,%d) places pixel outside proof bounds: %v", row, col, p)
									}
									got[index(p.X, p.Y)]++
								}
							}
						}
					}
				}
				covered := 0
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						i := index(x, y)
						if want[i] > 1 || got[i] != want[i] {
							t.Fatalf("owners at (%d,%d): got %d, independent polygon union %d", x, y, got[i], want[i])
						}
						if got[i] == 1 {
							covered++
						}
						// This rectangle lies strictly inside the patch boundary.
						// It must stay filled across all horizontal and diagonal seams.
						if x >= 96*density && x < 426*density && y >= 40*(firstRow+2)*density && y < 40*(firstRow+4)*density && got[i] != 1 {
							t.Fatalf("interior gap at (%d,%d)", x, y)
						}
					}
				}
				if expected := cells * 5680 * density * density; covered != expected {
					t.Fatalf("covered pixels = %d, want %d", covered, expected)
				}
			})
		}
	}
}

func TestPrepareSourceClearsOnlyPerimeterConnectedBlack(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 7, 7))
	black := color.NRGBA{A: 255}
	paint := color.NRGBA{R: 140, G: 170, B: 90, A: 255}
	draw.Draw(src, src.Bounds(), image.NewUniform(black), image.Point{}, draw.Src)
	draw.Draw(src, image.Rect(1, 1, 6, 6), image.NewUniform(paint), image.Point{}, draw.Src)
	src.SetNRGBA(3, 3, black) // Enclosed dark detail must survive.
	src.SetNRGBA(2, 2, black) // Diagonally adjacent details remain enclosed.
	nearBlack := color.NRGBA{B: 1, A: 255}
	src.SetNRGBA(0, 0, nearBlack) // This must not become a global color key.
	got := prepareSource(src)
	if got.Bounds() != src.Bounds() {
		t.Fatalf("prepared bounds = %v, want %v", got.Bounds(), src.Bounds())
	}
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			want := src.NRGBAAt(x, y)
			if want == black && (x == 0 || x == 6 || y == 0 || y == 6) {
				want = color.NRGBA{}
			}
			if pixel := got.NRGBAAt(x, y); pixel != want {
				t.Fatalf("prepared pixel (%d,%d) = %v, want %v", x, y, pixel, want)
			}
		}
	}
	if src.NRGBAAt(0, 3) != black {
		t.Fatal("preparing source modified the supplied image")
	}
}

func TestPrepareSourceDoesNotFloodThroughDiagonalContact(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 7, 7))
	paint := color.NRGBA{R: 140, G: 170, B: 90, A: 255}
	draw.Draw(src, src.Bounds(), image.NewUniform(paint), image.Point{}, draw.Src)
	black := color.NRGBA{A: 255}
	for _, p := range []image.Point{{0, 3}, {1, 3}, {2, 2}} {
		src.SetNRGBA(p.X, p.Y, black)
	}
	got := prepareSource(src)
	if got.NRGBAAt(0, 3) != (color.NRGBA{}) || got.NRGBAAt(1, 3) != (color.NRGBA{}) {
		t.Fatal("perimeter-connected black corridor remains visible")
	}
	if got.NRGBAAt(2, 2) != black {
		t.Fatal("diagonal contact incorrectly removed an enclosed black detail")
	}
}

func TestSampleSourcePreservesColorAtTransparentFringe(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	paint := color.NRGBA{R: 200, G: 100, B: 50, A: 255}
	src.SetNRGBA(0, 0, paint)
	src.SetNRGBA(0, 1, paint)
	for _, tc := range []struct {
		name        string
		sx          float64
		alpha       int
		transparent bool
	}{
		{"opaque", .5, 255, false},
		{"half-covered", 1, 128, false},
		{"quarter-covered", 1.25, 64, false},
		{"transparent", 1.5, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Invert the declared source registration to address source pixel centers.
			wx := (tc.sx - 42 + 48.0*123/142) * 142 / 123
			got := sampleSource(src, wx, 40.0/35)
			if tc.transparent {
				if got != (color.NRGBA{}) {
					t.Fatalf("transparent sample = %v, want zero RGBA", got)
				}
				return
			}
			if got.R != paint.R || got.G != paint.G || got.B != paint.B {
				t.Fatalf("fringe color = %v, want unchanged RGB %v", got, paint)
			}
			// Inverting the affine can move an exact half-alpha tie by an ulp.
			if delta := int(got.A) - tc.alpha; delta < -1 || delta > 1 {
				t.Fatalf("fringe alpha = %d, want %d (+/-1)", got.A, tc.alpha)
			}
		})
	}
}

func TestRenderTileUsesOpaqueInteriorAndZeroOutsidePolygon(t *testing.T) {
	paint := color.NRGBA{R: 170, G: 120, B: 60, A: 255}
	src := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	draw.Draw(src, src.Bounds(), image.NewUniform(paint), image.Point{}, draw.Src)
	for _, density := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("%dx", density), func(t *testing.T) {
			got, missing := renderTile(src, 71, 40, density)
			if missing != 0 {
				t.Fatalf("fully painted tile reports %d missing pixels", missing)
			}
			if want := image.Rect(0, 0, 96*density, 80*density); got.Bounds() != want {
				t.Fatalf("tile bounds = %v, want %v", got.Bounds(), want)
			}
			for y := range got.Bounds().Dy() {
				for x := range got.Bounds().Dx() {
					want := color.NRGBA{}
					if testTileOwned(x, y, density) {
						want = paint
					}
					if pixel := got.NRGBAAt(x, y); pixel != want {
						t.Fatalf("tile pixel (%d,%d) = %v, want %v", x, y, pixel, want)
					}
				}
			}
		})
	}
}
