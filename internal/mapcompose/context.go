// Package mapcompose turns a level's terrain into the ordered tile layers the map
// view draws, independently of rendering. It selects neighbour-dependent context
// tiles for ground, forest and hills, registers compound buildings and reports
// every tile it cannot select exactly, so gaps stay visible instead of invented.
//
// The context contract is a port of the offline exporters' descriptors
// (tools/art-review): the key of a tile is the SHA-256 of its descriptor and its
// runtime art is context/<family>/<key> in assets/index.json.
// TestDescriptorMatchesRecipeProfiles pins the key to the exporters' profiles.
package mapcompose

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Family is a group of terrains whose tiles depend on their neighbours.
type Family string

const (
	Ground Family = "ground" // mud, stone and the winter rough grounds over grass
	Forest Family = "forest"
	Hills  Family = "hills"
)

// Families lists the context families in a fixed order.
var Families = []Family{Ground, Forest, Hills}

// Phase period of the hills material, in map pixels.
const phaseX, phaseY = 568, 320

// directions names the six neighbours in hexgrid.Around order.
var directions = [6]string{"NE", "SE", "S", "SW", "NW", "N"}

// member returns the season and base terrain of id within family f, or ok false
// when id cannot appear in a context of f.
func member(f Family, id string) (season, base string, ok bool) {
	switch f {
	case Ground:
		switch id {
		case "grass", "mud", "stone":
			return "summer", "grass", true
		case "grass_w", "rough_w", "stoney_w", "shelled_w":
			return "winter", "grass_w", true
		}
	case Forest:
		switch id {
		case "forest", "grass":
			return "summer", "grass", true
		case "forest_w", "grass_w":
			return "winter", "grass_w", true
		}
	case Hills:
		if id == "grass" || id == "hills" {
			return "summer", "grass", true
		}
	}
	return "", "", false
}

// FamilyOf returns the context family whose tile for a centre of terrain id
// depends on its neighbours. Base terrains (grass) have none: their context is
// always the canonical default tile.
func FamilyOf(id string) (Family, bool) {
	for _, f := range Families {
		if _, base, ok := member(f, id); ok && id != base {
			return f, true
		}
	}
	return "", false
}

// Phase returns the hills material phase of the hex at (line, col): its tile
// origin modulo the material period.
func Phase(line, col int) (x, y int) {
	x = hexgrid.ColPitch*col + hexgrid.HalfPitch*(line&1)
	y = hexgrid.LinePitch * line
	return ((x % phaseX) + phaseX) % phaseX, ((y % phaseY) + phaseY) % phaseY
}

// Descriptor returns the context descriptor of a centre terrain with neighbours
// in NE, SE, S, SW, NW, N order. Phase is used by hills only. ok is false when
// the centre or a neighbour does not belong to f in the centre's season: the
// offline contract defines no tile for that combination.
func Descriptor(f Family, centre string, neighbors [6]string, phase [2]int) (string, bool) {
	season, base, ok := member(f, centre)
	if !ok {
		return "", false
	}
	for _, n := range neighbors {
		if s, _, ok := member(f, n); !ok || s != season {
			return "", false
		}
	}
	if centre == base {
		// A base centre looks the same whatever surrounds it.
		neighbors = [6]string{base, base, base, base, base, base}
		phase = [2]int{}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s-v1\nseason=%s\ncenter=%s\n", f, season, centre)
	if f == Hills {
		fmt.Fprintf(&b, "phase=%d,%d\n", phase[0], phase[1])
	}
	for i, d := range directions {
		fmt.Fprintf(&b, "%s=%s\n", d, neighbors[i])
	}
	return b.String(), true
}

// Key is the context key of a descriptor.
func Key(descriptor string) string {
	sum := sha256.Sum256([]byte(descriptor))
	return hex.EncodeToString(sum[:])
}

// TileID is the runtime asset ID of a terrain or infrastructure tile.
func TileID(runtimeID string) string { return "tile/" + runtimeID }

// ContextID is the runtime asset ID of a context tile.
func ContextID(f Family, key string) string { return "context/" + string(f) + "/" + key }
