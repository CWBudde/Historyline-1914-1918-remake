package candidatereview

import (
	"flag"
	"fmt"
	"io"
)

// Run renders raw candidates beside the installed art in their game context.
func Run(args []string, stdout, stderr io.Writer) error {
	var o options
	f := flag.NewFlagSet("hl review candidates", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&o.root, "root", ".", "repository root")
	f.StringVar(&o.catalogue, "catalogue", "build/review/look-audit/agent-inventory/catalogue.tsv", "candidate catalogue (TSV)")
	f.StringVar(&o.raw, "raw", "build/review/candidates/raw", "local copy of raw candidates, laid out <session>/<exec>.png")
	f.StringVar(&o.out, "out", "build/review/candidates/site", "output directory under build/")
	f.StringVar(&o.only, "only", "", "comma-separated kinds: unit, portrait, interface, screen, terrain, other")
	f.StringVar(&o.decisions, "terrain-decisions", "build/review/candidates/terrain-decisions.tsv", "optional terrain decisions: exec, keep|delete, reason (TSV)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	o.stderr = stderr
	return run(o)
}
