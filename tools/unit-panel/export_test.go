package unitpanel

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStrictInputsAndCLI(t *testing.T) {
	path := "../../art/pipeline/unit-panel-v1.json"
	in, err := loadInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Items) != 11 {
		t.Fatalf("inputs=%d", len(in.Items))
	}
	for _, args := range [][]string{{}, {"-manifest", path, "-out", "x", "extra"}, {"-manifest", path, "-out", "x", "-wat"}, {"-manifest", path, "-manifest", path, "-out", "x"}} {
		if err := Run(args, io.Discard, io.Discard); err == nil {
			t.Fatalf("accepted args %q", args)
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	base := sourceBase(path)
	tests := []struct {
		name string
		mut  func([]byte) []byte
	}{
		{"trailing", func(b []byte) []byte { return append(b, []byte(" {}")...) }},
		{"case-alias", func(b []byte) []byte { return bytes.Replace(b, []byte(`"profile":`), []byte(`"PROFILE":`), 1) }},
		{"nested-case-alias", func(b []byte) []byte { return bytes.Replace(b, []byte(`"source":`), []byte(`"SOURCE":`), 1) }},
		{"unknown", func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"profile":`), []byte(`"surprise": 1, "profile":`), 1)
		}},
		{"duplicate", func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"profile":`), []byte(`"profile": "unit-panel-v1", "profile":`), 1)
		}},
		{"pin", func(b []byte) []byte {
			return bytes.Replace(b, []byte("c18494baa7ea35b8dbfac3861787d360b9c0ab91232562371a09e6fbf09aaa40"), []byte(strings.Repeat("0", 64)), 1)
		}},
		{"escape", func(b []byte) []byte {
			return bytes.Replace(b, []byte("art/fonts/Go-Bold.ttf"), []byte("../Go-Bold.ttf"), 1)
		}},
		{"absolute", func(b []byte) []byte {
			return bytes.Replace(b, []byte("art/fonts/Go-Bold.ttf"), []byte("/tmp/Go-Bold.ttf"), 1)
		}},
		{"legacy", func(b []byte) []byte {
			return bytes.Replace(b, []byte("art/fonts/Go-Bold.ttf"), []byte("original/Go-Bold.ttf"), 1)
		}},
	}
	// Parse altered documents directly; no source directory needs to be writable.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, e := parseInputs(base, tt.mut(raw)); e == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

func TestBundleDeterministicAndSafe(t *testing.T) {
	in, err := loadInputs("../../art/pipeline/unit-panel-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	if err = exportBundle(in, a); err != nil {
		t.Fatal(err)
	}
	if err = exportBundle(in, b); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	err = filepath.WalkDir(a, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(a, path)
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		files[rel] = raw
		other, e := os.ReadFile(filepath.Join(b, rel))
		if e != nil {
			return e
		}
		if !bytes.Equal(raw, other) {
			t.Errorf("nondeterministic %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 70 {
		t.Fatalf("file count=%d", len(files))
	}
	for _, d := range []string{"2x", "4x"} {
		if !bytes.Equal(files["fixtures/empty-"+d+".png"], files["fixtures/hidden-"+d+".png"]) {
			t.Fatal("hidden pixel leak")
		}
	}
	if err = exportBundle(in, a); err == nil {
		t.Fatal("overwrote output")
	}
	for _, item := range in.Items {
		if !bytes.Equal(item.Bytes, files[item.Output]) {
			t.Errorf("snapshot mismatch %s", item.Output)
		}
	}
	var report Report
	if err = json.Unmarshal(files["report.json"], &report); err != nil {
		t.Fatal(err)
	}
	for _, f := range report.Fixtures {
		if len(f.Draws) == 0 {
			t.Fatal("missing draw records")
		}
		for _, r := range f.Draws {
			if r.Kind == "text" && (r.Bounds[0] < r.Clip[0] || r.Bounds[1] < r.Clip[1] || r.Bounds[2] > r.Clip[2] || r.Bounds[3] > r.Clip[3]) {
				t.Fatalf("text outside region %+v", r)
			}
		}
	}
	if !bytes.Equal([]byte(productHTML(report.Fixtures[14].Draws)), []byte(productHTML(report.Fixtures[16].Draws))) {
		t.Fatal("hidden subtree leak")
	}
	bad := *in
	bad.Items = append([]Input(nil), in.Items...)
	bad.Items[0].Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
	out := filepath.Join(root, "failed")
	if err = exportBundle(&bad, out); err == nil {
		t.Fatal("invalid cached input accepted")
	}
	if _, err = os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("failed output remains")
	}
}

func TestUnicodeFitAndVectorGrammar(t *testing.T) {
	in, err := loadInputs("../../art/pipeline/unit-panel-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	face, err := newFace(in.Fonts["regular"], 14, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = face.Close() }()
	got := fitText(face, "Übersetzung 日本語 très longue désignation", 80)
	if !utf8.ValidString(got) || !strings.HasSuffix(got, "…") || textWidth(face, got) > 80 {
		t.Fatalf("bad ellipsis %q", got)
	}
	vs := vectors()
	if len(vs) != 12 {
		t.Fatalf("vectors %d", len(vs))
	}
	for _, v := range vs {
		s := v.svg()
		d := xml.NewDecoder(strings.NewReader(s))
		for {
			tok, e := d.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				t.Fatal(e)
			}
			if tag, ok := tok.(xml.StartElement); ok {
				if tag.Name.Local != "svg" && tag.Name.Local != "path" {
					t.Fatalf("unsupported SVG element %s", tag.Name.Local)
				}
				for _, a := range tag.Attr {
					switch a.Name.Local {
					case "xmlns", "width", "height", "viewBox", "fill", "d":
					default:
						t.Fatalf("unsupported SVG attribute %s", a.Name.Local)
					}
				}
			}
		}
		for _, bad := range []string{"<text", "<image", "foreignObject", "href=", "url("} {
			if strings.Contains(s, bad) {
				t.Fatalf("unsafe svg %s", bad)
			}
		}
	}
}

func TestRollbackKeepsReplacement(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "out")
	if e := os.Mkdir(out, 0o755); e != nil {
		t.Fatal(e)
	}
	owned, e := os.Lstat(out)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Rename(out, filepath.Join(root, "old")); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(out, 0o755); e != nil {
		t.Fatal(e)
	}
	sentinel := filepath.Join(out, "keep")
	if e = os.WriteFile(sentinel, []byte("replacement"), 0o644); e != nil {
		t.Fatal(e)
	}
	cleanupOwned(out, owned)
	if _, e = os.Stat(sentinel); e != nil {
		t.Fatal("replacement removed", e)
	}
	current, e := os.Lstat(out)
	if e != nil {
		t.Fatal(e)
	}
	cleanupOwned(out, current)
	if _, e = os.Stat(out); !os.IsNotExist(e) {
		t.Fatal("owned directory not removed")
	}
	if e = os.Symlink(filepath.Join(root, "old"), out); e != nil {
		t.Fatal(e)
	}
	cleanupOwned(out, owned)
	if _, e = os.Lstat(out); e != nil {
		t.Fatal("replacement symlink removed", e)
	}
}

func TestWrongDimensionsFontsAndSymlink(t *testing.T) {
	in, e := loadInputs("../../art/pipeline/unit-panel-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	for _, i := range in.Items {
		p := filepath.Join(root, i.Source)
		if e = os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(p, i.Bytes, 0o644); e != nil {
			t.Fatal(e)
		}
	}
	tiny, e := pngBytes(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name    string
		content []byte
		font    bool
		link    bool
	}{{"dimensions", tiny, false, false}, {"not-png", []byte("invalid PNG"), false, false}, {"malformed-font", []byte("invalid font"), true, false}, {"symlink-escape", in.Items[0].Bytes, false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			m := in.Manifest
			r := &m.Portrait.Two
			if tc.font {
				r = &m.Fonts.Regular
			}
			r.Source = "assets/invalid-" + tc.name
			r.SHA256 = pin(tc.content)
			p := filepath.Join(root, r.Source)
			if tc.link {
				outside := filepath.Join(t.TempDir(), "outside")
				if e := os.WriteFile(outside, tc.content, 0o644); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(outside, p); e != nil {
					t.Fatal(e)
				}
			} else if e := os.WriteFile(p, tc.content, 0o644); e != nil {
				t.Fatal(e)
			}
			raw, e := json.Marshal(m)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = parseInputs(root, raw); e == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
	raw, e := json.Marshal(in.Manifest)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := parseInputs(root, raw)
	if e != nil {
		t.Fatal(e)
	}
	// Later source path mutation must not affect snapshots, pins, or renders.
	if e = os.WriteFile(filepath.Join(root, in.Items[0].Source), []byte("changed after load"), 0o644); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(t.TempDir(), "bundle")
	if e = exportBundle(loaded, out); e != nil {
		t.Fatal(e)
	}
	snapshot, e := os.ReadFile(filepath.Join(out, loaded.Items[0].Output))
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(snapshot, in.Items[0].Bytes) {
		t.Fatal("source was reread")
	}
}

func TestVectorDrawInputsResolve(t *testing.T) {
	in, e := loadInputs("../../art/pipeline/unit-panel-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	fs, e := makeFixtures()
	if e != nil {
		t.Fatal(e)
	}
	records, e := compose(in, fs[0].Panel, 2, map[string]Asset{})
	if e != nil {
		t.Fatal(e)
	}
	names := map[string]bool{}
	for _, v := range vectors() {
		names[v.Name] = true
	}
	for _, r := range records {
		if strings.HasPrefix(r.Input, "vector-source.json#") {
			name := strings.TrimPrefix(r.Input, "vector-source.json#")
			if !names[name] {
				t.Fatalf("unresolved vector input %s", r.Input)
			}
		}
	}
}
