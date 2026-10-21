package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"sort"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Reuse proofs keep the same central tile in six different constellations.
// Only the existing land-vertex profiles are candidates; a shared world corner
// must have one occupancy, including the three-cell contacts around the ring.
func shoreReuseScene(seed int) (shoreScene, error) {
	g := sharedConnectionGeometry()
	origin := connectionPoint{}
	if seed%2 != 0 {
		origin = connectionPoint{71, 40}
	}
	positions := []connectionPoint{origin}
	for _, n := range g.Neighbors {
		positions = append(positions, origin.add(n))
	}
	candidates := shoreVariants()
	orders := make([][]shoreVariant, 7)
	state := uint32(seed + 1)
	for i := range orders {
		orders[i] = slices.Clone(candidates)
		for j := len(candidates) - 1; j > 0; j-- {
			state = state*1664525 + 1013904223
			k := int(state % uint32(j+1))
			orders[i][j], orders[i][k] = orders[i][k], orders[i][j]
		}
	}
	orders[0] = []shoreVariant{{"lakeside23", []int{2, 3, 4, 5}}}
	fixedCorners := map[connectionPoint]bool{}
	chosen := make([]shoreVariant, 7)
	var search func(int) bool
	search = func(index int) bool {
		if index == len(positions) {
			return true
		}
		for _, v := range orders[index] {
			var added []connectionPoint
			valid := true
			for i, p := range g.Polygon {
				at := positions[index].add(p)
				land := slices.Contains(v.LandVertices, i)
				if previous, ok := fixedCorners[at]; ok {
					if previous != land {
						valid = false
						break
					}
				} else {
					fixedCorners[at] = land
					added = append(added, at)
				}
			}
			chosen[index] = v
			if valid && search(index+1) {
				return true
			}
			for _, p := range added {
				delete(fixedCorners, p)
			}
		}
		return false
	}
	if !search(0) {
		return shoreScene{}, errors.New("no compatible shore reuse arrangement")
	}
	s := shoreScene{CenterID: "lakeside23", CenterLineParity: seed % 2, CanvasOrigin: connectionPoint{1e9, 1e9}}
	for i, v := range chosen {
		s.Cells = append(s.Cells, shoreCell{positions[i], v.ID, v.LandVertices})
		s.CanvasOrigin[0] = min(s.CanvasOrigin[0], positions[i][0])
		s.CanvasOrigin[1] = min(s.CanvasOrigin[1], positions[i][1])
	}
	sort.Slice(s.Cells, func(i, j int) bool {
		a, b := s.Cells[i].Origin, s.Cells[j].Origin
		return a[1] < b[1] || a[1] == b[1] && a[0] < b[0]
	})
	return s, nil
}

func (b *shoreBundle) addReuseProof() error {
	seen := map[string]bool{}
	var pool []shoreScene
	for seed := range 256 {
		s, err := shoreReuseScene(seed)
		if err != nil {
			return err
		}
		var key strings.Builder
		for _, cell := range s.Cells {
			key.WriteString(cell.RuntimeID + "/")
		}
		if seen[key.String()] {
			continue
		}
		seen[key.String()] = true
		pool = append(pool, s)
	}
	for len(b.report.ReuseScenes) < 6 && len(pool) > 0 {
		// Prefer the arrangement that changes the most cells relative to its
		// closest selected example, rather than six nearly identical solutions.
		best, bestDistance := 0, -1
		for i, candidate := range pool {
			distance := 7
			for _, earlier := range b.report.ReuseScenes {
				changed := 0
				for j, cell := range candidate.Cells {
					if cell.RuntimeID != earlier.Cells[j].RuntimeID {
						changed++
					}
				}
				distance = min(distance, changed)
			}
			if distance > bestDistance {
				best, bestDistance = i, distance
			}
		}
		if bestDistance < 2 {
			break
		}
		s := pool[best]
		pool = slices.Delete(pool, best, best+1)
		s.ID = fmt.Sprintf("shore-reuse-%02d", len(b.report.ReuseScenes)+1)
		for _, density := range []int{2, 4} {
			output, err := b.addPNG(fmt.Sprintf("%s-%dx.png", s.ID, density), renderShoreScene(b.pixels, s, density))
			if err != nil {
				return err
			}
			if density == 2 {
				s.Output = output
			} else {
				s.Output4x = output
			}
		}
		b.report.ReuseScenes = append(b.report.ReuseScenes, s)
	}
	if len(b.report.ReuseScenes) != 6 {
		return errors.New("shore reuse proof needs six distinct compatible arrangements")
	}
	for _, density := range []int{2, 4} {
		out, err := b.addPNG(fmt.Sprintf("shore-catalogue-%dx.png", density), b.renderReuseCatalogue(density))
		if err != nil {
			return err
		}
		b.report.ReuseCatalogues = append(b.report.ReuseCatalogues, out)
	}
	return nil
}

// Every tile and patch is copied at its exported density without resizing.
// Text is merely a presentation label; it never modifies the terrain artwork.
func (b *shoreBundle) renderReuseCatalogue(density int) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 786*density, 932*density))
	draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{27, 34, 32, 255}), image.Point{}, draw.Src)
	label := func(x, y int, text string) { shoreProofLabel(im, density, x, y, text) }
	label(12, 8, "Reusable shore tiles - native exports")
	for i, v := range shoreRoster() {
		x, y := 12+i%3*254, 32+i/3*104
		src := b.pixels[fmt.Sprintf("tiles/%s-%dx.png", v.ID, density)]
		at := image.Pt((x+71)*density, y*density)
		draw.Draw(im, src.Bounds().Add(at), src, src.Bounds().Min, draw.Over)
		label(x+71, y+84, v.ID)
	}
	label(12, 352, "Same lakeside23 tile, six compatible arrangements")
	for i, s := range b.report.ReuseScenes {
		x, y := 12+i%3*254, 380+i/3*268
		src := b.pixels[fmt.Sprintf("%s-%dx.png", s.ID, density)]
		at := image.Pt(x*density, y*density)
		draw.Draw(im, src.Bounds().Add(at), src, src.Bounds().Min, draw.Over)
		label(x, y+244, fmt.Sprintf("Arrangement %d", i+1))
	}
	return im
}

func shoreProofLabel(im *image.NRGBA, density, x, y int, text string) {
	mask := image.NewAlpha(image.Rect(0, 0, len(text)*7, 13))
	d := font.Drawer{Dst: mask, Src: image.White, Face: basicfont.Face7x13, Dot: fixed.P(0, 11)}
	d.DrawString(text)
	ink := image.NewUniform(color.NRGBA{226, 230, 216, 255})
	for py := range mask.Bounds().Dy() {
		for px := range mask.Bounds().Dx() {
			if mask.AlphaAt(px, py).A == 0 {
				continue
			}
			r := image.Rect((x+px)*density, (y+py)*density, (x+px+1)*density, (y+py+1)*density)
			draw.Draw(im, r, ink, image.Point{}, draw.Src)
		}
	}
}
