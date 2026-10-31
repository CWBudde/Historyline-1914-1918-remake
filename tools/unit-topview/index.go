package unittopview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
)

// encodePNG returns the PNG bytes of im.
func encodePNG(im image.Image) ([]byte, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// indexLine matches the path and pin of one assets/index.json entry; the index
// keeps one entry per line.
var indexLine = regexp.MustCompile(`"path":"([^"]+)","sha256":"([0-9a-f]{64})"`)

// install writes files (paths relative to the assets directory) and re-pins
// their entries in assets/index.json. Every file must already have an entry
// with its size, so IDs, packs and anchors stay as they are; nothing is
// written when one is missing.
func install(assetsDir string, files map[string]*image.NRGBA) error {
	indexPath := filepath.Join(assetsDir, "index.json")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	encoded := map[string][]byte{}
	pins := map[string]string{}
	for path, im := range files {
		b, err := encodePNG(im)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		encoded[path], pins[path] = b, hex.EncodeToString(sum[:])
	}
	seen := map[string]bool{}
	out := indexLine.ReplaceAllFunc(index, func(m []byte) []byte {
		sub := indexLine.FindSubmatch(m)
		path := string(sub[1])
		pin, ok := pins[path]
		if !ok {
			return m
		}
		seen[path] = true
		return []byte(fmt.Sprintf(`"path":"%s","sha256":"%s"`, path, pin))
	})
	for path := range files {
		if !seen[path] {
			return fmt.Errorf("assets/index.json has no entry for %s", path)
		}
	}
	for path, b := range encoded {
		if err := os.WriteFile(filepath.Join(assetsDir, path), b, 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(indexPath, out, 0o644)
}
