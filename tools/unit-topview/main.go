// Command unit-topview exports a unit family drawn from straight above, such as
// the bunker: one master for all six facings (a top view does not turn, so its
// marking stays upright). The neutral master is the base; each owner's master is
// an edit of it that paints the national insignia, of which only the centre disc
// is inlaid into the base, so the outline never changes between owners.
//
//	go run ./tools/unit-topview -manifest art/pipeline/unit-bunker-v2.json
//
// writes assets/units/<family>/<palette>/<facing>-<state>-<d>x.png and re-pins
// their assets/index.json entries; -preview DIR (under build/) also renders the
// family on the map.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

type source struct {
	Source string `json:"source"`
	SHA256 string `json:"source_sha256"`
}

// owner is how one palette marks the base: unmarked (neutral), or the disc of
// radius logical pixels around the centre taken from the inlay master.
type owner struct {
	Radius float64 `json:"radius,omitempty"`
	Inlay  *source `json:"inlay,omitempty"`
}

// manifest is art/pipeline/unit-<family>-v<N>.json. The base's solid box (alpha
// above 128, after clearing alpha below alpha_cut) is centred on the anchor and
// its larger side scaled to diameter logical pixels.
type manifest struct {
	Schema    int              `json:"schema_version"`
	Family    string           `json:"family"`
	Base      source           `json:"base"`
	AlphaCut  uint8            `json:"alpha_cut"`
	Diameter  float64          `json:"diameter"`
	Owners    map[string]owner `json:"owners"`
	Densities []int            `json:"densities"`
}

func readManifest(root, path string) (manifest, *image.NRGBA, error) {
	var m manifest
	b, err := os.ReadFile(path)
	if err != nil {
		return m, nil, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return m, nil, fmt.Errorf("%s: %w", path, err)
	}
	if m.Schema != 1 {
		return m, nil, fmt.Errorf("%s: schema_version %d, want 1", path, m.Schema)
	}
	if !slices.Contains(unitsprite.AllFamilies(), m.Family) {
		return m, nil, fmt.Errorf("%s: no sprite family %q", path, m.Family)
	}
	for _, p := range unitsprite.Palettes {
		if _, ok := m.Owners[string(p)]; !ok {
			return m, nil, fmt.Errorf("%s: no owner %s", path, p)
		}
	}
	if len(m.Owners) != len(unitsprite.Palettes) {
		return m, nil, fmt.Errorf("%s: owners must be exactly %v", path, unitsprite.Palettes)
	}
	if m.Diameter <= 0 || m.Diameter > logicalSize || len(m.Densities) == 0 {
		return m, nil, fmt.Errorf("%s: diameter or densities out of range", path)
	}
	base, err := readSource(root, m.Base, m.AlphaCut)
	return m, base, err
}

// readSource decodes a pinned master and clears its alpha haze.
func readSource(root string, s source, cut uint8) (*image.NRGBA, error) {
	raw, err := os.ReadFile(filepath.Join(root, s.Source))
	if err != nil {
		return nil, err
	}
	if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != s.SHA256 {
		return nil, fmt.Errorf("%s: sha256 %x, manifest pins %s", s.Source, sum, s.SHA256)
	}
	im, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", s.Source, err)
	}
	return cutHaze(im, cut), nil
}

// build returns every runtime file of the family by its path below assets/; it
// notes each inlay's registration on log.
func build(root string, m manifest, base *image.NRGBA, log io.Writer) (map[string]*image.NRGBA, error) {
	f, err := frameOf(base, m.Diameter)
	if err != nil {
		return nil, err
	}
	files := map[string]*image.NRGBA{}
	palettes := make([]string, 0, len(m.Owners))
	for p := range m.Owners {
		palettes = append(palettes, p)
	}
	sort.Strings(palettes)
	for _, p := range palettes {
		master := base
		if o := m.Owners[p]; o.Inlay != nil {
			edit, err := readSource(root, *o.Inlay, m.AlphaCut)
			if err != nil {
				return nil, err
			}
			r := o.Radius * f.scale
			if edit.Bounds() != base.Bounds() || r <= 0 || r > f.scale*m.Diameter/2 {
				return nil, fmt.Errorf("owner %s: the inlay must match the base's size and its radius lie inside the diameter", p)
			}
			var note string
			master, note = inlay(base, edit, f.centreX, f.centreY, r, 10)
			_, _ = fmt.Fprintf(log, "%s inlay: %s\n", p, note)
		}
		for _, d := range m.Densities {
			normal := render(master, f, d)
			states := map[unitsprite.State]*image.NRGBA{unitsprite.Normal: normal, unitsprite.Spent: spent(normal)}
			for face := range unitsprite.Facings {
				for st, im := range states {
					files[fmt.Sprintf("units/%s/%s/%s-%s-%dx.png", m.Family, p, face, st, d)] = im
				}
			}
		}
	}
	return files, nil
}

func run() error {
	path := flag.String("manifest", "", "top-view unit manifest (art/pipeline/unit-<family>-v<N>.json)")
	root := flag.String("root", ".", "repository root the manifest's sources are relative to")
	preview := flag.String("preview", "", "also render the installed family on the map into this directory under build/")
	flag.Parse()
	if *path == "" || flag.NArg() != 0 {
		return errors.New("usage: unit-topview -manifest FILE [-root DIR] [-preview build/DIR]")
	}
	m, base, err := readManifest(*root, *path)
	if err != nil {
		return err
	}
	files, err := build(*root, m, base, os.Stdout)
	if err != nil {
		return err
	}
	assetsDir := filepath.Join(*root, "assets")
	if err := install(assetsDir, files); err != nil {
		return err
	}
	fmt.Printf("installed %d files of unit/%s\n", len(files), m.Family)
	if *preview != "" {
		return renderPreview(*root, *preview, m.Family)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "unit-topview:", err)
		os.Exit(1)
	}
}
