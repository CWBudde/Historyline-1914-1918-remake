package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/hajimehoshi/ebiten/v2"
)

// A replay records as a video of the player's view (-record): the split
// battle screen at a fixed size, a video frame per Update, each advancing
// the game a fixed number of ticks without input, so the video takes the
// same frames however fast the machine draws them.

// recordConfig is what -record and its -record-* flags ask for.
type recordConfig struct {
	out       string        // the video, under build/
	w, h      int           // the frame size, even
	fps       int           // video frames per second
	speed     int           // game ticks per video frame (2 at 30 fps is real time)
	tail      time.Duration // video after the replay is done
	crf       int           // x264 quality
	maxFrames int           // 0 records to the end
}

// The record flags beside -record itself.
var recordFlags = []string{"record-size", "record-fps", "record-speed", "record-tail", "record-crf", "record-max-frames"}

// parseRecordSize reads WxH, both even and positive.
func parseRecordSize(s string) (w, h int, err error) {
	ws, hs, ok := strings.Cut(s, "x")
	if ok {
		w, err = strconv.Atoi(ws)
		if err == nil {
			h, err = strconv.Atoi(hs)
		}
	}
	if !ok || err != nil || w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("-record-size %q: want WxH, such as 1920x1080", s)
	}
	if w%2 != 0 || h%2 != 0 {
		return 0, 0, fmt.Errorf("-record-size %q: width and height must be even", s)
	}
	return w, h, nil
}

// check validates c against the other flags: -record needs -replay
// and excludes -capture and -reduced-motion, the -record-* flags need
// -record; set holds the names of the flags given.
func (c recordConfig) check(replay, capture string, reduced bool, set map[string]bool) error {
	if c.out == "" {
		for _, f := range recordFlags {
			if set[f] {
				return fmt.Errorf("-%s requires -record", f)
			}
		}
		return nil
	}
	clean := filepath.ToSlash(filepath.Clean(c.out))
	switch {
	case replay == "":
		return errors.New("-record requires -replay")
	case capture != "":
		return errors.New("-record and -capture exclude each other")
	case reduced:
		return errors.New("-record shows normal motion: drop -reduced-motion")
	case !strings.HasPrefix(clean, "build/") || strings.Contains(clean, ".."):
		return fmt.Errorf("-record %q: the video must be under ignored build/", c.out)
	case strings.ToLower(filepath.Ext(clean)) != ".mp4":
		return fmt.Errorf("-record %q: the video must be an .mp4 file", c.out)
	case c.fps < 1:
		return fmt.Errorf("-record-fps %d: must be at least 1", c.fps)
	case c.speed < 1:
		return fmt.Errorf("-record-speed %d: must be at least 1 tick per frame", c.speed)
	case c.tail < 0:
		return fmt.Errorf("-record-tail %v: must not be negative", c.tail)
	case c.crf < 0 || c.crf > 51:
		return fmt.Errorf("-record-crf %d: want 0 to 51", c.crf)
	case c.maxFrames < 0:
		return fmt.Errorf("-record-max-frames %d: must not be negative", c.maxFrames)
	}
	return nil
}

// tailFrames is the tail in video frames.
func (c recordConfig) tailFrames() int { return int(math.Round(c.tail.Seconds() * float64(c.fps))) }

// frameSink takes the recorded frames (llmplay.VideoWriter).
type frameSink interface {
	WriteFrame(pix []byte) error
}

// recorder is the core of a recording: advance, draw, write.
type recorder struct {
	g      *game
	sink   frameSink
	cfg    recordConfig
	target *ebiten.Image
	pix    []byte
	read   func(*ebiten.Image, []byte) // the frame's pixels: ReadPixels
	frames int
	tail   int  // video frames still to write after the replay is done, -1 before
	done   bool // the video is complete, or failed (err)
	err    error
}

// newRecorder lays g out at the record size and records it into sink.
func newRecorder(g *game, sink frameSink, cfg recordConfig) *recorder {
	g.applyLayout(cfg.w, cfg.h, 1)
	return &recorder{
		g: g, sink: sink, cfg: cfg,
		target: ebiten.NewImage(cfg.w, cfg.h),
		pix:    make([]byte, 4*cfg.w*cfg.h),
		read:   (*ebiten.Image).ReadPixels,
		tail:   -1,
	}
}

// step advances the game by the speed's ticks and writes the frame it
// shows. It reports false once the video is complete or failed: the
// replay done and the tail written, the frame limit reached, a replay
// mismatch or a write error (err).
func (r *recorder) step() bool {
	if r.done {
		return false
	}
	g := r.g
	for range r.cfg.speed {
		g.advance()
	}
	r.target.Clear()
	g.Draw(r.target)
	if g.artErr != "" || g.terrainErr != "" {
		return r.fail(fmt.Errorf("record rendering: %s %s", g.artErr, g.terrainErr))
	}
	r.read(r.target, r.pix)
	if err := r.sink.WriteFrame(r.pix); err != nil {
		return r.fail(err)
	}
	r.frames++
	switch {
	case g.dos.replayErr() != nil:
		return r.fail(fmt.Errorf("replay stopped at frame %d: %w", r.frames, g.dos.replayErr()))
	case r.cfg.maxFrames > 0 && r.frames >= r.cfg.maxFrames:
		r.done = true
	case r.tail < 0 && g.dos.replayDone():
		r.tail = r.cfg.tailFrames()
		r.done = r.tail == 0
	case r.tail > 0:
		r.tail--
		r.done = r.tail == 0
	}
	return !r.done
}

func (r *recorder) fail(err error) bool {
	r.err, r.done = err, true
	return false
}

// recordGame runs a recorder under ebiten.RunGame: Update records a frame,
// Draw previews the last one on the window. It never reads input.
type recordGame struct{ *recorder }

func (r *recordGame) Update() error {
	if !r.step() {
		return ebiten.Termination
	}
	return nil
}

func (*recordGame) Layout(int, int) (int, int) { return 1280, 720 }

func (r *recordGame) Draw(screen *ebiten.Image) {
	var op ebiten.DrawImageOptions
	s := math.Min(1280/float64(r.cfg.w), 720/float64(r.cfg.h))
	op.GeoM.Scale(s, s)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(r.target, &op)
}

// recordReplay records g's replay into cfg.out as fast as the machine
// draws and reports the video on out. The video is closed also when the
// replay fails.
func recordReplay(g *game, cfg recordConfig, out io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(cfg.out), 0o755); err != nil {
		return err
	}
	v, err := llmplay.NewVideoWriter(context.Background(), cfg.out, cfg.w, cfg.h, cfg.fps, cfg.crf)
	if err != nil {
		return err
	}
	r := newRecorder(g, v, cfg)
	ebiten.SetTPS(ebiten.SyncWithFPS) // an Update per drawn frame, not per 1/60 s
	ebiten.SetVsyncEnabled(false)
	err = errors.Join(ebiten.RunGame(&recordGame{r}), r.err)
	if err == nil && !r.done {
		err = fmt.Errorf("recording interrupted after %d frames", r.frames)
	}
	err = errors.Join(err, v.Close())
	secs := float64(r.frames) / float64(cfg.fps)
	if _, perr := fmt.Fprintf(out, "recorded %d frames (%.1f s at %d fps) to %s\n", r.frames, secs, cfg.fps, cfg.out); perr != nil {
		err = errors.Join(err, perr)
	}
	return err
}
