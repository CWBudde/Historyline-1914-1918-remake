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
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlasset"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
	"golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"
)

type QualityPair struct {
	Lite FileRef `json:"lite"`
	Full FileRef `json:"full"`
}
type ImageEntry struct {
	ID      string     `json:"id"`
	File    FileRef    `json:"file"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Density int        `json:"density"`
	Anchor  [2]float64 `json:"logicalAnchor"`
}
type ImageCatalog struct {
	Schema  int          `json:"schema"`
	Entries []ImageEntry `json:"entries"`
}

// catalog consumes the validated native packs, retaining their exact geometry.
// Only the unit catalogue lacks 1x; its 2x images supply the derived small pack.
func (b blobWriter) catalog(ctx context.Context, dir, name string) (_ map[int]QualityPair, _ map[string]bool, err error) {
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, nil, err
	}
	defer func() { err = errors.Join(err, r.Close()) }()
	byDensity := map[int][]assets.Manifest{}
	byDir := map[string]assets.Manifest{}
	err = fs.WalkDir(r.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "manifest.json" {
			return nil
		}
		sub, err := fs.Sub(r.FS(), path.Dir(p))
		if err != nil {
			return err
		}
		pack, err := assets.Open(sub, "manifest.json")
		if err != nil {
			return err
		}
		m := pack.Manifest()
		density := m.Entries[0].Density
		if !slices.Contains([]int{1, 2, 4}, density) {
			return fmt.Errorf("unsupported density %d", density)
		}
		byDensity[density] = append(byDensity[density], m)
		byDir[path.Dir(p)] = m
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	want := []int{1, 2, 4}
	if name == "screens" {
		want = []int{1, 2}
	}
	if name == "units" && len(byDensity[1]) == 0 {
		byDensity[1] = byDensity[2]
	}
	out := map[int]QualityPair{}
	ids := map[string]bool{}
	var previous map[string]ImageEntry
	for _, density := range want {
		if len(byDensity[density]) == 0 {
			return nil, nil, fmt.Errorf("%s: missing %dx packs", name, density)
		}
		full, lite := ImageCatalog{Schema: 1}, ImageCatalog{Schema: 1}
		seen := map[string]ImageEntry{}
		var dirs []string
		for p, m := range byDir {
			d := m.Entries[0].Density
			if d == density || name == "units" && density == 1 && d == 2 {
				dirs = append(dirs, p)
			}
		}
		slices.Sort(dirs)
		// An explicitly supplied 1x unit catalog takes precedence over deriving it.
		if name == "units" && density == 1 {
			hasOne := false
			for _, p := range dirs {
				if byDir[p].Entries[0].Density == 1 {
					hasOne = true
				}
			}
			if hasOne {
				dirs = slices.DeleteFunc(dirs, func(p string) bool { return byDir[p].Entries[0].Density != 1 })
			}
		}
		for _, p := range dirs {
			sub, err := fs.Sub(r.FS(), p)
			if err != nil {
				return nil, nil, err
			}
			pack, err := assets.Open(sub, "manifest.json")
			if err != nil {
				return nil, nil, err
			}
			m := pack.Manifest()
			start := len(full.Entries)
			full.Entries = append(full.Entries, make([]ImageEntry, len(m.Entries))...)
			lite.Entries = append(lite.Entries, make([]ImageEntry, len(m.Entries))...)
			var g errgroup.Group
			g.SetLimit(8)
			for _, e := range m.Entries {
				if _, ok := seen[e.ID]; ok {
					return nil, nil, fmt.Errorf("duplicate %s at %dx", e.ID, density)
				}
				seen[e.ID] = ImageEntry{Width: e.Width / e.Density, Height: e.Height / e.Density, Anchor: e.Anchor}
				ids[e.ID] = true
			}
			for i, e := range m.Entries {
				g.Go(func() error {
					if err := ctx.Err(); err != nil {
						return err
					}
					raw, err := fs.ReadFile(r.FS(), path.Join(p, e.Path))
					if err != nil {
						return err
					}
					if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != e.SHA256 {
						return fmt.Errorf("asset changed during export: %s", e.ID)
					}
					im, err := pack.Image(e.ID, e.Density)
					if err != nil {
						return err
					}
					ext := strings.TrimPrefix(path.Ext(e.Path), ".")
					entry := ImageEntry{ID: e.ID, Width: e.Width / e.Density * density, Height: e.Height / e.Density * density, Density: density, Anchor: e.Anchor}
					if e.Density != density {
						n := image.NewNRGBA(image.Rect(0, 0, entry.Width, entry.Height))
						draw.ApproxBiLinear.Scale(n, n.Bounds(), im, im.Bounds(), draw.Src, nil)
						im = n
						var buf bytes.Buffer
						if err := png.Encode(&buf, n); err != nil {
							return err
						}
						raw = buf.Bytes()
						ext = "png"
					}
					ref, err := b.fullArt(ext, im, raw)
					if err != nil {
						return err
					}
					entry.File = ref
					full.Entries[start+i] = entry
					if e.ID != worldmaterial.MeadowID && e.ID != worldmaterial.ScrubID {
						compressed, err := cachedLite(im, raw, hlasset.LiteWebP)
						if err != nil {
							return err
						}
						if len(compressed.Data) < len(raw) {
							entry.File, err = b.put(compressed.Ext, compressed.Data)
							if err != nil {
								return err
							}
						}
					}
					lite.Entries[start+i] = entry
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", p, err)
			}
		}
		if previous != nil {
			if len(previous) != len(seen) {
				return nil, nil, fmt.Errorf("%s: different density coverage", name)
			}
			for id, e := range seen {
				if old, ok := previous[id]; !ok || old != e {
					return nil, nil, fmt.Errorf("%s: inconsistent density geometry: %s", name, id)
				}
			}
		}
		previous = seen
		less := func(a, c ImageEntry) int { return strings.Compare(a.ID, c.ID) }
		slices.SortFunc(full.Entries, less)
		slices.SortFunc(lite.Entries, less)
		pair := QualityPair{}
		if pair.Full, err = b.json(full); err != nil {
			return nil, nil, err
		}
		if pair.Lite, err = b.json(lite); err != nil {
			return nil, nil, err
		}
		out[density] = pair
	}
	return out, ids, nil
}

type encodedImage struct {
	Ext  string
	Data []byte
}

// cachedLite compresses im for the browser, cached by its PNG bytes raw and the
// WebP quality it is encoded at (a JPEG for an opaque image has one quality).
func cachedLite(im image.Image, raw []byte, quality int) (_ encodedImage, err error) {
	h := sha256.New()
	h.Write(raw)
	h.Write([]byte{byte(quality)})
	sum := h.Sum(nil)
	cache, err := os.UserCacheDir()
	if err != nil {
		return encodedImage{}, err
	}
	dir := filepath.Join(cache, "historyline", "webapp-lite-v2")
	name := filepath.Join(dir, hex.EncodeToString(sum)+".json")
	if data, err := os.ReadFile(name); err == nil {
		var v encodedImage
		if json.Unmarshal(data, &v) == nil && slices.Contains([]string{"jpg", "webp"}, v.Ext) && len(v.Data) > 0 {
			return v, nil
		}
	}
	ext, data, err := hlasset.EncodeAt(im, quality)
	if err != nil {
		return encodedImage{}, err
	}
	v := encodedImage{strings.TrimPrefix(ext, "."), data}
	encoded, err := json.Marshal(v)
	if err != nil {
		return v, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return v, err
	}
	f, err := os.CreateTemp(dir, ".lite-")
	if err != nil {
		return v, err
	}
	defer func() { err = errors.Join(err, removeTemporary(f.Name())) }()
	if _, err := f.Write(encoded); err != nil {
		return v, errors.Join(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return v, err
	}
	return v, os.Rename(f.Name(), name)
}

// cachedLossless stores the lossless WebP twin of a PNG, cached by the PNG's
// bytes: an unchanged image is not encoded again by the next export. Bump the
// directory's version when hlasset.LosslessWebP changes.
func cachedLossless(im image.Image, raw []byte) (_ []byte, err error) {
	sum := sha256.Sum256(raw)
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cache, "historyline", "webapp-lossless-v1")
	name := filepath.Join(dir, hex.EncodeToString(sum[:])+".webp")
	if data, err := os.ReadFile(name); err == nil && len(data) > 0 {
		return data, nil
	}
	data, err := hlasset.EncodeLossless(im)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return data, err
	}
	f, err := os.CreateTemp(dir, ".lossless-")
	if err != nil {
		return data, err
	}
	defer func() { err = errors.Join(err, removeTemporary(f.Name())) }()
	if _, err := f.Write(data); err != nil {
		return data, errors.Join(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return data, err
	}
	return data, os.Rename(f.Name(), name)
}
