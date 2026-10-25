package dosscreen

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
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
		levels []int
	}{
		{1280, 720, 1, []int{0, 1, 2, 3, 4}},
		{1280, 720, 1.5, []int{2, 3, 4}},
		{1280, 720, 2, []int{2, 3, 4}},
		{1920, 1080, 1, []int{0, 1, 2, 3, 4}},
		{1920, 1080, 2, []int{1, 2, 3, 4}},
		{3840, 2160, 1, []int{0, 1, 2, 3, 4}},
		{3840, 2160, 2, []int{0, 1, 2, 3, 4}},
		{640, 400, 1, []int{2, 3, 4}},
	}
	for _, tc := range tests {
		if got := Layout(tc.w, tc.h, tc.ui).ZoomLevels(); !slices.Equal(got, tc.levels) {
			t.Errorf("%dx%d@%v: levels %v, want %v", tc.w, tc.h, tc.ui, got, tc.levels)
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

// The teeth meet the cell edges at every usable zoom level of real layouts, in
// both panels, inside the map and at its edge.
func TestZoomTeethMeetCellEdges(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	for _, w := range [][2]int{{1280, 720}, {1920, 1080}, {3840, 2160}, {1024, 768}} {
		for _, ui := range []float64{1, 1.5, 2} {
			s := Layout(w[0], w[1], ui)
			for _, i := range s.ZoomLevels() {
				z := s.Zoomed(i)
				name := fmt.Sprintf("%dx%d@%v zoom %d", w[0], w[1], ui, i)
				checkTeeth(t, name+" left", z.View(0, 2, 3), g)
				checkTeeth(t, name+" right", z.View(1, -1, 0), g)
			}
		}
	}
}

// A zoomed Screen is a new cache key for the panel frames but keeps the chrome,
// which does not change with the zoom; going back reuses the cached panels.
func TestFrameCacheZoom(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	s := Layout(1280, 720, 1)
	out := s.Zoomed(DefaultZoom - 1)
	if out == s || out.Zoom == s.Zoom {
		t.Fatalf("zoomed screen %+v equals the layout", out)
	}
	dst := ebiten.NewImage(1280, 720)
	var f Frame
	defer f.Release()
	draw := func(s Screen) {
		f.Draw(dst, s, [2]View{s.View(0, 5, 5), s.View(1, 9, 9)}, g)
	}
	draw(s)
	if f.painted != 3 {
		t.Fatalf("first draw painted %d images, want chrome and two panels", f.painted)
	}
	draw(out)
	if f.painted != 5 {
		t.Errorf("zooming out painted %d images, want two new panels and no chrome", f.painted)
	}
	draw(s)
	if f.painted != 5 {
		t.Errorf("zooming back repainted: %d", f.painted)
	}
}
