package artreview

import (
	"io"
	"strings"
	"testing"
)

// A missing dispatcher or a value-based exclusivity check must fail these cases.
func TestUnitCLIExclusiveExplicitFlags(t *testing.T) {
	for _, args := range [][]string{
		{"-unit", "", "-out", "new"},
		{"-unit", "m.json"},
		{"-unit", "m.json", "-out", "new", "-hills", ""},
		{"-unit", "m.json", "-out", "new", "-reference", ""},
		{"-unit", "m.json", "-out", "new", "positional"},
	} {
		err := run(args, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "-unit requires -out and excludes") {
			t.Errorf("%v: got %v", args, err)
		}
	}
}
