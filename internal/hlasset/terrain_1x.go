package hlasset

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

// derivedDensity is the terrain density build-terrain derives from the graded
// 2x packs instead of reading it from the index. A map drawn at 1x effective
// zoom then shows the 2x detail filtered once over its whole area, not the
// bilinear samples of a 2x tile drawn at half size, which alias.
const derivedDensity = 1

// deriveTerrain writes out/1x/<group>/ next to every written 2x pack: the same
// entries in the same parts, each 1x texel the alpha-weighted area average of
// its 2×2 graded 2x texels in linear light (halve).
func deriveTerrain(out string, packs []terrainPack) ([]terrainPack, error) {
	root, err := os.OpenRoot(out)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	var derived []terrainPack
	for _, src := range packs {
		if src.Density != 2*derivedDensity {
			continue
		}
		sub, err := fs.Sub(root.FS(), terrainDir(src.Density, src.Group))
		if err != nil {
			return nil, err
		}
		p, err := assets.Open(sub, "manifest.json", assets.Options{AllowCandidates: true})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.Manifest.ID, err)
		}
		dir := terrainDir(derivedDensity, src.Group)
		if err := mkdirAllIn(root, path.Join(dir, "images")); err != nil {
			return nil, err
		}
		m := src.Manifest
		m.ID = strings.TrimSuffix(m.ID, fmt.Sprintf("-%dx", src.Density)) + fmt.Sprintf("-%dx", derivedDensity)
		m.Entries = make([]assets.Entry, len(src.Manifest.Entries))
		for i, e := range src.Manifest.Entries {
			im, err := p.Image(e.ID, e.Density)
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			if err := png.Encode(&buf, halve(im)); err != nil {
				return nil, err
			}
			e.Path = fmt.Sprintf("images/%04d.png", i)
			e.SHA256 = fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
			e.Width, e.Height, e.Density = e.Width/2, e.Height/2, derivedDensity
			if err := writeNew(root, path.Join(dir, e.Path), func(w io.Writer) error { _, err := w.Write(buf.Bytes()); return err }); err != nil {
				return nil, err
			}
			m.Entries[i] = e
		}
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := writeNew(root, path.Join(dir, "manifest.json"), func(w io.Writer) error { _, err := w.Write(append(b, '\n')); return err }); err != nil {
			return nil, err
		}
		derived = append(derived, terrainPack{Group: src.Group, Density: derivedDensity, Manifest: m})
	}
	return derived, nil
}

// mkdirAllIn creates the missing directories of the slash path dir in root.
func mkdirAllIn(root *os.Root, dir string) error {
	if dir == "." {
		return nil
	}
	if err := mkdirAllIn(root, path.Dir(dir)); err != nil {
		return err
	}
	if err := root.Mkdir(filepath.FromSlash(dir), 0o755); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// srgbLinear maps an 8-bit sRGB value to linear light.
var srgbLinear = func() (t [256]float64) {
	for v := range t {
		c := float64(v) / 255
		if c <= 0.04045 {
			t[v] = c / 12.92
		} else {
			t[v] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return t
}()

func srgbEncode(v float64) uint8 {
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	return uint8(math.Round(255 * math.Max(0, math.Min(1, v))))
}

// halve returns im at half its (even) size: each texel the area average of
// a 2×2 block, its colour weighted by alpha and averaged in linear light, so
// bright detail keeps its brightness and transparent texels lend no colour.
// Tile edges stay where the 2x tile put them: a block half covered becomes a
// half transparent texel.
func halve(im image.Image) *image.NRGBA {
	b := im.Bounds()
	src, ok := im.(*image.NRGBA)
	if !ok || b.Min != (image.Point{}) {
		src = image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(src, src.Rect, im, b.Min, draw.Src)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx()/2, b.Dy()/2))
	for y := range dst.Rect.Dy() {
		for x := range dst.Rect.Dx() {
			var r, g, bl, a float64
			for _, o := range [4]int{0, 4, src.Stride, src.Stride + 4} {
				i := src.PixOffset(2*x, 2*y) + o
				w := float64(src.Pix[i+3])
				r += srgbLinear[src.Pix[i]] * w
				g += srgbLinear[src.Pix[i+1]] * w
				bl += srgbLinear[src.Pix[i+2]] * w
				a += w
			}
			if a == 0 {
				continue
			}
			j := dst.PixOffset(x, y)
			dst.Pix[j], dst.Pix[j+1], dst.Pix[j+2] = srgbEncode(r/a), srgbEncode(g/a), srgbEncode(bl/a)
			dst.Pix[j+3] = uint8(math.Round(a / 4))
		}
	}
	return dst
}
