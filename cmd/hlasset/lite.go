package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
	"github.com/gen2brain/webp"
	"golang.org/x/sync/errgroup"
)

// Qualities of the compressed catalogs.
const liteJPEG, liteWebP = 85, 80

// liteStats counts what liteCatalog wrote.
type liteStats struct {
	packs, jpeg, webp, png int
	before, after          int64 // bytes of the images
}

// liteCatalog writes a compressed copy of the pack catalog in dir, every
// <path>/manifest.json under it, to the new directory out, for the web build:
// an opaque image as a JPEG, any other as a lossy WebP with its alpha, and the
// PNG where that is smaller anyway. Periodic world materials always retain
// lossless PNG pixels. IDs, sizes and anchors stay.
func liteCatalog(dir, out string) (liteStats, error) {
	var st liteStats
	r, err := os.OpenRoot(dir)
	if err != nil {
		return st, err
	}
	defer func() { _ = r.Close() }()
	var manifests []string
	if err := fs.WalkDir(r.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == "manifest.json" {
			manifests = append(manifests, path.Dir(p))
		}
		return err
	}); err != nil {
		return st, err
	}
	if len(manifests) == 0 {
		return st, fmt.Errorf("%s has no pack manifests", dir)
	}
	o, err := newOutput(out)
	if err != nil {
		return st, err
	}
	defer o.close()
	for _, d := range manifests {
		sub, err := fs.Sub(r.FS(), d)
		if err != nil {
			return st, err
		}
		p, err := assets.Open(sub, "manifest.json", assets.Options{AllowCandidates: true})
		if err != nil {
			return st, fmt.Errorf("%s: %w", d, err)
		}
		if err := o.writeLitePack(d, sub, p, &st); err != nil {
			return st, fmt.Errorf("%s: %w", d, err)
		}
		st.packs++
	}
	o.done()
	return st, nil
}

// writeLitePack writes the compressed copy of pack p, read from src, under dir.
func (o *output) writeLitePack(dir string, src fs.FS, p *assets.Pack, st *liteStats) error {
	m := p.Manifest()
	type file struct {
		ext  string
		data []byte
	}
	files := make([]file, len(m.Entries))
	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())
	for i, e := range m.Entries {
		g.Go(func() error {
			orig, err := fs.ReadFile(src, e.Path)
			if err != nil {
				return err
			}
			im, err := p.Image(e.ID, e.Density)
			if err != nil {
				return err
			}
			var ext string
			var b []byte
			if e.ID == worldmaterial.MeadowID || e.ID == worldmaterial.ScrubID {
				var buf bytes.Buffer
				err = png.Encode(&buf, im)
				ext, b = ".png", buf.Bytes()
			} else {
				ext, b, err = liteEncode(im)
			}
			if err != nil {
				return fmt.Errorf("%s at %dx: %w", e.ID, e.Density, err)
			}
			if len(b) >= len(orig) && path.Ext(e.Path) == ".png" {
				ext, b = ".png", orig
			}
			files[i] = file{ext, b}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	if err := o.mkdirAll(path.Join(dir, "images")); err != nil {
		return err
	}
	for i := range m.Entries {
		e, f := &m.Entries[i], files[i]
		info, err := fs.Stat(src, e.Path)
		if err != nil {
			return err
		}
		st.before += info.Size()
		st.after += int64(len(f.data))
		switch f.ext {
		case ".jpg":
			st.jpeg++
		case ".webp":
			st.webp++
		default:
			st.png++
		}
		e.Path = fmt.Sprintf("images/%04d%s", i, f.ext)
		sum := sha256.Sum256(f.data)
		e.SHA256 = hex.EncodeToString(sum[:])
		if err := o.writeNew(path.Join(dir, e.Path), func(w io.Writer) error { _, err := w.Write(f.data); return err }); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return o.writeNew(path.Join(dir, "manifest.json"), func(w io.Writer) error {
		_, err := w.Write(append(b, '\n'))
		return err
	})
}

// liteEncode compresses im: a JPEG when it is opaque, else a lossy WebP.
func liteEncode(im image.Image) (ext string, b []byte, err error) {
	n := image.NewNRGBA(im.Bounds())
	draw.Draw(n, n.Rect, im, im.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if opaque(n) {
		err = jpeg.Encode(&buf, n, &jpeg.Options{Quality: liteJPEG})
		return ".jpg", buf.Bytes(), err
	}
	err = webp.Encode(&buf, n, webp.Options{Quality: liteWebP, Method: 4})
	return ".webp", buf.Bytes(), err
}

func opaque(n *image.NRGBA) bool {
	for y := range n.Rect.Dy() {
		row := n.Pix[y*n.Stride : y*n.Stride+4*n.Rect.Dx()]
		for i := 3; i < len(row); i += 4 {
			if row[i] != 255 {
				return false
			}
		}
	}
	return true
}

func (s liteStats) String() string {
	var parts []string
	for _, c := range []struct {
		n    int
		name string
	}{{s.jpeg, "JPEG"}, {s.webp, "WebP"}, {s.png, "PNG"}} {
		if c.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c.n, c.name))
		}
	}
	return fmt.Sprintf("%d packs (%s): %.1f MB, from %.1f MB", s.packs, strings.Join(parts, ", "), float64(s.after)/1e6, float64(s.before)/1e6)
}
