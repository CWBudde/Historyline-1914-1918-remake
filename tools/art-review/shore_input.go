package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
)

type shoreVariant struct {
	ID           string `json:"id"`
	LandVertices []int  `json:"land_vertices"`
}
type shoreManifest struct {
	BankProfile   string                        `json:"bank_profile,omitempty"`
	Season        string                        `json:"season,omitempty"`
	SchemaVersion int                           `json:"schema_version"`
	Materials     map[string]connectionMaterial `json:"materials"`
	Variants      []shoreVariant                `json:"variants"`
}
type shoreInputs struct {
	Manifest  shoreManifest
	Bytes     []byte
	Hash      string
	Materials map[string]compoundSource
}

func shoreRoster() []shoreVariant {
	return []shoreVariant{
		{"lakeside12", []int{1, 2, 3}},
		{"lakeside16", []int{0, 1, 2}},
		{"lakeside23", []int{2, 3, 4, 5}},
		{"lakeside34", []int{3, 4, 5}},
		{"lakeside56", []int{0, 1, 5}},
		{"lakeside123", []int{1, 2, 3, 4}},
		{"lakeside126", []int{0, 1, 2, 3}},
		{"lakeside156", []int{0, 1, 2, 5}},
		{"lakeside3456", []int{0, 1, 3, 4, 5}},
	}
}

func shoreVariants() []shoreVariant {
	return append(shoreRoster(), shoreVariant{"grass", []int{0, 1, 2, 3, 4, 5}}, shoreVariant{"water", []int{}})
}

// Exact keys and non-null values supplement the shared duplicate/document/type checks.
func shoreObject(raw json.RawMessage, names ...string) (map[string]json.RawMessage, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	if obj == nil || len(obj) != len(names) {
		return nil, errors.New("shore object requires exactly its canonical fields")
	}
	for _, name := range names {
		v, ok := obj[name]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("shore requires non-null canonical field %q", name)
		}
	}
	return obj, nil
}

func decodeShoreManifest(b []byte, m *shoreManifest) error {
	if err := decodeConnectionJSON(b, m); err != nil {
		return err
	}
	fields := []string{"schema_version", "materials", "variants"}
	var optional map[string]json.RawMessage
	if err := json.Unmarshal(b, &optional); err != nil {
		return err
	}
	if _, ok := optional["bank_profile"]; ok {
		fields = append(fields, "bank_profile")
	}
	if _, ok := optional["season"]; ok {
		fields = append(fields, "season")
	}
	root, err := shoreObject(b, fields...)
	if err != nil {
		return err
	}
	materials, err := shoreObject(root["materials"], "grass", "water", "bank")
	if err != nil {
		return err
	}
	for _, name := range []string{"grass", "water", "bank"} {
		if _, err = shoreObject(materials[name], "source", "source_sha256"); err != nil {
			return err
		}
	}
	var variants []json.RawMessage
	if err = json.Unmarshal(root["variants"], &variants); err != nil {
		return err
	}
	for _, v := range variants {
		obj, e := shoreObject(v, "id", "land_vertices")
		if e != nil {
			return e
		}
		var corners []json.RawMessage
		if e = json.Unmarshal(obj["land_vertices"], &corners); e != nil {
			return e
		}
		for _, c := range corners {
			if bytes.Equal(bytes.TrimSpace(c), []byte("null")) {
				return errors.New("shore vertex must be an integer")
			}
		}
	}
	return nil
}

func loadShoreInputs(path string) (shoreInputs, error) {
	in := shoreInputs{Materials: map[string]compoundSource{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeShoreManifest(in.Bytes, &in.Manifest); err != nil {
		return in, err
	}
	if in.Manifest.SchemaVersion != 1 || len(in.Manifest.Variants) != 9 {
		return in, errors.New("shore requires schema_version 1 and exactly the nine known variants")
	}
	if in.Manifest.BankProfile != "" && in.Manifest.BankProfile != authoredBankProfile {
		return in, errors.New("shore bank_profile must be authored-rocks-v1 when set")
	}
	if in.Manifest.Season != "" && (in.Manifest.Season != "winter" || in.Manifest.BankProfile != authoredBankProfile) {
		return in, errors.New("shore season must be winter, with the authored-rocks-v1 bank, when set")
	}
	expected := map[string][]int{}
	for _, v := range shoreRoster() {
		expected[v.ID] = v.LandVertices
	}
	seen := map[string]bool{}
	for _, v := range in.Manifest.Variants {
		want, ok := expected[v.ID]
		if !ok || seen[v.ID] || !slices.Equal(want, v.LandVertices) {
			return in, fmt.Errorf("shore variant %q requires its unique exact sorted land_vertices profile", v.ID)
		}
		seen[v.ID] = true
	}
	for _, name := range []string{"grass", "water", "bank"} {
		src, e := compoundReadSource(sourceBase(path), in.Manifest.Materials[name], true)
		if e != nil {
			return in, fmt.Errorf("shore %s: %w", name, e)
		}
		in.Materials[name] = src
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
