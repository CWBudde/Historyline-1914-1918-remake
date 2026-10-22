// Command hlcampaign writes levels/campaign: the remake level of every
// imported original map of levels/original (origview.Campaign), named like
// the shipped levels, level4_cp.json for map 03.
//
//	go run ./cmd/hlcampaign [-from levels/original] [-out levels/campaign]
//
// just levels-import runs it after importing the maps.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
)

func main() {
	from := flag.String("from", "levels/original", "directory of the imported maps (full/NN.json)")
	out := flag.String("out", "levels/campaign", "output directory")
	flag.Parse()
	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	files, err := generate(os.DirFS(*from), set)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.RemoveAll(*out); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	for name, b := range files {
		if err := os.WriteFile(filepath.Join(*out, name), b, 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("wrote %d levels to %s\n", len(files), *out)
}

// generate returns the campaign level files, by file name, of the maps
// full/NN.json in src.
func generate(src fs.FS, set *data.Set) (map[string][]byte, error) {
	names, err := fs.Glob(src, "full/[0-9][0-9].json")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("no imported maps")
	}
	files := map[string][]byte{}
	for _, name := range names {
		raw, err := fs.ReadFile(src, name)
		if err != nil {
			return nil, err
		}
		l, err := finmap.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path.Base(name), err)
		}
		stem, err := origview.FileName(l)
		if err != nil {
			return nil, err
		}
		lv, err := origview.Campaign(l, set)
		if err != nil {
			return nil, err
		}
		b, err := lv.Marshal()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", stem, err)
		}
		files[stem+".json"] = b
	}
	return files, nil
}
