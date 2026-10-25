package dosscreen

// Zooms are the zoom levels as factors on the fitted cell scale, from far out to
// close in, about a tenth apart (seven steps per doubling) so one wheel step
// zooms gently. Zooms[DefaultZoom] is 1, the scale Layout fits to the panel.
var Zooms = [...]float64{0.5, 0.55, 0.61, 0.67, 0.74, 0.82, 0.91, 1, 1.1, 1.22, 1.35, 1.49, 1.64, 1.81, 2}

// DefaultZoom is the index in Zooms of the fitted scale.
const DefaultZoom = 7

// ZoomCols and ZoomRows are the fewest columns and cells per column a panel
// shows when zoomed in, fewer than the smallest window keeps (MinCols ×
// MinRows), so a fitted panel can still zoom in.
const (
	ZoomCols = 4
	ZoomRows = 4
)

// Zoomed returns s at zoom level i: the same rects with the fitted cell scale
// times Zooms[i] and the columns and rows refitted. An unusable level is moved to
// the nearest usable one (ClampZoom), so Zoomed can restore a saved level after
// the window changed.
func (s Screen) Zoomed(i int) Screen {
	s.Zoom = s.ClampZoom(i)
	s.CellScale = s.zoomScale(s.Zoom)
	s.Cols, s.Rows = Fit(s.Panels[0], s.CellScale)
	return s
}

// ZoomLevels returns the usable zoom levels in rising order: those at which the
// cell scale stays at or above MinScale × ui and each panel still shows
// ZoomCols × ZoomRows cells. DefaultZoom is always one of them, and they form a run.
func (s Screen) ZoomLevels() []int {
	lo, hi := s.zoomRange()
	out := make([]int, 0, hi-lo+1)
	for i := lo; i <= hi; i++ {
		out = append(out, i)
	}
	return out
}

// ClampZoom returns the usable zoom level nearest to i.
func (s Screen) ClampZoom(i int) int {
	lo, hi := s.zoomRange()
	return min(max(i, lo), hi)
}

// zoomRange is the first and last usable zoom level. Zooming out only adds cells,
// so the levels below the default end at the scale floor; zooming in only takes
// cells away, so the levels above it end at the first one short of cells.
func (s Screen) zoomRange() (lo, hi int) {
	lo, hi = DefaultZoom, DefaultZoom
	for lo > 0 && s.zoomScale(lo-1) >= MinScale*s.UI-eps {
		lo--
	}
	for hi < len(Zooms)-1 {
		sc := s.zoomScale(hi + 1)
		if cols, rows := Fit(s.Panels[0], sc); sc < MinScale*s.UI-eps || cols < ZoomCols || rows < ZoomRows {
			break
		}
		hi++
	}
	return lo, hi
}

// zoomScale is the cell scale at zoom level i.
func (s Screen) zoomScale(i int) float64 {
	if i == DefaultZoom {
		return s.FitScale
	}
	return s.FitScale * Zooms[i]
}
