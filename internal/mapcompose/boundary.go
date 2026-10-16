package mapcompose

import (
	"strings"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
)

var boundaryTerrains = sync.OnceValue(func() map[string]data.Terrain {
	tables, err := data.Load()
	if err != nil {
		return nil
	}
	return tables.Terrain
})

// boundaryTerrain recognizes same-season terrain outside the active material
// family. Each family retreats to its seasonal base at such a boundary. Water
// has no seasonal suffix and is valid in both seasons. Unknown IDs and mixed
// seasons remain visible errors instead of silently acquiring a context.
func boundaryTerrain(id, season string) bool {
	if _, ok := boundaryTerrains()[id]; !ok {
		return false
	}
	if id == "water" || id == "waterStones" {
		return true
	}
	winter := strings.HasSuffix(id, "_w")
	return winter == (season == "winter")
}
