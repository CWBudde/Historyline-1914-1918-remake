package hlwebapp

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSize(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "blobs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, size := range map[string]int{"index.html": 1_000_000, "blobs/a.png": 1_500_000} {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(limit string) (string, error) {
		var out, errs bytes.Buffer
		err := Size(context.Background(), []string{"-dir", dir, "-max-mb", limit}, &out, &errs)
		return out.String(), err
	}
	out, err := run("3")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"html", "png", "total", "2.5 MB"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if _, err := run("2"); err == nil || !strings.Contains(err.Error(), "above the 2 MB limit") {
		t.Errorf("over the limit: %v", err)
	}
	var out2, errs bytes.Buffer
	if err := Size(context.Background(), []string{"-dir", t.TempDir()}, &out2, &errs); err == nil {
		t.Error("an empty site must fail")
	}
}
