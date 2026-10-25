package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hillsTestIdentity(id string, n [6]string, p connectionPoint) (string, string) {
	if id == "grass" {
		p = connectionPoint{}
		n = [6]string{"grass", "grass", "grass", "grass", "grass", "grass"}
	}
	x, y := math.Mod(p[0], 568), math.Mod(p[1], 320)
	if x < 0 {
		x += 568
	}
	if y < 0 {
		y += 320
	}
	s := fmt.Sprintf("hills-v1\nseason=summer\ncenter=%s\nphase=%.0f,%.0f\n", id, x, y)
	for i, label := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
		s += label + "=" + n[i] + "\n"
	}
	sum := sha256.Sum256([]byte(s))
	return s, hex.EncodeToString(sum[:])
}

// Exported scene pixels must be ordinary Over of decoded files, with real phase/context identity.
func TestHillsBundleActualPNGsTopologyAndSnapshots(t *testing.T) {
	path := hillsTestFixture(t)
	in, err := loadHillsInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded := append([]byte(nil), in.Sources["height"].Bytes...)
	groundTestWrite(t, filepath.Join(filepath.Dir(path), "height.png"), []byte("changed after load"))
	b, err := buildHillsBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	actual := map[string]*image.NRGBA{}
	sources := 0
	for _, a := range b.artifacts {
		if strings.HasSuffix(a.name, ".png") {
			if strings.HasPrefix(a.name, "sources/") {
				sources++
				if a.name == "sources/height.png" && !bytes.Equal(a.data, loaded) {
					t.Fatal("source reread")
				}
				continue
			}
			im, err := decodeHillsTestPNG(a.data)
			if err != nil {
				t.Fatal(err)
			}
			actual[a.name] = im
		}
	}
	if sources != 2 || len(b.report.Scenes) != 8 {
		t.Fatal("bounded source/scene roster")
	}
	if len(b.report.Profiles) < 96 || len(b.report.Exports) != 2*len(b.report.Profiles)+16 {
		t.Fatal("derived profile/export count")
	}
	profiles := map[string]hillsProfileReport{}
	quietPorts := map[int][]uint8{}
	for _, scale := range []int{2, 4} {
		ports := make([]uint8, 96*80*scale*scale)
		for y := range 80 * scale {
			for x := range 96 * scale {
				for n, d := range groundTestDeltas {
					if groundTestDistance((float64(x)+.5)/float64(scale)-float64(d[0]), (float64(y)+.5)/float64(scale)-float64(d[1])) <= 4 {
						ports[y*96*scale+x] |= 1 << n
					}
				}
			}
		}
		quietPorts[scale] = ports
	}
	phase0 := 0
	canonical := 0
	isolatedPhases := map[connectionPoint]bool{}
	for _, p := range b.report.Profiles {
		var grassPorts uint8
		for n, id := range p.NeighborIDs {
			if id == "grass" {
				grassPorts |= 1 << n
			}
		}
		desc, key := hillsTestIdentity(p.RuntimeID, p.NeighborIDs, p.Phase)
		if p.Descriptor != desc || p.ContextKey != key {
			t.Fatal("wrong profile descriptor")
		}
		if _, ok := profiles[key]; ok {
			t.Fatal("duplicate profile")
		}
		profiles[key] = p
		if p.RuntimeID == "hills" && p.Phase == (connectionPoint{}) {
			phase0++
		}
		if p.Canonical {
			canonical++
			if p.Phase != (connectionPoint{}) {
				t.Fatal("nonzero canonical")
			}
		}
		isolated := p.RuntimeID == "hills"
		for _, id := range p.NeighborIDs {
			isolated = isolated && id == "grass"
		}
		if isolated {
			isolatedPhases[p.Phase] = true
			if !p.IsolatedCue.Enabled {
				t.Fatal("isolated phase missing cue metadata")
			}
		} else if p.IsolatedCue != (hillsCue{}) {
			t.Fatal("nonisolated profile has cue")
		}
		for _, out := range p.Outputs {
			im := actual[out.Path]
			if im == nil {
				t.Fatal("missing actual PNG")
			}
			scale := im.Bounds().Dx() / 96
			for y := range 80 * scale {
				for x := range 96 * scale {
					pixel := im.NRGBAAt(x, y)
					if groundTestOwned(x, y, scale) {
						if pixel.A != 255 {
							t.Fatal("actual PNG hole")
						}
						if p.RuntimeID == "hills" && grassPorts&quietPorts[scale][y*96*scale+x] != 0 && pixel != actual[fmt.Sprintf("tiles/grass-%dx.png", scale)].NRGBAAt(x, y) {
							t.Fatalf("actual PNG final quiet band %s phase %v at %d,%d", out.Path, p.Phase, x, y)
						}
					} else if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 || pixel.A != 0 {
						t.Fatal("actual PNG outside")
					}
				}
			}
		}
	}
	if phase0 != 64 || canonical != 2 {
		t.Fatalf("phase0/canonical %d/%d", phase0, canonical)
	}
	if len(isolatedPhases) != 32 {
		t.Fatalf("exported isolated phases: %d", len(isolatedPhases))
	}
	for _, s := range b.report.Scenes {
		if len(s.Sites) != 19 || len(s.Cells) != 7 || len(s.Outputs) != 2 {
			t.Fatal("scene roster")
		}
		if s.CenterOrigin != (connectionPoint{float64(71 * s.CenterLineParity), float64(40 * s.CenterLineParity)}) {
			t.Fatal("fake parity")
		}
		world := map[connectionPoint]string{}
		for _, site := range s.Sites {
			world[site.Origin] = site.RuntimeID
			hill := strings.Contains(s.ID, "homogeneous") || strings.Contains(s.ID, "isolated") && site.Origin == s.CenterOrigin || strings.Contains(s.ID, "clearing") && site.Origin != s.CenterOrigin || strings.Contains(s.ID, "frontier") && site.Origin[0] >= s.CenterOrigin[0]
			if (site.RuntimeID == "hills") != hill {
				t.Fatal("scenario topology")
			}
		}
		for _, c := range s.Cells {
			if c.RuntimeID != world[c.Origin] {
				t.Fatal("cell origin ID")
			}
			for n, d := range groundTestDeltas {
				if c.NeighborIDs[n] != world[c.Origin.add(connectionPoint{float64(d[0]), float64(d[1])})] {
					t.Fatal("actual neighbor topology")
				}
			}
			_, key := hillsTestIdentity(c.RuntimeID, c.NeighborIDs, c.Origin)
			if c.ContextKey != key {
				t.Fatal("fixture phase key")
			}
			if _, ok := profiles[key]; !ok {
				t.Fatal("missing fixture phase profile")
			}
		}
		for _, scale := range []int{2, 4} {
			reassembled := image.NewNRGBA(image.Rect(0, 0, 238*scale, 240*scale))
			for _, c := range s.Cells {
				p := profiles[c.ContextKey]
				tile := actual[p.Outputs[0].Path]
				for _, out := range p.Outputs {
					if actual[out.Path].Bounds().Dx() == 96*scale {
						tile = actual[out.Path]
					}
				}
				at := image.Pt(int(c.Origin[0]-s.CanvasOrigin[0])*scale, int(c.Origin[1]-s.CanvasOrigin[1])*scale)
				draw.Draw(reassembled, tile.Bounds().Add(at), tile, image.Point{}, draw.Over)
			}
			saved := actual[s.Outputs[0].Path]
			for _, out := range s.Outputs {
				if actual[out.Path].Bounds().Dx() == 238*scale {
					saved = actual[out.Path]
				}
			}
			if saved == nil || !bytes.Equal(reassembled.Pix, saved.Pix) {
				t.Fatalf("decoded PNG assembly %s %dx", s.ID, scale)
			}
			for y := range 240 * scale {
				for x := range 238 * scale {
					owned := false
					for _, c := range s.Cells {
						owned = owned || groundTestOwned(x-int(c.Origin[0]-s.CanvasOrigin[0])*scale, y-int(c.Origin[1]-s.CanvasOrigin[1])*scale, scale)
					}
					if owned && saved.NRGBAAt(x, y).A != 255 {
						t.Fatal("fixture ownership hole")
					}
				}
			}
		}
	}
	if _, err := loadHillsInputs(path); err == nil {
		t.Fatal("changed source accepted")
	}
	t.Logf("hills bundle: %d profiles, %d scenes, %d PNGs, %d artifacts", len(b.report.Profiles), len(b.report.Scenes), len(b.report.Exports), len(b.artifacts))
	// Report serialization must retain the explicit numeric phase and literal descriptor.
	data, err := json.Marshal(b.report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"phase":`)) {
		t.Fatal("missing phase metadata")
	}
}

func TestHillsOutputRollbackAndExclusiveDirectory(t *testing.T) {
	path := hillsTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "out")
	if err := run([]string{"-hills", path, "-out", out}, groundTestBrokenWriter{}); err == nil {
		t.Fatal("write failure expected")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("rollback left output", err)
	}
	if err := os.Mkdir(out, 0o700); err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, filepath.Join(out, "keep"), []byte("keep"))
	if err := runHills(path, out, &bytes.Buffer{}); err == nil {
		t.Fatal("existing dir accepted")
	}
	if string(groundTestRead(t, filepath.Join(out, "keep"))) != "keep" {
		t.Fatal("existing file altered")
	}
	out2 := filepath.Join(filepath.Dir(path), "replacement")
	if err := runHills(path, out2, groundTestReplacingWriter{t, out2, filepath.Join(filepath.Dir(path), "moved")}); err == nil {
		t.Fatal("replacement sink expected failure")
	}
	if _, err := os.Stat(out2); err != nil {
		t.Fatal("removed replacement", err)
	}
}

func decodeHillsTestPNG(data []byte) (*image.NRGBA, error) {
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	im := image.NewNRGBA(decoded.Bounds())
	draw.Draw(im, im.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	return im, nil
}
