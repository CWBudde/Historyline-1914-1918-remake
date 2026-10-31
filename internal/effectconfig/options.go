// Package effectconfig defines native preview options without graphics dependencies.
package effectconfig

import (
	"flag"
	"fmt"
	"io"
)

// Options configures the native unit movement and effects preview.
type Options struct {
	Repo, Mode, Family, Out string
	SelfCheck               bool
}

// Parse reads preview flags and rejects unsupported modes and positional arguments.
func Parse(args []string, stderr io.Writer) (Options, error) {
	var o Options
	f := flag.NewFlagSet("hl preview effects", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&o.Repo, "repo", ".", "repository containing offline candidates")
	f.BoolVar(&o.SelfCheck, "self-check", false, "save bounded native screenshots and evidence, then exit")
	f.StringVar(&o.Mode, "mode", "normal", "normal, reduced or skip")
	f.StringVar(&o.Family, "family", "infantry", "delivered unit-family slug")
	f.StringVar(&o.Out, "out", "build/review/effect-preview", "fresh output directory for self-check evidence")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	if f.NArg() != 0 || (o.Mode != "normal" && o.Mode != "reduced" && o.Mode != "skip") {
		return o, fmt.Errorf("requires -mode normal|reduced|skip and no positional arguments")
	}
	return o, nil
}
