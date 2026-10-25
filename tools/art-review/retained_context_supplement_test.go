//go:build retainedgfx

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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetainedContextBundleRejectsSymlinkIntoActiveArt(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(filepath.Join(root, "build"), "retained-symlink-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove ignored symlink fixture: %v", err)
		}
	}()
	link := filepath.Join(dir, "bundle")
	if err = os.Symlink(filepath.Join(root, "art"), link); err != nil {
		t.Fatal(err)
	}
	manifest, _ := groundTestFixture(t)
	err = augmentRetainedContextBundle("ground", manifest, filepath.Join(t.TempDir(), "recipe.json"), link)
	if err == nil || !strings.Contains(err.Error(), "context bundle") {
		t.Fatalf("active art symlink was not rejected at confinement boundary: %v", err)
	}
}

// The external directory stands in for active art without touching it. A
// confined bundle root must also reject a nested profiles/ symlink escape.
func TestRetainedContextBundleRejectsNestedProfilesSymlink(t *testing.T) {
	manifest, _ := groundTestFixture(t)
	bundle := t.TempDir()
	input := groundTestRead(t, manifest)
	groundTestWrite(t, filepath.Join(bundle, "manifest.json"), input)
	sum := sha256.Sum256(input)
	report, _ := json.Marshal(map[string]any{"schema_version": 1, "manifest_sha256": hex.EncodeToString(sum[:]), "profiles": []any{}, "exports": []any{}})
	groundTestWrite(t, filepath.Join(bundle, "report.json"), report)
	c, err := newGroundContext("mud", []string{"mud", "mud", "mud", "mud", "mud", "mud"})
	if err != nil {
		t.Fatal(err)
	}
	profile := groundProfileReport{RuntimeID: "mud", Season: "summer", ContextKey: c.key(), Descriptor: c.descriptor(), NeighborIDs: c.NeighborIDs}
	recipe := filepath.Join(t.TempDir(), "recipe.json")
	data, _ := json.Marshal(map[string]any{"profiles": []groundProfileReport{profile}})
	groundTestWrite(t, recipe, data)
	external := t.TempDir()
	sentinel := []byte("external art must remain unchanged")
	groundTestWrite(t, filepath.Join(external, "sentinel"), sentinel)
	if err = os.Symlink(external, filepath.Join(bundle, "profiles")); err != nil {
		t.Fatal(err)
	}
	if err = augmentRetainedContextBundle("ground", manifest, recipe, bundle); err == nil {
		t.Fatal("nested profiles symlink escaped the bundle without an error")
	}
	if !bytes.Equal(report, groundTestRead(t, filepath.Join(bundle, "report.json"))) {
		t.Fatal("escaped artifact updated report before rejection")
	}
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "sentinel" || !bytes.Equal(sentinel, groundTestRead(t, filepath.Join(external, "sentinel"))) {
		t.Fatal("artifact writes escaped into external art directory")
	}
}

// Missing supplement profiles, stale output hashes, lost hills phases, or a
// second run appending duplicates all break the promotion/runtime contract.
func TestRetainedContextSupplementFamilies(t *testing.T) {
	for _, family := range []string{"ground", "forest", "hills"} {
		t.Run(family, func(t *testing.T) {
			var manifest, runtime, season, descriptor, key string
			var neighbors [6]string
			phase := connectionPoint{}
			switch family {
			case "ground":
				manifest, _ = groundTestFixture(t)
				runtime, season = "mud", "summer"
				neighbors = [6]string{"mud", "mud", "mud", "mud", "mud", "mud"}
				c, e := newGroundContext(runtime, neighbors[:])
				if e != nil {
					t.Fatal(e)
				}
				descriptor, key = c.descriptor(), c.key()
				var m map[string]any
				if e = json.Unmarshal(groundTestRead(t, manifest), &m); e != nil {
					t.Fatal(e)
				}
				m["mud_feature"] = "circular-earthwork"
				data, _ := json.Marshal(m)
				groundTestWrite(t, manifest, data)
			case "forest":
				manifest = forestTestFixture(t)
				runtime, season = "forest", "summer"
				neighbors = [6]string{"forest", "grass", "forest", "grass", "forest", "grass"}
				c, e := newForestContext(runtime, neighbors[:])
				if e != nil {
					t.Fatal(e)
				}
				descriptor, key = c.descriptor(), c.key()
			case "hills":
				manifest = hillsTestFixture(t)
				runtime, season = "hills", "summer"
				neighbors = [6]string{"hills", "grass", "hills", "grass", "hills", "grass"}
				phase = connectionPoint{213, 40}
				c, e := newHillsContext(runtime, neighbors[:], phase)
				if e != nil {
					t.Fatal(e)
				}
				descriptor, key = c.descriptor(), c.key()
			}
			bundle := filepath.Join(t.TempDir(), "bundle")
			if e := os.Mkdir(bundle, 0o755); e != nil {
				t.Fatal(e)
			}
			input := groundTestRead(t, manifest)
			groundTestWrite(t, filepath.Join(bundle, "manifest.json"), input)
			sum := sha256.Sum256(input)
			report := map[string]any{"schema_version": 1, "manifest_sha256": hex.EncodeToString(sum[:]), "profiles": []any{}, "exports": []any{}}
			data, _ := json.Marshal(report)
			groundTestWrite(t, filepath.Join(bundle, "report.json"), data)
			profile := map[string]any{"runtime_id": runtime, "season": season, "neighbor_ids": neighbors, "context_key": key, "descriptor": descriptor, "phase": phase, "canonical": false}
			recipe := filepath.Join(t.TempDir(), "recipe.json")
			data, _ = json.Marshal(map[string]any{"profiles": []any{profile}})
			groundTestWrite(t, recipe, data)
			before := groundTestRead(t, recipe)
			// Wrong input bytes or stale retained identities must fail before
			// supplement files or report changes become visible.
			groundTestWrite(t, filepath.Join(bundle, "manifest.json"), append(append([]byte(nil), input...), '\n'))
			if e := augmentRetainedContextBundle(family, manifest, recipe, bundle); e == nil {
				t.Fatal("accepted mismatched fresh manifest snapshot")
			}
			groundTestWrite(t, filepath.Join(bundle, "manifest.json"), input)
			originalReport := groundTestRead(t, filepath.Join(bundle, "report.json"))
			profile["context_key"] = "invalid-retained-key"
			data, _ = json.Marshal(map[string]any{"profiles": []any{profile}})
			groundTestWrite(t, recipe, data)
			if e := augmentRetainedContextBundle(family, manifest, recipe, bundle); e == nil {
				t.Fatal("accepted retained descriptor/key mismatch")
			}
			if !bytes.Equal(originalReport, groundTestRead(t, filepath.Join(bundle, "report.json"))) {
				t.Fatal("invalid context changed fresh report")
			}
			if _, e := os.Stat(filepath.Join(bundle, "profiles")); !os.IsNotExist(e) {
				t.Fatalf("invalid context wrote PNGs: %v", e)
			}
			groundTestWrite(t, recipe, before)
			for range 2 {
				if e := augmentRetainedContextBundle(family, manifest, recipe, bundle); e != nil {
					t.Fatal(e)
				}
			}
			var got struct {
				Profiles []struct {
					RuntimeID  string          `json:"runtime_id"`
					Descriptor string          `json:"descriptor"`
					Phase      connectionPoint `json:"phase"`
					Outputs    []packOutput    `json:"outputs"`
				}
				Exports []packOutput `json:"exports"`
			}
			if e := json.Unmarshal(groundTestRead(t, filepath.Join(bundle, "report.json")), &got); e != nil {
				t.Fatal(e)
			}
			if len(got.Profiles) != 1 || len(got.Exports) != 2 {
				t.Fatalf("missing or duplicate supplement: %d profiles, %d outputs", len(got.Profiles), len(got.Exports))
			}
			p := got.Profiles[0]
			if p.RuntimeID != runtime || p.Descriptor != descriptor || (family == "hills" && p.Phase != phase) {
				t.Fatalf("context identity changed: %+v", p)
			}
			for _, out := range p.Outputs {
				b := groundTestRead(t, filepath.Join(bundle, out.Path))
				sum := sha256.Sum256(b)
				if hex.EncodeToString(sum[:]) != out.SHA256 {
					t.Fatal("stale PNG hash")
				}
				im, e := png.Decode(bytes.NewReader(b))
				if e != nil {
					t.Fatal(e)
				}
				if im.Bounds().Dx() != out.Dimensions.Width {
					t.Fatal("wrong output dimensions")
				}
				if family == "ground" && out.Dimensions.Width == 384 {
					// The opt-in circular earthwork leaves grass at the top guard;
					// forgetting the manifest hook would instead fill it with mud.
					in, e := loadGroundInputs(manifest)
					if e != nil {
						t.Fatal(e)
					}
					materials := map[string]*image.NRGBA{}
					for id, s := range in.Sources {
						materials[id] = s.Image
					}
					r := newGroundRenderer(materials)
					base, e := newGroundContext("grass", []string{"grass", "grass", "grass", "grass", "grass", "grass"})
					if e != nil {
						t.Fatal(e)
					}
					grass, e := r.tiles(base)
					if e != nil {
						t.Fatal(e)
					}
					if got, want := color.NRGBAModel.Convert(im.At(192, 8)), grass[4].NRGBAAt(192, 8); got != want {
						t.Fatalf("supplement ignored mud_feature grass guard: %v != %v", got, want)
					}
				}
			}
			if len(p.Outputs) != 2 || p.Outputs[0].Path != fmt.Sprintf("profiles/%s-2x.png", key) {
				t.Fatal("wrong profile output paths")
			}
			if !bytes.Equal(before, groundTestRead(t, recipe)) {
				t.Fatal("retained recipe modified")
			}
		})
	}
}
