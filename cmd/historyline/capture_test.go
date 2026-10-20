package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
)

// Catch captures that accidentally change the world area when only output
// resolution changes, or ignore the HUD's independent display scale.
func TestCaptureFraming(t *testing.T) {
	world := camera.Size{W: 1200, H: 2000}
	for _, tc := range []struct {
		w, h, ui int
		want     float64
	}{
		{1280, 720, 1, .3}, {1920, 1080, 1, .45}, {3840, 2160, 3, .3},
	} {
		if got := captureZoom(world, tc.w, tc.h, tc.ui); got != tc.want {
			t.Errorf("%dx%d at UI %d: zoom %v, want %v", tc.w, tc.h, tc.ui, got, tc.want)
		}
	}
}
