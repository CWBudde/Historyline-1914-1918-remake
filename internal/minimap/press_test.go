package minimap

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
)

// Only a press that started on the minimap drives it, for as long as it is held,
// even when it is dragged off the minimap; a press from the map never does.
func TestPressMustStartOnMinimap(t *testing.T) {
	type step struct{ down, inside, drives bool }
	for _, tc := range []struct {
		name  string
		steps []step
	}{
		{"click on the minimap", []step{{true, true, true}, {false, true, false}}},
		{"drag across the minimap", []step{{true, true, true}, {true, true, true}, {true, false, true}, {true, true, true}, {false, false, false}}},
		{"drag from the map onto it", []step{{true, false, false}, {true, true, false}, {true, true, false}, {false, true, false}}},
		{"hover without a press", []step{{false, true, false}, {false, false, false}}},
		{"new press after release", []step{{true, false, false}, {false, true, false}, {true, true, true}}},
		{"release outside ends the drag", []step{{true, true, true}, {true, false, true}, {false, false, false}, {true, false, false}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var p Press
			for i, s := range tc.steps {
				if got := p.Update(s.down, s.inside); got != s.drives {
					t.Fatalf("step %d (%+v): drives = %v", i, s, got)
				}
				if p.Captured() != s.drives {
					t.Fatalf("step %d: Captured = %v", i, p.Captured())
				}
			}
		})
	}
}

func TestClampKeepsDraggedPositionOnMinimap(t *testing.T) {
	m := Minimap{World: camera.Size{W: 400, H: 200}, Pos: camera.Point{X: 10, Y: 20}, Scale: 0.5}
	for _, tc := range []struct{ x, y, wx, wy float64 }{
		{50, 50, 50, 50},
		{0, 0, 10, 20},
		{500, 500, 210, 120},
	} {
		if x, y := m.Clamp(tc.x, tc.y); x != tc.wx || y != tc.wy {
			t.Errorf("Clamp(%v, %v) = %v, %v, want %v, %v", tc.x, tc.y, x, y, tc.wx, tc.wy)
		}
	}
}
