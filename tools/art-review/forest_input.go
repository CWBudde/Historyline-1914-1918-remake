package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
)

type forestManifest struct {
	SchemaVersion int                           `json:"schema_version"`
	Grounds       map[string]connectionMaterial `json:"grounds"`
	Atlases       map[string]connectionMaterial `json:"atlases"`
	SummerProfile string                        `json:"summer_profile,omitempty"`
}

const forestSummerCanopyProfile = "canopy-26.5-30-v1"

type forestInputs struct {
	Manifest forestManifest
	Bytes    []byte
	Hash     string
	Grounds  map[string]compoundSource
	Atlases  map[string]compoundSource
}

func decodeForestManifest(data []byte, m *forestManifest) error {
	if err := decodeConnectionJSON(data, m); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	keys := []string{"schema_version", "grounds", "atlases"}
	if value, ok := raw["summer_profile"]; ok {
		if string(value) != `"canopy-26.5-30-v1"` {
			return errors.New("forest summer_profile requires canopy-26.5-30-v1")
		}
		keys = append(keys, "summer_profile")
	}
	root, err := groundObject(data, keys...)
	if err != nil {
		return err
	}
	if string(root["schema_version"]) != "1" {
		return errors.New("forest requires integer schema_version 1")
	}
	for _, group := range []string{"grounds", "atlases"} {
		refs, err := groundObject(root[group], "summer", "winter")
		if err != nil {
			return err
		}
		for _, season := range []string{"summer", "winter"} {
			ref, err := groundObject(refs[season], "source", "source_sha256")
			if err != nil {
				return err
			}
			for _, key := range []string{"source", "source_sha256"} {
				var value string
				if err := json.Unmarshal(ref[key], &value); err != nil || value == "" {
					return fmt.Errorf("forest %s %s requires nonempty %s", group, season, key)
				}
			}
		}
	}
	return nil
}
func validateForestAtlas(im *image.NRGBA) error {
	n := im.Bounds().Dx()
	if n != im.Bounds().Dy() || n < 512 || n > 4096 || n%2 != 0 {
		return errors.New("forest atlas must be even square 512..4096")
	}
	panel := n / 2
	margin := panel * 8 / 100
	for i := range 4 {
		opaque, transparentMargin := 0, 0
		for y := range panel {
			for x := range panel {
				a := im.NRGBAAt(i%2*panel+x, i/2*panel+y).A
				guard := x < margin || y < margin || x >= panel-margin || y >= panel-margin
				if a == 0 && guard {
					transparentMargin++
				}
				if a >= 250 {
					opaque++
					if guard {
						return fmt.Errorf("forest atlas panel %d opaque art enters 8%% guard", i)
					}
				}
			}
		}
		if opaque < 64 || transparentMargin == 0 {
			return fmt.Errorf("forest atlas panel %d requires opaque art and transparent margin", i)
		}
	}
	return nil
}
func loadForestInputs(path string) (forestInputs, error) {
	in := forestInputs{Grounds: map[string]compoundSource{}, Atlases: map[string]compoundSource{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeForestManifest(in.Bytes, &in.Manifest); err != nil {
		return in, fmt.Errorf("forest manifest: %w", err)
	}
	for _, season := range []string{"summer", "winter"} {
		ground, err := compoundReadSource(sourceBase(path), in.Manifest.Grounds[season], true)
		if err != nil {
			return in, fmt.Errorf("forest ground %s: %w", season, err)
		}
		atlas, err := compoundReadSource(sourceBase(path), in.Manifest.Atlases[season], false)
		if err != nil {
			return in, fmt.Errorf("forest atlas %s: %w", season, err)
		}
		if err = validateForestAtlas(atlas.Image); err != nil {
			return in, fmt.Errorf("forest atlas %s: %w", season, err)
		}
		in.Grounds[season], in.Atlases[season] = ground, atlas
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
