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
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type packManifest struct {
	SchemaVersion int         `json:"schema_version"`
	ArtDirection  string      `json:"art_direction"`
	Assets        []packAsset `json:"assets"`
}

type packAsset struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	SourceSHA256 string `json:"source_sha256"`
	Description  string `json:"description"`
}

type packCrop struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type packOutput struct {
	Path       string     `json:"path"`
	SHA256     string     `json:"sha256"`
	Dimensions dimensions `json:"dimensions"`
}

type packAssetReport struct {
	packAsset
	SourceDimensions dimensions   `json:"source_dimensions"`
	CropRectangle    packCrop     `json:"crop_rectangle"`
	SourcePreview    string       `json:"source_preview"`
	Outputs          []packOutput `json:"outputs"`
}

type packReport struct {
	SchemaVersion    int               `json:"schema_version"`
	ManifestSHA256   string            `json:"manifest_sha256"`
	ArtDirection     string            `json:"art_direction"`
	Status           string            `json:"status"`
	Resampling       string            `json:"resampling"`
	CropMethod       string            `json:"crop_method"`
	ResamplingMethod string            `json:"resampling_method"`
	ResamplingChain  []string          `json:"resampling_chain"`
	Acceptance       string            `json:"acceptance"`
	Assets           []packAssetReport `json:"assets"`
}

// runPack deliberately uses only manifest sources, never reference geometry or pixels.
// Validate and encode all inputs before claiming a fresh output directory.
func runPack(manifestPath, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return fmt.Errorf("pack output directory already exists: %s", out)
		}
		return err
	}
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var manifest packManifest
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("pack manifest: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return errors.New("pack manifest must contain exactly one JSON object")
	}
	if manifest.SchemaVersion != 1 || strings.TrimSpace(manifest.ArtDirection) == "" || len(manifest.Assets) == 0 {
		return errors.New("pack requires schema_version 1, art_direction, and at least one asset")
	}
	manifestHash := sha256.Sum256(manifestBytes)
	report := packReport{
		ManifestSHA256: hex.EncodeToString(manifestHash[:]),
		SchemaVersion:  1, ArtDirection: manifest.ArtDirection, Status: "candidate", Resampling: "lanczos",
		CropMethod:       "Largest exact 6:5 integer rectangle (6k by 5k, k=min(width/6,height/5)); centered with floor offsets, discarding an odd extra pixel at right/bottom. No stretching or upscaling.",
		ResamplingMethod: "Existing alpha-aware antialiased Lanczos3: sinc(t)*sinc(t/3), widened support, pixel centers, clipped and renormalized edges; premultiplied byte RGB, float64 accumulation, clamp then unpremultiply, round half up once per reduction. No linear-light conversion.",
		ResamplingChain:  []string{"source -> centered 6:5 crop (no resampling)", "crop -> raw4x (384x320): lanczos3", "raw4x -> raw2x (192x160): lanczos3"},
		Acceptance:       "Candidate only. Rectangle exports preserve overscan. Renderer clipping and multi-tile joins acceptance pending; no seamlessness or in-game acceptance claim.",
	}
	type artifact struct {
		name string
		data []byte
	}
	var artifacts []artifact
	seen := make(map[string]bool)
	safeID := regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
	for _, asset := range manifest.Assets {
		if !safeID.MatchString(asset.ID) || seen[strings.ToLower(asset.ID)] {
			return fmt.Errorf("pack asset id must be safe and unique: %q", asset.ID)
		}
		seen[strings.ToLower(asset.ID)] = true
		if strings.TrimSpace(asset.Description) == "" || strings.TrimSpace(asset.Source) == "" {
			return fmt.Errorf("asset %s requires source and description", asset.ID)
		}
		wantHash, err := hex.DecodeString(asset.SourceSHA256)
		if err != nil || len(wantHash) != sha256.Size {
			return fmt.Errorf("asset %s requires a 64-digit SHA256", asset.ID)
		}
		sourcePath := asset.Source
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(sourceBase(manifestPath), sourcePath)
		}
		resolved, err := filepath.EvalSymlinks(sourcePath)
		if err != nil {
			return fmt.Errorf("asset %s source: %w", asset.ID, err)
		}
		absolute, err := filepath.Abs(resolved)
		if err != nil {
			return err
		}
		for _, part := range strings.Split(filepath.ToSlash(absolute), "/") {
			if strings.EqualFold(part, "original") {
				return fmt.Errorf("asset %s: original installation sources are excluded from pack mode", asset.ID)
			}
		}
		sourceBytes, err := os.ReadFile(resolved)
		if err != nil {
			return fmt.Errorf("asset %s source: %w", asset.ID, err)
		}
		actualHash := sha256.Sum256(sourceBytes)
		if !bytes.Equal(actualHash[:], wantHash) {
			return fmt.Errorf("asset %s source SHA256 mismatch", asset.ID)
		}
		// Decode the verified bytes, rather than reopening a mutable input path.
		cfg, err := png.DecodeConfig(bytes.NewReader(sourceBytes))
		if err != nil {
			return fmt.Errorf("asset %s PNG: %w", asset.ID, err)
		}
		if cfg.Width < 384 || cfg.Height < 320 || int64(cfg.Width)*int64(cfg.Height) > maxInputPixels {
			return fmt.Errorf("asset %s PNG must be at least 384x320 and at most 64 megapixels", asset.ID)
		}
		decoded, err := png.Decode(bytes.NewReader(sourceBytes))
		if err != nil {
			return fmt.Errorf("asset %s PNG: %w", asset.ID, err)
		}
		k := min(cfg.Width/6, cfg.Height/5)
		crop := packCrop{(cfg.Width - 6*k) / 2, (cfg.Height - 5*k) / 2, 6 * k, 5 * k}
		src := image.NewNRGBA(image.Rect(0, 0, crop.Width, crop.Height))
		// Opacity applies to the entire source, including the discarded overscan.
		for y := 0; y < cfg.Height; y++ {
			for x := 0; x < cfg.Width; x++ {
				p := decoded.At(x+decoded.Bounds().Min.X, y+decoded.Bounds().Min.Y)
				if _, _, _, a := p.RGBA(); a != 65535 {
					return fmt.Errorf("asset %s must be opaque (pixel %d,%d)", asset.ID, x, y)
				}
				if x >= crop.X && x < crop.X+crop.Width && y >= crop.Y && y < crop.Y+crop.Height {
					src.Set(x-crop.X, y-crop.Y, p)
				}
			}
		}
		raw4x, err := resizeLanczos(src, 384, 320)
		if err != nil {
			return err
		}
		raw2x, err := resizeLanczos(raw4x, 192, 160)
		if err != nil {
			return err
		}
		asset.SourceSHA256 = hex.EncodeToString(actualHash[:])
		assetReport := packAssetReport{packAsset: asset, SourceDimensions: dimensions{cfg.Width, cfg.Height}, CropRectangle: crop, SourcePreview: asset.ID + "-source.png"}
		artifacts = append(artifacts, artifact{assetReport.SourcePreview, sourceBytes})
		for _, output := range []struct {
			suffix string
			img    *image.NRGBA
		}{{"2x", raw2x}, {"4x", raw4x}} {
			var encoded bytes.Buffer
			if err := png.Encode(&encoded, output.img); err != nil {
				return err
			}
			name := asset.ID + "-" + output.suffix + ".png"
			sum := sha256.Sum256(encoded.Bytes())
			assetReport.Outputs = append(assetReport.Outputs, packOutput{name, hex.EncodeToString(sum[:]), dimensions{output.img.Bounds().Dx(), output.img.Bounds().Dy()}})
			artifacts = append(artifacts, artifact{name, encoded.Bytes()})
		}
		report.Assets = append(report.Assets, assetReport)
	}
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	artifacts = append(artifacts, artifact{"report.json", append(reportBytes, '\n')})
	gallery, err := template.New("pack").Parse(packGallery)
	if err != nil {
		return err
	}
	var html bytes.Buffer
	if err := gallery.Execute(&html, report); err != nil {
		return err
	}
	artifacts = append(artifacts, artifact{"index.html", html.Bytes()})
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(out, 0755); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, os.RemoveAll(out))
		}
	}()
	for _, artifact := range artifacts {
		if err := saveExclusive(filepath.Join(out, artifact.name), func(w io.Writer) error { _, err := w.Write(artifact.data); return err }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d candidate assets to %s; renderer clipping and joins acceptance pending.\n", len(report.Assets), out)
	return err
}

const packGallery = `<!doctype html>
<html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Candidate terrain pack</title>
<style>body{font:16px system-ui;background:#20251f;color:#eee;margin:2rem}article{border-top:1px solid #77816f;padding:1rem 0}section{display:flex;flex-wrap:wrap;gap:1rem;align-items:flex-start}figure{margin:0}img{display:block;max-width:100%;height:auto}figure.source{width:384px}figcaption{padding:.4rem 0}a{color:#bce0a4}</style>
<h1>Candidate terrain pack</h1><p>{{.ArtDirection}}</p><p>{{.Acceptance}}</p>
<p>Full sources and rectangular 2x / 4x exports below. No footprint clipping has been applied. <a href="report.json">Reproducible build report</a>.</p>
{{range .Assets}}<article><h2>{{.ID}}</h2><p>{{.Description}}</p><section>
<figure class="source"><a href="{{.SourcePreview}}"><img src="{{.SourcePreview}}" alt="{{.ID}} full source"></a><figcaption>Full source · {{.SourceDimensions.Width}}×{{.SourceDimensions.Height}}</figcaption></figure>
{{range .Outputs}}<figure><a href="{{.Path}}"><img src="{{.Path}}" alt="{{.Path}}"></a><figcaption>{{.Path}} · {{.Dimensions.Width}}×{{.Dimensions.Height}}</figcaption></figure>{{end}}
</section></article>{{end}}
</html>
`
