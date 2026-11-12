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
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"
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

// backdrop writes a painting's tiles and their pyramid, which the browser fetches only for its battle's month.
func (b blobWriter) backdrop(ctx context.Context, im image.Image) (FileRef, error) {
	p, err := b.pyramid(ctx, im)
	if err != nil {
		return FileRef{}, err
	}
	return b.json(p)
}

// Qualities of a painting's WebP tiles: the full-resolution level (about three
// quarters of the bytes) matches what the desktop game's JPEG shows, the
// smaller levels, seen zoomed out, are lighter.
const (
	backdropQuality    = 92
	backdropMipQuality = 85
	// backdropWorkers bounds the tiles encoded at once: each encoder instance holds a few MB.
	backdropWorkers = 4
)

func (b blobWriter) pyramid(ctx context.Context, im image.Image) (Pyramid, error) {
	p := Pyramid{Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), TileSize: 512}
	for scale := 1; ; scale *= 2 {
		m := Mip{Scale: scale, Width: im.Bounds().Dx(), Height: im.Bounds().Dy()}
		quality := backdropMipQuality
		if scale == 1 {
			quality = backdropQuality
		}
		for y := 0; y < m.Height; y += p.TileSize {
			for x := 0; x < m.Width; x += p.TileSize {
				t := Tile{X: x, Y: y, Width: min(p.TileSize, m.Width-x), Height: min(p.TileSize, m.Height-y)}
				// A two-pixel neighbour gutter prevents clamped texture filtering
				// from introducing seams between the independently loaded tiles.
				src := image.Rect(max(0, x-2), max(0, y-2), min(m.Width, x+t.Width+2), min(m.Height, y+t.Height+2))
				t.SourceX, t.SourceY, t.SourceWidth, t.SourceHeight = src.Min.X, src.Min.Y, src.Dx(), src.Dy()
				m.Tiles = append(m.Tiles, t)
			}
		}
		var g errgroup.Group
		g.SetLimit(backdropWorkers)
		for i := range m.Tiles {
			t := &m.Tiles[i]
			g.Go(func() error {
				if err := ctx.Err(); err != nil {
					return err
				}
				src := image.Rect(t.SourceX, t.SourceY, t.SourceX+t.SourceWidth, t.SourceY+t.SourceHeight)
				pic := image.NewNRGBA(image.Rect(0, 0, src.Dx(), src.Dy()))
				draw.Draw(pic, pic.Rect, im, src.Min.Add(im.Bounds().Min), draw.Src)
				var buf bytes.Buffer
				if err := webp.Encode(&buf, pic, webp.Options{Quality: quality, Method: 4}); err != nil {
					return err
				}
				var err error
				t.File, err = b.put("webp", buf.Bytes())
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return p, err
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
