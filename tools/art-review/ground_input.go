package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
)

var (
	groundMaterialIDs = [...]string{"grass", "grass_w", "mud", "stone", "rough_w", "stoney_w", "shelled_w"}
	groundTargetIDs   = [...]string{"mud", "stone", "rough_w", "stoney_w", "shelled_w"}
)

type groundManifest struct {
	SchemaVersion  int                           `json:"schema_version"`
	Materials      map[string]connectionMaterial `json:"materials"`
	MudFeature     string                        `json:"mud_feature,omitempty"`
	ShelledFeature string                        `json:"shelled_feature,omitempty"`
	// LatticeMaterials are the target materials sampled unmirrored and
	// lattice-periodic, by source px per world px (newLatticeGroundRenderer).
	LatticeMaterials map[string]float64 `json:"lattice_materials,omitempty"`
}
type groundInputs struct {
	Manifest groundManifest
	Bytes    []byte
	Hash     string
	Sources  map[string]compoundSource
}

func groundObject(raw []byte, keys ...string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	if object == nil || len(object) != len(keys) {
		return nil, errors.New("ground object requires exactly its canonical fields")
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return nil, fmt.Errorf("ground object missing canonical field %q", key)
		}
	}
	return object, nil
}

func decodeGroundManifest(data []byte, m *groundManifest) error {
	if err := decodeConnectionJSON(data, m); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	keys := []string{"schema_version", "materials"}
	if value, ok := raw["mud_feature"]; ok {
		if string(value) != `"circular-earthwork"` {
			return errors.New("mud_feature requires circular-earthwork")
		}
		keys = append(keys, "mud_feature")
	}
	if value, ok := raw["shelled_feature"]; ok {
		if string(value) != `"circular-earthwork"` {
			return errors.New("shelled_feature requires circular-earthwork")
		}
		keys = append(keys, "shelled_feature")
	}
	if _, ok := raw["lattice_materials"]; ok {
		if len(m.LatticeMaterials) == 0 {
			return errors.New("lattice_materials requires at least one material")
		}
		for id, k := range m.LatticeMaterials {
			if !slices.Contains(groundTargetIDs[:], id) || k <= 0 {
				return fmt.Errorf("lattice_materials: %s requires a target material and a positive scale", id)
			}
		}
		keys = append(keys, "lattice_materials")
	}
	root, err := groundObject(data, keys...)
	if err != nil {
		return err
	}
	if string(root["schema_version"]) != "1" {
		return errors.New("ground requires integer schema_version 1")
	}
	materials, err := groundObject(root["materials"], groundMaterialIDs[:]...)
	if err != nil {
		return err
	}
	for _, id := range groundMaterialIDs {
		ref, err := groundObject(materials[id], "source", "source_sha256")
		if err != nil {
			return fmt.Errorf("ground material %s: %w", id, err)
		}
		for _, key := range []string{"source", "source_sha256"} {
			var value string
			if err := json.Unmarshal(ref[key], &value); err != nil || value == "" {
				return fmt.Errorf("ground material %s requires nonempty string %s", id, key)
			}
		}
	}
	return nil
}

func loadGroundInputs(path string) (groundInputs, error) {
	in := groundInputs{Sources: map[string]compoundSource{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeGroundManifest(in.Bytes, &in.Manifest); err != nil {
		return in, fmt.Errorf("ground manifest: %w", err)
	}
	for _, id := range groundMaterialIDs {
		// The loader reads each file once; its exact pinned bytes also supply decoding and snapshots.
		source, err := compoundReadSource(sourceBase(path), in.Manifest.Materials[id], true)
		if err != nil {
			return in, fmt.Errorf("ground material %s: %w", id, err)
		}
		in.Sources[id] = source
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
