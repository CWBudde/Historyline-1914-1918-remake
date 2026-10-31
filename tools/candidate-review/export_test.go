package candidatereview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnitExportValidatesWithoutGoToolchain(t *testing.T) {
	root := t.TempDir()
	manifestDir := filepath.Join(root, "art", "pipeline")
	if err := os.MkdirAll(manifestDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestDir, "unit-infantry-v1.json"), []byte(`{"schema_version":999,"family":"infantry","palettes":{},"grounds":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	raw := filepath.Join(root, "raw.png")
	if err := os.WriteFile(raw, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", "")
	r := renderer{o: options{root: root}, work: filepath.Join(root, "build", "review", "work")}
	_, reason := r.exportUnit("infantry", &candidate{Exec: "test", raw: raw})
	if !strings.Contains(reason, "schema") {
		t.Fatalf("expected exporter validation without a Go subprocess, got %q", reason)
	}
}
