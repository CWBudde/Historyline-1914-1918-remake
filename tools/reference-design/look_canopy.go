package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

type canopyPlane struct {
	w, h  int
	alpha []float64
}

type lookCanopy struct {
	Available   bool              `json:"available"`
	Reason      string            `json:"reason,omitempty"`
	Definition  string            `json:"definition"`
	MaskFile    string            `json:"mask_file,omitempty"`
	MaskSHA256  string            `json:"mask_sha256,omitempty"`
	Unsupported []string          `json:"unsupported_cells"`
	Groups      []lookCanopyGroup `json:"groups"`
	Cells       []lookCanopyCell  `json:"cells"`
	Gaps        int               `json:"compositor_gaps"`
	Issues      int               `json:"compositor_issues"`
}

type lookCanopyGroup struct {
	Group    string    `json:"group"`
	Cells    int       `json:"cells"`
	Coverage lookValue `json:"coverage_fraction"`
}

type lookCanopyCell struct {
	Cell     string    `json:"cell"`
	Group    string    `json:"group"`
	Coverage lookValue `json:"coverage_fraction"`
}

// Alpha means survive fractional crown edges; opaque-only colour reduction
// would discard them. Numbers retain the full average, without 8-bit rounding.
func averageCanopy(im *image.Alpha, density int) (*canopyPlane, error) {
	if im == nil || density != 2 && density != 4 {
		return nil, fmt.Errorf("canopy needs a 2x or 4x alpha plane")
	}
	b := im.Bounds()
	if b.Empty() || b.Dx()%density != 0 || b.Dy()%density != 0 {
		return nil, fmt.Errorf("canopy has incomplete density blocks")
	}
	p := &canopyPlane{w: b.Dx() / density, h: b.Dy() / density}
	p.alpha = make([]float64, p.w*p.h)
	for y := range p.h {
		for x := range p.w {
			sum := 0
			for j := range density {
				for i := range density {
					sum += int(im.AlphaAt(b.Min.X+x*density+i, b.Min.Y+y*density+j).A)
				}
			}
			p.alpha[y*p.w+x] = float64(sum) / float64(255*density*density)
		}
	}
	return p, nil
}

func (p *project) canopyGroup(c cell) string {
	h, ok := p.hex(c)
	if !ok {
		return ""
	}
	if h.Terrain == "grass" || h.Terrain == "grass_w" {
		return "open_grass"
	}
	if h.Terrain != "forest" && h.Terrain != "forest_w" {
		return ""
	}
	for _, n := range hexgrid.Around(hexgrid.Coord{Line: c.Row - 1, Col: c.Col - 1}) {
		other, ok := p.hex(cell{n.Line + 1, n.Col + 1})
		if !ok || other.Terrain != h.Terrain {
			return "forest_frontier"
		}
	}
	return "forest_interior"
}

func (p *project) measureCanopy(plane *canopyPlane, unsupported []hexgrid.Coord) lookCanopy {
	result := lookCanopy{Definition: "visible contribution of crown-only runtime alpha after later standing occlusion; shadows and units excluded; exact area mean at 1x; open means clean grass cells; interior has six forest neighbours; this is coverage, not a woodland-perimeter measurement", Unsupported: []string{}, Groups: []lookCanopyGroup{}, Cells: []lookCanopyCell{}}
	w, h := maprender.Size(p.level.Grid())
	if plane == nil || plane.w != w || plane.h != h || len(plane.alpha) != w*h {
		result.Reason = "invalid_canopy_plane"
		return result
	}
	missing := map[hexgrid.Coord]bool{}
	for _, c := range unsupported {
		missing[c] = true
		result.Unsupported = append(result.Unsupported, cell{c.Line + 1, c.Col + 1}.String())
	}
	type aggregate struct {
		cells, samples int
		sum            float64
	}
	groups := map[string]*aggregate{}
	total := 0
	for _, c := range p.cells() {
		hex, _ := p.hex(c)
		group := p.canopyGroup(c)
		if p.excluded[c.Row] || group == "" || len(hex.Infra.IDs) > 0 || hex.Content != nil {
			continue
		}
		entry := lookCanopyCell{Cell: c.String(), Group: group}
		a := groups[group]
		if a == nil {
			a = &aggregate{}
			groups[group] = a
		}
		a.cells++
		if missing[hexgrid.Coord{Line: c.Row - 1, Col: c.Col - 1}] {
			entry.Coverage.Reason = "canopy_layer_unavailable"
			result.Cells = append(result.Cells, entry)
			continue
		}
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		sum, n := 0., 0
		for y := range hexgrid.HexHeight {
			for x := range hexgrid.HexWidth {
				px, py := ox+x, oy+y
				if px < 0 || py < 0 || px >= w || py >= h || !insideTile(x, y, 1) {
					continue
				}
				value := plane.alpha[py*w+px]
				if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
					result.Reason = "invalid_canopy_alpha"
					return result
				}
				sum += value
				n++
			}
		}
		if n > 0 {
			entry.Coverage = lookMeasured(sum/float64(n), n)
			a.sum += sum
			a.samples += n
			total += n
		} else {
			entry.Coverage.Reason = "insufficient_samples"
		}
		result.Cells = append(result.Cells, entry)
	}
	for _, group := range []string{"forest_interior", "forest_frontier", "open_grass"} {
		a := groups[group]
		if a == nil {
			continue
		}
		entry := lookCanopyGroup{Group: group, Cells: a.cells}
		if a.samples > 0 {
			entry.Coverage = lookMeasured(a.sum/float64(a.samples), a.samples)
		} else {
			entry.Coverage.Reason = "canopy_layer_unavailable"
		}
		result.Groups = append(result.Groups, entry)
	}
	result.Available = total > 0
	if !result.Available {
		result.Reason = "insufficient_samples"
	}
	return result
}

// The PNG is a quantized diagnostic; statistics use the unrounded plane.
func writeCanopyPlane(out, name string, p *canopyPlane) (string, error) {
	im := image.NewGray(image.Rect(0, 0, p.w, p.h))
	for i, a := range p.alpha {
		im.Pix[i] = uint8(math.Round(255 * a))
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return "", err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(out, name), b.Bytes(), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), nil
}
