package dosscreen

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

// The zoom factors rise from far out to close in, with the fitted scale at
// DefaultZoom; Layout starts there and Zoomed(DefaultZoom) changes nothing.
func TestZooms(t *testing.T) {
	if Zooms[DefaultZoom] != 1 {
		t.Fatalf("Zooms[DefaultZoom] = %v, want 1", Zooms[DefaultZoom])
	}
	for i := 1; i < len(Zooms); i++ {
		if Zooms[i] <= Zooms[i-1] {
			t.Errorf("Zooms not rising at %d: %v", i, Zooms)
		}
	}
	for _, w := range windows {
		for _, ui := range uiScales {
			s := Layout(w[0], w[1], ui)
			if s.Zoom != DefaultZoom || s.FitScale != s.CellScale {
				t.Errorf("%dx%d@%v: layout at zoom %d, scale %v of %v", w[0], w[1], ui, s.Zoom, s.CellScale, s.FitScale)
			}
			if z := s.Zoomed(DefaultZoom); z != s {
				t.Errorf("%dx%d@%v: Zoomed(DefaultZoom) = %+v, want %+v", w[0], w[1], ui, z, s)
			}
		}
	}
}

// A zoom level is usable when its scale stays at or above MinScale × ui and the
// panel still shows ZoomCols × ZoomRows cells; the fitted level always is. The
// usable levels are a run around DefaultZoom.
func TestZoomLevels(t *testing.T) {
	for _, w := range windows {
		for _, ui := range uiScales {
			s := Layout(w[0], w[1], ui)
			name := fmt.Sprintf("%dx%d@%v", w[0], w[1], ui)
			levels := s.ZoomLevels()
			if !slices.Contains(levels, DefaultZoom) {
				t.Errorf("%s: levels %v miss the default", name, levels)
			}
			for k := 1; k < len(levels); k++ {
				if levels[k] != levels[k-1]+1 {
					t.Errorf("%s: levels %v not a run", name, levels)
				}
			}
			for i, f := range Zooms {
				sc := s.FitScale * f
				cols, rows := Fit(s.Panels[0], sc)
				want := i == DefaultZoom || sc >= MinScale*ui-1e-9 && cols >= ZoomCols && rows >= ZoomRows
				if got := slices.Contains(levels, i); got != want {
					t.Errorf("%s: level %d (scale %.4f, %d×%d) usable %v, want %v (levels %v)", name, i, sc, cols, rows, got, want, levels)
				}
				if !want {
					continue
				}
				z := s.Zoomed(i)
				if z.Zoom != i || math.Abs(z.CellScale-s.FitScale*Zooms[i]) > 1e-9 || z.FitScale != s.FitScale {
					t.Errorf("%s: Zoomed(%d) = zoom %d scale %.4f, fitted %.4f", name, i, z.Zoom, z.CellScale, s.FitScale)
				}
				if z.Panels != s.Panels || z.Status != s.Status || z.Divider != s.Divider || z.Bolt != s.Bolt {
					t.Errorf("%s: Zoomed(%d) moved the pieces", name, i)
				}
				if c, r := Fit(z.Panels[0], z.CellScale); c != z.Cols || r != z.Rows {
					t.Errorf("%s: Zoomed(%d) shows %d×%d, fit %d×%d", name, i, z.Cols, z.Rows, c, r)
				}
			}
			// The levels do not depend on the zoom they are asked from.
			for _, i := range levels {
				if got := s.Zoomed(i).ZoomLevels(); !slices.Equal(got, levels) {
					t.Errorf("%s: levels from zoom %d %v, from the layout %v", name, i, got, levels)
				}
			}
		}
	}
}

func TestZoomLevelExamples(t *testing.T) {
	tests := []struct {
		w, h   int
		ui     float64
		lo, hi float64 // the factors of the outermost and closest levels
	}{
		{1280, 720, 1, 0.5, 1.81},
		{1280, 720, 1.5, 0.82, 1.81},
		{1280, 720, 2, 1, 1.64},
		{1920, 1080, 1, 0.5, 1.81},
		{1920, 1080, 2, 0.74, 1.81},
		{3840, 2160, 1, 0.5, 1.81},
		{3840, 2160, 2, 0.5, 1.81},
		{640, 400, 1, 1, 1.81},
	}
	for _, tc := range tests {
		got := Layout(tc.w, tc.h, tc.ui).ZoomLevels()
		if lo, hi := Zooms[got[0]], Zooms[got[len(got)-1]]; lo != tc.lo || hi != tc.hi {
			t.Errorf("%dx%d@%v: levels %v (%v..%v), want %v..%v", tc.w, tc.h, tc.ui, got, lo, hi, tc.lo, tc.hi)
		}
	}
}

// ClampZoom moves any index to the nearest usable level, and Zoomed only ever
// returns usable levels.
func TestClampZoom(t *testing.T) {
	for _, w := range windows {
		for _, ui := range uiScales {
			s := Layout(w[0], w[1], ui)
			levels := s.ZoomLevels()
			lo, hi := levels[0], levels[len(levels)-1]
			for i := -3; i < len(Zooms)+3; i++ {
				want := min(max(i, lo), hi)
				if got := s.ClampZoom(i); got != want {
					t.Errorf("%dx%d@%v: ClampZoom(%d) = %d, want %d (levels %v)", w[0], w[1], ui, i, got, want, levels)
				}
				if z := s.Zoomed(i); z.Zoom != want {
					t.Errorf("%dx%d@%v: Zoomed(%d) at zoom %d, want %d", w[0], w[1], ui, i, z.Zoom, want)
				}
			}
		}
	}
}
