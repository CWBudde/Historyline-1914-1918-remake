package artreview

// Seasonal exports. Autumn and spring change only how terrain looks: with
// -season a terrain mode runs its summer export unchanged on the seasonal
// edits of its sources and suffixes every runtime ID and runtime file name
// (_au, _sp). Geometry, masks, context keys and ports stay those of summer.
//
// A seasonal edit lies beside the source it repaints and differs only in the
// season word: art/terrain/grass/grass-summer-v1.png is repainted as
// grass-autumn-v<N>.png, the newest version taken. Summer and shared sources
// with an edit are substituted (the pinned summer bytes are checked first and
// the edit must keep their dimensions); the others, and every winter source,
// are used as pinned. Reports record each substituted source with its SHA-256.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// exportSeason is the -season flag: empty, autumn or spring.
var exportSeason string

var seasonSuffixes = map[string]string{"autumn": "_au", "spring": "_sp"}

// seasonModes are the flags of the terrain modes that accept -season.
var seasonModes = map[string]bool{"connected": true, "mixed": true, "shore": true, "compound": true, "ground": true, "forest": true, "forest-layered": true, "hills": true}

var seasonSourceName = regexp.MustCompile(`^(.+)-(summer|winter|autumn|spring|shared)-v([1-9][0-9]*)\.png$`)

func seasonActive() bool { return exportSeason != "" }

// seasonID is id with the runtime suffix of the export season.
func seasonID(id string) string { return id + seasonSuffixes[exportSeason] }

// seasonExports narrows the seasons a combined summer/winter exporter writes
// to summer in a seasonal export.
func seasonExports(seasons ...string) []string {
	if seasonActive() {
		return []string{"summer"}
	}
	return seasons
}

// seasonExported reports whether an export writes the forms of season: all
// of them normally, summer alone in a seasonal export.
func seasonExported(season string) bool { return !seasonActive() || season == "summer" }

// seasonReport marks a seasonal bundle: its runtime files and IDs carry
// IDSuffix, while context keys and descriptors stay those of summer.
type seasonReport struct {
	Season   string `json:"season,omitempty"`
	IDSuffix string `json:"id_suffix,omitempty"`
}

func currentSeasonReport() seasonReport {
	return seasonReport{exportSeason, seasonSuffixes[exportSeason]}
}

// seasonRef returns ref with its source replaced by the newest seasonal edit
// beside it, pinned by that edit's SHA-256, and whether it was replaced.
func seasonRef(base string, ref connectionMaterial) (connectionMaterial, bool, error) {
	if !seasonActive() {
		return ref, false, nil
	}
	dir, name := path.Split(filepath.ToSlash(ref.Source))
	m := seasonSourceName.FindStringSubmatch(name)
	if m == nil || m[2] != "summer" && m[2] != "shared" {
		return ref, false, nil
	}
	entries, err := os.ReadDir(filepath.Join(base, filepath.FromSlash(dir)))
	if err != nil {
		return ref, false, err
	}
	edit, version := "", 0
	for _, e := range entries {
		em := seasonSourceName.FindStringSubmatch(e.Name())
		if em == nil || em[1] != m[1] || em[2] != exportSeason {
			continue
		}
		if v, _ := strconv.Atoi(em[3]); v > version {
			edit, version = e.Name(), v
		}
	}
	if edit == "" {
		return ref, false, nil
	}
	pinned, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(ref.Source)))
	if err != nil {
		return ref, false, err
	}
	if sum := sha256.Sum256(pinned); hex.EncodeToString(sum[:]) != strings.ToLower(ref.SourceSHA256) {
		return ref, false, fmt.Errorf("%s: pinned source SHA256 mismatch", ref.Source)
	}
	data, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(dir+edit)))
	if err != nil {
		return ref, false, err
	}
	a, err := png.DecodeConfig(bytes.NewReader(pinned))
	if err != nil {
		return ref, false, fmt.Errorf("%s: %w", ref.Source, err)
	}
	b, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ref, false, fmt.Errorf("%s: %w", dir+edit, err)
	}
	if a.Width != b.Width || a.Height != b.Height {
		return ref, false, fmt.Errorf("%s is %dx%d, its %s edit %s %dx%d: a seasonal edit keeps the source's geometry", ref.Source, a.Width, a.Height, exportSeason, edit, b.Width, b.Height)
	}
	sum := sha256.Sum256(data)
	return connectionMaterial{Source: dir + edit, SourceSHA256: hex.EncodeToString(sum[:])}, true, nil
}

// seasonPreflight rejects a seasonal export of a manifest when neither it nor
// a manifest it pins has a seasonal edit of any source.
func seasonPreflight(manifest string) error {
	n, err := seasonEdits(manifest, map[string]bool{})
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("-season %s: no %s edit (<name>-%s-v<N>.png) beside any summer or shared source of %s", exportSeason, exportSeason, exportSeason, manifest)
	}
	return nil
}

func seasonEdits(manifest string, seen map[string]bool) (int, error) {
	data, err := os.ReadFile(manifest)
	if err != nil {
		return 0, err
	}
	var root any
	if err = json.Unmarshal(data, &root); err != nil {
		return 0, fmt.Errorf("%s: %w", manifest, err)
	}
	base := sourceBase(manifest)
	count := 0
	var walk func(any) error
	walk = func(v any) error {
		switch v := v.(type) {
		case map[string]any:
			if source, ok := v["source"].(string); ok && source != "" && !filepath.IsAbs(source) {
				pin, _ := v["source_sha256"].(string)
				if strings.HasSuffix(source, ".json") {
					p := filepath.Join(base, filepath.FromSlash(source))
					if !seen[p] {
						seen[p] = true
						n, err := seasonEdits(p, seen)
						if err != nil {
							return err
						}
						count += n
					}
				} else if _, ok, err := seasonRef(base, connectionMaterial{source, pin}); err != nil {
					return err
				} else if ok {
					count++
				}
			}
			for _, x := range v {
				if err := walk(x); err != nil {
					return err
				}
			}
		case []any:
			for _, x := range v {
				if err := walk(x); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return count, walk(root)
}

// errSeasonWinter rejects a seasonal export of a winter-only manifest.
var errSeasonWinter = errors.New("-season exports the summer forms; this manifest has only winter ones")
