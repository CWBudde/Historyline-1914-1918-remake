package hlwebapp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

func TestBlobsPinBytesAndDeduplicate(t *testing.T) {
	b := blobWriter{root: t.TempDir()}
	data := []byte("selected bytes")
	r, err := b.put("txt", data)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if r.SHA256 != hex.EncodeToString(sum[:]) || r.Bytes != len(data) {
		t.Fatalf("unbound reference: %+v", r)
	}
	got, err := os.ReadFile(filepath.Join(b.root, r.Path))
	if err != nil || !bytes.Equal(data, got) {
		t.Fatalf("blob: %q, %v", got, err)
	}
	again, err := b.put("txt", data)
	if err != nil || again != r {
		t.Fatalf("not stable: %+v %v", again, err)
	}
	for _, ext := range []string{"../escape", "png/escape", ""} {
		if _, err := b.put(ext, data); err == nil {
			t.Fatalf("accepted extension %q", ext)
		}
	}
}

func TestExportCatalogKeepsDensitiesAndGeometry(t *testing.T) {
	src := t.TempDir()
	for _, d := range []int{2, 4} {
		dir := filepath.Join(src, fmt.Sprintf("%dx/test", d))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		im := image.NewNRGBA(image.Rect(0, 0, 8*d, 6*d))
		im.SetNRGBA(d, d, color.NRGBA{90, 100, 30, 180})
		var raw bytes.Buffer
		if err := png.Encode(&raw, im); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw.Bytes())
		m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("test/%d", d), Entries: []assets.Entry{{ID: "unit/test/cp/n-normal", Path: "image.png", SHA256: hex.EncodeToString(sum[:]), Width: 8 * d, Height: 6 * d, Density: d, Anchor: [2]float64{4, 3}}}}
		manifest, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		for name, data := range map[string][]byte{"manifest.json": manifest, "image.png": raw.Bytes()} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	b := blobWriter{root: t.TempDir()}
	refs, ids, err := b.catalog(context.Background(), src, "units")
	if err != nil {
		t.Fatal(err)
	}
	if !ids["unit/test/cp/n-normal"] || len(refs) != 3 {
		t.Fatalf("incomplete catalog: %+v", refs)
	}
	for _, d := range []int{1, 2, 4} {
		for _, r := range []FileRef{refs[d].Lite, refs[d].Full} {
			raw, err := os.ReadFile(filepath.Join(b.root, r.Path))
			if err != nil {
				t.Fatal(err)
			}
			var c ImageCatalog
			if err := json.Unmarshal(raw, &c); err != nil {
				t.Fatal(err)
			}
			if c.Schema != 1 || len(c.Entries) != 1 {
				t.Fatalf("catalog: %+v", c)
			}
			e := c.Entries[0]
			if e.Density != d || e.Width != 8*d || e.Height != 6*d || e.Anchor != [2]float64{4, 3} {
				t.Fatalf("geometry: %+v", e)
			}
			pic, err := os.ReadFile(filepath.Join(b.root, e.File.Path))
			if err != nil {
				t.Fatal(err)
			}
			if sum := sha256.Sum256(pic); hex.EncodeToString(sum[:]) != e.File.SHA256 {
				t.Fatal("hash differs")
			}
		}
	}
}

func TestBackdropPyramidCoversOddEdges(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 1027, 515))
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			im.SetRGBA(x, y, color.RGBA{80, 100, 120, 255})
		}
	}
	b := blobWriter{root: t.TempDir()}
	p, err := b.pyramid(context.Background(), im)
	if err != nil {
		t.Fatal(err)
	}
	if p.Width != 1027 || p.Height != 515 || p.TileSize != 512 || len(p.Mips) != 3 {
		t.Fatalf("pyramid: %+v", p)
	}
	for i, mip := range p.Mips {
		w, h := (1027+(1<<i)-1)>>i, (515+(1<<i)-1)>>i
		if mip.Width != w || mip.Height != h {
			t.Fatalf("mip %d: %dx%d", i, mip.Width, mip.Height)
		}
		covered := 0
		for _, tile := range mip.Tiles {
			raw, err := os.ReadFile(filepath.Join(b.root, tile.File.Path))
			if err != nil {
				t.Fatal(err)
			}
			pic, err := jpeg.Decode(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			if pic.Bounds().Dx() != tile.SourceWidth || pic.Bounds().Dy() != tile.SourceHeight || tile.X+tile.Width > w || tile.Y+tile.Height > h {
				t.Fatal("tile geometry")
			}
			if tile.SourceX != max(0, tile.X-2) || tile.SourceY != max(0, tile.Y-2) || tile.SourceX+tile.SourceWidth != min(w, tile.X+tile.Width+2) || tile.SourceY+tile.SourceHeight != min(h, tile.Y+tile.Height+2) {
				t.Fatal("missing neighbour gutter")
			}
			covered += tile.Width * tile.Height
		}
		if covered != w*h {
			t.Fatalf("mip %d: covered %d of %d", i, covered, w*h)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := b.pyramid(ctx, im); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestAssetsRejectsBeforeWriting(t *testing.T) {
	for _, args := range [][]string{{"unexpected"}, {"-bad"}} {
		dir := filepath.Join(t.TempDir(), "absent")
		var out bytes.Buffer
		if err := Assets(context.Background(), append([]string{"-out", dir}, args...), &out, &out); err == nil {
			t.Fatal("accepted invalid arguments")
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatal("wrote output")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	if err := Assets(ctx, nil, &out, &out); err == nil {
		t.Fatal("ignored cancellation")
	}
}
