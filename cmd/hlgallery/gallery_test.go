package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPublicGallery generates the real gallery into a temp copy of the layout
// and checks structure, image hashes, prompts and the absence of reference art.
func TestPublicGallery(t *testing.T) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	if err := root.MkdirAll("build/scratch/gallery-test", 0755); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join("../..", "build/scratch/gallery-test")
	// t.Cleanup runs after the deferred root.Close, so remove through the plain path.
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := run([]string{"-root", "../..", "-out", "build/scratch/gallery-test"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cat struct {
		Items []item `json:"items"`
	}
	if err := json.Unmarshal(raw, &cat); err != nil {
		t.Fatal(err)
	}
	if len(cat.Items) != 390 {
		t.Fatalf("got %d assets, want 390", len(cat.Items))
	}
	prompted := 0
	for _, it := range cat.Items {
		if it.Reference != nil {
			t.Fatalf("%s: reference image in public gallery", it.ID)
		}
		if len(it.Prompts) > 0 {
			prompted++
		}
		for _, p := range it.Prompts {
			if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
				t.Fatalf("%s: %v", it.ID, err)
			}
		}
		for _, v := range it.Variants {
			for d, pic := range v.Images {
				b, err := os.ReadFile(filepath.Join(dir, pic.URL))
				if err != nil {
					t.Fatal(err)
				}
				if fmt.Sprintf("%x", sha256.Sum256(b)) != pic.SHA {
					t.Fatalf("%s/%s/%s: hash mismatch", it.ID, v.ID, d)
				}
			}
		}
	}
	if prompted < 100 {
		t.Fatalf("only %d assets link a prompt", prompted)
	}
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.Contains(p, "/media/") {
			t.Errorf("content-addressed media in public gallery: %s", p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
