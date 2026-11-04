package llmplay

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEncodeMP4Args(t *testing.T) {
	got := encodeMP4Args("g/frames/*.png", "g/game.mp4", 0, 1440, 28)
	want := []string{
		"-y", "-framerate", "1", "-pattern_type", "glob", "-i", "g/frames/*.png",
		"-vf", "scale=1440:-2", "-c:v", "libx264", "-preset", "veryfast", "-crf", "28",
		"-pix_fmt", "yuv420p", "-movflags", "+faststart", "g/game.mp4",
	}
	if !slices.Equal(got, want) {
		t.Errorf("args %q\nwant %q", got, want)
	}
}

func TestVideoWriterArgs(t *testing.T) {
	got := strings.Join(videoWriterArgs("out.mp4", 64, 48, 30, 23), " ")
	want := "-y -hide_banner -loglevel error -f rawvideo -pix_fmt rgba -s 64x48 -framerate 30 -i - " +
		"-c:v libx264 -preset veryfast -crf 23 -pix_fmt yuv420p -movflags +frag_keyframe+empty_moov+default_base_moof out.mp4"
	if got != want {
		t.Errorf("args %q\nwant %q", got, want)
	}
}

// Odd or empty sizes and a frame rate below 1 are refused before ffmpeg
// starts.
func TestNewVideoWriterRefuses(t *testing.T) {
	for _, c := range []struct{ w, h, fps int }{{63, 48, 2}, {64, 47, 2}, {0, 48, 2}, {64, 48, 0}} {
		if _, err := NewVideoWriter(context.Background(), filepath.Join(t.TempDir(), "x.mp4"), c.w, c.h, c.fps, 23); err == nil {
			t.Errorf("%dx%d at %d fps: no error", c.w, c.h, c.fps)
		}
	}
}

// WriteFrame takes exactly w*h*4 bytes.
func TestWriteFrameSize(t *testing.T) {
	v := &VideoWriter{w: 4, h: 2}
	for _, n := range []int{0, 31, 33} {
		if err := v.WriteFrame(make([]byte, n)); err == nil || !strings.Contains(err.Error(), "want 32") {
			t.Errorf("%d bytes: err %v", n, err)
		}
	}
	if v.Frames() != 0 {
		t.Errorf("frames %d after refusals", v.Frames())
	}
}

func needFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
}

func testFrame(w, h, i int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{uint8(x * 4), uint8(y * 5), uint8(i * 80), 255})
		}
	}
	return img
}

func TestVideoWriterEncodes(t *testing.T) {
	needFFmpeg(t)
	out := filepath.Join(t.TempDir(), "v.mp4")
	v, err := NewVideoWriter(context.Background(), out, 64, 48, 2, 28)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 3 {
		if err := v.WriteFrame(testFrame(64, 48, i).Pix); err != nil {
			t.Fatal(err)
		}
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if v.Frames() != 3 {
		t.Errorf("frames %d, want 3", v.Frames())
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Fatalf("video: %v, %v", st, err)
	}
	if err := v.WriteFrame(testFrame(64, 48, 0).Pix); err == nil {
		t.Error("frame after Close taken")
	}
}

func TestEncodeMP4(t *testing.T) {
	needFFmpeg(t)
	dir := t.TempDir()
	for i := range 3 {
		f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%04d.png", i)))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, testFrame(64, 48, i)); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(dir, "game.mp4")
	if err := EncodeMP4(context.Background(), filepath.Join(dir, "*.png"), out, 2, 32, 28); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Fatalf("video: %v, %v", st, err)
	}
}

// A failing ffmpeg run reports the tail of its stderr.
func TestEncodeMP4Error(t *testing.T) {
	needFFmpeg(t)
	dir := t.TempDir()
	err := EncodeMP4(context.Background(), filepath.Join(dir, "*.png"), filepath.Join(dir, "game.mp4"), 2, 32, 28)
	if err == nil || !strings.HasPrefix(err.Error(), "ffmpeg: exit status") || !strings.Contains(err.Error(), "game.mp4") {
		t.Errorf("err %v, want ffmpeg's exit with its stderr", err)
	}
}

// tailBuffer keeps whole lines of the last stderrTail bytes.
func TestTailBuffer(t *testing.T) {
	var b tailBuffer
	for i := range 200 {
		if _, err := fmt.Fprintf(&b, "line %03d\n", i); err != nil {
			t.Fatal(err)
		}
	}
	got := b.String()
	if len(got) > stderrTail || !strings.HasPrefix(got, "line ") || !strings.HasSuffix(got, "line 199\n") {
		t.Errorf("tail %d bytes %q…", len(got), got[:min(len(got), 20)])
	}
}

// TestVideoWriterKilled: a recording killed mid-way leaves a readable video.
func TestVideoWriterKilled(t *testing.T) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("no ffprobe")
	}
	out := filepath.Join(t.TempDir(), "killed.mp4")
	v, err := NewVideoWriter(context.Background(), out, 64, 48, 30, 23)
	if errors.Is(err, ErrNoFFmpeg) {
		t.Skip("no ffmpeg")
	}
	if err != nil {
		t.Fatal(err)
	}
	pix := make([]byte, 64*48*4)
	for f := range 600 { // past the keyframe interval of 250 frames
		for i := range pix {
			pix[i] = byte(i*f + i>>3)
		}
		if err := v.WriteFrame(pix); err != nil {
			t.Fatal(err)
		}
	}
	readable := func() bool {
		return exec.CommandContext(context.Background(), ffprobe, "-v", "error", "-show_entries", "format=duration", out).Run() == nil
	}
	for end := time.Now().Add(10 * time.Second); !readable() && time.Now().Before(end); {
		time.Sleep(50 * time.Millisecond)
	}
	if err := v.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err == nil {
		t.Error("Close reports no error for a killed ffmpeg")
	}
	raw, err := exec.CommandContext(context.Background(), ffprobe, "-v", "error", "-select_streams", "v", "-count_packets",
		"-show_entries", "stream=nb_read_packets", "-of", "csv=p=0", out).Output()
	if err != nil {
		t.Fatalf("killed video unreadable: %v", err)
	}
	if n, err := strconv.Atoi(strings.TrimSpace(strings.Split(string(raw), "\n")[0])); err != nil || n == 0 {
		t.Errorf("killed video holds %q frames", raw)
	}
}
