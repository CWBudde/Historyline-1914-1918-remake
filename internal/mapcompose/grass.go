package mapcompose

import "fmt"

// GrassVariationID identifies one of sixteen optional summer-grass interiors.
// The coordinate permutation changes the authored material phase along every lattice
// direction. All sixteen exports retain the canonical material at their edges.
func GrassVariationID(line, col int) string {
	variant := ((line%4+4)%4)*4 + ((col+3*line)%4+4)%4
	return fmt.Sprintf("tile/grass-detail-%02d", variant)
}
