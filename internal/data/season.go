package data

import (
	"encoding/json"
	"strings"
	"sync"
)

// winterForms are the winter terrains not named kind+"_w": the original's
// winter set draws hills as rough ground, a mud field as shelled ground and
// stone as stony ground, on the same terrain records.
var winterForms = map[string]string{"hills": "rough_w", "mud": "shelled_w", "stone": "stoney_w"}

// terrainTable is the terrain table by terrain id.
var terrainTable = sync.OnceValue(func() map[string]Terrain {
	var t map[string]Terrain
	raw, err := files.ReadFile("terrain.json")
	if err == nil {
		err = json.Unmarshal(raw, &t)
	}
	if err != nil {
		panic("data: terrain.json: " + err.Error())
	}
	return t
})

// InSeason is terrain kind ("grass", "hills") in season ("summer",
// "autumn", "spring" or "winter"): in winter its winter form ("grass_w",
// "rough_w") where the terrain table has one, kind otherwise.
func InSeason(kind, season string) string {
	if season != "winter" {
		return kind
	}
	w, ok := winterForms[kind]
	if !ok {
		w = kind + "_w"
	}
	if _, ok := terrainTable()[w]; ok {
		return w
	}
	return kind
}

// Kind is terrain t without its season, the reverse of InSeason: "rough_w"
// is "hills", "grass_w" "grass".
func Kind(t string) string {
	base, ok := strings.CutSuffix(t, "_w")
	if !ok {
		return t
	}
	for kind, w := range winterForms {
		if w == t {
			return kind
		}
	}
	return base
}

// HasTerrain reports whether the terrain table has terrain t.
func HasTerrain(t string) bool {
	_, ok := terrainTable()[t]
	return ok
}
