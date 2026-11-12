package main

import (
	"io/fs"
	"log"
	"os"
)

// openArtDir opens a pack catalog directory. done releases it; the packs
// snapshot their bytes, so it may be called once they are open.
func openArtDir(dir string) (fsys fs.FS, done func(), err error) {
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, nil, err
	}
	return r.FS(), func() {
		if err := r.Close(); err != nil {
			log.Printf("close art directory: %v", err)
		}
	}, nil
}
