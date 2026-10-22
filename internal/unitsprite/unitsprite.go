// Package unitsprite names the runtime unit sprite catalog: the family that draws
// each unit type, the six authored facings, the owner palette, the moved state and
// the stable visual IDs. cmd/hlasset builds packs under these IDs and the game
// looks sprites up by them, so both sides share one table.
package unitsprite

import (
	"fmt"
	"math"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

// families maps every unit type of internal/data to its sprite family. A family
// slug names the art folders art/units/<family>/ and art/portraits/<family>/;
// both sides of a class share one family and differ only by owner palette.
var families = map[string]string{
	"CP_infantry": "infantry", "EC_infantry": "infantry",
	"CP_eliteInf": "elite-infantry", "EC_eliteInf": "elite-infantry",
	"CP_cavalry": "cavalry", "EC_cavalry": "cavalry",
	"CP_lArtillery": "light-artillery", "EC_lArtillery": "light-artillery",
	"CP_mArtillery": "medium-artillery", "EC_mArtillery": "medium-artillery",
	"CP_hArtillery": "heavy-artillery", "EC_hArtillery": "heavy-artillery",
	"CP_supplyCar": "supply-car", "EC_supplyCar": "supply-car",
	"CP_bunker": "bunker", "EC_bunker": "bunker",
	"EC_charron":    "charron",
	"CP_armCar":     "armored-car",
	"CP_patrolBoat": "patrol-boat", "EC_patrolBoat": "patrol-boat",
	"CP_transTrain": "transport-train",
	"CP_armTrain":   "armored-train", "EC_armTrain": "armored-train",
	"CP_trainArt": "train-artillery", "EC_trainArt": "train-artillery",
	"CP_sappers": "sappers", "EC_sappers": "sappers",
}

// Family returns the sprite family of a unit type (a key of data.Set.Units).
func Family(unitType string) (string, bool) {
	f, ok := families[unitType]
	return f, ok
}

// Unbound art families have complete exports but no gameplay-table alias yet.
var unboundFamilies = []string{
	"a7v", "mark-i", "mark-iv", "renault-ft17", "st-chamond",
	"torpedo-boat", "destroyer", "battleship", "transport-ship", "submarine",
	"fokker-e-i", "fokker-e-iii", "albatros-d-iii", "fokker-dr-i", "fokker-d-vii", "junkers-j4",
	"morane-saulnier-n", "nieuport-xvii", "spad-vii", "spad-xiii", "se5a", "sopwith-camel",
	"dh2", "voisin-iii", "dh4", "gotha-g-vb", "handley-page-0400", "zeppelin-staaken-r-vi",
	"construction-unit", "static-aa", "mobile-aa", "anti-tank", "static-balloon",
}

// Families returns every exported family slug, sorted, including unbound art.
func Families() []string {
	out := slices.Clone(unboundFamilies)
	for _, f := range families {
		if !slices.Contains(out, f) {
			out = append(out, f)
		}
	}
	slices.Sort(out)
	return out
}

// ClassIDs returns the unit types drawn by family, sorted.
func ClassIDs(family string) []string {
	var out []string
	for id, f := range families {
		if f == family {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}

// Palette is the owner colouring of a sprite. It follows the current owner, never
// the unit's country, so captured units change colour.
type Palette string

const (
	CP      Palette = "cp"
	EC      Palette = "ec"
	Neutral Palette = "neutral"
)

// Palettes lists the authored palettes.
var Palettes = []Palette{CP, EC, Neutral}

// PaletteOf is the palette of a unit owned by p.
func PaletteOf(p game.Player) Palette {
	switch p {
	case game.CentralPowers:
		return CP
	case game.Entente:
		return EC
	}
	return Neutral
}

// State is the authored sprite state: normal, or spent (greyscale) once a unit has
// acted this turn. Remembered units reuse the spent sprite.
type State string

const (
	Normal State = "normal"
	Spent  State = "spent"
)

// States lists the authored states.
var States = []State{Normal, Spent}

// Facing is one of the six authored headings, numbered as the art exports'
// facing_index: clockwise from north is 0, 300, 240, 180, 120, 60 degrees.
type Facing int

const (
	N Facing = iota
	NW
	SW
	S
	SE
	NE
)

// Facings is the number of authored facings; range over it for all of them.
const Facings Facing = 6

var facingLabels = [Facings]string{"n", "nw", "sw", "s", "se", "ne"}

func (f Facing) String() string {
	if f < 0 || f >= Facings {
		return fmt.Sprintf("Facing(%d)", int(f))
	}
	return facingLabels[f]
}

// ParseFacing reads an export facing label (n, nw, sw, s, se, ne).
func ParseFacing(label string) (Facing, bool) {
	i := slices.Index(facingLabels[:], label)
	return Facing(i), i >= 0
}

// headings are the map-pixel steps to the neighbour each facing looks at (the hex
// geometry of the art reports: flat-topped 96×80 hexes, 71/40 px apart).
var headings = [Facings][2]float64{{0, -80}, {-71, -40}, {-71, 40}, {0, 80}, {71, 40}, {71, -40}}

// Toward returns the authored facing closest in angle to the map-pixel direction
// (dx, dy), y pointing down. A zero vector has no facing.
func Toward(dx, dy float64) (Facing, bool) {
	if dx == 0 && dy == 0 {
		return N, false
	}
	best, bestCos := N, math.Inf(-1)
	for f, h := range headings {
		if c := (dx*h[0] + dy*h[1]) / math.Hypot(h[0], h[1]); c > bestCos {
			best, bestCos = Facing(f), c
		}
	}
	return best, true
}

// ID is the stable visual ID of one sprite, e.g. unit/light-artillery/cp/nw-spent.
// Densities of the same ID live in separate packs.
func ID(family string, p Palette, f Facing, s State) string {
	return fmt.Sprintf("unit/%s/%s/%s-%s", family, p, f, s)
}
