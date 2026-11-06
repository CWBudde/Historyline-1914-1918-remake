package hlwebapp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// FileRef pins an exported file. Every path is relative to data/index.json.
type FileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

type blobWriter struct{ root string }

func (b blobWriter) put(ext string, raw []byte) (_ FileRef, err error) {
	if ext == "" || strings.IndexFunc(ext, func(r rune) bool { return (r < 'a' || r > 'z') && (r < '0' || r > '9') }) >= 0 {
		return FileRef{}, fmt.Errorf("invalid blob extension %q", ext)
	}
	sum := sha256.Sum256(raw)
	r := FileRef{SHA256: hex.EncodeToString(sum[:]), Bytes: len(raw)}
	r.Path = "blobs/" + r.SHA256 + "." + ext
	name := filepath.Join(b.root, filepath.FromSlash(r.Path))
	if old, err := os.ReadFile(name); err == nil && bytes.Equal(old, raw) {
		return r, nil
	}
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return r, err
	}
	// Readers only see complete hash-pinned files, including during an export.
	f, err := os.CreateTemp(filepath.Dir(name), ".blob-")
	if err != nil {
		return r, err
	}
	tmp := f.Name()
	defer func() { err = errors.Join(err, removeTemporary(tmp)) }()
	if _, err = f.Write(raw); err != nil {
		return r, errors.Join(err, f.Close())
	}
	if err = f.Close(); err != nil {
		return r, err
	}
	return r, os.Rename(tmp, name)
}

func removeTemporary(name string) error {
	err := os.Remove(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (b blobWriter) json(v any) (FileRef, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return FileRef{}, err
	}
	return b.put("json", raw)
}

func (b blobWriter) png(im image.Image) (FileRef, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, im); err != nil {
		return FileRef{}, err
	}
	return b.put("png", buf.Bytes())
}

// Pyramid coordinates are pixels in each mip; Scale maps mip pixels to source pixels.
type Pyramid struct {
	Width    int   `json:"width"`
	Height   int   `json:"height"`
	TileSize int   `json:"tileSize"`
	Mips     []Mip `json:"mips"`
}
type Mip struct {
	Scale  int    `json:"scale"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"`
}
type Tile struct {
	X            int     `json:"x"`
	Y            int     `json:"y"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	SourceX      int     `json:"sourceX"`
	SourceY      int     `json:"sourceY"`
	SourceWidth  int     `json:"sourceWidth"`
	SourceHeight int     `json:"sourceHeight"`
	File         FileRef `json:"file"`
}

func (b blobWriter) pyramid(ctx context.Context, im image.Image) (Pyramid, error) {
	p := Pyramid{Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), TileSize: 512}
	for scale := 1; ; scale *= 2 {
		m := Mip{Scale: scale, Width: im.Bounds().Dx(), Height: im.Bounds().Dy()}
		for y := 0; y < m.Height; y += p.TileSize {
			for x := 0; x < m.Width; x += p.TileSize {
				if err := ctx.Err(); err != nil {
					return p, err
				}
				t := Tile{X: x, Y: y, Width: min(p.TileSize, m.Width-x), Height: min(p.TileSize, m.Height-y)}
				// A two-pixel neighbour gutter prevents clamped texture filtering
				// from introducing seams between the independently loaded tiles.
				src := image.Rect(max(0, x-2), max(0, y-2), min(m.Width, x+t.Width+2), min(m.Height, y+t.Height+2))
				t.SourceX, t.SourceY, t.SourceWidth, t.SourceHeight = src.Min.X, src.Min.Y, src.Dx(), src.Dy()
				pic := image.NewRGBA(image.Rect(0, 0, src.Dx(), src.Dy()))
				draw.Copy(pic, image.Point{}, im, src.Add(im.Bounds().Min), draw.Src, nil)
				var buf bytes.Buffer
				if err := jpeg.Encode(&buf, pic, &jpeg.Options{Quality: 85}); err != nil {
					return p, err
				}
				var err error
				if t.File, err = b.put("jpg", buf.Bytes()); err != nil {
					return p, err
				}
				m.Tiles = append(m.Tiles, t)
			}
		}
		p.Mips = append(p.Mips, m)
		if m.Width <= p.TileSize && m.Height <= p.TileSize {
			break
		}
		next := image.NewRGBA(image.Rect(0, 0, (m.Width+1)/2, (m.Height+1)/2))
		draw.ApproxBiLinear.Scale(next, next.Bounds(), im, im.Bounds(), draw.Src, nil)
		im = next
	}
	return p, nil
}
