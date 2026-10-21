package mapcompose

import (
	"fmt"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Corner indices match the shoreline export's clockwise polygon, starting at
// its top-left vertex. These are the canonical masks, including lakeside23's
// four land corners; a stream mouth does not change their shared endpoints.
var shoreLandVertices = map[string][]int{
	"lakeside12":   {1, 2, 3},
	"lakeside16":   {0, 1, 2},
	"lakeside23":   {2, 3, 4, 5},
	"lakeside34":   {3, 4, 5},
	"lakeside56":   {0, 1, 5},
	"lakeside123":  {1, 2, 3, 4},
	"lakeside126":  {0, 1, 2, 3},
	"lakeside156":  {0, 1, 2, 5},
	"lakeside3456": {0, 1, 3, 4, 5},
}

// selectShoreOutlet joins a shore to one neighboring stream that faces it.
// Mouth exports preserve the two land corners of that edge and open only its
// middle to the stream. Missing art and multiple facing streams keep the base
// tile; choosing one of several mouths would leave the other streams blocked.
func selectShoreOutlet(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord, terrain string, has func(string) bool) string {
	land, shore := shoreLandVertices[terrain]
	if !shore {
		return ""
	}
	port := 0
	for i, neighbor := range hexgrid.Around(c) {
		if !g.Contains(neighbor) {
			continue
		}
		opposite := (i+3)%6 + 1
		for _, infra := range at(neighbor).Infra {
			ports, stream := strings.CutPrefix(infra, "stream")
			if !stream || ports == "" || strings.Trim(ports, "123456") != "" || !strings.ContainsRune(ports, rune('0'+opposite)) {
				continue
			}
			if port != 0 {
				return ""
			}
			port = i + 1
			break
		}
	}
	if port == 0 || !slices.Contains(land, port%6) || !slices.Contains(land, (port+1)%6) {
		return ""
	}
	id := TileID(fmt.Sprintf("%s-outlet%d", terrain, port))
	if !has(id) {
		return ""
	}
	return id
}
