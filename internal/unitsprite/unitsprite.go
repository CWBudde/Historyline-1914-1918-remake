// Package unitsprite names the runtime unit sprite catalog: the family that draws
// each unit type, the six authored facings, the owner palette, the moved state and
// the stable visual IDs. internal/hlasset builds packs under these IDs and the game
// looks sprites up by them, so both sides share one table.
package unitsprite

import (
	"fmt"
	"maps"
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
	"EC_charron": "charron",
	"CP_armCar":  "armored-car", "EC_armCar": "armored-car",
	"CP_patrolBoat": "patrol-boat", "EC_patrolBoat": "patrol-boat",
	"CP_transTrain": "transport-train", "EC_transTrain": "transport-train",
	"CP_armTrain": "armored-train", "EC_armTrain": "armored-train",
	"CP_trainArt": "train-artillery", "EC_trainArt": "train-artillery",
	"CP_sappers": "sappers", "EC_sappers": "sappers",
	"CP_a7v": "a7v", "EC_stChamond": "st-chamond",
	"CP_ft17": "renault-ft17", "EC_ft17": "renault-ft17",
	"CP_markI": "mark-i", "EC_markI": "mark-i",
	"CP_markIV": "mark-iv", "EC_markIV": "mark-iv",
	"CP_torpedoBoat": "torpedo-boat", "EC_torpedoBoat": "torpedo-boat",
	"CP_destroyer": "destroyer", "EC_destroyer": "destroyer",
	"CP_battleship": "battleship", "EC_battleship": "battleship",
	"CP_transShip": "transport-ship", "EC_transShip": "transport-ship",
	"CP_submarine": "submarine", "EC_submarine": "submarine",
	"CP_fokkerEI": "fokker-e-i", "CP_fokkerEIII": "fokker-e-iii",
	"CP_albatrosDIII": "albatros-d-iii", "CP_fokkerDrI": "fokker-dr-i",
	"CP_fokkerDVII": "fokker-d-vii", "CP_junkersJI": "junkers-j4",
	"EC_moraneN": "morane-saulnier-n", "EC_nieuport17": "nieuport-xvii",
	"EC_spad7": "spad-vii", "EC_spad13": "spad-xiii",
	"EC_se5a": "se5a", "EC_camel": "sopwith-camel",
	"EC_dh2": "dh2", "EC_voisin": "voisin-iii", "EC_dh4": "dh4",
	"CP_gothaGV": "gotha-g-vb", "EC_handleyPage": "handley-page-0400", "CP_staaken": "zeppelin-staaken-r-vi",
	"CP_construction": "construction-unit", "EC_construction": "construction-unit",
	"CP_staticAA": "static-aa", "EC_staticAA": "static-aa",
	"CP_mobileAA": "mobile-aa", "EC_mobileAA": "mobile-aa",
	"CP_antiTank": "anti-tank", "EC_antiTank": "anti-tank",
	"CP_balloon": "static-balloon", "EC_balloon": "static-balloon",
}

// Family returns the sprite family of a unit type (a key of data.Set.Units).
func Family(unitType string) (string, bool) {
	f, ok := families[unitType]
	return f, ok
}

// longs maps the family of each two-cell ship of the original (battleship,
// destroyer, submarine) to its long family: the same ship drawn across its
// head's and its rear's hex, on a LongSize canvas whose anchor is the midpoint
// between the two hex centres.
var longs = map[string]string{"battleship": "battleship-long", "destroyer": "destroyer-long", "submarine": "submarine-long"}

// LongSize is the logical width and height of a long family's sprites; every
// other family's are 96.
const LongSize = 160

// Long returns the long family of a two-cell ship's family.
func Long(family string) (string, bool) {
	l, ok := longs[family]
	return l, ok
}

// squadFamily is the family whose sprites show one man per point of strength:
// SquadSize men at full strength, drawn by the family itself; each weaker strength
// has its own family <family>-<n> with n men.
const squadFamily = "infantry"

// SquadSize is the strength at which the squad family shows all its men.
const SquadSize = 6

// teamFamily is the family drawn as a small team of 1..TeamLow men strength: the
// family itself shows the full team, <family>-1 the weakened one.
const teamFamily = "elite-infantry"

// TeamLow is the highest strength drawn by the weakened team family.
const TeamLow = 3

// Squad returns the family that draws family at the given strength: for the squad
// family a strength below SquadSize has its own family, for the team family a
// strength up to TeamLow; every other strength and family is drawn by family.
func Squad(family string, size int) string {
	switch {
	case family == squadFamily && size >= 1 && size < SquadSize:
		return fmt.Sprintf("%s-%d", family, size)
	case family == teamFamily && size >= 1 && size <= TeamLow:
		return family + "-1"
	}
	return family
}

// SquadFamilies returns the families of the squad family's weaker strengths.
func SquadFamilies() []string {
	var out []string
	for n := 1; n < SquadSize; n++ {
		out = append(out, Squad(squadFamily, n))
	}
	out = append(out, teamFamily+"-1")
	slices.Sort(out)
	return out
}

// Size is the logical width and height of family's sprites.
func Size(family string) int {
	for _, l := range longs {
		if l == family {
			return LongSize
		}
	}
	return 96
}

// Families returns every exported family slug, sorted; the long families are
// LongFamilies.
func Families() []string {
	var out []string
	for _, f := range families {
		if !slices.Contains(out, f) {
			out = append(out, f)
		}
	}
	slices.Sort(out)
	return out
}

// LongFamilies returns the long families, sorted.
func LongFamilies() []string {
	return slices.Sorted(maps.Values(longs))
}

// AllFamilies returns Families, LongFamilies and SquadFamilies, sorted: every family with a
// sprite pack.
func AllFamilies() []string {
	return slices.Sorted(slices.Values(slices.Concat(Families(), LongFamilies(), SquadFamilies())))
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
