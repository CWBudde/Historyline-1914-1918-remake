// Command hlasset builds and validates local runtime candidate packs from the
// runtime art in assets/ (assets/index.json).
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

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: hlasset build|build-units|build-terrain|build-screens|lite|terrain-gaps|validate [flags]")
	}
	switch args[0] {
	case "build":
		flags := flag.NewFlagSet("build", flag.ContinueOnError)
		dir := flags.String("assets", "assets", "runtime asset directory with index.json")
		out := flags.String("out", "", "new output directory (required)")
		density := flags.Int("density", 2, "exact asset density: 1, 2 or 4")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return fmt.Errorf("build requires -out NEW_DIRECTORY and no positional arguments")
		}
		p, err := build(*dir, *out, *density)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "built %s: %d PNG variants (candidate; not runtime approved)\n", p.ID, len(p.Entries))
		return err
	case "build-units":
		flags := flag.NewFlagSet("build-units", flag.ContinueOnError)
		dir := flags.String("assets", "assets", "runtime asset directory with index.json")
		out := flags.String("out", "", "new output directory (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return fmt.Errorf("build-units requires -out NEW_DIRECTORY and no positional arguments")
		}
		packs, err := buildUnits(*dir, *out)
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
	case "build-terrain":
		flags := flag.NewFlagSet("build-terrain", flag.ContinueOnError)
		dir := flags.String("assets", "assets", "runtime asset directory with index.json")
		out := flags.String("out", "", "new output directory (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return fmt.Errorf("build-terrain requires -out NEW_DIRECTORY and no positional arguments")
		}
		packs, err := buildTerrain(*dir, *out)
		if err != nil {
			return err
		}
		tiles := 0
		for _, p := range packs {
			tiles += len(p.Manifest.Entries)
		}
		_, err = fmt.Fprintf(stdout, "built %d terrain packs (%d PNG variants; candidate, not runtime approved)\n", len(packs), tiles)
		return err
	case "build-screens":
		flags := flag.NewFlagSet("build-screens", flag.ContinueOnError)
		dir := flags.String("assets", "assets", "runtime asset directory with index.json")
		art := flags.String("art", "art", "art directory with the scene paintings (screens/<family>/illustration.png)")
		out := flags.String("out", "", "new output directory (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *out == "" {
			return fmt.Errorf("build-screens requires -out NEW_DIRECTORY and no positional arguments")
		}
		packs, err := buildScreens(*dir, *art, *out)
		if err != nil {
			return err
		}
		variants := 0
		for _, p := range packs {
			variants += len(p.Entries)
		}
		_, err = fmt.Fprintf(stdout, "built %d screen packs (%d PNG variants; candidate, not runtime approved)\n", len(packs), variants)
		return err
	case "lite":
		flags := flag.NewFlagSet("lite", flag.ContinueOnError)
		in := flags.String("in", "", "pack catalog built by build-units, build-terrain or build-screens (required)")
		out := flags.String("out", "", "new output directory (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *in == "" || *out == "" {
			return fmt.Errorf("lite requires -in DIRECTORY -out NEW_DIRECTORY and no positional arguments")
		}
		st, err := liteCatalog(*in, *out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "compressed %s\n", st)
		return err
	case "terrain-gaps":
		flags := flag.NewFlagSet("terrain-gaps", flag.ContinueOnError)
		dir := flags.String("terrain", "", "terrain catalog built by build-terrain (required)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *dir == "" {
			return fmt.Errorf("terrain-gaps requires -terrain DIRECTORY and no positional arguments")
		}
		return terrainGaps(*dir, stdout)
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
		return fmt.Errorf("unknown command %q; use build, build-units, build-terrain, build-screens, lite, terrain-gaps or validate", args[0])
	}
}

// interfacePack selects the interface pack of one density from the index. Every
// interface ID must exist at that density.
func interfacePack(entries []indexEntry, density int) (assets.Manifest, error) {
	if density != 1 && density != 2 && density != 4 {
		return assets.Manifest{}, fmt.Errorf("density must be 1, 2 or 4")
	}
	id := fmt.Sprintf("interface-v1-%dx", density)
	packs := indexPacks(entries, "interface-v1-")
	i := slices.IndexFunc(packs, func(m assets.Manifest) bool { return m.ID == id })
	if i < 0 {
		return assets.Manifest{}, fmt.Errorf("index has no %s pack", id)
	}
	m := packs[i]
	have := map[string]bool{}
	for _, e := range m.Entries {
		if e.Density != density {
			return assets.Manifest{}, fmt.Errorf("%s: %s has density %dx", id, e.ID, e.Density)
		}
		have[e.ID] = true
	}
	// Follow index order so missing-ID diagnostics are deterministic.
	for _, p := range packs {
		for _, e := range p.Entries {
			if !have[e.ID] {
				return assets.Manifest{}, fmt.Errorf("asset %q missing requested density %dx", e.ID, density)
			}
		}
	}
	return m, nil
}

// build writes the interface pack of one density to the new directory out.
func build(assetsDir, out string, density int) (assets.Manifest, error) {
	entries, err := readIndex(assetsDir)
	if err != nil {
		return assets.Manifest{}, err
	}
	m, err := interfacePack(entries, density)
	if err != nil {
		return assets.Manifest{}, err
	}
	written, err := writePacks(assetsDir, out, []assets.Manifest{m}, func(int) string { return "." })
	if err != nil {
		return assets.Manifest{}, err
	}
	return written[0], nil
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
