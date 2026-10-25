// Command unit-panel exports a deterministic infantry inspector art pilot.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func run(args []string) error {
	fs := flag.NewFlagSet("unit-panel", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
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
		return fmt.Errorf("usage: unit-panel -manifest FILE -out FRESH_DIR")
	}
	in, e := loadInputs(*manifest)
	if e != nil {
		return e
	}
	return exportBundle(in, *out)
}

func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
