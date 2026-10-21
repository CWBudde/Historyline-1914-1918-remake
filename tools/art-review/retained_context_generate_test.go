//go:build retainedgfx

package main

// Opt-in offline authoring hook. It supplements a freshly generated ignored
// bundle; it never promotes or modifies active art or the context profiles.
//
// From repository root:
// go test -tags retainedgfx ./tools/art-review -run '^TestGenerateRetainedContextBundle$' -count=1 -args -context-family ground -context-manifest art/pipeline/ground-terrain-v1.json -context-profiles art/pipeline/contexts/ground.json -context-bundle build/review/fresh-ground
import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var retainedContextFamily = flag.String("context-family", "", "ground, forest, or hills")
var retainedContextManifest = flag.String("context-manifest", "", "current pinned input manifest; absolute or repository-relative")
var retainedContextProfiles = flag.String("context-profiles", "", "context profiles file (art/pipeline/contexts/<family>.json); absolute or repository-relative")
var retainedContextBundle = flag.String("context-bundle", "", "fresh ignored CLI bundle to supplement; absolute or repository-relative")

func TestGenerateRetainedContextBundle(t *testing.T) {
	if *retainedContextFamily == "" && *retainedContextManifest == "" && *retainedContextProfiles == "" && *retainedContextBundle == "" {
		t.Skip("explicit context flags required")
	}
	if err := augmentRetainedContextBundle(*retainedContextFamily, *retainedContextManifest, *retainedContextProfiles, *retainedContextBundle); err != nil {
		t.Fatal(err)
	}
}

func retainedContextPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("all four context flags are required")
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join("..", "..", path))
}

// Append only absent identities; rendering replaces stale PNG metadata from
// the profiles file. Re-running the helper leaves existing profiles intact.
func retainedAppendProfiles[P any](profilesFile []byte, profiles *[]P, key func(P) string, render func(P) (P, error)) error {
	var retained struct {
		Profiles []P `json:"profiles"`
	}
	if err := json.Unmarshal(profilesFile, &retained); err != nil {
		return err
	}
	if retained.Profiles == nil {
		return fmt.Errorf("context profiles file requires profiles")
	}
	known := map[string]bool{}
	for _, p := range *profiles {
		known[key(p)] = true
	}
	for _, p := range retained.Profiles {
		if known[key(p)] {
			continue
		}
		updated, err := render(p)
		if err != nil {
			return err
		}
		*profiles = append(*profiles, updated)
		known[key(p)] = true
	}
	return nil
}

func retainedContextOutputs(key string, tiles map[int]*image.NRGBA, add func(string, *image.NRGBA) (packOutput, error)) ([]packOutput, error) {
	var outputs []packOutput
	for _, scale := range []int{2, 4} {
		if tiles[scale] == nil {
			return nil, fmt.Errorf("missing %dx render for %s", scale, key)
		}
		out, err := add(fmt.Sprintf("profiles/%s-%dx.png", key, scale), tiles[scale])
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, out)
	}
	return outputs, nil
}

func augmentRetainedContextBundle(family, manifest, profilesFile, bundle string) error {
	if family != "ground" && family != "forest" && family != "hills" {
		return fmt.Errorf("context-family requires ground, forest, or hills")
	}
	var err error
	manifest, err = retainedContextPath(manifest)
	if err != nil {
		return err
	}
	profilesFile, err = retainedContextPath(profilesFile)
	if err != nil {
		return err
	}
	bundle, err = retainedContextPath(bundle)
	if err != nil {
		return err
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return err
	}
	// Resolve existing directories before confinement: a lexical build path
	// may otherwise point through a symlink into active art.
	bundle, err = filepath.EvalSymlinks(bundle)
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	temporary, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		return err
	}
	inside := func(parent, path string) bool {
		rel, e := filepath.Rel(parent, path)
		return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	if !inside(filepath.Join(root, "build"), bundle) && !inside(temporary, bundle) {
		return fmt.Errorf("context bundle must be ignored build output or a temporary directory")
	}
	if inside(filepath.Join(root, "art"), bundle) || inside(filepath.Join(root, "assets"), bundle) {
		return fmt.Errorf("active art is not a context bundle")
	}
	bundleRoot, err := os.OpenRoot(bundle)
	if err != nil {
		return err
	}
	defer func() { _ = bundleRoot.Close() }()
	input, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}
	snapshot, err := bundleRoot.ReadFile("manifest.json")
	if err != nil {
		return err
	}
	if !bytes.Equal(input, snapshot) {
		return fmt.Errorf("fresh bundle manifest snapshot does not match pinned input")
	}
	reportBytes, err := bundleRoot.ReadFile("report.json")
	if err != nil {
		return err
	}
	var header struct {
		ManifestSHA256 string `json:"manifest_sha256"`
	}
	if err = json.Unmarshal(reportBytes, &header); err != nil {
		return err
	}
	sum := sha256.Sum256(input)
	if header.ManifestSHA256 != hex.EncodeToString(sum[:]) {
		return fmt.Errorf("fresh bundle report manifest hash mismatch")
	}
	profileBytes, err := os.ReadFile(profilesFile)
	if err != nil {
		return err
	}
	var report any
	var artifacts []connectionArtifact
	switch family {
	case "ground":
		in, e := loadGroundInputs(manifest)
		if e != nil {
			return e
		}
		b := groundBundle{pixels: map[string]*image.NRGBA{}}
		if e = json.Unmarshal(reportBytes, &b.report); e != nil {
			return e
		}
		materials := map[string]*image.NRGBA{}
		for id, s := range in.Sources {
			materials[id] = s.Image
		}
		r := newGroundRenderer(materials)
		r.mudFeature = in.Manifest.MudFeature
		e = retainedAppendProfiles(profileBytes, &b.report.Profiles, func(p groundProfileReport) string { return p.ContextKey }, func(p groundProfileReport) (groundProfileReport, error) {
			c, e := newGroundContext(p.RuntimeID, p.NeighborIDs[:])
			if e != nil {
				return p, e
			}
			season, _ := groundSeason(p.RuntimeID)
			if c.key() != p.ContextKey || c.descriptor() != p.Descriptor || p.Season != season {
				return p, fmt.Errorf("retained ground identity mismatch %s", p.ContextKey)
			}
			tiles, e := r.tiles(c)
			if e != nil {
				return p, e
			}
			p.Outputs, e = retainedContextOutputs(p.ContextKey, tiles, b.addPNG)
			return p, e
		})
		if e != nil {
			return e
		}
		report, artifacts = b.report, b.artifacts
	case "forest":
		in, e := loadForestInputs(manifest)
		if e != nil {
			return e
		}
		b := forestBundle{pixels: map[string]*image.NRGBA{}}
		if e = json.Unmarshal(reportBytes, &b.report); e != nil {
			return e
		}
		grounds, atlases := map[string]*image.NRGBA{}, map[string]*image.NRGBA{}
		for id, s := range in.Grounds {
			grounds[id] = s.Image
		}
		for id, s := range in.Atlases {
			atlases[id] = s.Image
		}
		r, e := newForestRenderer(grounds, atlases, in.Manifest.SummerProfile)
		if e != nil {
			return e
		}
		e = retainedAppendProfiles(profileBytes, &b.report.Profiles, func(p forestProfileReport) string { return p.ContextKey }, func(p forestProfileReport) (forestProfileReport, error) {
			c, e := newForestContext(p.RuntimeID, p.NeighborIDs[:])
			if e != nil {
				return p, e
			}
			season, _ := forestSeason(p.RuntimeID)
			if c.key() != p.ContextKey || c.descriptor() != p.Descriptor || p.Season != season {
				return p, fmt.Errorf("retained forest identity mismatch %s", p.ContextKey)
			}
			tiles, e := r.tiles(c)
			if e != nil {
				return p, e
			}
			p.Instances = r.instances(c)
			p.Outputs, e = retainedContextOutputs(p.ContextKey, tiles, b.addPNG)
			return p, e
		})
		if e != nil {
			return e
		}
		report, artifacts = b.report, b.artifacts
	case "hills":
		in, e := loadHillsInputs(manifest)
		if e != nil {
			return e
		}
		b := hillsBundle{pixels: map[string]*image.NRGBA{}}
		if e = json.Unmarshal(reportBytes, &b.report); e != nil {
			return e
		}
		r := newHillsRenderer(in.Sources["grass"].Image, in.Sources["height"].Image)
		e = retainedAppendProfiles(profileBytes, &b.report.Profiles, func(p hillsProfileReport) string { return p.ContextKey }, func(p hillsProfileReport) (hillsProfileReport, error) {
			c, e := newHillsContext(p.RuntimeID, p.NeighborIDs[:], p.Phase)
			if e != nil {
				return p, e
			}
			if c.key() != p.ContextKey || c.descriptor() != p.Descriptor || c.Phase != p.Phase || p.Season != "summer" {
				return p, fmt.Errorf("retained hills identity mismatch %s", p.ContextKey)
			}
			tiles, e := r.tiles(c)
			if e != nil {
				return p, e
			}
			p.IsolatedCue = r.cueFor(c)
			p.Outputs, e = retainedContextOutputs(p.ContextKey, tiles, b.addPNG)
			return p, e
		})
		if e != nil {
			return e
		}
		report, artifacts = b.report, b.artifacts
	}
	for _, a := range artifacts {
		if old, e := bundleRoot.ReadFile(a.name); e == nil {
			if !bytes.Equal(old, a.data) {
				return fmt.Errorf("existing fresh bundle PNG differs: %s", a.name)
			}
			continue
		} else if !os.IsNotExist(e) {
			return e
		}
		if err = bundleRoot.MkdirAll(filepath.Dir(a.name), 0755); err != nil {
			return err
		}
		if err = bundleRoot.WriteFile(a.name, a.data, 0644); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	tmpName := ".retained-report-" + rand.Text() + ".json"
	tmp, err := bundleRoot.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = bundleRoot.Remove(tmpName) }()
	if _, err = tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return bundleRoot.Rename(tmpName, "report.json")
}
