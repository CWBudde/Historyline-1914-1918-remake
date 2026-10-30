package llmplay

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

// Project turns the omniscient TrueView into a level maprender can draw:
// terrain from the level plus sapper-reworked hexes, the unit on each
// hex, and each building entrance's current owner. Strength/experience
// are not representable in level.Hex; RenderFrame adds HP pips itself.
// Recording uses Project; prompts never do (they stay fair).
func Project(s *dosplay.Session) *level.Level {
	out := &level.Level{MapName: s.Level.MapName, MapNo: s.Level.MapNo, Player: s.Level.Player, Extra: s.Level.Extra}
	out.Lines = make([][]level.Hex, len(s.Level.Lines))
	for l, line := range s.Level.Lines {
		out.Lines[l] = make([]level.Hex, len(line))
		for c, h := range line {
			h.Unit = ""
			if h.Content != nil {
				ct := *h.Content
				h.Content = &ct
			}
			h.Infra.IDs = slices.Clone(h.Infra.IDs)
			out.Lines[l][c] = h
		}
	}
	for c, h := range s.Reworked() {
		u := out.Lines[c.Line][c.Col].Unit
		out.Lines[c.Line][c.Col] = h
		out.Lines[c.Line][c.Col].Unit = u
	}
	tv := s.TrueView()
	for at, u := range tv.Units {
		if at.Line < len(out.Lines) && at.Col < len(out.Lines[at.Line]) && u.ID != "" {
			out.Lines[at.Line][at.Col].Unit = u.ID
		}
	}
	for at, bv := range tv.Buildings {
		if at.Line >= len(out.Lines) || at.Col >= len(out.Lines[at.Line]) {
			continue
		}
		h := &out.Lines[at.Line][at.Col]
		if h.Content == nil {
			h.Content = &level.Content{}
		}
		h.Content.Owner = int(bv.Owner)
		var held []string
		for _, u := range bv.Units {
			if u.ID != "" {
				held = append(held, u.ID)
			}
		}
		if held != nil {
			h.Content.Units = held
		}
	}
	return out
}

// hexTopLeft is the logical-pixel origin of a hex in maprender's layout
// (Size: (Width-1)*142+96 by (Lines-1)*40+80). Neighbouring lines sit
// 40px apart, columns 142px, odd lines half a hex right.
func hexTopLeft(c hexgrid.Coord) (int, int) {
	x := c.Col * 142
	if c.Line%2 == 1 {
		x += 71
	}
	return x, c.Line * 40
}

// pip draws a filled bar of size/full over the unit at c.
func pip(img *image.NRGBA, c hexgrid.Coord, size, full, density int, col color.Color) {
	x, y := hexTopLeft(c)
	w := 60 * density
	h := 4 * density
	x0 := x*density + 8*density
	y0 := y*density + 4*density
	fill := w
	if full > 0 {
		fill = w * size / full
	}
	draw.Draw(img, image.Rect(x0, y0, x0+w, y0+h), &image.Uniform{color.NRGBA{0, 0, 0, 200}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(x0, y0, x0+fill, y0+h), &image.Uniform{col}, image.Point{}, draw.Src)
}

// RenderFrame draws the current true state at density (2 or 4) with a
// strength pip per unit: green for side 0, red for side 1, grey neutral.
func RenderFrame(r *maprender.Renderer, s *dosplay.Session, density int) (*image.NRGBA, error) {
	m, err := r.Render(Project(s), density, true)
	if err != nil {
		return nil, err
	}
	cols := [3]color.Color{
		color.NRGBA{40, 200, 60, 255}, color.NRGBA{230, 50, 40, 255}, color.NRGBA{170, 170, 170, 255},
	}
	for at, u := range s.TrueView().Units {
		pip(m.Image, at, u.Size, u.Full, density, cols[u.Side])
	}
	return m.Image, nil
}

// Downscale box-averages img by an integer factor (1 keeps it). A
// density-2 RHEIN frame is ~9000x3900px / 48MB PNG; factor 4 gives a
// watchable ~2250x960 recording at a few MB per frame.
func Downscale(img image.Image, factor int) *image.NRGBA {
	b := img.Bounds()
	if factor <= 1 {
		out := image.NewNRGBA(b)
		draw.Draw(out, b, img, b.Min, draw.Src)
		return out
	}
	w, h := b.Dx()/factor, b.Dy()/factor
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	n := uint32(factor * factor)
	for y := range h {
		for x := range w {
			var r, g, bl, a uint32
			for dy := range factor {
				for dx := range factor {
					pr, pg, pb, pa := img.At(b.Min.X+x*factor+dx, b.Min.Y+y*factor+dy).RGBA()
					r, g, bl, a = r+pr, g+pg, bl+pb, a+pa
				}
			}
			out.SetNRGBA(x, y, color.NRGBA{uint8(r / n >> 8), uint8(g / n >> 8), uint8(bl / n >> 8), uint8(a / n >> 8)})
		}
	}
	return out
}

// FrameRecorder returns a Game.Frames hook writing frames/NNNN.png into
// the game dir, downscaled by div. Renderer construction (assets tree)
// is the caller's.
func FrameRecorder(dir string, r *maprender.Renderer, density, div int) func(*dosplay.Session, int) (string, error) {
	return func(s *dosplay.Session, resolution int) (string, error) {
		full, err := RenderFrame(r, s, density)
		if err != nil {
			return "", err
		}
		img := Downscale(full, div)
		name := filepath.Join("frames", fmt.Sprintf("%04d.png", resolution))
		if err := os.MkdirAll(filepath.Join(dir, "frames"), 0o755); err != nil {
			return "", err
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			return "", err
		}
		defer func() { _ = f.Close() }()
		if err := png.Encode(f, img); err != nil {
			return "", err
		}
		return name, nil
	}
}

// WriteGIF stitches the frame PNGs of a game dir into game.gif at
// delayCS centiseconds per frame, downscaling by factor (1 = full size).
// Stdlib only: the repo has no ffmpeg dependency.
func WriteGIF(dir string, delayCS, factor int) (int, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "frames", "*.png"))
	if err != nil || len(paths) == 0 {
		return 0, fmt.Errorf("llmplay: no frames in %s", dir)
	}
	slices.Sort(paths)
	g := &gif.GIF{}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return 0, err
		}
		img, err := png.Decode(f)
		_ = f.Close()
		if err != nil {
			return 0, err
		}
		g.Image = append(g.Image, paletted(img, factor))
		g.Delay = append(g.Delay, delayCS)
	}
	out, err := os.Create(filepath.Join(dir, "game.gif"))
	if err != nil {
		return 0, err
	}
	defer func() { _ = out.Close() }()
	if err := gif.EncodeAll(out, g); err != nil {
		return 0, err
	}
	return len(g.Image), nil
}

func paletted(src image.Image, factor int) *image.Paletted {
	b := src.Bounds()
	if factor < 1 {
		factor = 1
	}
	w, h := b.Dx()/factor, b.Dy()/factor
	dst := image.NewPaletted(image.Rect(0, 0, w, h), palette.Plan9)
	for y := range h {
		for x := range w {
			dst.Set(x, y, src.At(b.Min.X+x*factor, b.Min.Y+y*factor))
		}
	}
	return dst
}
