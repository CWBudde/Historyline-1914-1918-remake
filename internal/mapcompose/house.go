package mapcompose

import (
	"fmt"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// MaxHouseVariants bounds the numbered house tiles beside tile/house:
// tile/house2 … tile/house8 and their winter forms tile/house2_w …, numbered
// from 2 without gaps.
const MaxHouseVariants = 8

// houseTiles are the tiles a house hex of terrain may draw: tile/house and the
// numbered variants of its season the catalog has; nil for other terrains.
func houseTiles(terrain string, has func(string) bool) []string {
	var sfx string
	switch terrain {
	case "house":
	case "house_w":
		sfx = "_w"
	default:
		return nil
	}
	ids := []string{TileID(terrain)}
	for v := 2; v <= MaxHouseVariants; v++ {
		id := TileID(fmt.Sprintf("house%d%s", v, sfx))
		if !has(id) {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

// houseTile is the tile of a house hex (line, col), seeded by the
// coordinates; ok is false for other terrains.
func houseTile(terrain string, line, col int, has func(string) bool) (string, bool) {
	ids := houseTiles(terrain, has)
	if ids == nil {
		return "", false
	}
	return ids[(coordinateHash(line, col)>>32)%uint64(len(ids))], true
}

// varyHouses gives neighbouring houses different farmsteads where the catalog
// has enough: in line, then column order, a house drawing the same tile as a
// house on an earlier line beside it takes the next variant that differs.
// Harbour and hangar hexes keep their ground.
func (s *Scene) varyHouses(at func(hexgrid.Coord) Hex, has func(string) bool) {
	for l := range s.Cells {
		for col := range s.Cells[l] {
			c := hexgrid.Coord{Line: l, Col: col}
			ids := houseTiles(at(c).Terrain, has)
			cell := &s.Cells[l][col]
			i := slices.Index(ids, cell.Background)
			if len(ids) < 2 || i < 0 {
				continue
			}
			var taken []string
			for _, n := range s.Grid.Neighbors(c) {
				if n.Line < l {
					taken = append(taken, s.Cells[n.Line][n.Col].Background)
				}
			}
			for range ids {
				if !slices.Contains(taken, ids[i]) {
					break
				}
				i = (i + 1) % len(ids)
			}
			cell.Background = ids[i]
		}
	}
}
