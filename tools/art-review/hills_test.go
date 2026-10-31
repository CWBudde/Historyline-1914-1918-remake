package artreview

import (
	"bytes"
	"strings"
	"testing"
)

// Dispatching an empty or competing flag to another exporter violates the mode boundary.
func TestHillsModeBoundary(t *testing.T) {
	for _, args := range [][]string{
		{"-hills", "", "-out", "unused"},
		{"-hills", "missing.json"},
		{"-hills", "missing.json", "-out", "unused", "extra"},
		{"-hills", "missing.json", "-out", "unused", "-forest", ""},
		{"-hills", "missing.json", "-out", "unused", "-ground", "x"},
		{"-hills", "missing.json", "-out", "unused", "-resampling", "area"},
		{"-hills", "missing.json", "-out", "unused", "-reference", ""},
	} {
		var log bytes.Buffer
		err := run(args, &log)
		if err == nil || !strings.Contains(err.Error(), "-hills requires") {
			t.Fatalf("mode boundary %v: %v; %s", args, err, &log)
		}
	}
}
