package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

func TestRetainedInterfaceBuildValidateAndRepeat(t *testing.T) {
	source := assetsRoot
	root := t.TempDir()
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	m, err := build(source, a, 2)
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != "candidate" || len(m.Entries) != 117 {
		t.Fatalf("unexpected runtime catalog: %s %d", m.Status, len(m.Entries))
	}
	ids := make(map[string]bool)
	for _, e := range m.Entries {
		ids[e.ID] = true
	}
	if len(ids) != 117 {
		t.Fatal("lost enabled stable IDs", len(ids))
	}
	if _, err := build(source, b, 2); err != nil {
		t.Fatal(err)
	}
	files := 0
	if err := fs.WalkDir(os.DirFS(a), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		left, err := os.ReadFile(filepath.Join(a, name))
		if err != nil {
			return err
		}
		right, err := os.ReadFile(filepath.Join(b, name))
		if err != nil {
			return err
		}
		if !bytes.Equal(left, right) {
			t.Errorf("nonrepeatable output %s", name)
		}
		if name != "manifest.json" && !strings.HasSuffix(name, ".png") {
			t.Errorf("non-runtime file copied: %s", name)
		}
		files++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if files != 118 {
		t.Fatal("unexpected file count", files)
	}
	var output bytes.Buffer
	if err := run([]string{"validate", "-root", a, "-allow-candidates"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "117 PNG variants; status=candidate") {
		t.Fatal("missing validation result", output.String())
	}
	if err := run([]string{"validate", "-root", a}, &output); err == nil || !strings.Contains(err.Error(), "opt-in") {
		t.Fatal("candidate silently accepted", err)
	}
	before, err := os.ReadFile(filepath.Join(a, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := build(source, a, 2); err == nil {
		t.Fatal("existing pack overwritten")
	}
	after, err := os.ReadFile(filepath.Join(a, "manifest.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed rebuild changed existing pack", err)
	}
	// Output byte corruption must fail validation through the public command.
	path := filepath.Join(a, m.Entries[0].Path)
	if err := os.WriteFile(path, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"validate", "-root", a, "-allow-candidates"}, &output); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatal("corruption passed CLI", err)
	}
}

func TestEveryDensityPreservesStableIDsAndSourceBytes(t *testing.T) {
	source := assetsRoot
	index, err := readIndex(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{1, 2, 4} {
		original, err := interfacePack(index, density)
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(t.TempDir(), "pack")
		m, err := build(source, out, density)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Entries) != 117 {
			t.Fatal("missing density variants", density, len(m.Entries))
		}
		for i, e := range m.Entries {
			before := original.Entries[i]
			if e.ID != before.ID || e.Density != density || e.Anchor != before.Anchor {
				t.Fatal("stable identity or registration changed", e)
			}
			sourceBytes, err := os.ReadFile(filepath.Join(source, before.Path))
			if err != nil {
				t.Fatal(err)
			}
			outputBytes, err := os.ReadFile(filepath.Join(out, e.Path))
			if err != nil || !bytes.Equal(sourceBytes, outputBytes) {
				t.Fatal("runtime build changed source PNG bytes", e.ID, density, err)
			}
		}
	}
}

func TestBuildFailureLeavesNoOutput(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "pack")
	pin := strings.Repeat("0", 64)
	for _, index := range []string{
		`{"schema_version":2,"assets":[]}`,
		`{"schema_version":1,"assets":[]}`,
		`{"schema_version":1,"assets":[{"id":"bad","pack":"interface-v1-2x","density":2,"path":"../escape-2x.png","sha256":"` + pin + `","width":2,"height":2,"logical_anchor":[0,0],"art":"interface/bad"}]}`,
		`{"schema_version":1,"assets":[{"id":"bad","pack":"interface-v1-2x","density":2,"path":"bad-2x.png","sha256":"` + pin + `","width":2,"height":2,"logical_anchor":[0,0],"art":"interface/bad"}]}`,
	} {
		if err := os.WriteFile(filepath.Join(source, "index.json"), []byte(index), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := build(source, out, 2); err == nil {
			t.Fatal("bad source accepted")
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatal("invalid build left output", err)
		}
	}
}

func TestCommandErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"review"}, {"build"}, {"build", "-unknown"}, {"validate", "extra"}, {"validate", "-unknown"}} {
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Fatal("invalid command succeeded", args)
		}
	}
	if _, err := interfacePack(nil, 2); err == nil {
		t.Fatal("empty index accepted")
	}
	if _, err := interfacePack(nil, 3); err == nil {
		t.Fatal("unsupported density accepted")
	}
	anchor := [2]float64{}
	partial := []indexEntry{
		{ID: "kept", Pack: "interface-v1-2x", Density: 2, Path: "kept-2x.png", Anchor: &anchor},
		{ID: "missing", Pack: "interface-v1-1x", Density: 1, Path: "missing-1x.png", Anchor: &anchor},
	}
	if _, err := interfacePack(partial, 2); err == nil || !strings.Contains(err.Error(), "missing requested density") {
		t.Fatal("missing density silently omitted", err)
	}
	// The runtime manifest is a separate format; exporter metadata is not accepted.
	if _, err := assets.ParseManifest([]byte(`{"schema_version":1,"profile":"interface-v1"}`)); err == nil {
		t.Fatal("export report accepted as runtime manifest")
	}
}
