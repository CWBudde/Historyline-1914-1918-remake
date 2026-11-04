package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestParseRecordSize(t *testing.T) {
	for _, tt := range []struct {
		in   string
		w, h int
		err  string
	}{
		{"1920x1080", 1920, 1080, ""},
		{"640x360", 640, 360, ""},
		{"1921x1080", 0, 0, "even"},
		{"1920x1081", 0, 0, "even"},
		{"1920", 0, 0, "WxH"},
		{"0x1080", 0, 0, "WxH"},
		{"ax2", 0, 0, "WxH"},
	} {
		w, h, err := parseRecordSize(tt.in)
		if tt.err == "" && (err != nil || w != tt.w || h != tt.h) {
			t.Errorf("%q: %dx%d, %v; want %dx%d", tt.in, w, h, err, tt.w, tt.h)
		}
		if tt.err != "" && (err == nil || !strings.Contains(err.Error(), tt.err)) {
			t.Errorf("%q: error %v, want one about %s", tt.in, err, tt.err)
		}
	}
}

func TestRecordCheck(t *testing.T) {
	ok := recordConfig{out: "build/review/llmplay/g/screen.mp4", w: 1920, h: 1080, fps: 30, speed: 2, tail: 3 * time.Second, crf: 23}
	with := func(f func(*recordConfig)) recordConfig {
		c := ok
		f(&c)
		return c
	}
	const replay = "build/review/llmplay/g"
	for _, tt := range []struct {
		name    string
		c       recordConfig
		replay  string
		capture string
		reduced bool
		set     []string
		err     string
	}{
		{name: "valid", c: ok, replay: replay},
		{name: "no record", c: recordConfig{}},
		{name: "record flag without record", c: recordConfig{}, set: []string{"record-fps"}, err: "-record-fps requires -record"},
		{name: "no replay", c: ok, err: "requires -replay"},
		{name: "capture", c: ok, replay: replay, capture: "build/x", err: "exclude"},
		{name: "reduced", c: ok, replay: replay, reduced: true, err: "reduced-motion"},
		{name: "outside build", c: with(func(c *recordConfig) { c.out = "screen.mp4" }), replay: replay, err: "under ignored build/"},
		{name: "escapes build", c: with(func(c *recordConfig) { c.out = "build/../x.mp4" }), replay: replay, err: "under ignored build/"},
		{name: "not mp4", c: with(func(c *recordConfig) { c.out = "build/x.mkv" }), replay: replay, err: ".mp4"},
		{name: "fps", c: with(func(c *recordConfig) { c.fps = 0 }), replay: replay, err: "-record-fps"},
		{name: "speed", c: with(func(c *recordConfig) { c.speed = 0 }), replay: replay, err: "-record-speed"},
		{name: "tail", c: with(func(c *recordConfig) { c.tail = -time.Second }), replay: replay, err: "-record-tail"},
		{name: "crf", c: with(func(c *recordConfig) { c.crf = 52 }), replay: replay, err: "-record-crf"},
		{name: "max frames", c: with(func(c *recordConfig) { c.maxFrames = -1 }), replay: replay, err: "-record-max-frames"},
	} {
		set := map[string]bool{}
		for _, f := range tt.set {
			set[f] = true
		}
		err := tt.c.check(tt.replay, tt.capture, tt.reduced, set)
		if tt.err == "" && err != nil {
			t.Errorf("%s: %v", tt.name, err)
		}
		if tt.err != "" && (err == nil || !strings.Contains(err.Error(), tt.err)) {
			t.Errorf("%s: error %v, want one about %q", tt.name, err, tt.err)
		}
	}
	if got := ok.tailFrames(); got != 90 {
		t.Errorf("3 s at 30 fps: %d tail frames, want 90", got)
	}
}

// countSink counts the frames and checks their size.
type countSink struct {
	t      *testing.T
	size   int
	frames int
}

func (s *countSink) WriteFrame(pix []byte) error {
	if len(pix) != s.size {
		s.t.Fatalf("frame of %d bytes, want %d", len(pix), s.size)
	}
	s.frames++
	return nil
}

// recordTest records the replay in dir into sink at cfg through the
// recorder's core loop. Outside ebiten.RunGame the pixels cannot be read
// back, so the frames stay blank; drawing and the frame count are real.
func recordTest(t *testing.T, dir string, sink frameSink, cfg recordConfig) *recorder {
	t.Helper()
	rp, err := llmplay.OpenReplay(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := newRecorder(replayGame(rp, false), sink, cfg)
	r.read = func(*ebiten.Image, []byte) {}
	for r.step() {
		if r.frames > 100000 {
			t.Fatalf("no end after %d frames", r.frames)
		}
	}
	return r
}

// TestRecordReplay: a replay records to its end plus the tail, the same
// frames every run at a fixed speed, and the frame limit cuts it short.
func TestRecordReplay(t *testing.T) {
	dir := recordRhein(t, 2)
	cfg := recordConfig{w: 320, h: 180, fps: 30, speed: 16, tail: time.Second, crf: 23}
	var runs [2]int
	for i := range runs {
		sink := &countSink{t: t, size: 4 * cfg.w * cfg.h}
		r := recordTest(t, dir, sink, cfg)
		if r.err != nil {
			t.Fatal(r.err)
		}
		if !r.g.dos.replayDone() || r.g.dos.replay.at != 2 {
			t.Fatalf("recording ended before the replay: line %d", r.g.dos.replay.at)
		}
		if sink.frames != r.frames || r.tail != 0 {
			t.Fatalf("sink took %d of %d frames, tail left %d", sink.frames, r.frames, r.tail)
		}
		runs[i] = r.frames
	}
	t.Logf("%d frames", runs[0])
	if runs[0] != runs[1] {
		t.Errorf("two runs recorded %d and %d frames", runs[0], runs[1])
	}
	if runs[0] <= cfg.tailFrames() {
		t.Errorf("%d frames: no more than the tail", runs[0])
	}
	cfg.speed = 32
	if r := recordTest(t, dir, &countSink{t: t, size: 4 * cfg.w * cfg.h}, cfg); r.frames >= runs[0] || r.frames <= cfg.tailFrames() {
		t.Errorf("twice the speed: %d frames, at the speed before %d", r.frames, runs[0])
	}
	cfg.maxFrames = 7
	if r := recordTest(t, dir, &countSink{t: t, size: 4 * cfg.w * cfg.h}, cfg); r.err != nil || r.frames != 7 {
		t.Errorf("limit 7: %d frames, %v", r.frames, r.err)
	}
}

// TestRecordStopsOnMismatch: a replay mismatch ends the recording with
// the error.
func TestRecordStopsOnMismatch(t *testing.T) {
	dir := recordRhein(t, 2)
	path := filepath.Join(dir, "transcript.jsonl")
	_, lines, err := llmplay.ReadTranscript(dir)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bad := strings.Repeat("0", len(lines[0].StateAfter))
	raw = []byte(strings.Replace(string(raw), `"stateAfter":"`+lines[0].StateAfter, `"stateAfter":"`+bad, 1))
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := recordConfig{w: 320, h: 180, fps: 30, speed: 16, tail: time.Second}
	r := recordTest(t, dir, &countSink{t: t, size: 4 * cfg.w * cfg.h}, cfg)
	if r.err == nil || !strings.Contains(r.err.Error(), "transcript has "+bad) {
		t.Fatalf("recording error %v, want the mismatch", r.err)
	}
}

// TestRecordVideo: the recorder writes a playable video through ffmpeg.
func TestRecordVideo(t *testing.T) {
	dir := recordRhein(t, 1)
	out := filepath.Join(t.TempDir(), "screen.mp4")
	cfg := recordConfig{w: 320, h: 180, fps: 30, speed: 16, tail: time.Second, crf: 23, maxFrames: 12}
	v, err := llmplay.NewVideoWriter(context.Background(), out, cfg.w, cfg.h, cfg.fps, cfg.crf)
	if errors.Is(err, llmplay.ErrNoFFmpeg) {
		t.Skip("ffmpeg not found")
	}
	if err != nil {
		t.Fatal(err)
	}
	r := recordTest(t, dir, v, cfg)
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if r.err != nil || v.Frames() != 12 {
		t.Fatalf("%d frames written, %v", v.Frames(), r.err)
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Fatalf("video %s: %v", out, err)
	}
}
