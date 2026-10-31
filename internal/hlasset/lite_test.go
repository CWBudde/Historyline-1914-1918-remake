package hlasset

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

// writeCatalog writes one pack per manifest directory under dir.
func writeCatalog(t *testing.T, dir string, packs map[string]map[string]image.Image) map[string]assets.Manifest {
	t.Helper()
	out := map[string]assets.Manifest{}
	for d, images := range packs {
		m := assets.Manifest{Schema: 1, ID: path.Base(d), Status: "candidate"}
		if err := os.MkdirAll(filepath.Join(dir, d, "images"), 0o755); err != nil {
			t.Fatal(err)
		}
		for id, im := range images {
			var b bytes.Buffer
			if err := png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			name := "images/" + path.Base(id) + ".png"
			if err := os.WriteFile(filepath.Join(dir, d, name), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			m.Entries = append(m.Entries, assets.Entry{
				ID: id, Path: name, SHA256: hex.EncodeToString(sum[:]),
				Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), Density: 2, Anchor: [2]float64{3, 4},
			})
		}
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, d, "manifest.json"), b, 0o644); err != nil {
			t.Fatal(err)
		}
		out[d] = m
	}
	return out
}

// noise is a w×h image of random colours, alpha a everywhere but a clear
// top-left pixel when a < 255.
func noise(w, h int, a uint8) image.Image {
	r := rand.New(rand.NewPCG(1, 2))
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			im.SetNRGBA(x, y, color.NRGBA{uint8(r.IntN(256)), uint8(r.IntN(256)), uint8(r.IntN(256)), a})
		}
	}
	if a < 255 {
		im.SetNRGBA(0, 0, color.NRGBA{})
	}
	return im
}

func TestLiteCatalog(t *testing.T) {
	in, out := t.TempDir(), filepath.Join(t.TempDir(), "lite")
	tiny := image.NewNRGBA(image.Rect(0, 0, 8, 8)) // one colour: the PNG is smaller
	src := writeCatalog(t, in, map[string]map[string]image.Image{
		"2x/backdrop": {"screen/backdrop": noise(64, 40, 255)},
		"2x/button":   {"button/up": noise(32, 16, 200), "button/dot": tiny},
	})
	st, err := liteCatalog(in, out)
	if err != nil {
		t.Fatal(err)
	}
	if st.packs != 2 || st.jpeg != 1 || st.webp != 1 || st.png != 1 || st.after >= st.before {
		t.Errorf("stats %+v", st)
	}
	want := map[string]string{"screen/backdrop": ".jpg", "button/up": ".webp", "button/dot": ".png"}
	for d, m := range src {
		p, err := assets.Open(os.DirFS(filepath.Join(out, d)), "manifest.json", assets.Options{AllowCandidates: true})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Validate(); err != nil {
			t.Fatal(err)
		}
		got := p.Manifest()
		if got.ID != m.ID || got.Status != m.Status || len(got.Entries) != len(m.Entries) {
			t.Fatalf("%s: manifest %+v", d, got)
		}
		for i, e := range got.Entries {
			o := m.Entries[i]
			if e.ID != o.ID || e.Width != o.Width || e.Height != o.Height || e.Density != o.Density || e.Anchor != o.Anchor {
				t.Errorf("%s: entry %+v, was %+v", d, e, o)
			}
			if path.Ext(e.Path) != want[e.ID] {
				t.Errorf("%s is %s, want %s", e.ID, e.Path, want[e.ID])
			}
			im, err := p.Image(e.ID, e.Density)
			if err != nil {
				t.Fatal(err)
			}
			if a := color.NRGBAModel.Convert(im.At(0, 0)).(color.NRGBA).A; e.ID == "button/up" && a != 0 {
				t.Errorf("%s lost its alpha: %d", e.ID, a)
			}
		}
	}
	if _, err := liteCatalog(in, out); err == nil {
		t.Error("wrote over an existing directory")
	}
	if _, err := liteCatalog(t.TempDir(), filepath.Join(t.TempDir(), "x")); err == nil {
		t.Error("compressed a directory without packs")
	}
}
