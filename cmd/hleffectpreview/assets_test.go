package main

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCandidateBundleAndTamperRejection(t *testing.T) {
	repo := filepath.Join("..", "..")
	a, err := loadAssets(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.animations) != 4 || len(a.images) != 22 {
		t.Fatalf("unexpected coverage: %d animations, %d images", len(a.animations), len(a.images))
	}
	// Copy exactly the bytes used by the preview, with no source/reference art.
	root := t.TempDir()
	for path := range a.hashes {
		data, err := os.ReadFile(filepath.Join(repo, path))
		if err != nil {
			t.Fatal(err)
		}
		dest := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := loadAssets(root); err != nil {
		t.Fatal(err)
	}
	frame := filepath.Join(root, "art/exports/interface-v1/effect/hit-1x.png")
	original, err := os.ReadFile(frame)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(frame, append(original, 'x'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAssets(root); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("tampered pinned frame accepted: %v", err)
	}
	if err := os.WriteFile(frame, original, 0o600); err != nil {
		t.Fatal(err)
	}
	var tiny bytes.Buffer
	if err := png.Encode(&tiny, image.NewNRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	unit := filepath.Join(root, "art/exports/unit-infantry-v1/sprites/infantry/cp/n-normal-2x.png")
	if err := os.WriteFile(unit, tiny.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAssets(root); err == nil || !strings.Contains(err.Error(), "bad dimensions") {
		t.Fatalf("bad unit dimensions accepted: %v", err)
	}
}

func TestRejectEscapingCandidatePaths(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "art/exports/interface-v1")
	if err := os.MkdirAll(dest, 0o700); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("../../art/exports/interface-v1/report.json")
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte(`"path": "action/attack-disabled-1x.png"`), []byte(`"path": "../legacy/forbidden.png"`), 1)
	if err := os.WriteFile(filepath.Join(dest, "report.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAssets(root); err == nil || !strings.Contains(err.Error(), "invalid or duplicate interface path") {
		t.Fatalf("escaping path accepted: %v", err)
	}
}

func TestAllDeliveredFamiliesShareSupportedGeometry(t *testing.T) {
	seen := make(map[string]bool)
	for _, family := range families {
		t.Run(family.slug, func(t *testing.T) {
			if seen[family.slug] {
				t.Fatal("duplicate family slug")
			}
			seen[family.slug] = true
			a, err := loadFamilyAssets(filepath.Join("..", ".."), family.slug)
			if err != nil {
				t.Fatal(err)
			}
			if a.family != family.slug {
				t.Fatal("family omitted from evidence inputs")
			}
			for _, facing := range []string{"n", "nw", "sw", "s", "se", "ne"} {
				if a.images["unit/"+facing].Bounds() != image.Rect(0, 0, 192, 192) {
					t.Fatal("unsupported geometry", facing)
				}
			}
		})
	}
	if len(seen) != 15 {
		t.Fatalf("expected 15 delivered families, got %d", len(seen))
	}
	for _, invalid := range []string{"../infantry", "INFANTRY", "missing", ""} {
		if _, err := loadFamilyAssets(filepath.Join("..", ".."), invalid); err == nil || !strings.Contains(err.Error(), "unknown unit family") {
			t.Fatalf("unvalidated family accepted: %q %v", invalid, err)
		}
	}
}
