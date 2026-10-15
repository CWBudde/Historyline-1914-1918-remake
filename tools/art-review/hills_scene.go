package main

import (
	"errors"
	"fmt"
	"sort"
)

type hillsCell struct {
	groundCell
	Phase connectionPoint `json:"phase"`
}
type hillsScene struct {
	ID               string          `json:"id"`
	CenterLineParity int             `json:"center_line_parity"`
	CenterOrigin     connectionPoint `json:"center_origin"`
	CanvasOrigin     connectionPoint `json:"canvas_origin"`
	Sites            []groundSite    `json:"sites"`
	Cells            []hillsCell     `json:"cells"`
	Outputs          []packOutput    `json:"outputs"`
}

func (s hillsScene) ground() groundScene {
	out := groundScene{ID: s.ID, CenterLineParity: s.CenterLineParity, CenterOrigin: s.CenterOrigin, CanvasOrigin: s.CanvasOrigin, Sites: s.Sites, Outputs: s.Outputs}
	for _, c := range s.Cells {
		out.Cells = append(out.Cells, c.groundCell)
	}
	return out
}
func hillsSceneFor(scenario string, parity int) (hillsScene, error) {
	if parity < 0 || parity > 1 {
		return hillsScene{}, errors.New("invalid hills fixture parity")
	}
	center := connectionPoint{float64(71 * parity), float64(40 * parity)}
	scene := hillsScene{ID: fmt.Sprintf("summer-%s-parity%d", scenario, parity), CenterLineParity: parity, CenterOrigin: center, CanvasOrigin: center}
	world := map[connectionPoint]string{center: "grass"}
	front := []connectionPoint{center}
	deltas := sharedConnectionGeometry().Neighbors
	for range 2 {
		var next []connectionPoint
		for _, p := range front {
			for _, d := range deltas {
				q := p.add(d)
				if _, ok := world[q]; !ok {
					world[q] = "grass"
					next = append(next, q)
				}
			}
		}
		front = next
	}
	for p := range world {
		hill := false
		switch scenario {
		case "isolated":
			hill = p == center
		case "homogeneous":
			hill = true
		case "clearing":
			hill = p != center
		case "frontier":
			hill = p[0] >= center[0]
		default:
			return scene, errors.New("unknown hills fixture scenario")
		}
		if hill {
			world[p] = "hills"
		}
	}
	for p, id := range world {
		scene.Sites = append(scene.Sites, groundSite{Origin: p, RuntimeID: id})
	}
	sort.Slice(scene.Sites, func(i, j int) bool {
		a, b := scene.Sites[i].Origin, scene.Sites[j].Origin
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		return a[0] < b[0]
	})
	visible := append([]connectionPoint{center}, make([]connectionPoint, 6)...)
	for i, d := range deltas {
		visible[i+1] = center.add(d)
	}
	for i, p := range visible {
		c := hillsCell{groundCell: groundCell{Label: []string{"C", "NE", "SE", "S", "SW", "NW", "N"}[i], Origin: p, RuntimeID: world[p]}}
		for n, d := range deltas {
			id, ok := world[p.add(d)]
			if !ok {
				return scene, errors.New("hills fixture missing actual neighbor site")
			}
			c.NeighborIDs[n] = id
		}
		context, err := newHillsContext(c.RuntimeID, c.NeighborIDs[:], p)
		if err != nil {
			return scene, err
		}
		c.ContextKey = context.key()
		c.Phase = context.effective().Phase
		scene.Cells = append(scene.Cells, c)
		scene.CanvasOrigin[0] = min(scene.CanvasOrigin[0], p[0])
		scene.CanvasOrigin[1] = min(scene.CanvasOrigin[1], p[1])
	}
	return scene, nil
}
