package main

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
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
