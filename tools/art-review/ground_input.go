package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var groundMaterialIDs = [...]string{"grass", "grass_w", "mud", "stone", "rough_w", "stoney_w", "shelled_w"}
var groundTargetIDs = [...]string{"mud", "stone", "rough_w", "stoney_w", "shelled_w"}

type groundManifest struct {
	SchemaVersion int                           `json:"schema_version"`
	Materials     map[string]connectionMaterial `json:"materials"`
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
	root, err := groundObject(data, "schema_version", "materials")
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
		source, err := compoundReadSource(filepath.Dir(path), in.Manifest.Materials[id], true)
		if err != nil {
			return in, fmt.Errorf("ground material %s: %w", id, err)
		}
		in.Sources[id] = source
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
