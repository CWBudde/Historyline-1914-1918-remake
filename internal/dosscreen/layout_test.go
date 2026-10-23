package dosscreen

import (
	"fmt"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

var (
	windows = [][2]int{
		{1280, 720}, {1366, 768}, {1600, 900}, {1920, 1080}, {2560, 1440}, {3840, 2160},
		{3440, 1440}, {1024, 768}, {1280, 1024}, {1080, 1920}, {800, 600}, {640, 400},
	}
	uiScales = []float64{1, 1.25, 1.5, 1.75, 2}
)

// The pieces tile the window: two equal panels side by side with the divider
// between them, the status band below; nothing overlaps or leaves the window.
func TestLayoutFillsWindow(t *testing.T) {
	for _, w := range windows {
		for _, ui := range uiScales {
			s := Layout(w[0], w[1], ui)
			name := fmt.Sprintf("%dx%d@%v", w[0], w[1], ui)
			l, r := s.Panels[0], s.Panels[1]
			if l.W != r.W || l.H != r.H || l.Y != 0 || r.Y != 0 {
				t.Errorf("%s: panels %+v %+v not equal", name, l, r)
			}
			if l.X != 0 || l.Right() != s.Divider.X || s.Divider.Right() != r.X || r.Right() != w[0] {
				t.Errorf("%s: panels and divider %+v %+v %+v do not span the width", name, l, s.Divider, r)
			}
			if s.Divider.W < scaled(dividerNative, ui) || s.Divider.W > scaled(dividerNative, ui)+1 {
				t.Errorf("%s: divider width %d", name, s.Divider.W)
			}
			for i, st := range s.Status {
				p := s.Panels[i]
				if st.X != p.X || st.W != p.W || st.Y != p.Bottom()+scaled(gapNative, ui) {
					t.Errorf("%s: status %d %+v not under panel %+v", name, i, st, p)
				}
				if st.H != scaled(statusNative, ui) || st.Bottom()+scaled(footNative, ui) != w[1] {
					t.Errorf("%s: status %d %+v height or foot", name, i, st)
				}
			}
			if s.Bolt.X < s.Divider.X || s.Bolt.Right() > s.Divider.Right() ||
				s.Bolt.Y < s.Status[0].Y || s.Bolt.Bottom() > s.Status[0].Bottom() {
				t.Errorf("%s: bolt %+v not at the foot of the divider %+v", name, s.Bolt, s.Divider)
			}
			pieces := []Rect{l, r, s.Divider, s.Bolt, s.Status[0], s.Status[1]}
			for i, a := range pieces {
				if a.W <= 0 || a.H <= 0 || a.X < 0 || a.Y < 0 || a.Right() > w[0] || a.Bottom() > w[1] {
					t.Errorf("%s: piece %d %+v outside the window", name, i, a)
				}
				for _, b := range pieces[i+1:] {
					if a.Overlaps(b) {
						t.Errorf("%s: %+v overlaps %+v", name, a, b)
					}
				}
			}
		}
	}
}

// The panel height sets the cell scale for 7½ rows; extra width shows more
// columns, and a narrow window keeps 7 columns and shows more rows instead.
func TestLayoutCellScale(t *testing.T) {
	for _, w := range windows {
		for _, ui := range uiScales {
			s := Layout(w[0], w[1], ui)
			name := fmt.Sprintf("%dx%d@%v", w[0], w[1], ui)
			p := s.Panels[0]
			bw, bh := BlockSize(s.Cols, s.Rows)
			minW, minH := MinWindow(ui)
			floorBound := math.Abs(s.CellScale-MinScale*ui) < 1e-9
			switch {
			case w[0] < minW || w[1] < minH:
				// Below the smallest window only the fit is checked.
			case floorBound:
				if s.Cols < MinCols || s.Rows < MinRows {
					t.Errorf("%s: %d×%d cells at the smallest scale", name, s.Cols, s.Rows)
				}
			default:
				if s.Cols < MinCols || s.Rows < TargetRows || bh < 7.5*hexgrid.HexHeight {
					t.Errorf("%s: %d×%d cells, want at least %d columns and 7½ rows", name, s.Cols, s.Rows, MinCols)
				}
				// Either the height decides (exactly 7½ rows) or the width (7 columns).
				heightBound := math.Abs(bh*s.CellScale-float64(p.H)) < 1e-6 && s.Rows == TargetRows
				if !heightBound && s.Cols != MinCols {
					t.Errorf("%s: scale %.4f set by neither height nor width (%d×%d)", name, s.CellScale, s.Cols, s.Rows)
				}
			}
			if bw*s.CellScale > float64(p.W)+1e-6 || bh*s.CellScale > float64(p.H)+1e-6 {
				t.Errorf("%s: %d×%d block at %.4f does not fit panel %+v", name, s.Cols, s.Rows, s.CellScale, p)
			}
			// No letterboxing: one more column or row does not fit.
			if w1, _ := BlockSize(s.Cols+1, s.Rows); w1*s.CellScale <= float64(p.W) {
				t.Errorf("%s: room for column %d", name, s.Cols+1)
			}
			if _, h1 := BlockSize(s.Cols, s.Rows+1); h1*s.CellScale <= float64(p.H) {
				t.Errorf("%s: room for row %d", name, s.Rows+1)
			}
			if v := s.View(1, 3, 2); v.Cols != s.Cols || v.Rows != s.Rows || v.Panel != s.Panels[1] || v.Scale != s.CellScale {
				t.Errorf("%s: view %+v does not match the layout", name, v)
			}
		}
	}
}

func TestLayoutExamples(t *testing.T) {
	tests := []struct {
		w, h       int
		ui         float64
		panel      Rect
		cols, rows int
		scale      float64
	}{
		{1280, 720, 1, Rect{0, 0, 632, 680}, 7, 7, 680.0 / 600},
		{1920, 1080, 1, Rect{0, 0, 952, 1040}, 7, 7, 1040.0 / 600},
		{1920, 1080, 2, Rect{0, 0, 944, 1000}, 7, 7, 1000.0 / 600},
		{3840, 2160, 2, Rect{0, 0, 1904, 2080}, 7, 7, 2080.0 / 600},
		{3440, 1440, 1, Rect{0, 0, 1712, 1400}, 9, 7, 1400.0 / 600},
		{1024, 768, 1, Rect{0, 0, 504, 728}, 7, 8, 504.0 / 522},
	}
	for _, tc := range tests {
		s := Layout(tc.w, tc.h, tc.ui)
		if s.Panels[0] != tc.panel || s.Cols != tc.cols || s.Rows != tc.rows || math.Abs(s.CellScale-tc.scale) > 1e-9 {
			t.Errorf("%dx%d@%v: panel %+v %d×%d scale %.4f, want %+v %d×%d %.4f", tc.w, tc.h, tc.ui,
				s.Panels[0], s.Cols, s.Rows, s.CellScale, tc.panel, tc.cols, tc.rows, tc.scale)
		}
	}
}

// The default zoom shows a hex at the same share of the screen at 1080p and 4K,
// at any UI scale: 8.3–8.9 % of the width (the column step 6.1–6.6 %) and about
// 13 % of the height per cell, from the 7½ rows fitted to the panel height.
func TestDefaultHexShare(t *testing.T) {
	for _, tc := range []struct {
		w, h int
		ui   float64
	}{{1920, 1080, 1}, {1920, 1080, 2}, {2560, 1440, 1}, {3840, 2160, 1}, {3840, 2160, 2}} {
		s := Layout(tc.w, tc.h, tc.ui)
		hex := hexgrid.HexWidth * s.CellScale / float64(tc.w)
		step := hexgrid.HalfPitch * s.CellScale / float64(tc.w)
		cell := hexgrid.HexHeight * s.CellScale / float64(tc.h)
		if s.Zoom != DefaultZoom || hex < .083 || hex > .089 || step < .061 || step > .066 || cell < .123 || cell > .135 {
			t.Errorf("%dx%d@%v: hex %.2f %%, column step %.2f %% of the width, cell %.2f %% of the height", tc.w, tc.h, tc.ui, 100*hex, 100*step, 100*cell)
		}
	}
}

// The smallest window keeps 7×6 cells per panel at the smallest cell scale, and
// is tight: one device pixel less in either direction loses a column or a row.
func TestMinWindow(t *testing.T) {
	prevW, prevH := 0, 0
	for _, ui := range uiScales {
		w, h := MinWindow(ui)
		if w <= prevW || h <= prevH {
			t.Errorf("ui %v: min window %dx%d does not grow with the UI scale", ui, w, h)
		}
		prevW, prevH = w, h
		if s := Layout(w, h, ui); s.Cols < MinCols || s.Rows < MinRows {
			t.Errorf("ui %v: %dx%d shows %d×%d cells", ui, w, h, s.Cols, s.Rows)
		}
		if s := Layout(w-2, h, ui); s.Cols >= MinCols {
			t.Errorf("ui %v: %dx%d still shows %d columns", ui, w-2, h, s.Cols)
		}
		if s := Layout(w, h-1, ui); s.Rows >= MinRows {
			t.Errorf("ui %v: %dx%d still shows %d rows", ui, w, h-1, s.Rows)
		}
		if s := Layout(w*3, h, ui); s.CellScale < MinScale*ui-1e-9 || s.Cols <= MinCols {
			t.Errorf("ui %v: wide min-height window %+v", ui, s)
		}
	}
}

func TestLayoutDegenerate(t *testing.T) {
	for _, w := range [][2]int{{0, 0}, {1, 1}, {50, 30}, {200, 100}, {-5, 10}} {
		s := Layout(w[0], w[1], 0)
		if s.Cols < 1 || s.Rows < 1 || s.CellScale <= 0 || s.Panels[0].W < 1 || s.Panels[0].H < 1 {
			t.Errorf("%v: %+v", w, s)
		}
	}
}
