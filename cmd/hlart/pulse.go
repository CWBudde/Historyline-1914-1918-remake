package main

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

const pulseLevel = "levels/level1_cp.json"

// pulseMap is out/pulse.json, the facts the art book prints beside the maps.
type pulseMap struct {
	Level   string `json:"level"`
	Density int    `json:"density"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Hexes   int    `json:"hexes"`
	Units   int    `json:"units"`
	Gaps    int    `json:"gaps"`
	Issues  int    `json:"issues"`
}

// pulse renders the PULSE level with the runtime art at 4x into out/pulse-ground.png
// (terrain and infrastructure) and out/pulse-units.png (with the starting units).
func pulse(root, out string) error {
	lvl, err := level.Load(filepath.Join(root, filepath.FromSlash(pulseLevel)))
	if err != nil {
		return err
	}
	r, err := maprender.New(os.DirFS(filepath.Join(root, "assets")))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	const density = 4
	w, h := maprender.Size(lvl.Grid())
	info := pulseMap{Level: pulseLevel, Density: density, Width: w, Height: h, Hexes: lvl.Grid().Size()}
	for _, units := range []bool{false, true} {
		m, err := r.Render(lvl, density, units)
		if err != nil {
			return err
		}
		name := "pulse-ground.png"
		if units {
			name, info.Units = "pulse-units.png", m.Units
		}
		info.Gaps, info.Issues = len(m.Gaps), len(m.Issues)
		if err := writePNG(filepath.Join(out, name), m.Image); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "pulse.json"), append(raw, '\n'), 0o644)
}

func writePNG(path string, img image.Image) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(w, img); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
