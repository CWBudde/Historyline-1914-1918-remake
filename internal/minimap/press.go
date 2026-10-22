package minimap

// Press tracks one press (mouse button or touch) for the minimap: only a press that
// started on the minimap drives it, for as long as it is held, even when it is
// dragged off it (pass the position through Minimap.Clamp). A press that started on
// the map never does, so dragging the map across the minimap does not jump the
// camera. The first frame seen down counts as the start of the press.
type Press struct {
	held, onMinimap bool
}

// Update takes this frame's button state and whether the pointer is on the
// minimap, and reports whether the pointer drives the minimap this frame.
func (p *Press) Update(down, inside bool) bool {
	switch {
	case !down:
		p.held, p.onMinimap = false, false
	case !p.held:
		p.held, p.onMinimap = true, inside
	}
	return p.onMinimap
}

// Captured reports whether the press in progress belongs to the minimap, so the
// map should ignore it.
func (p *Press) Captured() bool { return p.onMinimap }
