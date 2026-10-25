package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

//go:embed pulse.html
var pulsePage string

type pulseStudy struct {
	Generated   string             `json:"generated"`
	LevelHash   string             `json:"level_sha256"`
	Width       int                `json:"width"`
	Height      int                `json:"height"`
	Hexes       int                `json:"hexes"`
	Units       int                `json:"units"`
	Gaps        []mapcompose.Gap   `json:"gaps"`
	Issues      []mapcompose.Issue `json:"issues"`
	Maps        map[string]picture `json:"maps"`
	GroundMaps  map[string]picture `json:"ground_maps"`
	CaptureNote string             `json:"capture_note"`
	Captures    []pulseCapture     `json:"captures"`
}
type pulseCapture struct {
	File    string  `json:"file"`
	View    string  `json:"view"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Zoom    float64 `json:"zoom"`
	UI      int     `json:"ui_scale"`
	Density int     `json:"asset_density"`
	Reveal  bool    `json:"reveal"`
	Picture picture `json:"picture"`
}

func (c *collector) pulseImage(im image.Image, source, method string) (picture, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return picture{}, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
	p := picture{URL: "media/" + hash + ".png", Source: source, SHA: hash, Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), Method: method}
	return p, c.out.WriteFile(p.URL, b.Bytes(), 0o644)
}

func (c *collector) pulse(captures string) error {
	raw, err := c.read("art/reference/pulse/level.json")
	if err != nil {
		return err
	}
	lvl, err := level.Parse(raw)
	if err != nil {
		return err
	}
	if lvl.MapName != "Pulse" {
		return fmt.Errorf("PULSE study requires the Pulse level")
	}
	grid := lvl.Grid()
	w, h := maprender.Size(grid)
	s := pulseStudy{Generated: time.Now().UTC().Format(time.RFC3339), LevelHash: fmt.Sprintf("%x", sha256.Sum256(raw)), Width: w, Height: h, Hexes: grid.Size(), Maps: map[string]picture{}, GroundMaps: map[string]picture{}}
	assets, err := fs.Sub(c.root.FS(), "assets")
	if err != nil {
		return err
	}
	r, err := maprender.New(assets)
	if err != nil {
		return err
	}
	// All fog is lifted for the offline art comparison.
	for _, density := range []int{2, 4} {
		ground, err := r.Render(lvl, density, false)
		if err != nil {
			return err
		}
		s.Gaps, s.Issues = ground.Gaps, ground.Issues
		if err := c.pulseDensity(s.GroundMaps, ground.Image, density, "Shared compositor terrain and infrastructure; no units, owner bars or UI"); err != nil {
			return err
		}
		full, err := r.Render(lvl, density, true)
		if err != nil {
			return err
		}
		s.Units = full.Units
		if err := c.pulseDensity(s.Maps, full.Image, density, "Shared compositor terrain, infrastructure and authored units; owner bars/UI omitted"); err != nil {
			return err
		}
	}
	if err := c.pulseCaptures(&s, captures, lvl); err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err = c.out.WriteFile("pulse.json", b, 0o644); err != nil {
		return err
	}
	return c.out.WriteFile("pulse.html", []byte(strings.Replace(pulsePage, "/*STUDY*/null", string(b), 1)), 0o644)
}

func (c *collector) pulseDensity(images map[string]picture, im *image.NRGBA, density int, method string) error {
	p, err := c.pulseImage(im, "art/reference/pulse/level.json + assets/index.json", fmt.Sprintf("%s; authored %dx exports", method, density))
	if err != nil {
		return err
	}
	images[strconv.Itoa(density)] = p
	if density == 2 {
		reduced, err := resizeLanczos(im, im.Bounds().Dx()/2, im.Bounds().Dy()/2)
		if err != nil {
			return err
		}
		p, err = c.pulseImage(reduced, p.Source, method+"; full 2x composition reduced with premultiplied Lanczos3; no authored 1x pack")
		if err != nil {
			return err
		}
		images["1"] = p
	}
	return nil
}

func (c *collector) pulseCaptures(s *pulseStudy, dir string, lvl *level.Level) error {
	s.CaptureNote = "No game captures available. Run the documented historyline -capture command, then regenerate."
	if dir == "" {
		return nil
	}
	if !local(dir) || !strings.HasPrefix(dir, "build/") {
		return fmt.Errorf("captures must be under build/")
	}
	_, err := c.root.Stat(path.Join(dir, "captures.json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	var manifest struct {
		Level     string         `json:"level"`
		LevelHash string         `json:"canonical_level_sha256"`
		Generated string         `json:"generated"`
		Frames    []pulseCapture `json:"frames"`
	}
	if err := c.json(path.Join(dir, "captures.json"), &manifest); err != nil {
		return err
	}
	if manifest.Level != "Pulse" || len(manifest.Frames) != 9 {
		return fmt.Errorf("capture manifest must contain nine PULSE frames")
	}
	levelBytes, err := lvl.Marshal()
	if err != nil {
		return err
	}
	if manifest.LevelHash != fmt.Sprintf("%x", sha256.Sum256(levelBytes)) {
		return fmt.Errorf("captures use a different PULSE level; recapture before regenerating")
	}
	seen := map[string]bool{}
	for _, f := range manifest.Frames {
		key := fmt.Sprintf("%d/%s", f.Width, f.View)
		if seen[key] || map[int]int{1280: 720, 1920: 1080, 3840: 2160}[f.Width] != f.Height ||
			(f.View != "overview" && f.View != "detail" && f.View != "fog") || f.Reveal != (f.View != "fog") || f.Zoom <= 0 || f.UI <= 0 || (f.Density != 2 && f.Density != 4) {
			return fmt.Errorf("invalid/duplicate capture view %s", key)
		}
		seen[key] = true
		if !local(f.File) || strings.Contains(f.File, "/") {
			return fmt.Errorf("invalid capture filename")
		}
		p, e := c.snapshot(path.Join(dir, f.File), "", image.Rectangle{}, 0, 0)
		if e != nil {
			return e
		}
		if p.Width != f.Width || p.Height != f.Height {
			return fmt.Errorf("capture dimensions disagree")
		}
		f.Picture = p
		s.Captures = append(s.Captures, f)
	}
	s.CaptureNote = "Actual Go Draw framebuffers, captured " + manifest.Generated + ". Initial game state; overview/detail reveal the map for art review, fog uses the player's real starting vision. Current debug HUD is part of the capture; screen designs remain separate fixtures."
	return nil
}
