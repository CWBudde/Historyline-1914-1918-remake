package main

import (
	"bytes"
	"encoding/json"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetainedBundlePinsGeometryAndRepeat(t *testing.T) {
	root := filepath.Join("..", "..", "art", "exports", "interface-v1")
	raw, e := os.ReadFile(filepath.Join(root, "report.json"))
	if e != nil {
		t.Fatal(e)
	}
	var report Report
	if e = json.Unmarshal(raw, &report); e != nil {
		t.Fatal(e)
	}
	pngs, svgs := 0, 0
	ids := map[string]int{}
	ranks := map[string]bool{}
	for _, f := range report.Files {
		data, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if e != nil {
			t.Fatal(e)
		}
		if hash(data) != f.SHA256 {
			t.Fatalf("bad saved pin %s", f.Path)
		}
		if strings.HasSuffix(f.Path, ".svg") {
			svgs++
		}
		if f.Density == 0 {
			continue
		}
		pngs++
		ids[f.ID]++
		im, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			t.Fatal(e)
		}
		if im.Bounds().Dx() != f.Width || im.Bounds().Dy() != f.Height {
			t.Fatalf("wrong saved geometry %s", f.Path)
		}
		for y := 0; y < f.Height; y++ {
			for x := 0; x < f.Width; x++ {
				p := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if p.A == 0 && (p.R != 0 || p.G != 0 || p.B != 0) {
					t.Fatalf("hidden RGB %s", f.Path)
				}
			}
		}
		if f.Density == 1 && strings.HasPrefix(f.ID, "rank/") {
			if ranks[f.SHA256] {
				t.Fatal("indistinguishable experience levels")
			}
			ranks[f.SHA256] = true
		}
	}
	if pngs != 351 || svgs != 93 || len(ids) != 117 || len(ranks) != 11 {
		t.Fatalf("incomplete version1 bundle: %d PNGs, %d SVGs, %d IDs, %d ranks", pngs, svgs, len(ids), len(ranks))
	}
	for id, n := range ids {
		if n != 3 {
			t.Fatalf("incomplete densities for %s", id)
		}
	}
	out := filepath.Join(t.TempDir(), "repeat")
	if e = export(filepath.Join("..", "..", "art", "interface-v1.json"), out); e != nil {
		t.Fatal(e)
	}
	count := 0
	if e = filepath.WalkDir(root, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if entry.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		want, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		got, e := os.ReadFile(filepath.Join(out, rel))
		if e != nil {
			return e
		}
		if !bytes.Equal(want, got) {
			t.Errorf("repeat differs %s", rel)
		}
		count++
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if count != len(report.Files)+1 {
		t.Fatalf("unexpected unpinned files: %d", count)
	}
	if e = export(filepath.Join("..", "..", "art", "interface-v1.json"), out); e == nil {
		t.Fatal("accepted existing output")
	}
	unchanged, e := os.ReadFile(filepath.Join(out, "report.json"))
	if e != nil || !bytes.Equal(raw, unchanged) {
		t.Fatal("existing output modified")
	}
	t.Logf("verified %d byte-identical build files, %d PNG exports and %d SVG masters", count, pngs, svgs)
}

func TestGalleryLocalReferencesExist(t *testing.T) {
	// Check the actual emitted static element links.
	root := filepath.Join("..", "..", "art", "exports", "interface-v1")
	raw, e := os.ReadFile(filepath.Join(root, "index.html"))
	if e != nil {
		t.Fatal(e)
	}
	s := string(raw)
	for _, prefix := range []string{`src="`, `href="`} {
		rest := s
		for {
			start := strings.Index(rest, prefix)
			if start < 0 {
				break
			}
			rest = rest[start+len(prefix):]
			end := strings.IndexByte(rest, '"')
			if end < 0 {
				t.Fatal("unterminated link")
			}
			path := rest[:end]
			rest = rest[end+1:]
			if strings.HasPrefix(path, "#") || strings.Contains(path, ":") {
				continue
			}
			if _, e = os.Stat(filepath.Join(root, filepath.FromSlash(path))); e != nil {
				t.Fatalf("broken link %s", path)
			}
		}
	}
}
