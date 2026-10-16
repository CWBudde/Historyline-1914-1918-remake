// Command hlasset builds and validates local runtime candidate packs.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: hlasset build|build-units|validate [flags]")
	}
	switch args[0] {
	case "build":
		flags := flag.NewFlagSet("build", flag.ContinueOnError)
		source := flags.String("source", "art/exports/interface-v1", "interface-v1 export directory")
		out := flags.String("out", "", "new output directory (required)")
		density := flags.Int("density", 2, "exact asset density: 1, 2 or 4")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return fmt.Errorf("build requires -out NEW_DIRECTORY and no positional arguments")
		}
		p, err := build(*source, *out, *density)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "built %s: %d PNG variants (candidate; not runtime approved)\n", p.ID, len(p.Entries))
		return err
	case "build-units":
		flags := flag.NewFlagSet("build-units", flag.ContinueOnError)
		recipes := flags.String("recipes", "art/recipes", "directory with the unit-*-v1.json recipes")
		exports := flags.String("exports", "art/exports", "directory with the unit-*-v1 export directories")
		out := flags.String("out", "", "new output directory (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return fmt.Errorf("build-units requires -out NEW_DIRECTORY and no positional arguments")
		}
		packs, err := buildUnits(*recipes, *exports, *out)
		if err != nil {
			return err
		}
		families, densities, variants := map[string]bool{}, map[int]bool{}, 0
		for _, p := range packs {
			families[p.Family], densities[p.Density] = true, true
			variants += len(p.Manifest.Entries)
		}
		_, err = fmt.Fprintf(stdout, "built %d unit packs (%d families × %d densities, %d PNG variants; candidate, not runtime approved)\n",
			len(packs), len(families), len(densities), variants)
		return err
	case "validate":
		flags := flag.NewFlagSet("validate", flag.ContinueOnError)
		root := flags.String("root", ".", "pack filesystem root")
		manifest := flags.String("manifest", "manifest.json", "root-relative manifest path")
		candidates := flags.Bool("allow-candidates", false, "explicitly allow offline candidates")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return fmt.Errorf("validate accepts no positional arguments")
		}
		r, err := os.OpenRoot(*root)
		if err != nil {
			return err
		}
		defer func() { _ = r.Close() }()
		p, err := assets.Open(r.FS(), *manifest, assets.Options{AllowCandidates: *candidates})
		if err != nil {
			return err
		}
		if err := p.Validate(); err != nil {
			return err
		}
		m := p.Manifest()
		_, err = fmt.Fprintf(stdout, "validated %s: %d PNG variants; status=%s\n", m.ID, len(m.Entries), m.Status)
		return err
	default:
		return fmt.Errorf("unknown command %q; use build, build-units or validate", args[0])
	}
}

// The first adapter deliberately supports only the interface-v1 export report.
// Screen reviews, contact sheets and context-dependent terrain are not runtime art.
func importInterface(b []byte, density int) (assets.Manifest, error) {
	if density != 1 && density != 2 && density != 4 {
		return assets.Manifest{}, fmt.Errorf("density must be 1, 2 or 4")
	}
	var report struct {
		Schema  int    `json:"schema_version"`
		Profile string `json:"profile"`
		Files   []struct {
			ID      string     `json:"id"`
			Path    string     `json:"path"`
			SHA256  string     `json:"sha256"`
			Width   int        `json:"width"`
			Height  int        `json:"height"`
			Density int        `json:"density"`
			Anchor  [2]float64 `json:"logical_anchor"`
		} `json:"files"`
	}
	if len(b) > 2<<20 {
		return assets.Manifest{}, fmt.Errorf("export report too large")
	}
	if err := json.Unmarshal(b, &report); err != nil {
		return assets.Manifest{}, err
	}
	if report.Schema != 1 || report.Profile != "interface-v1" {
		return assets.Manifest{}, fmt.Errorf("build currently requires an interface-v1 schema-1 report")
	}
	m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("interface-v1-%dx", density), Status: "candidate"}
	allIDs := make(map[string]bool)
	selected := make(map[string]bool)
	for _, f := range report.Files {
		if !strings.HasSuffix(f.Path, ".png") || f.Density == 0 {
			continue
		}
		allIDs[f.ID] = true
		if f.Density == density {
			m.Entries = append(m.Entries, assets.Entry{ID: f.ID, Path: f.Path, SHA256: f.SHA256, Width: f.Width, Height: f.Height, Density: f.Density, Anchor: f.Anchor})
			selected[f.ID] = true
		}
	}
	// Follow report order so missing-ID diagnostics are deterministic too.
	for _, f := range report.Files {
		if allIDs[f.ID] && !selected[f.ID] {
			return assets.Manifest{}, fmt.Errorf("asset %q missing requested density %dx", f.ID, density)
		}
	}
	return m, nil
}

func build(source, out string, density int) (assets.Manifest, error) {
	r, err := os.OpenRoot(source)
	if err != nil {
		return assets.Manifest{}, err
	}
	defer func() { _ = r.Close() }()
	f, err := r.Open("report.json")
	if err != nil {
		return assets.Manifest{}, err
	}
	b, err := io.ReadAll(io.LimitReader(f, (2<<20)+1))
	closeErr := f.Close()
	if err != nil {
		return assets.Manifest{}, err
	}
	if closeErr != nil {
		return assets.Manifest{}, closeErr
	}
	m, err := importInterface(b, density)
	if err != nil {
		return assets.Manifest{}, err
	}
	p, err := assets.Load(r.FS(), m, assets.Options{AllowCandidates: true})
	if err != nil {
		return assets.Manifest{}, err
	}
	if err := p.Validate(); err != nil {
		return assets.Manifest{}, err
	}
	o, err := newOutput(out)
	if err != nil {
		return assets.Manifest{}, err
	}
	defer o.close()
	if m, err = o.writePack(".", p, m); err != nil {
		return assets.Manifest{}, err
	}
	o.done()
	return m, nil
}

// output is a build directory reserved exclusively. Until done is called, close
// removes only the paths this build created; existing user directories are never
// replaced.
type output struct {
	dir      string
	root     *os.Root
	created  []string // files and directories, in creation order
	complete bool
}

func newOutput(dir string) (*output, error) {
	if err := os.Mkdir(dir, 0o755); err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(dir)
	if err != nil {
		_ = os.Remove(dir)
		return nil, err
	}
	return &output{dir: dir, root: r}, nil
}

func (o *output) done() { o.complete = true }

func (o *output) close() {
	if !o.complete {
		for i := len(o.created) - 1; i >= 0; i-- {
			_ = o.root.Remove(filepath.FromSlash(o.created[i]))
		}
	}
	_ = o.root.Close()
	if !o.complete {
		_ = os.Remove(o.dir)
	}
}

// mkdirAll creates the missing directories of the slash path dir.
func (o *output) mkdirAll(dir string) error {
	if dir == "." || slices.Contains(o.created, dir) {
		return nil
	}
	if err := o.mkdirAll(path.Dir(dir)); err != nil {
		return err
	}
	if err := o.root.Mkdir(filepath.FromSlash(dir), 0o755); err != nil {
		return err
	}
	o.created = append(o.created, dir)
	return nil
}

// writePack writes p's exact PNG snapshots and manifest m under dir. A flat
// images/NNNN.png layout avoids collisions with metadata and source directories.
// It returns m with the new paths.
func (o *output) writePack(dir string, p *assets.Pack, m assets.Manifest) (assets.Manifest, error) {
	m.Entries = slices.Clone(m.Entries)
	if err := o.mkdirAll(path.Join(dir, "images")); err != nil {
		return m, err
	}
	for i := range m.Entries {
		e := &m.Entries[i]
		name := path.Join(dir, fmt.Sprintf("images/%04d.png", i))
		if err := o.writeNew(name, func(w io.Writer) error { return p.WritePNG(w, e.ID, e.Density) }); err != nil {
			return m, err
		}
		e.Path = fmt.Sprintf("images/%04d.png", i)
	}
	manifestBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return m, err
	}
	err = o.writeNew(path.Join(dir, "manifest.json"), func(w io.Writer) error {
		_, err := w.Write(append(manifestBytes, '\n'))
		return err
	})
	return m, err
}

func (o *output) writeNew(name string, write func(io.Writer) error) error {
	if err := writeNew(o.root, name, write); err != nil {
		return err
	}
	o.created = append(o.created, name)
	return nil
}

func writeNew(root *os.Root, name string, write func(io.Writer) error) error {
	f, err := root.OpenFile(filepath.FromSlash(name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	err = write(f)
	closeErr := f.Close()
	if err != nil {
		_ = root.Remove(name)
		return err
	}
	if closeErr != nil {
		_ = root.Remove(name)
	}
	return closeErr
}
