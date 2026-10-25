package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime/debug"
)

type Asset struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Density    int    `json:"density,omitempty"`
	Provenance string `json:"provenance"`
}
type Report struct {
	Schema           int             `json:"schema_version"`
	Profile          string          `json:"profile"`
	RuleProfile      string          `json:"rule_profile"`
	Backend          string          `json:"backend"`
	RasterProvenance string          `json:"raster_input_provenance"`
	Inputs           []Input         `json:"inputs"`
	Assets           []Asset         `json:"assets"`
	Fixtures         []FixtureRender `json:"fixtures"`
}

func jsonBytes(v any) ([]byte, error) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}

func pngBytes(im image.Image) ([]byte, error) {
	var b bytes.Buffer
	e := png.Encode(&b, im)
	return b.Bytes(), e
}

func imageVersion() string {
	if b, ok := debug.ReadBuildInfo(); ok {
		for _, d := range b.Deps {
			if d.Path == "golang.org/x/image" {
				return d.Version
			}
		}
	}
	return "unknown"
}

func exportBundle(in *Inputs, out string) (err error) {
	if _, e := os.Lstat(out); e == nil {
		return fmt.Errorf("output already exists: %s", out)
	} else if !os.IsNotExist(e) {
		return e
	}
	for _, i := range in.Items {
		if pin(i.Bytes) != i.SHA256 {
			return fmt.Errorf("loaded input changed: %s", i.Key)
		}
	}
	if len(in.Items) != 11 {
		return fmt.Errorf("incomplete inputs")
	}
	if err = os.Mkdir(out, 0o755); err != nil {
		return err
	}
	owned, statErr := os.Lstat(out)
	if statErr != nil {
		return statErr
	}
	defer func() {
		if err != nil {
			cleanupOwned(out, owned)
		}
	}()
	write := func(path string, b []byte) error {
		p := filepath.Join(out, filepath.FromSlash(path))
		if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
			return e
		}
		return os.WriteFile(p, b, 0o644)
	}
	writeJSON := func(path string, v any) error {
		b, e := jsonBytes(v)
		if e != nil {
			return e
		}
		return write(path, b)
	}
	report := Report{Schema: 1, Profile: "unit-panel-v1", RuleProfile: "Standard", Backend: "golang.org/x/image " + imageVersion() + " vector + opentype DPI72 HintingNone; native 2x/4x; PNG draw.Over", RasterProvenance: "Existing portrait and tactical exports retain their Lanczos export provenance; same-density unscaled blits only.", Inputs: in.Items}
	images := map[string]image.Image{}
	assets := map[string]Asset{}
	for _, i := range in.Items {
		if err = write(i.Output, i.Bytes); err != nil {
			return err
		}
		a := Asset{i.Output, i.SHA256, i.Width, i.Height, 0, "pinned input snapshot"}
		assets[i.Output] = a
		report.Assets = append(report.Assets, a)
		if i.Image != nil {
			images[i.Output] = i.Image
		}
	}
	vs := vectors()
	for _, v := range vs {
		name := v.Name
		if name != "frame" {
			name = "icons/" + name
		}
		path := "assets/" + name + ".svg"
		b := []byte(v.svg())
		if err = write(path, b); err != nil {
			return err
		}
		a := Asset{path, pin(b), v.Width, v.Height, 1, "vector-source.json#" + v.Name}
		assets[path] = a
		report.Assets = append(report.Assets, a)
		for _, d := range []int{2, 4} {
			path = fmt.Sprintf("assets/%s.%dx.png", name, d)
			im := v.raster(d)
			b, err = pngBytes(im)
			if err != nil {
				return err
			}
			if err = write(path, b); err != nil {
				return err
			}
			a = Asset{path, pin(b), v.Width * d, v.Height * d, d, "native vector raster; no scaling"}
			assets[path] = a
			report.Assets = append(report.Assets, a)
			images[path] = im
		}
	}
	fs, e := makeFixtures()
	if e != nil {
		return e
	}
	for _, f := range fs {
		for _, d := range []int{2, 4} {
			records, e := compose(in, f.Panel, d, assets)
			if e != nil {
				return e
			}
			im, e := renderPanel(in, records, images, d)
			if e != nil {
				return e
			}
			b, e := pngBytes(im)
			if e != nil {
				return e
			}
			path := fmt.Sprintf("fixtures/%s-%dx.png", f.ID, d)
			if err = write(path, b); err != nil {
				return err
			}
			report.Fixtures = append(report.Fixtures, FixtureRender{f.ID, d, path, pin(b), records})
			report.Assets = append(report.Assets, Asset{path, pin(b), 400 * d, 640 * d, d, "ordered report draw records"})
		}
	}
	manifest := struct {
		Schema  int     `json:"schema_version"`
		Profile string  `json:"profile"`
		Inputs  []Input `json:"inputs"`
	}{1, "unit-panel-v1", in.Items}
	for _, v := range []struct {
		path string
		data any
	}{{"manifest.json", manifest}, {"fixtures.json", fs}, {"vector-source.json", vs}, {"report.json", report}} {
		if err = writeJSON(v.path, v.data); err != nil {
			return err
		}
	}
	if err = write("index.html", []byte(galleryHTML(fs, report.Fixtures))); err != nil {
		return err
	}
	return nil
}

// A competing process may replace a failed export's path. Only remove our directory.
func cleanupOwned(path string, owned os.FileInfo) {
	current, err := os.Lstat(path)
	if err == nil && current.Mode()&os.ModeSymlink == 0 && os.SameFile(owned, current) {
		_ = os.RemoveAll(path)
	}
}
