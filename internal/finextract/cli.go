package finextract

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
)

// Run extracts local original maps as versioned level JSON for inspection.
func Run(args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl levels extract", flag.ContinueOnError)
	f.SetOutput(stderr)
	original := f.String("original", "", "original/ directory (default: found above the working directory)")
	out := f.String("out", "", "output directory (default: build/scratch/levels in the repository)")
	names := f.String("build", "full,demo", "comma-separated builds to import")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if *original == "" {
		o, err := origmaps.Find(".")
		if err != nil {
			return err
		}
		*original = o
	}
	if *out == "" {
		root, err := repoRoot(".")
		if err != nil {
			return err
		}
		*out = filepath.Join(root, "build", "scratch", "levels")
	}
	builds, err := selectBuilds(origmaps.Builds(*original), *names)
	if err != nil {
		return err
	}
	n, err := extract(builds, *out, stdout)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "wrote %d levels to %s\n", n, *out)
	return err
}
