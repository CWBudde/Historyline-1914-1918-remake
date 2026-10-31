package referencedesign

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func comparisonOptions(t *testing.T) options {
	t.Helper()
	p := loadFixture(t)
	p.ref.Image.Width, p.ref.Image.Height = 300, 160
	writeImage(t, p)
	level, err := os.ReadFile(filepath.Join("testdata", p.ref.Level))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.root, p.ref.Level), level, 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(p.ref)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.root, "reference.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	assets, err := filepath.Abs("../../assets")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(assets, filepath.Join(p.root, "assets")); err != nil {
		t.Fatal(err)
	}
	return options{root: p.root, ref: "reference.json", compare: filepath.Join(t.TempDir(), "compare")}
}

func TestCompareOutputFormat(t *testing.T) {
	for _, tc := range []struct {
		name, format, extension string
		signature               []byte
	}{
		{"default JPEG", "", ".jpg", []byte{0xff, 0xd8, 0xff}},
		{"explicit JPEG", "jpg", ".jpg", []byte{0xff, 0xd8, 0xff}},
		{"lossless PNG", "png", ".png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := comparisonOptions(t)
			o.compareFormat = tc.format
			if err := run(o); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(o.compare, "compare.json"))
			if err != nil {
				t.Fatal(err)
			}
			var cmp comparison
			if err := json.Unmarshal(raw, &cmp); err != nil {
				t.Fatal(err)
			}
			files := append([]string{}, cmp.Bands...)
			for _, family := range cmp.Families {
				files = append(files, family.Crops...)
			}
			for _, feature := range cmp.Features {
				files = append(files, feature.Crop)
			}
			if len(files) == 0 {
				t.Fatal("comparison has no images")
			}
			for _, name := range files {
				if filepath.Ext(name) != tc.extension {
					t.Errorf("image %s has extension %s, want %s", name, filepath.Ext(name), tc.extension)
				}
				raw, err := os.ReadFile(filepath.Join(o.compare, name))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.HasPrefix(raw, tc.signature) {
					t.Errorf("image %s lacks the expected %s signature", name, tc.extension)
				}
			}
		})
	}
}

func TestCompareRejectsInvalidFormatWithoutReplacingOutput(t *testing.T) {
	o := comparisonOptions(t)
	o.compareFormat = "webp"
	if err := os.MkdirAll(o.compare, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(o.compare, "keep.txt")
	if err := os.WriteFile(keep, []byte("previous comparison"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(o); err == nil || !strings.Contains(err.Error(), "compare-format") {
		t.Fatalf("run error = %v, want invalid compare-format", err)
	}
	raw, err := os.ReadFile(keep)
	if err != nil || string(raw) != "previous comparison" {
		t.Fatalf("previous comparison = %q, %v; want preserved", raw, err)
	}
	entries, err := os.ReadDir(o.compare)
	if err != nil || len(entries) != 1 {
		t.Fatalf("comparison entries = %v, %v; want only previous output", entries, err)
	}
}
