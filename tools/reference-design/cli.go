package referencedesign

import (
	"flag"
	"fmt"
	"io"
)

// Run checks the reference design or renders its review and measurement reports.
func Run(args []string, stdout, stderr io.Writer) error {
	var o options
	f := flag.NewFlagSet("hl review reference", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&o.root, "root", ".", "repository root that reference paths resolve against")
	f.StringVar(&o.ref, "ref", "art/reference/pulse/reference.json", "reference data, relative to -root")
	f.BoolVar(&o.check, "check", false, "validate reference data and targets (default without output flags)")
	f.BoolVar(&o.dump, "dump-cells", false, "print terrain and infrastructure of every cell")
	f.BoolVar(&o.drift, "drift", false, "with -check: print render metrics against targets per family")
	f.StringVar(&o.out, "out", "", "render tiles, comparison, index.html and targets.json")
	f.StringVar(&o.lookOut, "look-out", "", "write partial Look 0 measurements and provenance to DIR/look.json")
	f.BoolVar(&o.lookCheck, "look-check", false, "with -look-out: fail unless all required acceptance rules pass")
	f.StringVar(&o.compare, "compare", "", "write only comparison bands, crops and compare.json; skipped without image")
	f.StringVar(&o.compareFormat, "compare-format", "jpg", "comparison image format: png or jpg")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	o.stdout = stdout
	return run(o)
}
