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

// Shore marks a Gap of a neighbour-derived shore tile. It is not a context
// family: the tile ID is the descriptor.
const Shore Family = "shore"

// Water hexes draw a shore derived from their neighbours, so a coast never
// meets land along a straight hex edge. Each of a hex's six vertices (the
// shoreline export's order, clockwise from the top left) is land when any of
// the three hexes at it claims it: a land hex claims all its vertices, a
// lakeside hex its authored land vertices, open water none. Every hex at a
// vertex computes the same state, so neighbouring tiles agree, and authored
// coasts that already agree keep their tiles.

// vertexShares[v] are the two neighbours at vertex v, as hexgrid.Around
// indices, and the vertex's index in each neighbour.
var vertexShares = func() (s [6][2][2]int) {
	for v := range 6 {
		s[v] = [2][2]int{{(v + 5) % 6, (v + 4) % 6}, {(v + 4) % 6, (v + 2) % 6}}
	}
	return s
}()

// edgeVertices are the vertices of the edge toward hexgrid.Around index i
// (port i+1).
func edgeVertices(i int) (int, int) { return (i + 1) % 6, (i + 2) % 6 }

// wet reports whether terrain is open water or a shore.
func wet(terrain string) bool {
	_, shore := shoreLandVertices[terrain]
	return shore || terrain == "water" || terrain == "waterStones"
}

// claims reports whether a hex of terrain claims its vertex v as land.
func claims(terrain string, v int) bool {
	if land, shore := shoreLandVertices[terrain]; shore {
		return slices.Contains(land, v)
	}
	return !wet(terrain)
}

// ShoreShape is a water hex's derived shore: Land has bit v set for each land
// vertex, Channels bit e for each edge e (from vertex e to e+1) whose two land
// vertices leave water open through its middle, toward a neighbour that is
// water too.
type ShoreShape struct{ Land, Channels uint8 }

// shoreShape derives the shore of the water hex c.
func shoreShape(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord) ShoreShape {
	centre := at(c).Terrain
	around := hexgrid.Around(c)
	terrain := func(i int) (string, bool) {
		if !g.Contains(around[i]) {
			return "", false
		}
		return at(around[i]).Terrain, true
	}
	var s ShoreShape
	for v := range 6 {
		land := claims(centre, v)
		for _, n := range vertexShares[v] {
			if t, ok := terrain(n[0]); ok && claims(t, n[1]) {
				land = true
			}
		}
		if land {
			s.Land |= 1 << v
		}
	}
	for i := range 6 {
		a, b := edgeVertices(i)
		t, ok := terrain(i)
		if !ok || !wet(t) || s.Land&(1<<a) == 0 || s.Land&(1<<b) == 0 {
			continue
		}
		// The neighbour sees the same edge from the opposite side.
		na, nb := edgeVertices((i + 3) % 6)
		authored := func(terrain string, a, b int) bool {
			_, shore := shoreLandVertices[terrain]
			return shore && claims(terrain, a) && claims(terrain, b)
		}
		if !authored(centre, a, b) && !authored(t, na, nb) {
			s.Channels |= 1 << a
		}
	}
	return s
}

// ID is the runtime ID of the shore tile: one of the nine lakeside shores
// where it matches one, else shore-<land>[-c<channels>] in hex.
func (s ShoreShape) ID() string {
	if s.Channels == 0 {
		for id, land := range shoreLandVertices {
			if mask(land) == s.Land {
				return id
			}
		}
	}
	if s.Channels == 0 {
		return fmt.Sprintf("shore-%02x", s.Land)
	}
	return fmt.Sprintf("shore-%02x-c%02x", s.Land, s.Channels)
}

func mask(vertices []int) uint8 {
	var m uint8
	for _, v := range vertices {
		m |= 1 << v
	}
	return m
}

// selectShore returns the runtime terrain ID drawn on water hex c (its own
// terrain where nothing needs to change), or a gap when its derived shore tile
// is not in the catalog.
func selectShore(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord, has func(string) bool) (string, *Gap) {
	terrain := at(c).Terrain
	if terrain == "waterStones" || !wet(terrain) {
		return terrain, nil
	}
	s := shoreShape(g, at, c)
	if s.Land == 0 {
		return "water", nil
	}
	id := s.ID()
	// The nine lakeside shores are base tiles, drawn like any terrain's.
	if _, lakeside := shoreLandVertices[id]; lakeside || has(TileID(id)) {
		return id, nil
	}
	return terrain, &Gap{At: c, Terrain: terrain, Family: Shore, Reason: Missing, Descriptor: id}
}

// ParseShore reads a derived shore tile ID (ShoreShape.ID) back into its shape.
func ParseShore(id string) (ShoreShape, bool) {
	if land, shore := shoreLandVertices[id]; shore {
		return ShoreShape{Land: mask(land)}, true
	}
	var s ShoreShape
	if n, _ := fmt.Sscanf(id, "shore-%02x-c%02x", &s.Land, &s.Channels); n == 2 && s.ID() == id {
		return s, true
	}
	if n, _ := fmt.Sscanf(id, "shore-%02x", &s.Land); n == 1 && s.Channels == 0 && s.ID() == id {
		return s, true
	}
	return ShoreShape{}, false
}
