package main

import "testing"

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

func TestLayoutHUDFitsScreen(t *testing.T) {
	for _, sz := range [][2]int{{400, 700}, {320, 200}, {1280, 720}, {1800, 1000}, {50, 50}} {
		w, h := sz[0], sz[1]
		l := layoutHUD(w, h)
		if l.MenuX+l.MenuW > w && w > 2*hudMargin {
			t.Errorf("%dx%d: menu overflows", w, h)
		}
		if l.Report.X < 0 || l.Report.X+l.Report.W > w {
			t.Errorf("%dx%d: report %+v outside screen", w, h, l.Report)
		}
		if l.MinimapMax < minimapMin || l.MinimapMax > minimapCap {
			t.Errorf("%dx%d: minimap max %d", w, h, l.MinimapMax)
		}
		if l.Banner.W != w || l.Banner.Y != (h-bannerHeight)/2 {
			t.Errorf("%dx%d: banner %+v", w, h, l.Banner)
		}
	}
	l := layoutHUD(1800, 1000)
	if l.MenuW != menuMaxWidth || l.MinimapMax != minimapCap || l.Report.X != (1800-reportMaxW)/2 {
		t.Errorf("wide layout %+v", l)
	}
	if narrow := layoutHUD(300, 700); narrow.MenuW != 284 || narrow.Report.W != 292 {
		t.Errorf("narrow layout %+v", narrow)
	}
}

func TestMenuRowRect(t *testing.T) {
	l := layoutHUD(800, 600)
	if r := l.menuRowRect(2, 3); r.Y != 600-hudMargin-menuRowHeight || r.X != hudMargin {
		t.Errorf("last row %+v", r)
	}
	if a, b := l.menuRowRect(0, 3), l.menuRowRect(1, 3); b.Y-a.Y != menuRowHeight {
		t.Errorf("rows not contiguous: %+v %+v", a, b)
	}
}
