package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func forestTestWritePNG(t *testing.T, path string, im *image.NRGBA) connectionMaterial {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, im); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data.Bytes())
	return connectionMaterial{filepath.Base(path), hex.EncodeToString(sum[:])}
}

func forestTestAtlas() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	for panel := range 4 {
		for y := 24; y < 232; y++ {
			for x := 24; x < 232; x++ {
				// Asymmetric silhouettes expose accidental reflection, panel mixing or rotation.
				if (x-128)*(x-128)+(y-128)*(y-128) < 90*90 || x > 150 && x < 218 && y > 80 && y < 111 {
					im.SetNRGBA(panel%2*256+x, panel/2*256+y, color.NRGBA{uint8(110 + panel*31), uint8(38 + x/3), uint8(25 + y/4), 255})
				}
			}
		}
	}
	return im
}

func forestTestFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	grounds, atlases := map[string]connectionMaterial{}, map[string]connectionMaterial{}
	for i, season := range []string{"summer", "winter"} {
		ground := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := range 320 {
			for x := range 384 {
				ground.SetNRGBA(x, y, color.NRGBA{uint8(11 + x/4 + i*50), uint8(37 + y/3), uint8(19 + (x+y)/8), 255})
			}
		}
		grounds[season] = forestTestWritePNG(t, filepath.Join(dir, "ground-"+season+".png"), ground)
		atlas := forestTestAtlas()
		if i == 1 {
			for y := range 512 {
				for x := range 512 {
					p := atlas.NRGBAAt(x, y)
					if p.A != 0 {
						p.B += 80
						atlas.SetNRGBA(x, y, p)
					}
				}
			}
		}
		atlases[season] = forestTestWritePNG(t, filepath.Join(dir, "atlas-"+season+".png"), atlas)
	}
	data, err := json.Marshal(forestManifest{SchemaVersion: 1, Grounds: grounds, Atlases: atlases})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "manifest.json")
	if err = os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestForestStrictManifestAndAtlasGuards(t *testing.T) {
	path := forestTestFixture(t)
	original := groundTestRead(t, path)
	for name, change := range map[string]func(string) string{
		"schema_float": func(s string) string { return strings.Replace(s, `"schema_version":1`, `"schema_version":1.0`, 1) },
		"case_alias":   func(s string) string { return strings.Replace(s, `"grounds"`, `"Grounds"`, 1) },
		"null_season":  func(s string) string { return strings.Replace(s, `"atlases":{`, `"atlases":{"unknown":null,`, 1) },
		"extra_field":  func(s string) string { return strings.Replace(s, `"grounds":{`, `"extra":0,"grounds":{`, 1) },
		"duplicate": func(s string) string {
			return strings.Replace(s, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1)
		},
		"traversal": func(s string) string {
			return strings.Replace(s, `"source":"ground-summer.png"`, `"source":"../ground-summer.png"`, 1)
		},
		"bad_pin":  func(s string) string { return strings.Replace(s, `"source_sha256":"`, `"source_sha256":"0`, 1) },
		"null_pin": func(s string) string { return strings.Replace(s, `"source":"ground-summer.png"`, `"source":null`, 1) },
	} {
		t.Run(name, func(t *testing.T) {
			groundTestWrite(t, path, []byte(change(string(original))))
			if _, err := loadForestInputs(path); err == nil {
				t.Fatal("accepted malformed source manifest")
			}
		})
	}
	groundTestWrite(t, path, original)
	in, err := loadForestInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, season := range []string{"summer", "winter"} {
		source := in.Atlases[season]
		if !bytes.Equal(source.Bytes, groundTestRead(t, filepath.Join(filepath.Dir(path), source.Ref.Source))) {
			t.Fatal("loaded source bytes changed")
		}
	}
	for _, name := range []string{"size", "empty_panel", "guard", "no_transparency"} {
		t.Run(name, func(t *testing.T) {
			im := forestTestAtlas()
			switch name {
			case "size":
				im = image.NewNRGBA(image.Rect(0, 0, 514, 512))
			case "empty_panel":
				for y := range 256 {
					for x := range 256 {
						im.SetNRGBA(x, y, color.NRGBA{})
					}
				}
			case "guard":
				im.SetNRGBA(19, 128, color.NRGBA{1, 2, 3, 250})
			case "no_transparency":
				for y := range 256 {
					for x := range 256 {
						im.SetNRGBA(x, y, color.NRGBA{1, 2, 3, 255})
					}
				}
			}
			if err := validateForestAtlas(im); err == nil {
				t.Fatal("accepted invalid atlas")
			}
		})
	}
	// Alpha below the declared guard threshold is allowed; opaque pixels at
	// exactly floor(8%*256)=20 are outside that guard.
	im := forestTestAtlas()
	im.SetNRGBA(19, 128, color.NRGBA{1, 2, 3, 249})
	im.SetNRGBA(20, 128, color.NRGBA{1, 2, 3, 250})
	if err := validateForestAtlas(im); err != nil {
		t.Fatal(err)
	}
}

func TestForestLoadedBytesSnapshotAndResolvedSourceIdentity(t *testing.T) {
	path := forestTestFixture(t)
	in, err := loadForestInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(filepath.Dir(path), in.Manifest.Atlases["summer"].Source)
	loaded := append([]byte(nil), in.Atlases["summer"].Bytes...)
	groundTestWrite(t, sourcePath, []byte("source replaced after the one authorized read"))
	bundle, err := buildForestBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, artifact := range bundle.artifacts {
		if artifact.name == "sources/atlases-summer.png" {
			found = true
			if !bytes.Equal(artifact.data, loaded) {
				t.Fatal("reread changed source while snapshotting")
			}
		}
	}
	if !found {
		t.Fatal("missing retained source")
	}
	if _, err := loadForestInputs(path); err == nil {
		t.Fatal("accepted changed source bytes under old pin")
	}
	groundTestWrite(t, sourcePath, loaded)
	outside := filepath.Join(t.TempDir(), "outside.png")
	groundTestWrite(t, outside, loaded)
	link := filepath.Join(filepath.Dir(path), "escaped.png")
	if err = os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	m := in.Manifest
	ref := m.Atlases["summer"]
	ref.Source = "escaped.png"
	m.Atlases["summer"] = ref
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, data)
	if _, err := loadForestInputs(path); err == nil {
		t.Fatal("accepted symlink outside the source root")
	}
}

func TestForestRollbackAndExistingDirectory(t *testing.T) {
	path := forestTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "rollback")
	if err := run([]string{"-forest", path, "-out", out}, groundTestBrokenWriter{}); err == nil {
		t.Fatal("expected write failure")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatalf("failed export left partial output: %v", err)
	}
	if err := os.Mkdir(out, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(out, "keep.txt")
	groundTestWrite(t, marker, []byte("untouched"))
	if err := run([]string{"-forest", path, "-out", out}, &bytes.Buffer{}); err == nil {
		t.Fatal("existing output accepted")
	}
	if string(groundTestRead(t, marker)) != "untouched" {
		t.Fatal("existing output changed")
	}
	// A directory concurrently replaced by the log sink belongs to its creator;
	// rollback must remove only the inode originally created by this export.
	out2 := filepath.Join(filepath.Dir(path), "replacement")
	err := run([]string{"-forest", path, "-out", out2}, groundTestReplacingWriter{t, out2, filepath.Join(filepath.Dir(path), "moved-owned")})
	if err == nil {
		t.Fatal("expected replacement sink failure")
	}
	if _, err = os.Stat(out2); err != nil {
		t.Fatal("rollback removed replacement directory")
	}
}

func TestForestPanelRequiresTransparencyInItsMargin(t *testing.T) {
	im := forestTestAtlas()
	// A zero-alpha hole in the middle cannot substitute for a transparent
	// margin around the complete crown panel.
	for y := range 256 {
		for x := range 256 {
			p := im.NRGBAAt(x, y)
			if p.A == 0 {
				p = color.NRGBA{1, 2, 3, 249}
				im.SetNRGBA(x, y, p)
			}
		}
	}
	im.SetNRGBA(128, 128, color.NRGBA{})
	if err := validateForestAtlas(im); err == nil {
		t.Fatal("accepted translucent filled panel with no transparent margin")
	}
}
