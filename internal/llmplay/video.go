package llmplay

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// ErrNoFFmpeg is returned when ffmpeg is not on the PATH.
var ErrNoFFmpeg = errors.New("ffmpeg not found")

// stderrTail is how much of ffmpeg's stderr an error carries at most.
const stderrTail = 1 << 10

func ffmpegPath() (string, error) {
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoFFmpeg, err)
	}
	return p, nil
}

// encodeMP4Args are ffmpeg's arguments for EncodeMP4.
func encodeMP4Args(framesGlob, out string, fps, width, crf int) []string {
	return []string{
		"-y",
		"-framerate", fmt.Sprint(max(fps, 1)),
		"-pattern_type", "glob", "-i", framesGlob,
		"-vf", fmt.Sprintf("scale=%d:-2", width),
		"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprint(crf),
		"-pix_fmt", "yuv420p", "-movflags", "+faststart", out,
	}
}

// EncodeMP4 encodes the PNG frames matching framesGlob into the H.264
// video out, scaled to width (the height follows the frames). It returns
// ErrNoFFmpeg without ffmpeg, and ffmpeg's error with the tail of its
// stderr.
func EncodeMP4(ctx context.Context, framesGlob, out string, fps, width, crf int) error {
	ffmpeg, err := ffmpegPath()
	if err != nil {
		return err
	}
	var stderr tailBuffer
	cmd := exec.CommandContext(ctx, ffmpeg, encodeMP4Args(framesGlob, out, fps, width, crf)...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return ffmpegError(err, &stderr)
	}
	return nil
}

// VideoWriter streams RGBA frames into an H.264 video through ffmpeg.
type VideoWriter struct {
	w, h   int
	frames int
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr tailBuffer
	closed bool
}

// videoWriterArgs are ffmpeg's arguments for a VideoWriter. The file is
// fragmented, so all but its last fragment stays playable when the
// recording is killed.
func videoWriterArgs(out string, w, h, fps, crf int) []string {
	return []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-f", "rawvideo", "-pix_fmt", "rgba", "-s", fmt.Sprintf("%dx%d", w, h),
		"-framerate", fmt.Sprint(fps), "-i", "-",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprint(crf),
		"-pix_fmt", "yuv420p", "-movflags", "+frag_keyframe+empty_moov+default_base_moof", out,
	}
}

// NewVideoWriter starts ffmpeg encoding w×h RGBA frames at fps into out.
// The size must be even (x264 with yuv420p). It returns ErrNoFFmpeg
// without ffmpeg.
func NewVideoWriter(ctx context.Context, out string, w, h, fps, crf int) (*VideoWriter, error) {
	if w <= 0 || h <= 0 || w%2 != 0 || h%2 != 0 {
		return nil, fmt.Errorf("video size %dx%d: width and height must be positive and even", w, h)
	}
	if fps < 1 {
		return nil, fmt.Errorf("video frame rate %d: must be at least 1", fps)
	}
	ffmpeg, err := ffmpegPath()
	if err != nil {
		return nil, err
	}
	v := &VideoWriter{w: w, h: h}
	v.cmd = exec.CommandContext(ctx, ffmpeg, videoWriterArgs(out, w, h, fps, crf)...)
	v.cmd.Stderr = &v.stderr
	if v.stdin, err = v.cmd.StdinPipe(); err != nil {
		return nil, err
	}
	if err := v.cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w", err)
	}
	return v, nil
}

// WriteFrame writes one frame of w*h*4 RGBA bytes.
func (v *VideoWriter) WriteFrame(pix []byte) error {
	if v.closed {
		return errors.New("video writer closed")
	}
	if want := v.w * v.h * 4; len(pix) != want {
		return fmt.Errorf("video frame of %d bytes, want %d (%dx%d RGBA)", len(pix), want, v.w, v.h)
	}
	if _, err := v.stdin.Write(pix); err != nil {
		return fmt.Errorf("ffmpeg: frame %d: %w", v.frames, err)
	}
	v.frames++
	return nil
}

// Frames is the number of frames written.
func (v *VideoWriter) Frames() int { return v.frames }

// Close ends the stream and waits for ffmpeg to finish the file. It
// returns ffmpeg's error with the tail of its stderr.
func (v *VideoWriter) Close() error {
	if v.closed {
		return nil
	}
	v.closed = true
	cerr := v.stdin.Close()
	if err := v.cmd.Wait(); err != nil {
		return ffmpegError(err, &v.stderr)
	}
	return cerr
}

func ffmpegError(err error, stderr *tailBuffer) error {
	if tail := strings.TrimSpace(stderr.String()); tail != "" {
		return fmt.Errorf("ffmpeg: %w: %s", err, tail)
	}
	return fmt.Errorf("ffmpeg: %w", err)
}

// tailBuffer keeps the last stderrTail bytes written to it.
type tailBuffer struct {
	b   []byte
	cut bool
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.b = append(t.b, p...)
	if over := len(t.b) - stderrTail; over > 0 {
		t.b, t.cut = append(t.b[:0], t.b[over:]...), true
	}
	return len(p), nil
}

// String is the kept tail, from its first whole line once cut.
func (t *tailBuffer) String() string {
	b := t.b
	if i := bytes.IndexByte(b, '\n'); t.cut && i >= 0 {
		b = b[i+1:]
	}
	return string(b)
}
