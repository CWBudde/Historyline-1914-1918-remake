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

type fogManifest struct {
	SchemaVersion int                           `json:"schema_version"`
	Materials     map[string]connectionMaterial `json:"materials"`
}

type fogInputs struct {
	Manifest fogManifest
	Bytes    []byte
	Hash     string
	Sources  map[string]compoundSource
}

// Inspect exact required keys as encoding/json otherwise accepts field-name
// aliases and null values for non-pointer struct fields.
func fogObject(raw []byte, keys ...string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	if object == nil || len(object) != len(keys) {
		return nil, errors.New("fog object requires exactly its canonical fields")
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return nil, fmt.Errorf("fog object missing canonical field %q", key)
		}
	}
	return object, nil
}

func decodeFogManifest(data []byte, m *fogManifest) error {
	if err := decodeConnectionJSON(data, m); err != nil {
		return err
	}
	root, err := fogObject(data, "schema_version", "materials")
	if err != nil {
		return err
	}
	if string(root["schema_version"]) != "1" {
		return errors.New("fog requires integer schema_version 1")
	}
	materials, err := fogObject(root["materials"], "unseen", "summer", "winter")
	if err != nil {
		return err
	}
	for _, id := range []string{"unseen", "summer", "winter"} {
		ref, err := fogObject(materials[id], "source", "source_sha256")
		if err != nil {
			return fmt.Errorf("fog material %s: %w", id, err)
		}
		for _, key := range []string{"source", "source_sha256"} {
			var value string
			if err := json.Unmarshal(ref[key], &value); err != nil || value == "" {
				return fmt.Errorf("fog material %s requires nonempty string %s", id, key)
			}
		}
	}
	return nil
}

func loadFogInputs(path string) (fogInputs, error) {
	in := fogInputs{Sources: map[string]compoundSource{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeFogManifest(in.Bytes, &in.Manifest); err != nil {
		return in, fmt.Errorf("fog manifest: %w", err)
	}
	for _, id := range []string{"unseen", "summer", "winter"} {
		// All three materials obey the established pinned opaque ground-source
		// loader, including minimum dimensions and resolved path safeguards.
		source, err := compoundReadSource(filepath.Dir(path), in.Manifest.Materials[id], true)
		if err != nil {
			return in, fmt.Errorf("fog material %s: %w", id, err)
		}
		in.Sources[id] = source
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
