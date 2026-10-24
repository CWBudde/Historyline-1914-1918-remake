package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// A winter mixed manifest derives the winter footbridges and level crossings
// from the pinned summer mixed manifest: each listed variant keeps its
// geometry, timber infill and steel, and every component it is built from
// (the crossing's layers, the footbridge's approach path) is swapped for the
// component of the same ID with a "_w" suffix from the pinned winter
// connected manifests. Exports carry the runtime ID with "_w".
// mixedWinterInfillSnow blends the timber infill of a crossing over a winter
// track toward the winter ground.
const mixedWinterInfillSnow = .3

type winterMixedManifest struct {
	SchemaVersion int                  `json:"schema_version"`
	Base          connectionMaterial   `json:"base"`
	Manifests     []connectionMaterial `json:"manifests"`
	Variants      []string             `json:"variants"`
}

type winterMixedInputs struct {
	base     mixedInputs
	variants []mixedVariant
	winter   map[string]mixedComponent // summer component ID -> its winter form
}

// winterMixedComponentIDs lists the summer components a footbridge or
// crossing is rendered from.
func winterMixedComponentIDs(v mixedVariant) []string {
	if v.Kind == "footbridge" {
		return []string{"path" + mixedAxis(v.Ports)}
	}
	return v.Layers
}

func loadWinterMixedInputs(path string) (winterMixedInputs, error) {
	w := winterMixedInputs{winter: map[string]mixedComponent{}}
	b, err := os.ReadFile(path)
	if err != nil {
		return w, err
	}
	var m winterMixedManifest
	if err := decodeConnectionJSON(b, &m); err != nil {
		return w, fmt.Errorf("winter mixed manifest: %w", err)
	}
	if m.SchemaVersion != 1 || len(m.Manifests) == 0 || len(m.Variants) == 0 {
		return w, errors.New("winter mixed requires schema_version 1, a base, winter manifests and variants")
	}
	resolve := func(ref connectionMaterial) (string, error) {
		if ref.Source == "" || filepath.IsAbs(ref.Source) {
			return "", errors.New("winter mixed references must be relative")
		}
		if _, err := hex.DecodeString(ref.SourceSHA256); err != nil || len(ref.SourceSHA256) != 64 {
			return "", fmt.Errorf("%s: winter mixed reference requires SHA256", ref.Source)
		}
		return filepath.Join(sourceBase(path), ref.Source), nil
	}
	basePath, err := resolve(m.Base)
	if err != nil {
		return w, err
	}
	if w.base, err = loadMixedInputs(basePath); err != nil {
		return w, err
	}
	if !strings.EqualFold(w.base.hash, m.Base.SourceSHA256) {
		return w, errors.New("winter mixed base manifest SHA256 mismatch")
	}
	components := map[string]mixedComponent{}
	for _, ref := range m.Manifests {
		p, err := resolve(ref)
		if err != nil {
			return w, err
		}
		src, err := loadPinnedMixedConnection(p, ref.SourceSHA256)
		if err != nil {
			return w, fmt.Errorf("%s: %w", ref.Source, err)
		}
		for _, f := range src.manifest.Families {
			for _, v := range f.Variants {
				if _, ok := components[v.ID]; ok {
					return w, fmt.Errorf("duplicate winter component %s", v.ID)
				}
				components[v.ID] = mixedComponent{src, f, v}
			}
		}
	}
	roster := map[string]mixedVariant{}
	for _, v := range w.base.manifest.Variants {
		roster[v.ID] = v
	}
	seen := map[string]bool{}
	for _, id := range m.Variants {
		v, ok := roster[id]
		if !ok || seen[id] || v.Kind != "footbridge" && v.Kind != "crossing" {
			return w, fmt.Errorf("winter mixed variant %q must be a footbridge or crossing of the base, listed once", id)
		}
		seen[id] = true
		for _, cid := range winterMixedComponentIDs(v) {
			summer, ok := w.base.components[cid]
			if !ok {
				return w, fmt.Errorf("%s: base has no component %s", id, cid)
			}
			c, ok := components[cid+"_w"]
			if !ok {
				return w, fmt.Errorf("%s: no winter component %s_w", id, cid)
			}
			if c.family.ID != summer.family.ID || !reflect.DeepEqual(c.variant.Ports, summer.variant.Ports) || !mixedProfileCompatible(c.family) {
				return w, fmt.Errorf("%s: winter component %s_w differs from %s in family, ports or profile", id, cid, cid)
			}
			w.winter[cid] = c
		}
		w.variants = append(w.variants, v)
	}
	return w, nil
}

// inputs is the base mixed input with every component swapped for its
// winter form.
func (w winterMixedInputs) inputs() mixedInputs {
	in := w.base
	in.components = map[string]mixedComponent{}
	for id, c := range w.base.components {
		in.components[id] = c
	}
	for id, c := range w.winter {
		in.components[id] = c
	}
	return in
}
