package hlmaterial

import (
	"flag"
	"fmt"
	"io"
)

// Run exports the selected reusable periodic world materials.
func Run(args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl export materials", flag.ContinueOnError)
	f.SetOutput(stderr)
	profile := f.String("profile", "art/pipeline/world-materials-v1.json", "pinned source profile")
	out := f.String("out", "build/review/world-materials", "output directory (paths match assets/)")
	strict := f.Bool("check", false, "fail unless every meadow check passes (the report is written anyway)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	return run(*profile, *out, *strict)
}
