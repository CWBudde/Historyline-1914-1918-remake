package main

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

const pulseLevel = "levels/level1_cp.json"

// renderedMap is the metadata printed beside a map in the art book.
type renderedMap struct {
	Name    string `json:"name"`
	Level   string `json:"level"`
	Density int    `json:"density"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Hexes   int    `json:"hexes"`
	Units   int    `json:"units"`
	Gaps    int    `json:"gaps"`
	Issues  int    `json:"issues"`
	Detail  []int  `json:"detail,omitempty"` // logical x, y, width, height
}

// pulse renders the PULSE level with the runtime art at 4x into out/pulse-ground.png
// (terrain and infrastructure) and out/pulse-units.png (with the starting units).
func pulse(root, out string) error {
	return renderLevel(root, out, pulseLevel, "pulse", false)
}

// renderLevel uses the same runtime compositor for PULSE and other shipped maps:
// a level file, or a scenario setup "levels/monthly/1914-08-pulse.json:cp".
// The optional detail is a centred crop of the ground image, without resampling.
func renderLevel(root, out, levelPath, prefix string, detail bool) error {
	var lvl *level.Level
	var err error
	if file, mode, ok := strings.Cut(levelPath, ":"); ok {
		lvl, err = scenario.LoadLevel(filepath.Join(root, filepath.FromSlash(file)), scenario.Mode(mode))
	} else {
		lvl, err = level.Load(filepath.Join(root, filepath.FromSlash(levelPath)))
	}
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
	info := renderedMap{Name: lvl.MapName, Level: levelPath, Density: density, Width: w, Height: h, Hexes: lvl.Grid().Size()}
	for _, units := range []bool{false, true} {
		m, err := r.Render(lvl, density, units)
		if err != nil {
			return err
		}
		name := prefix + "-ground.png"
		if units {
			name, info.Units = prefix+"-units.png", m.Units
		}
		info.Gaps, info.Issues = len(m.Gaps), len(m.Issues)
		if err := writePNG(filepath.Join(out, name), m.Image); err != nil {
			return err
		}
		if detail && !units {
			cw, ch := min(600, w), min(600, h)
			x, y := (w-cw)/2, (h-ch)/2
			info.Detail = []int{x, y, cw, ch}
			crop := image.Rect(x*density, y*density, (x+cw)*density, (y+ch)*density)
			if err := writePNG(filepath.Join(out, prefix+"-detail.png"), m.Image.SubImage(crop)); err != nil {
				return err
			}
		}
	}
	raw, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, prefix+".json"), append(raw, '\n'), 0o644)
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
