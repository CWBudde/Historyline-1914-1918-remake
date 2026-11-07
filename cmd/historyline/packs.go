package main

import (
	"os"
	"path/filepath"
)

// packsDir is where just packs (hl assets packs -out build/packs) puts the art.
const packsDir = "build/packs"

// defaultPack is the built pack of that name below packsDir, "" when it is
// not there (the web build passes its packs explicitly).
func defaultPack(name string) string {
	dir := filepath.Join(packsDir, name)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return ""
	}
	return dir
}
