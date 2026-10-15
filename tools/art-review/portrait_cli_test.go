package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func portraitTestFixture(t *testing.T, width, height int) (string, []byte, []byte) {
	t.Helper()
	dir := t.TempDir()
	im := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			im.SetNRGBA(x, y, color.NRGBA{uint8(x % 251), uint8(y % 241), uint8((x + y) % 239), 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, im); err != nil {
		t.Fatal(err)
	}
	source := encoded.Bytes()
	sum := sha256.Sum256(source)
	manifest, err := json.Marshal(map[string]any{"schema_version": 1, "family": "infantry", "source": map[string]string{"source": "source.png", "source_sha256": hex.EncodeToString(sum[:])}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "input.json")
	unitTestWrite(t, path, manifest)
	unitTestWrite(t, filepath.Join(dir, "source.png"), source)
	return path, manifest, source
}

func TestPortraitCLIExportsOpaquePortrait(t *testing.T) {
	path, _, _ := portraitTestFixture(t, 1024, 1024)
	out := filepath.Join(t.TempDir(), "bundle")
	if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err != nil {
		t.Fatalf("portrait CLI: %v", err)
	}
	entries, err := os.ReadDir(out)
	if err != nil || len(entries) != 7 {
		t.Fatalf("bundle must contain exactly seven files: %v, %d", err, len(entries))
	}
}

func TestPortraitCLIExclusiveExplicitFlags(t *testing.T) {
	path, _, _ := portraitTestFixture(t, 840, 752)
	for _, extra := range [][]string{
		{"-unit="}, {"-pack="}, {"-compound="}, {"-fog="}, {"-ground="}, {"-forest="}, {"-hills="}, {"-shore="}, {"-connected="}, {"-mixed="}, {"-neighborhood="}, {"-candidate="}, {"-reference="}, {"-reference-platform=repo-provisional"}, {"-resampling=area"}, {"positional"},
	} {
		t.Run(extra[0], func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "absent")
			args := append([]string{"-portrait", path, "-out", out}, extra...)
			if run(args, io.Discard) == nil {
				t.Fatal("conflicting explicit flag accepted")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("rejected CLI created output")
			}
		})
	}
	for _, args := range [][]string{{"-portrait="}, {"-portrait", path}, {"-portrait=", "-out", "unused"}, {"-portrait", path, "-out="}} {
		if run(args, io.Discard) == nil {
			t.Fatalf("missing required argument accepted: %v", args)
		}
	}
}
