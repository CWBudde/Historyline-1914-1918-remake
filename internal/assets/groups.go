package assets

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"
)

// PackGroups discovers contiguous numeric parts of known groups at every requested
// density. The first part keeps the group's name, followed by -2, -3, etc.
// Optional groups may be absent, but an existing group must be complete.
func PackGroups(fsys fs.FS, required, optional []string, densities []int) ([]string, error) {
	if len(densities) == 0 {
		return nil, fmt.Errorf("no pack densities requested")
	}
	bases := slices.Concat(required, optional)
	byDensity := make([]map[string]map[int]string, len(densities))
	for i, d := range densities {
		dir := fmt.Sprintf("%dx", d)
		entries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			return nil, err
		}
		groups := map[string]map[int]string{}
		for _, entry := range entries {
			name := entry.Name()
			base, part := name, 1
			if !slices.Contains(bases, base) {
				cut := strings.LastIndexByte(name, '-')
				if cut < 0 {
					return nil, fmt.Errorf("unknown pack group %s/%s", dir, name)
				}
				base = name[:cut]
				part, err = strconv.Atoi(name[cut+1:])
				if err != nil || part < 2 || strconv.Itoa(part) != name[cut+1:] || !slices.Contains(bases, base) {
					return nil, fmt.Errorf("unknown or malformed pack group %s/%s", dir, name)
				}
			}
			if !entry.IsDir() {
				return nil, fmt.Errorf("pack group %s/%s is not a directory", dir, name)
			}
			info, err := fs.Stat(fsys, path.Join(dir, name, "manifest.json"))
			if err != nil {
				return nil, err
			}
			if !info.Mode().IsRegular() {
				return nil, fmt.Errorf("pack %s/%s has no regular manifest", dir, name)
			}
			if groups[base] == nil {
				groups[base] = map[int]string{}
			}
			groups[base][part] = name
		}
		byDensity[i] = groups
	}
	var result []string
	for _, base := range bases {
		parts := byDensity[0][base]
		if len(parts) == 0 && slices.Contains(required, base) {
			return nil, fmt.Errorf("missing required pack group %s", base)
		}
		for i, groups := range byDensity {
			if len(groups[base]) != len(parts) {
				return nil, fmt.Errorf("pack group %s differs at %dx", base, densities[i])
			}
		}
		var numbers []int
		for n := range parts {
			numbers = append(numbers, n)
		}
		slices.Sort(numbers)
		for i, n := range numbers {
			if n != i+1 {
				return nil, fmt.Errorf("pack group %s has noncontiguous parts", base)
			}
			for j, groups := range byDensity {
				if groups[base][n] != parts[n] {
					return nil, fmt.Errorf("pack group %s missing part %d at %dx", base, n, densities[j])
				}
			}
			result = append(result, parts[n])
		}
	}
	return result, nil
}

// CheckGroupDensities checks a catalogue of individually validated manifests.
// Each group supplies one manifest per requested density, in the same order.
// IDs must belong to exactly one group and keep their logical geometry.
func CheckGroupDensities(groups map[string][]Manifest, densities []int) error {
	if len(densities) == 0 {
		return fmt.Errorf("no pack densities requested")
	}
	seen := make([]map[string]bool, len(densities))
	for i := range seen {
		if densities[i] <= 0 {
			return fmt.Errorf("invalid pack density %d", densities[i])
		}
		seen[i] = map[string]bool{}
	}
	var names []string
	for name := range groups {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, group := range names {
		manifests := groups[group]
		if len(manifests) != len(densities) {
			return fmt.Errorf("group %s has incomplete densities", group)
		}
		for i, d := range densities {
			for _, e := range manifests[i].Entries {
				if e.Density != d {
					return fmt.Errorf("group %s: %s has density %d, want %d", group, e.ID, e.Density, d)
				}
				if seen[i][e.ID] {
					return fmt.Errorf("duplicate terrain ID %s at %dx", e.ID, d)
				}
				seen[i][e.ID] = true
			}
		}
	}
	for _, group := range names {
		first := map[string]Entry{}
		for i, d := range densities {
			m := groups[group][i]
			for _, e := range m.Entries {
				if i == 0 {
					first[e.ID] = e
					continue
				}
				prior, ok := first[e.ID]
				if !ok || e.Width/d != prior.Width/densities[0] || e.Height/d != prior.Height/densities[0] || e.Anchor != prior.Anchor {
					return fmt.Errorf("group %s: %s has unpaired density geometry", group, e.ID)
				}
			}
			if len(m.Entries) != len(first) {
				return fmt.Errorf("group %s has unpaired density entries", group)
			}
		}
	}
	return nil
}
