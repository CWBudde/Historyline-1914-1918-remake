package artreview

import (
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
)

// A two-cell ship (the original's battleship, destroyer and submarine) stands
// on two hexes, its head's and its rear's. Its long export draws the six
// authored panels of the family's single-hex atlas on a two-hex canvas: each
// panel's hull is centred on the canvas centre, the midpoint between the two
// hex centres, and scaled so that it spans hull_logical px along its length.
// The runtime family is <family>-long.

// shipManifest names the family's single-hex unit manifest (atlas pin and
// owner palettes) and the hull length in logical px.
type shipManifest struct {
	SchemaVersion int     `json:"schema_version"`
	Unit          string  `json:"unit"`
	Hull          float64 `json:"hull_logical"`
}

// shipCanvas is the square long-ship canvas in logical px; its centre is the
// logical anchor.
const shipCanvas = 160

// shipFrame is one facing's registration: the hull's centre and length in
// source panel px, and the source-to-4x scale.
type shipFrame struct {
	Facing  string          `json:"facing"`
	Centre  connectionPoint `json:"source_hull_centre"`
	Length  float64         `json:"source_hull_length"`
	Scale4x float64         `json:"source_to_4x_scale"`
}

// shipInputs are a decoded ship manifest and its unit inputs.
type shipInputs struct {
	Manifest shipManifest
	Unit     unitInputs
}

func loadShipInputs(path string) (shipInputs, error) {
	var in shipInputs
	b, err := os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeConnectionJSON(b, &in.Manifest); err != nil {
		return in, err
	}
	m := in.Manifest
	if m.SchemaVersion != 1 || m.Unit == "" || m.Hull < 96 || m.Hull > shipCanvas-24 {
		return in, errors.New("ship manifest requires schema_version 1, its unit manifest and hull_logical in 96..136")
	}
	in.Unit, err = loadUnitInputs(filepath.Join(sourceBase(path), m.Unit))
	return in, err
}

// shipHull measures the hull of panel plane n: the pixels with alpha of at
// least 128, their principal axis, and the midpoint and length of their extent
// along it.
func shipHull(n *image.NRGBA) (connectionPoint, float64, error) {
	var sx, sy, sxx, syy, sxy, count float64
	b := n.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if n.NRGBAAt(x, y).A < 128 {
				continue
			}
			fx, fy := float64(x)+.5, float64(y)+.5
			sx, sy, sxx, syy, sxy, count = sx+fx, sy+fy, sxx+fx*fx, syy+fy*fy, sxy+fx*fy, count+1
		}
	}
	if count < 100 {
		return connectionPoint{}, 0, errors.New("ship panel has no hull")
	}
	mx, my := sx/count, sy/count
	cxx, cyy, cxy := sxx/count-mx*mx, syy/count-my*my, sxy/count-mx*my
	angle := .5 * math.Atan2(2*cxy, cxx-cyy)
	ax, ay := math.Cos(angle), math.Sin(angle)
	lo, hi := math.Inf(1), math.Inf(-1)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if n.NRGBAAt(x, y).A < 128 {
				continue
			}
			t := (float64(x)+.5-mx)*ax + (float64(y)+.5-my)*ay
			lo, hi = min(lo, t-.5), max(hi, t+.5)
		}
	}
	mid := (lo + hi) / 2
	return connectionPoint{mx + mid*ax, my + mid*ay}, hi - lo, nil
}

// shipSprites renders the long sprites of in, keyed by
// <family>-long/<palette>/<facing>-<state>-<d>x.png, and the frames.
func shipSprites(in shipInputs) (map[string]*image.NRGBA, []shipFrame, error) {
	unit := in.Unit
	cell := unit.Atlas.Image.Bounds().Dx() / 2
	size := shipCanvas * 4
	out := map[string]*image.NRGBA{}
	var frames []shipFrame
	for face, label := range unitFaces {
		rect := image.Rect(face%2*cell, face/2*cell, (face%2+1)*cell, (face/2+1)*cell)
		n, k := unitPlanes(unit.Atlas.Image.SubImage(rect).(*image.NRGBA))
		centre, length, err := shipHull(n)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}
		scale := in.Manifest.Hull * 4 / length
		offset := connectionPoint{float64(size)/2 - scale*centre[0], float64(size)/2 - scale*centre[1]}
		frames = append(frames, shipFrame{label, centre, length, scale})
		n4, err := unitAffineTo(n, size, scale, offset)
		if err != nil {
			return nil, nil, err
		}
		k4, err := unitAffineTo(k, size, scale, offset)
		if err != nil {
			return nil, nil, err
		}
		n2, err := resizeLanczos(n4, size/2, size/2)
		if err != nil {
			return nil, nil, err
		}
		k2, err := resizeLanczos(k4, size/2, size/2)
		if err != nil {
			return nil, nil, err
		}
		planes := map[int][2]*image.NRGBA{2: {n2, k2}, 4: {n4, k4}}
		for _, palette := range []string{"cp", "ec", "neutral"} {
			for _, state := range []string{"normal", "spent"} {
				for d, p := range planes {
					im := unitTint(p[0], p[1], unit.Manifest.Palettes[palette], state == "spent")
					if _, err := unitGuardAlpha(im, d, 8); err != nil {
						return nil, nil, fmt.Errorf("%s %s %s %dx: %w", label, palette, state, d, err)
					}
					out[fmt.Sprintf("%s-long/%s/%s-%s-%dx.png", unit.Manifest.Family, palette, label, state, d)] = im
				}
			}
		}
	}
	return out, frames, nil
}
