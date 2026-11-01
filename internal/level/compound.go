package level

import (
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Compound is a terrain that covers a fixed group of hexes, each a piece
// named Prefix plus the piece's name ("cmpN", "cmpW", …), with the season
// suffix "_w" on a winter hex. The N piece anchors the group.
type Compound struct {
	Prefix string
	// Entrance is the building piece that holds the building's content and
	// takes its container record; "" for a compound that is no building.
	Entrance string
	// Pieces walk from the N piece to each piece through hexgrid.Around
	// directions.
	Pieces map[string][]int
}

// hexgrid.Around directions.
const (
	dirSE, dirS, dirSW = 1, 2, 3
)

func diamond() map[string][]int {
	return map[string][]int{"N": nil, "W": {dirSW}, "E": {dirSE}, "S": {dirS}}
}

// Compounds are the HQ (seven hexes), the depot or camp and the factory
// (diamonds) and the mountain quarters, which form no building.
var Compounds = []Compound{
	{"hq", "N", map[string][]int{
		"N": nil, "NW": {dirSW}, "NE": {dirSE}, "C": {dirS},
		"SW": {dirS, dirSW}, "SE": {dirS, dirSE}, "S": {dirS, dirS},
	}},
	{"cmp", "N", diamond()},
	{"factory", "E", diamond()},
	{"mount", "", diamond()},
}

// At is the hex of piece p of the group whose N piece is at n.
func (cp *Compound) At(n hexgrid.Coord, p string) hexgrid.Coord {
	for _, d := range cp.Pieces[p] {
		n = hexgrid.Around(n)[d]
	}
	return n
}

// CompoundPiece splits a terrain into its compound, piece name and season
// suffix; nil for a terrain that is no compound piece.
func CompoundPiece(terrain string) (cp *Compound, piece, suffix string) {
	base, winter := strings.CutSuffix(terrain, "_w")
	if winter {
		suffix = "_w"
	}
	for i := range Compounds {
		p, ok := strings.CutPrefix(base, Compounds[i].Prefix)
		if _, known := Compounds[i].Pieces[p]; ok && known {
			return &Compounds[i], p, suffix
		}
	}
	return nil, "", ""
}
