package hlcampaign

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Run writes the original maps' campaign folders, keeping backdrops, front pages
// and a map's season override.
func Run(args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl levels campaign", flag.ContinueOnError)
	f.SetOutput(stderr)
	dir := f.String("levels", "levels", "the levels directory (map and setup files are overwritten)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	set, err := data.Load()
	if err != nil {
		return err
	}
	maps, err := origmaps.Full()
	if err != nil {
		return err
	}
	out, err := generate(maps, set)
	if err != nil {
		return err
	}
	for _, line := range out.report {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return err
		}
	}
	names := make([]string, 0, len(out.folders))
	for name := range out.folders {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		folder := out.folders[name]
		if old, err := scenario.Load(filepath.Join(*dir, name)); err == nil && !folder.Winter() {
			folder.MapSeason = old.MapSeason
		}
		if err := folder.Write(filepath.Join(*dir, name)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	_, err = fmt.Fprintf(stdout, "wrote %d level folders to %s\n", len(out.folders), *dir)
	return err
}
