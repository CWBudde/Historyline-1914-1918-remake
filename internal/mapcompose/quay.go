package mapcompose

import (
	"slices"
	"strings"
)

// quays returns the infrastructure drawn on a water hex. The levels keep the
// original's quays and piers as a road bridge carrying a road; on water that
// pair draws as the stone quay on its axis (tile/quay<ports>) where the
// catalog has it. The rules still see the bridge and the road.
func quays(infra []string, has func(string) bool) []string {
	var out []string
	for _, id := range infra {
		ports, bridge := strings.CutPrefix(id, "rBridge")
		if !bridge || !has(TileID("quay"+ports)) {
			continue
		}
		out = slices.DeleteFunc(slices.Clone(infra), func(i string) bool { return i == id || i == "road"+ports })
		return slices.Insert(out, 0, "quay"+ports)
	}
	return infra
}
