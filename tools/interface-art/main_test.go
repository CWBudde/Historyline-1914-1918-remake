package interfaceart

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWeaponFitPreservesWholeSubjectAndAspect(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 1600, 700))
	for y := 100; y < 600; y++ {
		for x := 300; x < 1300; x++ {
			im.SetNRGBA(x, y, color.NRGBA{173, 95, 42, 255})
		}
	}
	got, e := fitWeapon(im, [4]int{290, 90, 1020, 520})
	if e != nil {
		t.Fatal(e)
	}
	if got.Bounds() != image.Rect(0, 0, 1008, 200) {
		t.Fatal(got.Bounds())
	}
	var visible image.Rectangle
	for y := 0; y < 200; y++ {
		for x := 0; x < 1008; x++ {
			if got.NRGBAAt(x, y).A > 127 {
				visible = visible.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if visible.Dx() < 366 || visible.Dx() > 372 || visible.Dy() < 182 || visible.Dy() > 188 {
		t.Fatalf("subject stretched or clipped: %v", visible)
	}
	if _, e = fitWeapon(im, [4]int{310, 110, 900, 400}); e == nil {
		t.Fatal("accepted amputated subject")
	}
	tiny := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	tiny.SetNRGBA(20, 20, color.NRGBA{255, 255, 255, 255})
	if _, e = fitWeapon(tiny, [4]int{0, 0, 40, 40}); e == nil {
		t.Fatal("accepted upscaling")
	}
}

func TestTransparentMarkersAndFactoryStayInsideCanvas(t *testing.T) {
	for _, a := range catalog() {
		if a.Family != "map" && a.Family != "ownership" {
			continue
		}
		t.Run(a.ID, func(t *testing.T) {
			for _, s := range a.Shapes {
				for _, c := range s.Commands {
					for i := 0; i < len(c.Points); i += 2 {
						x, y := c.Points[i], c.Points[i+1]
						if x < 0 || y < 0 || x > float64(a.Width) || y > float64(a.Height) {
							t.Fatalf("clipped path (%g,%g)", x, y)
						}
					}
				}
			}
			im := a.raster(4)
			if a.Family == "map" && a.ID != "map/reachable" && a.ID != "map/attackable" && a.ID != "map/path" && a.ID != "map/from" && a.ID != "map/attack" && a.ID != "map/invalid" {
				if im.NRGBAAt(240, 200).A != 0 {
					t.Fatal("marker conceals underlying unit")
				}
			}
		})
	}
}

func TestAnimationFramesResolveAndPlaybackIsFinite(t *testing.T) {
	known := map[string]bool{}
	for _, a := range catalog() {
		known[a.ID] = true
	}
	for i := 0; i < 9; i++ {
		known["effect/explosion-0"+string(rune('0'+i))] = true
	}
	for _, a := range animations() {
		if a.Loop || !a.Skippable || len(a.Frames) != len(a.FrameMS) || len(a.ReducedFrames) != len(a.ReducedMS) || len(a.ReducedFrames) != 1 {
			t.Fatalf("unbounded or inaccessible track: %+v", a)
		}
		total := 0
		for i, id := range a.Frames {
			if !known[id] || a.FrameMS[i] <= 0 {
				t.Fatalf("invalid frame: %s", id)
			}
			total += a.FrameMS[i]
		}
		if total > 1000 {
			t.Fatal("long playback")
		}
		if !known[a.ReducedFrames[0]] || a.ReducedMS[0] > 250 {
			t.Fatal("invalid reduced frame")
		}
	}
}

func TestPinnedInputRejectsEscapeLegacyAndTampering(t *testing.T) {
	dir := t.TempDir()
	im := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	var b bytes.Buffer
	if e := png.Encode(&b, im); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(dir, "art.png"), b.Bytes(), 0o644); e != nil {
		t.Fatal(e)
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _ = root.Close() }()
	valid := Input{ID: "weapon/bayonet", Source: "art.png", SHA256: hash(b.Bytes())}
	if _, e = loadInput(root, valid, true); e != nil {
		t.Fatal(e)
	}
	for _, path := range []string{"../art.png", "legacy/art.png", "original/art.png", "C:/art.png", "/art.png", "a\\art.png"} {
		bad := valid
		bad.Source = path
		if _, e = loadInput(root, bad, true); e == nil {
			t.Fatalf("accepted %s", path)
		}
	}
	valid.SHA256 = strings.Repeat("0", 64)
	if _, e = loadInput(root, valid, true); e == nil {
		t.Fatal("accepted wrong pin")
	}
}

func TestManifestAndCLIRejectIncompleteOrAmbiguousInputs(t *testing.T) {
	for _, b := range []string{`{"unknown":1}`, `{} {}`, `{} trailing`} {
		var m Manifest
		if e := decodeStrict([]byte(b), &m); e == nil {
			t.Fatalf("accepted %s", b)
		}
	}
	for _, args := range [][]string{{}, {"-out", "a", "-out", "b"}, {"-bad", "x"}, {"-manifest", "x", "-out", "y", "extra"}} {
		if e := Run(args, io.Discard, io.Discard); e == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if e := validateBindings(Manifest{}); e == nil {
		t.Fatal("accepted missing weapons")
	}
}

func TestLanczosSignedImpulseAndTransparentRGB(t *testing.T) {
	src := image.NewNRGBA(image.Rect(4, 7, 12, 11))
	for y := 7; y < 11; y++ {
		for x := 4; x < 12; x++ {
			p := color.NRGBA{A: 255}
			if x == 7 {
				p.R = 255
			}
			src.SetNRGBA(x, y, p)
		}
	}
	im, e := resizeLanczos(src, 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	for x, r := range []uint8{51, 29} {
		if im.NRGBAAt(x, 0) != (color.NRGBA{R: r, A: 255}) {
			t.Fatal("incorrect Lanczos support")
		}
	}
	transparent := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for i := range transparent.Pix {
		transparent.Pix[i] = 255
	}
	for i := 3; i < len(transparent.Pix); i += 4 {
		transparent.Pix[i] = 0
	}
	out, e := resizeLanczos(transparent, 4, 4)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range out.Pix {
		if v != 0 {
			t.Fatal("hidden RGB leaked")
		}
	}
}
