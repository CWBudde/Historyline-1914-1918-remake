package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pipelineManifest is the path of a committed bundle manifest.
func pipelineManifest(name string) string {
	return filepath.Join("..", "..", "art", "pipeline", name+".json")
}

// freshExport exports a committed manifest into a new temporary directory.
func freshExport(t *testing.T, name string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), name)
	if e := export(pipelineManifest(name), out); e != nil {
		t.Fatal(e)
	}
	return out
}

// runtimePins returns the SHA-256 of every runtime file in assets/index.json,
// keyed by "<kind>|<id>|<density>" (kind is the first part of its art key).
func runtimePins(t *testing.T) map[string]string {
	t.Helper()
	raw, e := os.ReadFile(filepath.Join("..", "..", "assets", "index.json"))
	if e != nil {
		t.Fatal(e)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Density int    `json:"density"`
			SHA256  string `json:"sha256"`
			Art     string `json:"art"`
		} `json:"assets"`
	}
	if e = json.Unmarshal(raw, &index); e != nil {
		t.Fatal(e)
	}
	pins := map[string]string{}
	for _, a := range index.Assets {
		pins[fmt.Sprintf("%s|%s|%d", strings.SplitN(a.Art, "/", 2)[0], a.ID, a.Density)] = a.SHA256
	}
	return pins
}

// checkRuntimePins requires every selected export file to be byte-identical to
// its runtime copy in assets/ and returns how many were compared.
func checkRuntimePins(t *testing.T, kind string, files []Output, selected func(Output) bool) int {
	t.Helper()
	pins := runtimePins(t)
	n := 0
	for _, f := range files {
		if !selected(f) {
			continue
		}
		key := fmt.Sprintf("%s|%s|%d", kind, f.ID, f.Density)
		if want, ok := pins[key]; !ok || want != f.SHA256 {
			t.Errorf("%s: export differs from the runtime asset (%s)", f.Path, key)
		}
		n++
	}
	return n
}
