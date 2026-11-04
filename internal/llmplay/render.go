package llmplay

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

// RenderOptions controls RenderReplay.
type RenderOptions struct {
	Every    int         // render every Nth resolution (1 = all)
	Density  int         // maprender density (2 or 4)
	Div      int         // integer downscale of the rendered frame (8 ≈ 1120px wide on RHEIN)
	Backdrop image.Image // painted level backdrop instead of terrain tiles, nil for tiles
}

// RenderReplay replays a recorded game (verifying every state hash as
// VerifyReplay does) and writes frames/NNNN.png for the opening
// position and every Nth resolution — the raw material for a GIF or a
// video (see hlllplay render). It returns the number of frames.
func RenderReplay(dir string, r *maprender.Renderer, o RenderOptions) (int, error) {
	if o.Every < 1 {
		o.Every = 1
	}
	if o.Density == 0 {
		o.Density = 2
	}
	if o.Div < 1 {
		o.Div = 1
	}
	rp, err := OpenReplay(dir)
	if err != nil {
		return 0, err
	}
	s, lines := rp.Session, rp.Lines
	framesDir := filepath.Join(dir, "frames")
	if err := os.MkdirAll(framesDir, 0o755); err != nil {
		return 0, err
	}
	write := func(n int) (err error) {
		var img *image.NRGBA
		if o.Backdrop != nil {
			img, err = RenderFrameBackdrop(r, s, o.Density, o.Backdrop)
		} else {
			img, err = RenderFrame(r, s, o.Density)
		}
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(framesDir, fmt.Sprintf("%04d.png", n)))
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, f.Close()) }()
		return png.Encode(f, Downscale(img, o.Div))
	}
	n := 0
	if err := write(0); err != nil {
		return 0, fmt.Errorf("frame 0: %w", err)
	}
	n++
	for _, l := range lines {
		if _, err := s.Replay(l.Commands); err != nil {
			return n, fmt.Errorf("resolution %d: %w", l.Resolution, err)
		}
		if err := CheckResolution(s, l); err != nil {
			return n, err
		}
		if l.Resolution%o.Every == 0 || l.Resolution == len(lines) {
			if err := write(l.Resolution); err != nil {
				return n, fmt.Errorf("frame %d: %w", l.Resolution, err)
			}
			n++
		}
	}
	return n, nil
}
