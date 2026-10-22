package main

import (
	"fmt"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

// splitTerrain keeps both densities in identical parts, bounded by the larger
// images. Stable IDs, anchors and source paths are unchanged.
func splitTerrain(packs []terrainPack) ([]terrainPack, error) {
	byGroup := map[string]map[int]terrainPack{}
	var groups []string
	for _, p := range packs {
		if byGroup[p.Group] == nil {
			byGroup[p.Group] = map[int]terrainPack{}
			groups = append(groups, p.Group)
		}
		if _, exists := byGroup[p.Group][p.Density]; exists {
			return nil, fmt.Errorf("duplicate terrain group %s at %dx", p.Group, p.Density)
		}
		byGroup[p.Group][p.Density] = p
	}
	seen := map[string]string{}
	var result []terrainPack
	for _, group := range groups {
		paired := make([]map[string]assets.Entry, len(terrainDensities))
		var ids []string
		for i, d := range terrainDensities {
			p, ok := byGroup[group][d]
			if !ok {
				return nil, fmt.Errorf("terrain group %s missing %dx", group, d)
			}
			paired[i] = map[string]assets.Entry{}
			for _, e := range p.Manifest.Entries {
				if _, exists := paired[i][e.ID]; exists {
					return nil, fmt.Errorf("duplicate terrain ID %s at %dx", e.ID, d)
				}
				if e.Density != d || e.Width <= 0 || e.Height <= 0 || e.Width > 8192 || e.Height > 8192 || e.Width%d != 0 || e.Height%d != 0 {
					return nil, fmt.Errorf("terrain %s has invalid %dx geometry", e.ID, d)
				}
				paired[i][e.ID] = e
				if i == 0 {
					if prior, exists := seen[e.ID]; exists {
						return nil, fmt.Errorf("duplicate terrain ID %s in %s and %s", e.ID, prior, group)
					}
					seen[e.ID] = group
					ids = append(ids, e.ID)
				}
			}
			if len(paired[i]) != len(ids) {
				return nil, fmt.Errorf("terrain group %s has unpaired density entries", group)
			}
		}
		slices.Sort(ids)
		parts := [][]string{{}}
		var decoded int64
		for _, id := range ids {
			first := paired[0][id]
			var size int64
			for i, d := range terrainDensities {
				e, ok := paired[i][id]
				if !ok || e.Width/d != first.Width/first.Density || e.Height/d != first.Height/first.Density || e.Anchor != first.Anchor {
					return nil, fmt.Errorf("terrain %s has unpaired density geometry", id)
				}
				size = max(size, int64(e.Width)*int64(e.Height)*8)
			}
			if size > assets.MaxDecodedBytes {
				return nil, fmt.Errorf("terrain %s exceeds the decoded pack budget", id)
			}
			last := len(parts) - 1
			if len(parts[last]) == assets.MaxEntriesPerPack || decoded+size > assets.MaxDecodedBytes {
				parts = append(parts, nil)
				last++
				decoded = 0
			}
			parts[last] = append(parts[last], id)
			decoded += size
		}
		for n, part := range parts {
			name := group
			if n > 0 {
				name = fmt.Sprintf("%s-%d", group, n+1)
			}
			for i, d := range terrainDensities {
				m := byGroup[group][d].Manifest
				m.ID = fmt.Sprintf("terrain-%s-%dx", name, d)
				m.Entries = make([]assets.Entry, 0, len(part))
				for _, id := range part {
					m.Entries = append(m.Entries, paired[i][id])
				}
				result = append(result, terrainPack{Group: name, Density: d, Manifest: m})
			}
		}
	}
	return result, nil
}
