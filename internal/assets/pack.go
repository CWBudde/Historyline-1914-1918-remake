// Package assets loads bounded, hash-pinned image packs independently of rendering.
// Images are PNG, or JPEG and WebP in the web build's compressed packs. Packs are
// immutable snapshots; callers must treat returned images as read-only.
package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"math"
	"path"
	"strings"
	"sync"

	"golang.org/x/image/webp"
)

const (
	maxManifest = 2 << 20
	maxFile     = 32 << 20
	maxEncoded  = 128 << 20
)

// MaxDecodedBytes bounds each pack's decoded images at eight bytes per pixel.
const MaxDecodedBytes = 256 << 20

// MaxEntriesPerPack bounds the number of exact ID/density entries in one pack.
const MaxEntriesPerPack = 4096

// Manifest is versioned independently of authoring recipes and export reports.
// Paths are relative to the supplied filesystem root, not the manifest directory.
type Manifest struct {
	Schema  int     `json:"schema_version"`
	ID      string  `json:"id"`
	Status  string  `json:"status"` // candidate or approved; never inferred from an export
	Entries []Entry `json:"entries"`
}

// Entry describes one exact density of a stable visual ID. Width/Height are
// pixels; Anchor is in logical coordinates. No density fallback or scaling occurs.
type Entry struct {
	ID      string     `json:"id"`
	Path    string     `json:"path"`
	SHA256  string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Density int        `json:"density"`
	Anchor  [2]float64 `json:"logical_anchor"`
}

// formats are the image formats a pack may hold, by file extension.
var formats = map[string]struct {
	name   string
	config func(io.Reader) (image.Config, error)
	decode func(io.Reader) (image.Image, error)
}{
	".png":  {"PNG", png.DecodeConfig, png.Decode},
	".jpg":  {"JPEG", jpeg.DecodeConfig, jpeg.Decode},
	".webp": {"WebP", webp.DecodeConfig, webp.Decode},
}

type Options struct {
	AllowCandidates bool
}

type key struct {
	id      string
	density int
}

type sprite struct {
	entry Entry
	bytes []byte
	once  sync.Once
	image image.Image
	err   error
}

// Pack owns verified encoded bytes and a lazy, concurrency-safe decoded cache.
// A different Pack has a separate cache, so switching packs cannot reuse stale IDs.
type Pack struct {
	manifest Manifest
	sprites  map[key]*sprite
}

// MissingError identifies a missing exact ID/density pair. There are no silent
// placeholder images, reference-art fallbacks or invented directions.
type MissingError struct {
	PackID  string
	ID      string
	Density int
}

func (e *MissingError) Error() string {
	return fmt.Sprintf("pack %q: missing asset %q at %dx", e.PackID, e.ID, e.Density)
}

// ParseManifest rejects unknown fields and trailing JSON values.
func ParseManifest(b []byte) (Manifest, error) {
	var m Manifest
	if len(b) > maxManifest {
		return m, fmt.Errorf("manifest exceeds %d bytes", maxManifest)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, fmt.Errorf("manifest: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return m, fmt.Errorf("manifest must contain exactly one JSON value")
	}
	return m, nil
}

// Open reads a manifest then snapshots its images. Use os.Root.FS for disk packs
// to contain symlinks, or an embed.FS for distribution packs.
func Open(fsys fs.FS, name string, options Options) (*Pack, error) {
	if !safePath(name) {
		return nil, fmt.Errorf("invalid manifest path %q", name)
	}
	b, err := readBounded(fsys, name, maxManifest)
	if err != nil {
		return nil, err
	}
	m, err := ParseManifest(b)
	if err != nil {
		return nil, err
	}
	return Load(fsys, m, options)
}

// Load checks the entire catalog, hashes and image headers before returning any
// assets. Validate also fully decodes every image, including a PNG's pixel stream/CRC.
func Load(fsys fs.FS, m Manifest, options Options) (*Pack, error) {
	if m.Schema != 1 || !safePath(m.ID) {
		return nil, fmt.Errorf("unsupported pack schema or invalid ID %q", m.ID)
	}
	if m.Status != "approved" && m.Status != "candidate" {
		return nil, fmt.Errorf("pack %q: invalid status %q", m.ID, m.Status)
	}
	if m.Status == "candidate" && !options.AllowCandidates {
		return nil, fmt.Errorf("pack %q: candidate requires explicit opt-in", m.ID)
	}
	if len(m.Entries) == 0 || len(m.Entries) > MaxEntriesPerPack {
		return nil, fmt.Errorf("pack %q: need 1..%d entries", m.ID, MaxEntriesPerPack)
	}
	p := &Pack{manifest: m, sprites: make(map[key]*sprite)}
	p.manifest.Entries = append([]Entry(nil), m.Entries...)
	paths := make(map[string]bool)
	geometry := make(map[string]Entry)
	decoded := int64(0)
	// Check the complete catalog before file reads or image allocation.
	for _, e := range m.Entries {
		k := key{e.ID, e.Density}
		if _, ok := formats[path.Ext(e.Path)]; !ok || !safePath(e.ID) || !safePath(e.Path) {
			return nil, fmt.Errorf("asset %q: invalid image path or ID %q (PNG, JPEG or WebP)", e.ID, e.Path)
		}
		if _, ok := p.sprites[k]; ok {
			return nil, fmt.Errorf("duplicate asset %q at %dx", e.ID, e.Density)
		}
		if paths[strings.ToLower(e.Path)] {
			return nil, fmt.Errorf("duplicate asset path %q", e.Path)
		}
		paths[strings.ToLower(e.Path)] = true
		if e.Density != 1 && e.Density != 2 && e.Density != 4 {
			return nil, fmt.Errorf("asset %q: unsupported density %d", e.ID, e.Density)
		}
		if e.Width <= 0 || e.Height <= 0 || e.Width > 8192 || e.Height > 8192 ||
			e.Width%e.Density != 0 || e.Height%e.Density != 0 {
			return nil, fmt.Errorf("asset %q: invalid pixel dimensions", e.ID)
		}
		w, h := e.Width/e.Density, e.Height/e.Density
		if !finite(e.Anchor[0]) || !finite(e.Anchor[1]) || e.Anchor[0] < 0 || e.Anchor[1] < 0 || e.Anchor[0] > float64(w) || e.Anchor[1] > float64(h) {
			return nil, fmt.Errorf("asset %q: invalid logical anchor", e.ID)
		}
		if prior, ok := geometry[e.ID]; ok && (prior.Width/prior.Density != w || prior.Height/prior.Density != h || prior.Anchor != e.Anchor) {
			return nil, fmt.Errorf("asset %q: density geometry disagrees", e.ID)
		}
		geometry[e.ID] = e
		sum, err := hex.DecodeString(e.SHA256)
		if err != nil || len(sum) != sha256.Size || strings.ToLower(e.SHA256) != e.SHA256 {
			return nil, fmt.Errorf("asset %q: invalid SHA-256", e.ID)
		}
		// PNG decoding uses at most eight bytes per pixel for 16-bit RGBA.
		decoded += int64(e.Width) * int64(e.Height) * 8
		if decoded > MaxDecodedBytes {
			return nil, fmt.Errorf("pack %q: decoded budget exceeds %d bytes", m.ID, MaxDecodedBytes)
		}
		p.sprites[k] = &sprite{entry: e}
	}
	encoded := 0
	for _, e := range p.manifest.Entries {
		b, err := readBounded(fsys, e.Path, maxFile)
		if err != nil {
			return nil, fmt.Errorf("asset %q: %w", e.ID, err)
		}
		encoded += len(b)
		if encoded > maxEncoded {
			return nil, fmt.Errorf("pack %q: encoded budget exceeds %d bytes", m.ID, maxEncoded)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != e.SHA256 {
			return nil, fmt.Errorf("asset %q: hash mismatch: %s", e.ID, e.Path)
		}
		f := formats[path.Ext(e.Path)]
		cfg, err := f.config(bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("asset %q: %s header: %w", e.ID, f.name, err)
		}
		if cfg.Width != e.Width || cfg.Height != e.Height {
			return nil, fmt.Errorf("asset %q: bad dimensions for %s: got %dx%d, want %dx%d", e.ID, e.Path, cfg.Width, cfg.Height, e.Width, e.Height)
		}
		p.sprites[key{e.ID, e.Density}].bytes = b
	}
	return p, nil
}

func (p *Pack) Manifest() Manifest {
	m := p.manifest
	m.Entries = append([]Entry(nil), m.Entries...)
	return m
}

func (p *Pack) Entry(id string, density int) (Entry, error) {
	s, err := p.lookup(id, density)
	if err != nil {
		return Entry{}, err
	}
	return s.entry, nil
}

// WritePNG writes the exact verified snapshot of a PNG, without recompression.
func (p *Pack) WritePNG(w io.Writer, id string, density int) error {
	s, err := p.lookup(id, density)
	if err != nil {
		return err
	}
	if path.Ext(s.entry.Path) != ".png" {
		return fmt.Errorf("asset %q at %dx is not a PNG", id, density)
	}
	_, err = io.Copy(w, bytes.NewReader(s.bytes))
	return err
}

// Image decodes an exact variant at most once. Repeated calls share a read-only
// image; encoded bytes were snapshotted by Load, so disk edits cannot race it.
func (p *Pack) Image(id string, density int) (image.Image, error) {
	s, err := p.lookup(id, density)
	if err != nil {
		return nil, err
	}
	s.once.Do(func() {
		f := formats[path.Ext(s.entry.Path)]
		s.image, s.err = f.decode(bytes.NewReader(s.bytes))
		if s.err != nil {
			s.err = fmt.Errorf("asset %q at %dx: decode %s: %w", id, density, f.name, s.err)
		}
	})
	return s.image, s.err
}

func (p *Pack) lookup(id string, density int) (*sprite, error) {
	s, ok := p.sprites[key{id, density}]
	if !ok {
		return nil, &MissingError{PackID: p.manifest.ID, ID: id, Density: density}
	}
	return s, nil
}

// Validate fully decodes every declared image in manifest order for stable errors.
func (p *Pack) Validate() error {
	for _, e := range p.manifest.Entries {
		if _, err := p.Image(e.ID, e.Density); err != nil {
			return err
		}
	}
	return nil
}

func readBounded(fsys fs.FS, name string, limit int) ([]byte, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > limit {
		return nil, fmt.Errorf("%s exceeds %d bytes", name, limit)
	}
	return b, nil
}

func safePath(name string) bool {
	if name == "." || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:<>\"|?*") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if strings.EqualFold(part, "legacy") || strings.EqualFold(part, "original") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return false
		}
		for _, char := range part {
			if char < 32 {
				return false
			}
		}
		base, _, _ := strings.Cut(strings.ToUpper(part), ".")
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" ||
			(len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return false
		}
	}
	return true
}

func finite(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) }
