package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func (c *collector) pulseGrassVariations(result map[string]map[string]picture) error {
	b, err := c.root.ReadFile("art/recipes/terrain/refdesign-v1/grass-variations.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var recipe struct {
		Variants []struct {
			ID      string `json:"runtime_id"`
			Outputs []struct {
				Path string `json:"path"`
				SHA  string `json:"sha256"`
			} `json:"outputs"`
		} `json:"variants"`
	}
	if err := json.Unmarshal(b, &recipe); err != nil {
		return err
	}
	wanted := map[string]bool{}
	for i := range 16 {
		wanted[fmt.Sprintf("grass-detail-%02d", i)] = true
	}
	for _, v := range recipe.Variants {
		if !wanted[v.ID] {
			return fmt.Errorf("unknown or duplicate grass detail %q", v.ID)
		}
		delete(wanted, v.ID)
		images := map[string]picture{}
		for _, o := range v.Outputs {
			for _, d := range []string{"2", "4"} {
				if strings.HasSuffix(o.Path, "-"+d+"x.png") {
					if images[d].Source != "" || len(o.SHA) != 64 {
						return fmt.Errorf("invalid grass detail export %s/%s", v.ID, d)
					}
					images[d] = picture{Source: o.Path, SHA: o.SHA}
				}
			}
		}
		if len(images) != 2 {
			return fmt.Errorf("grass detail %s requires both densities", v.ID)
		}
		result["tile/"+v.ID] = images
	}
	if len(wanted) != 0 {
		return errors.New("grass detail recipe requires sixteen variants")
	}
	return nil
}
