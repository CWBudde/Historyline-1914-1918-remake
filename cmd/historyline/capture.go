package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/hajimehoshi/ebiten/v2"
)

// captureGame drives the real Draw method at explicit framebuffer sizes. It
// never advances gameplay or reads live keyboard/mouse state.
type captureGame struct {
	*game
	out     string
	index   int
	err     error
	records []frameCapture
	frames  []reviewFrame
}

type frameCapture struct {
	File          string  `json:"file"`
	View          string  `json:"view"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	Zoom          float64 `json:"zoom"`
	EffectiveZoom float64 `json:"effective_zoom"`
	UI            int     `json:"ui_scale"`
	Density       int     `json:"asset_density"`
	Reveal        bool    `json:"reveal"`
}

type reviewFrame struct {
	width, height int
	view          string
	reveal        bool
	effectiveZoom float64 // zero keeps the existing overview/detail framing
}

func reviewFramePlan(look bool) []reviewFrame {
	var frames []reviewFrame
	if look {
		for _, size := range [][2]int{{1920, 1080}, {3840, 2160}} {
			for _, zoom := range []float64{1, 1.6} {
				for _, reveal := range []bool{true, false} {
					view := "fog"
					if reveal {
						view = "revealed"
					}
					frames = append(frames, reviewFrame{size[0], size[1], "look-" + fmt.Sprintf("%gx-", zoom) + view, reveal, zoom})
				}
			}
		}
		return frames
	}
	for _, size := range [][2]int{{1280, 720}, {1920, 1080}, {3840, 2160}} {
		for _, view := range []string{"overview", "detail", "fog"} {
			frames = append(frames, reviewFrame{size[0], size[1], view, view != "fog", 0})
		}
	}
	return frames
}

func (f reviewFrame) filename() string {
	return fmt.Sprintf("pulse-%s-%dx%d.png", f.view, f.width, f.height)
}

func (f reviewFrame) cameraZoom(world camera.Size, ui int) float64 {
	if f.effectiveZoom > 0 {
		return f.effectiveZoom / float64(ui)
	}
	if f.view == "overview" {
		return captureZoom(world, f.width, f.height, ui)
	}
	return float64(f.height) / 720 / float64(ui)
}

func captureZoom(world camera.Size, w, h, ui int) float64 {
	return math.Min(float64(w)/world.W, float64(h)/world.H) * 5 / 6 / float64(ui)
}

func (c *captureGame) Update() error {
	if c.err != nil {
		return c.err
	}
	if c.index == len(c.frames) {
		return ebiten.Termination
	}
	return nil
}

func (*captureGame) Layout(int, int) (int, int) { return 1280, 720 }

func (c *captureGame) Draw(screen *ebiten.Image) {
	if c.err != nil || c.index == len(c.frames) {
		return
	}
	frame := c.frames[c.index]
	w, h := frame.width, frame.height
	g := c.game
	g.applyLayout(w, h, 1)
	g.reveal = frame.reveal
	g.cam.Zoom = frame.cameraZoom(g.cam.World, g.view.UI)
	cx, cy := g.cam.World.W/2, g.cam.World.H/2
	if frame.view != "overview" {
		// Same HQ/unit neighbourhood at each resolution and effective zoom.
		cx, cy = g.cam.World.W*.42, g.cam.World.H*.22
	}
	g.cam.CenterOn(cx, cy)
	target := ebiten.NewImage(w, h)
	defer target.Deallocate()
	g.Draw(target)
	if g.artErr != "" || g.terrainErr != "" {
		c.err = fmt.Errorf("capture rendering: %s %s", g.artErr, g.terrainErr)
		return
	}
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	target.ReadPixels(im.Pix)
	name := frame.filename()
	file, err := os.Create(filepath.Join(c.out, name))
	if err != nil {
		c.err = err
		return
	}
	encodeErr := png.Encode(file, im)
	closeErr := file.Close()
	if encodeErr != nil {
		c.err = encodeErr
		return
	}
	if closeErr != nil {
		c.err = closeErr
		return
	}
	c.records = append(c.records, frameCapture{
		File: name, View: frame.view, Width: w, Height: h, Zoom: g.cam.Zoom,
		EffectiveZoom: g.cam.Zoom * float64(g.view.UI), UI: g.view.UI,
		Density: g.drawnDensity(), Reveal: g.reveal,
	})
	c.index++
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(1280/float64(w), 720/float64(h))
	screen.DrawImage(target, &op)
}

func writeReviewFrames(g *game, out string, look bool) error {
	clean := filepath.ToSlash(filepath.Clean(out))
	if !strings.HasPrefix(clean, "build/") || strings.Contains(clean, "..") {
		return fmt.Errorf("capture output must be under ignored build/")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	c := &captureGame{game: g, out: out, frames: reviewFramePlan(look)}
	if err := ebiten.RunGame(c); err != nil {
		return err
	}
	if c.err != nil {
		return c.err
	}
	if len(c.records) != len(c.frames) {
		return fmt.Errorf("capture interrupted after %d of %d frames", len(c.records), len(c.frames))
	}
	levelBytes, err := g.lvl.Marshal()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(struct {
		Level          string         `json:"level"`
		LevelHash      string         `json:"canonical_level_sha256"`
		Generated      string         `json:"generated"`
		TerrainGaps    int            `json:"terrain_gaps"`
		CompoundIssues int            `json:"compound_issues"`
		Frames         []frameCapture `json:"frames"`
	}{g.lvl.MapName, fmt.Sprintf("%x", sha256.Sum256(levelBytes)), time.Now().UTC().Format(time.RFC3339), len(g.terrain.scene.Gaps), len(g.terrain.scene.Issues), c.records}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "captures.json"), b, 0644)
}
