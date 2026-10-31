// Package unitpanel exports a deterministic infantry inspector art pilot.
package unitpanel

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// Run exports the unit inspector panel from a strict pinned manifest.
func Run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("hl export unit-panel", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifest := fs.String("manifest", "", "strict pinned manifest")
	out := fs.String("out", "", "fresh output directory")
	seen := map[string]bool{}
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			key := strings.SplitN(strings.TrimLeft(a, "-"), "=", 2)[0]
			if seen[key] {
				return fmt.Errorf("repeated flag %s", key)
			}
			seen[key] = true
		}
	}
	if e := fs.Parse(args); e != nil {
		return e
	}
	if fs.NArg() != 0 || *manifest == "" || *out == "" {
		return fmt.Errorf("usage: hl export unit-panel -manifest FILE -out FRESH_DIR")
	}
	in, e := loadInputs(*manifest)
	if e != nil {
		return e
	}
	return exportBundle(in, *out)
}
