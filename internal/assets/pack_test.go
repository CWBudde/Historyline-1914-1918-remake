package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"math"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

func fixture(t *testing.T, c color.NRGBA) (fstest.MapFS, Manifest) {
	t.Helper()
	f := fstest.MapFS{}
	m := Manifest{Schema: 1, ID: "fixture", Status: "approved"}
	for _, density := range []int{1, 2, 4} {
		im := image.NewNRGBA(image.Rect(0, 0, 3*density, 2*density))
		for y := range im.Bounds().Dy() {
			for x := range im.Bounds().Dx() {
				im.SetNRGBA(x, y, c)
			}
		}
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		name := "sprite-" + string(rune('0'+density)) + ".png"
		f[name] = &fstest.MapFile{Data: b.Bytes()}
		sum := sha256.Sum256(b.Bytes())
		m.Entries = append(m.Entries, Entry{ID: "unit/fixture/cp/n-normal", Path: name, SHA256: hex.EncodeToString(sum[:]), Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), Density: density, Anchor: [2]float64{1.5, 2}})
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	f["manifest.json"] = &fstest.MapFile{Data: b}
	return f, m
}

func TestExactVariantsSnapshotCacheAndPackSwitch(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	f, m := fixture(t, red)
	p, err := Open(f, "manifest.json", Options{})
	if err != nil {
		t.Fatal(err)
	}
	// Mutating caller metadata and disk bytes must not affect a loaded snapshot.
	m.Entries[0].Anchor[0] = 99
	copyOfManifest := p.Manifest()
	copyOfManifest.Entries[0].ID = "changed"
	f[m.Entries[0].Path].Data = []byte("changed after loading")
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{1, 2, 4} {
		entry, err := p.Entry("unit/fixture/cp/n-normal", density)
		if err != nil || entry.Anchor != [2]float64{1.5, 2} || entry.Width/density != 3 {
			t.Fatalf("logical registration changed: %+v %v", entry, err)
		}
		im, err := p.Image(entry.ID, density)
		if err != nil || im.Bounds() != image.Rect(0, 0, 3*density, 2*density) {
			t.Fatalf("wrong exact image: %v %v", im, err)
		}
		if color.NRGBAModel.Convert(im.At(0, 0)) != red {
			t.Fatal("snapshot pixels changed")
		}
		var group sync.WaitGroup
		for range 16 {
			group.Go(func() {
				cached, err := p.Image(entry.ID, density)
				if err != nil || cached != im {
					t.Errorf("cache was not shared: %v", err)
				}
			})
		}
		group.Wait()
		var copied bytes.Buffer
		if err := p.WritePNG(&copied, entry.ID, density); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(copied.Bytes())
		if hex.EncodeToString(sum[:]) != entry.SHA256 {
			t.Fatal("copied bytes differ from pinned snapshot")
		}
	}
	for _, query := range []key{{"unit/fixture/cp/n-normal", 3}, {"missing", 1}, {"unit/fixture/cp/s-normal", 1}} {
		_, err := p.Image(query.id, query.density)
		var missing *MissingError
		if !errors.As(err, &missing) || missing.PackID != "fixture" || missing.ID != query.id || missing.Density != query.density {
			t.Fatalf("missing exact variant needs a diagnostic, got %v", err)
		}
	}
	greenFS, greenManifest := fixture(t, color.NRGBA{G: 255, A: 255})
	other, err := Load(greenFS, greenManifest, Options{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := p.Image("unit/fixture/cp/n-normal", 1)
	b, _ := other.Image("unit/fixture/cp/n-normal", 1)
	if a == b || a.At(0, 0) == b.At(0, 0) {
		t.Fatal("pack switch reused the prior pack's cache")
	}
}

func TestRejectMalformedCatalog(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Manifest)
		want string
	}{
		{"schema", func(m *Manifest) { m.Schema = 2 }, "schema"},
		{"empty", func(m *Manifest) { m.Entries = nil }, "entries"},
		{"too-many", func(m *Manifest) { m.Entries = make([]Entry, maxEntries+1) }, "entries"},
		{"status", func(m *Manifest) { m.Status = "accepted-ish" }, "status"},
		{"candidate", func(m *Manifest) { m.Status = "candidate" }, "opt-in"},
		{"duplicate", func(m *Manifest) { m.Entries = append(m.Entries, m.Entries[0]) }, "duplicate asset"},
		{"shared-path", func(m *Manifest) { m.Entries[1].Path = m.Entries[0].Path }, "duplicate asset path"},
		{"case-path", func(m *Manifest) { m.Entries[1].Path = strings.ToUpper(m.Entries[0].Path[:1]) + m.Entries[0].Path[1:] }, "duplicate asset path"},
		{"density", func(m *Manifest) { m.Entries[0].Density = 3 }, "density"},
		{"dimensions", func(m *Manifest) { m.Entries[0].Height = 0 }, "dimensions"},
		{"huge", func(m *Manifest) { m.Entries[0].Width = 9000 }, "dimensions"},
		{"fractional-logical", func(m *Manifest) { m.Entries[1].Width = 7 }, "dimensions"},
		{"density-size", func(m *Manifest) { m.Entries[1].Width = 8 }, "geometry"},
		{"density-anchor", func(m *Manifest) { m.Entries[1].Anchor[0] = 2 }, "geometry"},
		{"negative-anchor", func(m *Manifest) { m.Entries[0].Anchor[0] = -1 }, "anchor"},
		{"outside-anchor", func(m *Manifest) { m.Entries[0].Anchor[0] = 4 }, "anchor"},
		{"nan-anchor", func(m *Manifest) { m.Entries[0].Anchor[0] = math.NaN() }, "anchor"},
		{"infinite-anchor", func(m *Manifest) { m.Entries[0].Anchor[1] = math.Inf(1) }, "anchor"},
		{"short-hash", func(m *Manifest) { m.Entries[0].SHA256 = "abcdef" }, "SHA-256"},
		{"nonhex-hash", func(m *Manifest) { m.Entries[0].SHA256 = strings.Repeat("z", 64) }, "SHA-256"},
		{"uppercase-hash", func(m *Manifest) { m.Entries[0].SHA256 = strings.ToUpper(m.Entries[0].SHA256) }, "SHA-256"},
		{"budget", func(m *Manifest) { m.Entries = m.Entries[:1]; m.Entries[0].Width = 8192; m.Entries[0].Height = 8192 }, "budget"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, m := fixture(t, color.NRGBA{})
			c.edit(&m)
			if _, err := Load(f, m, Options{}); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q, got %v", c.want, err)
			}
		})
	}
	for _, bad := range []string{"../escape.png", "/absolute.png", "C:/drive.png", `folder\escape.png`, "original/a.png", "Legacy/a.png", "a/../b.png", "a//b.png", ".", "a?/b.png", "a./b.png", "a /b.png", "CON.png", "com1/a.png", "a\x00.png"} {
		t.Run(bad, func(t *testing.T) {
			f, m := fixture(t, color.NRGBA{})
			m.Entries[0].Path = bad
			if _, err := Load(f, m, Options{}); err == nil {
				t.Fatal("accepted nonportable/reference path")
			}
			if _, err := Open(f, bad, Options{}); err == nil {
				t.Fatal("accepted invalid manifest path")
			}
		})
	}
}

func TestByteIntegrityAndFullPNGDecode(t *testing.T) {
	for _, mode := range []string{"missing", "hash", "header", "dimensions", "pixel-crc"} {
		t.Run(mode, func(t *testing.T) {
			f, m := fixture(t, color.NRGBA{A: 255})
			e := &m.Entries[0]
			b := f[e.Path].Data
			switch mode {
			case "missing":
				delete(f, e.Path)
			case "hash":
				f[e.Path].Data = append(b, 'x')
			case "header":
				f[e.Path].Data = []byte("not PNG")
			case "dimensions":
				m.Entries = m.Entries[:1]
				e.Width++
			case "pixel-crc":
				b = bytes.Clone(b)
				idat := bytes.Index(b, []byte("IDAT"))
				if idat < 0 {
					t.Fatal("fixture has no pixel stream")
				}
				b[idat+4] ^= 1
				f[e.Path].Data = b
			}
			if mode == "header" || mode == "pixel-crc" {
				sum := sha256.Sum256(f[e.Path].Data)
				e.SHA256 = hex.EncodeToString(sum[:])
			}
			p, err := Load(f, m, Options{})
			if mode != "pixel-crc" {
				if err == nil {
					t.Fatal("invalid bytes accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("header inspection should succeed before full decode", err)
			}
			if err := p.Validate(); err == nil {
				t.Fatal("full validation accepted corrupt pixel stream")
			}
			if _, err := p.Image(e.ID, e.Density); err == nil {
				t.Fatal("cached decode error lost")
			}
		})
	}
}

func TestStrictManifestAndCandidateOptIn(t *testing.T) {
	for _, b := range []string{`{}`, `{"extra":true}`, `null null`, `[]`, strings.Repeat(" ", maxManifest+1)} {
		m, err := ParseManifest([]byte(b))
		if err == nil {
			_, err = Load(fstest.MapFS{}, m, Options{})
		}
		if err == nil {
			t.Fatalf("invalid manifest accepted: %.40s", b)
		}
	}
	f, m := fixture(t, color.NRGBA{})
	m.Status = "candidate"
	if _, err := Load(f, m, Options{AllowCandidates: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(fstest.MapFS{}, "manifest.json", Options{}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("missing manifest error lost", err)
	}
}

func TestConcurrentFirstDecode(t *testing.T) {
	f, m := fixture(t, color.NRGBA{B: 255, A: 255})
	p, err := Load(f, m, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	results := make(chan image.Image, 32)
	for range 32 {
		group.Go(func() {
			im, err := p.Image(m.Entries[0].ID, 1)
			if err != nil {
				t.Error(err)
			}
			results <- im
		})
	}
	group.Wait()
	close(results)
	first := <-results
	for im := range results {
		if im != first {
			t.Fatal("concurrent cache misses decoded different images")
		}
	}
}
