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

type hillsManifest struct {
	SchemaVersion int                           `json:"schema_version"`
	Materials     map[string]connectionMaterial `json:"materials"`
}
type hillsInputs struct {
	Manifest hillsManifest
	Bytes    []byte
	Hash     string
	Sources  map[string]compoundSource
}

func decodeHillsManifest(data []byte, m *hillsManifest) error {
	if err := decodeConnectionJSON(data, m); err != nil {
		return err
	}
	root, err := groundObject(data, "schema_version", "materials")
	if err != nil {
		return err
	}
	if string(root["schema_version"]) != "1" {
		return errors.New("hills requires integer schema_version 1")
	}
	ids := []string{"grass", "height"}
	var materialKeys map[string]json.RawMessage
	if err := json.Unmarshal(root["materials"], &materialKeys); err != nil {
		return err
	}
	if _, ok := materialKeys["scrub"]; ok {
		ids = append(ids, "scrub")
	}
	refs, err := groundObject(root["materials"], ids...)
	if err != nil {
		return err
	}
	for _, id := range ids {
		ref, err := groundObject(refs[id], "source", "source_sha256")
		if err != nil {
			return err
		}
		for _, key := range []string{"source", "source_sha256"} {
			var s string
			if err := json.Unmarshal(ref[key], &s); err != nil || s == "" {
				return fmt.Errorf("hills %s requires nonempty %s", id, key)
			}
		}
	}
	return nil
}

func validateHillsHeight(im *image.NRGBA) error {
	n := im.Bounds().Dx()
	if n != im.Bounds().Dy() || n < 512 || n > 4096 {
		return errors.New("hills height must be square 512..4096")
	}
	for y := range n {
		for x := range n {
			p := im.NRGBAAt(x, y)
			if p.A != 255 || int(max(p.R, p.G, p.B))-int(min(p.R, p.G, p.B)) > 8 {
				return errors.New("hills height must be opaque with RGB channel spread <=8")
			}
		}
	}
	return nil
}

func loadHillsInputs(path string) (hillsInputs, error) {
	in := hillsInputs{Sources: map[string]compoundSource{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeHillsManifest(in.Bytes, &in.Manifest); err != nil {
		return in, fmt.Errorf("hills manifest: %w", err)
	}
	ids := []string{"grass", "height"}
	if _, ok := in.Manifest.Materials["scrub"]; ok {
		ids = append(ids, "scrub")
	}
	for _, id := range ids {
		source, err := compoundReadSource(sourceBase(path), in.Manifest.Materials[id], true)
		if err != nil {
			return in, fmt.Errorf("hills %s: %w", id, err)
		}
		if id == "height" {
			if err = validateHillsHeight(source.Image); err != nil {
				return in, err
			}
		}
		in.Sources[id] = source
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
