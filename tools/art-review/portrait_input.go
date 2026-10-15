package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

type portraitManifest struct {
	SchemaVersion int                `json:"schema_version"`
	Family        string             `json:"family"`
	Source        connectionMaterial `json:"source"`
}

type portraitInputs struct {
	Manifest portraitManifest
	Bytes    []byte
	Hash     string
	Source   compoundSource
	Crop     packCrop
}

func decodePortraitManifest(b []byte, m *portraitManifest) error {
	if err := decodeConnectionJSON(b, m); err != nil {
		return err
	}
	root, err := unitObject(b, "schema_version", "family", "source")
	if err != nil {
		return err
	}
	if _, err = unitObject(root["source"], "source", "source_sha256"); err != nil {
		return err
	}
	if m.SchemaVersion != 1 || (m.Family != "infantry" && m.Family != "lightArtillery" && m.Family != "eliteInfantry") {
		return errors.New("portrait requires schema_version 1 and canonical family infantry, lightArtillery or eliteInfantry")
	}
	return nil
}

func portraitCrop(bounds image.Rectangle) (packCrop, error) {
	w, h := bounds.Dx(), bounds.Dy()
	k := min(w/105, h/94)
	if k < 8 {
		return packCrop{}, errors.New("portrait largest exact 105:94 crop must be at least 840x752; upscaling excluded")
	}
	return packCrop{X: (w - 105*k) / 2, Y: (h - 94*k) / 2, Width: 105 * k, Height: 94 * k}, nil
}

func loadPortraitInputs(path string) (portraitInputs, error) {
	var in portraitInputs
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodePortraitManifest(in.Bytes, &in.Manifest); err != nil {
		return in, err
	}
	// The unchanged reader's opaque branch validates the entire decoded image,
	// including overscan later discarded by the portrait crop. Its minimum
	// ground size is weaker than the exact portrait crop minimum below.
	in.Source, err = compoundReadSource(filepath.Dir(path), in.Manifest.Source, true)
	if err != nil {
		return in, err
	}
	// The shared reader converts to 8-bit NRGBA before its opacity check.
	// A 16-bit alpha such as 65534 can round to 255 there. Validate original
	// precision from the retained bytes, without rereading the source file.
	if in.Source.Bytes[24] == 16 {
		decoded, decodeErr := png.Decode(bytes.NewReader(in.Source.Bytes))
		if decodeErr != nil {
			return in, decodeErr
		}
		bounds := decoded.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if _, _, _, alpha := decoded.At(x, y).RGBA(); alpha != 65535 {
					return in, errors.New("portrait source must be fully opaque at original 16-bit precision, including discarded overscan")
				}
			}
		}
	}
	in.Crop, err = portraitCrop(in.Source.Image.Bounds())
	if err != nil {
		return in, err
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
