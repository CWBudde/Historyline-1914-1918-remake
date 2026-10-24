package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type mixedVariant struct {
	ID        string   `json:"id"`
	RuntimeID string   `json:"runtime_id"`
	Kind      string   `json:"kind"`
	Ports     []int    `json:"ports,omitempty"`
	Layers    []string `json:"layers,omitempty"`
}
type mixedManifest struct {
	SchemaVersion int                  `json:"schema_version"`
	BridgeArtwork string               `json:"bridge_artwork,omitempty"`
	Manifests     []connectionMaterial `json:"manifests"`
	Variants      []mixedVariant       `json:"variants"`
}

func mixedRoster() []mixedVariant {
	return []mixedVariant{
		{ID: "bridge25", RuntimeID: "bridge25", Kind: "footbridge", Ports: []int{2, 5}},
		{ID: "bridge36", RuntimeID: "bridge36", Kind: "footbridge", Ports: []int{3, 6}},
		{ID: "rbridge14", RuntimeID: "rBridge14", Kind: "road-bridge", Ports: []int{1, 4}},
		{ID: "rbridge25", RuntimeID: "rBridge25", Kind: "road-bridge", Ports: []int{2, 5}},
		{ID: "rbridge36", RuntimeID: "rBridge36", Kind: "road-bridge", Ports: []int{3, 6}},
		{ID: "cross1", RuntimeID: "cross1", Kind: "crossing", Layers: []string{"path14", "rail36"}},
		{ID: "cross2", RuntimeID: "cross2", Kind: "crossing", Layers: []string{"road14", "rail36"}},
		{ID: "cross3", RuntimeID: "cross3", Kind: "crossing", Layers: []string{"road25", "rail36"}},
		{ID: "cross4", RuntimeID: "cross4", Kind: "crossing", Layers: []string{"path36", "rail14"}},
	}
}
func decodeMixedManifest(b []byte) (mixedManifest, error) {
	var m mixedManifest
	if err := decodeConnectionJSON(b, &m); err != nil {
		return m, err
	}
	// Go's case-insensitive field aliases are deliberately excluded, including
	// present-but-empty kind-specific fields.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return m, err
	}
	for k := range raw {
		if k != "schema_version" && k != "bridge_artwork" && k != "manifests" && k != "variants" {
			return m, fmt.Errorf("noncanonical mixed field %q", k)
		}
	}
	if _, present := raw["bridge_artwork"]; present && m.BridgeArtwork != "standing" {
		return m, errors.New("bridge_artwork must be standing when specified")
	}
	var vs []map[string]json.RawMessage
	if err := json.Unmarshal(raw["variants"], &vs); err != nil {
		return m, err
	}
	for i, v := range vs {
		for k := range v {
			if k != "id" && k != "runtime_id" && k != "kind" && k != "ports" && k != "layers" {
				return m, fmt.Errorf("noncanonical variant field %q", k)
			}
		}
		_, ports := v["ports"]
		_, layers := v["layers"]
		if m.Variants[i].Kind == "crossing" {
			if ports || !layers {
				return m, errors.New("crossing requires layers and excludes ports")
			}
		} else if layers || !ports {
			return m, errors.New("bridge requires ports and excludes layers")
		}
	}
	var refs []map[string]json.RawMessage
	if err := json.Unmarshal(raw["manifests"], &refs); err != nil {
		return m, err
	}
	for _, r := range refs {
		for k := range r {
			if k != "source" && k != "source_sha256" {
				return m, fmt.Errorf("noncanonical manifest reference field %q", k)
			}
		}
	}
	if m.SchemaVersion != 1 || len(m.Variants) != 9 || len(m.Manifests) != 3 {
		return m, errors.New("mixed requires schema_version 1, three manifests and the nine bounded variants")
	}
	roster := map[string]mixedVariant{}
	for _, v := range mixedRoster() {
		roster[v.ID] = v
	}
	seen := map[string]bool{}
	for _, v := range m.Variants {
		if seen[v.ID] || !reflect.DeepEqual(v, roster[v.ID]) {
			return m, fmt.Errorf("unsupported or duplicate mixed variant %q", v.ID)
		}
		seen[v.ID] = true
	}
	return m, nil
}

// mixedProfileCompatible reports whether a component family keeps the width
// and border the bridges and crossings are built for.
func mixedProfileCompatible(f connectionFamily) bool {
	expected := map[string][2]float64{"road": {14, 2}, "stream": {18, 2.5}, "path": {7, .75}, "rail": {11, 1}}
	profile, ok := expected[f.ID]
	if f.ID == "road" && (f.Markings == "dashed" || f.Markings == "ruts") {
		profile = [2]float64{22, 1}
		if f.Border == 2 {
			profile[1] = 2
		}
	}
	if f.ID == "stream" && f.BorderMaterial == "bank" && f.Border == 3.5 {
		profile[1] = 3.5
	}
	return ok && f.Width == profile[0] && f.Border == profile[1]
}

type mixedComponent struct {
	input   connectionInputs
	family  connectionFamily
	variant connectionVariant
}
type mixedInputs struct {
	manifest     mixedManifest
	manifestData []byte
	hash         string
	merged       connectionInputs
	components   map[string]mixedComponent
	definitions  map[string][]byte
}

func (in mixedInputs) retainedManifest() connectionArtifact {
	return connectionArtifact{"mixed-manifest.json", in.manifestData}
}

func loadPinnedMixedConnection(path, pin string) (connectionInputs, error) {
	in, err := loadConnectionInputs(path)
	if err != nil {
		return in, err
	}
	if !strings.EqualFold(in.hash, pin) {
		return in, errors.New("mixed loaded source manifest SHA256 mismatch")
	}
	return in, nil
}

func loadMixedInputs(path string) (mixedInputs, error) {
	in := mixedInputs{components: map[string]mixedComponent{}, definitions: map[string][]byte{}, merged: connectionInputs{manifest: connectionManifest{SchemaVersion: 1, Geometry: sharedConnectionGeometry(), Materials: map[string]connectionMaterial{}}, materials: map[string]*image.NRGBA{}, sources: map[string][]byte{}, grassSampling: true}}
	b, err := os.ReadFile(path)
	if err != nil {
		return in, err
	}
	in.manifest, err = decodeMixedManifest(b)
	if err != nil {
		return in, err
	}
	sum := sha256.Sum256(b)
	in.hash = hex.EncodeToString(sum[:])
	in.manifestData = b
	seen := map[string]bool{}
	for _, ref := range in.manifest.Manifests {
		if ref.Source == "" || filepath.IsAbs(ref.Source) {
			return in, errors.New("mixed manifest references must be relative")
		}
		p := filepath.Join(sourceBase(path), ref.Source)
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return in, err
		}
		absolute, err := filepath.Abs(resolved)
		if err != nil {
			return in, err
		}
		for _, s := range []string{p, absolute} {
			for _, part := range strings.Split(filepath.ToSlash(s), "/") {
				if strings.EqualFold(part, "original") || strings.EqualFold(part, "pilot") {
					return in, errors.New("original and pilot sources excluded from mixed")
				}
			}
		}
		if seen[absolute] {
			return in, errors.New("duplicate mixed manifest reference")
		}
		seen[absolute] = true
		data, err := os.ReadFile(resolved)
		if err != nil {
			return in, err
		}
		sum := sha256.Sum256(data)
		want, err := hex.DecodeString(ref.SourceSHA256)
		if err != nil || len(want) != 32 || !bytes.Equal(sum[:], want) {
			return in, errors.New("mixed source manifest SHA256 mismatch")
		}
		src, err := loadPinnedMixedConnection(p, ref.SourceSHA256)
		if err != nil {
			return in, err
		}
		in.definitions[ref.Source] = data
		for id, mat := range src.manifest.Materials {
			old, ok := in.merged.manifest.Materials[id]
			if ok && old.SourceSHA256 != mat.SourceSHA256 {
				return in, fmt.Errorf("conflicting material %s", id)
			}
			in.merged.manifest.Materials[id] = mat
			in.merged.materials[id] = src.materials[id]
			in.merged.sources[id] = src.sources[id]
		}
		for _, f := range src.manifest.Families {
			if !mixedProfileCompatible(f) {
				return in, fmt.Errorf("mixed incompatible %s profile", f.ID)
			}
			for _, v := range f.Variants {
				if _, ok := in.components[v.ID]; ok {
					return in, fmt.Errorf("duplicate component %s", v.ID)
				}
				in.components[v.ID] = mixedComponent{src, f, v}
			}
		}
	}
	for _, family := range []string{"path", "road", "rail", "stream"} {
		for _, axis := range []string{"14", "25", "36"} {
			id := family + axis
			c, ok := in.components[id]
			if !ok || c.family.ID != family || !reflect.DeepEqual(c.variant.Ports, []int{int(axis[0] - '0'), int(axis[1] - '0')}) {
				return in, fmt.Errorf("missing/wrong straight component %s", id)
			}
		}
	}
	for _, id := range []string{"grass", "road", "water", "earth", "path", "rail", "timber", "steel"} {
		if in.merged.materials[id] == nil {
			return in, fmt.Errorf("missing material %s", id)
		}
	}
	return in, nil
}
