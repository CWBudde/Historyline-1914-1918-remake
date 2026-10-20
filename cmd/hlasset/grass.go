package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

const grassVariationRecipe = "terrain/refdesign-v1/grass-variations.json"

// Optional detail tiles share the background pack and its normal registration.
// A catalog made before the grass study remains usable without this recipe.
func planGrassVariations(recipes string) ([]assets.Entry, error) {
	b, err := readFileBounded(filepath.Join(recipes, grassVariationRecipe), maxRecipe)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var recipe struct {
		Variants []struct {
			ID      string `json:"runtime_id"`
			Outputs []struct {
				Path string `json:"path"`
			} `json:"outputs"`
		} `json:"variants"`
	}
	if err := json.Unmarshal(b, &recipe); err != nil {
		return nil, err
	}
	pins := map[string]pin{}
	if err := recipePins(b, pins); err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for i := range 16 {
		wanted[fmt.Sprintf("grass-detail-%02d", i)] = true
	}
	var entries []assets.Entry
	for _, variant := range recipe.Variants {
		if !wanted[variant.ID] {
			return nil, fmt.Errorf("unknown or duplicate grass variation %q", variant.ID)
		}
		delete(wanted, variant.ID)
		have := map[int]bool{}
		for _, output := range variant.Outputs {
			e, err := terrainEntry("tile/"+variant.ID, output.Path, pins)
			if err != nil {
				return nil, err
			}
			if (e.Density != 2 && e.Density != 4) || have[e.Density] {
				return nil, fmt.Errorf("grass variation %s repeats or has invalid density", variant.ID)
			}
			have[e.Density] = true
			entries = append(entries, e)
		}
		if !have[2] || !have[4] {
			return nil, fmt.Errorf("grass variation %s lacks both densities", variant.ID)
		}
	}
	if len(wanted) != 0 {
		return nil, errors.New("grass variation recipe must provide sixteen variants")
	}
	return entries, nil
}
