package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

type portraitReport struct {
	SchemaVersion     int                      `json:"schema_version"`
	Profile           string                   `json:"profile"`
	Family            string                   `json:"family"`
	Classes           []string                 `json:"classes"`
	LogicalDimensions dimensions               `json:"logical_dimensions"`
	Source            connectionMaterialReport `json:"source"`
	ManifestSHA256    string                   `json:"manifest_sha256"`
	ManifestSnapshot  string                   `json:"manifest_snapshot"`
	CropRectangle     packCrop                 `json:"crop_rectangle"`
	CropPolicy        string                   `json:"crop_policy"`
	Resampling        string                   `json:"resampling"`
	Opacity           string                   `json:"opacity"`
	Semantics         string                   `json:"semantics"`
	Acceptance        string                   `json:"acceptance"`
	Outputs           []packOutput             `json:"outputs"`
	CandidatePNGCount int                      `json:"candidate_png_count"`
	PhysicalPNGCount  int                      `json:"physical_png_count"`
	TotalFileCount    int                      `json:"total_file_count"`
}

type portraitArtifact struct {
	name string
	data []byte
}

type portraitBundle struct {
	report    portraitReport
	artifacts []portraitArtifact
}

func buildPortraitBundle(in portraitInputs) (portraitBundle, error) {
	b := portraitBundle{report: portraitReport{
		SchemaVersion:     1,
		Profile:           "portrait-infantry-v1",
		Family:            "infantry",
		Classes:           []string{"CP_infantry", "EC_infantry"},
		LogicalDimensions: dimensions{Width: 210, Height: 188},
		Source: connectionMaterialReport{
			connectionMaterial: in.Source.Ref,
			Dimensions:         dimensions{Width: in.Source.Image.Bounds().Dx(), Height: in.Source.Image.Bounds().Dy()},
			RetainedSource:     "infantry-source.png",
		},
		ManifestSHA256:    in.Hash,
		ManifestSnapshot:  "manifest.json",
		CropRectangle:     in.Crop,
		CropPolicy:        "Largest exact 105*k by 94*k rectangle, k=min(source_width/105,source_height/94) using integer division; k>=8. Center with floor offsets, discarding odd extra pixels on right/bottom. Copy crop without resampling to a zero-based NRGBA image before filtering; no stretch or upscaling.",
		Resampling:        "Unchanged resizeLanczos: antialiased premultiplied byte-RGB Lanczos3 with signed lobes, retained support normalized at edges, output-only half-up byte rounding and alpha/color clamping. Fixed chain cropped source -> 4x 840x752 -> 2x 420x376 -> 1x 210x188; no quantized upsampling.",
		Opacity:           "Every decoded source pixel must have alpha 255, including discarded overscan; all exports remain opaque. No alpha editing, background flattening, or magenta key treatment.",
		Semantics:         "One owner-neutral illustrated infantry portrait shared only by CP_infantry and EC_infantry. No faction recolouring, moved/spent grayscale, flags, text, frame, icons, or UI baked into this illustration. Source is the opaque art target, not a faction mask.",
		Acceptance:        "Provisional offline candidate. No runtime installation, original-game similarity target, native 96x77 resource alias binding, full UI acceptance, or whole graphics acceptance.",
		CandidatePNGCount: 3,
		PhysicalPNGCount:  4,
		TotalFileCount:    7,
	}}
	b.artifacts = append(b.artifacts, portraitArtifact{"infantry-source.png", in.Source.Bytes}, portraitArtifact{"manifest.json", in.Bytes})
	c := in.Crop
	current := image.NewNRGBA(image.Rect(0, 0, c.Width, c.Height))
	draw.Draw(current, current.Bounds(), in.Source.Image, image.Pt(c.X, c.Y), draw.Src)
	for _, density := range []int{4, 2, 1} {
		var err error
		current, err = resizeLanczos(current, 210*density, 188*density)
		if err != nil {
			return b, err
		}
		var encoded bytes.Buffer
		if err = png.Encode(&encoded, current); err != nil {
			return b, err
		}
		name := fmt.Sprintf("infantry-%dx.png", density)
		data := encoded.Bytes()
		sum := sha256.Sum256(data)
		b.artifacts = append(b.artifacts, portraitArtifact{name, data})
		b.report.Outputs = append(b.report.Outputs, packOutput{Path: name, SHA256: hex.EncodeToString(sum[:]), Dimensions: dimensions{Width: current.Bounds().Dx(), Height: current.Bounds().Dy()}})
	}
	report, err := json.MarshalIndent(b.report, "", "  ")
	if err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, portraitArtifact{"report.json", append(report, '\n')})
	gallery, err := template.New("portrait").Parse(portraitGallery)
	if err != nil {
		return b, err
	}
	var page bytes.Buffer
	if err = gallery.Execute(&page, b.report); err != nil {
		return b, err
	}
	b.artifacts = append(b.artifacts, portraitArtifact{"index.html", page.Bytes()})
	return b, nil
}

func runPortrait(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("portrait output directory already exists")
		}
		return err
	}
	in, err := loadPortraitInputs(path)
	if err != nil {
		return err
	}
	bundle, err := buildPortraitBundle(in)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0755); err != nil {
		return err
	}
	owned, err := os.Lstat(out)
	if err != nil {
		return err
	}
	defer func() {
		if resultErr == nil {
			return
		}
		current, statErr := os.Lstat(out)
		if errors.Is(statErr, os.ErrNotExist) {
			return
		}
		if statErr != nil || !os.SameFile(owned, current) {
			resultErr = errors.Join(resultErr, errors.New("portrait output identity changed; replacement preserved"), statErr)
			return
		}
		resultErr = errors.Join(resultErr, os.RemoveAll(out))
	}()
	for _, a := range bundle.artifacts {
		if err = saveExclusive(filepath.Join(out, a.name), func(w io.Writer) error { _, e := w.Write(a.data); return e }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote 3 portrait candidate PNGs, 4 physical PNGs and 7 total files to %s (provisional).\n", out)
	return err
}

const portraitGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Infantry portrait candidate</title><style>body{background:#252a25;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}img{image-rendering:auto;max-width:100%;height:auto}figure{margin:1rem 0}article{border-top:1px solid #888;padding:1rem 0}</style><h1>Infantry portrait candidate</h1><p>{{.Acceptance}}</p><p>One opaque owner-neutral infantry illustration shared by CP_infantry and EC_infantry.</p><p><a href="report.json">Export geometry and PNG pins</a> · <a href="manifest.json">Immutable input manifest</a> · <a href="{{.Source.RetainedSource}}">Opaque source art</a></p>{{range .Outputs}}<article><h2>{{.Dimensions.Width}} × {{.Dimensions.Height}}</h2><figure><a href="{{.Path}}"><img src="{{.Path}}" width="{{.Dimensions.Width}}" height="{{.Dimensions.Height}}" alt="Opaque painted infantry portrait"></a></figure></article>{{end}}<h2>Source art target</h2><a href="{{.Source.RetainedSource}}"><img src="{{.Source.RetainedSource}}" width="{{.Source.Dimensions.Width}}" height="{{.Source.Dimensions.Height}}" alt="Opaque authored infantry artwork"></a><details><summary>Fixed crop and resampling contract</summary><p>{{.CropPolicy}}</p><p>{{.Resampling}}</p><p>{{.Opacity}}</p><p>{{.Semantics}}</p></details></html>`
