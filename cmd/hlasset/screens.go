package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/image/draw"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

// screenDensities are the densities of the screen catalog: the backdrop and
// the scene paintings have no 4x.
var screenDensities = []int{1, 2}

const screenW, screenH = 640, 400 // logical size of every full-screen image

// screenFile is one image of the screen catalog: a runtime file of
// assets/index.json or a scene painting scaled from art/.
type screenFile struct {
	Family  string // art/screens/<family>
	ID      string
	Density int
	Width   int
	Height  int
	Anchor  [2]float64
	read    func() ([]byte, error)
}

// planScreens returns the files of the screen catalog at density d: every
// unpacked runtime file under screens/ at d, and scene/<family> for each
// art/screens/<family>/illustration.png.
func planScreens(entries []indexEntry, assetsDir, artDir string, d int) ([]screenFile, error) {
	var files []screenFile
	seen := map[string]bool{}
	for _, e := range entries {
		if e.Pack != "" || !strings.HasPrefix(e.Art, "screens/") || e.Density != d {
			continue
		}
		if seen[e.ID] {
			return nil, fmt.Errorf("screen asset %s at %dx listed twice", e.ID, d)
		}
		seen[e.ID] = true
		name := filepath.Join(assetsDir, filepath.FromSlash(e.Path))
		var anchor [2]float64
		if e.Anchor != nil {
			anchor = *e.Anchor
		}
		files = append(files, screenFile{Family: strings.TrimPrefix(e.Art, "screens/"), ID: e.ID, Density: d, Width: e.Width, Height: e.Height, Anchor: anchor,
			read: func() ([]byte, error) { return os.ReadFile(name) }})
	}
	if !seen["backdrop/western-front"] {
		return nil, fmt.Errorf("index has no backdrop/western-front at %dx", d)
	}
	paintings, err := filepath.Glob(filepath.Join(artDir, "screens", "*", "illustration.png"))
	if err != nil {
		return nil, err
	}
	slices.Sort(paintings)
	for _, src := range paintings {
		family := filepath.Base(filepath.Dir(src))
		files = append(files, screenFile{Family: family, ID: "scene/" + family, Density: d, Width: screenW * d, Height: screenH * d,
			read: func() ([]byte, error) { return scalePainting(src, screenW*d, screenH*d) }})
	}
	return files, nil
}

// scalePainting trims one pixel from each edge of the painting at src (the
// sources are 1586×992, a hair wider than 16:10) and scales it to w×h.
func scalePainting(src string, w, h int) ([]byte, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	im, err := png.Decode(io.LimitReader(f, 64<<20))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src, err)
	}
	b := im.Bounds()
	if b.Dx()-2 < w || b.Dy()-2 < h {
		return nil, fmt.Errorf("%s: %dx%d is smaller than %dx%d; no upscaling", src, b.Dx(), b.Dy(), w, h)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), im, b.Inset(1), draw.Src, nil)
	var out bytes.Buffer
	if err := png.Encode(&out, dst); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// buildScreens writes the screen catalog into the new directory out: one pack
// per family and density in out/<d>x/<family>/ (a frame and its painting share
// one), so each stays inside the loader's decoded-image budget. Every pack is
// loaded and fully validated before the build counts as done; a failed build
// leaves nothing behind.
func buildScreens(assetsDir, artDir, out string) ([]assets.Manifest, error) {
	entries, err := readIndex(assetsDir)
	if err != nil {
		return nil, err
	}
	plans := make([][]screenFile, len(screenDensities))
	for i, d := range screenDensities {
		if plans[i], err = planScreens(entries, assetsDir, artDir, d); err != nil {
			return nil, err
		}
	}
	o, err := newOutput(out)
	if err != nil {
		return nil, err
	}
	defer o.close()
	var written []assets.Manifest
	for i, d := range screenDensities {
		var families []string
		for _, f := range plans[i] {
			if !slices.Contains(families, f.Family) {
				families = append(families, f.Family)
			}
		}
		for _, family := range families {
			m, err := o.writeScreenPack(plans[i], d, family)
			if err != nil {
				return nil, err
			}
			written = append(written, m)
		}
	}
	o.done()
	return written, nil
}

// writeScreenPack writes the files of family at density d as the pack
// screen-<family>-<d>x into <d>x/<family>/ and validates it.
func (o *output) writeScreenPack(files []screenFile, d int, family string) (assets.Manifest, error) {
	dir := path.Join(fmt.Sprintf("%dx", d), family)
	m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("screen-%s-%dx", family, d), Status: "candidate"}
	if err := o.mkdirAll(path.Join(dir, "images")); err != nil {
		return m, err
	}
	for _, f := range files {
		if f.Family != family {
			continue
		}
		b, err := f.read()
		if err != nil {
			return m, fmt.Errorf("%s at %dx: %w", f.ID, d, err)
		}
		sum := sha256.Sum256(b)
		name := fmt.Sprintf("images/%04d.png", len(m.Entries))
		if err := o.writeNew(path.Join(dir, name), func(w io.Writer) error { _, err := w.Write(b); return err }); err != nil {
			return m, err
		}
		m.Entries = append(m.Entries, assets.Entry{ID: f.ID, Path: name, SHA256: hex.EncodeToString(sum[:]),
			Width: f.Width, Height: f.Height, Density: d, Anchor: f.Anchor})
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return m, err
	}
	if err := o.writeNew(path.Join(dir, "manifest.json"), func(w io.Writer) error {
		_, err := w.Write(append(b, '\n'))
		return err
	}); err != nil {
		return m, err
	}
	sub, err := fs.Sub(o.root.FS(), dir)
	if err != nil {
		return m, err
	}
	p, err := assets.Open(sub, "manifest.json", assets.Options{AllowCandidates: true})
	if err != nil {
		return m, err
	}
	return m, p.Validate()
}
