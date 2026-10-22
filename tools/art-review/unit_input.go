package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// Configurable choices are the explicit family, source pins, owner colours and
// one optional whole-atlas framing inset. Anchors, facing order and states stay fixed.
type unitManifest struct {
	SchemaVersion int                                      `json:"schema_version"`
	Family        string                                   `json:"family"`
	Atlas         connectionMaterial                       `json:"atlas"`
	Palettes      map[string][]int                         `json:"palettes"`
	Grounds       map[string]map[string]connectionMaterial `json:"grounds"`
	FramingInset  *float64                                 `json:"framing_inset,omitempty"`
}
type unitInputs struct {
	Manifest unitManifest
	Bytes    []byte
	Hash     string
	Atlas    compoundSource
	Grounds  map[string]map[int]compoundSource
}

func unitObject(raw json.RawMessage, names ...string) (map[string]json.RawMessage, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	if obj == nil || len(obj) != len(names) {
		return nil, errors.New("unit object requires exactly its canonical fields")
	}
	for _, name := range names {
		v, ok := obj[name]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("unit requires non-null canonical field %q", name)
		}
	}
	return obj, nil
}
func decodeUnitManifest(b []byte, m *unitManifest) error {
	if err := decodeConnectionJSON(b, m); err != nil {
		return err
	}
	fields := []string{"schema_version", "family", "atlas", "palettes", "grounds"}
	if m.FramingInset != nil {
		fields = append(fields, "framing_inset")
		if *m.FramingInset < .8 || *m.FramingInset > 1 {
			return errors.New("unit framing_inset must be between 0.8 and 1; one fixed inset applies to all six panels")
		}
	}
	root, err := unitObject(b, fields...)
	if err != nil {
		return err
	}
	if m.SchemaVersion != 1 {
		return errors.New("unit requires schema_version 1")
	}
	if _, err = lookupUnitFamily(m.Family); err != nil {
		return err
	}
	if _, err = unitObject(root["atlas"], "source", "source_sha256"); err != nil {
		return err
	}
	palettes, err := unitObject(root["palettes"], "cp", "ec", "neutral")
	if err != nil {
		return err
	}
	for _, name := range []string{"cp", "ec", "neutral"} {
		var slots []json.RawMessage
		if err = json.Unmarshal(palettes[name], &slots); err != nil {
			return err
		}
		if len(slots) != 3 {
			return errors.New("unit palette requires exactly three integer RGB slots")
		}
		for _, raw := range slots {
			var v int
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return errors.New("unit palette slot cannot be null")
			}
			if err = json.Unmarshal(raw, &v); err != nil {
				return err
			}
			if v < 0 || v > 255 {
				return errors.New("unit palette slot outside 0..255")
			}
		}
	}
	grounds, err := unitObject(root["grounds"], "grass", "grass_w", "mud")
	if err != nil {
		return err
	}
	for _, name := range []string{"grass", "grass_w", "mud"} {
		densities, e := unitObject(grounds[name], "2x", "4x")
		if e != nil {
			return e
		}
		for _, d := range []string{"2x", "4x"} {
			if _, e = unitObject(densities[d], "source", "source_sha256"); e != nil {
				return e
			}
		}
	}
	return nil
}
func unitDecodePNG(b []byte) (*image.NRGBA, error) {
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, im.Bounds().Dx(), im.Bounds().Dy()))
	draw.Draw(out, out.Bounds(), im, im.Bounds().Min, draw.Src)
	return out, nil
}
func unitValidateAtlas(im *image.NRGBA) error {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	cell := w / 2
	if w%2 != 0 || h%3 != 0 || cell != h/3 || cell < 384 || cell > 2048 {
		return errors.New("unit atlas requires six square cells, 2 columns by 3 rows, cell size 384..2048")
	}
	for face := range 6 {
		opaque, clear, key := 0, 0, 0
		ox, oy := face%2*cell, face/2*cell
		for y := range cell {
			for x := range cell {
				p := im.NRGBAAt(ox+x, oy+y)
				if p.A == 0 {
					clear++
				}
				if p.A >= 250 {
					if x == 0 || y == 0 || x == cell-1 || y == cell-1 {
						return fmt.Errorf("unit panel %d: opaque artwork reaches a source cell edge", face)
					}
					opaque++
				}
				if p.A > 0 && unitKey(p) > .5 {
					key++
				}
			}
		}
		if opaque < 64 || clear < 64 || key < 64 {
			return fmt.Errorf("unit panel %d requires substantial opaque art, transparent region and magenta key cloth (64 pixels each)", face)
		}
		if clear < cell*cell/4 {
			return fmt.Errorf("unit panel %d requires fully transparent background covering at least 25%% of its cell; got %d clear pixels", face, clear)
		}
	}
	return nil
}
func unitValidateGround(im *image.NRGBA, density int) error {
	if im.Bounds() != image.Rect(0, 0, 96*density, 80*density) {
		return errors.New("unit ground dimensions must match 96x80 at the pinned density")
	}
	g := sharedConnectionGeometry()
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			p := im.NRGBAAt(x, y)
			owned := connectionInside(g, connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)})
			if owned && p.A != 255 || !owned && p != (color.NRGBA{}) {
				return errors.New("unit ground requires opaque complementary owned pixels and zero RGBA outside")
			}
		}
	}
	return nil
}
func loadUnitInputs(path string) (unitInputs, error) {
	in := unitInputs{Grounds: map[string]map[int]compoundSource{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeUnitManifest(in.Bytes, &in.Manifest); err != nil {
		return in, err
	}
	in.Atlas, err = compoundReadSource(sourceBase(path), in.Manifest.Atlas, false)
	if err != nil {
		return in, fmt.Errorf("unit atlas: %w", err)
	}
	if err = unitValidateAtlas(in.Atlas.Image); err != nil {
		return in, err
	}
	for _, name := range []string{"grass", "grass_w", "mud"} {
		in.Grounds[name] = map[int]compoundSource{}
		for _, density := range []int{2, 4} {
			src, e := compoundReadSource(sourceBase(path), in.Manifest.Grounds[name][fmt.Sprintf("%dx", density)], false)
			if e != nil {
				return in, fmt.Errorf("unit ground %s %dx: %w", name, density, e)
			}
			if e = unitValidateGround(src.Image, density); e != nil {
				return in, e
			}
			in.Grounds[name][density] = src
		}
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
