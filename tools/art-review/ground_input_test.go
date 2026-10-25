package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func groundTestManifest(t *testing.T, path string, materials map[string]connectionMaterial) {
	t.Helper()
	data, err := json.Marshal(map[string]any{"schema_version": 1, "materials": materials})
	if err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, data)
}

func groundTestSavePNG(t *testing.T, path string, im image.Image) connectionMaterial {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, b.Bytes())
	sum := sha256.Sum256(b.Bytes())
	return connectionMaterial{filepath.Base(path), hex.EncodeToString(sum[:])}
}

func groundTestReject(t *testing.T, path, out string, needle string) {
	t.Helper()
	err := run([]string{"-ground", path, "-out", out}, io.Discard)
	if err == nil || (needle != "" && !strings.Contains(err.Error(), needle)) {
		t.Fatalf("invalid input error got %v, want %q", err, needle)
	}
	if _, e := os.Lstat(out); !os.IsNotExist(e) {
		t.Fatalf("failed input created output: %v", e)
	}
}

func TestGroundStrictManifest(t *testing.T) {
	path, refs := groundTestFixture(t)
	valid := string(groundTestRead(t, path))
	cases := map[string]string{
		"null": "null", "array": "[]", "empty": "{}",
		"version-null":       strings.Replace(valid, `"schema_version":1`, `"schema_version":null`, 1),
		"version-string":     strings.Replace(valid, `"schema_version":1`, `"schema_version":"1"`, 1),
		"version-wrong":      strings.Replace(valid, `"schema_version":1`, `"schema_version":2`, 1),
		"version-fraction":   strings.Replace(valid, `"schema_version":1`, `"schema_version":1.5`, 1),
		"version-alias":      strings.Replace(valid, `"schema_version"`, `"Schema_Version"`, 1),
		"version-duplicate":  strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		"unknown-field":      strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"profile":{}`, 1),
		"materials-alias":    strings.Replace(valid, `"materials"`, `"Materials"`, 1),
		"materials-null":     `{"schema_version":1,"materials":null}`,
		"materials-array":    `{"schema_version":1,"materials":[]}`,
		"extra-material":     strings.Replace(valid, `"materials":{`, `"materials":{"extra":null,`, 1),
		"duplicate-material": strings.Replace(valid, `"mud":{`, `"mud":null,"mud":{`, 1),
		"material-alias":     strings.Replace(valid, `"mud":{`, `"Mud":{`, 1),
		"missing-material":   strings.Replace(valid, `"mud":{`, `"snow":{`, 1),
		"ref-null":           strings.Replace(valid, `"mud":{`, `"mud":null,"other":{`, 1),
		"extra-json":         valid + ` {}`,
	}
	source := `"source":"mud.png"`
	pin := `"source_sha256":"` + refs["mud"].SourceSHA256 + `"`
	for name, replacement := range map[string]string{"source-null": `"source":null`, "source-array": `"source":[]`, "source-number": `"source":5`, "source-alias": `"Source":"mud.png"`, "source-empty": `"source":""`, "source-duplicate": source + `,` + source, "ref-extra": source + `,"x":true`} {
		cases[name] = strings.Replace(valid, source, replacement, 1)
	}
	cases["source-missing"] = strings.Replace(valid, source+`,`, "", 1)
	for name, replacement := range map[string]string{"pin-null": `"source_sha256":null`, "pin-array": `"source_sha256":[]`, "pin-number": `"source_sha256":1`, "pin-alias": `"Source_SHA256":"` + refs["mud"].SourceSHA256 + `"`, "pin-empty": `"source_sha256":""`, "pin-duplicate": pin + `,` + pin} {
		cases[name] = strings.Replace(valid, pin, replacement, 1)
	}
	cases["pin-missing"] = strings.Replace(valid, `,`+pin, "", 1)
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			groundTestWrite(t, path, []byte(bad))
			groundTestReject(t, path, filepath.Join(filepath.Dir(path), "out-"+name), "")
		})
	}
}

func TestGroundModeConflicts(t *testing.T) {
	for _, other := range []string{"reference", "candidate", "reference-platform", "neighborhood", "connected", "compound", "shore", "fog", "mixed", "pack", "resampling"} {
		for _, value := range []string{"", "x"} {
			t.Run(other+"/"+value, func(t *testing.T) {
				err := run([]string{"-ground", "m.json", "-out", "unused", "-" + other + "=" + value}, io.Discard)
				if err == nil || !strings.Contains(err.Error(), "excludes all other flags") {
					t.Fatalf("mode conflict %s=%q: %v", other, value, err)
				}
			})
		}
	}
	for _, args := range [][]string{{"-ground=", "-out", "unused"}, {"-ground", "x"}, {"-ground", "x", "-out="}, {"-ground", "x", "-out", "unused", "extra"}} {
		if err := run(args, io.Discard); err == nil || !strings.Contains(err.Error(), "-ground requires") {
			t.Fatalf("invalid CLI %v: %v", args, err)
		}
	}
}

func TestGroundSourcesAndFreshOutput(t *testing.T) {
	path, refs := groundTestFixture(t)
	dir := filepath.Dir(path)
	original := refs["grass"]
	for _, tt := range []struct{ name, source, pin string }{
		{"stale-pin", original.Source, strings.Repeat("0", 64)},
		{"upper-pin", original.Source, strings.ToUpper(original.SourceSHA256)},
		{"short-pin", original.Source, "ab"},
		{"empty-pin", original.Source, ""},
		{"absolute", filepath.Join(dir, original.Source), original.SourceSHA256},
		{"traversal", "../grass.png", original.SourceSHA256},
		{"dot", "./grass.png", original.SourceSHA256},
		{"backslash", `a\grass.png`, original.SourceSHA256},
		{"unclean", "a//grass.png", original.SourceSHA256},
		{"original", "original/grass.png", original.SourceSHA256},
		{"pilot", "PiLoT/grass.png", original.SourceSHA256},
		{"missing", "missing.png", original.SourceSHA256},
	} {
		t.Run(tt.name, func(t *testing.T) {
			refs["grass"] = connectionMaterial{tt.source, tt.pin}
			groundTestManifest(t, path, refs)
			groundTestReject(t, path, filepath.Join(dir, "out-"+tt.name), "")
		})
	}
	refs["grass"] = original
	for _, id := range []string{"grass", "grass_w", "mud", "stone", "rough_w", "stoney_w", "shelled_w"} {
		t.Run("nonopaque-"+id, func(t *testing.T) {
			im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
			draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{50, 60, 70, 255}), image.Point{}, draw.Src)
			im.SetNRGBA(383, 319, color.NRGBA{50, 60, 70, 254})
			old := refs[id]
			refs[id] = groundTestSavePNG(t, filepath.Join(dir, "nonopaque.png"), im)
			groundTestManifest(t, path, refs)
			groundTestReject(t, path, filepath.Join(dir, "out-alpha-"+id), "opaque")
			refs[id] = old
		})
	}
	for _, size := range []image.Point{{383, 320}, {384, 319}, {1, 1}} {
		im := image.NewNRGBA(image.Rectangle{Max: size})
		refs["grass"] = groundTestSavePNG(t, filepath.Join(dir, "small.png"), im)
		groundTestManifest(t, path, refs)
		groundTestReject(t, path, filepath.Join(dir, "out-small"), "at least 384x320")
	}
	refs["grass"] = original
	// Over-budget valid IHDR must be rejected before the short image body is decoded.
	huge := append([]byte(nil), groundTestRead(t, filepath.Join(dir, original.Source))...)
	binary.BigEndian.PutUint32(huge[16:20], 8193)
	binary.BigEndian.PutUint32(huge[20:24], 8192)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	groundTestWrite(t, filepath.Join(dir, "huge.png"), huge)
	sum := sha256.Sum256(huge)
	refs["grass"] = connectionMaterial{"huge.png", hex.EncodeToString(sum[:])}
	groundTestManifest(t, path, refs)
	groundTestReject(t, path, filepath.Join(dir, "out-huge"), "64 megapixels")
	refs["grass"] = original
	outside := t.TempDir()
	contents := groundTestRead(t, filepath.Join(dir, original.Source))
	groundTestWrite(t, filepath.Join(outside, "escape.png"), contents)
	for _, tt := range []struct{ name, target string }{{"escape", filepath.Join(outside, "escape.png")}, {"original-link", filepath.Join(dir, "original", "hidden.png")}, {"pilot-link", filepath.Join(dir, "pilot", "hidden.png")}} {
		if tt.name != "escape" {
			if err := os.MkdirAll(filepath.Dir(tt.target), 0o700); err != nil {
				t.Fatal(err)
			}
			groundTestWrite(t, tt.target, contents)
		}
		link := tt.name + ".png"
		if err := os.Symlink(tt.target, filepath.Join(dir, link)); err != nil {
			t.Fatal(err)
		}
		refs["grass"] = connectionMaterial{link, original.SourceSHA256}
		groundTestManifest(t, path, refs)
		groundTestReject(t, path, filepath.Join(dir, "out-"+tt.name), "")
	}
	refs["grass"] = original
	groundTestManifest(t, path, refs)
	existing := filepath.Join(dir, "existing")
	if err := os.Mkdir(existing, 0o700); err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, filepath.Join(existing, "sentinel"), []byte("preserve"))
	regular := filepath.Join(dir, "regular")
	groundTestWrite(t, regular, []byte("preserve"))
	link, dangling := filepath.Join(dir, "output-link"), filepath.Join(dir, "dangling")
	if err := os.Symlink(existing, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "absent"), dangling); err != nil {
		t.Fatal(err)
	}
	for _, out := range []string{existing, regular, link, dangling} {
		if err := run([]string{"-ground", path, "-out", out}, io.Discard); err == nil {
			t.Fatalf("accepted existing output %s", out)
		}
	}
	if string(groundTestRead(t, filepath.Join(existing, "sentinel"))) != "preserve" || string(groundTestRead(t, regular)) != "preserve" {
		t.Fatal("existing output modified")
	}
}

type groundTestBrokenWriter struct{}

func (groundTestBrokenWriter) Write([]byte) (int, error) {
	return 0, errors.New("ground test writer failure")
}

func TestGroundFailureCleanup(t *testing.T) {
	path, _ := groundTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "rollback")
	if err := run([]string{"-ground", path, "-out", out}, groundTestBrokenWriter{}); err == nil || !strings.Contains(err.Error(), "ground test writer failure") {
		t.Fatalf("writer failure: %v", err)
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatalf("failed new output retained: %v", err)
	}
}

func TestGroundLoadedBytesUsedForSnapshotsAndRendering(t *testing.T) {
	path, _ := groundTestFixture(t)
	in, err := loadGroundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	groundTestWrite(t, path, []byte("changed after load"))
	for _, source := range in.Sources {
		groundTestWrite(t, filepath.Join(filepath.Dir(path), source.Ref.Source), []byte("changed after load"))
	}
	bundle, err := buildGroundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	artifacts := map[string][]byte{}
	for _, artifact := range bundle.artifacts {
		artifacts[artifact.name] = artifact.data
	}
	if !bytes.Equal(artifacts["manifest.json"], in.Bytes) {
		t.Fatal("manifest bytes reread")
	}
	for id, source := range in.Sources {
		if !bytes.Equal(artifacts[bundle.report.Materials[id].RetainedSource], source.Bytes) {
			t.Fatal("source bytes reread")
		}
	}
	// Prove actual exported pixels came from the loaded image despite disk changes.
	var centerProfile groundProfileReport
	for _, p := range bundle.report.Profiles {
		if p.RuntimeID == "mud" && p.Canonical {
			centerProfile = p
		}
	}
	im := groundTestLoadBytes(t, artifacts[centerProfile.Outputs[1].Path])
	p := connectionPoint{48.125, 40.125}
	want := sampleConnectionMaterial(in.Sources["mud"].Image, p)
	if got := im.NRGBAAt(192, 160); got != want {
		t.Fatalf("loaded-source center pixel got %v want %v", got, want)
	}
}

func groundTestLoadBytes(t *testing.T, data []byte) *image.NRGBA {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(decoded.Bounds())
	draw.Draw(im, im.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	return im
}

// A failed invocation owns the directory it created, not an unrelated
// replacement at the same pathname. The log callback gives a deterministic
// interleaving after artifacts have been written and before failure cleanup.
type groundTestReplacingWriter struct {
	t             *testing.T
	out, previous string
}

func (w groundTestReplacingWriter) Write([]byte) (int, error) {
	if err := os.Rename(w.out, w.previous); err != nil {
		w.t.Fatal(err)
	}
	if err := os.Mkdir(w.out, 0o700); err != nil {
		w.t.Fatal(err)
	}
	groundTestWrite(w.t, filepath.Join(w.out, "sentinel"), []byte("unrelated replacement"))
	return 0, errors.New("replacement writer failure")
}

func TestGroundFailureCleanupPreservesReplacedDirectory(t *testing.T) {
	path, _ := groundTestFixture(t)
	dir := filepath.Dir(path)
	out := filepath.Join(dir, "raced-output")
	err := run([]string{"-ground", path, "-out", out}, groundTestReplacingWriter{t, out, filepath.Join(dir, "renamed-owned-output")})
	if err == nil {
		t.Fatal("writer error lost")
	}
	data, readErr := os.ReadFile(filepath.Join(out, "sentinel"))
	if readErr != nil || string(data) != "unrelated replacement" {
		t.Fatalf("failure cleanup deleted an unrelated replacement directory: %v", readErr)
	}
}
