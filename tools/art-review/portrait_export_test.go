package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPortraitActualBundlePinsRepeatAndGeometry(t *testing.T) {
	path, manifest, source := portraitTestFixture(t, 1024, 1024)
	outs := []string{filepath.Join(t.TempDir(), "first"), filepath.Join(t.TempDir(), "second")}
	for _, out := range outs {
		if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	files, err := os.ReadDir(outs[0])
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name())
		if f.IsDir() || !bytes.Equal(unitTestRead(t, filepath.Join(outs[0], f.Name())), unitTestRead(t, filepath.Join(outs[1], f.Name()))) {
			t.Fatalf("nondeterministic or nested artifact: %s", f.Name())
		}
	}
	wantNames := []string{"index.html", "infantry-1x.png", "infantry-2x.png", "infantry-4x.png", "infantry-source.png", "manifest.json", "report.json"}
	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("files: %v", names)
	}
	if !bytes.Equal(unitTestRead(t, filepath.Join(outs[0], "manifest.json")), manifest) || !bytes.Equal(unitTestRead(t, filepath.Join(outs[0], "infantry-source.png")), source) {
		t.Fatal("input snapshots were reencoded")
	}
	var report portraitReport
	if err := json.Unmarshal(unitTestRead(t, filepath.Join(outs[0], "report.json")), &report); err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 1 || report.Profile != "portrait-infantry-v1" || report.Family != "infantry" || !reflect.DeepEqual(report.Classes, []string{"CP_infantry", "EC_infantry"}) || report.LogicalDimensions != (dimensions{Width: 210, Height: 188}) || report.CropRectangle != (packCrop{X: 39, Y: 89, Width: 945, Height: 846}) {
		t.Fatalf("incorrect portrait identity/geometry: %+v", report)
	}
	if report.Source.Source != "source.png" || report.Source.RetainedSource != "infantry-source.png" || report.Source.Dimensions != (dimensions{Width: 1024, Height: 1024}) || report.ManifestSnapshot != "manifest.json" || len(report.Outputs) != 3 || report.CandidatePNGCount != 3 || report.PhysicalPNGCount != 4 || report.TotalFileCount != 7 {
		t.Fatalf("incorrect provenance/counts: %+v", report)
	}
	checkHash := func(name, pin string) {
		t.Helper()
		sum := sha256.Sum256(unitTestRead(t, filepath.Join(outs[0], name)))
		if hex.EncodeToString(sum[:]) != pin {
			t.Fatalf("artifact pin mismatch: %s", name)
		}
	}
	checkHash(report.ManifestSnapshot, report.ManifestSHA256)
	checkHash(report.Source.RetainedSource, report.Source.SourceSHA256)
	gallery := string(unitTestRead(t, filepath.Join(outs[0], "index.html")))
	for _, name := range wantNames[1:] {
		if !strings.Contains(gallery, `href="`+name+`"`) {
			t.Errorf("gallery missing %s link", name)
		}
	}
	if !strings.Contains(gallery, "image-rendering:auto") {
		t.Fatal("portrait gallery lacks continuous rendering")
	}
	for _, output := range report.Outputs {
		checkHash(output.Path, output.SHA256)
		im := unitTestLoad(t, filepath.Join(outs[0], output.Path))
		if im.Bounds().Dx() != output.Dimensions.Width || im.Bounds().Dy() != output.Dimensions.Height {
			t.Fatal("export dimensions disagree with report")
		}
		for i := 3; i < len(im.Pix); i += 4 {
			if im.Pix[i] != 255 {
				t.Fatal("export lost opacity")
			}
		}
	}
	// This scalar oracle independently evaluates only selected pixels, recursively
	// including intermediate byte rounding. It never calls the shared resampler.
	src := unitTestLoad(t, filepath.Join(outs[0], "infantry-source.png"))
	oracle := portraitTestOracle(src)
	for stage, density := range []int{4, 2, 1} {
		im := unitTestLoad(t, filepath.Join(outs[0], fmt.Sprintf("infantry-%dx.png", density)))
		if im.Bounds() != image.Rect(0, 0, 210*density, 188*density) {
			t.Fatalf("incorrect density %d", density)
		}
		for _, p := range []image.Point{{0, 0}, {im.Bounds().Dx() - 1, im.Bounds().Dy() - 1}, {17, 23}, {im.Bounds().Dx() / 2, im.Bounds().Dy() / 2}} {
			if got, want := im.NRGBAAt(p.X, p.Y), oracle(stage+1, p.X, p.Y); got != want {
				t.Fatalf("independent scalar crop/Lanczos oracle density %d at %v: got %v want %v", density, p, got, want)
			}
		}
	}
}

func portraitTestOracle(src *image.NRGBA) func(int, int, int) color.NRGBA {
	type key struct{ stage, x, y int }
	cache := map[key]color.NRGBA{}
	sizes := [][2]int{{945, 846}, {840, 752}, {420, 376}, {210, 188}}
	var pixel func(int, int, int) color.NRGBA
	pixel = func(stage, x, y int) color.NRGBA {
		if stage == 0 {
			return src.NRGBAAt(39+x, 89+y)
		}
		k := key{stage, x, y}
		if value, ok := cache[k]; ok {
			return value
		}
		sx := float64(sizes[stage-1][0]) / float64(sizes[stage][0])
		sy := float64(sizes[stage-1][1]) / float64(sizes[stage][1])
		cx, cy := (float64(x)+.5)*sx-.5, (float64(y)+.5)*sy-.5
		kernel := func(u float64) float64 {
			u = math.Abs(u)
			if u == 0 {
				return 1
			}
			if u >= 3 || u == 1 || u == 2 {
				return 0
			}
			v := math.Pi * u
			return 3 * math.Sin(v) * math.Sin(v/3) / (v * v)
		}
		var total, r, g, b float64
		for j := max(0, int(math.Ceil(cy-3*sy))); j <= min(sizes[stage-1][1]-1, int(math.Floor(cy+3*sy))); j++ {
			for i := max(0, int(math.Ceil(cx-3*sx))); i <= min(sizes[stage-1][0]-1, int(math.Floor(cx+3*sx))); i++ {
				weight := kernel((float64(i)-cx)/sx) * kernel((float64(j)-cy)/sy)
				p := pixel(stage-1, i, j)
				total += weight
				r += weight * float64(p.R)
				g += weight * float64(p.G)
				b += weight * float64(p.B)
			}
		}
		channel := func(v float64) uint8 { return uint8(math.Floor(min(255, max(0, v/total)) + .5)) }
		out := color.NRGBA{channel(r), channel(g), channel(b), 255}
		cache[k] = out
		return out
	}
	return pixel
}

func TestPortraitLoadedBytesSnapshots(t *testing.T) {
	path, manifest, source := portraitTestFixture(t, 840, 752)
	in, err := loadPortraitInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	unitTestWrite(t, path, []byte("changed manifest"))
	unitTestWrite(t, filepath.Join(filepath.Dir(path), "source.png"), []byte("changed source"))
	b, err := buildPortraitBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	snapshots := map[string][]byte{}
	for _, a := range b.artifacts {
		snapshots[a.name] = a.data
	}
	if !bytes.Equal(snapshots["manifest.json"], manifest) || !bytes.Equal(snapshots["infantry-source.png"], source) {
		t.Fatal("snapshot reread changed input")
	}
}

func TestPortraitCropBeforeFilterExcludesOverscanBleed(t *testing.T) {
	path, _, _ := portraitTestFixture(t, 1024, 1025)
	im := image.NewNRGBA(image.Rect(0, 0, 1024, 1025))
	art := color.NRGBA{37, 91, 123, 255}
	for y := range 1025 {
		for x := range 1024 {
			p := color.NRGBA{255, 0, 255, 255}
			if x >= 39 && x < 984 && y >= 89 && y < 935 {
				p = art
			}
			im.SetNRGBA(x, y, p)
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, im); err != nil {
		t.Fatal(err)
	}
	portraitTestReplaceSource(t, path, encoded.Bytes())
	out := filepath.Join(t.TempDir(), "bundle")
	if err := runPortrait(path, out, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{1, 2, 4} {
		got := unitTestLoad(t, filepath.Join(out, fmt.Sprintf("infantry-%dx.png", density)))
		for y := range got.Bounds().Dy() {
			for x := range got.Bounds().Dx() {
				if p := got.NRGBAAt(x, y); p != art {
					t.Fatalf("discarded overscan bled into %dx export at (%d,%d): %v", density, x, y, p)
				}
			}
		}
	}
}

func TestPortraitExclusiveOutputAndRollback(t *testing.T) {
	path, _, _ := portraitTestFixture(t, 840, 752)
	dir := t.TempDir()
	for _, kind := range []string{"file", "dir", "symlink"} {
		out := filepath.Join(dir, kind)
		switch kind {
		case "dir":
			if err := os.Mkdir(out, 0o700); err != nil {
				t.Fatal(err)
			}
		case "symlink":
			if err := os.Symlink(filepath.Join(dir, "missing"), out); err != nil {
				t.Fatal(err)
			}
		default:
			unitTestWrite(t, out, []byte("preserve"))
		}
		before, err := os.Lstat(out)
		if err != nil {
			t.Fatal(err)
		}
		if runPortrait(path, out, io.Discard) == nil {
			t.Fatalf("existing %s accepted", kind)
		}
		after, err := os.Lstat(out)
		if err != nil || !os.SameFile(before, after) {
			t.Fatalf("existing %s changed", kind)
		}
	}
	failed := filepath.Join(dir, "failed")
	if runPortrait(path, failed, unitTestFailWriter{}) == nil {
		t.Fatal("log failure ignored")
	}
	if _, err := os.Lstat(failed); !os.IsNotExist(err) {
		t.Fatal("failed output retained")
	}
	replaced := filepath.Join(dir, "replaced")
	if err := runPortrait(path, replaced, unitTestReplacingWriter{t: t, out: replaced}); err == nil || !strings.Contains(err.Error(), "identity changed") {
		t.Fatalf("replacement identity not reported: %v", err)
	}
	if got := string(unitTestRead(t, filepath.Join(replaced, "foreign-marker"))); got != "preserve" {
		t.Fatal("foreign replacement removed")
	}
}
