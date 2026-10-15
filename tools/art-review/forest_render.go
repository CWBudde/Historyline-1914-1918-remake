package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"
	"sort"
	"strings"
)

// Forest contexts are bounded to these four existing runtime IDs, with ordered
// NE,SE,S,SW,NW,N neighbors. Keys identify offline profiles, not runtime IDs.
type forestContext struct {
	RuntimeID   string
	NeighborIDs [6]string
}

func forestSeason(id string) (season, base string) {
	switch id {
	case "forest", "grass":
		return "summer", "grass"
	case "forest_w", "grass_w":
		return "winter", "grass_w"
	default:
		return "", ""
	}
}
func newForestContext(center string, neighbors []string) (forestContext, error) {
	c := forestContext{RuntimeID: center}
	season, _ := forestSeason(center)
	if season == "" || len(neighbors) != 6 {
		return c, errors.New("forest context requires an existing seasonal forest/base ID and exactly six NE,SE,S,SW,NW,N neighbors")
	}
	for i, id := range neighbors {
		s, _ := forestSeason(id)
		if s != season {
			return c, fmt.Errorf("forest neighbor %d has incompatible ID %q", i, id)
		}
		c.NeighborIDs[i] = id
	}
	return c, nil
}
func (c forestContext) effective() forestContext {
	_, base := forestSeason(c.RuntimeID)
	if c.RuntimeID == base {
		for i := range c.NeighborIDs {
			c.NeighborIDs[i] = base
		}
	}
	return c
}
func (c forestContext) descriptor() string {
	c = c.effective()
	season, _ := forestSeason(c.RuntimeID)
	var b strings.Builder
	fmt.Fprintf(&b, "forest-v1\nseason=%s\ncenter=%s\n", season, c.RuntimeID)
	for i, label := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
		fmt.Fprintf(&b, "%s=%s\n", label, c.NeighborIDs[i])
	}
	return b.String()
}
func (c forestContext) key() string {
	sum := sha256.Sum256([]byte(c.descriptor()))
	return hex.EncodeToString(sum[:])
}

// Literal outcome of bounded periodic rejection: uint64 LCG seed 0x19141918,
// state=state*6364136223846793005+1442695040888963407 (wrapping); take upper
// 32 bits modulo 384 then 320 for quarter-grid x,y. Accept owned centers at
// least ten world pixels from existing points and their six translations.
// The 36th accepted center occurs on attempt 1934 (budget 200000).
func forestTemplate() []connectionPoint {
	return []connectionPoint{
		{20.5, 71.5}, {53.25, 4.5}, {78.5, 18.5}, {29.25, .75}, {51, 47.5}, {67.25, 11.25},
		{64.25, 51.5}, {38.5, 12.5}, {37, 47.5}, {15.75, 45.75}, {31.25, 28.75}, {73.75, 29.5},
		{21.25, 17.25}, {50, 74}, {63.5, 62.25}, {41.25, 36.25}, {55.5, 28}, {20.5, 36},
		{46.75, 56.75}, {10.25, 25.5}, {50.75, 15.75}, {5.5, 40.75}, {84.5, 48.75}, {36.5, 63.5},
		{37, 73.5}, {73.75, 41}, {62.25, 77.75}, {26.25, 52.25}, {20.25, 60.75}, {43, 25.5},
		{41.5, 2.75}, {60.25, 40}, {74.5, 55}, {71.5, 69.25}, {30, 39.25}, {61.75, 19.75},
	}
}

type forestInstance struct {
	Center     connectionPoint `json:"center"`
	CellOrigin connectionPoint `json:"cell_origin"`
	PointIndex int             `json:"point_index"`
	Variant    int             `json:"variant"`
}

// Lanczos3 4x->2x has radius six 4x pixels = 1.5 world pixels
// per axis. Its diagonal support is included to preserve the FINAL quiet band.
func forestGateRadius() float64 { return math.Hypot(12, 12) + 4 + 1.5*math.Sqrt2 }

func forestInstances(c forestContext) []forestInstance {
	c = c.effective()
	_, base := forestSeason(c.RuntimeID)
	if c.RuntimeID == base {
		return nil
	}
	origins := [7]connectionPoint{}
	copy(origins[1:], sharedConnectionGeometry().Neighbors)
	ids := [7]string{c.RuntimeID}
	copy(ids[1:], c.NeighborIDs[:])
	var instances []forestInstance
	for site, origin := range origins {
		if ids[site] == base {
			continue
		}
		for i, point := range forestTemplate() {
			center := origin.add(point)
			omit := false
			for other, baseOrigin := range origins {
				if ids[other] == base && groundFilledDistance(center.sub(baseOrigin)) <= forestGateRadius() {
					omit = true
					break
				}
			}
			if !omit {
				instances = append(instances, forestInstance{center, origin, i, i % 4})
			}
		}
	}
	sort.Slice(instances, func(i, j int) bool {
		a, b := instances[i], instances[j]
		if a.Center[1] != b.Center[1] {
			return a.Center[1] < b.Center[1]
		}
		if a.Center[0] != b.Center[0] {
			return a.Center[0] < b.Center[0]
		}
		if a.PointIndex != b.PointIndex {
			return a.PointIndex < b.PointIndex
		}
		if a.CellOrigin[1] != b.CellOrigin[1] {
			return a.CellOrigin[1] < b.CellOrigin[1]
		}
		return a.CellOrigin[0] < b.CellOrigin[0]
	})
	return instances
}

type forestRenderer struct {
	grounds map[string]*image.NRGBA
	baseRaw map[string]*image.NRGBA
	crowns  map[string][4]*image.RGBA
}

func newForestRenderer(grounds, atlases map[string]*image.NRGBA) (*forestRenderer, error) {
	r := &forestRenderer{grounds: grounds, baseRaw: map[string]*image.NRGBA{}, crowns: map[string][4]*image.RGBA{}}
	for _, season := range []string{"summer", "winter"} {
		atlas := atlases[season]
		if atlas == nil || grounds[season] == nil {
			return nil, errors.New("forest renderer requires both seasonal grounds and atlases")
		}
		panelSize := atlas.Bounds().Dx() / 2
		var crowns [4]*image.RGBA
		for i := range 4 {
			panel := image.NewNRGBA(image.Rect(0, 0, panelSize, panelSize))
			draw.Draw(panel, panel.Bounds(), atlas, image.Pt(i%2*panelSize, i/2*panelSize), draw.Src)
			scaled, err := resizeLanczos(panel, 96, 96)
			if err != nil {
				return nil, err
			}
			crowns[i] = image.NewRGBA(scaled.Bounds())
			draw.Draw(crowns[i], crowns[i].Bounds(), scaled, image.Point{}, draw.Src)
		}
		r.crowns[season] = crowns
		raw := image.NewNRGBA(image.Rect(0, 0, 416, 352))
		for y := range 352 {
			for x := range 416 {
				raw.SetNRGBA(x, y, sampleConnectionMaterial(grounds[season], connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}))
			}
		}
		r.baseRaw[season] = raw
	}
	return r, nil
}
func (r *forestRenderer) raw(c forestContext) *image.NRGBA {
	season, _ := forestSeason(c.RuntimeID)
	im := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	copy(im.Pix, r.baseRaw[season].Pix)
	for _, instance := range forestInstances(c) {
		// Quarter-grid centers make every origin an exact integer at 4x.
		at := image.Pt(int((instance.Center[0]-12+4)*4), int((instance.Center[1]-12+4)*4))
		crown := r.crowns[season][instance.Variant]
		draw.Draw(im, crown.Bounds().Add(at), crown, image.Point{}, draw.Over)
	}
	return im
}
func (r *forestRenderer) tiles(c forestContext) (map[int]*image.NRGBA, error) {
	season, base := forestSeason(c.RuntimeID)
	if c.RuntimeID == base {
		return exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": r.grounds[season]}}, nil, nil)
	}
	raw4 := r.raw(c)
	raw2, err := resizeLanczos(raw4, 208, 176)
	if err != nil {
		return nil, err
	}
	tiles := map[int]*image.NRGBA{}
	for scale, raw := range map[int]*image.NRGBA{2: raw2, 4: raw4} {
		im := image.NewNRGBA(image.Rect(0, 0, 96*scale, 80*scale))
		for y := range 80 * scale {
			for x := range 96 * scale {
				if connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}) {
					im.SetNRGBA(x, y, raw.NRGBAAt(x+4*scale, y+4*scale))
				}
			}
		}
		tiles[scale] = im
	}
	return tiles, nil
}
