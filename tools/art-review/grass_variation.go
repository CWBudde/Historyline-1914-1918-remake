package artreview

import (
	"errors"
	"image"
)

const grassDetailStrength = .65

func grassVariationPhase(variant int) connectionPoint {
	return connectionPoint{float64(37 * variant % 142), float64(23 * variant % 80)}
}

// renderGrassVariation preserves the canonical shared material through eight
// world pixels from every edge, including outside the polygon. Distinct source
// phases enter with a smoothstep over 8..18 world pixels. The 2x Lanczos support
// therefore cannot reach a changed pixel in the six-pixel boundary guard.
// Both canonical and varied phases use the selected material and same sampler,
// preserving fine, common feature scale. No tint variation is used.
func renderGrassVariation(src *image.NRGBA, variant int) (map[int]*image.NRGBA, error) {
	return renderGrassDetailVariation(src, src, variant)
}

// Authored detail materials enter only the interior; every edge retains the
// exact canonical source. Equal image dimensions preserve material feature scale.
func renderGrassDetailVariation(src, detail *image.NRGBA, variant int) (map[int]*image.NRGBA, error) {
	if variant < 0 || variant >= 16 || src == nil || detail == nil || src.Bounds().Dx() < 384 || src.Bounds().Dy() < 320 || detail.Bounds().Size() != src.Bounds().Size() {
		return nil, errors.New("grass variation needs index 0..15 and equal-size sources at least 384x320")
	}
	g := sharedConnectionGeometry()
	phase := grassVariationPhase(variant)
	raw4 := renderGrassBaseRaw(src)
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			weight := connectionSmoothStep((mixedEdgeDistance(g, p) - 8) / 10)
			if weight == 0 {
				continue
			}
			meadow := sampleGrassMaterial(src, p.add(phase))
			paint := connectionBlend(meadow, sampleGrassMaterial(detail, p.add(phase)), grassDetailStrength)
			raw4.SetNRGBA(x, y, connectionBlend(raw4.NRGBAAt(x, y), paint, weight))
		}
	}
	return grassTilesFromRaw(raw4)
}

// grassSeasonDetailMaterial projects a seasonal edit of a detail painting
// onto its seasonal base's grid as summer does (Lanczos3) and shifts each
// channel's mean to the base's, so the edit adds its details without moving
// the meadow's tone off the seasonal base the tiles blend it into.
func grassSeasonDetailMaterial(base, painting *image.NRGBA) (*image.NRGBA, error) {
	material, err := resizeLanczos(painting, base.Bounds().Dx(), base.Bounds().Dy())
	if err != nil {
		return nil, err
	}
	mean := func(m *image.NRGBA) (s [3]float64) {
		for i := 0; i < len(m.Pix); i += 4 {
			for c := range 3 {
				s[c] += float64(m.Pix[i+c])
			}
		}
		for c := range 3 {
			s[c] /= float64(len(m.Pix) / 4)
		}
		return s
	}
	have, want := mean(material), mean(base)
	for i := 0; i < len(material.Pix); i += 4 {
		for c := range 3 {
			v := float64(material.Pix[i+c]) + want[c] - have[c]
			material.Pix[i+c] = uint8(max(0, min(255, v+.5)))
		}
	}
	return material, nil
}

func grassTilesFromRaw(raw4 *image.NRGBA) (map[int]*image.NRGBA, error) {
	g := sharedConnectionGeometry()
	raw2, err := resizeLanczos(raw4, 208, 176)
	if err != nil {
		return nil, err
	}
	result := map[int]*image.NRGBA{}
	for scale, raw := range map[int]*image.NRGBA{2: raw2, 4: raw4} {
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		for y := range 80 * scale {
			for x := range 96 * scale {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				if connectionInside(g, p) {
					im.SetNRGBA(x, y, raw.NRGBAAt(x+4*scale, y+4*scale))
				}
			}
		}
		result[scale] = im
	}
	return result, nil
}
