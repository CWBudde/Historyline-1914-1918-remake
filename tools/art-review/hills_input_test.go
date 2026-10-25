package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hillsTestHeight() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	for y := range 512 {
		for x := range 512 {
			v := uint8(40 + x/4 + y/8)
			im.SetNRGBA(x, y, color.NRGBA{v, v + 2, v + 4, 255})
		}
	}
	return im
}

func hillsTestFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	grass := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			grass.SetNRGBA(x, y, color.NRGBA{uint8(40 + x/5), uint8(65 + y/4), 31, 255})
		}
	}
	refs := map[string]connectionMaterial{"grass": forestTestWritePNG(t, filepath.Join(dir, "grass.png"), grass), "height": forestTestWritePNG(t, filepath.Join(dir, "height.png"), hillsTestHeight())}
	data, err := json.Marshal(hillsManifest{1, refs})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "manifest.json")
	groundTestWrite(t, path, data)
	return path
}

// A permissive JSON decoder or unpinned source read would accept these manifests.
func TestHillsStrictInputs(t *testing.T) {
	path := hillsTestFixture(t)
	original := groundTestRead(t, path)
	for name, change := range map[string]func(string) string{
		"decimal": func(s string) string { return strings.Replace(s, `"schema_version":1`, `"schema_version":1.0`, 1) },
		"case":    func(s string) string { return strings.Replace(s, `"materials"`, `"Materials"`, 1) },
		"duplicate": func(s string) string {
			return strings.Replace(s, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1)
		},
		"unknown": func(s string) string { return strings.Replace(s, `"materials":{`, `"extra":0,"materials":{`, 1) },
		"missing": func(s string) string { return strings.Replace(s, `"height":`, `"other":`, 1) },
		"null":    func(s string) string { return strings.Replace(s, `"source":"height.png"`, `"source":null`, 1) },
		"traversal": func(s string) string {
			return strings.Replace(s, `"source":"height.png"`, `"source":"../height.png"`, 1)
		},
		"pin": func(s string) string { return strings.Replace(s, `"source_sha256":"`, `"source_sha256":"0`, 1) },
	} {
		t.Run(name, func(t *testing.T) {
			groundTestWrite(t, path, []byte(change(string(original))))
			if _, err := loadHillsInputs(path); err == nil {
				t.Fatal("accepted malformed hills input")
			}
		})
	}
	groundTestWrite(t, path, original)
	in, err := loadHillsInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(in.Bytes, original) || len(in.Sources) != 2 {
		t.Fatal("did not retain inputs")
	}
	for name, im := range map[string]*image.NRGBA{"height": hillsTestHeight(), "nonsquare": image.NewNRGBA(image.Rect(0, 0, 512, 513)), "small": image.NewNRGBA(image.Rect(0, 0, 511, 511)), "large": image.NewNRGBA(image.Rect(0, 0, 4097, 1))} {
		err := validateHillsHeight(im)
		if (err == nil) != (name == "height") {
			t.Fatalf("height bounds %s: %v", name, err)
		}
	}
	im := hillsTestHeight()
	im.SetNRGBA(0, 0, color.NRGBA{100, 108, 104, 255})
	if err := validateHillsHeight(im); err != nil {
		t.Fatal("spread eight rejected", err)
	}
	im.SetNRGBA(0, 0, color.NRGBA{100, 109, 104, 255})
	if validateHillsHeight(im) == nil {
		t.Fatal("spread nine accepted")
	}
	im.SetNRGBA(0, 0, color.NRGBA{100, 100, 100, 254})
	if validateHillsHeight(im) == nil {
		t.Fatal("alpha accepted")
	}
	// Resolve symlinks before accepting provenance identity.
	source := in.Sources["height"]
	outside := filepath.Join(t.TempDir(), "outside.png")
	groundTestWrite(t, outside, source.Bytes)
	if err := os.Symlink(outside, filepath.Join(filepath.Dir(path), "escaped.png")); err != nil {
		t.Fatal(err)
	}
	ref := in.Manifest.Materials["height"]
	ref.Source = "escaped.png"
	in.Manifest.Materials["height"] = ref
	data, _ := json.Marshal(in.Manifest)
	groundTestWrite(t, path, data)
	if _, err := loadHillsInputs(path); err == nil {
		t.Fatal("escaped symlink accepted")
	}
}

// Losing absolute phase, negative modulo or order merges distinct relief profiles.
func TestHillsContextPhaseIdentity(t *testing.T) {
	all := []string{"grass", "grass", "grass", "grass", "grass", "grass"}
	c, err := newHillsContext("hills", all, connectionPoint{71, -40})
	if err != nil {
		t.Fatal(err)
	}
	want := "hills-v1\nseason=summer\ncenter=hills\nphase=71,280\nNE=grass\nSE=grass\nS=grass\nSW=grass\nNW=grass\nN=grass\n"
	sum := sha256.Sum256([]byte(want))
	if c.descriptor() != want || c.key() != hex.EncodeToString(sum[:]) {
		t.Fatalf("phase descriptor %q", c.descriptor())
	}
	for _, p := range []connectionPoint{{0, 40}, {1, 0}, {0, .1}, {math.Inf(1), 0}, {0, math.NaN()}} {
		if _, err := newHillsContext("hills", all, p); err == nil {
			t.Fatalf("invalid lattice origin %v", p)
		}
	}
	for _, id := range []string{"mud", "hills_w", "Hills", ""} {
		if _, err := newHillsContext(id, all, connectionPoint{}); err == nil {
			t.Fatal("invalid center", id)
		}
		bad := append([]string(nil), all...)
		bad[0] = id
		if _, err := newHillsContext("hills", bad, connectionPoint{}); err == nil {
			t.Fatal("invalid neighbor", id)
		}
	}
	if _, err := newHillsContext("hills", all[:5], connectionPoint{}); err == nil {
		t.Fatal("neighbor count")
	}
	seen := map[string]bool{}
	for mask := range 64 {
		n := append([]string(nil), all...)
		for i := range 6 {
			if mask&(1<<i) != 0 {
				n[i] = "hills"
			}
		}
		c, _ := newHillsContext("hills", n, connectionPoint{})
		if seen[c.key()] {
			t.Fatal("merged ordered mask")
		}
		seen[c.key()] = true
	}
	phases := map[connectionPoint]bool{}
	for line := range 8 {
		for col := range 4 {
			p := connectionPoint{float64(142*col + 71*(line%2)), float64(40 * line)}
			a, _ := newHillsContext("hills", all, p)
			phases[a.Phase] = true
			for _, d := range []connectionPoint{{568, 0}, {0, 320}, {-568, -320}} {
				b, _ := newHillsContext("hills", all, p.add(d))
				if a.key() != b.key() {
					t.Fatal("period key")
				}
			}
		}
	}
	if len(phases) != 32 {
		t.Fatal("phase cosets", len(phases))
	}
	base, _ := newHillsContext("grass", []string{"hills", "hills", "hills", "hills", "hills", "hills"}, connectionPoint{71, 40})
	canonical, _ := newHillsContext("grass", all, connectionPoint{})
	if base.key() != canonical.key() {
		t.Fatal("grass normalization")
	}
	a, _ := newHillsContext("hills", all, connectionPoint{})
	b, _ := newHillsContext("hills", all, connectionPoint{71, 40})
	if a.key() == b.key() || !a.canonical() || b.canonical() {
		t.Fatal("phase identity/canonical")
	}
}
