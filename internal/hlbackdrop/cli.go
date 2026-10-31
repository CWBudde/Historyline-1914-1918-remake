package hlbackdrop

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

// Run encodes or checks the backdrops' JPEGs against their lossless sources.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || (args[0] != "jpeg" && args[0] != "check") {
		return fmt.Errorf("usage: hl levels backdrop jpeg|check [flags]")
	}
	f := flag.NewFlagSet("hl levels backdrop "+args[0], flag.ContinueOnError)
	f.SetOutput(stderr)
	levels := f.String("levels", "levels", "the levels directory")
	var q int
	if args[0] == "jpeg" {
		f.IntVar(&q, "q", 0, "JPEG quality (default: the existing JPEG's own, 95 for a new one)")
	}
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if args[0] == "check" {
		if f.NArg() != 0 {
			return fmt.Errorf("check takes no positional arguments")
		}
		return errors.Join(check(*levels)...)
	}
	if f.NArg() == 0 {
		return fmt.Errorf("jpeg requires one or more level names")
	}
	for _, name := range f.Args() {
		b, err := find(*levels, name)
		if err != nil {
			return err
		}
		n, used, err := writeJPEG(b, q)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "%s: %.2f MB at quality %d\n", b.out, float64(n)/1e6, used); err != nil {
			return err
		}
	}
	return nil
}
