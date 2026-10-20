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
}

type frameCapture struct {
	File    string  `json:"file"`
	View    string  `json:"view"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Zoom    float64 `json:"zoom"`
	UI      int     `json:"ui_scale"`
	Density int     `json:"asset_density"`
	Reveal  bool    `json:"reveal"`
}

func captureZoom(world camera.Size, w, h, ui int) float64 {
	return math.Min(float64(w)/world.W, float64(h)/world.H) * 5 / 6 / float64(ui)
}

func (c *captureGame) Update() error {
	if c.err != nil {
		return c.err
	}
	if c.index == 9 {
		return ebiten.Termination
	}
	return nil
}

func (*captureGame) Layout(int, int) (int, int) { return 1280, 720 }

func (c *captureGame) Draw(screen *ebiten.Image) {
	if c.err != nil || c.index == 9 {
		return
	}
	sizes := [3][2]int{{1280, 720}, {1920, 1080}, {3840, 2160}}
	w, h := sizes[c.index/3][0], sizes[c.index/3][1]
	view := []string{"overview", "detail", "fog"}[c.index%3]
	g := c.game
	g.applyLayout(w, h, 1)
	g.reveal = view != "fog"
	g.cam.Zoom = captureZoom(g.cam.World, w, h, g.view.UI)
	cx, cy := g.cam.World.W/2, g.cam.World.H/2
	if view != "overview" {
		// Same HQ/unit neighbourhood at each resolution; effective zoom scales
		// with framebuffer size, independent of the HUD scale.
		g.cam.Zoom = float64(h) / 720 / float64(g.view.UI)
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
	name := fmt.Sprintf("pulse-%s-%dx%d.png", view, w, h)
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
	c.records = append(c.records, frameCapture{name, view, w, h, g.cam.Zoom, g.view.UI, g.drawnDensity(), g.reveal})
	c.index++
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(1280/float64(w), 720/float64(h))
	screen.DrawImage(target, &op)
}

func writeReviewFrames(g *game, out string) error {
	clean := filepath.ToSlash(filepath.Clean(out))
	if !strings.HasPrefix(clean, "build/") || strings.Contains(clean, "..") {
		return fmt.Errorf("capture output must be under ignored build/")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	c := &captureGame{game: g, out: out}
	if err := ebiten.RunGame(c); err != nil {
		return err
	}
	if c.err != nil {
		return c.err
	}
	if len(c.records) != 9 {
		return fmt.Errorf("capture interrupted after %d of 9 frames", len(c.records))
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
