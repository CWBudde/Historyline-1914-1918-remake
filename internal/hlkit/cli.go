package hlkit

import (
	"flag"
	"fmt"
	"io"
)

// Run prepares the selected world kits, optionally installing their runtime cells.
func Run(args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl export kits", flag.ContinueOnError)
	f.SetOutput(stderr)
	profile := f.String("profile", "art/pipeline/world-kits-v1.json", "pinned source profile")
	out := f.String("out", "build/review/world-kits", "output directory")
	strict := f.Bool("check", false, "fail unless every kit check passes (the report is written anyway)")
	assetsDir := f.String("install", "", "also install runtime cells and shadows into this assets tree")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if *assetsDir != "" {
		return install(*profile, *out, *assetsDir)
	}
	return run(*profile, *out, *strict)
}
