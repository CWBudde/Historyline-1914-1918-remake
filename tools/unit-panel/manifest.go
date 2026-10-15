package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/image/font/opentype"
	"image"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Ref struct {
	Source string `json:"source"`
	SHA256 string `json:"source_sha256"`
}
type Densities struct {
	Two  Ref `json:"2x"`
	Four Ref `json:"4x"`
}
type Manifest struct {
	Schema   int       `json:"schema_version"`
	Profile  string    `json:"profile"`
	Portrait Densities `json:"portrait"`
	Units    struct {
		CP      Densities `json:"cp"`
		EC      Densities `json:"ec"`
		Neutral Densities `json:"neutral"`
	} `json:"units"`
	Fonts struct {
		Regular Ref `json:"regular"`
		Bold    Ref `json:"bold"`
		License Ref `json:"license"`
	} `json:"fonts"`
}
type Input struct {
	Key    string      `json:"key"`
	Source string      `json:"source"`
	Output string      `json:"output"`
	SHA256 string      `json:"sha256"`
	Width  int         `json:"width,omitempty"`
	Height int         `json:"height,omitempty"`
	Bytes  []byte      `json:"-"`
	Image  image.Image `json:"-"`
}
type Inputs struct {
	Manifest Manifest
	Items    []Input
	Fonts    map[string]*opentype.Font
}

func pin(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func loadInputs(path string) (*Inputs, error) {
	raw, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	return parseInputs(filepath.Dir(path), raw)
}

// checkJSON rejects duplicate keys as well as trailing documents before typed decoding.
func checkJSON(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	var value func() error
	value = func() error {
		tok, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || seen[s] {
					return fmt.Errorf("duplicate/invalid key %v", k)
				}
				seen[s] = true
				if e = value(); e != nil {
					return e
				}
			}
		case '[':
			for d.More() {
				if e = value(); e != nil {
					return e
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter")
		}
		_, e = d.Token()
		return e
	}
	if e := value(); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}
func parseInputs(base string, raw []byte) (*Inputs, error) {
	if e := checkJSON(raw); e != nil {
		return nil, e
	}
	if e := manifestFields(raw); e != nil {
		return nil, e
	}
	var m Manifest
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&m); e != nil {
		return nil, e
	}
	if m.Schema != 1 || m.Profile != "unit-panel-v1" {
		return nil, fmt.Errorf("unsupported schema/profile")
	}
	root, e := filepath.Abs(base)
	if e != nil {
		return nil, e
	}
	root, e = filepath.EvalSymlinks(root)
	if e != nil {
		return nil, e
	}
	in := &Inputs{Manifest: m, Fonts: map[string]*opentype.Font{}}
	add := func(key string, r Ref, w, h int) error {
		if len(r.SHA256) != 64 || r.SHA256 != strings.ToLower(r.SHA256) {
			return fmt.Errorf("%s: invalid pin", key)
		}
		if _, e := hex.DecodeString(r.SHA256); e != nil {
			return fmt.Errorf("%s: invalid pin", key)
		}
		if r.Source == "" || filepath.IsAbs(r.Source) || strings.Contains(r.Source, "\\") || filepath.ToSlash(filepath.Clean(r.Source)) != r.Source {
			return fmt.Errorf("%s: invalid source path", key)
		}
		for _, part := range strings.Split(strings.ToLower(r.Source), "/") {
			if part == ".." || part == "original" || part == "gfx" || strings.Contains(part, "legacy") {
				return fmt.Errorf("%s: forbidden source", key)
			}
		}
		if !strings.HasPrefix(r.Source, "exports/") && !strings.HasPrefix(r.Source, "source/fonts/") {
			return fmt.Errorf("%s: source outside authored exports/fonts", key)
		}
		path, e := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(r.Source)))
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(root, path)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("source escapes manifest directory")
		}
		for _, part := range strings.Split(strings.ToLower(filepath.ToSlash(rel)), "/") {
			if part == "original" || part == "gfx" || strings.Contains(part, "legacy") {
				return fmt.Errorf("resolved legacy source")
			}
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if pin(b) != r.SHA256 {
			return fmt.Errorf("%s: hash mismatch", key)
		}
		ext := ".png"
		if w == 0 {
			ext = ".ttf"
		}
		if key == "license" {
			ext = ".txt"
		}
		item := Input{Key: key, Source: r.Source, Output: "sources/" + key + ext, SHA256: r.SHA256, Bytes: b}
		if w > 0 {
			cfg, format, e := image.DecodeConfig(bytes.NewReader(b))
			if e != nil || format != "png" || cfg.Width != w || cfg.Height != h {
				return fmt.Errorf("%s: expected PNG %dx%d", key, w, h)
			}
			item.Image, _, e = image.Decode(bytes.NewReader(b))
			if e != nil {
				return e
			}
			item.Width = w
			item.Height = h
		} else if key != "license" {
			f, e := opentype.Parse(b)
			if e != nil {
				return fmt.Errorf("%s: malformed font: %w", key, e)
			}
			in.Fonts[key] = f
		} else if len(b) == 0 {
			return fmt.Errorf("empty font license")
		}
		in.Items = append(in.Items, item)
		return nil
	}
	for _, v := range []struct {
		k    string
		d    Densities
		w, h int
	}{{"portrait", m.Portrait, 210, 188}, {"unit-cp", m.Units.CP, 96, 96}, {"unit-ec", m.Units.EC, 96, 96}, {"unit-neutral", m.Units.Neutral, 96, 96}} {
		if e := add(v.k+"-2x", v.d.Two, v.w*2, v.h*2); e != nil {
			return nil, e
		}
		if e := add(v.k+"-4x", v.d.Four, v.w*4, v.h*4); e != nil {
			return nil, e
		}
	}
	for _, v := range []struct {
		k string
		r Ref
	}{{"regular", m.Fonts.Regular}, {"bold", m.Fonts.Bold}, {"license", m.Fonts.License}} {
		if e := add(v.k, v.r, 0, 0); e != nil {
			return nil, e
		}
	}
	return in, nil
}

// encoding/json matches field names without regard to case. The manifest does not.
func manifestFields(raw []byte) error {
	object := func(b []byte, keys ...string) (map[string]json.RawMessage, error) {
		var m map[string]json.RawMessage
		if e := json.Unmarshal(b, &m); e != nil {
			return nil, e
		}
		if len(m) != len(keys) {
			return nil, fmt.Errorf("manifest object requires exactly %v", keys)
		}
		for _, key := range keys {
			if _, ok := m[key]; !ok {
				return nil, fmt.Errorf("missing exact field %q", key)
			}
		}
		return m, nil
	}
	root, e := object(raw, "schema_version", "profile", "portrait", "units", "fonts")
	if e != nil {
		return e
	}
	units, e := object(root["units"], "cp", "ec", "neutral")
	if e != nil {
		return e
	}
	fonts, e := object(root["fonts"], "regular", "bold", "license")
	if e != nil {
		return e
	}
	refs := []json.RawMessage{fonts["regular"], fonts["bold"], fonts["license"]}
	for _, b := range []json.RawMessage{root["portrait"], units["cp"], units["ec"], units["neutral"]} {
		d, e := object(b, "2x", "4x")
		if e != nil {
			return e
		}
		refs = append(refs, d["2x"], d["4x"])
	}
	for _, r := range refs {
		if _, e := object(r, "source", "source_sha256"); e != nil {
			return e
		}
	}
	return nil
}
func (in *Inputs) item(key string) Input {
	for _, i := range in.Items {
		if i.Key == key {
			return i
		}
	}
	return Input{}
}
