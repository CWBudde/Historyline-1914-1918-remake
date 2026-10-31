package artreview

import "io"

// Run exports an offline art comparison or a manifest-driven candidate bundle.
func Run(args []string, stdout, stderr io.Writer) error {
	return run(args, stderr)
}
