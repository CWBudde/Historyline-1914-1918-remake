package main

// Pure layout maths for the window and the HUD. Nothing here touches Ebitengine so
// it can be table-tested; main.go feeds it the window size and device scale factor.
// The map is drawn in device pixels and the HUD on a logical canvas magnified by
// the UI scale (see view).

const (
	// uiBaseW and uiBaseH are the physical window size per step of whole-number
	// UI magnification: every further 900x560 device pixels adds one step.
	uiBaseW = 900
	uiBaseH = 560

	hudMargin     = 8   // gap between HUD elements and the screen edge
	menuRowHeight = 18  // one action-menu row (the 16px debug font plus padding)
	menuMaxWidth  = 340 // widest action menu
	reportMaxW    = 360 // battle panel
	reportHeight  = 70
	reportTop     = 4
	minimapCap    = 280 // largest minimap edge
	minimapMin    = 80  // smallest minimap edge
	minimapMargin = 12
	bannerHeight  = 48
	glyphW        = 6 // ebitenutil debug font advance
)

// uiScale is the whole-number factor between logical HUD pixels and physical
// pixels. It follows the device scale factor (so text is as big as the OS intends)
// and grows with the window (so text stays readable at 4K) but never drops below 1.
func uiScale(outW, outH int, dpi float64) int {
	if dpi <= 0 {
		dpi = 1
	}
	physW, physH := float64(outW)*dpi, float64(outH)*dpi
	s := int(dpi + 0.5)
	if g := int(min(physW/uiBaseW, physH/uiBaseH)); g > s {
		s = g
	}
	return max(s, 1)
}

// logicalSize is what Layout returns: the window in physical pixels divided by the
// UI scale, so one logical pixel is a whole number of physical ones and the debug
// font stays crisp. It never returns less than 1x1.
func logicalSize(outW, outH int, dpi float64) (w, h int) {
	if dpi <= 0 {
		dpi = 1
	}
	s := float64(uiScale(outW, outH, dpi))
	w = int(float64(outW)*dpi/s + 0.5)
	h = int(float64(outH)*dpi/s + 0.5)
	return max(w, 1), max(h, 1)
}

// view is the screen geometry for one window size and device scale factor. The
// map is drawn at device resolution (Phys), the HUD at the logical size (Log) and
// magnified UI times, so the camera zoom, the UI scale and the asset density are
// independent of each other.
type view struct {
	PhysW, PhysH int // screen in device pixels (what Layout returns)
	UI           int // whole-number HUD magnification
	LogW, LogH   int // HUD canvas in logical pixels
}

func newView(outW, outH int, dpi float64) view {
	if dpi <= 0 {
		dpi = 1
	}
	v := view{
		PhysW: max(int(float64(outW)*dpi+0.5), 1),
		PhysH: max(int(float64(outH)*dpi+0.5), 1),
		UI:    uiScale(outW, outH, dpi),
	}
	v.LogW, v.LogH = logicalSize(outW, outH, dpi)
	return v
}

// toHUD maps a device-pixel cursor position to logical HUD pixels.
func (v view) toHUD(px, py int) (x, y int) {
	ui := max(v.UI, 1)
	return px / ui, py / ui
}

// assetDensity picks the pack density (1×, 2× or 4×) for a map drawn at eff device
// pixels per reference pixel: the smallest that is not magnified, capped at 4×.
func assetDensity(eff float64) int {
	switch {
	case eff <= 1:
		return 1
	case eff <= 2:
		return 2
	}
	return 4
}

// rect is an integer screen rectangle.
type rect struct{ X, Y, W, H int }

// hudLayout holds the screen rectangles of the HUD for one logical screen size.
type hudLayout struct {
	MenuX, MenuW int // action menu column; rows grow upwards from the bottom
	MenuBottom   int // y just below the last menu row
	Report       rect
	MinimapMax   int // longest minimap edge
	Banner       rect
	DevRight     int // right edge for the dev indicator text
}

// layoutHUD places the HUD for a w x h logical screen. Everything shrinks to fit
// narrow screens instead of using fixed positions.
func layoutHUD(w, h int) hudLayout {
	var l hudLayout
	l.MenuX = hudMargin
	l.MenuW = max(min(menuMaxWidth, w-2*hudMargin), 1)
	l.MenuBottom = h - hudMargin
	rw := max(min(reportMaxW, w-2*reportTop), 1)
	l.Report = rect{(w - rw) / 2, reportTop, rw, reportHeight}
	l.MinimapMax = max(min(minimapCap, min(w, h)*3/10), minimapMin)
	l.Banner = rect{0, (h - bannerHeight) / 2, w, bannerHeight}
	l.DevRight = w - hudMargin
	return l
}

// menuRowRect is the rectangle of row i of n.
func (l hudLayout) menuRowRect(i, n int) rect {
	return rect{l.MenuX, l.MenuBottom - (n-i)*menuRowHeight, l.MenuW, menuRowHeight}
}

// textWidth is the pixel width of s in the debug font.
func textWidth(s string) int { return glyphW * len(s) }
