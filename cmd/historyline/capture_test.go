package main

import (
	"math"
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

func TestLookCapturePlan(t *testing.T) {
	frames := reviewFramePlan(true)
	if len(frames) != 12 {
		t.Fatalf("got %d look frames, want 12", len(frames))
	}
	world := camera.Size{W: 1200, H: 2000}
	seen := map[string]bool{}
	for _, frame := range frames {
		if frame.width != 1920 && frame.width != 3840 || frame.height*16 != frame.width*9 {
			t.Errorf("unexpected look framebuffer: %+v", frame)
		}
		if frame.effectiveZoom != 0.5 && frame.effectiveZoom != 1 && frame.effectiveZoom != 1.6 {
			t.Errorf("unexpected look scale: %+v", frame)
		}
		if seen[frame.filename()] {
			t.Errorf("look frame overwrites %s", frame.filename())
		}
		seen[frame.filename()] = true
		for _, ui := range []int{1, 2, 3, 4} {
			if got := frame.cameraZoom(world, ui) * float64(ui); math.Abs(got-frame.effectiveZoom) > 1e-12 {
				t.Errorf("%s at UI %d: effective zoom %v, want %v", frame.filename(), ui, got, frame.effectiveZoom)
			}
		}
	}
	for _, filename := range []string{
		"pulse-look-0.5x-revealed-1920x1080.png", "pulse-look-0.5x-fog-1920x1080.png",
		"pulse-look-1x-revealed-1920x1080.png", "pulse-look-1x-fog-1920x1080.png",
		"pulse-look-1.6x-revealed-1920x1080.png", "pulse-look-1.6x-fog-1920x1080.png",
		"pulse-look-1x-revealed-3840x2160.png", "pulse-look-1x-fog-3840x2160.png",
		"pulse-look-1.6x-revealed-3840x2160.png", "pulse-look-1.6x-fog-3840x2160.png",
	} {
		if !seen[filename] {
			t.Errorf("missing %s", filename)
		}
	}
}

func TestDefaultCapturePlan(t *testing.T) {
	frames := reviewFramePlan(false)
	if len(frames) != 9 {
		t.Fatalf("got %d default frames, want 9", len(frames))
	}
	world := camera.Size{W: 1200, H: 2000}
	for i, want := range []struct {
		filename string
		zoom     float64
		reveal   bool
	}{
		{"pulse-overview-1280x720.png", .1, true},
		{"pulse-detail-1280x720.png", 1.0 / 3, true},
		{"pulse-fog-1280x720.png", 1.0 / 3, false},
		{"pulse-overview-1920x1080.png", .15, true},
		{"pulse-detail-1920x1080.png", .5, true},
		{"pulse-fog-1920x1080.png", .5, false},
		{"pulse-overview-3840x2160.png", .3, true},
		{"pulse-detail-3840x2160.png", 1, true},
		{"pulse-fog-3840x2160.png", 1, false},
	} {
		frame := frames[i]
		if frame.filename() != want.filename || math.Abs(frame.cameraZoom(world, 3)-want.zoom) > 1e-12 || frame.reveal != want.reveal {
			t.Errorf("frame %d: %s zoom %v reveal %v; want %+v", i, frame.filename(), frame.cameraZoom(world, 3), frame.reveal, want)
		}
	}
}
